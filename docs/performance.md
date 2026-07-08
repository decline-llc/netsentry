# NetSentry Performance Baseline

> Status: v0.1.0 development snapshot. These numbers are local microbenchmarks, not an end-to-end production throughput claim.

---

## Scope

The current benchmark target measures the parts that already have stable standalone harnesses:

- C Ethernet/VLAN/Q-in-Q IPv4 TCP parser microbenchmarks.
- C packet and heartbeat JSON formatting microbenchmarks.
- C Unix Domain Socket line write microbenchmark against a local drain listener.
- Go package benchmark command execution with `go test -bench=.`, although the current Go packages do not yet expose `Benchmark*` functions.
- A repeat-pcap end-to-end pressure smoke test across C capture, UDS, Go receiver, rule matching, SQLite aggregation, and API health/alerts checks.
- An optional local corpus pressure evidence script for sanitized `.pcap` and `.pcapng` files supplied by the operator.

The repeat-pcap pressure smoke is intended to catch obvious pipeline regressions and provide a local baseline. It is not a production traffic model and does not replace a broader benchmark corpus.

---

## Reproduction

```bash
make bench
```

The C benchmark iteration count defaults to `100000` and can be overridden:

```bash
BENCH_ITERATIONS=1000000 make bench
```

Some sandboxed environments block Unix socket `bind(2)` or tracing-sensitive sanitizer behavior. In that case, run the same command in a normal local shell.

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

## Current Interpretation

Parser and JSON formatting costs are not the obvious bottleneck in the current microbenchmarks. The UDS line write benchmark is materially slower than parser-only and JSON-only paths, which is expected because it crosses the socket boundary.

For v0.1.0, the remaining performance question is end-to-end throughput under more realistic rule sets and pcap corpora. The current pressure smoke reports:

- packets read from pcap
- packets delivered over UDS and processed by the worker
- Go receiver decode errors
- raw alerts generated
- SQLite aggregation rate
- total pcap-to-alert runtime

It now exposes worker match latency and alert write latency histograms, current and high-water packet queue depth, and process-lifetime packet/alert rate gauges through `/api/metrics`, but the current recorded baseline still does not include realistic pcap corpora. Until those results are recorded, the honest target remains functional correctness plus measured local benchmarks, not a published production PPS guarantee.

`make e2e-corpus-pressure` provides the release-candidate evidence path for those
realistic corpora once sanitized samples are available. Corpus results should be
interpreted as local evidence for the specific sample set, rule set, and machine.
