# Departmental SLO execution and artifact-review runbook

R90-124 consolidates the implemented measurement and review chain. **All commands
below are unexecuted handoff templates.** Behavioral testing and acceptance belong
to the specialist department under the Sep 25 user instruction. Neither staging
nor production has qualifying measurements. The [SLO contract](performance-slo.md)
is authoritative; tool success means only the documented local operation succeeded.

For optional offline sender-source correlation, see the
[sender reconstruction guide](slo-sender-reconstruct.md). It retains and joins
fixture/offer/submission bytes without traffic. Bundle/pair `--reconstruct-sender`
selects fresh sender replay; `--reconstruct-ledgers` selects adapter replay.
Both remain behaviorally unvalidated.

## 1. Freeze the run and resolve execution prerequisites

Testing takes place in this single Ubuntu VM, with separate working directories,
fresh SUT state/endpoints and isolated process groups. There is no external bench01
host or SSH prerequisite. Generator and SUT share hardware; record their allocations,
contention and generator headroom. Do not reuse production services, sockets,
databases or ports. The department must supply and retain the following before
live execution; this runbook does not choose unknown deployment values.

| Decision / evidence | Required handoff |
| --- | --- |
| Run identity and time | Unique bounded `RUN_ID`; canonical whole-second `UTC_ORIGIN` in `[2000,2100)`, equal to manifest `started_at`; origin allows startup/readiness before scheduled offers |
| Profile and isolation | `staging` or `prod`, actual CPU/RAM/storage/NIC allocation, pinning/RSS, generator allocation/headroom, isolated endpoints, owned processes and cleanup procedure |
| Source/build identity | Exact SUT/harness commits, build/toolchain identity, configuration/rules digests and retained bytes; distinguish tool source hashes from authenticated builds |
| Traffic and oracle | Frozen marked fixture, independently derived expected rules/events, eligible packet set, supported protocols, mean size/distribution, byte boundary and sustained/burst schedule |
| Measurement | Live-arrival boundary, clock normalization/resolution/elapsed-time uncertainty, durable completion evidence and WAL configuration; declarations alone cannot qualify them |
| Run policy | Duration, post-offer drain, minimum expected alerts, extended-tail sample justification, percentile method, one-minute cohorts and isolated 60-second bursts; freeze these before execution |
| Storage and review | New output names, existing parent/scratch directories, disk allocation for originals, nested copies and SQLite indexes; retained process logs/exits, final reviewer and unresolved gaps |

The reporter fixes nearest-rank p99 and treats missing completions as positive
infinity; JSON uses `p99_kind=unbounded_missing` and a null upper bound when that
rank is missing. Its chosen minimum duration must exceed 300 seconds; this floor
and a positive minimum alert count do not establish sufficient tail evidence.
At 15 alerts/minute, five minutes gives about 75 expected events. The department
must justify extended coverage without increasing the baseline alert rate silently.

The native measurement lane is untagged Ethernet/IPv4/nonfragmented UDP on one
explicit destination port. `NSLO1` markers change payload length and matching;
compute the oracle over marked payloads. Mean packet size uses complete Ethernet
frame bytes including padding and excludes FCS/preamble/inter-frame gap. TCP,
VLAN and mixed production workloads are not qualified by this lane. The general
contract permits a locally selected ingress, but this implementation requires
Ethernet datalink, host/microsecond timestamps and inbound capture support. Do not
assume loopback or sending on the capture interface provides that boundary.
The department must resolve and verify the local topology and traffic containment.

## 2. Acquire sender, capture and engine artifacts

Use a department-managed isolated working directory and the exact reviewed build.
The repository `make build` outputs `bin/netsentry-engine` and
`bin/netsentry-capture`; it is not invoked by this documentation increment.
The commands below are separate process invocations, not a sequential shell
script. Variables are unresolved department-supplied values. Relative evidence
paths resolve under the department's working directory; `REPO` identifies its
reviewed checkout. Parents must exist; output files/directories must be new.

Start the engine first, with an isolated configuration referencing the frozen
rules, dedicated UDS/API/state paths and WAL storage:

```bash
"$REPO/bin/netsentry-engine" --config "$ISOLATED_CONFIG" \
  --slo-output-dir engine-run --slo-run-id "$RUN_ID" --slo-origin "$UTC_ORIGIN"
```

Retain readiness and startup outcomes before starting capture. The engine emits
an `engine ready` log after receiver/worker/HTTP setup; the department must verify
usable isolated endpoints and absence of startup failure. A directory or socket
pathname alone is insufficient readiness evidence.

```bash
"$REPO/bin/netsentry-capture" -i "$CAPTURE_INTERFACE" -s "$ISOLATED_UDS_PATH" \
  --slo-run-id "$RUN_ID" --slo-port "$DESTINATION_PORT" \
  --slo-output capture-summary.json
```

All three capture measurement flags are required with live `-i`; offline `-r`
is rejected. Retain successful capture activation, inbound/filter setup and UDS
connection before offers. Do not substitute a fixed sleep for readiness proof.
The adapter collector defaults to a cumulative 64 GiB source-byte budget across
manifest, offered and events files; `--max-input-bytes` can set another positive
limit. Admission requires regular files and source metadata stability. This
budget excludes generated outputs and SQLite scratch. See the
[ingress contract](slo-ingress.md) for fixture schema, link fields and privileges.
Only after both components are ready, invoke the sender in its owned process group:

```bash
python3 "$REPO/scripts/slo_ingress.py" \
  --interface "$SEND_INTERFACE" --fixture fixture.jsonl --output-dir sender-run \
  --run-id "$RUN_ID" --origin "$UTC_ORIGIN" \
  --src-mac "$SOURCE_MAC" --dst-mac "$DESTINATION_MAC" \
  --src-ip "$SOURCE_IPV4" --dst-ip "$DESTINATION_IPV4" \
  --src-port "$SOURCE_PORT" --dst-port "$DESTINATION_PORT"
```

This invocation sends raw packets and needs appropriate AF_PACKET privileges.
An offer row is retained before kernel submission. Submission success proves
kernel acceptance only; reconcile scheduled versus actual offer times, submission
errors, independent offered-load evidence and generator headroom. Failed or
ambiguous offers must not be dropped from a qualifying denominator.

Retain the sender outcome, then observe through the predeclared run/drain boundary
while capture and engine remain active. End capture after that observation period,
retain its shutdown/exit result, and stop the owned engine after the department's
completion accounting. SIGINT/SIGTERM drive the implemented shutdown paths.
Engine shutdown cancels workers; it does **not** promise to drain queued packets.
Keep all outstanding packets and expected alerts in the cohort. Do not declare
`execution_complete` from a successful close receipt or shorten the recorded
observation interval to hide failures. Close every writer and record actual
process outcomes before offline review. Retain interrupted and failed runs too.

| Producer | Required retained artifacts / boundary |
| --- | --- |
| Sender | `sender-run/fixture.jsonl`, `offered.jsonl`, `submissions.jsonl`, `submission.json`; receipt written after successful submission/file lifecycle, not acceptance |
| Capture | `capture-summary.json`; host timestamp/inbound policy, run/port, drop/parse/marker/pcap counters and close state; counters are not cohort loss denominators |
| Engine | `engine-run/events.jsonl`, `close.json`; arrival/durable/processed rows, run/origin and exact raw digest/bytes/rows after workers stop |
| Department | Process logs/exits, readiness/drain observations, frozen config/rules/builds, oracle/acquisition/clock evidence, resource/isolation proof and persistent database/recovery evidence |

Engine arrival rows carry supplied capture timestamps; they are emitted only when
a worker observes a packet. Absent rows remain missing in the independent oracle.
Durable timestamps follow successful `WriteBatch` return with WAL/FULL settings,
including primary/shard handling. They are conservative upper bounds, not physical
flush timestamps. Aggregation/deduplication and VM/device persistence still need
verification. The event ID is the SHA-256 packet/rule correlation defined in
[slo-runtime.md](slo-runtime.md), not the database row ID. Serialized export and
sender ledger I/O introduce unmeasured overhead that must be retained in review.

## 3. Adapt and report finalized observations

Freeze the intended policy/configuration before the run. After actual completion
review, finalize `completed-manifest.json` using the exact
[adapter schema](slo-collect.md) and [report fields](slo-report.md): one consecutive
60-second window per minute, common origin, actual observation/drain boundary,
resource/provenance values and truthful completeness declarations. Adapter manifests
omit `expected_alerts` and contain no asserted packet counters. In this UDP lane,
`fixture_sha256` identifies the exact retained sender `fixture.jsonl` bytes.
An unfinished run cannot be converted into a completed report by inventing a manifest.
Lifecycle events beyond the declared observation boundary are rejected; retain and
investigate them rather than clipping rows or silently extending the frozen drain.

```bash
python3 "$REPO/scripts/slo_collect.py" --manifest completed-manifest.json \
  --offered sender-run/offered.jsonl --events engine-run/events.jsonl \
  --output-dir retained-run --scratch-dir "$SCRATCH_DIR"
```

Retain all five adapter files: `manifest.json`, `offered.jsonl`, `events.jsonl`,
`observations.json`, `receipt.json`. The adapter joins unique packet/event/rule
identities, includes every expected event and requires terminal success plus all
expected durable alerts before counting a packet as fully processed. Unknown or
duplicate identities and invalid time order are errors. Missing lifecycle records
remain missing; they cannot improve loss or latency results.

```bash
python3 "$REPO/scripts/slo_report.py" --input retained-run/observations.json
```

Default output uses the exact designated template:

```text
~/benchmarks/r90‑75/acceptance_{staging|prod}_{YYYYMMDD_HHMMSS}.json
```

The directory contains U+2011, not ASCII hyphen. The timestamp comes from
`started_at` in UTC. Resolve the actual profile and timestamp into `REPORT_PATH`
for subsequent commands. No placeholder acceptance artifact is supplied here.
`--output "$REPORT_PATH"` is an alternative explicit new destination; retain any
such deviation from the designated handoff location. Reports embed exact source
JSON and its SHA-256; verify that it matches adapter observations. Keep the entire
adapter package and external evidence as well as the report.

Review full-run, per-minute and trailing five-minute results as offer-time cohorts
finalized at the declared drain deadline. Separate sustained/burst denominators;
review each burst's loss, all raw expected/completed/missing/late counts and
violations alongside p99. Latency includes the declared uncertainty. Missing
alerts always fail and count as deadline violations, even if p99 remains finite.
A report's failed or inconclusive status survives every downstream review.

## 4. Retain context and review bundles

Prepare the exact [run-context declaration](slo-context.md) and supporting files
from the actual run. Unknown fields are explicit nulls; known values without
references remain gaps. The declaration binds run/profile/start and the SHA-256
of the exact adapter observations. Evidence files use catalog `<id>.bin` names.
The following operations consume existing files; they perform no discovery:

```bash
python3 "$REPO/scripts/slo_context.py" --declaration run-context.json \
  --observations retained-run/observations.json --evidence-dir context-evidence \
  --output-dir retained-context
python3 "$REPO/scripts/slo_bundle.py" --sender sender-run \
  --capture capture-summary.json --engine engine-run --adapter retained-run \
  --summary "$REPORT_PATH" --reconstruct-ledgers --scratch-dir "$SCRATCH_DIR" \
  --output-dir review-bundle
```

Context output contains `declaration.json`, `observations.json`, `evidence/<id>.bin`
and `context.json`. Bundle output contains fixed sender/capture/engine/adapter/report
snapshots plus `bundle.json`. Replay mode also retains `reconstruction/adapter/`,
`reconstruction/rebuilt/` and `reconstruction/reconstruction.json`. Original byte
copies, raw row counts, run/origin/port, adapter/report binding and report recomputation
must agree before bundle replay starts. All five fresh replay source entries bind
to this enclosing bundle; previous reconstruction receipts cannot satisfy this.

For separate derivation diagnostics, the existing standalone operation is:

```bash
python3 "$REPO/scripts/slo_reconstruct.py" --adapter retained-run \
  --output-dir reconstruction-review --scratch-dir "$SCRATCH_DIR"
```

It compares every observation root value after normalizing object order and
expected alerts by event identity; nulls and minute ordering remain significant.
Byte-only formatting differences may coexist with semantic equality. Its output
is not consumed as trusted proof by the bundle or pair modes, which replay afresh.
Source agreement cannot authenticate acquisition or a jointly rewritten oracle.

## 5. Compare two retained runs with an explicit policy

Supply distinct, nonoverlapping baseline/candidate runs and preserve their full
original bundles and contexts. The comparator implements exact declared
same-commit repeatability: source/tool/config/oracle differences are diagnostic
condition differences, not a performance regression verdict. It emits no speed
ratio or significance estimate. A staging/prod pair will differ in profile and
cannot establish matched repeatability. A changed-build comparison needs a
separately documented matched-baseline interpretation.

```bash
python3 "$REPO/scripts/slo_compare.py" \
  --baseline "$BASELINE_BUNDLE" --candidate "$CANDIDATE_BUNDLE" \
  --baseline-context "$BASELINE_CONTEXT" --candidate-context "$CANDIDATE_CONTEXT" \
  --require-context --reconstruct-ledgers --scratch-dir "$SCRATCH_DIR" \
  --output-dir comparison-review
```

| Operation / options | Output schema and policy |
| --- | --- |
| Bundle default | v1; supplied-artifact consistency, no ledger replay |
| Bundle adapter-only `--reconstruct-ledgers` | v2; `retained_adapter_replay_v1` |
| Pair default, no context or replay | v1; `exact_declared_repeatability_v1` |
| Pair context option(s) or `--require-context`, no replay | v2; `exact_declared_repeatability_with_context_v1` |
| Pair adapter replay, no context or sender | v3; `exact_declared_repeatability_with_reconstruction_v1` |
| Pair adapter replay plus context, no sender | v3; `exact_declared_repeatability_with_context_and_reconstruction_v1` |

Add `--reconstruct-sender` to either bundle or pair invocation to require fresh
fixture/offer/submission correlation. Bundle output becomes v3; pair output
becomes v4. Both replay flags can be selected together. The sender adds
`sender-reconstruction/sender/` and `sender-reconstruction/sender-reconstruction.json`
under the bundle (or each pair side's reconciled directory), with all four nested
inventory entries bound to the current sender snapshots. Sender and adapter
failures are recorded independently after the common base checks pass.

Original bundle v3 requires pair `--reconstruct-sender`, plus
`--reconstruct-ledgers` when its adapter requirement is true. Missing selections
are invalid evidence. See the [bundle](slo-bundle.md) and [pair](slo-compare.md)
guides for exact schemas, mode policies and unexecuted departmental cases.

Original bundle v2 requires explicit pair `--reconstruct-ledgers`; default and
context-only pair calls reject it. Replay mode accepts original v1/v2 or mixed
pairs. Each side retains original `bundle.json` and fresh `reconciled/` artifacts;
`comparison.json` records new replay/context references. Old nested reconstruction
files are not copied by pair review; preserve original bundles separately.
Original mismatch/incomplete/error results remain failures even if fresh checks
succeed. Current replay failures prevent that side from qualifying comparison
fields, metrics or context binding. Context requires original/fresh evidence and
exact bundle-observation binding. Null/null never qualifies as known equality;
unsupported and unavailable fields remain explicit. Equal supported declarations,
including false isolation claims, do not establish real-world adequacy.

## 6. Interpret outcomes, budget storage and recover

Retain stdout/stderr and process exit status alongside every receipt. Exit numbers
have tool-specific meanings; never use a universal `exit 0 = acceptance` rule.

| Tool | Exit semantics |
| --- | --- |
| Sender / adapter | 0 submission/adaptation completed; 2 input/operation failure; neither establishes a completed acceptance run |
| Reporter | 0 `review_required`; 1 `measurement_failed`; 2 input/publication error; 3 `inconclusive` |
| Context / default bundle | 0 `review_required`; 1 `mismatch`; 2 argument/I/O error; 3 `incomplete` |
| Standalone reconstruction / replay bundle | 0 `review_required`; 1 `mismatch`; 2 `error`; 3 `incomplete` |
| Pair | 0 `review_required`; 1 `conditions_differ`; 2 argument/I/O or recorded replay error; 3 `incomplete`; 4 `invalid_evidence` |

Reconstruction precedence is error > mismatch > incomplete > review-required.
Replay pair precedence is error > invalid-evidence > incomplete > conditions-differ
> review-required. All diagnostics remain significant, including simultaneous
context invalidity and replay errors. Reporter coverage gaps can take precedence
over numerical failures without erasing them. Component close flags, adapter
completion, context binding, replay completion and bundle completeness describe
different boundaries. None is SLO compliance, factual verification or comparability.

Default bundle retained-input budget is 64 GiB (`--max-bytes`). Pair has 64 GiB
per side (`--max-bytes-per-side`). Replay adds a separate 64 GiB per operation/side
(`--max-reconstruction-bytes`), plus rebuilt copies, metadata and SQLite scratch.
Sender replay adds a separate 64 GiB per operation/side
(`--max-sender-reconstruction-bytes`) for four input copies, excluding metadata.
Standalone reconstruction uses `--max-bytes`; its selected budget is passed to the nested collector. Context defaults to 256 MiB for
references (`--max-evidence-bytes`, pair `--max-context-evidence-bytes` per side),
with a 64 MiB per-reference ceiling. Receipt/declaration, observation/report and
JSONL limits remain in the detailed contracts. These are input limits, not total
disk quotas or demonstrated scale. Size all retained copies and transient indexes;
use an existing scratch directory. Do not raise limits to hide absent coverage.

On any failure, preserve originals, partial snapshots, receipts and process logs.
A late write/fsync/close error may leave a visible apparently complete receipt;
inspect both contents and exit status. Output directories are not atomic transactions.
Stop writers, diagnose the exact input/resource/binding failure, then retry into a
new destination. No command overwrites or automatically deletes retained evidence.
Missing-input modes can retain gaps but cannot manufacture a completed run. Keep
raw/private evidence local; only approved redacted conclusions belong in Git/Vault.

## 7. Departmental validation and acceptance handoff

Behavioral matrices remain **not run by the agent**. The department owns the cases
in [ingress](slo-ingress.md), [runtime](slo-runtime.md), [adapter](slo-collect.md),
[reporter](slo-report.md), [context](slo-context.md), [reconstruction](slo-reconstruct.md),
[bundle](slo-bundle.md) and [pair comparison](slo-compare.md). Cover missing/duplicate
identities, time/drain edges, original/current failures, all mode combinations,
physical boundaries, concurrent shutdown, partial publication, budgets and scale.
Store actual results separately from this unexecuted checklist.

| Acceptance requirement | Staging | Production |
| --- | --- | --- |
| KVM SUT allocation | 2 vCPU / 4 GiB | 8 vCPU / 16 GiB |
| Storage / NIC | 20 GB SSD / 1 Gbps virtual NIC | 100 GB dedicated SSD / 10 Gbps SR-IOV VF |
| Sustained / 60-second burst | 100 / 250 Mbps | 3 / 7 Gbps |
| Rules | 2,000 | 20,000 |
| Expected alerts | 15/min baseline, 120/min burst | 15/min baseline, 120/min burst |
| Live arrival to durable write | p99 <=180 ms, missing alerts included | Same |
| Sustained / burst packet loss | <=0.05% / <=0.2% | Same |
| Current qualification | No qualifying run; isolation/resources/clock/durability/load/sample evidence outstanding | Same; historical VM observation of about 7.70 GiB RAM is below the 16 GiB target |

The proposed 512-byte mean implies approximately 24,414/61,035 pps for staging
and 732,422/1,708,984 pps for production, excluding extra link overhead. Bandwidth
is primary; retain the actual distribution and byte convention. No current-host
resource discovery was performed for this runbook. Match actual profile resources
or obtain an explicit profile revision before qualifying execution; exploratory
results remain labeled deviations and do not block implementation work.

A completed handoff includes both the designated report and checksum-bound raw
companions, exact source/build/config/rules/fixture provenance, process outcomes,
resource/isolation and generator evidence, clock/durability validation, cohort
and expected-event reconciliation, all per-window counts/p99/violations, extended
sample justification, behavioral results, and a reviewer decision for the exact
profile with explicit failure/inconclusive reasons. Supply comparison evidence
under the agreed same-VM context or document a matched rebaseline. Summary JSON,
receipt equality or a clean compile cannot replace this evidence. R90-75 remains
outstanding until departmental review; this runbook changes no acceptance target
and authorizes no release, tag or registry publication.
