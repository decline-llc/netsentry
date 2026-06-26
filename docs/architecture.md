# NetSentry — Architecture

> **Version**: v0.1.0 (planned)
> **Status**: Pre-implementation design document

---

## 1. Overview

NetSentry is a production-grade lightweight IDS engine built on a **C/Go dual-language architecture**:

| Language | Role | Rationale |
|----------|------|-----------|
| **C** | libpcap packet capture + protocol parsing | libpcap is a C-native library; C module can be independently benchmarked and deployed as a network probe |
| **Go** | Rule engine + HTTP API + alert storage | goroutines are well-suited for concurrent pipelines; Go module can be independently upgraded and restarted |

---

## 2. Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                      NetSentry v0.1.0                            │
├────────────────────┬─────────────────────────────────────────────┤
│  Packet Capture    │  Rule Engine & Detection (Go)               │
│  (C)               │                                             │
│                    │  ┌──────────┐    ┌────────┐    ┌─────────┐  │
│  libpcap           │  │ Channel  │───▶│ Worker │───▶│ Alerts  │  │
│  pcap_open_offline │  │ (cap=    │    │ single │    │ channel │  │
│  Eth→VLAN→IP→TCP   │  │  10000)  │    │ gorout.│    │         │  │
│  protocol parsing  │  │          │    │        │    │ ┌─────┐ │  │
│  + boundary checks │  │ blocking │    │RWMutex │    │ │aggr.│ │  │
│  cJSON goto_cleanup│  │  send    │    │ AC     │    │ │60s  │ │  │
│  UDS send          │──│──────────│    │ autom. │    │ │UPSR.│ │  │
│                    │  │          │    │ + IP   │    │ └─────┘ │  │
│  heartbeat frames  │  │          │    │ blist  │    │   │     │  │
│  (drops/errors)    │  │          │    │        │    │   ▼     │  │
│  session_id        │  │          │    │        │    │ SQLite  │  │
│                    │  └──────────┘    └────────┘    │ WAL +   │  │
│                    │                                │ per-day │  │
│                    │                                │ sharding│  │
│                    │                                └─────────┘  │
│                    │  ┌──────────────────────────────────────┐   │
│                    │  │  HTTP API                             │   │
│                    │  │  /api/health?verbose=true             │   │
│                    │  │  /api/metrics (Prometheus)           │   │
│                    │  │  /api/alerts  (paginated, ATT&CK)    │   │
│                    │  │  /api/rules + /api/rules/batch        │   │
│                    │  │  /api/suppressions                   │   │
│                    │  └──────────────────────────────────────┘   │
├────────────────────┴─────────────────────────────────────────────┤
│              UDS + JSON line protocol (heartbeat + session_id)   │
│              C→Go: PacketInfo JSON + Heartbeat JSON              │
│              C reconnect: exponential backoff 1s→2s→4s→max 30s  │
├──────────────────────────────────────────────────────────────────┤
│         SQLite WAL mode (rules + alerts, MITRE ATT&CK fields)    │
│         Alerts sharded by day: alerts_YYYY-MM-DD.db → 7-day TTL │
│         Pre-write log: alert_wal.jsonl (fsync + event_id replay) │
└──────────────────────────────────────────────────────────────────┘
```

---

## 3. Data Flow

```
pcap offline file
   │
   ▼
[1] C module: pcap_open_offline → Eth (VLAN tag skip) → IP (fragment check) → TCP
   │  ⚠️ Boundary check before every field dereference (IHL/total_length/data_offset)
   │  ⚠️ VLAN tags (0x8100/0x88A8) looped over to locate real EtherType
   │  ⚠️ IPv4 fragments (MF=1 or offset>0) → skip transport layer, mark is_fragment=true
   │  cJSON serialization (goto cleanup pattern, zero leaks) → UDS line-by-line JSON
   │  Heartbeat frame every 5 s: {"type":"heartbeat","session_id":"...","parse_errors":N,...}
   │  UDS disconnect → exponential backoff reconnect (1s→2s→4s→max 30s)
   │
   ▼ UDS + JSON line protocol
[2] Go module: UDS receive (single goroutine reader)
   │  ├─ PacketInfo JSON → channel buffer (blocking send, no drop)
   │  ├─ Heartbeat JSON → atomic.Value directly (bypass channel)
   │  ▼
   │  Worker (single goroutine, holds RWMutex.RLock):
   │  ├─ ip_blacklist (Go map O(1)) + payload_match (AC automaton, depth=4096 configurable)
   │  ├─ match hit → Alert (with MITRE ATT&CK fields)
   │  ├─ update last_packet_at → atomic.Uint64
   │  └─ traffic stats → atomic.Uint64
   │  ▼
   │  Alert aggregator (single goroutine, in-memory map + periodic cleanup):
   │  ├─ UPSERT: INSERT ... ON CONFLICT(...) DO UPDATE
   │  ├─ same window dup → aggregated_count += 1, last_seen = now
   │  ├─ aggregated_count reaches cap → force finalize
   │  └─ timer every 10s cleans expired windows (last_seen < now - 60s)
   │  ▼
   │  critical alerts → dedicated critical channel + writer goroutine (micro-batch 1s/10)
   │  other alerts   → batch buffer (100 rows / 5 s) → SQLite + truncate pre-write log
   │  ▼
   │  HTTP API: paginated query / ATT&CK filter / Prometheus metrics
   │  /api/health?verbose=true returns full health snapshot
   │  ▼
   │  Rule hot-reload: POST /api/rules → RWMutex.Lock → rebuild AC + IP map → Unlock
   │
   ▼
[3] Graceful shutdown (9-step sequence, see §10)
```

---

## 4. IPC: C → Go Communication

### 4.1 Message Types

| Type | Format | Frequency |
|------|--------|-----------|
| Data frame | `{"timestamp_sec":..., "src_ip":"...", ...}` | Per packet |
| Heartbeat | `{"type":"heartbeat","session_id":"...","seq":N,...}` | Every 5 s |
| Handshake | `{"type":"hello","version":"0.1.0","session_id":"...","pid":N,...}` | Once on connect |

### 4.2 C-side UDS Reconnect State Machine

```
CONNECTED → (send() returns EPIPE/ECONNRESET) → DISCONNECTED
DISCONNECTED → wait backoff → CONNECTING → connect() ok → CONNECTED
                                         → connect() fail → DISCONNECTED (retry)

Backoff: 1s → 2s → 4s → 8s → 16s → 30s (cap)
Retries: unlimited (C reconnects until Go recovers)
While disconnected: incoming packets → dropped++ (no buffering, avoid OOM)
```

### 4.3 PacketInfo JSON Schema

```json
{
  "timestamp_sec":  1719300000,
  "timestamp_usec": 123456,
  "src_ip":         "192.168.1.100",
  "dst_ip":         "10.0.0.1",
  "src_port":       54321,
  "dst_port":       80,
  "protocol":       6,
  "tcp_flags":      "0x18",
  "payload_len":    1400,
  "payload_preview":"R0VUIC8...",
  "is_fragment":    false,
  "truncated":      false
}
```

`payload_preview` is Base64-encoded raw bytes. Go side decodes with `encoding/base64.StdEncoding.DecodeString()` before AC automaton matching.

---

## 5. Concurrency Design

### 5.1 Rule Hot-Reload — `atomic.Pointer` (lock-free)

```go
type RuleEngine struct {
    state atomic.Pointer[ruleState]  // immutable snapshot
}
type ruleState struct {
    acMatcher   *ahocorasick.Matcher
    ipBlacklist map[string]*Rule
}
// ReloadRules: build new ruleState → atomic.Pointer.Store()
// Match:       state := atomic.Pointer.Load() → read snapshot → match
```

### 5.2 Channel Back-Pressure

Blocking send (no `select/default` drop). When channel is full, the UDS reader goroutine blocks → UDS receive buffer fills → C-side `send()` blocks → C-side slows down. Natural end-to-end back-pressure.

```go
select {
case packetCh <- &pkt:
case <-ctx.Done():
    return
}
```

### 5.3 Go Module Pipeline Interfaces

```go
// pipeline/interfaces.go

type Matcher interface {
    ID() string
    Match(pkt *model.Packet) *model.Alert
}

type AlertWriter interface {
    Write(alert *model.Alert) error
    Close() error
}
// v0.1.0: SQLiteWriter
// v0.2.0: KafkaWriter, WebhookWriter, StdoutWriter
```

---

## 6. Alert Aggregation

**Aggregation key**: `(rule_id, src_ip, dst_ip, dst_port, window_start)`
**Time window**: 60 s (configurable)
**Cap per window**: 100 (configurable; forces finalize when reached)

```sql
INSERT INTO alerts (..., aggregated_count, ...)
VALUES (?, ..., 1, ...)
ON CONFLICT(rule_id, src_ip, dst_ip, dst_port, window_start)
DO UPDATE SET
    aggregated_count = aggregated_count + 1,
    last_seen = excluded.last_seen,
    payload_preview = excluded.payload_preview;
```

---

## 7. Storage: Per-Day SQLite Sharding

- Alert databases: `data/alerts_YYYY-MM-DD.db` (7-day TTL auto-delete)
- Rules / suppressions: `data/netsentry.db` (no rotation)
- Pre-write log: `data/alert_wal_YYYY-MM-DD.jsonl` (fsync + `event_id` idempotent replay)
- WAL mode; checkpoint on shutdown and when WAL > 10 MB

### Day-boundary switchover (atomic)

```
1. Detect UTC date change (checked every 30 s)
2. aggregator.FinalizeAll() → flush all in-memory windows
3. Create new day's SQLite file + run DDL
4. atomic.Pointer[sql.DB].Store(newDB)
5. Old DB: WAL checkpoint → close
6. Delete DB files older than 7-day TTL
```

---

## 8. Graceful Shutdown (9-step)

```
Signal (SIGINT/SIGTERM) → context.Cancel()

1. HTTP server.Shutdown(ctx)           — stop new API requests (5 s timeout)
2. UDS listener.Close()                — C-side send() → EPIPE
3. Drain packet channel                — process remaining packets (30 s timeout)
4. ticker.Stop()                       — stop batch-write timer
5. flush batch buffer                  — write buffered alerts
6. aggregator.FinalizeAll()            — finalize all windows (max 3 rounds × 5 s)
7. flush final alerts → SQLite
8. SQLite WAL checkpoint + db.Close()
9. wg.Wait()                           — wait all goroutines (5 s timeout)
→ exit
```

---

## 9. Prometheus Metrics (selected)

| Metric | Type | Description |
|--------|------|-------------|
| `netsentry_packets_received_total` | Counter | Total packets received |
| `netsentry_packets_dropped_total` | Counter | C-side drops (from heartbeat) |
| `netsentry_alerts_total` | CounterVec | Alerts by severity |
| `netsentry_channel_depth` | Gauge | Worker channel queue depth |
| `netsentry_capture_connected` | Gauge | C module connection state (0/1) |
| `netsentry_capture_restarts_total` | Counter | C process restarts (session_id change) |
| `netsentry_actual_throughput_pps` | Gauge | Actual Worker processing speed (PPS) |
| `netsentry_db_dir_free_bytes` | Gauge | Disk free space (bytes), updated every 30 s |
| `netsentry_db_emergency_mode` | Gauge | 1 = disk full, all SQLite writes paused |

Full metric list: see design document §4 (Prometheus Metrics table).

---

## 10. Performance Boundaries (v0.1.0)

| Item | Value |
|------|-------|
| Throughput (offline, single core) | ~50K PPS |
| Throughput (real-time) | ~30K PPS |
| Memory (idle) | ~30–50 MB |
| Memory (full load) | ~80 MB |
| "Zero-copy" scope | Protocol parsing layer only (C module pointer offsets). IPC path has 5 copies. |
| IPv6 | Not supported |
| TCP stream reassembly | Not supported (v0.3.0 roadmap) |
| TLS decryption | Not supported (deploy behind SSL offloader) |

For 10 Gbps line-rate or full protocol stack, use [Snort3](https://github.com/snort3/snort3) or [Suricata](https://github.com/OISF/suricata).

---

## 11. Scalability Path (distributed, for reference)

Single-node → 100-node enterprise IDS:

```
Probes (C module, per node)
    │ gRPC + protobuf
    ▼
Kafka / Redpanda  (decouple capture from detection)
    │
    ▼
Detection workers (Go module, horizontal scale)
    │
    ▼
Elasticsearch / ClickHouse  (time-series alert storage)
    │
    ▼
Grafana / Kibana
```

Key tradeoffs: Kafka for durable replay vs NATS for lower ops overhead; ClickHouse for high compression + fast aggregation vs Elasticsearch for full-text search on payloads.
