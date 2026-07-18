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

C sends one JSON object per line. String fields are escaped before serialization; `payload_preview` is Base64. The Go receiver caps a frame at 64 KiB and verifies IP fields, timestamp microseconds, Base64 validity, the 4096-byte payload ceiling, and decoded-length consistency.

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

- `internal/receiver`: UDS listener, hello validation, heartbeat state. Implemented in the current build; broader Go engine lifecycle integration remains future work.
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
  runs the same separate read-only `PRAGMA quick_check` used at primary
  startup. A corrupt or truncated shard is not opened for journal/schema
  initialization and remains byte-for-byte unchanged.
- Cross-shard alert querying and alert counting in daily-shard mode; time range filters narrow the shard files scanned before applying the regular SQL filters and API pagination across the merged result.
- Cross-shard query and count operations reuse the current shard's owned
  connection and open every non-current shard with a URL-safe SQLite
  `mode=ro` handle. Malformed historical input can fail a read but cannot be
  modified by that read-only open; active WAL-backed shards remain visible.
- Row-level TTL pruning in the opened database using `engine.alert_retention_days`.
- Startup cleanup of old `netsentry-YYYY-MM-DD.db` daily shard files and their WAL/SHM sidecars when retention is enabled.
- Before journal or schema initialization, an existing non-empty primary
  database must pass read-only SQLite `quick_check`. A failed check stops
  startup with `ErrDatabaseIntegrity`; NetSentry does not repair, truncate,
  rename, or overwrite the rejected file.
- Storage health tracking marks the store degraded after ordinary SQLite write/query errors and emergency after disk-full, quota, read-only filesystem, or disk I/O failures. Emergency mode stops retrying SQLite writes in the current process after the recovery log is updated when possible, and exposes that state through verbose health and Prometheus gauges.

Remaining v0.1.0 storage work:

- Automatic disk cleanup or restart-free recovery after emergency mode.

All SQL values must use placeholders. Do not format user-controlled values into SQL strings.

For a startup integrity failure, keep NetSentry stopped and preserve the
database together with any `-wal`, `-shm`, and alert recovery-log sidecars.
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

Alert storage tests cover SQLite aggregation windows, JSONL recovery-log replay idempotency, query index creation, SQL-backed filtering/pagination, daily-shard cross-file querying/counting, corrupt/truncated historical-shard read/write preservation, active WAL-backed read-only access, out-of-order writes, aggregation key separation, canceled write contexts, emergency storage mode and restart replay, journal mode validation, daily shard pathing, row TTL pruning, and old daily shard cleanup. API tests also cover health and metrics counts backed by a real daily-shard SQLite store.

The v0.1.0 IPC serializer decision is to retain the current bounded handwritten C JSON formatter instead of adding cJSON. The formatter is narrow, fails closed on buffer exhaustion, Base64-encodes payload previews, and is already exercised through unit tests, microbenchmarks, deterministic fuzz smoke, and e2e heartbeat assertions. Replacing it remains a future option only if sustained fuzzing or production evidence shows a concrete defect.

Remaining validation gaps:

- Sustained external C fuzz campaigns with larger parser and formatter corpora.
- Realistic pcap corpora for throughput, query tuning, and alert-volume behavior beyond synthetic repeat-pcap smoke runs.
- Broader SQLite corruption and fault-injection scenarios beyond current disk-full, read-only, I/O, recovery replay, and emergency-mode tests.
