# NetSentry Performance Evidence

> Status: local benchmark inventory and historical development snapshot. These
> numbers are not an end-to-end production throughput claim or a portable
> regression threshold.

---

## Scope

The current benchmark target measures the parts that already have stable standalone harnesses:

- C Ethernet/VLAN/Q-in-Q IPv4 TCP parser microbenchmarks.
- C packet and heartbeat JSON formatting microbenchmarks.
- C Unix Domain Socket line write microbenchmark against a local drain listener.
- Go Aho-Corasick `Match` microbenchmarks for deterministic no-hit and
  multi-hit payloads.
- Go full rule-engine `Match` microbenchmarks for deterministic no-hit and
  multi-hit packets across payload, IP, and port rules.
- Go primary SQLite alert-store microbenchmarks for durable single and
  32-alert writes plus indexed exact-rule and timestamp-range queries over a
  fixed 512-row fixture.
- A repeat-pcap end-to-end pressure smoke test across C capture, UDS, Go receiver, rule matching, SQLite aggregation, and API health/alerts checks.
- An optional local corpus pressure evidence script for sanitized `.pcap` and `.pcapng` files supplied by the operator.

The repeat-pcap pressure smoke is intended to catch obvious pipeline regressions and provide a local baseline. It is not a production traffic model and does not replace a broader benchmark corpus.

---

## Reproduction

```bash
make bench
```

The root target runs each Go sub-benchmark for ten seconds and reports
allocations. Its Go command uses `-run '^$'`, so ordinary tests remain in the
separate `make test` correctness/race gate and do not run concurrently with
long benchmark packages. For quick discovery and bounded local execution of
only the rule-matching cases:

```bash
cd engine
go test -run '^$' -bench 'Benchmark(Matcher|Engine)Match$' \
  -benchtime=100x -benchmem ./internal/rule/...
```

Both benchmark families build their matcher or immutable rule state and check
the expected result set outside timed regions. The full-engine fixtures also
prepare Base64 packet payloads before timing. Final results are kept local to
the benchmark and retained explicitly, so no shared mutable sink is required.

For bounded discovery and execution of only the alert-store cases:

```bash
cd engine
go test -run '^$' -bench 'BenchmarkStore(WriteBatch|Query)$' \
  -benchtime=10x -benchmem ./internal/alert
```

The write cases use the public primary-store `WriteBatch` path with the real
recovery append, sync, SQLite commit, and recovery-log clear lifecycle. Each
timed operation receives deterministic unique event identities, while direct
cardinality checks and row cleanup run with the timer stopped so the store
stays bounded at one aggregate row and at most 32 event rows. Query setup
writes 512 deterministic rows through the same production path once; exact
rule and timestamp-range plans must name their expected SQLite indexes before
timing, and result cardinality is checked before and after timing.

The C benchmark iteration count defaults to `100000` and can be overridden:

```bash
BENCH_ITERATIONS=1000000 make bench
```

Some sandboxed environments block Unix socket `bind(2)` or tracing-sensitive sanitizer behavior. In that case, run the same command in a normal local shell.

To capture the same complete C and Go surface in one machine-readable local
evidence envelope, run:

```bash
make benchmark-evidence
```

The command records the full Git commit and tree IDs, clean/dirty state,
OS/kernel/architecture and toolchain fingerprint, explicit C/Go parameters,
redacted raw output, and parsed metrics for every named case. It fails closed
on a missing, duplicate, unknown, malformed, or failed benchmark. Output
defaults to ignored `docs/evidence/local/benchmark/`; override it with
`BENCHMARK_EVIDENCE_OUTPUT=/safe/local/path/result.json`. Local repository,
home, and temporary paths are redacted by default.

Use `BENCH_ITERATIONS` and `GO_BENCHTIME` to select an explicit bounded run:

```bash
BENCH_ITERATIONS=10000 GO_BENCHTIME=100x make benchmark-evidence
```

These captures are local synthetic microbenchmark evidence. A successful
capture applies no numeric threshold and grants no production-throughput,
cross-host, release, tag, or publication claim.

The checked-in R90-74 same-host observation set can be independently
recomputed from all five retained raw samples:

```bash
make benchmark-baseline-check
```

The versioned aggregation API requires at least five individually valid clean
captures with identical commit/tree, environment, parameters, commands, and
metric surfaces. It binds sample SHA-256 values and reports median, inclusive
quartiles/IQR, arithmetic mean, sample standard deviation, coefficient of
variation, and range relative to median for every reported metric. An edited
raw sample, digest, or aggregate value fails revalidation.

For an end-to-end pressure smoke:

```bash
make e2e-pressure
```

The default run repeats the six-packet synthetic pcap pattern 1000 times, for 6000 packets and 5000 generated alerts before SQLite aggregation. Increase the size with:

```bash
PRESSURE_REPEATS=10000 make e2e-pressure
```

Larger local runs may need extra time for the worker and SQLite aggregation to
drain after capture exits. Tune the post-capture wait loop with:

```bash
PRESSURE_REPEATS=10000 PRESSURE_WAIT_ATTEMPTS=1200 make e2e-pressure
```

The script reports elapsed time, packet throughput, alert throughput, and verifies:

- expected packets received and processed
- expected raw alerts generated
- zero decode and alert write errors
- five SQLite aggregated alert rows
- aggregated alert counts equal the raw alert total

For local sanitized corpus evidence:

```bash
PCAP_CORPUS=/path/to/sanitized-pcaps make e2e-corpus-pressure
```

`PCAP_CORPUS` may be a single pcap file or a directory containing `.pcap` and
`.pcapng` files. The script writes JSON and Markdown summaries under
`docs/evidence/local/` by default. That directory is ignored because private
traffic filenames, paths, and operator notes may be sensitive.

---

## Local Baseline

Run date: 2026-06-29

Environment:

```text
OS: Linux VMware-Virtual-Platform 6.17.0-35-generic x86_64
Go: go1.22.2 linux/amd64
GCC: Ubuntu 13.3.0
Iterations: 100000
```

Results:

| Benchmark | Result |
| --- | ---: |
| `bench_parser/tcp_plain` | 212.40 ns/packet, 4,708,124 pps |
| `bench_parser/tcp_vlan` | 214.59 ns/packet, 4,660,096 pps |
| `bench_parser/tcp_qinq` | 206.32 ns/packet, 4,846,956 pps |
| `bench_uds_sender/format_packet_json` | 298.26 ns/op, 3,352,798 ops/sec |
| `bench_uds_sender/format_heartbeat_json` | 215.01 ns/op, 4,650,978 ops/sec |
| `bench_uds_sender/uds_send_line` | 1,856.36 ns/op, 538,689 ops/sec |

The C UDS sender reported:

```text
avg_json_serialize_us=0.24 write_errors=0
```

R90-70 and R90-71 added reproducible Go matching and SQLite-store benchmark
coverage. R90-74 now retains their numeric output together with the C surface
as five matched observations at exact clean commit
`b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`. The historical C table predates
those cases and uses a different Go toolchain, so it remains separate from the
current complete-surface record.

The end-to-end pressure smoke prints a result line like:

```text
[pressure] ok: packets=6000 alerts=5000 aggregated_rows=5 elapsed_sec=... pps=... alerts_per_sec=...
```

Local pressure-smoke samples for the current machine and configuration:

| Run date | Repeats | Packets | Raw alerts | Aggregated rows | Elapsed | Packet rate | Alert rate |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 2026-06-29 | 1000 | 6000 | 5000 | 5 | 4.582 s | 1309 pps | 1091 alerts/sec |
| 2026-06-30 | 10000 | 60000 | 50000 | 5 | 42.788 s | 1402 pps | 1169 alerts/sec |
| 2026-07-06 | 10000 | 60000 | 50000 | 5 | 108.615 s | 552 pps | 460 alerts/sec |

---

## Evidence and Comparability

The current surfaces measure different boundaries and must remain separate:

- C parser, formatter, and local-drain UDS results are standalone custom
  microbenchmarks.
- Go Aho-Corasick, full-rule, durable-write, and indexed-query results use
  `testing.B` with allocations and deliberately bounded fixtures.
- `make e2e-pressure` measures one repeated six-packet synthetic pipeline mix,
  including drain time; it is a functional pressure smoke, not a microbenchmark.
- `make e2e-corpus-pressure` reports evidence for the exact supplied corpus,
  rules, commit, and host. It is local-only by default and does not establish a
  portable traffic model.
- `/api/metrics` exposes process-lifetime rates, latency histograms, queue
  depth, and error/state counters for observation. These values are not a
  benchmark sample set by themselves.

The three historical repeat-pcap samples range from 552 to 1,402 pps. That
spread demonstrates that unmatched local runs cannot support the previously
aspirational 10% portable regression figure. R90-73 delivered versioned
complete-surface capture, and R90-74 records five sequential default-parameter
observations from one clean commit and unchanged environment under
[`r90-74-single-host-benchmark-baseline/`](evidence/r90-74-single-host-benchmark-baseline/).
The highest recorded coefficient of variation is 11.252396% for
`BenchmarkMatcherMatch/no_hit` `ns/op`; this is an observation, not a failure
or proposed budget. A portable/same-host/observation-only budget decision
remains blocked in R90-75 until comparable-environment evidence and explicit
product/SLO scope exist.

See
[`performance-evidence-audit-20260805.md`](audit/performance-evidence-audit-20260805.md)
for the execution-boundary and evidence-gap reconciliation.

## Current Interpretation

Parser and JSON formatting costs are not the obvious bottleneck in the current microbenchmarks. The UDS line write benchmark is materially slower than parser-only and JSON-only paths, which is expected because it crosses the socket boundary.

The complete local microbenchmark surface now has one five-sample matched-host
observation record. Remaining questions concern repeatability across later
same-host sessions and, separately, comparable environments and representative
authorized corpora. The current pressure smoke reports:

- packets read from pcap
- packets delivered over UDS and processed by the worker
- Go receiver decode errors
- raw alerts generated
- SQLite aggregation rate
- total pcap-to-alert runtime

It also exposes worker match latency and alert write latency histograms,
current and high-water packet queue depth, and process-lifetime packet/alert
rate gauges through `/api/metrics`. One approved public real-traffic pressure
record exists, but it remains evidence for that exact corpus and environment.
The honest target remains functional correctness plus explicitly classified
local measurements, not a published production PPS guarantee.

`make e2e-corpus-pressure` provides the release-candidate evidence path for those
realistic corpora once sanitized samples are available. Corpus results should be
interpreted as local evidence for the specific sample set, rule set, and machine.
