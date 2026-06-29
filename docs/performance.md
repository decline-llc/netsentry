# NetSentry Performance Baseline

> Status: v0.1.0 development snapshot. These numbers are local microbenchmarks, not an end-to-end production throughput claim.

---

## Scope

The current benchmark target measures the parts that already have stable standalone harnesses:

- C Ethernet/VLAN/Q-in-Q IPv4 TCP parser microbenchmarks.
- C packet and heartbeat JSON formatting microbenchmarks.
- C Unix Domain Socket line write microbenchmark against a local drain listener.
- Go package benchmark command execution with `go test -bench=.`, although the current Go packages do not yet expose `Benchmark*` functions.

It does not yet measure full pcap-to-SQLite throughput across the C capture process, UDS transport, Go receiver, rule engine, worker, SQLite writer, and HTTP API.

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

---

## Current Interpretation

Parser and JSON formatting costs are not the obvious bottleneck in the current microbenchmarks. The UDS line write benchmark is materially slower than parser-only and JSON-only paths, which is expected because it crosses the socket boundary.

For v0.1.0, the next performance question is end-to-end throughput under realistic rule sets and SQLite writes. That requires a dedicated pressure harness that reports:

- packets read from pcap
- packets delivered over UDS
- Go receiver decode errors
- packet queue depth
- worker match latency
- alert write latency
- SQLite aggregation rate
- total pcap-to-alert runtime

Until that exists, the honest target remains functional correctness plus measured microbenchmarks, not a published production PPS guarantee.
