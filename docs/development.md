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

Pinned supply-chain tools used by CI:

```bash
(cd engine && go install golang.org/x/vuln/cmd/govulncheck@v1.6.0)
(cd engine && go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12)
```

The exact tool versions, upstream commits, Go toolchain, and Action commits are recorded in `.github/supply-chain-lock.json`.

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
make workflow-check # validate GitHub Actions syntax and expressions
make supply-chain-check # validate immutable CI/toolchain/fixture locks and reachable vulnerabilities
make docs-check    # scan public docs for retired stale wording
make shell-check   # run shell script syntax checks
make python-check  # run Python script syntax checks
make config-check  # validate repository config, rule, and suppression files
make bench         # C parser/UDS microbenchmarks + Go benchmarks
make fuzz-parser   # deterministic ASan fuzz smoke for the C frame parser
make fuzz-parser-long # longer deterministic ASan fuzz pass for the C frame parser
make fuzz-sustained # sustained ASan parser fuzz evidence
make e2e-smoke     # deterministic pcap -> SQLite -> API smoke test
make e2e-pressure  # repeat-pcap end-to-end throughput smoke test
make e2e-corpus-pressure # local sanitized pcap corpus pressure evidence
make sanitize-pcap # sanitize an Ethernet pcap before sharing it
make dist          # build a local release archive under dist/
make docker-build  # build a local Docker image
make rc-check      # release-candidate verification bundle
make release-gate  # reviewed external fuzz/pcap evidence gate
make lint          # go vet + optional staticcheck
make quickstart    # build, generate pcap, run engine/capture, print alerts
make asan-test     # C parser tests under AddressSanitizer
make clean
```

Local Docker image builds are available through `make docker-build` and are covered by `make rc-check`. GitHub Actions workflows are present for release-candidate checks, tag-driven GitHub Release publication, and GHCR image publishing; both publication workflows rerun `make rc-check` and `make release-gate` before publishing named assets.

CI runs `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check` before `make rc-check`. The supply-chain gate verifies full Action SHAs, the exact Go patch toolchain, pinned tool install commands, `actionlint`, `govulncheck`, and all nine external fixture/license hashes without retaining fixture bytes in Git or the runner workspace.

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
engine:
  uds_socket_path: "/tmp/netsentry.sock"
  uds_socket_mode: "0600"
  channel_buffer_size: 10000
  worker_count: 1
  db_dir: "data"
  db_path: "data/netsentry.db"
  db_journal_mode: "WAL"
  db_busy_timeout: 5000
  alert_recovery_log_path: ""
  rules_seed_file: "configs/rules.json"
  suppressions_file: "configs/suppressions.json"
  api_listen_host: "127.0.0.1"
  api_port: 8080
  api_auth_enabled: false
  api_auth_token: "${NETSENTRY_API_TOKEN:}"
  health_freshness_limit_seconds: 30
  pprof_enabled: false
  pprof_addr: "127.0.0.1:6060"

logging:
  format: "json"
```

Environment expansion supports `${ENV_VAR}` and `${ENV_VAR:default}`. Missing variables expand to their configured default. The loader rejects unknown top-level and nested YAML fields, so configuration typos and retired reserved fields fail at startup instead of silently retaining defaults. Every accepted YAML field configures the Go engine; the standalone C capture binary is configured explicitly with `-r` or `-i`, `-s`, and `-c` command-line arguments. `engine.uds_socket_mode` must be a non-zero octal permission mode no greater than `0777`; `logging.format` is `json` or `console`. Validation also rejects invalid API ports, worker/channel ranges, empty tokens when authentication is enabled, and any non-loopback API listener without authentication.

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
  e2e_corpus_pressure.sh
  fuzz_sustained.sh
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

The loader still accepts legacy top-level arrays and legacy `payload_match`, `ip_blacklist`, and MITRE scalar fields while old files are migrated. Reload rejects null or duplicate rules, empty enabled match sets, unsupported types/severities, and non-canonical MITRE tuples. The v0.1 alert schema permits at most one MITRE technique per rule.

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
# Optional longer post-capture drain wait for larger local runs:
PRESSURE_REPEATS=10000 PRESSURE_WAIT_ATTEMPTS=1200 make e2e-pressure
```

For release-candidate evidence against local sanitized traffic samples, run:

```bash
PCAP_CORPUS=/path/to/sanitized-pcaps make e2e-corpus-pressure
# Optional output directory:
PCAP_CORPUS=/path/to/sanitized-pcaps CORPUS_OUTPUT_DIR=/tmp/netsentry-corpus-evidence make e2e-corpus-pressure
```

`PCAP_CORPUS` may point to a single `.pcap`/`.pcapng` file or a directory. The
script starts the engine once, runs the capture binary over each file, waits for
the pipeline to drain, then writes JSON and Markdown evidence. The default output
directory is `docs/evidence/local/`, which is ignored because corpus paths and
operator notes can be sensitive. Corpus paths are redacted by default; set
`NETSENTRY_EVIDENCE_INCLUDE_PATHS=1` only for private local debugging evidence.
The summaries include packet/alert counts and rates, alert match rate, sampled
peak engine RSS, engine error-log line count, API health, metrics, and an alerts
query snapshot. Sanitize pcaps before sharing them.

For C parser hardening work, run the deterministic ASan fuzz smoke:

```bash
make fuzz-parser
# Longer local pass:
FUZZ_LONG_ITERATIONS=1000000 make fuzz-parser-long
# Evidence-producing sustained run:
make fuzz-sustained
# Optional external corpus replay:
FUZZ_CORPUS=/path/to/local-corpus make fuzz-sustained
```

The harness starts from built-in Ethernet/IP/TCP/UDP, VLAN, Q-in-Q, fragment, short-frame, and malformed TCP-offset seeds, then applies deterministic mutations.
`make fuzz-sustained` records JSON and Markdown evidence under
`docs/evidence/local/` by default. That directory is ignored because external
corpus paths may be sensitive. Corpus paths are redacted by default; set
`NETSENTRY_EVIDENCE_INCLUDE_PATHS=1` only for private local debugging evidence.
Use `FUZZ_SUSTAINED_ITERATIONS` and `FUZZ_OUTPUT_DIR` to tune duration and output
location.

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

The `e2e-smoke` step uses a temporary config, Unix socket, API port, and SQLite database, then asserts that the synthetic pcap produces 6 processed packets, 5 alerts, 8 loaded rules, capture heartbeat metrics, process-lifetime packet/alert rate metrics, and rule/write latency histogram observations.

To create a local release archive:

```bash
make dist
VERSION=0.1.0-rc1 make dist
make release-artifacts VERSION=0.1.0
```

The archive and SHA-256 checksum are written to `dist/`. Generated release archives are ignored by Git. The archive includes generated `RELEASE_NOTES.md` with package contents, quick verification, v0.1.0 boundaries, release-candidate evidence, and links to packaged docs.
`make release-artifacts` is the stricter release helper: it requires a SemVer
`VERSION` without the leading `v`, then delegates to `make dist`.

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

To create a standard repository-provided synthetic corpus without any
third-party packet library:

```bash
make gen-sanitized-corpus
make gen-sanitized-corpus CORPUS_DIR=/tmp/netsentry-sanitized-corpus
```

The generator writes `payload-rules`, `protocol-mix`, and `background-traffic`
pcap/pcapng pairs per requested set, plus `MANIFEST.json`. For example,
`make gen-sanitized-corpus CORPUS_DIR=/tmp/netsentry-synthetic-100
CORPUS_SETS=100` emits 600 differentiated files. Each set contains only fixed
RFC 5737 documentation addresses, fixed local MAC addresses, deterministic
timestamps, synthetic payloads, and a unique synthetic marker. Repeated runs
are byte-identical. Keep the output outside the repository unless a reviewed
public evidence package explicitly requires it; synthetic output never replaces
external fuzz or realistic production traffic evidence.

The sanitizer preserves pcap timestamps, packet framing, Ethernet/VLAN/IPv4/TCP/UDP structure, ports, and lengths. It replaces MAC addresses, maps IPv4 addresses into the `198.18.0.0/15` benchmark range, overwrites TCP/UDP payload bytes, and zeroes unsupported captured frames.

Current validation baseline:

- `make test-unit` runs C/Go unit and race tests followed serially by C ASan tests.
- `make test-integration` verifies the pinned PcapPlusPlus/Zeek fixture manifest, processes supported external pcaps, and checks invalid CLI/non-Ethernet rejection.
- `make test-e2e` covers pcap -> UDS -> worker pool -> SQLite -> API; `make test-stress` runs configurable repeat-pcap pressure.
- Go tests cover receiver frame validation/lifecycle, worker-pool shutdown, panic isolation, rule/MITRE semantics, API limits, SQLite aggregation, daily shards, recovery-log replay, and storage degraded/emergency behavior.
- Release-candidate checks run syntax checks, repository configuration validation, dependency verification, C/Go tests, coverage snapshot, deterministic C parser fuzz smoke, e2e smoke, release archive checks, Docker image content smoke, and Docker runtime health smoke.

The C-side JSON line formatter is intentionally kept as a bounded handwritten v0.1.0 implementation. It avoids a new C dependency, rejects truncation, escapes JSON strings, Base64-encodes packet payload previews, and is covered by the UDS sender tests and current smoke checks. A cJSON migration should be reopened only with a concrete defect or fuzzing result.

Release readiness for v0.1.0:

The canonical release gate checklist and evidence handling notes are maintained
in `docs/release-readiness.md`.

Ready:

- `make rc-check` includes syntax checks, config validation, dependency verification, tests, coverage, deterministic fuzz smoke, e2e smoke, release archive checks, Docker image content smoke, and Docker runtime health smoke.
- GitHub Actions CI, tag-driven GitHub Release publication, and GHCR publishing workflows are checked in.
- The v0.1.0 release gate has a reviewed, version-scoped exception for real production-derived pcap evidence; it expires before v0.1.1. The separate R90-04 exception permits only anonymized public real-traffic PCAP evidence after approved privacy, provenance, sanitization, and sensitive-metadata reviews.
- The R90-04 exception is now recorded as expired at its completion commit and
  is rejected by `make release-gate`; it cannot authorize R90-05 or any tag or
  image publication.
- `make dist` produces a local release archive, checksum, and generated release notes.
- `make release-artifacts VERSION=0.1.0` validates release-version format before building publishable archive assets.
- `make docker-build` builds the local runtime image.
- Latest local full sudo Docker RC validation passed on 2026-07-08, covering the complete `make rc-check` bundle including Docker build, image content smoke, and runtime `/api/health` smoke.
- Latest local non-Docker RC validation passed on 2026-07-10 with `SKIP_DOCKER=1 make rc-check`, covering syntax, docs, Python, config, dependencies, C/Go tests, race tests, coverage 74.2%, ASan fuzz smoke, e2e smoke, dist archive smoke, and release notes smoke.

Release result:

- The signed `v0.1.0` tag, GitHub Release assets, tag-triggered Docker workflow, and public `ghcr.io/decline-llc/netsentry:v0.1.0` manifest were verified on 2026-07-11. The v0.1.0 exception does not carry into v0.1.1; R90-04 alone may use the separately approved anonymized public real-traffic alternative.
- v0.1.1 production-derived PCAP evidence uses `make pcap-evidence` and
  `make pcap-evidence-check`. The generator records path-redacted inventory
  facts; named reviewers must supply provenance, privacy, sanitization,
  sensitive-metadata, and final approval decisions. The release gate reparses
  and hashes the exact local corpus with `RELEASE_EXCEPTION=none`.
- The separately approved R90-05 exception accepts only the exact synthetic
  corpus digest and packet count in `docs/audit/release_exception_r9005.yaml`.
  The evidence must remain labeled synthetic and non-production-derived, and
  the exception expires before R90-06.

Exception record:

- `docs/audit/release_exception_v0.1.0.yaml` records the explicit v0.1.0
  exception. `docs/audit/release_exception_r9004.yaml` separately permits
  R90-04-only anonymized public real-traffic evidence after its required
  reviews. Synthetic/generated traffic is prohibited under that exception; the
  separate R90-05 exception is digest-scoped and later increments still require
  production-derived evidence.

Use `docs/evidence/release-evidence-template.md` for the sanitized public
release evidence record. Keep generated local evidence under
`docs/evidence/local/` out of Git.

Remaining test gaps:

- Sustained external C fuzz campaign results from larger parser and formatter corpora.
- Realistic pcap corpora for throughput and query tuning beyond repeat-pcap smoke runs.
- Broader SQLite corruption/fault-injection scenarios beyond the current disk-full, read-only, I/O, recovery replay, and emergency-mode tests.
- Active-load full-engine shutdown drills that combine receiver, worker, HTTP, and storage teardown.

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
