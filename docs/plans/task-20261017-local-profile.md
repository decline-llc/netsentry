# R90-194: versioned local-v1 evidence evaluation

Baseline: 9d95c1b2f232ceaa14af0afd0abf36382cc0e4ac, freshly fetched clean main.
R90-193 exact hosted run 39016262049 completed/success; local delivery record,
631-commit Vault index, iteration and current local-v1 notes establish delivery.
Four-week history moved from delegated bounded repairs to full local/hosted CI
and current-resource acceptance. Historical delegation is superseded by AGENTS.
The forward queue has 201 complete unique row/Definition pairs; R90-194 is the
highest dependency-ready increment. R90-75 stays deferred. No competing edits.

## Acceptance and evidence map

1. Introduce observation/manifest schema 2 with contract_id=local-v1 and an
   independent evaluator. Keep legacy schema 1 staging/prod constants, strict
   diagnostics and summaries unchanged. Evidence: frozen legacy golden summaries
   and rejection regressions, exact contract/profile/version tests.
2. Evaluate 100 rules, 600 seconds, 30-second drain, one expected event/second,
   100/500 pps minute schedule (bursts at 180 and 480), +/-5% actual minute load,
   <=1% loss in sustained and each isolated burst, 500ms nearest-rank p99 plus
   <=10ms established clock uncertainty. Evidence: direct boundary/failure tests.
   All oracle events remain, missing completion is infinity; never hide numerical
   failures behind an inconclusive status. Record extra/duplicate/late lifecycle
   identities and capture/submission errors. No passing verdict on incomplete
   evidence, unsuitable allocations, bad clock/load or oracle coverage.
3. Adapt collect/reconstruction, context/binding, bundle/recomputation and pair
   comparison to explicit contract identity. Preserve raw input and anomalies;
   never inflate completed counts with duplicates. Reject mismatched contracts
   and cross-profile pairing. Evidence: public API/CLI retained-ledger replay,
   bundle/context/comparison tests, damaged identity and incomplete controls.
4. Separate measurement_failed, inconclusive, review_required and explicitly
   reviewed_local_acceptance for a single supplied run. Review is a hash-bound
   local declaration of all required evidence checks, never automated L1/L2 or
   production certification. Evidence: invalid/stale review and CLI exit tests.
5. Integrate regressions in the full CI gate, document schema/CLI boundaries and
   refresh active authority. Preflight exact tools, module root and Docker
   frontend/base manifests; focused tests, knowledge-check, full make ci-check,
   static/JSON/roadmap multisets/links/diff/sensitive review. One non-force batch
   commit/push; fresh exact refs, exact hosted success, Vault range/index/MOC,
   stable knowledge and idempotent replay. Post-push facts local only.

## Scope, risks and stop conditions

Own SLO Python tools, their direct regression/fixtures, CI test entry, AGENTS
current tooling-status sentence, README/development and SLO docs,
current status/roadmap and this plan/state. No detector/capture/storage runtime
change, live traffic, namespace fixture, L1/L2, new dependencies, tags or release.
R90-195 is next, including the existing 1000ms capture-buffer risk; not started.
Schema confusion, omitted failure identities, silent legacy drift and false
review authority are principal risks. Stop upload on required check failure or
ambiguous evidence; diagnose and rerun, never weaken checks. Review cannot turn
failed/inconclusive measurements into acceptance. Keep raw data and machine
paths outside Git/Vault. Exactly one increment in this trigger.


## Implementation checkpoint (2026-10-18)

All behavior criteria are implemented; focused boundary and retained-evidence
regressions, syntax/docs and 33 knowledge tests passed. Golden legacy summaries
were generated with the unchanged baseline source, not the new evaluator.
Local helper source digest is included in new summaries and comparison conditions.
Single-run review re-derives its summary from exact raw bytes before acceptance.
Legacy envelope review fields are preserved as compatibility metadata; active
local reports and AGENTS require local review, not an external department.
Initial fixture and overlapping numeric/validity assertions were corrected;
no gate was weakened. Complete CI and exact delivery remain required.

The first full gate failed before runtime checks on a raw-GitHub TLS handshake
timeout for the third pinned asset. Direct identical-URL byte/hash verification
succeeded; the unchanged full gate was restarted. This external transport
failure is retained, not reclassified as a code failure or bypassed. The new
Oct18 environment date rolls only the unfinished horizon forward one day.

Reusable skill refinement: when an evidence entrypoint delegates evaluation to
another source module, include that module's digest in provenance and pair
conditions. The local skill's Markdown/front matter passed structural review;
its change is separate from the NetSentry feature commit.

Final compatibility review found legacy context kind/profile diagnostic order
had shifted through the shared identity helper. Restore the original order and
add a direct both-invalid and profile-only regression. The first successful
complete retry (82.3% coverage including Docker health) predates that repair;
its log is retained but is not the final source gate. Focused/knowledge checks
passed again; a clean complete ci-check is rerunning on the final source.


## Acceptance comparison and local completion

Every planned tooling boundary maps to a direct executed method in
`docs/evidence/r90-194-local-profile.md`. The final full ci-check passed, including
all supply-chain fetches, no reachable vulnerabilities, uncached full race and
82.3% coverage, 12 SLO/42 evidence/33 knowledge methods, C tests/ASan fuzz, E2E,
archive and Docker content/health. Final implementation hashes remain unchanged.
Only evidence/status prose follows the gate and receives affected static and
knowledge checks. Retained sparse-ledger integrations prove derivation, not
load; single-run review declarations do not prove measured L1/L2. Required
scope stayed within R90-194. External TLS retry and restored legacy diagnostic
precedence are fully recorded; no unresolved numeric/tooling ambiguity remains.

One batch delivery remains: resolve its exact SHA through Git, fresh fetch,
hosted completed/success, exact Vault range/index/MOC, stable-topic reconciliation
and replay. Persist these facts locally with no extra closure push. Once verified,
R90-195 is next; no fixture or live acceptance work starts in this trigger.
