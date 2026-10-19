# Bounded rootless local-v1 Ethernet fixture

R90-195 provides a no-send frozen fixture, an isolated runner and a budgeted
adapter. It does not certify L1/L2. The 10pps tooling smoke completed with exact
packet/event accounting; its ~692/694ms alert delays expose a buffering risk.
R90-198 must resolve that risk before R90-196 runs the frozen L1 workload.

## Run modes

Use the existing Ubuntu VM with Linux user/network namespaces, `ip`, user-systemd
and delegated cgroup-v2 cpu/memory controllers. No sudo or host interface is used.
The runner requires six available logical CPUs, >=8GiB free disk and an existing
private evidence root. Build with the repository-pinned Go toolchain first.
Choose a different run ID for every invocation; destinations are never reused.

```bash
make build
mkdir -m 700 /tmp/netsentry-local-v1-evidence
python3 scripts/local_fixture.py dry-run --root /tmp/netsentry-local-v1-evidence --run-id plan-01
python3 scripts/local_fixture.py smoke --root /tmp/netsentry-local-v1-evidence --run-id smoke-01
```

`dry-run` freezes 100 rules and the full 108000-row planned oracle without opening
a network socket or starting the SUT. `smoke` has a separate frozen 2s/10pps
schedule, 20 frames/two positives and a 3s observation; it cannot be derived as
local-v1 L1. The following full-run commands belong to the subsequent acceptance
increment, after R90-198; they were **not** executed in R90-195:

```bash
python3 scripts/local_fixture.py run --root /tmp/netsentry-local-v1-evidence --run-id l1-01
python3 scripts/local_fixture.py derive --root /tmp/netsentry-local-v1-evidence --run-id l1-01
```

Full `run` freezes 600s, 100pps, 500pps at seconds 180–239 and 480–539, 108000
planned packets, one positive/second and a 30s observation. Every frame is exactly
512 bytes without FCS/preamble/IFG, untagged Ethernet/IPv4/DF/UDP, containing the
existing NSLO1 marker. Each positive matches exactly one of 100 distinct bounded
keywords. `oracle.jsonl` closes before SUT startup; its hash binds packet identity,
schedule, frame bytes and independently expected event IDs. Intent precedes send;
offer/submission ledgers record actual times and kernel return, including errors.
A missing outcome after interruption is never assumed successful. All partial
artifacts remain. Successful submission alone does not prove capture or delivery.

## Isolation, readiness and cleanup

The launcher enters a fresh rootless user/network namespace, verifies a different
network identity and loopback-only inventory via sockets, then creates one veth
pair. It accepts no interface argument and never enters the host namespace.
A separate Ethernet probe verifies exact inbound peer bytes before SUT startup.
Loopback/API port, UDS, database, run ID and process sessions are private/new.
Engine API/100 rules and UDS readiness, capture's post-filter activation marker,
and a future frozen origin must all hold before offering measurement traffic.
Early child exit is observed while waiting; timeout/cancellation retains failure.

SUT engine/capture share a cgroup with 3GiB memory.max, no swap, four logical CPUs
and a four-CPU quota. Generator has 1GiB/no swap, two disjoint CPUs and a two-CPU
quota. Readback, aggregate cgroup peaks/events and actual process affinities are
sampled every 100ms. This is a shared VM allocation, not exclusive hardware.

After observation, SIGTERM is followed by one separately retained control frame
with `fixture-control` run identity to wake idle libpcap. That noneligible control
packet is excluded by run-marker filtering and never enters measurement offered
counts. Failure to stop within 5s kills only the owned process session and leaves
an incomplete result. Namespace interfaces disappear on namespace exit. Socket
cleanup uses empty-directory removal only; replacement files/symlinks/targets
are preserved. No recursive deletion of evidence or neighboring services occurs.

## Budgets and retained evidence

Use one private evidence root for all runs, copies and scratch. An exclusive
root-directory lock serializes runner/adapter operations; do not add unrelated
files concurrently. Admission sums all regular-file logical bytes, including
retained failures and copies, rejects links/special artifacts, requires 8GiB
free disk and reserves 1GiB inside the cumulative 4GiB ceiling. A fixed run uses
fewer than 32 files, each child bounded by hard 32MiB RLIMIT_FSIZE (including
SQLite/recovery sidecars and logs). The supervisor samples counts/bytes and stops
on envelope overrun. The only outside-root temporary object is a short private
UDS directory/socket, containing no evidence or regular scratch bytes.

`derive` reserves another GiB under the same lock, verifies sealed raw hashes,
keeps adapter copies and SQLite scratch under the root, caps each derived file at
64MiB, and preserves incomplete output on failure. The fixed adapter has fewer
than eight output/scratch files. No full-run adapter capacity or L1 measurement
is claimed by synthetic adapter regression. Keep later reconstruction/bundle
copies under the same root and perform the same admission/reservation; arbitrary
external commands are not constrained by this runner's budget lock. Never copy
evidence elsewhere to evade accounting. Raw files stay outside Git and Vault.

Retained files include frozen/oracle/rules/config/provenance, binary/source hashes,
intent/offered/submission ledgers, capture close metadata, engine lifecycle and
close receipts, SQLite/recovery files, API snapshot, resource/clock samples and
a final hash inventory. Read-only URI-encoded SQLite observation checks integrity,
event count and per-rule aggregate counts separately from API/lifecycle data.
The accounting audit retains missing/extra/duplicate durable identities; exact
packet equality is diagnostic, not a substitute for L1's <=1% segmented policy.

Clock samples bracket host realtime with monotonic reads, retain resolution and
read-only Linux adjtimex metadata. The candidate interval uncertainty sums sampled
realtime/monotonic drift, syscall brackets, clock resolution and libpcap's 1us
quantization. Absolute NTP UTC error is common to both endpoints and retained
separately; it is not added to same-clock intervals. Bounds above 10ms or failed
probes stop execution as inconclusive. Sampling cannot exclude a step and inverse
step entirely between probes: review raw samples and operating conditions before
acceptance. Neither this bound nor an exit-zero smoke is a physical NIC claim.

A finalized full fixture writes the schema-2 collection manifest. `derive` uses
R90-194's collector and evaluator; no reviewed acceptance is issued automatically.
Review full raw ledgers, submitted outcomes, clocks, resource peaks, API/SQLite
and exact raw hashes under [local-v1](acceptance-local-v1.md). Numerical failure
and inconclusive evidence remain visible. L1 requires two separately reviewed
full runs and L2 remains a later increment.
