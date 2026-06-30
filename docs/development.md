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
make build-asan    # compile C capture with AddressSanitizer
make test          # C parser/UDS tests + Go race tests
make bench         # C parser/UDS microbenchmarks + Go benchmarks
make e2e-smoke     # deterministic pcap -> SQLite -> API smoke test
make e2e-pressure  # repeat-pcap end-to-end throughput smoke test
make dist          # build a local release archive under dist/
make docker-build  # build a local Docker image
make rc-check      # release-candidate verification bundle
make lint          # go vet + optional staticcheck
make quickstart    # build, generate pcap, run engine/capture, print alerts
make asan-test     # C parser tests under AddressSanitizer
make clean
```

Planned but not implemented yet:

- C fuzz targets
- `make sanitize-pcap`
- published Docker image workflow

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

Latest local quickstart verification:

```text
Run date: 2026-06-30
Result: generated 6 packets, processed them through C capture -> UDS -> Go engine -> SQLite -> API, and returned 5 alerts.
```

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
  health_freshness_limit_seconds: 30
  pprof_enabled: false
  pprof_addr: "127.0.0.1:6060"

logging:
  level: "info"
  format: "json"
  engine_log: "logs/engine.log"
```

Environment expansion supports `${ENV_VAR}` and `${ENV_VAR:default}`. At the moment, missing required variables expand to an empty string; validation only rejects an empty API token when `engine.api_auth_enabled` is true.

---

## 5. Current Source Layout

Tracked implementation areas today:

```text
capture/
  include/              public C parser, packet, and UDS sender headers
  src/                  capture CLI, Ethernet/VLAN parser, passthrough parser, UDS sender
  tests/                parser tests, UDS sender tests, and C microbenchmarks

engine/
  cmd/netsentry/        engine entrypoint and process wiring
  internal/alert/       SQLite store, aggregation, suppressor, payload redaction
  internal/api/         HTTP router, handlers, pagination, errors, audit middleware
  internal/config/      YAML config loading, environment expansion, validation
  internal/pipeline/    single-worker packet processing loop
  internal/receiver/    Unix socket listener and heartbeat/session state
  internal/rule/        rule loader, rule engine, and Aho-Corasick matcher
  internal/signal/      shutdown signal context helper
  internal/stats/       atomic counters and Prometheus text rendering
  pkg/model/            shared packet, alert, and rule models

configs/
  config.yaml
  rules.json
  rules.example.json

scripts/
  e2e_smoke.sh
  e2e_pressure.sh
  gen_test_pcap.py
  package_release.sh
  rc_check.sh
```

Empty future directories may exist locally; treat only directories with tracked source files as implemented.

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
make build-asan
make quickstart
```

For parser performance changes, also run:

```bash
make bench
```

For changes that may affect the full offline pipeline, also run:

```bash
make e2e-pressure
# Optional larger run:
PRESSURE_REPEATS=10000 make e2e-pressure
```

The current benchmark scope, local baseline, and pressure smoke behavior are documented in `docs/performance.md`.

For release-candidate checks, run:

```bash
make rc-check
```

This runs shell syntax checks, `make test`, `make e2e-smoke`, `make dist`, release archive checksum/content smoke checks, `make docker-build`, a minimal Docker image content smoke check, and a Docker runtime `/api/health` smoke check. If Docker is unavailable in the current environment, use:

```bash
SKIP_DOCKER=1 make rc-check
```

The `e2e-smoke` step uses a temporary config, Unix socket, API port, and SQLite database, then asserts that the synthetic pcap produces 6 processed packets, 5 alerts, 8 loaded rules, and capture heartbeat metrics.

To create a local release archive:

```bash
make dist
VERSION=0.1.0-rc1 make dist
```

The archive and SHA-256 checksum are written to `dist/`. Generated release archives are ignored by Git.

To build the local Docker image:

```bash
make docker-build
IMAGE=netsentry:0.1.0-rc1 make docker-build
DOCKER="sudo docker" make docker-build
```

The image contains both `netsentry-engine` and `netsentry-capture`. The default entrypoint starts the engine with `configs/config.yaml`; use `docker run --entrypoint netsentry-capture ...` when you need to run the capture binary from the same image.

Planned tests:

- More C parser unit tests for malformed input.
- Broader UDS sender tests for edge-case write failures.
- C fuzz targets.
- Broader UDS receiver integration tests for multi-session lifecycle.
- SQLite aggregation tests for alert storage changes.
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
| W6 | payload rule semantics, remaining rule semantics, and Prometheus metrics |
| W7 | SQLite storage and alert aggregation |
| W8 | stable REST API contract |
| W9 | auth, audit, verbose health, pprof |
| W10-W12 | integration, graceful shutdown, pressure tests, release prep |
