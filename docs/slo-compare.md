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
max_bytes_per_side=64*1024**3, baseline_context=None, candidate_context=None,
require_context=False, max_context_evidence_bytes=256*1024**2,
reconstruct_ledgers=False, max_reconstruction_bytes=64*1024**3, scratch_dir=None,
reconstruct_sender=False, max_sender_reconstruction_bytes=64*1024**3)`, using
`pathlib.Path` directory arguments.
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

Default mode requires the exact v1 bundle schema; explicit reconstruction mode
also accepts v2; sender mode accepts v3 as described below. All require coherent status/completion
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
does not replay the packet join in default mode. No mode verifies that fixtures/rules, clocks, physical
packet arrival, storage durability or process exits reflect reality. Original tool
source digests are compared between sides, not asserted to be authenticated builds.
Current reconciliation uses the installed implementation; changed interpretation
can reject an old report. Freeze tools and retain the source versions for review.

## Conservative comparison policy

Without context or reconstruction options, output schema v1 uses `comparison_policy`
`exact_declared_repeatability_v1`: same-commit repeatability
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
qualification gaps identify absent or unverified CPU/pinning, storage/NIC, VM allocation,
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

## Supplied environment context (R90-121)

An unexecuted department invocation with retained R90-120 packages:

```bash
python3 scripts/slo_compare.py \
  --baseline baseline-bundle --candidate candidate-bundle \
  --baseline-context baseline-context --candidate-context candidate-context \
  --require-context --output-dir context-comparison-review
```

Either context argument or `--require-context` enables context mode. When neither
argument is supplied and context is not required, the existing v1 policy/output
path remains when reconstruction is disabled. Context-only mode uses output `schema_version: 2` and policy
`exact_declared_repeatability_with_context_v1`; a missing side is an explicit gap.
`--require-context` alone therefore cannot produce a completed declared comparison.

Each side additionally retains its original `context/context.json` and creates
`context/reconciled/` using [the context retention API](slo-context.md), including
the raw declaration, observations, opaque references and fresh receipt. The original
receipt is strictly schema-checked: flags, identities, 25 field cells, known/unique
inventory names, finite sizes/budgets, hash syntax and diagnostic/status coherence.
Original and fresh run identities and declared values/references must agree. Every
file complete in both inventories must have identical byte counts/digests; differing
complete files invalidate the context. Missing or partial files cannot prove a field.
Receipt paths never select inputs: fixed context filenames and validated declaration
IDs determine copies. An invalid declaration cannot select reference files.

Declaration and observation files must be complete and bound to the original
receipt. Then run/profile/start, observation SHA-256 and byte count must match the
already reconciled bundle for that side. A context from another run or observation
file cannot supply comparison conditions. No fields from an invalid context or
unqualified bundle become eligible, even if a supplied receipt claims success.
The original and current context source receipts remain retained for review.

All 25 context fields appear in `context_comparisons`. Each side of a row contains
its value, supporting evidence IDs and one of these states:

| State | Meaning |
| --- | --- |
| `known` | Non-null declaration with nonempty references, every referenced file proven complete and identical in original/fresh inventories, and context bound to this bundle |
| `unknown` | Explicit null; even two nulls never count as a match |
| `unsupported` | Non-null value lacks references or one of its references lacks original/fresh byte proof |
| `unavailable` | Context receipt, declaration, observation or bundle binding could not qualify the field |

Rows have `equal: null` unless both sides are known; otherwise exact value equality
is used. Evidence IDs/digests are provenance and are not equality conditions for
machine properties. Different reference files can support equal declarations;
identical reference files still cannot authenticate those declarations. False is
a known boolean value when supported; equality of false isolation claims does
not establish isolation adequacy.

Consistent incomplete contexts may contribute individually supported fields; their
original/current gaps remain and cannot be upgraded away. Known differences remain
in `differing_fields` with the `context.` prefix even if another field is unknown.
Existing invalid > incomplete > differing > review-required precedence applies.
A partial context mode never silently falls back to the v1 success path.

Schema v2 adds `sides.<side>.context` receipt references and `bound_to_bundle`,
`context_comparisons`, `context_conditions_match`, `facts_verified: false`, the
context evidence budget and consumer source-code digest. `context_conditions_match`
is false if a known difference exists, null if any fields remain unqualified and
none differ, otherwise true. Overall `declared_conditions_match` follows the same
rule across bundle and context conditions, remaining null when bundle comparison
is unavailable. A true value only describes available declaration equality;
status can still be incomplete because of original receipt or run-overlap gaps.
Measurement statuses/metrics retain their original meaning and are never upgraded.

`--max-context-evidence-bytes` defaults to 256 MiB per side for opaque references.
Each side additionally needs at most 1 MiB for its original context receipt,
1 MiB for its declaration and 64 MiB for observations, plus generated metadata.
R90-120's 64 MiB per-reference limit remains. These allocations are additional to
bundle budgets. Original/fresh partial outputs stay retained, with no overwrite
or automatic deletion. Context, side, comparison and parent directory entries are
synced in order before reporting success; a late failure may leave visible output.

Field presence does not remove factual qualification requirements. Evidence
relevance, machine state, compiler/build provenance, physical clock/durability,
real offered workload, isolation and raw packet derivation remain unverified.
No machine discovery, command execution, live traffic, tests or acceptance occurs.
All comparability/regression/SLO assertions remain false in context mode too.

### Additional departmental validation — not executed

Cover v1 invocation/schema compatibility; absent/one/two context packages and
explicit requirement; each receipt schema/type/flag/inventory condition; unknown,
unsupported and unavailable fields; null/null never equal; false/false with
references; individually supported fields in incomplete packages; old versus
fresh digest/value/identity drift; swapped run/profile/start/observation sources;
known allocation mismatch; partial budgets; original missing and newly supplied
reference files; different IDs/files with equal values; simultaneous known
mismatch and unknown field; preserved failed/inconclusive metrics; no context
success fallback; modes, no-overwrite, source changes, read/write/fsync/close/
interrupt failures and retained partial manifests. No such case was executed by
the agent; static implementation delivery does not establish runtime correctness.

## Fresh reconstruction in pair review (R90-123)

Unexecuted invocation template:

```bash
python3 scripts/slo_compare.py --baseline baseline-bundle --candidate candidate-bundle \
  --reconstruct-ledgers --scratch-dir scratch --output-dir reconstructed-pair
```

`--reconstruct-ledgers` requires fresh reconstruction for both sides through
[the v2 bundle mode](slo-bundle.md). Original v1 or v2 bundles can be supplied,
including one of each. Original v2 bundles require this explicit flag; supplying
them in default or context-only mode is invalid evidence, with no silent fallback.
No supplied reconstruction path or old nested receipt selects the new replay.

The original bundle manifest is retained and strictly checked. V2 adds exact
policy, summary field/type/path/hash checks, error-list/status coherence, and
complete/source-bound/matching replay requirements for a successful original
receipt. Independently, the current reconciler copies the fixed companion files,
finishes ordinary checks and replays its own retained adapter inputs. All five
nested source inventories bind to the enclosing bundle; the original bundle's
shared inventory/run/report fields must also match the fresh reconciliation.
The old nested reconstruction files are not copied or authenticated; retain them
with the original source bundle for historical review. An old summary is never
proof of fresh reconstruction. Different old/new replay tool digests remain
reviewable provenance, not authenticated builds or automatic acceptance.

Original mismatch/incomplete/error statuses remain failures even if fresh replay
succeeds. A current replay failure prevents that side from supplying condition
fields, metrics or a qualified identity for context binding. A valid side retains
its failed/inconclusive measurement status; replay agreement never upgrades it.
Known comparisons already eligible under the existing rules remain diagnostic.

Adapter replay without sender selection uses `schema_version: 3` and one of two policies:

| Context enabled | Policy |
| --- | --- |
| No | `exact_declared_repeatability_with_reconstruction_v1` |
| Yes | `exact_declared_repeatability_with_context_and_reconstruction_v1` |

V3 adds `reconstruction_required: true`, `errors`,
`max_reconstruction_bytes_per_side` and `sides.<side>.reconstruction`. Each summary
points to a fresh receipt relative to the pair output, at
`<side>/reconciled/reconstruction/reconstruction.json` when available. Original
receipt and new bundle references remain retained. V1/v2 calls without replay
keep their previous output structure; source-code digests naturally change.

`error` / exit 2 now also represents recorded original/current replay operation
errors. Precedence is `error > invalid_evidence > incomplete > conditions_differ > review_required`,
with other diagnostic lists retained. Invalid original v2
schemas remain invalid evidence, not trusted operation-error declarations.
Physical qualification, comparability/regression/SLO flags and context unknown
semantics retain their prior boundaries.

`--max-reconstruction-bytes` defaults to 64 GiB per side, additional to bundle
and context budgets. Each nested reconstruction also needs rebuilt copies and
SQLite scratch space. Sides replay sequentially using `--scratch-dir` (or system
temporary storage); original and nested outputs remain retained. Source/operation/
late fsync failures require preserving partial outputs and checking process exit.

### Additional departmental validation — not executed

Cover original v1/v2 and mixed pairs; v2 without the explicit flag; all context/
replay combinations; every v2 root/summary type/policy/path/flag/error constraint;
old receipt claims versus changed raw sources; fresh observation mismatch; each
original/current incomplete/mismatch/error outcome; source binding and output-path
references; budget/scratch/disk failures; context eligibility after failed replay;
unchanged failed/inconclusive metrics and permanent qualification flags. Verify
v1/v2 default output compatibility and retained diagnostics in v3. In combined
mode, context errors must remain invalid evidence and must neither overwrite nor
be promoted into reconstruction operation errors; verify both error classes together.

## Fresh sender reconstruction in pair review (R90-127)

Unexecuted handoff template:

```bash
python3 scripts/slo_compare.py --baseline baseline-bundle --candidate candidate-bundle \
  --reconstruct-sender --reconstruct-ledgers --output-dir sender-replayed-pair
```

`--reconstruct-sender` requires both sides to run fresh sender reconstruction
through [bundle schema v3](slo-bundle.md). Adapter replay remains an independent
`--reconstruct-ledgers` selection; context selection also remains independent.
Default and adapter-only modes retain their prior output structures.

| Supplied original bundle | Required selections |
| --- | --- |
| v1 | None; either or both fresh replays may be requested |
| v2 | `--reconstruct-ledgers`; sender replay is optional |
| v3, adapter requirement false | `--reconstruct-sender`; adapter replay is optional |
| v3, adapter requirement true | Both replay flags |

Missing a required flag yields invalid evidence. Original v3 must have the exact
sender policy/fields and boolean adapter requirement; the latter controls whether
adapter fields must be present. Replay summaries validate types, completion,
nullable matching values, fixed manifest paths, hashes and diagnostic coherence.
A nonempty shared errors list must correspond to an error in at least one selected
operation; it does not imply both operations failed. No source path in an old
receipt chooses fresh inputs, and no old summary qualifies a side.

The current bundle is rebuilt from its fixed companion files and runs the selected
replays against its own snapshots. Sender replay binds all four nested inventory
entries. Original and fresh complete bundle identities/inventories must then bind
before the side contributes comparison conditions, metrics or context identity.
Original incomplete/mismatch/error statuses remain failures after fresh success;
invalid original schemas remain invalid evidence. The valid side's failed or
inconclusive measurement metrics retain their meaning.

Sender-enabled output is schema v4. Its policy is
`exact_declared_repeatability_with_` followed by optional `context_and_`, optional
`reconstruction_and_` (adapter replay), and `sender_reconstruction_v1`, in that
order. V4 includes `sender_reconstruction_required: true`,
`adapter_reconstruction_required`, `errors`,
`max_sender_reconstruction_bytes_per_side`, and
`sides.<side>.sender_reconstruction`. Existing adapter/context fields remain
when selected. Sender manifest references point to
`<side>/reconciled/sender-reconstruction/sender-reconstruction.json` when available.
Original nested replay files must remain with their original bundle; they are
not copied or trusted by the current replay.

The default sender budget is another 64 GiB per side, independent of bundle,
adapter and context budgets, and excludes generated metadata. Source, operation,
publication and interruption failures require retaining all partial output and
checking process exit. Status precedence remains error > invalid_evidence >
incomplete > conditions_differ > review_required. Physical facts, comparability,
regression and SLO assertions remain false; sender replay certifies none of them.

### Additional departmental validation — not executed

Cover mixed v1/v2/v3 inputs and every flag/context combination; exact v3 root,
policy, summary, nullable/type/path and error coherence; original/current failures
and asymmetry; independent fresh sender/adapter inventory drift; old receipt
substitution; context eligibility after sender failure; unchanged measurement
outcomes; separate budgets, I/O/close/fsync/interruption and partial manifests.
The [implementation plan](plans/task-20260930-slo-sender-integration.md) records
static acceptance mapping. No tests, CLI execution or knowledge suite ran.
