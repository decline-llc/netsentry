# R90-75 Production SLO Acceptance Contract

Status: **proposed acceptance targets; validation outstanding for both profiles**.
The user selected production SLO evaluation on 2026-09-23 and formally adopted
the three measurement requirements below. Neither profile represents
demonstrated capacity. No qualifying measurement evidence currently exists.

## Development and testing ownership

On 2026-09-25 the user instructed the agent to skip tests, delegate testing to
specialist staff and continue development. Agent-run unit, integration,
benchmark, acceptance and knowledge test suites are deferred. Static source,
syntax and diff review remain part of development delivery; they are not test
results. Missing acceptance hardware or completed measurements does not block
implementation. The department owns actual run setup, behavioral validation,
profile/resource verification and the evidence required for compliance claims.
No performance target, failure-accounting requirement or evidence truthfulness
is waived by this division of work.

## Target profiles

These are the reviewed starting profiles, not an enabled numeric gate. Record
and freeze the exact tested configuration before execution; report deviations.

| Requirement | Low-traffic staging | Production baseline |
| --- | --- | --- |
| Deployment | KVM VM | KVM VM |
| CPU / RAM | 2 vCPU / 4 GiB | 8 vCPU / 16 GiB |
| Storage | 20 GB SSD | 100 GB dedicated SSD |
| NIC | 1 Gbps virtual NIC | 10 Gbps SR-IOV VF |
| Sustained / burst traffic | 100 / 250 Mbps | 3 / 7 Gbps |
| Corresponding sustained / burst packet rate | approximately 24,414 / 61,035 pps | approximately 732,422 / 1,708,984 pps |
| Mean packet size for rate conversion | 512 bytes | 512 bytes |
| Active rules | 2,000 | 20,000 |
| Expected alerts | 15/min baseline; 120/min burst | 15/min baseline; 120/min burst |
| Burst duration | 60 seconds | 60 seconds |
| Live arrival to durable alert write | p99 <=180 ms | p99 <=180 ms |
| Packet loss: sustained / burst | <=0.05% / <=0.2% | <=0.05% / <=0.2% |
| Evaluation and reporting | Rolling 5 minutes; 1-minute telemetry | Rolling 5 minutes; 1-minute telemetry |

Bandwidth is the primary target in this starting profile. Packet rates use
`bandwidth_bits_per_second / (512 * 8)` and exclude additional link overhead.
The initially suggested 1.2/2.8 million pps at a 512-byte mean would require
4.9152/11.4688 Gbps before additional overhead; those packet rates are not
simultaneous acceptance requirements for the 3/7 Gbps profile. Freeze the
actual packet-size distribution and the byte-counting boundary in each run.
A typical packet size alone does not establish a mean.

CPU pinning, VM allocation, storage persistence settings, NIC/RSS setup,
worker count, queue capacity, rule mix, database size, background workload and
traffic protocol mix must be retained with the run configuration. An SR-IOV
VF or multiple queues alone does not establish application throughput.
The production case with 20,000 rules carries greater unvalidated risk.

## Formal measurement requirements

### 1. End-to-end latency

Latency starts at **live packet arrival** and ends upon **successful durable
persistence** of the corresponding expected alert event. Include capture
buffering, UDS transfer, queueing, matching and persistence delays. Existing
matching-path and write-path histograms cannot substitute for this interval;
adding their percentiles also cannot establish an end-to-end percentile.

The harness must correlate offered test-packet identity, observed live arrival,
expected rule-match event and durable completion. Aggregation must not erase
individual expected alert events. Packet timestamps inherited from an offline
PCAP are not live-arrival measurements. Starting the clock at the userspace
capture callback would omit earlier capture buffering. Specify and validate
the arrival timestamp source, clock relationship, resolution and uncertainty,
including the relationship between the live-arrival clock and persistence clock. Invalid
or uncorrelatable timestamps prevent a passing latency claim.

A successful API read or a matching counter alone does not prove a durable
completion timestamp. Freeze the durability configuration and verify the
measurement's completion boundary against the persistent event outcome.

### 2. Offered-versus-completed loss and missing alerts

Compute packet loss by comparing **offered eligible packets** with
**fully-processed results** for the same identified cohort. Define eligibility
before the run from the frozen supported protocol/packet and rule fixture.
Do not remove packets from the denominator because they were lost, rejected,
delayed or caused an error after being offered. Duplicates cannot inflate
completed counts; account for outstanding work at window boundaries and after
the declared drain deadline. Keep sustained and burst loss denominators
separate so a five-minute average cannot conceal a failed burst window.

Missing expected security alerts count as failures and must not be filtered
out of latency calculations. Report expected, completed, late and missing
alert events; include missing events in deadline-violation and failed-event
accounting. A percentile computed only from successful writes is insufficient.
The implementation must specify its representation of missing latency values
and demonstrate that missing events cannot improve the SLO verdict.

### 3. Counts, deadline violations and extended runs

Publish raw alert counts and deadline-violation counters **alongside p99** in
each report. At 15 alerts/minute, a five-minute window contains approximately
75 expected alerts, so a short window cannot characterize the tail reliably.
Extended-duration acceptance runs are required. Freeze the run duration,
minimum event/sample criterion and percentile estimator before execution;
these details remain to be specified. Do not silently increase the alert rate
to accumulate samples and then label the result as the baseline workload.
Retain per-window observations and the full-run summary, including insufficient-
sample and invalid-measurement states. One-minute reporting must preserve the
underlying data needed to recompute the five-minute windows and tail results.

## Local execution and retained artifacts

The user clarified on 2026-09-23 that **all tests execute inside this single
Ubuntu VM**. There is no external bench01 host. The NetSentry test instance and
load generator run on this VM; no SSH connection is needed. This clarification
supersedes the earlier external-runner description and replaces R90-75's
independently provisioned environment prerequisite with an explicitly approved
**isolated local execution context**:

- Separate test working directory and per-run evidence/state directories.
- Isolated test process groups with cleanup limited to those processes.
- No reuse of production service runtime, sockets, databases, ports or state.
- A recorded local ingress choice: loopback or the VM's selected main interface,
  with link type, routing, traffic containment and packet counting verified
  before a live run. Merely finding an interface does not select it for traffic.

These controls establish process/state isolation, not independent hardware.
Generator and SUT share CPU, memory, storage and networking resources; record
both allocations and generator headroom. A generator-limited run cannot prove
SUT capacity. Results must be labeled single-VM local evidence, with no
cross-host portability or independent-machine claim. Any pass applies only to
the exact tested profile and configuration.

Read-only local discovery found x86_64, 16 visible logical CPUs, and
8,078,816 KiB of guest RAM (approximately 7.70 GiB). Interfaces include `lo`,
`ens33`, `ens37` and `docker0`; no interface has been selected or exercised.
Visible CPUs do not prove dedicated allocation. The current VM cannot satisfy
the proposed production profile's 16 GiB guest-memory requirement. Production
load attempts on this allocation would be exploratory evidence with a hardware
deviation, not a qualifying pass for that profile. Staging resource isolation,
SSD/NIC properties and generator headroom also remain unverified. Profile
compliance requires the actual resources to match the accepted profile or a
separately recorded revision of that profile before execution.

The designated output template for completed acceptance runs is, verbatim:

```text
~/benchmarks/r90‑75/acceptance_{staging|prod}_{YYYYMMDD_HHMMSS}.json
```

The directory uses U+2011 (NON-BREAKING HYPHEN) between `r90` and `75`, as
supplied. Braces express alternatives/placeholders; each actual file identifies
one profile and one timestamp. Do not silently replace the directory with
ASCII `r90-75`. Resolve `~` in the local test execution account, agree timestamp
timezone, and avoid overwriting prior results. No placeholder result is written
to this completed-run destination.

Full retained artifacts must allow independent review and recomputation:

- Run identity, profile, start/end/timezone, pass/fail/inconclusive outcome and
  explicit reasons; capture failed and incomplete execution evidence too.
- Exact SUT/harness commits, clean/build identities, toolchains, configuration,
  rules/fixture hashes and approved traffic provenance; actual hardware,
  resource allocation, local process/state isolation and offered-load schedule.
- Eligible offered and uniquely completed packet counts; sustained/burst
  denominators, errors, outstanding work, loss and queue/drain observations.
- Expected, durably completed, missing and late alert counts; deadline
  violations, p99 estimator/result, sample count and measurement validity.
- Timestamp/correlation method and clock uncertainty; underlying timestamped
  measurements, per-minute and rolling-window results, extended-run coverage
  and durability-boundary verification.
- Full raw logs/samples embedded or retained as checksum-bound companion
  artifacts. A summary JSON alone is insufficient if its source data are lost.

R90-114 implements [the supplied-observation report tool](slo-report.md) in
`scripts/slo_report.py`: strict input validation, loss/latency/deadline summaries,
source-byte retention and review-required JSON output. Its behavioral tests
are delegated and have not been executed by the agent. This is a report
implementation, not a live collector or an automatic SLO compliance gate.

## Current evidence gaps and execution prerequisites

- `capture/src/main.c` forwards libpcap timestamps and records send/drop
  counters. It does not establish the required offered-to-completed cohort
  ledger or verified live-arrival clock boundary.
- `engine/internal/pipeline/worker.go` increments the processed counter before
  matching finishes, and independently times matching and `WriteBatch`.
  Neither that counter nor those timers proves this acceptance contract.
- `engine/internal/stats/stats.go` exposes those component metrics;
  `scripts/e2e_pressure.sh` validates a synthetic fixture and total elapsed
  time. Neither supplies expected-event end-to-end latency evidence.
- The earlier SSH preflight failed name resolution before any remote command.
  The user then clarified there is no remote host; SSH/DNS is no longer an
  execution prerequisite. No acceptance traffic ran.
- The department still needs local test resource allocation and isolated
  service endpoints/ingress, frozen fixtures/configuration, extended-run policy,
  live measurement collection and verified artifacts. These acceptance
  prerequisites no longer block agent implementation work.
- The observed guest RAM is below the proposed production profile. Match that
  profile before qualifying execution, or explicitly revise its hardware scope.

Both profiles remain **unvalidated**. SLO compliance can only be asserted after
successful execution in the agreed isolated local VM context with full retained
artifacts at the designated location and verification against the exact tested
profile and this contract. Documentation checks and the existing
[R90-74 microbenchmark baseline](evidence/r90-74-single-host-benchmark-baseline/README.md)
are not production acceptance evidence. The comparison study now uses the
approved isolated same-host context; preserve exact commit/toolchain and
resource comparability or explicitly plan a new matched baseline. This change
in environment scope does not create completed measurements or prove hardware
independence.

See [performance evidence](performance.md) and the
[active roadmap](plans/rolling-90-day-roadmap.md) for delivery status.
