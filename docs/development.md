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
make fuzz-uds-formatter # deterministic ASan fuzz smoke for C UDS JSON formatters
make fuzz-sustained # sustained ASan parser and formatter fuzz evidence
make e2e-smoke     # deterministic pcap -> SQLite -> API smoke test
make e2e-pressure  # repeat-pcap end-to-end throughput smoke test
make e2e-corpus-pressure # local sanitized pcap corpus pressure evidence
make sanitize-pcap # sanitize an Ethernet pcap before sharing it
make dist          # build a local release archive under dist/
make docker-build  # build a local Docker image
make rc-check      # release-candidate verification bundle
make release-gate  # reviewed non-PCAP release evidence gate
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

The capture binary accepts `-c <connect_retries>` to bound initial UDS connection attempts. Offline mode defaults to 5 attempts so a missing engine fails clearly instead of retrying forever; live mode keeps retrying unless `-c` is set. Every initial or replacement UDS connection sends hello before packet or heartbeat traffic. The receiver closes a connection after packet-before-hello, heartbeat-before-hello, duplicate hello, or a heartbeat session mismatch without closing other valid clients.

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
  uds_max_connections: 4
  uds_read_timeout_seconds: 30
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

Environment expansion supports `${ENV_VAR}` and `${ENV_VAR:default}`. Missing variables expand to their configured default. The loader rejects unknown top-level and nested YAML fields, so configuration typos and retired reserved fields fail at startup instead of silently retaining defaults. Every accepted YAML field configures the Go engine; the standalone C capture binary is configured explicitly with `-r` or `-i`, `-s`, and `-c` command-line arguments. `engine.uds_socket_mode` must be a non-zero octal permission mode no greater than `0777`, `engine.uds_max_connections` must be between 1 and 1024, and `engine.uds_read_timeout_seconds` must be between 1 and 3600. The receiver closes newly accepted excess connections while all handler slots are occupied, expires clients that deliver no complete frame within the configured interval, refreshes the deadline after each complete frame, and makes the slot available again after disconnect or expiry. Idle expiry is not counted as a frame decode error. `logging.format` is `json` or `console`. Validation also rejects invalid API ports, worker/channel ranges, empty tokens when authentication is enabled, and any non-loopback API listener without authentication.

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

The Go portion of `make bench` excludes ordinary tests and discovers both the
rule-matching and primary SQLite alert-store benchmark families. The store
cases retain production recovery durability, bound write rows between timed
operations, and seed their fixed query corpus through `WriteBatch`; they do
not publish a portable or production throughput threshold. Use `make test`
separately for the complete correctness and race gate.

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

Rule CRUD and explicit reload each build and atomically publish a complete
engine snapshot, but their full file-backed transactions are not currently
serialized with each other. Two handlers can derive candidates from the same
old snapshot and a later full-file replacement can lose an earlier successful
mutation. R90-77 is the direct concurrency boundary; R90-78 separately covers
short-write, sync, close, rename, and parent-directory durability outcomes.

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

The suppression manager already serializes mutations, but its temporary-file
replacement has no direct short-write, file-sync, rename, or parent-directory
sync fault evidence. R90-79 keeps that durability work separate from the rule
transaction fix so each active-memory and on-disk outcome remains reviewable.

### SQLite startup integrity and recovery

An existing non-empty primary alerts database receives a read-only SQLite
`PRAGMA quick_check` and required-schema inspection before journal-mode or
schema initialization. The schema check requires the current `alerts` and
`alert_events` columns, types, primary-key/non-null roles, and binary-collated
aggregation uniqueness contract. Required table, column, and index-key
identifiers are matched case-insensitively, as SQLite resolves them. The check
also rejects extra uniqueness constraints, generated columns, triggers,
`CHECK` constraints, and foreign-key relationships that attach to or reference
either write-critical table and can alter or reject NetSentry's fixed writes.
Corrupt, truncated,
unrelated, or incompatible input fails startup with `ErrDatabaseIntegrity` and
an explicit statement that the file was not modified. Missing query indexes
and extensions confined to unrelated operator tables are compatible. New paths
and existing empty files continue through normal initialization.

If the integrity preflight fails:

1. Keep NetSentry stopped; do not delete or retry writes against the database.
2. Preserve and copy the database plus matching `-wal`, `-shm`, and alert
   recovery-log sidecars while the process is stopped.
3. Run SQLite integrity/recovery tools against a copy, never the only original.
4. Keep the original for rollback or forensic review, and configure a new or
   operator-recovered database path only after the recovered copy is reviewed.

NetSentry does not automatically repair, replace, quarantine, or delete a
database that fails this check.

### Restart-free emergency recovery test contract

R90-60 implements the operator-triggered design from R90-57. Emergency mode is
sticky until restart or a successful authenticated `POST
/api/storage/recovery` request. The implementation has no timer, free-space
poller, implicit write retry, or automatic cleanup.

Tests synchronize on observable state/lifecycle boundaries without fixed
sleeps. The durable contract covers:

1. **Ownership:** exactly one request changes `emergency` to `recovering`; a
   concurrent request receives a conflict and does not wait or start later.
2. **Lifecycle exclusion:** active reads and writes drain before handle
   replacement; new database operations wait or cancel; health observation
   remains available; no operation reaches a closed or second writable handle.
3. **Preflight preservation:** malformed/truncated recovery input, incompatible
   schema, corrupt database, corrupt WAL, and inconsistent SHM each fail through
   separate read-only handles before writable open, with database, WAL, SHM,
   and recovery-log bytes compared before and after.
4. **Fault remains:** disk-full, quota, read-only filesystem, and I/O failures
   during writable replay return to `emergency`, retain the complete recovery
   log, record that SQLite sidecars may have changed after the writable boundary,
   and never schedule a retry.
5. **Successful replay:** a repaired store reopens with one writable owner,
   replays every pending event once, truncates the log only after the complete
   commit set, marks healthy, and permits the next normal write.
6. **Committed-prefix retry:** cancellation or an injected later-shard failure
   after an earlier commit retains the full log; a second explicit request uses
   event IDs to finish without duplicate events or aggregate inflation.
7. **Cancellation and shutdown:** cancellation before lifecycle readiness,
   during read-only preflight, and during active replay all release ownership;
   shutdown waits for release and never publishes a transient healthy state.
8. **Empty-log proof:** recovery still performs a write-capable SQLite probe
   before marking healthy: begin an immediate transaction, insert a reserved
   nonce-scoped `alert_events` event, verify it inside the transaction, and
   roll back. A read-only query or free-space increase alone is insufficient
   proof, and the probe must leave no durable application row.
9. **Daily shards and encoded paths:** direct and space-containing database
   paths, healthy active WAL, current and historical shards, and a failure in a
   later shard preserve serialization and idempotent retry behavior.
10. **Operator surface:** missing/invalid authentication, healthy-state calls,
    duplicate calls, timeout, preflight rejection, writable failure, success,
    and audit-log redaction have stable direct regressions.

Repeated Go reliability runs must use `-count=1`. Every failure test that
promises preservation must compare artifacts through independent read-only
handles before any writable open.

R90-62 exercises the committed-prefix boundary with real SQLite behavior. It
orders daily-shard replay paths deterministically, holds the later compatible
shard with an independent reserved writer, and observes the earlier committed
event through a separate read-only handle. Direct failure and active-context
cancellation cases both retain the complete log and sticky emergency state;
after the lock is released, an explicit retry proves one event and aggregate
count one per input.

R90-66 applies the same direct-evidence rule to ordinary primary writes. Both
cases open one read-only observer before the writer, reserve the primary
database with an independent `BEGIN IMMEDIATE`, and compare the exact canonical
recovery record after failure. The contention case waits for the real SQLite
busy result; the active-cancellation case waits for both the complete synced
log and an in-use store connection before cancelling. Neither case observes an
event or aggregate before retry. Transaction failures preserve
`context.Canceled` or the active deadline in the error chain alongside the
driver diagnostic; after the reservation is released, one normal retry drains
duplicate event records idempotently, leaves aggregate count one, clears the
log, and restores healthy state.

The primary startup check, non-current write preflight, and historical
query/count paths share one URL-safe SQLite `mode=ro` helper. The helper
resolves database symlinks before inspecting sidecars and constructing the URI,
so its WAL/SHM lookup matches the Unix VFS target location. If either WAL or
SHM already exists, NetSentry also requests `readonly_shm=1`; the default Unix
VFS therefore cannot update reader marks or create, truncate, or rebuild the
existing SHM in place. A database with no sidecars retains ordinary `mode=ro`
behavior because forcing read-only SHM in that state would reject a healthy
clean reopen. Direct fault coverage rejects an unsupported WAL version during
primary startup and an active-owner inconsistent SHM during encoded-path
historical query, count, and write preflight, with database, WAL, and SHM byte
comparisons.

The same preservation rule applies when a running daily-shard store receives
an alert for an existing non-current shard. NetSentry checks that shard through
a separate read-only handle before writable initialization; a corrupt or
truncated shard, or a structurally valid shard with an unrelated/incompatible
schema, rejects the write with `ErrDatabaseIntegrity` and is left byte-for-byte
unchanged. Preserve the shard and any sidecars, inspect only a copy, and restore
or replace it through an operator-controlled recovery path.

Historical-shard query and count paths are also non-mutating at the connection
boundary: each non-current shard uses that shared read-only helper, while the
current shard continues to use the store-owned connection. Healthy detached
file sets, shards owned by an active WAL writer, and an active database reached
through a symlink remain queryable without changing database, WAL, or SHM
bytes.
Persisted alert rows with a destination port outside `0..65535` or an aggregate
count below one also return a field-specific read error. Preserve the database,
inspect a copy, and repair or replace it only through the operator-controlled
recovery process; NetSentry does not clamp or rewrite those values.
Stored severity must be exactly `low`, `medium`, `high`, or `critical`; empty,
case-variant, and unsupported text returns the same kind of field-specific read
error rather than being normalized or classified under another severity.
Stored aggregate timestamps must also use the exact canonical UTC RFC3339Nano
text emitted by the SQLite writer and satisfy
`window_start <= first_seen <= last_seen`. Parseable explicit or non-UTC
offsets and redundant fractional precision fail before ordering or identity
validation because SQLite compares these columns as text. Every aggregation,
ordering, inclusive time-filter, and retention inequality converts canonical
text to a shared fixed-width nanosecond key first. The writable primary
database creates an optional expression index for global `last_seen`
order/range scans; read-only legacy shards use the same comparison without
index creation or byte modification. The row-ordering check preserves
compatibility with historical aggregation-window durations while rejecting
reversed event ranges and window starts after the first event.
Stored alert IDs must equal the canonical identity derived from `rule_id`,
`src_ip`, `dst_ip`, `dst_port`, and `window_start`. Empty or altered IDs fail
row decoding; the reader does not repair the row.
Stored `event_id`, `rule_id`, `rule_name`, `protocol`, `src_ip`, and `dst_ip`
values must also contain non-whitespace text. Optional payload and match fields
remain valid when empty. Stored MITRE tactic, technique ID, and technique name
must be either exactly all empty or all nonblank. Partial and whitespace-only
tuple members fail shared row decoding; complete values are returned unchanged
without being checked against the current catalog.
Stored protocol text must also equal the canonical writer name: `TCP`, `UDP`,
`ICMP`, or `PROTO_<number>` for another uint8 IP protocol. Case variants,
arbitrary names, malformed or out-of-range numbers, leading-zero forms, and
numeric aliases of named protocols fail without normalizing or rewriting the
row.
Stored source and destination addresses must additionally parse as strict IPv4.
Malformed, ordinary IPv6, and IPv4-mapped IPv6 text returns a field-specific
read error before dependent aggregation-identity validation. Historical reads
retain the read-only shard boundary and do not repair or rewrite the row.

Recovery-log replay requires newline-terminated JSONL records. NetSentry reads
and validates the complete log before writing any recovered alert; malformed
JSON or a missing final newline fails startup with `ErrRecoveryLogIntegrity`,
leaves the log byte-for-byte unchanged, and does not persist a valid prefix.
Syntactically valid records reject repeated top-level JSON names before model
decoding, including exact duplicates and case-variant aliases that Go would map
to the same durable field. Every top-level name must exactly match the current
alert writer's JSON vocabulary; unknown scalar or nested members and
case-variant supported names are rejected. Object member names inside accepted
field values are not recursively inspected. At package initialization,
NetSentry derives one immutable recovery contract from the declared exported
`model.Alert` fields, their `json` tags, and supported writer types. Per-record
validation uses that contract for field order, canonical names,
required/optional status, JSON kinds, and integral encoding without reflection.
Ignored and unexported fields are excluded; embedded fields,
case-insensitive name conflicts, composite types, and custom marshalers fail
contract construction rather than silently weakening validation. Every field
the writer cannot omit under the module toolchain's supported `omitempty` and
`omitzero` behavior must be present before model decoding; `raw_payload`
remains the only current optional field. Every present value must use the
non-null top-level JSON kind emitted by the writer: strings for text and
timestamps, and numbers for `dst_port` plus `aggregated_count`. A present
`raw_payload` must be a string. Both numeric fields must use canonical unsigned
base-10 integer JSON spelling without an exponent, fractional component, sign,
or multi-digit leading zero; alternate representations fail before model
decoding.
Each decoded record must contain the durable identity (`id`, `event_id`, and
`rule_id`), a nonblank `rule_name`, timestamps/window, aggregate count, and
network tuple (`src_ip`, `dst_ip`, and `protocol`) emitted by the normalized
recovery writer. The `event_id` must equal the deterministic identity used by
the event ledger, the durable `id` must equal the normalized aggregation
identity, `first_seen` and `last_seen` must equal `timestamp`, `window_start`
must match the configured aggregation window, `aggregated_count` must equal
one, severity must be exactly `low`, `medium`, `high`, or `critical`, and both
addresses must be strict IPv4. Each of the four timestamp strings must equal
the canonical UTC RFC3339Nano value emitted by the writer; parseable explicit
or non-UTC offsets and redundant fractional precision are rejected before
representation-dependent identity checks. Missing, empty, and whitespace-only
rule names are rejected without trimming nonblank names. Empty, case-variant,
and unsupported severities are rejected rather than normalized. MITRE tactic,
technique ID, and technique name must be either exactly all empty or all
nonblank; complete tuple text is preserved without trimming or current-catalog
validation. Protocol text must equal the same canonical writer name enforced
for stored rows: `TCP`, `UDP`, `ICMP`, or `PROTO_<number>` for another uint8 IP
protocol. A missing, empty, altered, inconsistent, alternate timestamp
encoding, partial or whitespace-only MITRE tuple, noncanonical protocol,
malformed, ordinary IPv6, or IPv4-mapped IPv6 identity/address field reports
its record number and fails through the same preservation boundary before
replay starts.
This complete preflight occurs before database-directory creation and writable
SQLite initialization. NetSentry replays the validated in-memory record set
after initialization without a second read; invalid input therefore leaves a
missing target database absent and a compatible existing database unchanged.
Preserve the rejected log, inspect only a copy, and repair or replace it only
through an operator-controlled recovery path. A valid log is truncated only
after every recovered alert is successfully persisted.
Normal runtime writes first validate any existing recovery log, then validate
the complete newly normalized batch against the same durable contract before
appending its first record. A later invalid current record therefore cannot
partially append an earlier valid record, modify SQLite, or mark healthy
storage degraded; an existing valid pending log remains byte-for-byte
unchanged.

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

For C parser and UDS JSON formatter hardening, run the deterministic ASan fuzz
smokes:

```bash
make fuzz-parser
make fuzz-uds-formatter
# Longer local pass:
FUZZ_LONG_ITERATIONS=1000000 make fuzz-parser-long
# Evidence-producing sustained run (both harnesses):
make fuzz-sustained
# Optional external corpus replay for the byte-oriented parser only:
FUZZ_CORPUS=/path/to/local-corpus make fuzz-sustained
```

The parser harness starts from built-in Ethernet/IP/TCP/UDP, VLAN, Q-in-Q,
fragment, short-frame, and malformed TCP-offset seeds, then applies
deterministic mutations. The formatter harness derives bounded structured
packet, heartbeat, and hello inputs; it covers escaping, payload and integer
edges, proves exact-fit success plus undersized-buffer rejection under ASan,
and independently decodes representative JSONL frames with Python's standard
library. `FUZZ_FORMATTER_ITERATIONS` controls its deterministic mutation count.
`make fuzz-sustained` forcibly rebuilds both ASan harnesses, runs them serially
at the same `FUZZ_SUSTAINED_ITERATIONS` budget, and records independently
validated parser/formatter statuses, elapsed times, exact iteration reports,
and sanitizer finding counts. JSON and Markdown evidence defaults to
`docs/evidence/local/`; that directory is ignored because external corpus paths
may be sensitive. Optional corpus replay applies only to the byte-oriented
parser. Corpus paths are redacted by default; set
`NETSENTRY_EVIDENCE_INCLUDE_PATHS=1` only for private local debugging evidence.
Use `FUZZ_OUTPUT_DIR` to select another local output location. Generated
evidence is explicitly `local_synthetic` and cannot establish production
traffic, throughput, release, tag, or publication authority. The reviewed
R90-64 baseline summary is in
[`docs/evidence/r90-64-sustained-fuzz-20260803.md`](evidence/r90-64-sustained-fuzz-20260803.md).

The current benchmark scope, local baseline, and pressure smoke behavior are
documented in `docs/performance.md`. `make bench` executes the existing C
microbenchmarks and invokes
`go test -run '^$' -bench=. -benchtime=10s -benchmem ./...`. Disabling ordinary
tests in this dedicated target prevents long all-package benchmarks from
interfering with timing-sensitive tests; `make test` remains the separate
correctness/race gate. The Go module defines deterministic Aho-Corasick and full
rule-engine `Match` cases for no-hit and multi-hit payloads; setup and
correctness checks remain outside the timed loops and every case reports
allocations. R90-71 adds production-durable primary SQLite single/32-alert
write cases and indexed rule/time queries over a fixed fixture. The R90-72
audit finds no versioned, repeated matched-environment complete-surface sample
set, so R90-73/R90-74 build that local evidence before the separately blocked
R90-75 budget decision. These local synthetic measurements are not production
throughput claims.

For release-candidate checks, run:

```bash
make rc-check
DOCKER="sudo docker" make rc-check
```

This runs `make shell-check`, `make docs-check`, `make python-check`, `make config-check`, `make deps-check`, `make test`, `make test-coverage`, `make fuzz-parser`, `make fuzz-uds-formatter`, `make e2e-smoke`, `make dist`, release archive checksum/content smoke checks, `make docker-build`, a minimal Docker image content smoke check, and a Docker runtime `/api/health` smoke check. If Docker is unavailable in the current environment, use:

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
- Receiver contract tests enforce strict IPv4 packet addresses, including
  ordinary and IPv4-mapped IPv6 rejection with one decode error and no queued
  packet.
- Go tests cover receiver frame validation/lifecycle, connection caps and read-idle expiry, worker-pool shutdown, panic isolation, rule/MITRE semantics, API limits, SQLite aggregation including nanosecond timestamp aggregation/order/filter/pruning and query-plan coverage, daily shards, bounded recovery-log encoding/replay plus canonical-field-vocabulary, required-value-kind, duplicate-field, timestamp-encoding, event-identity, severity, rule-name, MITRE-tuple, protocol-name, and semantic validation, corrupt/truncated/write-blocking-schema startup and historical-shard preservation, cross-process corrupt WAL/SHM three-file preservation, non-binary aggregation, generated-column, `CHECK`-constraint, and write-critical foreign-key rejection, compatible case-variant required identifiers and ordinary column/index/unrelated-table extensions, collation-independent exact alert filters, persisted numeric/severity/timestamp-encoding/timestamp-order/aggregation-identity/required-text/MITRE-tuple/IPv4-address rejection, direct and symlinked active WAL-backed read-only access without SHM mutation, clean no-sidecar reopen compatibility, storage degraded/emergency behavior, ordinary primary contention and active cancellation after durable recovery append, explicit recovery ownership/cancellation including committed-prefix later-shard failure and active-replay cancellation, preservation-safe multi-shard preflight, empty-log write probing, retained-log idempotent retry, authenticated API status mapping, recovering health observation, and redacted audit phase metadata.
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
- Historical v0.1.1 production-derived PCAP evidence used `make pcap-evidence` and
  `make pcap-evidence-check`. The generator records path-redacted inventory
  facts; named reviewers must supply provenance, privacy, sanitization,
  sensitive-metadata, and final approval decisions for that optional workflow.
- The historical R90-05 exception accepted only the exact synthetic
  corpus digest and packet count in `docs/audit/release_exception_r9005.yaml`.
  The evidence must remain labeled synthetic and non-production-derived, and
  the exception expires before R90-06.
- As of 2026-07-16, `docs/audit/pcap_release_gate_waiver.yaml` removes every
  PCAP requirement from release-gate acceptance. PCAP generation, sanitization,
  manifest, and pressure commands remain optional diagnostics. Raw PCAP bytes
  and private paths still stay outside Git and the Vault.

Exception record:

- `docs/audit/release_exception_v0.1.0.yaml` records the explicit v0.1.0
  exception. `docs/audit/release_exception_r9004.yaml` separately permits
  R90-04-only anonymized public real-traffic evidence after its required
  reviews. Those historical exception boundaries are retained as delivery
  evidence but no longer affect the globally waived PCAP release gate.

Use `docs/evidence/release-evidence-template.md` for the sanitized public
release evidence record. Keep generated local evidence under
`docs/evidence/local/` out of Git.

Remaining test gaps:

- Sustained parser and formatter campaign results from larger reviewed external
  corpora. The R90-64 one-million-iteration result is the accepted local
  synthetic dual-harness baseline, not external corpus evidence.
- Additional diverse, alert-bearing realistic pcap corpora for throughput,
  query tuning, and alert-volume behavior. R90-04 already records one approved
  public real-traffic pressure run; new corpus work is an optional
  external-input diagnostic, not a release-gate prerequisite.
- R90-66 through R90-68 directly cover primary interruption, recovery-log
  append open/short-write/sync/close faults, and post-commit clear
  open/truncate/sync/close faults across primary and encoded daily-shard paths.
  Retained and already-cleared outcomes both have one explicit idempotent
  recovery proof. No later local storage-fault increment is currently queued;
  completed corruption, schema, sidecar, and replay boundaries are not reopened.
- R90-70 and R90-71 complete the local C/Go microbenchmark inventory, but the
  numeric Go results are not versioned and the historical C/pressure samples
  are not repeated matched-environment observations. R90-72 therefore rejects
  an immediate portable threshold: R90-73 will add tested evidence capture,
  R90-74 will record a repeated single-host observation baseline, and R90-75
  remains blocked on comparable-environment evidence plus explicit product/SLO
  budget scope.

The full-engine lifecycle regression now combines the real UDS receiver,
pipeline worker, HTTP API, and SQLite store under active load. It verifies that
shutdown waits for an in-flight match and closes receiver and API listeners
before the store-close boundary.

---

## 8. Development Roadmap

The versioned delivery authority is
[`docs/plans/rolling-90-day-roadmap.md`](plans/rolling-90-day-roadmap.md).
`$netsentry-next` audits that file against Git, task-state, remote, and Vault
evidence before selecting exactly one increment. Any ignored local master plan
is private working material only and cannot override the versioned roadmap.

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
