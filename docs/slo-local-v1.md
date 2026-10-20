# local-v1 versioned evidence tools

R90-194 supports **supplied** local-v1 L1 observations. The evaluator does not
send traffic, establish clock accuracy, prove durable storage, or complete L1/L2.
Use the frozen [acceptance contract](acceptance-local-v1.md) and R90-195
[isolated runner](local-fixture.md). R90-198 repairs measurement buffering and idle
shutdown before two measured runs R90-196, then stability/recovery R90-197.

## Versions and identity

Observation and collection manifest **schema_version=2**, **profile=local-v1**,
**contract_id=local-v1** must occur together. Unknown/mismatched versions, members,
contracts and profiles fail closed. Schema 1 still means historical staging/prod:
its 180ms, resource/bandwidth constants, diagnostics and summary algorithm remain
unchanged. Eight golden summaries were frozen from commit `9d95c1b2f232ceaa14af0afd0abf36382cc0e4ac`
for review, failure, inconclusive and combined outcomes across both old profiles.
They are synthetic compatibility fixtures, not performance measurements.

Schema versions belong to individual artifact types. Existing adapter receipts,
sender/capture/engine envelopes, bundle and reconstruction schemas retain their
own versions. A bundle v2 means adapter reconstruction; it does not mean local
observation v2. Context declarations/receipts use version 2 for local-v1 and bind
contract ID as well as run/profile/origin/observation digest. Pair comparison
rejects pairs crossing local-v1 and legacy profiles or mismatched contracts as
invalid evidence; legacy staging-versus-prod retains `conditions_differ`.

Legacy envelope `departmental_review_required` fields remain serialized for
compatibility; for a local-v1 run their review is performed locally under AGENTS,
without an external department. They never override the explicitly local report
review state. Bundle/context `review_required` means retained-artifact consistency,
and pair `review_required` means matching declared conditions, even when the
underlying measurements fail or are inconclusive. These statuses are never
numerical acceptance; inspect the retained report and each side's measurement status.

## Observation schema 2

All members are required; unknown members are rejected. See `observation()` in
[scripts/test_slo_local.py](../scripts/test_slo_local.py) for an executable synthetic
example. Tests build deliberately synthetic evidence and must not be saved as a
measured local-v1 run.

Common observation members have their schema-1 meanings: `run_id`, `started_at`
(UTC), `duration_seconds`, `drain_seconds`, `observed_through_ns`,
`execution_complete`, `policy`, `measurement`, `resources`, `provenance`,
`packet_windows`, `expected_alerts`. Identity adds `contract_id`. Differences:

| Member | Required representation |
| --- | --- |
| policy | Exactly `minimum_duration_seconds: 600`, `minimum_expected_alerts: 600`; cannot lower minima per run |
| measurement | Existing `live_arrival`, `durable_completion`, `clock_method`, plus `clock_checked` boolean; `clock_uncertainty_ns` integer or explicit null |
| resources | `sut_vcpus`, `sut_memory_bytes`, `active_rules`, plus `generator_vcpus`, `generator_memory_bytes`; positive integral allocation limits |
| integrity | Integer `capture_errors`, `submission_errors`; boolean `raw_evidence_complete`, `identity_accounting_complete` |
| packet_windows | Existing ordered one-minute rows plus `submitted_eligible`; unique completed <= submitted <= offered; actual eligible byte totals |
| expected_alerts | All oracle event/packet/rule IDs with offer time and nullable arrival/durable times; same strict identity and ordering rules as v1 |
| anomalies | Array of `{kind, identity, at_ns, reason}`; kind arrival/processed/durable; reason unexpected/duplicate/after_observation; retain every occurrence |

The evaluator requires exactly 600 workload seconds plus 30 observation seconds,
100 active rules, one oracle event per workload second, 100 pps steady and 500 pps
at seconds 180–239 and 480–539. Actual offered count in **each minute** must lie
within inclusive +/-5% of 6000/30000. Planned total is 108000; actual offered,
submitted, completed, bytes, mean frame length and load are reported separately.
Submission/capture errors, incomplete execution/accounting/raw evidence, missing
clock proof or uncertainty >10ms prevent acceptance. Allocation limits are <=4
SUT CPUs/3GiB and <=2 generator CPUs/1GiB; these do not prove runtime headroom.

Loss <=1% is evaluated across sustained minutes together and independently for
each 60-second burst. Duplicates cannot inflate completed counts. Every oracle
event participates in nearest-rank p99, with missing completion as positive
infinity. Missing events fail even when the full-run p99 rank stays finite.
Full-run and sustained/individual-burst latency summaries include the clock bound;
upper bounds >500ms fail. Individual late events remain counted even when their
fraction is below p99. Unexpected, duplicate or after-observation lifecycle records
fail and remain visible. A numerical failure takes priority over inconclusive
validity; both diagnostic lists are retained.

## Collection and reconstruction

Use existing `slo_collect.py --manifest ... --offered ... --events ... --output-dir ...`.
The schema-2 manifest omits `expected_alerts` and `anomalies`; each window contains
only `start_second`, `phase`, `observation_complete`. All other declarations remain.
Local offered rows contain `packet_id`, `offered_ns`, `offered_bytes`,
`expected_alerts` and **submitted** boolean. Every attempted offer, including a
failed submission, stays in this ledger. Materialize submitted outcomes from the
actual sender ledger; never presume success from a planned schedule. The current
v1 reference sender does not generate this new row contract; the bounded local
runner in R90-195 now supplies actual local submitted outcomes.

Events retain the existing arrival/processed/durable encoding and nanosecond run
origin. An unknown ID, duplicate record or timestamp past the observation boundary
is retained in the anomaly list, never silently discarded or counted as another
successful completion. A late durable record leaves its expected completion null;
raw bytes and the late identity/time survive. Invalid causal order still rejects
collection; retain partial output and retry in a new directory. Capture-error
counts are supplied by the manifest; failed submissions inferred from offered
rows can only increase its submission-error count.

Adapter replay uses the same local collector and compares every field, including
anomalies and counters, against retained raw ledgers. Bundle summary checks
recompute the local evaluator and verify source bytes/digests. Use
`--reconstruct-ledgers` for derivation proof. The separate legacy sender-frame
replay accepts v1 offered rows; it does not certify new local sender rows.
A bundle can retain evidence of a failed/inconclusive run without declaring it
accepted. Existing per-tool byte limits remain active; cumulative 4GiB retention,
copy/scratch accounting and disk/resource enforcement belong to the runner.

Context schema 2 adds contract ID and admits deployment_type `vm` without
requiring KVM. Existing nullable hardware/toolchain/isolation declarations and
opaque evidence references remain; false (for example dedicated storage) is a
known declaration, not an automatic failure. Unknown/unreferenced facts remain
gaps. Declared CPU/memory limits bind to the observation snapshot. Context
consistency does not authenticate opaque evidence or establish clock accuracy.

## Reporting and explicit review

```bash
python3 scripts/slo_report.py --input observations.json --output report.json
python3 scripts/slo_report.py --input observations.json --review review.json --output reviewed-report.json
make slo-check
```

Without `--output`, local reports use `~/benchmarks/local-v1/<run_id>/report.json`.
Existing destinations are never overwritten. The reporter retains exact input
bytes and SHA-256. Exit 1 means `measurement_failed`, 3 `inconclusive`, 2 invalid
input/I/O, and 0 either `review_required` or `reviewed_local_acceptance`. Exit 0
alone is never evidence of acceptance.

Review JSON has exactly `schema_version: 2`, `contract_id: local-v1`, `run_id`,
`observations_sha256`, bounded ASCII `reviewer`, UTC `reviewed_at`,
`evidence_sha256` with bundle/context/raw_manifest SHA-256 values, and `checks`
with all seven booleans true: oracle, clock, durability, load, resources,
raw_evidence, identity_accounting. The reviewer must inspect these retained
artifacts and supporting raw data; this is an explicit human/agent local review
declaration, not cryptographic authentication. Hashes are declarations to make
that review auditable, not automatic proof of the referenced artifacts' truth.

Review binds the exact observation bytes and identity and cannot accept a failed
or inconclusive measurement. The resulting `local_run_accepted: true` and
`reviewed_local_acceptance` apply only to that supplied **single** local run.
`slo_compliance_asserted` stays false; two independent runs and L2 recovery still
need their own retained/reviewed evidence. Synthetic regression review declarations
are only tests. No measured local acceptance has been executed by R90-194.
