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
  -> Go engine minimal UDS listener
  -> atomic.Pointer rule engine
  -> in-memory alert store
  -> GET /api/alerts
```

Current implementation notes:

- `capture/src/main.c` opens offline pcaps and sends parsed frames to `/tmp/netsentry.sock`.
- `capture/src/eth_parser.c` handles Ethernet, VLAN/Q-in-Q, IPv4, TCP, and UDP with bounds checks.
- `capture/src/uds_sender.c` formats JSON frames with explicit string escaping, Base64 payload preview encoding, full-line UDS writes, write-error counters, and bounded initial reconnect support.
- `engine/cmd/netsentry/main.go` currently hosts the UDS listener, in-memory alert store, and minimal HTTP API. This is intentionally temporary and will be split into `internal/receiver`, `internal/pipeline`, `internal/alert`, and `internal/api`.
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

Known implementation gaps to close before v0.1.0:

- `protocols`, `ports`, `direction`, `depth`, and `offset` are part of the schema but are not fully enforced for `payload_match` yet.
- Mixed case-sensitive and case-insensitive payload rules currently share one matcher; this needs per-rule-correct behavior.

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

- `internal/receiver`: UDS listener, hello validation, heartbeat state.
- `internal/pipeline`: worker lifecycle and alert flow.
- `internal/alert`: aggregation, SQLite store, optional WAL replay.
- `internal/api`: router, pagination, errors, health, metrics, auth.
- `internal/stats`: counters and Prometheus collectors.

---

## 6. Backpressure and Shutdown

Target behavior:

- Packet channel sends should block rather than silently drop.
- Blocking sends must also listen to `ctx.Done()` so shutdown cannot leak goroutines.
- C reconnect uses exponential backoff, can bound initial offline connection attempts, and counts write errors/dropped frames while disconnected.
- Full graceful shutdown drains packet and alert buffers with explicit timeouts.

The current development build has basic signal handling and HTTP shutdown, but not the full v0.1.0 drain sequence.

---

## 7. Storage Target

Current build: in-memory alerts only.

v0.1.0 target:

- SQLite using `modernc.org/sqlite`.
- UPSERT aggregation by `(rule_id, src_ip, dst_ip, dst_port, window_start)`.
- Daily alert DB files when implemented.
- No unbounded memory buffering on disk-full conditions.

All SQL values must use placeholders. Do not format user-controlled values into SQL strings.

---

## 8. Observability Target

Current build: zap startup and match logs.

v0.1.0 target:

- `/api/metrics` Prometheus endpoint.
- `/api/health?verbose=true` with capture, engine, storage, and throughput status.
- Structured JSON logs.
- Localhost-only pprof server.

---

## 9. Testing Target

Current build has Go tests for rule matching/Aho-Corasick, C parser tests for short frames, TCP, UDP, VLAN, Q-in-Q, fragments, malformed TCP data offsets, C UDS sender tests for JSON formatting and bounded connection failure, plus C parser microbenchmarks for plain TCP, VLAN TCP, and Q-in-Q TCP frames.

Next layers:

- Broader C parser tests for additional malformed frames.
- JSON serialization and UDS send microbenchmarks.
- UDS reconnect integration tests against a real listener lifecycle.
- Full ASan capture binary target.
- UDS receiver tests for hello, heartbeat, bad JSON, context cancellation.
- End-to-end quickstart regression.
- Race tests for rule reload and matching.
