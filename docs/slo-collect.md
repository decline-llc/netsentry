# Packet ledger adapter and departmental collection handoff

`scripts/slo_collect.py` converts finalized external packet/oracle and lifecycle
ledgers into the [reporter's observation schema](slo-report.md). This adapter
implements correlation and retention, **not live instrumentation**. Tests are
**not run, delegated by the user**. No measured capacity or SLO compliance is
claimed. The [acceptance contract](performance-slo.md) remains authoritative.

## Department workflow

Freeze the manifest and complete the actual run plus drain before exporting.
Retain every offered eligible packet, including packets with no arrival or
completion. Build expected alerts from the fixture/rules oracle independently
of observed alerts; deriving this list only from received alerts hides failures.

```bash
# All three inputs must already exist; the output's parent must exist.
python3 scripts/slo_collect.py --manifest run.json --offered offered.jsonl \
  --events events.jsonl --output-dir retained-run --scratch-dir scratch
python3 scripts/slo_report.py --input retained-run/observations.json
```

These commands are handoff instructions, not executed validation. The second
command retains the existing default
`~/benchmarks/r90‑75/acceptance_{staging|prod}_{YYYYMMDD_HHMMSS}.json` path.
Retain that report **and the entire adapter bundle**, plus the external clock,
fixture, configuration, durability and workload evidence. The report's
`source.sha256` must equal the receipt's digest for `observations.json`.

The Python API is `scripts.slo_collect.collect(manifest, offered, events,
output, scratch_dir=None)`, with `pathlib.Path` arguments. It returns the
receipt, not a compliance result. CLI exit 0 means adaptation completed; exit 2
means input/filesystem/SQLite failure. Neither means an SLO pass.

## Exact input formats

JSON must be UTF-8, with no duplicate members, NaN or Infinity. Unknown fields
are rejected. IDs use the reporter's bounded ASCII identifier grammar. Times
are integer nanoseconds relative to one documented run origin, with a common
clock/error bound. No timestamp conversion or clock calibration occurs here.

### Manifest

Use the reporter's exact root fields, **omit `expected_alerts`**, and replace
each `packet_windows` entry with only:

```json
{"start_second": 0, "phase": "sustained", "observation_complete": true}
```

All other metadata, policy, resources, provenance and completion constraints
are exactly those in [the report schema](slo-report.md). Windows cover the run
in consecutive 60-second intervals; burst windows are isolated, nonadjacent
60-second intervals. `execution_complete` must be true. The adapter computes
counts; it does not accept asserted counters in the manifest. Completeness
flags remain externally asserted and require raw acquisition-log review.

### Offered ledger: one JSON object per line

Each eligible offered packet occurs exactly once:

```json
{"packet_id":"pkt-1","offered_ns":1000000,"offered_bytes":512,"expected_alerts":[{"event_id":"event-1","rule_id":"rule-7"}]}
```

This is schema illustration, not measurement evidence. `expected_alerts` may
be empty. Packet IDs and event IDs are globally unique in the run; a packet/rule
pair occurs at most once. `offered_bytes` is a positive integer using the frozen
workload's byte boundary. `offered_ns` is in `[0, duration_seconds * 1e9)`.
Packets omitted from this ledger cannot be recovered by the adapter: the
department must reconcile it against the generator's complete eligible offers.

### Event ledger: one JSON object per line

Exactly these three row types are supported:

```json
{"kind":"arrival","packet_id":"pkt-1","at_ns":2000000}
{"kind":"durable","event_id":"event-1","at_ns":9000000}
{"kind":"processed","packet_id":"pkt-1","at_ns":10000000}
```

- `arrival`: live packet arrival, including subsequent buffering and queueing
  in the latency interval. A matching-start timestamp cannot stand in for it.
- `durable`: successful durable persistence for the expected alert. A queue
  enqueue, write attempt or unverified store return cannot establish this.
- `processed`: terminal successful processing of the unique packet, after all
  expected alert durable writes. The current worker's early processed counter
  is unsuitable. Failure/no terminal observation is represented by absence.

Physical instrumentation must produce and validate these boundaries. The
[opt-in engine exporter](slo-runtime.md) now emits these lifecycle rows from
supplied live-arrival metadata. Native capture identity integration is pending.
File order may be arbitrary; all offers are loaded before events. Arrival must follow offer; durable must
follow arrival; terminal processing must follow arrival and any known durable
writes. All lifecycle times are at or before `observed_through_ns`, including
valid zero. Duplicate observations, even identical ones, and events for unknown
packet/event IDs reject the bundle. Timestamp inconsistencies reject it too.

Missing observations are retained, not manufactured: no arrival/durable means
null in the expected-alert record. No terminal marker means packet failure;
a terminal marker with any missing expected durable alert also means packet
failure. Durable alerts without a terminal marker remain observed for latency,
while the packet remains incomplete. Multi-rule packets require every expected
alert's durable observation. Extra unexpected detections are not a supported
row type; preserve them in external logs for separate correctness review.

## Accounting and retained evidence

The adapter uses unique offered packet identities for minute denominators and
unique terminal-success identities for numerators. Minute membership uses offer
time; completions through the declared drain boundary apply to that original
cohort. This is retrospective accounting, not a realtime monitoring endpoint.
The downstream reporter keeps missing expected alerts in p99 and deadline
counts and computes phase/minute/rolling-window summaries.

A new private output directory contains:

| File | Meaning |
| --- | --- |
| `manifest.json` | Exact consumed manifest bytes |
| `offered.jsonl` | Exact consumed oracle ledger bytes |
| `events.jsonl` | Exact consumed lifecycle ledger bytes |
| `observations.json` | Validated reporter input, including all expected alerts |
| `receipt.json` | Published last; file byte counts/SHA-256, ledger row counts, adapter source SHA-256 and explicit review/no-compliance flags |

Sources must be finalized, closed exports; do not append or modify them during
adaptation. Receipts bind consumed bytes, not the acquisition process, oracle
completeness or authenticity. Record the adapter Git commit alongside the
receipt; its source digest identifies the exact script used. The manifest's
harness commit identifies the external collector. No private input path is
included in the receipt.

Existing output directories are rejected. Files are exclusively created and
synchronized, and the final receipt uses the reporter's exclusive publication.
On error, partial output is retained for inspection and is not automatically
deleted or reused. Without a complete receipt and matching file hashes it is
not a completed bundle. A late sync error can leave a complete visible receipt;
inspect it before retrying to a fresh destination. The bundle is not published
as one atomic directory transaction.

JSONL lines are bounded to 256 KiB (including any newline); blank/malformed
lines are rejected. Manifest and observation JSON must fit the reporter's
64 MiB limit. The alert construction budget is conservative and may reject an
output near that limit. A temporary SQLite index holds packet/event identities;
only minute summaries and bounded alert data remain in application memory.
The index is discarded after adaptation; the retained ledgers permit rebuilding.
Use a scratch filesystem with sufficient free space. Raw copies and index grow
with packet volume; no production-scale speed, disk budget or capacity has
been demonstrated. The adapter does not modify SUT/runtime/CI configuration.

## Required departmental validation (not executed)

| Boundary | Required coverage |
| --- | --- |
| JSON/manifest | Unknown/missing fields, duplicate members, wrong types, nonfinite values, oversized lines/documents, blank/truncated rows, invalid IDs/digests, profile/policy/window constraints |
| Identity | Duplicate packets/events/packet-rule pairs, duplicate lifecycle rows, unknown references, multiple rules, zero events, oracle completeness independent of observations |
| Time | Shuffled row order, zero and exact run/drain boundaries, offer after arrival, durable without/before arrival, terminal before arrival/durable, clocks and uncertainty |
| Missing data | Missing arrival, durable or terminal; terminal with one missing expected alert; durable without terminal; packets with no expected alerts; missing-as-infinite downstream p99 and deadline counters |
| Cohorts | Minute-edge offers, drain completions, separate sustained/burst phases, complete byte/packet denominators, no duplicate numerator inflation |
| Retention | Recompute all hashes and output from retained sources; cross-check report source digest; existing directory preservation; permissions; interrupted/read/disk-full/fsync errors; receipt-last behavior |
| Scale/integration | SQLite scratch cleanup, bounded memory and actual disk/throughput cost, reporter compatibility, long acceptance runs, real arrival/durable/terminal instrumentation |

No row in this handoff represents a passed test. Full ingress instrumentation
and both profiles' acceptance remain outstanding with the specialist department.
