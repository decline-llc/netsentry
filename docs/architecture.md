# NetSentry Architecture

> **Status**: v0.1.0 development snapshot. This document is aligned with the local master plan and the current repository state.

---

## 1. Scope

NetSentry uses a C capture process and a Go engine process connected by a Unix Domain Socket. The current development build proves the core pcap-to-alert path. The v0.1.0 target is to harden that path with tests, storage, metrics, auth, and graceful shutdown.

Hard boundaries for v0.1.0:

- Offline pcap analysis first.
- IPv4 only.
- Ethernet/VLAN/Q-in-Q, TCP and UDP passthrough.
- Per-packet cleartext payload matching only.
- No TLS decryption, TCP stream reassembly, IP fragment reassembly, or IPv6.

---

## 2. Current Runtime Path

```text
pcap file
  -> C capture (libpcap, frame parsing, Base64 payload preview)
  -> UDS JSON lines (hello, heartbeat, packet frames)
  -> Go receiver packet channel
  -> configurable pipeline worker pool
  -> atomic.Pointer rule engine
  -> SQLite alert store with UPSERT aggregation
  -> GET /api/alerts
```

Current implementation notes:

- `capture/src/main.c` opens offline pcaps and sends parsed frames to `/tmp/netsentry.sock`.
- `capture/src/eth_parser.c` handles Ethernet, VLAN/Q-in-Q, IPv4, TCP, and UDP with bounds checks.
- `capture/src/uds_sender.c` formats JSON frames with explicit string escaping, Base64 payload preview encoding, full-line UDS writes, write-error counters, and bounded initial reconnect support.
- `engine/internal/receiver` owns the UDS listener, hello/heartbeat state, and context-aware packet channel.
- `engine/internal/pipeline` is driven by a configurable worker pool that consumes the shared bounded channel, calls the thread-safe rule engine, timestamps alerts, and writes through an `AlertWriter`. SQLite/recovery writes are serialized inside the store.
- `engine/internal/alert` owns the SQLite alert store, aggregation, and SQL-backed alert querying; `engine/internal/api` owns the minimal HTTP router, pagination, request validation, and error envelopes.
- `engine/internal/rule` already uses immutable rule snapshots via `atomic.Pointer[ruleState]`.

---

## 3. IPC Contract

C sends one JSON object per line. String fields are escaped before serialization; `payload_preview` is Base64. Every connection starts with exactly one valid hello; packet and heartbeat frames before hello, duplicate hello frames, and heartbeats whose session ID differs from that connection's hello are rejected by closing only that connection. The C sender repeats hello immediately after every successful reconnect. The Go receiver caps a frame at 64 KiB and verifies strict IPv4 source/destination fields, timestamp microseconds, Base64 validity, the 4096-byte payload ceiling, and decoded-length consistency. Ordinary IPv6 and IPv4-mapped IPv6 text are rejected before queueing.

Control frames:

```json
{"type":"hello","version":"0.1.0","session_id":"...","pid":1234,"hostname":"...","max_payload_len":4096}
```

```json
{"type":"heartbeat","session_id":"...","seq":1,"sent":10,"dropped":0,"parse_errors":0,"buf_util_pct":0,"avg_json_serialize_us":0,"uds_write_errors":0}
```

Packet frame:

```json
{
  "timestamp_sec": 1719300000,
  "timestamp_usec": 123456,
  "src_ip": "10.0.0.3",
  "dst_ip": "10.0.0.2",
  "src_port": 54322,
  "dst_port": 80,
  "protocol": 6,
  "tcp_flags": "ACK|PSH",
  "payload_len": 54,
  "payload_preview": "R0VUIC8...",
  "is_fragment": false,
  "truncated": false
}
```

`payload_preview` is Base64. The Go rule engine decodes it before payload matching.

---

## 4. Rule Engine

The rule engine owns an immutable `ruleState` snapshot:

```go
type Engine struct {
    state atomic.Pointer[ruleState]
}
```

Reload builds a full new state and swaps it with `Store`. Match reads one snapshot with `Load` and does not lock.

Supported rule types in the current code:

- `payload_match`
- `ip_blacklist`
- `port_blacklist`

Current rule semantics:

- `payload_match` enforces `protocols`, `ports`, `direction`, `depth`, and `offset` per rule. Mixed case-sensitive and case-insensitive payload rules are verified per rule after AC candidate matching.
- `ip_blacklist` enforces `ips`, `direction`, and optional `protocols` per rule. Exact IPs and CIDRs stay scoped to the owning rule.
- `port_blacklist` enforces `ports`, `direction`, and optional `protocols` per rule.

Current rule management:

- Rule management can list the active snapshot, create/update/delete rules with seed-file persistence, and hot reload from the configured seed file.

---

## 5. Planned v0.1.0 Architecture

```text
C capture
  -> UDS receiver (internal/receiver)
  -> context-aware packet channel
  -> single worker pipeline
  -> rule engine
  -> alert aggregator
  -> SQLite store
  -> REST API and Prometheus metrics
```

Planned modules:

- `internal/receiver`: UDS listener, hello validation, heartbeat state, and
  bounded connection lifecycle. The active-load full-engine regression
  integrates receiver, pipeline workers, HTTP API, and SQLite startup/shutdown.
- `internal/pipeline`: configurable worker-pool lifecycle and alert flow. Matching can run concurrently while the store serializes recovery-log/SQLite write critical sections.
- `internal/alert`: aggregation, SQLite store, JSONL recovery-log replay, indexed SQL-backed alert filtering/pagination, daily-shard timestamp-based writes, cross-file querying/counting, TTL pruning, old shard cleanup, payload redaction, and file-backed suppressions.
- `internal/api`: router, pagination request parsing, rule CRUD/reload, suppressions API, PSK auth for mutations, errors, health, audit middleware, and metrics.
- `internal/stats`: counters and Prometheus text rendering for process, queue, rule, alert, worker, and capture heartbeat metrics.

---

## 6. Backpressure and Shutdown

Target behavior:

- Packet channel sends should block rather than silently drop.
- Blocking sends must also listen to `ctx.Done()` so shutdown cannot leak goroutines.
- Concurrent UDS handlers are capped by `engine.uds_max_connections`; excess
  accepted clients are closed immediately, and disconnected handlers release
  their slot for capture reconnects.
- Each accepted UDS connection receives the finite
  `engine.uds_read_timeout_seconds` deadline before its first frame. Every
  complete frame refreshes the deadline; idle expiry closes the handler and
  releases its slot without counting a protocol decode failure.
- C reconnect uses exponential backoff, can bound initial offline connection attempts, and counts write errors/dropped frames while disconnected.
- HTTP API bind failures are returned synchronously during startup. Engine
  shutdown waits for the UDS receiver accept loop/connection handlers, every
  pipeline worker, and graceful HTTP API shutdown before the alert store is
  closed.

The active-load full-engine shutdown regression uses the real receiver,
pipeline worker, HTTP API, and SQLite store. It persists one alert, cancels with
a second match deliberately in flight, and verifies bounded teardown, listener
closure, and zero writes after the store-close boundary.

---

## 7. Alert Suppression

Current build:

- `internal/alert` includes a CIDR/exact-IP suppressor component and in-memory suppression manager.
- Suppressions can be scoped by rule ID and source, destination, or either-side IP ranges.
- Suppressions load from `engine.suppressions_file` at startup.
- `/api/suppressions` can list, add, replace, and delete suppressions that apply to newly generated alerts; mutations are persisted to `engine.suppressions_file` when configured.
- `/api/suppressions/reload` hot reloads suppressions from disk and swaps the active filter after validation succeeds.

---

## 8. Storage

Current build:

- SQLite using `modernc.org/sqlite`.
- UPSERT aggregation by `(rule_id, src_ip, dst_ip, dst_port, window_start)`.
- Fixed aggregation window from `engine.alert_aggregation_window`.
- Optional daily shard pathing with `engine.db_shard_daily`, which writes each alert to `engine.db_dir/netsentry-YYYY-MM-DD.db` based on the alert timestamp.
- Before opening an existing non-current daily shard for a write, the store
  runs the same separate read-only integrity and required-schema preflight used
  at primary startup. A corrupt, truncated, unrelated, or incompatible shard is
  not opened for journal/schema initialization and remains byte-for-byte
  unchanged.
- Primary startup, non-current write preflight, and non-current query/count all
  use the same URL-safe SQLite `mode=ro` helper. The helper resolves database
  symlinks before inspecting sidecars and constructing the URI, matching the
  Unix VFS location for WAL/SHM files. When either WAL or SHM already exists,
  it also requests `readonly_shm=1`, so the default Unix VFS cannot update
  reader marks or create, truncate, or rebuild the operator's SHM in place.
  Databases without sidecars retain ordinary `mode=ro` compatibility. Direct
  faults cover an unsupported primary WAL version and an active-owner
  inconsistent SHM under an encoded historical path; rejection preserves the
  database and both sidecars.
- Cross-shard alert querying and alert counting in daily-shard mode; time range filters narrow the shard files scanned before applying the regular SQL filters and API pagination across the merged result.
- Rule, severity, source, and destination query filters explicitly use SQLite
  binary comparison, preserving case-sensitive exact-match behavior even when
  a compatible operator schema declares a different column collation. Protocol
  and MITRE filters remain intentionally case-insensitive.
- Cross-shard query and count operations reuse the current shard's owned
  connection and open every non-current shard with a URL-safe SQLite
  read-only handle. Healthy detached file sets and shards owned by an active
  WAL writer, including an active database reached through a symlink, remain
  visible without changing database, WAL, or SHM bytes.
- Alert row decoding rejects persisted destination ports outside `0..65535`
  before `uint16` conversion and rejects aggregate counts below one. NetSentry
  reports the invalid field instead of clamping, wrapping, repairing, or
  returning a state its writer cannot produce.
- Persisted severity must be exactly `low`, `medium`, `high`, or `critical`.
  Empty, case-variant, and unsupported values fail row decoding rather than
  being normalized, substituted, or exposed outside the public alert enum.
- Persisted aggregate timestamps must satisfy `window_start <= first_seen <=
  last_seen` and must use the exact canonical UTC RFC3339Nano text emitted by
  the SQLite writer. Parseable explicit or non-UTC offsets and redundant
  fractional precision fail before ordering or identity validation because
  SQLite compares these stored columns as text. Aggregation earliest/latest
  selection, latest payload/match selection, ordering/pagination, inclusive
  time filters, and retention pruning convert canonical text to one fixed-width
  nanosecond key before comparison. The primary database has an optional
  expression index for global `last_seen` order/range scans; legacy historical
  shards remain correct through the same expression without writable index
  creation. The reader validates ordering without assuming that historical
  rows used the current aggregation-window duration.
- Persisted alert IDs must equal the canonical identity derived from
  `(rule_id, src_ip, dst_ip, dst_port, window_start)`. Writer normalization and
  row decoding share that derivation; mismatches fail instead of exposing an
  unrelated public identity.
- Persisted `event_id`, `rule_id`, `rule_name`, `protocol`, `src_ip`, and
  `dst_ip` values must contain non-whitespace text. Payload and match text
  remain optional and may be empty. MITRE tactic, technique ID, and technique
  name must be exactly all empty or each contain non-whitespace text; complete
  tuple text is returned unchanged without current-catalog revalidation.
- Persisted protocol text must equal the shared writer format: `TCP`, `UDP`,
  `ICMP`, or `PROTO_<number>` for an otherwise unnamed uint8 IP protocol.
  Case variants, arbitrary names, malformed or out-of-range numbers, and
  numeric aliases of the three named protocols fail shared primary and
  historical row decoding.
- Persisted `src_ip` and `dst_ip` values must also be strict IPv4 addresses.
  Malformed, ordinary IPv6, and IPv4-mapped IPv6 text fails shared primary and
  historical row decoding before aggregation-identity derivation.
- Row-level TTL pruning in the opened database using `engine.alert_retention_days`.
- Startup cleanup of old `netsentry-YYYY-MM-DD.db` daily shard files and their WAL/SHM sidecars when retention is enabled.
- Before journal or schema initialization, an existing non-empty primary
  database must pass read-only SQLite `quick_check` plus required `alerts` and
  `alert_events` table/column definitions and the binary-collated aggregation
  uniqueness contract. Table, column, and index-key identifiers are matched
  case-insensitively, following SQLite name-resolution semantics. Unknown
  nullable columns and unknown `NOT NULL` columns with a non-NULL default
  remain compatible; an unknown mandatory column without a usable default is
  rejected because NetSentry's fixed inserts cannot populate it. Generated
  columns, triggers, `CHECK` constraints, and foreign-key
  relationships attached to or referencing either write-critical table are
  rejected because their expressions, relationships, or side effects can alter
  or reject valid fixed-column writes; equivalent extensions
  confined to unrelated operator tables remain compatible. A failed check
  stops startup with `ErrDatabaseIntegrity`;
  NetSentry does not repair, migrate, truncate, rename, or overwrite the
  rejected file. Missing query indexes remain compatible and are created by
  normal writable initialization after the preflight succeeds.
- Recovery JSONL is read and structurally validated in full before replay.
  Malformed JSON and a non-empty final record without its terminating newline
  stop startup with `ErrRecoveryLogIntegrity`; no valid prefix is persisted and
  the log remains byte-for-byte unchanged. Syntactically valid records also
  reject exact duplicate top-level names and case-variant aliases that target
  the same durable model field before Go's last-value decoding can discard the
  earlier member. Every top-level name must exactly match the current alert
  writer's JSON vocabulary; unknown scalar or nested members and case-variant
  supported names are rejected. One immutable contract is built at package
  initialization from the declared exported `model.Alert` fields, their
  `json` tags, and supported writer types; per-record validation uses that
  contract for names, declaration order, presence, JSON kind, and integral
  encoding policy without runtime reflection. Ignored and unexported fields
  stay outside the contract, while embedded fields, case-insensitive name
  conflicts, composite types, and custom marshalers fail contract construction
  instead of silently bypassing validation. Every field the writer cannot omit
  under the module toolchain's supported `omitempty` and `omitzero` behavior
  must be present; only `raw_payload` may currently be omitted. Every present
  value must use the non-null top-level JSON kind emitted by the writer:
  strings for text and timestamps, and numbers for `dst_port` plus
  `aggregated_count`. Both numeric fields must use canonical unsigned base-10
  integer JSON spelling without an exponent, fractional component, sign, or
  multi-digit leading zero. A present `raw_payload` must be a string. Names
  inside accepted field values are not recursively inspected. Valid logs are
  truncated only after SQLite commits.
- Each decoded recovery record must retain the required identity, time, count,
  and network fields emitted by normalized alert writes. Its `event_id` must
  match the deterministic writer identity used by the replay ledger, and its
  durable `id` must match the normalized aggregation identity. Its `rule_name`
  must contain a non-whitespace character without being normalized.
  `first_seen` and `last_seen` must equal `timestamp`; `window_start` must match
  the configured aggregation window; `aggregated_count` must equal one;
  severity must be exactly `low`, `medium`, `high`, or `critical`; and `src_ip`
  plus `dst_ip` must be strict IPv4 addresses. All four timestamp strings must
  also equal the exact canonical UTC RFC3339Nano form emitted by the recovery
  writer; parseable explicit or non-UTC offsets and redundant fractional
  precision fail before representation-dependent identity validation. MITRE
  tactic, technique ID, and technique name must be either exactly all empty or
  all nonblank; complete tuple text is preserved without current-catalog
  revalidation. Protocol text must equal the shared canonical writer format
  (`TCP`, `UDP`, `ICMP`, or `PROTO_<number>` for another uint8 protocol). An
  altered event identity, missing or blank rule name, empty, case-variant, or
  unsupported severity, alternate timestamp encoding, partial or
  whitespace-only MITRE tuple, noncanonical protocol, or malformed, ordinary
  IPv6, and IPv4-mapped IPv6 address text fails through
  `ErrRecoveryLogIntegrity` before replay begins.
- The complete recovery preflight runs before database-directory creation or
  any writable SQLite open. Startup replays that validated in-memory snapshot
  after initialization rather than rereading the file, so rejected input
  cannot create a missing database or modify a compatible existing database.
- Normal runtime writes revalidate the complete existing recovery log inside
  the serialized write critical section before appending the current batch.
  Structural or semantic failure therefore leaves both the rejected log and
  SQLite unchanged; valid pending records are still persisted with the current
  batch before truncation.
- After the existing log passes, normal runtime writes validate the complete
  newly normalized batch against the same durable record contract before
  appending its first record. Invalid current input therefore cannot partially
  append a valid prefix, change SQLite, or degrade healthy storage, and any
  existing valid pending log remains unchanged.
- Recovery writer batches are fully JSON-encoded and size-checked before the
  log is opened. The reader and writer share a 4 MiB per-record limit, allowing
  valid records above 64 KiB while rejecting oversized writer output before
  any prefix can be appended.
- Storage health tracking marks the store degraded after ordinary SQLite write/query errors and emergency after disk-full, quota, read-only filesystem, or disk I/O failures. Emergency mode stops retrying SQLite writes in the current process after the recovery log is updated when possible, and exposes that state through verbose health and Prometheus gauges.

### Operator-triggered restart-free recovery

R90-60 implements the fail-closed contract designed in R90-57. Recovery is
explicit: no timer, health read, free-space poll, or ordinary write may start
an attempt. An authenticated operator request is the only trigger, and one
request owns the complete synchronous attempt.

| State | Accepted event | Guard and serialized action | Result |
| --- | --- | --- | --- |
| `healthy` | ordinary storage failure | Record the failure without changing recovery ownership. | `degraded` |
| `healthy` | emergency-class storage failure | Preserve the normalized batch in the recovery log when possible, record the failure, and stop SQLite retries. | `emergency` |
| `healthy` | operator recovery request | Recovery is unnecessary; do not touch SQLite or the recovery log. | Conflict; remain `healthy` |
| `degraded` | successful write or full list | Clear the ordinary failure through existing health behavior. | `healthy` |
| `degraded` | emergency-class storage failure | Preserve the batch when possible and stop SQLite retries. | `emergency` |
| `degraded` | operator recovery request | Restart-free recovery is reserved for sticky emergency mode. | Conflict; remain `degraded` |
| `emergency` | ordinary alert batch | Under the existing write critical section, validate and append the batch when possible; do not attempt SQLite. | `ErrStorageEmergency`; remain `emergency` |
| `emergency` | operator recovery request | Atomically claim the sole recovery owner, publish `recovering`, then acquire an exclusive store-lifecycle barrier before preflight or replay. | `recovering` |
| `recovering` | second operator request | Reject immediately; do not queue another attempt or alter evidence. | Conflict; remain `recovering` |
| `recovering` | ordinary store operation | Health reads may report progress; database reads and writes wait on the lifecycle barrier or return on context cancellation. No operation may cross the exclusive recovery boundary. | Remain `recovering` |
| `recovering` | read-only preflight rejection | Before any writable open, validate the complete recovery log and every database/shard required by it through preservation-safe read-only handles. | Record the error; return to `emergency` with rejected artifacts byte-for-byte unchanged |
| `recovering` | writable replay failure, cancellation, or shutdown | Close any candidate handle, retain the complete recovery log, do not clean up database/WAL/SHM files, and record whether writable probing may have changed SQLite sidecars. | Return to `emergency`, or continue shutdown |
| `recovering` | complete success | With no second writable owner, use the store-owned primary handle and serial shard handles to replay the preflighted snapshot idempotently, truncate the recovery log only after all target commits succeed, then publish healthy state. An empty log instead receives a rolled-back write probe. | `healthy` |
| any state | process shutdown | Cancel any recovery owner, drain store operations, and close handles without starting recovery or deleting evidence. | `closed` |
| `closed` | operator recovery request | The store lifecycle has ended. | Unavailable; remain `closed` |

The implementation preserves these invariants:

1. Recovery ownership is a compare-and-set transition from `emergency`; at
   most one caller can own it. A separate lifecycle barrier drains active
   database operations and prevents any goroutine from crossing the recovery
   boundary.
2. No background goroutine retries recovery. A failed or cancelled request
   returns to sticky `emergency`; another attempt requires another authenticated
   operator action.
3. Read-only preflight completes before any writable open. Structural,
   semantic, schema, or sidecar rejection leaves the database, WAL, SHM, and
   recovery-log bytes unchanged.
4. Once writable replay begins, SQLite may legitimately update database or
   sidecar bytes. Failure never triggers deletion, repair, rollback-by-copy, or
   recovery-log truncation. The health record must disclose that boundary.
5. The recovery log is the retry authority. Event IDs make a committed prefix
   safe to replay after cancellation or a later-shard failure; truncation occurs
   only after the entire snapshot is durable.
   When the log is empty, write capability is proven in one `BEGIN IMMEDIATE`
   transaction by inserting a reserved, nonce-scoped `alert_events` probe and
   rolling the transaction back; the probe must leave no durable application
   row. A read-only query is never sufficient recovery proof.
6. Daily-shard recovery preflights every referenced shard before the first
   writable shard open. Shard commits remain serial; a later failure retains
   the complete log so an operator retry can idempotently finish the batch.
7. Shutdown cancels the owner and waits for it to release the lifecycle barrier.
   It must not mark the store healthy, launch a retry, or truncate evidence.
8. Health and metrics may expose `recovering`, attempt time, and the last
   result, but observation alone never changes the state.

Threat boundaries are explicit: mandatory authentication prevents an
unauthenticated caller from forcing storage churn; compare-and-set ownership
turns request replay or trigger spam into conflicts; the lifecycle barrier and
event identity prevent concurrent or duplicate writes; preflight-before-open
closes the rejected-input mutation path; retained logs and the prohibition on
cleanup preserve forensic evidence; and cancellation/shutdown cannot silently
promote a partially recovered store.

Remaining v0.1.0 storage work:

- Automatic disk cleanup remains intentionally unsupported.

All SQL values must use placeholders. Do not format user-controlled values into SQL strings.

For a database or recovery-log startup integrity failure, keep NetSentry
stopped and preserve the database together with any `-wal`, `-shm`, and alert
recovery-log sidecars.
Inspect a copy with SQLite recovery tooling, retain the original as evidence,
and point NetSentry at a new or operator-recovered path only after review.

---

## 9. Observability Target

Current build: zap startup and match logs, verbose health with storage status and available bytes, Prometheus metrics for process/current and high-water queue depth/process-lifetime packet and alert rates/rule latency/alert write latency/alert/storage/worker/capture heartbeat state, structured audit logs for non-GET API requests, optional localhost-only pprof, SQLite JSONL recovery-log replay, and configurable payload preview redaction before alert writes.

v0.1.0 target:

- `/api/metrics` Prometheus endpoint with process counters, process-lifetime packet and alert rate gauges, rule match and alert write latency buckets, current/high-water queue depth, rule/alert/storage gauges, worker counters, and capture heartbeat gauges.
- `/api/health?verbose=true` with capture heartbeat freshness, engine queue/rule counts, storage status and available bytes, and throughput counters.
- Structured JSON logs.
- Localhost-only pprof server.

---

## 10. Testing Target

Current build has Go tests for rule matching/Aho-Corasick including payload protocol/port/direction/depth/offset semantics, engine worker shutdown orchestration, `internal/receiver`, and `internal/pipeline`, C parser tests for short frames, TCP, UDP, VLAN, Q-in-Q, fragments, malformed TCP data offsets, C UDS sender tests for JSON formatting, bounded connection failure, and reconnect lifecycle behavior, plus C microbenchmarks for parser, JSON serialization, and UDS line writes. Receiver tests cover reconnects, blocked channel cancellation, single and multiple active connection shutdown, and package-level goroutine leak checks.

Alert storage tests cover SQLite aggregation windows, nanosecond timestamp aggregation/order/filter/pruning, JSONL recovery-log replay idempotency and structural/semantic validation including model-derived canonical fields, required/optional status, JSON kinds, integral encoding policy, fail-closed unsupported model shapes, duplicate top-level fields, canonical timestamp encodings, severity, rule names, MITRE tuples, and protocol names with byte preservation, required-schema plus non-binary aggregation/write-blocking uniqueness/trigger/generated-column/constraint/foreign-key rejection with byte preservation, compatible case-variant required identifiers and ordinary column/index/unrelated-table extensions, collation-independent exact filters, persisted numeric/severity/timestamp-encoding/timestamp-order/aggregation-identity/required-text/MITRE-tuple validation, optional query-index recreation and timestamp query plans, SQL-backed filtering/pagination, daily-shard cross-file querying/counting, corrupt/truncated/incompatible historical-shard read/write preservation, cross-process corrupt WAL/SHM three-file preservation, active WAL-backed read-only access without SHM mutation, deterministic committed-prefix recovery under later-shard failure and active cancellation with idempotent retry, out-of-order writes, aggregation key separation, canceled write contexts, emergency storage mode and restart replay, journal mode validation, daily shard pathing, row TTL pruning, and old daily shard cleanup. API tests also cover health and metrics alert counts backed by a real daily-shard SQLite store.

The v0.1.0 IPC serializer decision is to retain the current bounded handwritten C JSON formatter instead of adding cJSON. The formatter is narrow, fails closed on buffer exhaustion, Base64-encodes payload previews, and is exercised through unit tests, microbenchmarks, a deterministic ASan boundary for packet/heartbeat/hello escaping, payload, integer, exact-fit, and undersized-buffer inputs, independent JSONL decoding, and e2e heartbeat assertions. Replacing it remains a future option only if sustained fuzzing or production evidence shows a concrete defect.

Remaining validation gaps:

- Sustained external C fuzz campaigns with larger parser and formatter corpora.
- Realistic pcap corpora for throughput, query tuning, and alert-volume behavior beyond synthetic repeat-pcap smoke runs.
- Broader SQLite corruption and fault-injection scenarios beyond current disk-full, read-only, I/O, recovery replay, and emergency-mode tests.
