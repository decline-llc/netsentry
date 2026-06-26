# NetSentry Development Guide

> **Status**: v0.1.0 development snapshot.

---

## 1. Prerequisites

Required for the current repository:

```bash
sudo apt install -y build-essential gcc make libpcap-dev golang-go python3 curl
```

Optional:

- Scapy, used by `scripts/gen_test_pcap.py` when installed. The script has a stdlib fallback.
- `staticcheck`, used by `make lint` when installed.

---

## 2. Current Build Targets

These targets exist today:

```bash
make build-c       # compile C capture
make build-go      # compile Go engine
make build         # build both binaries
make test          # C parser/UDS tests + Go race tests
make bench         # C parser/UDS microbenchmarks + Go benchmarks
make lint          # go vet + optional staticcheck
make quickstart    # build, generate pcap, run engine/capture, print alerts
make asan-test     # C parser tests under AddressSanitizer
make clean
```

Planned but not implemented yet:

- full `make build-asan` capture binary target
- C fuzz targets
- `make sanitize-pcap`
- Docker/release targets

---

## 3. Quickstart

```bash
make quickstart
```

Expected current behavior:

- Builds both binaries.
- Generates `/tmp/netsentry_test.pcap`.
- Starts the Go engine with `configs/config.yaml`.
- Runs C capture against the generated pcap.
- Prints the JSON response from `/api/alerts`.

Current seed rules should produce 5 alerts.

The capture binary accepts `-c <connect_retries>` to bound initial UDS connection attempts. Offline mode defaults to 5 attempts so a missing engine fails clearly instead of retrying forever; live mode keeps retrying unless `-c` is set.

---

## 4. Configuration

`configs/config.yaml` now matches `engine/internal/config/config.go`:

```yaml
capture:
  mode: "offline"
  offline_file: "/tmp/netsentry_test.pcap"
  payload_preview_len: 4096
  uds_socket_path: "/tmp/netsentry.sock"
  uds_socket_mode: "0600"
  heartbeat_interval: 5

engine:
  uds_socket_path: "/tmp/netsentry.sock"
  channel_buffer_size: 10000
  worker_count: 1
  db_dir: "data"
  db_path: "data/netsentry.db"
  db_journal_mode: "WAL"
  db_busy_timeout: 5000
  rules_seed_file: "configs/rules.json"
  api_port: 8080
  api_auth_enabled: false
  api_auth_token: "${NETSENTRY_API_TOKEN:}"

logging:
  level: "info"
  format: "json"
  engine_log: "logs/engine.log"
```

Environment expansion supports `${ENV_VAR}` and `${ENV_VAR:default}`. At the moment, missing required variables expand to an empty string; validation only rejects an empty API token when `engine.api_auth_enabled` is true.

---

## 5. Current Source Layout

Tracked implementation files today:

```text
capture/
  include/
  src/main.c
  src/eth_parser.c
  src/parser_registry.c
  src/passthrough_parser.c
  src/uds_sender.c

engine/
  cmd/netsentry/main.go
  internal/config/
  internal/pipeline/interfaces.go
  internal/rule/
  internal/signal/
  pkg/model/

configs/
  config.yaml
  rules.json
  rules.example.json

scripts/
  gen_test_pcap.py
```

Empty future directories may exist locally, but only tracked files above should be treated as implemented.

---

## 6. Rule Files

`configs/rules.json` uses the canonical wrapped schema:

```json
{
  "rules": [
    {
      "id": "rule-001",
      "type": "payload_match",
      "config": {
        "keywords": ["union select"],
        "case_insensitive": true
      },
      "mitre_techniques": []
    }
  ]
}
```

The loader still accepts legacy top-level arrays and legacy `payload_match`, `ip_blacklist`, and MITRE scalar fields while old files are migrated.

---

## 7. Testing

Current verification before committing:

```bash
make test
make asan-test
make quickstart
```

For parser performance changes, also run:

```bash
make bench
```

Planned tests:

- More C parser unit tests for malformed input.
- Broader UDS sender tests for edge-case write failures.
- Fuzz targets and full capture ASan build.
- UDS receiver integration tests.
- SQLite aggregation tests once storage is implemented.
- Full graceful shutdown tests.

---

## 8. Development Roadmap

The local authoritative roadmap is `First/NETSENTRY_MASTER_PLAN.md` and is intentionally ignored by Git.

Public summary:

| Stage | Focus |
| --- | --- |
| W2 | C parser tests, ASan target, parser microbenchmarks |
| W3 | serializer hardening, heartbeat fields, reconnect behavior |
| W4 | move UDS listener into `internal/receiver` |
| W5 | explicit worker pipeline |
| W6 | rule semantics and Prometheus metrics |
| W7 | SQLite storage and alert aggregation |
| W8 | stable REST API contract |
| W9 | auth, audit, verbose health, pprof |
| W10-W12 | integration, graceful shutdown, pressure tests, release prep |
