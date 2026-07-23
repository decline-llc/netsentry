# Changelog

All notable changes to NetSentry are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
NetSentry uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- C capture binary for offline pcap analysis with Ethernet, VLAN, Q-in-Q, IPv4, TCP, and UDP parsing.
- JSON line protocol over Unix Domain Socket with hello, packet, and heartbeat frames.
- Go receiver and single-worker packet pipeline with context-aware shutdown and packet-level panic isolation.
- Lock-free rule engine using immutable `atomic.Pointer[ruleState]` snapshots.
- Pure Go Aho-Corasick matcher for `payload_match` rules.
- Rule types for `payload_match`, `ip_blacklist`, and `port_blacklist`, including protocol, port, direction, offset, depth, case-insensitive matching, exact IP, and CIDR support.
- SQLite-backed alert store with UPSERT aggregation, JSONL recovery-log replay, startup TTL pruning, optional daily shard pathing, cross-shard alert querying/counting in daily-shard mode, and old daily shard file cleanup.
- Payload preview redaction before alert writes when `engine.redact_sensitive_fields` is enabled.
- REST API for health, alerts, metrics, rules CRUD/reload, and file-backed suppressions.
- Alert queries support SQLite-backed exact-match filters, RFC3339 time ranges, MITRE tactic/technique filters, matched-keyword substring filtering, minimum aggregate-count filtering, pagination, daily-shard cross-file querying/counting, and indexes for common exact/range filters.
- Unified API error envelope, pagination envelope, request IDs, method-aware 405 responses, optional PSK Bearer auth, non-GET audit logs, and localhost-only pprof.
- Prometheus text metrics for current packet, process-lifetime packet/alert rates, alert, current/high-water queue depth, rule latency, alert write latency, storage, worker, and capture heartbeat counters, with HELP text for exported gauges.
- Alert storage health tracking after SQLite write/query errors, including sticky emergency mode for disk-full, quota, read-only filesystem, and disk I/O failures, surfaced through verbose health with storage available bytes and `netsentry_storage_healthy`.
- Suppression rules can load from `engine.suppressions_file`; suppression create, update, delete, and reload operations persist or reload that file before updating the active filter.
- Deterministic synthetic pcap generator with a Python stdlib fallback when Scapy is unavailable.
- Non-interactive end-to-end smoke test via `make e2e-smoke`, including capture heartbeat and process rate metrics assertions.
- Repeat-pcap end-to-end throughput smoke test via `make e2e-pressure`.
- Local sanitized pcap corpus pressure evidence via `make e2e-corpus-pressure`, with ignored JSON/Markdown evidence output.
- Pcap sanitization helper via `make sanitize-pcap INPUT=... OUTPUT=...`.
- Deterministic AddressSanitizer fuzz smoke for the C frame parser via `make fuzz-parser`.
- Broader deterministic C parser fuzz seeds cover TCP, UDP, VLAN, Q-in-Q, IPv4 fragments, short frames, and malformed TCP data offsets; `make fuzz-parser-long` runs a longer local ASan pass.
- Sustained ASan C parser fuzz evidence entrypoint via `make fuzz-sustained`, with optional local corpus replay and ignored JSON/Markdown evidence output.
- Receiver lifecycle tests for multiple active UDS connections during context cancellation, with goleak coverage for the receiver package.
- Configurable concurrent UDS receiver connection cap with excess-client rejection and capacity-reuse coverage.
- SQLite aggregation tests now cover recovery-log replay idempotency, query index creation, SQL-backed filtering/pagination, out-of-order alert writes, rule/source/destination/port aggregation key separation, canceled write contexts, emergency mode restart replay, and unsupported journal mode validation.
- API tests cover health and metrics alert counts backed by a real daily-shard SQLite store.
- Full C capture AddressSanitizer build target via `make build-asan`.
- Local release archive packaging via `make dist`, including SHA-256 checksum generation.
- Generated release archive notes document package contents, quick verification, v0.1.0 boundaries, release-candidate evidence, and packaged documentation references.
- Local Docker image build via `make docker-build`.
- Release-candidate verification bundle via `make rc-check`, including documentation consistency, dependency verification, coverage snapshot, fuzz smoke, release archive, Docker image content, and runtime health smoke checks.
- Latest local full sudo Docker release-candidate validation passed on 2026-07-08.
- GitHub Actions workflows for release-candidate checks and GHCR Docker image publishing.
- Native coverage snapshot target via `make test-coverage`, running C tests and a Go coverage summary without external tooling.
- Native dependency integrity check via `make deps-check`, using `go mod verify`.
- Native documentation consistency check via `make docs-check`, scanning public docs for retired stale wording.
- Native shell syntax check via `make shell-check`, reused by release-candidate checks.
- Native Python syntax check via `make python-check`, reused by release-candidate checks.
- Native repository configuration check via `make config-check`, validating checked-in config, rules, and suppressions through current Go parsers.
- Public v0.1.0 release readiness checklist separating wired release gates from remaining blockers.
- Public release-readiness evidence index documenting v0.1.0 gates, evidence commands, local-only evidence handling, and final tag/publish checklist.
- Tag-driven GitHub Release publication workflow that reruns `make rc-check` and uploads the `make dist` tarball plus checksum.

### Changed
- Public rule samples now use the canonical wrapped schema while the loader remains backward compatible with legacy rule files.
- Public documentation now separates implemented behavior from planned v0.1.0 goals.
- Release documentation now distinguishes checked-in publication workflows from the final act of pushing a passing release tag.
- Quickstart clears the demo SQLite database before running so repeated runs keep returning the deterministic seed-alert set.
- Makefile Go targets now default `GOCACHE` to `/tmp/netsentry-go-cache`, while still allowing `GOCACHE=...` overrides, so build, test, lint, and benchmark targets work when the home-directory Go cache is read-only.
- Development and architecture testing notes now separate the current validation baseline from remaining test gaps.
- C-side JSON line formatting is documented as a bounded handwritten v0.1.0 implementation, so cJSON migration is no longer listed as required release-candidate work.
- Repeat-pcap pressure smoke can now tune the post-capture drain wait with `PRESSURE_WAIT_ATTEMPTS`, making larger local validation runs less prone to false failures while the worker and SQLite aggregation catch up.
- Engine shutdown now stops and waits for the UDS receiver and waits for the pipeline worker before returning from `main`, so deferred alert store close happens after worker teardown.

### Fixed
- C parser and UDS sender edge cases are covered by unit tests, ASan tests, and microbenchmarks.
- UDS reconnect behavior is tested across listener restart and receiver reconnection paths.
- Receiver shutdown closes single and multiple active Unix socket connections and removes the socket path.
- UDS receiver connections now have a validated finite read-idle timeout that
  refreshes after every complete frame and releases handler capacity on expiry
  without inflating decode-error metrics.
- UDS connections now require exactly one hello before packet or heartbeat
  traffic; reconnecting C capture sessions resend hello first, and session
  violations close only the offending connection.
- Engine worker shutdown orchestration is covered by a focused `cmd/netsentry` test.
- Engine startup now reports HTTP bind failures synchronously, and shutdown
  waits for the receiver, pipeline workers, and HTTP API before SQLite closes;
  an active-load integration regression covers the complete ordering.
- Worker panic recovery no longer terminates the worker loop after a single bad packet.
- Alert aggregation preserves earliest `first_seen`, latest `last_seen`, and latest payload/match fields when older events arrive after newer events in the same aggregation window.
- Daily-shard alert storage writes cross-day alerts to `netsentry-YYYY-MM-DD.db` files based on each alert timestamp during a running process.
- Existing non-current daily shards now receive the same read-only integrity
  preflight before a write; corrupt or truncated shards are rejected without
  modifying their bytes.
- Cross-shard query and count operations now open non-current daily shards with
  URL-safe SQLite read-only handles, including active WAL-backed shards.
- Disk-full/read-only/I/O storage failures now enter sticky emergency mode, stop retrying SQLite writes in the current process after recovery logging when possible, and replay pending recovery-log alerts after operator cleanup and restart.
- Existing non-empty SQLite databases now receive a read-only startup integrity
  preflight; corrupt or truncated files fail with a stable error and remain
  byte-for-byte unchanged for operator-led recovery.
- Existing primary databases and daily-shard write targets now also require the
  NetSentry table, column, and aggregation-uniqueness schema contract during
  that read-only preflight; unrelated or incompatible databases are rejected
  without modifying their bytes.
- Existing primary databases and daily-shard write targets reject extra unique
  indexes that do not contain a canonical alert or event identity before
  writable initialization; ordinary indexes and redundant identity-containing
  uniqueness extensions remain compatible.
- Malformed or unterminated JSONL recovery logs now fail startup with a stable
  integrity error before partial replay; rejected logs remain byte-for-byte
  unchanged and valid logs truncate only after successful persistence.
- Syntactically valid recovery records now must contain the durable normalized
  alert identity, timestamps/window/count, and network tuple before replay;
  semantic failures preserve the complete log without persisting a valid prefix.
- Recovery logs now receive their complete structural and semantic preflight
  before writable SQLite initialization, so rejected input cannot create a
  missing database or modify a compatible existing database.

### Known Gaps
- Automatic disk cleanup or restart-free recovery after storage emergency mode is not implemented.
- End-to-end pressure coverage currently includes repeat-pcap runs up to 60,000 packets locally; realistic pcap corpora can now be recorded with `make e2e-corpus-pressure` but release evidence is still pending.
- Sustained external C fuzz evidence is still pending; `make fuzz-sustained` now provides the local evidence entrypoint.

---

## [0.1.0] - Planned

### Target deliverables
- Offline pcap to alert workflow through C capture, UDS transport, Go receiver, rule engine, SQLite storage, and REST API.
- Honest documentation of v0.1.0 protocol and detection boundaries.
- Local binary release archive and Docker image.
- Repeatable release-candidate checks for tests, smoke tests, archive checksum/content, image content, and Docker runtime health.
- Unit and integration test coverage for implemented components.
