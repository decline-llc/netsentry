# R90-121: Context-aware supplied bundle comparison

## Selection and authority

Clean fetched baseline `11c82f8273635e60222948dcd2b7f9f2a019aed5` equals
HEAD/origin/main/FETCH_HEAD; R90-120 feature/closure Vault index/MOC are verified.
The recent contract-to-context chain remains delivered with testing delegated by
the user. Select dependency-ready R90-121 only. Sep 29–Dec 27 remains the current
horizon; R90-75 acceptance is departmental and R90-59 publication stays separately
blocked. No acceptance hardware or missing real artifacts block implementation.

## Bounded design and evidence map

Extend the existing comparison API/CLI with optional baseline/candidate context
packages plus an explicit require-context option. Any context input enables the
new context comparison policy/output schema; neither input and no requirement
preserves the existing v1 path. Retain each original context receipt, rerun the
R90-120 retention/validation against fixed source paths, and bind declaration,
observation and usable reference inventory to the original receipt. Then bind
run/profile/start and exact observation digest/bytes to the already reconciled
bundle for that side. Missing/partial/tampered input remains explicit.

Compare all 25 declared context values with states known/unknown/unsupported/
unavailable. Null never counts as a match. A known value requires nonempty
references whose original and fresh retained bytes agree; evidence IDs/digests
are provenance, not property-equality conditions. An incomplete but consistent
context may contribute individually supported fields, while its gaps remain.
Invalid context never qualifies its fields. Preserve invalid > incomplete >
different > review-required status precedence and all known differences; do not
upgrade failed/inconclusive measurements or certify facts/acceptance. Physical
truth, relevant evidence contents, full packet replay and clock/durability remain
review requirements even when all declarations match.

| Criterion | Planned evidence |
| --- | --- |
| Fresh context retention/binding | Source/AST review of strict original receipt schema, computed paths, byte/digest/run binding, original/fresh inventory proof and bounded inputs |
| Known/unknown comparison | Manual review of per-field eligibility, null/unsupported/unavailable handling, known differences and status precedence |
| Compatibility and handoff | Document v1 default/v2 context policy, new flags/limits/output paths, permanent qualification limits and departmental fault matrix |
| Delivery | Static AST/schema/JSON/roadmap/history/link/diff/token-pattern review; exact Git push/fetch/Vault ranges and stable prose replay |

## Scope and limits

Intended paths (11): scripts/slo_compare.py, new scripts/slo_context_compare.py,
scripts/slo_context.py (remove stale consumer limitation), Makefile syntax
registration, docs/slo-compare.md, docs/slo-context.md, docs/performance-slo.md,
roadmap, this plan, new context-compare task state and active R90-75 state.
No tests, CLI smoke, benchmark, acceptance/knowledge suites, traffic, resource
discovery, fixtures, measured artifacts, new dependencies, CI or release changes.
Stop only for new external/private authority or undocumented input; implement from
versioned schemas and explicit unknowns. No additional product SLO decision is
inferred; comparison policy remains conservative exact declared repeatability.

## Delivery and next boundary

Persist this plan/state before edits. Implement and statically review one increment,
verify feature push/fetch/Vault and stable prose replay, then one docs-only closure.
Queue R90-122 for bounded offline raw-ledger reconstruction against retained
observations, addressing the remaining derivation gap without live execution or
certification. Do not start it in this increment.

## Static review checkpoint

Implementation and manual source review are complete. `make python-check` parsed
source AST only; the new receipt field set matches R90-120's result dictionary.
All 25 context fields match the declaration documentation, all 138 task JSONs
parse, and 126 roadmap rows/Definitions match as multisets without duplicates.
History order, local Markdown links, diff formatting and token-pattern review
passed. Per-field eligibility, original/fresh proof, exact bundle binding,
invalid/incomplete precedence and conditional v1/v2 paths were manually reviewed.
No implementation was imported or executed; no runtime or acceptance evidence was
generated. User-delegated testing is the planned deviation from normal gates.
No other scope change or reusable skill change was needed. Feature Git/Vault
delivery and one documentation closure remain; R90-122 is not started.
