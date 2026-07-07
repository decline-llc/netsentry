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

For a non-interactive release smoke check, run:

```bash
make e2e-smoke
```

For a local coverage snapshot, run:

```bash
make test-coverage
```

For a local end-to-end throughput smoke check, run:

```bash
make e2e-pressure
# Optional larger run:
PRESSURE_REPEATS=10000 make e2e-pressure
```

For a C parser fuzz smoke:

```bash
make fuzz-parser
# Longer local pass:
FUZZ_LONG_ITERATIONS=1000000 make fuzz-parser-long
```

To create a local binary release archive:

```bash
make dist
```

For a local release-candidate verification bundle:

```bash
make rc-check
# If your Docker daemon requires elevated privileges:
DOCKER="sudo docker" make rc-check
```

The release-candidate bundle includes shell and Python syntax checks,
configuration sample validation, documentation consistency checks, dependency
verification, tests, a coverage snapshot, fuzz smoke, e2e smoke, release archive
checks, Docker image smoke, and Docker runtime health smoke.

To build the local Docker image:

```bash
make docker-build
# If your Docker daemon requires elevated privileges:
DOCKER="sudo docker" make docker-build
```

The Makefile writes Go build cache data to `/tmp/netsentry-go-cache` by default
so build, test, lint, and benchmark targets work in environments where the home
directory cache is read-only. Set `GOCACHE=/path/to/cache` to override it.

The repository also includes GitHub Actions workflows for release-candidate checks
and GHCR image publishing. Docker publishing runs the same `make rc-check`
bundle first, then only pushes on version tags or an explicit manual workflow run.

### v0.1.0 Release Readiness

Ready gates:

- Local source build, tests, coverage snapshot, deterministic fuzz smoke, e2e smoke, release archive checks, Docker image content smoke, and Docker runtime health smoke are wired into `make rc-check`.
- GitHub Actions CI reuses the release-candidate bundle.
- GHCR publishing workflow is present for version tags or explicit manual publishing.
- Release archive generation includes binaries, configs, docs, checksum, and generated release notes.

Remaining blockers before tagging v0.1.0:

- Run and record a final full `DOCKER="sudo docker" make rc-check` on the release candidate.
- Record sustained external C fuzz evidence against larger parser and formatter corpora.
- Record realistic pcap corpus pressure/query evidence, separate from synthetic repeat-pcap smoke runs.
- Create the named GitHub Release and publish the named registry image from a version tag.

---

## Implemented Today

- C capture skeleton with offline pcap reading, Ethernet/VLAN/IPv4/TCP/UDP parsing, Base64 payload preview, JSON line frames, hello and heartbeat frames.
- Go rule engine using `atomic.Pointer[ruleState]` immutable snapshots.
- Rule types: `payload_match`, `ip_blacklist`, `port_blacklist`.
- A self-contained Aho-Corasick matcher.
- Minimal Go UDS receiver, CIDR alert suppressor component, and SQLite alert store with UPSERT aggregation, JSONL recovery-log replay, startup TTL pruning, optional daily DB shard pathing/cleanup, cross-shard alert querying/counting in daily-shard mode, and degraded/emergency health tracking after storage errors.
- Minimal HTTP endpoints: `/api/health` with verbose component snapshot including storage status and available bytes, paginated `/api/alerts` with exact-match, time range, MITRE, matched-keyword, and aggregate-count filters, `/api/metrics`, rule listing, rule create/update/delete, rule reload, file-backed suppression create/update/delete/reload, method-aware error envelopes, optional PSK Bearer auth for modifying endpoints, non-GET audit logs, optional localhost-only pprof, storage health gauges, and payload preview redaction before alert writes.
- Seed rules in canonical wrapped JSON schema, with legacy schema compatibility retained in the loader.
- C-side JSON line formatting remains a bounded handwritten formatter for v0.1.0. It is covered by escaping, truncation, Base64 payload, UDS sender, microbenchmark, fuzz-smoke, and e2e heartbeat checks; a cJSON migration is not required unless later fuzzing exposes a concrete parser or formatting defect.

---

## Not Implemented Yet

These are v0.1.0 goals, not current behavior:

- Automatic disk cleanup or restart-free recovery after disk-full emergency mode.
- Full Prometheus metric coverage beyond the current process, current/high-water queue, rule/write latency, alert, storage, worker, and capture heartbeat metrics.
- Remaining large-corpus query tuning.
- Sustained external C fuzz campaigns against larger parser and formatter corpora.
- Published registry image for a named release.

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
curl "http://localhost:8080/api/alerts?severity=high&page=1&per_page=20" | python3 -m json.tool
```

Current alert responses use the stable list envelope:

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 5
  }
}
```

Supported filters are documented in `docs/api-reference.md`.

Performance microbenchmark scope, local baseline, and the repeat-pcap end-to-end pressure smoke are documented in `docs/performance.md`.

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
configs/    Runtime config, seed rules, and seed suppressions
docs/       Architecture, API, and development notes
scripts/    Quickstart pcap generator, e2e checks, release packaging, pcap sanitizer
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
