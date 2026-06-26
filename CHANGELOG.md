# Changelog

All notable changes to NetSentry are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
NetSentry uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- Aho-Corasick multi-pattern matcher (pure Go, no CGO, `engine/internal/rule/ahocorasick`)
- Lock-free rule engine using `atomic.Pointer[ruleState]`
- Packet model (`PacketInfo`, `Alert`, `Rule`) in `engine/pkg/model`
- Configuration loader with `${ENV_VAR:default}` expansion
- Pipeline interfaces (`Matcher`, `AlertWriter`)
- Graceful shutdown signal handler
- C capture module skeleton: Ethernet/VLAN/IP/TCP/UDP parser, UDS sender, passthrough parser
- 8 seed detection rules covering SQLi, XSS, path traversal, Log4Shell, reverse shell, C2 IP blacklist

---

## [0.1.0] — Planned 2026-09

### Target deliverables
- Full C → UDS → Go packet pipeline (offline pcap mode)
- Rule engine: `payload_match` + `ip_blacklist` + MITRE ATT&CK mapping
- Alert aggregation: SQLite WAL, 60-second deduplication window, per-day DB sharding
- REST API: `/api/alerts`, `/api/rules`, `/api/health`, `/api/metrics` (Prometheus)
- PSK authentication middleware
- Docker single-image deployment
- Complete test suite: unit + integration + fuzz
