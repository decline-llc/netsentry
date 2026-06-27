# NetSentry

> **Status**: v0.1.0 development / pre-alpha

NetSentry is a lightweight C/Go network intrusion detection and pcap forensics engine. The current repository is a working development build: it can compile the C capture binary and Go engine, generate a synthetic pcap, pass packets over a Unix Domain Socket, match seed rules, persist aggregated alerts in SQLite, and expose them over a minimal HTTP API.

The project goal remains a small, honest IDS for offline pcap analysis and edge deployments. It is not intended to replace Suricata or Zeek for 10 Gbps production IDS workloads.

---

## Current Working Path

```bash
make quickstart
```

This currently does the following:

1. Builds `bin/netsentry-capture` and `bin/netsentry-engine`.
2. Generates `/tmp/netsentry_test.pcap` using Scapy when available, or a Python stdlib fallback.
3. Starts the Go engine on `:8080` and `/tmp/netsentry.sock`.
4. Runs the C capture binary against the sample pcap.
5. Prints alerts from `GET /api/alerts`.

Expected result in the current seed setup: `5` alerts from SQL injection, Log4Shell, reverse shell, shell command injection, and scanner user-agent rules.

---

## Implemented Today

- C capture skeleton with offline pcap reading, Ethernet/VLAN/IPv4/TCP/UDP parsing, Base64 payload preview, JSON line frames, hello and heartbeat frames.
- Go rule engine using `atomic.Pointer[ruleState]` immutable snapshots.
- Rule types: `payload_match`, `ip_blacklist`, `port_blacklist`.
- A self-contained Aho-Corasick matcher.
- Minimal Go UDS receiver and SQLite alert store with UPSERT aggregation and optional daily DB shard pathing.
- Minimal HTTP endpoints: `/api/health`, `/api/alerts`, and `/api/metrics`.
- Seed rules in canonical wrapped JSON schema, with legacy schema compatibility retained in the loader.

---

## Not Implemented Yet

These are v0.1.0 goals, not current behavior:

- Runtime cross-day DB rotation, cross-day alert querying, WAL replay, TTL cleanup.
- Full Prometheus metric coverage beyond the current basic `/api/metrics`.
- Full `/api/health?verbose=true` component snapshot.
- PSK authentication, audit logs, payload redaction.
- Rules CRUD API, suppressions API, pagination and unified error envelope.
- C-side cJSON serializer, C unit tests, ASan/fuzz targets.
- Docker image and binary release packaging.

---

## Build From Source

Prerequisites: Go 1.21+, GCC 9+, libpcap development headers, make, Python 3.

```bash
sudo apt install -y build-essential gcc make libpcap-dev golang-go python3 curl
make quickstart
```

Scapy is optional for quickstart; `scripts/gen_test_pcap.py` has a stdlib fallback.

---

## API In This Build

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/alerts | python3 -m json.tool
```

Current alert responses are in the development shape:

```json
{
  "alerts": [ ... ],
  "total": 5
}
```

The planned stable API envelope is documented in `docs/api-reference.md` as roadmap work.

---

## Detection Boundaries

| Item | v0.1.0 boundary |
| --- | --- |
| Primary mode | Offline pcap analysis |
| Protocols | Ethernet/VLAN/Q-in-Q, IPv4, TCP, UDP passthrough |
| Rule types | `payload_match`, `ip_blacklist`, `port_blacklist` |
| Not supported | IPv6, TLS decryption, TCP stream reassembly, IP fragment reassembly |
| Known bypasses | Split TCP segments, URL/Unicode encoding, SQL comment insertion |

Payload matching runs on per-packet cleartext payload only. If an attack string is split across TCP segments, NetSentry v0.1.0 will miss it.

---

## Project Layout

```text
capture/    C capture and packet parsing code
engine/     Go engine, rule matcher, models, minimal API
configs/    Runtime config and seed rules
docs/       Architecture, API, and development notes
scripts/    Quickstart pcap generator
```

---

## Development Roadmap

The local authority for future work is `First/NETSENTRY_MASTER_PLAN.md`, which is intentionally ignored by Git and not pushed. Public docs summarize the same direction:

- W2/W3: C parser tests, serializer hardening, heartbeat and reconnect behavior.
- W4/W5: modular Go receiver and pipeline worker.
- W6: rule semantics and Prometheus metrics.
- W7: SQLite alert aggregation and storage.
- W8/W9: full API contract, health, auth, audit, pprof.
- W10-W12: integration tests, graceful shutdown, pressure tests, release prep.

---

## License

MIT. See [LICENSE](LICENSE).
