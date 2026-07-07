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

The root Makefile defaults Go build cache writes to `/tmp/netsentry-go-cache`
so the standard targets work in restricted home-directory environments. Override
it with `GOCACHE=/path/to/cache` when you want to use a different cache.

---

## 2. Current Build Targets

These targets exist today:

```bash
make build-c       # compile C capture
make build-go      # compile Go engine
make build         # build both binaries
make build-asan    # compile C capture with AddressSanitizer
make test          # C parser/UDS tests + Go race tests
make test-coverage # C tests + Go coverage summary
make deps-check    # verify Go module dependency cache integrity
make docs-check    # scan public docs for retired stale wording
make shell-check   # run shell script syntax checks
make python-check  # run Python script syntax checks
make config-check  # validate repository config, rule, and suppression files
make bench         # C parser/UDS microbenchmarks + Go benchmarks
make fuzz-parser   # deterministic ASan fuzz smoke for the C frame parser
make fuzz-parser-long # longer deterministic ASan fuzz pass for the C frame parser
make e2e-smoke     # deterministic pcap -> SQLite -> API smoke test
make e2e-pressure  # repeat-pcap end-to-end throughput smoke test
make sanitize-pcap # sanitize an Ethernet pcap before sharing it
make dist          # build a local release archive under dist/
make docker-build  # build a local Docker image
make rc-check      # release-candidate verification bundle
make lint          # go vet + optional staticcheck
make quickstart    # build, generate pcap, run engine/capture, print alerts
make asan-test     # C parser tests under AddressSanitizer
make clean
```

Local Docker image builds are available through `make docker-build` and are covered by `make rc-check`. GitHub Actions workflows are present for release-candidate checks and GHCR image publishing; the publish workflow runs `make rc-check` before building the registry image, then only pushes on version tags or explicit manual approval.

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
  alert_recovery_log_path: ""
  rules_seed_file: "configs/rules.json"
  suppressions_file: "configs/suppressions.json"
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
  suppressions.json

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

`configs/suppressions.json` uses the canonical wrapped schema:

```json
{
  "suppressions": [
    {
      "id": "internal-subnet",
      "enabled": true,
      "rule_ids": ["rule-001"],
      "src_cidrs": ["10.0.0.0/24"],
      "dst_cidrs": [],
      "any_cidrs": []
    }
  ]
}
```

The engine loads this file at startup. Suppression create, update, and delete requests persist the full file before swapping the active in-memory filter. `POST /api/suppressions/reload` reloads the file from disk and swaps the active filter after validation succeeds.

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

For C parser hardening work, run the deterministic ASan fuzz smoke:

```bash
make fuzz-parser
# Longer local pass:
FUZZ_LONG_ITERATIONS=1000000 make fuzz-parser-long
```

The harness starts from built-in Ethernet/IP/TCP/UDP, VLAN, Q-in-Q, fragment, short-frame, and malformed TCP-offset seeds, then applies deterministic mutations.

The current benchmark scope, local baseline, and pressure smoke behavior are documented in `docs/performance.md`.

For release-candidate checks, run:

```bash
make rc-check
DOCKER="sudo docker" make rc-check
```

This runs `make shell-check`, `make docs-check`, `make python-check`, `make config-check`, `make deps-check`, `make test`, `make test-coverage`, `make fuzz-parser`, `make e2e-smoke`, `make dist`, release archive checksum/content smoke checks, `make docker-build`, a minimal Docker image content smoke check, and a Docker runtime `/api/health` smoke check. If Docker is unavailable in the current environment, use:

```bash
SKIP_DOCKER=1 make rc-check
```

The `e2e-smoke` step uses a temporary config, Unix socket, API port, and SQLite database, then asserts that the synthetic pcap produces 6 processed packets, 5 alerts, 8 loaded rules, capture heartbeat metrics, and rule/write latency histogram observations.

To create a local release archive:

```bash
make dist
VERSION=0.1.0-rc1 make dist
```

The archive and SHA-256 checksum are written to `dist/`. Generated release archives are ignored by Git. The archive includes generated `RELEASE_NOTES.md` with package contents, quick verification, v0.1.0 boundaries, release-candidate evidence, and links to packaged docs.

To build the local Docker image:

```bash
make docker-build
IMAGE=netsentry:0.1.0-rc1 make docker-build
DOCKER="sudo docker" make docker-build
```

The image contains both `netsentry-engine` and `netsentry-capture`. The default entrypoint starts the engine with `configs/config.yaml`; use `docker run --entrypoint netsentry-capture ...` when you need to run the capture binary from the same image.

For a local coverage snapshot:

```bash
make test-coverage
COVERPROFILE=/tmp/custom-netsentry-coverage.out make test-coverage
```

The target runs the existing C tests, then writes a Go coverage profile to
`/tmp/netsentry-coverage.out` by default and prints the total Go coverage line.
It does not enforce a threshold yet.

To sanitize an Ethernet pcap before sharing it for tests:

```bash
make sanitize-pcap INPUT=/path/to/input.pcap OUTPUT=/tmp/sanitized.pcap
```

The sanitizer preserves pcap timestamps, packet framing, Ethernet/VLAN/IPv4/TCP/UDP structure, ports, and lengths. It replaces MAC addresses, maps IPv4 addresses into the `198.18.0.0/15` benchmark range, overwrites TCP/UDP payload bytes, and zeroes unsupported captured frames.

Current validation baseline:

- C parser and UDS sender unit tests cover malformed frames, VLAN/Q-in-Q, IPv4 fragments, TCP offset errors, reconnect behavior, and write-error accounting.
- Go unit and integration tests cover receiver lifecycle, worker panic isolation, rule semantics, API validation, SQLite aggregation, daily shards, recovery-log replay, and storage degraded/emergency behavior.
- Release-candidate checks run syntax checks, repository configuration validation, dependency verification, C/Go tests, coverage snapshot, deterministic C parser fuzz smoke, e2e smoke, release archive checks, Docker image content smoke, and Docker runtime health smoke.

Remaining test gaps:

- Sustained external C fuzz campaigns against a larger corpus.
- Realistic pcap corpora for throughput and query tuning beyond repeat-pcap smoke runs.
- Broader SQLite corruption/fault-injection scenarios beyond the current disk-full, read-only, I/O, recovery replay, and emergency-mode tests.

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
