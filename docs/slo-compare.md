# Declared comparability of SLO bundles

R90-119 adds `scripts/slo_compare.py`, a standard-library API/CLI comparing two
supplied [R90-118 evidence bundles](slo-bundle.md). It retains both inputs,
reconciles their companion files again, and diagnoses differences in declared
conditions. This implementation has received static review only. Tests, actual
measurements and acceptance remain assigned to the specialist department.

## Department invocation

Stop all input writers first. The following is an unexecuted template:

```bash
python3 scripts/slo_compare.py \
  --baseline baseline-bundle \
  --candidate candidate-bundle \
  --output-dir comparison-review
```

API: `compare(output, *, baseline=None, candidate=None,
max_bytes_per_side=64*1024**3)`, using `pathlib.Path` directory arguments.
Either input can be omitted to record missing evidence. Output must be new,
with an existing parent. Existing outputs are never overwritten or deleted.
The CLI does not send traffic, start services, execute benchmarks or tests,
or discover actual machine resources. It consumes already supplied files only.

## Retention and trust boundary

Each side retains its original `bundle.json` and creates `reconciled/` containing
fresh copies of the thirteen fixed R90-118 companion files and a freshly computed
bundle manifest. `comparison.json` is published last and records the original
manifest bytes/digest, the new manifest digest, input checks, per-side metrics,
field comparisons, differences and qualification gaps. All references to retained
files are relative to the new output directory; source paths are not published.

Original manifests require the exact current v1 schema, coherent status/completion
flags, bounded known inventory paths, unique entries, hashes/byte/row counts and
valid run/origin declarations. R90-118 reconciliation reads only fixed companion
basenames, checks receipts and raw source identity, and recomputes the report from
its retained observations. A side qualifies for field comparison only when its
original manifest and current reconciliation both say `review_required`, and their
complete inventories, run/origin identities and reported measurement status agree.
A claimed successful old manifest cannot bypass current source checks. Missing,
partial or malformed inputs remain visible, with retained raw copies where readable.
Original mismatch/incomplete statuses cannot be upgraded by the comparison.

Source hashes identify supplied bytes; they do not authenticate them. The checker
does not replay the packet join or verify that the fixtures/rules, clocks, physical
packet arrival, storage durability or process exits reflect reality. Original tool
source digests are compared between sides, not asserted to be authenticated builds.
Current reconciliation uses the installed implementation; changed interpretation
can reject an old report. Freeze tools and retain the source versions for review.

## Conservative comparison policy

`comparison_policy` is `exact_declared_repeatability_v1`: same-commit repeatability
under exactly matching available declarations. There are no implicit tolerances,
waivers, regression thresholds or statistical conclusions. Commit/config/tool
changes are reported differences. A revision experiment needs a separate explicit
policy before those differences can be treated as an intentional variable.

| Category | Equality conditions |
| --- | --- |
| Profile/window | `profile`, run duration, drain duration, minimum duration/sample policy |
| Resources | Declared SUT vCPUs, memory bytes and active rule count |
| Measurement | Live-arrival/durable flags, clock method and uncertainty |
| Provenance | SUT/harness commit, config/rules/fixture hashes |
| Workload | Exact minute phase/offered packets/offered bytes/completeness, expected event/packet/rule identities, expected alert counts per minute and full run |
| Ingress/export | Sender link and byte boundary; capture timestamp type/precision/direction/port; engine clock, durable and event identity boundary strings |
| Tooling | Original sender, adapter, bundle and reporter source-code digests |

Workload rows are summarized with SHA-256 to keep the comparison artifact small.
Hash each canonical JSON row (sorted keys, compact separators, ASCII escaping,
no nonfinite values), followed by LF, in sequence:

- Offered cohorts: objects with `start_second`, `phase`, `offered_eligible`,
  `offered_bytes`, `observation_complete`, in existing minute order.
- Oracle identity: arrays `[event_id, packet_id, rule_id]`, lexicographically sorted.
- Expected alert counts: one integer per minute, including zero minutes.

Full data remain in each side's `reconciled/adapter/observations.json`. Event
arrival/durable times and fully-processed results are outcomes and do not enter
condition equality. Sub-minute offered timing jitter is not compared; fixture
identity, actual offered cohorts and expected alert cohorts are compared. Even a
small offered-count difference is exposed, rather than silently tolerated. Sender
link or run-marker length changes can also change the workload and byte accounting.

Run IDs must differ. UTC starts identify runs and are not equality conditions.
Intervals from start through declared observation/drain end are treated as half-open;
overlap is a pair qualification gap in the approved shared-VM execution context.
Identical run IDs also create a gap. A copied bundle therefore cannot establish
independent runs simply by being supplied twice.

## Outcomes and metrics

| Status / exit | Meaning |
| --- | --- |
| `review_required` / 0 | Two bound, reconciled inputs, distinct nonoverlapping runs and all available equality conditions match; review still required |
| `conditions_differ` / 1 | Complete inputs, but at least one compared declaration differs |
| error / 2 | Invalid CLI/API argument or I/O failure; inspect retained partial output and retry to a new directory |
| `incomplete` / 3 | Missing/partial evidence or pair qualification gaps prevent a complete declared comparison |
| `invalid_evidence` / 4 | Invalid original manifest, recorded mismatch, failed current reconciliation or original/source binding mismatch |

Invalid evidence takes precedence over incomplete evidence, which takes precedence
over condition differences. All diagnosed differences/gaps remain in the artifact.
`declared_conditions_match` is null until both sides qualify for field comparison;
otherwise it reflects field equality only, even if run identity/overlap gaps exist.
It never means actual comparability has been established.

Each qualifying side exposes its original run identity, observed boundary,
recomputed measurement status and full-run metrics: offered/fully-processed/lost
packets, expected/observed/missing alerts, deadline violations and p99 bounds.
The exact reporter fields and missing-event semantics are retained. Minute and
rolling-window detail stays in the copied report. A side whose report says
`measurement_failed` or `inconclusive` retains that status even if every comparison
condition matches; original measurement outcomes are never upgraded.

Every comparison sets `comparability_established`, `slo_compliance_asserted` and
`regression_asserted` to false and requires departmental review. Always-present
qualification gaps identify absent CPU/pinning, storage/NIC, VM allocation,
isolation and toolchain evidence plus unverified traffic/oracle, clock/durability
and process outcomes. Equality of the available fields cannot fill those gaps.
No speed ratio, significance estimate or automatic performance policy is emitted.

## Bounds and persistence

Each side defaults to a 64 GiB source-copy budget, configurable through
`--max-bytes-per-side`; its original manifest has a separate 1 MiB limit. A pair
therefore needs space for up to twice the budget plus 2 MiB and generated outputs.
R90-118 per-file/JSONL bounds and regular-file restrictions apply. Missing or
budget-truncated inputs produce explicit gaps; an invalid retained prefix can
also produce a mismatch during reconciliation. Source modification is detected
through snapshot metadata and manifest/digest binding, without an adversarial
filesystem guarantee. Use quiescent trusted inputs.

Directories are private and original snapshots use R90-118 mode 0600. Large raw
ledgers stream through R90-118; observations and oracle sorting/report recomputation
use memory proportional to the bounded observation JSON. Counts/hashes keep the
pair summary small. Actual time, memory and disk overhead remain unmeasured.
Per-side directories are synced through reconciliation; comparison publication is
exclusive and synced, followed by output-parent sync. This is not an atomic whole-
directory transaction. Interruptions and late write/fsync/close failures can leave
partial output or a visible manifest; inspect process status, preserve evidence
and use a fresh output directory for retries.

## Required departmental validation — not executed

| Boundary | Required coverage |
| --- | --- |
| Inputs | Both complete, every missing file, partial/malformed manifests, schema/type/unknown-field errors, duplicate inventory names, false completion, changed bytes/digests/rows, mismatched old/new identities and statuses |
| Pair | Same bundle twice, identical run IDs with different files, overlapping/disjoint/exact-touching intervals, missing side, mixed invalid/missing/differing evidence and status precedence |
| Conditions | Each field/category mismatch, same and changed commit/tool versions, phase/count/byte/oracle drift, zero-alert minutes, row-order normalization, outcome changes excluded from conditions |
| Metrics | Failed/inconclusive reports remain so, raw alert/loss/deadline counts and missing-as-unbounded p99, equal conditions with different outcomes, no ratio/gate/compliance upgrade |
| Retention/scale | Original and current hashes, no-overwrite/private modes, source mutation/symlinks, limits/large ledgers, disk/read/fsync/close faults and interruptions, retained partial outcomes and replayable reports |

No real or synthetic comparison artifact was generated by the agent. Both staging
and production acceptance remain outstanding under the [SLO contract](performance-slo.md).

## Supplied environment context

R90-120 provides a [run-context declaration and retention tool](slo-context.md)
for explicit hardware/toolchain/isolation values, null unknowns and checksum-bound
opaque references. It binds the declaration to exact observation bytes. This
comparator does not yet consume the package; its environment qualification gaps
remain. R90-121 will add consumer wiring with fresh source/run binding; declaration
completeness or equality must never imply verified environment facts.
