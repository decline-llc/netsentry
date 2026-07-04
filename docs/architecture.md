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
  -> single pipeline worker
  -> atomic.Pointer rule engine
  -> SQLite alert store with UPSERT aggregation
  -> GET /api/alerts
```

Current implementation notes:

- `capture/src/main.c` opens offline pcaps and sends parsed frames to `/tmp/netsentry.sock`.
- `capture/src/eth_parser.c` handles Ethernet, VLAN/Q-in-Q, IPv4, TCP, and UDP with bounds checks.
- `capture/src/uds_sender.c` formats JSON frames with explicit string escaping, Base64 payload preview encoding, full-line UDS writes, write-error counters, and bounded initial reconnect support.
- `engine/internal/receiver` owns the UDS listener, hello/heartbeat state, and context-aware packet channel.
- `engine/internal/pipeline` owns the single worker that consumes packets, calls the rule engine, timestamps alerts, and writes them through an `AlertWriter`.
- `engine/internal/alert` owns the SQLite alert store, aggregation, and SQL-backed alert querying; `engine/internal/api` owns the minimal HTTP router, pagination, request validation, and error envelopes.
- `engine/internal/rule` already uses immutable rule snapshots via `atomic.Pointer[ruleState]`.

---

## 3. IPC Contract

C sends one JSON object per line. String fields are escaped before serialization; `payload_preview` is Base64.

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
- `internal/pipeline`: worker lifecycle and alert flow. Implemented as a single worker in the current build.
- `internal/alert`: aggregation, SQLite store, indexed SQL-backed alert filtering/pagination, daily-shard timestamp-based writes, cross-file querying/counting, TTL pruning, old shard cleanup, payload redaction, and file-backed suppressions. WAL replay remains future work.
- `internal/api`: router, pagination request parsing, rule CRUD/reload, suppressions API, PSK auth for mutations, errors, health, audit middleware, and metrics.
- `internal/stats`: counters and Prometheus text rendering for process, queue, rule, alert, worker, and capture heartbeat metrics.

---

## 6. Backpressure and Shutdown

Target behavior:

- Packet channel sends should block rather than silently drop.
- Blocking sends must also listen to `ctx.Done()` so shutdown cannot leak goroutines.
- C reconnect uses exponential backoff, can bound initial offline connection attempts, and counts write errors/dropped frames while disconnected.
- Full graceful shutdown drains packet and alert buffers with explicit timeouts.

The current development build has basic signal handling and HTTP shutdown, but not the full v0.1.0 drain sequence.

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
- Cross-shard alert querying and alert counting in daily-shard mode; time range filters narrow the shard files scanned before applying the regular SQL filters and API pagination across the merged result.
- Row-level TTL pruning in the opened database using `engine.alert_retention_days`.
- Startup cleanup of old `netsentry-YYYY-MM-DD.db` daily shard files and their WAL/SHM sidecars when retention is enabled.
- Basic storage health tracking marks the store degraded after SQLite write/query errors and exposes that state through verbose health and Prometheus gauges.

Remaining v0.1.0 storage work:

- WAL replay, if the write-ahead JSONL path is kept.
- Automatic disk-full recovery and a fuller emergency-mode policy.

All SQL values must use placeholders. Do not format user-controlled values into SQL strings.

---

## 9. Observability Target

Current build: zap startup and match logs, verbose health with storage status and available bytes, Prometheus metrics for process/current and high-water queue depth/rule latency/alert write latency/alert/storage/worker/capture heartbeat state, structured audit logs for non-GET API requests, optional localhost-only pprof, and configurable payload preview redaction before alert writes.

v0.1.0 target:

- `/api/metrics` Prometheus endpoint with process counters, rule match and alert write latency buckets, current/high-water queue depth, rule/alert/storage gauges, worker counters, and capture heartbeat gauges.
- `/api/health?verbose=true` with capture heartbeat freshness, engine queue/rule counts, storage status and available bytes, and throughput counters.
- Structured JSON logs.
- Localhost-only pprof server.

---

## 10. Testing Target

Current build has Go tests for rule matching/Aho-Corasick including payload protocol/port/direction/depth/offset semantics, `internal/receiver`, and `internal/pipeline`, C parser tests for short frames, TCP, UDP, VLAN, Q-in-Q, fragments, malformed TCP data offsets, C UDS sender tests for JSON formatting, bounded connection failure, and reconnect lifecycle behavior, plus C microbenchmarks for parser, JSON serialization, and UDS line writes. Receiver tests cover reconnects, blocked channel cancellation, single and multiple active connection shutdown, and package-level goroutine leak checks.

Alert storage tests cover SQLite aggregation windows, query index creation, SQL-backed filtering/pagination, daily-shard cross-file querying/counting, out-of-order writes, aggregation key separation, canceled write contexts, journal mode validation, daily shard pathing, row TTL pruning, and old daily shard cleanup. API tests also cover health and metrics counts backed by a real daily-shard SQLite store.

Next layers:

- Broader C parser tests for additional malformed frames.
- Broader reconnect integration tests against the Go engine lifecycle.
- Longer ASan/fuzz runs with broader parser corpora.
- Broader full-engine lifecycle tests across receiver, worker, HTTP, and storage shutdown behavior.
- End-to-end quickstart regression.
- Race tests for rule reload and matching.
