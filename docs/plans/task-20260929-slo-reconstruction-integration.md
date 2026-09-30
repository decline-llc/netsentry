# R90-123: Fresh reconstruction in bundle and pair review

## Selection and authority

Clean fetched baseline `e8b0734cf45b7c75ec8f1e19658346cb2d962fdd` equals
HEAD/origin/main/FETCH_HEAD. R90-122 feature/closure Vault note/index/MOC and
current stable prose are verified. September contract-through-reconstruction
history is delivered with tests delegated; no missing delivery changes priority.
R90-123 is the sole ready local increment. Sep 29–Dec 27 horizon remains current;
R90-75 acceptance and R90-59 publication retain separate boundaries.

## Design and evidence map

Add explicit reconstruct_ledgers=False, max_reconstruction_bytes=64*1024**3 and
scratch_dir=None options to bundle/pair APIs and corresponding CLI flags. Default
bundle schema v1 and pair v1/v2 paths remain. Enabled bundle output is schema v2,
policy retained_adapter_replay_v1; enabled pair output is schema v3 with distinct
policies for optional environment context. Fresh bundle reconciliation must finish
its existing checks before invoking R90-122 on its own retained adapter files.
Rebind all five newly retained adapter inputs to the enclosing bundle inventory.
Retain nested reconstruction output and summarize completion, match and binding.
Missing/partial inputs never qualify a replay; returned diagnostics and local
operation errors remain explicit. Publish enclosing receipts after nested output.

The pair accepts original v1 or v2 bundles in reconstruction mode and performs
fresh replay for each side. Original v2 requires this explicit mode; old calls
reject it rather than silently downgrading its contract. Strict v2 receipt checks
cover policy, summary, errors and coherent completion/status. Bind shared original
inventory/run/report fields to fresh reconciliation; original mismatch/incomplete/
error outcomes never upgrade. A prior reconstruction receipt is never opened as
proof, and neither its path nor source directory selects new inputs. Fresh replay
failures prevent a side from qualifying for conditions or context binding.
Status precedence adds error above invalid/mismatch, incomplete and differences.

| Criterion | Planned evidence |
| --- | --- |
| Fresh source derivation | Static call/data-flow review of retained sources, five-file digest/row/run binding and no old receipt trust |
| Compatibility | AST/manual review of opt-in flags, lazy import boundary, v1/v2 original schemas and v1/v2/v3 pair outputs |
| Error/qualification | Manual review of incomplete/mismatch/error propagation, original failures, context eligibility and permanent authority flags |
| Department handoff | CLI/API/output/budget docs and unexecuted fault matrix; source/AST/schema/JSON/roadmap/link/diff review |
| Delivery | Exact Git push/fetch/Vault feature range and one documentation closure; stable prose replay |

## Scope and non-goals

Twelve intended paths: scripts/slo_bundle.py, scripts/slo_compare.py,
scripts/slo_reconstruct.py, docs/slo-bundle.md, docs/slo-compare.md,
docs/slo-reconstruct.md, docs/slo-collect.md, docs/performance-slo.md, roadmap,
this plan, new integration task state and active R90-75 task state.
No adapter/reporter algorithm, dependency, live execution, SLO threshold, CI or
release changes. No tests, CLI smoke, fixtures, synthetic evidence, resource
probes, traffic, benchmark/acceptance/knowledge suites or business-logic execution.
Tests remain user-delegated; static delivery cannot prove runtime correctness,
physical acquisition, capacity or compliance. Stop only for new private/external
authority or undocumented formats. Input/resource errors remain diagnostic.

## Next boundary

Finish this implementation plus one docs-only closure. Queue R90-124 for a
consolidated departmental execution and artifact-review runbook spanning the
implemented chain, with explicit proposed targets and outstanding validation;
do not execute it or start that increment here.

## Static review checkpoint

The twelve-path implementation is complete. Manual source review covered lazy
imports, five-file binding, v1/v2 original validation, conditional v1/v2/v3 output,
base-check gating, partial/resource failure propagation, context qualification and
relative receipt paths. It caught and corrected context diagnostic variable reuse
that could overwrite replay operation errors; both classes now remain separate.
The department matrix explicitly includes their combined behavior.
`make python-check` parsed AST only. Producer/consumer receipt and summary field
sets agree; all 140 task JSONs parse; 128 roadmap rows/Definitions match without
duplicates. History, links, intended scope, diff and token-pattern review pass.
No implementation was imported/executed and no fixtures or measurement artifacts
were generated. User-delegated behavioral tests remain the planned deviation;
no other scope or skill change was needed. Git/Vault feature delivery and one
docs-only closure remain. R90-124 is not started.

## Verified implementation delivery

Feature `820de2527b97e9467975b36bf693e05b6cd99a62` contains exactly the twelve planned paths.
Push without force/tags and a fresh fetch established clean
HEAD/origin/main/FETCH_HEAD equality. Exact range
`e8b0734cf45b7c75ec8f1e19658346cb2d962fdd..820de2527b97e9467975b36bf693e05b6cd99a62`
was synchronized to the sole local Vault; iteration note, full index and MOC
were verified. Nine current stable notes were reconciled while preserving all
immutable iteration bytes. Identical-range replay preserved Markdown SHA-256
`30d600ea80fbd9254dcb2a5a0c51fe9c34c22fc9b28f28401a50b83bcedd542d`.
This single docs-only closure records those facts and receives its own verified
push/fetch/Vault range. R90-124 is ready and unstarted. Execution tests and
acceptance remain delegated; no qualifying measurements or SLO claim exist.
