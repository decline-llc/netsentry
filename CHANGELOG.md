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
- SQLite-backed alert store with UPSERT aggregation, startup TTL pruning, optional daily shard pathing, and old daily shard file cleanup.
- Payload preview redaction before alert writes when `engine.redact_sensitive_fields` is enabled.
- REST API for health, alerts, metrics, rules CRUD/reload, and file-backed suppressions.
- Alert queries support exact-match filters, RFC3339 time ranges, MITRE tactic/technique filters, matched-keyword substring filtering, and minimum aggregate-count filtering.
- Unified API error envelope, pagination envelope, request IDs, method-aware 405 responses, optional PSK Bearer auth, non-GET audit logs, and localhost-only pprof.
- Prometheus text metrics for current packet, alert, queue, rule latency, alert write latency, storage, worker, and capture heartbeat counters, with HELP text for exported gauges.
- Basic alert storage health tracking after SQLite write/query errors, surfaced through verbose health and `netsentry_storage_healthy`.
- Suppression rules can load from `engine.suppressions_file`; suppression create, update, delete, and reload operations persist or reload that file before updating the active filter.
- Deterministic synthetic pcap generator with a Python stdlib fallback when Scapy is unavailable.
- Non-interactive end-to-end smoke test via `make e2e-smoke`, including capture heartbeat metrics assertions.
- Repeat-pcap end-to-end throughput smoke test via `make e2e-pressure`.
- Pcap sanitization helper via `make sanitize-pcap INPUT=... OUTPUT=...`.
- Deterministic AddressSanitizer fuzz smoke for the C frame parser via `make fuzz-parser`.
- Receiver lifecycle tests for multiple active UDS connections during context cancellation, with goleak coverage for the receiver package.
- SQLite aggregation tests now cover out-of-order alert writes, rule/source/destination/port aggregation key separation, canceled write contexts, and unsupported journal mode validation.
- Full C capture AddressSanitizer build target via `make build-asan`.
- Local release archive packaging via `make dist`, including SHA-256 checksum generation.
- Local Docker image build via `make docker-build`.
- Release-candidate verification bundle via `make rc-check`, including fuzz smoke, release archive, Docker image content, and runtime health smoke checks.
- GitHub Actions workflows for release-candidate checks and GHCR Docker image publishing.

### Changed
- Public rule samples now use the canonical wrapped schema while the loader remains backward compatible with legacy rule files.
- Public documentation now separates implemented behavior from planned v0.1.0 goals.
- Quickstart clears the demo SQLite database before running so repeated runs keep returning the deterministic seed-alert set.

### Fixed
- C parser and UDS sender edge cases are covered by unit tests, ASan tests, and microbenchmarks.
- UDS reconnect behavior is tested across listener restart and receiver reconnection paths.
- Receiver shutdown closes single and multiple active Unix socket connections and removes the socket path.
- Worker panic recovery no longer terminates the worker loop after a single bad packet.
- Alert aggregation preserves earliest `first_seen`, latest `last_seen`, and latest payload/match fields when older events arrive after newer events in the same aggregation window.

### Known Gaps
- Runtime cross-day database rotation and cross-day alert querying are not implemented.
- WAL JSONL replay and automatic disk-full recovery are not implemented.
- End-to-end pressure coverage currently includes repeat-pcap runs up to 60,000 packets locally, but realistic pcap corpora are still pending.
- Longer C fuzz runs with broader seed corpora are still pending.

---

## [0.1.0] - Planned

### Target deliverables
- Offline pcap to alert workflow through C capture, UDS transport, Go receiver, rule engine, SQLite storage, and REST API.
- Honest documentation of v0.1.0 protocol and detection boundaries.
- Local binary release archive and Docker image.
- Repeatable release-candidate checks for tests, smoke tests, archive checksum/content, image content, and Docker runtime health.
- Unit and integration test coverage for implemented components.
