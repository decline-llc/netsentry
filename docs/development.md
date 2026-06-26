# NetSentry — Development Guide

> **Version**: v0.1.0 (planned)

---

## 1. Prerequisites

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install -y build-essential gcc make libpcap-dev golang-go python3 python3-pip curl jq
pip3 install scapy

# Verify
gcc --version        # >= 9.0
go version           # >= 1.21
pcap-config --version
```

| Dependency | Minimum | Purpose |
|------------|---------|---------|
| OS | Linux 4.x+, WSL2 (experimental), macOS (experimental) | Runtime |
| Go | 1.21+ | Compile `engine/` |
| gcc | 9.0+ (`-std=c11`) | Compile `capture/` |
| libpcap-dev | 1.9+ | Packet capture headers |
| make | 4.0+ | Build system |
| Python 3 + Scapy | 3.8+ | `make quickstart` test pcap generation |
| curl + jq | any | API verification |
| Disk | ~200 MB | Build deps + test data |

---

## 2. Quick Start

```bash
git clone https://github.com/yourusername/netsentry.git
cd netsentry
make quickstart
```

`make quickstart` does the following automatically:

```
[1/5] Building C capture module...   (gcc -std=c11 -O2)
[2/5] Building Go engine...           (go build -race)
[3/5] Generating test pcap (scapy)... (25 packets)
[4/5] Starting engine + capture...
[5/5] Verifying API...                ok
=== Ready! ===
curl http://localhost:8080/api/alerts | jq .
```

### Docker (recommended for end users)

```bash
docker run --rm -v $(pwd)/mypcap.pcap:/data/input.pcap:ro \
    ghcr.io/yourusername/netsentry:v0.1.0 \
    --pcap /data/input.pcap --output table
```

No compiler required. Image target size < 20 MB.

### CLI Output Modes

| Flag | Output | Use Case |
|------|--------|----------|
| `--output table` (default) | ANSI color table in terminal | Quick pcap triage |
| `--output json` | One JSON object per line | `\| jq` pipelines |
| `--output jsonl` | Same as `json` | Log ingestion |
| `--serve` | Start HTTP server on port 8080 | Persistent deployment |

CLI mode (`--output table/json/jsonl`) does **not** start SQLite — alerts go to stdout only. REST API mode (`--serve`) enables full SQLite persistence.

### Binary Release

```bash
curl -L https://github.com/yourusername/netsentry/releases/download/v0.1.0/netsentry-linux-amd64.tar.gz | tar xz
./netsentry --quickstart
curl http://localhost:8080/api/alerts | jq .
```

---

## 3. Build Targets

```makefile
make quickstart       # compile + generate pcap + start + verify
make build            # compile C + Go (no -race)
make build-race       # compile Go with -race (development)
make build-asan       # compile C with ASan (memory debugging)
make test             # run all tests
make bench            # run Go benchmarks (W2 micro-benchmarks)
make fuzz             # LibFuzzer on C parsers (clang required)
make sanitize-pcap    # anonymize raw/ pcaps → sanitized/
make lint             # golangci-lint + gcc -Werror
make clean
```

---

## 4. Configuration (`config.yaml`)

```yaml
capture:
  mode: "offline"                          # "offline" | "live" (v0.2.0)
  offline_file: "testdata/attack_traffic.pcap"
  payload_preview_len: 4096                # max payload bytes sent over UDS
  uds_socket_path: "/tmp/netsentry.sock"
  uds_socket_mode: "0600"                  # socket file permissions
  heartbeat_interval: 5                    # heartbeat every N seconds

engine:
  uds_socket_path: "/tmp/netsentry.sock"
  channel_buffer_size: 10000
  worker_count: 1                          # single goroutine in v0.1.0
  db_dir: "data/"
  db_path: "data/netsentry.db"
  db_journal_mode: "WAL"
  db_busy_timeout: 5000
  rules_seed_file: "configs/rules.json"
  api_port: 8080
  cors_allowed_origins: ["http://localhost:3000"]
  alert_aggregation_window: 60             # seconds
  alert_aggregation_max_count: 100
  alert_retention_days: 7
  api_auth_token: "${NETSENTRY_API_TOKEN}" # env var injection
  api_auth_enabled: true
  redact_sensitive_fields: true

logging:
  level: "info"
  format: "json"                           # "json" (prod) | "text" (dev)
  engine_log: "logs/engine.log"
```

### Environment Variable Syntax

```yaml
# Required — fatal if not set
api_auth_token: "${NETSENTRY_API_TOKEN}"

# Optional with default
db_dir: "${NETSENTRY_DATA_DIR:/var/lib/netsentry}"
```

---

## 5. Project Structure

```
netsentry/
├── capture/                     # C module
│   ├── src/
│   │   ├── main.c
│   │   ├── eth_parser.c         # VLAN tag skip loop
│   │   ├── ip_parser.c          # fragment check
│   │   ├── tcp_parser.c         # per-field boundary checks
│   │   ├── json_serializer.c    # goto cleanup + cJSON_PrintBuffered
│   │   ├── uds_client.c         # data + heartbeat, exponential backoff
│   │   ├── heartbeat.c          # timerfd_create + poll
│   │   └── signal_handler.c     # volatile sig_atomic_t
│   ├── include/
│   │   ├── packet_types.h
│   │   ├── parser_registry.h    # (protocol, port) key
│   │   └── uds_client.h
│   └── tests/
│       ├── test_parser.c        # malformed + VLAN + fragment cases
│       └── test_serializer.c    # full-quote escape boundary
├── engine/                      # Go module
│   ├── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── receiver/            # UDS listener, heartbeat (atomic.Value)
│   │   ├── pipeline/
│   │   │   └── interfaces.go    # Matcher + AlertWriter interfaces
│   │   ├── rule/
│   │   │   ├── engine.go        # atomic.Pointer[ruleState] lock-free reload
│   │   │   ├── ip_blacklist.go
│   │   │   ├── payload_match.go # depth/offset/case_insensitive
│   │   │   └── ac_builder.go
│   │   ├── alert/
│   │   │   ├── aggregator.go    # UPSERT + expiry window cleanup
│   │   │   ├── suppressor.go    # CIDR Trie matching
│   │   │   └── walog.go         # fsync + idempotent replay + rotation
│   │   ├── api/
│   │   │   ├── router.go, auth.go
│   │   │   ├── metrics.go       # Prometheus endpoint
│   │   │   ├── validator.go     # CIDR / ATT&CK whitelist validation
│   │   │   ├── errors.go        # unified error response format
│   │   │   └── audit.go         # audit log middleware
│   │   └── signal/              # context.WithCancel + WaitGroup
│   ├── pkg/model/               # packet.go, alert.go, rule.go
│   └── tests/
│       ├── test_matcher.go
│       ├── test_concurrency.go  # hot-reload + concurrent match
│       ├── test_aggregator.go   # UPSERT idempotency
│       └── test_walog.go        # truncated line tolerance + replay
├── configs/
│   ├── rules.json               # seed rules (10–15 curated)
│   └── rules.example.json
├── testdata/
│   ├── raw/                     # .gitignored, must sanitize before use
│   └── sanitized/               # safe to commit
├── scripts/
│   ├── gen_test_pcap.py
│   ├── sanitize_pcap.py
│   └── quickstart.sh
└── data/                        # runtime: .db files, UDS socket, WAL logs
```

---

## 6. Testing

### Run all tests

```bash
make test
# equivalent to:
go test ./... -race -count=1
cd capture && make test   # Unity-based C tests
```

### Test layers

| Layer | Tool | When |
|-------|------|------|
| Unit | C: Unity / Go: `testing` + `-race` | Every commit |
| Integration | Go: `httptest` + temp dir | Every push |
| Fuzzing | LibFuzzer / AFL++ | CI daily |
| Race detection | `go test -race -count=100` | Every commit |
| Memory | Valgrind + ASan | Every push |
| Benchmarks | `go test -bench=.` | Every release |

### Key test scenarios

| Test | What it verifies |
|------|-----------------|
| VLAN tag | Single / Q-in-Q double tag → correct IP header offset |
| IPv4 fragment | First / non-first / no fragment → `is_fragment` flag correct |
| cJSON leak | Valgrind 10 000 packets → 0 leaks |
| Full-quote escape | 4096-byte all-`"` payload → 12 KB buffer safe, valid JSON |
| Rule hot-reload race | `go test -race`: Worker match + 1000× concurrent `POST /api/rules` → 0 races |
| UPSERT idempotency | Same key × 100 inserts → 1 row, `aggregated_count = 100` |
| Pre-write log | Truncated line → skip; duplicate `event_id` → idempotent skip |
| UDS reconnect | Kill Go → C exponential backoff → Go restart → resume |
| Graceful shutdown | SIGINT → 9-step sequence complete → 0 goroutine leaks |

### Benchmarks (W2 targets)

```bash
make bench
# P99 < 30 μs end-to-end, throughput >= 45 K PPS over 60 s
```

---

## 7. pcap Sanitization

Never commit raw production pcaps. Always sanitize first:

```bash
# Place raw pcaps in testdata/raw/ (git-ignored)
make sanitize-pcap
# Output: testdata/sanitized/*.pcap (safe to commit)
```

The sanitizer (`scripts/sanitize_pcap.py`) masks the last IP octet, randomizes ports, and replaces payload bytes with `0x41` while preserving protocol structure and packet length.

---

## 8. CI / CD

```yaml
# .github/workflows/ci.yml — checks per push
- go vet + golangci-lint
- go test -race -cover ./...
- gcc -Wall -Wextra -Werror (C module)
- ASan build + malformed packet tests
- Valgrind memory leak check
- go-licenses check (license compliance)
- LibFuzzer run (1 hour, daily schedule)
```

```yaml
# .github/workflows/release.yml — on tag push
- Build static C binary:  gcc -std=c11 -O2 -static
- Build static Go binary: CGO_ENABLED=0 go build -ldflags="-s -w"
- Build Docker multi-arch (linux/amd64, linux/arm64) → ghcr.io
- Package tarball → GitHub Release
```

---

## 9. Git Workflow

- **Branches**: `main` (stable), `feat/*`, `fix/*`
- **Commits**: conventional commits (`feat:`, `fix:`, `perf:`, `docs:`)
- **PRs**: template at `.github/PULL_REQUEST_TEMPLATE.md`
- **Issues**: `bug_report.md` and `feature_request.md` templates

Good First Issues are labeled `good first issue` with detailed acceptance criteria. See [CONTRIBUTING.md](../CONTRIBUTING.md).

---

## 10. Development Milestones (v0.1.0, 12 weeks)

| Week | Focus | Deliverable |
|------|-------|-------------|
| W1 | Scaffold + Makefile + CI + Unity integration | Buildable skeleton |
| W2 | C: Eth/IP/TCP parsing + boundary checks + **micro-benchmarks** | Protocol parsers + performance baseline |
| W3 | C: cJSON goto cleanup + 12 KB buffer + heartbeat + UDS reconnect | Serialization zero-leaks + reconnect FSM |
| W4 | IPC: UDS data + heartbeat → Go receive + health monitor | End-to-end UDS link |
| W5 | Go: channel + Worker + **atomic.Pointer rule engine** + pipeline interfaces | Concurrent-safe detection pipeline |
| W6 | Go: ip_blacklist + AC automaton + Prometheus endpoint | Matchable rules + metrics visible |
| W7 | Go: alert aggregator (UPSERT) + suppressor (CIDR Trie) + per-day SQLite + MITRE ATT&CK model | Alert dedup + storage + ATT&CK mapping |
| W8 | API: pagination + unified error format + input validation + cross-day query | Full API implementation |
| W9 | `/api/health?verbose` + structured logs + audit log + PSK auth + pprof | Observability + security baseline |
| W10 | End-to-end integration + graceful shutdown tests + goroutine leak tests + race tests | Integration validated |
| W11–12 | Pressure tests + performance tuning + docs | v0.1.0 release candidate |

**v0.1.1** (separate milestone): config encryption tool, SIGHUP hot-reload, pcap sanitization tool.

**v0.2.0** (separate milestone, 6–8 weeks): live capture + FlatBuffers IPC + HTTP/DNS parsers + JWT auth + multi-worker pipeline.

**v0.3.0** (separate milestone, 8–10 weeks): TCP stream reassembly + Web dashboard + Docker + cross-day query via SQLite ATTACH.

---

## 11. Troubleshooting

### `libpcap not found`

```bash
sudo apt install libpcap-dev       # Ubuntu/Debian
sudo yum install libpcap-devel     # RHEL/CentOS
```

### Port 8080 already in use

```bash
lsof -i :8080
# or change: config.yaml → engine.api_port: 8081
```

### UDS socket permission denied

```bash
# Stale socket from unclean shutdown
rm /tmp/netsentry.sock
make quickstart
```

### `sqlite3: database is locked`

Delete stale WAL files and restart:

```bash
rm data/netsentry.db-wal data/netsentry.db-shm
make quickstart
```

### C module segfault

```bash
make build-asan
./netsentry-capture-asan --pcap crash.pcap
# Then check dmesg and open a GitHub issue with the pcap (sanitized)
```

### Go module dependency download failure

```bash
# China mainland
GOPROXY=https://goproxy.cn,direct go mod download

# International
GOPROXY=https://proxy.golang.org,direct go mod download
```
