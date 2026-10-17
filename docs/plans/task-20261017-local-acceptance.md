# R90-193: scope acceptance to existing single-VM resources

User authority (2026-10-17): reduce acceptance requirements to the hardware
currently available, adjust repository rules, and plan subsequent work. This
supersedes hardware/scale requirements as current delivery blockers; it does not
authorize fabricated test results, weakened correctness/security CI, or a release.

Baseline: 28c23b54abde8194ed899734518ba03c4079cafe, fresh fetched main.
Prior-turn local edits: AGENTS, both READMEs, active roadmap and dated status.
These are assistant-authored documentation for the same ongoing user request;
retain their delivery policy and integrate them into this complete policy change.
Recent history: many bounded repairs and delegated execution; Oct17 full CI repair
verified local/hosted success. R90-75 remains unmeasured, and its large profiles
are inappropriate as current readiness blockers on a ~7.7 GiB shared VM.

Acceptance map:
1. Establish active local-v1 scope, explicit numbers, validity requirements and
   staged L0/L1/L2 completion. Evidence: acceptance-local-v1 document and AGENTS.
2. Defer historical staging/prod targets without claiming completion or silently
   lowering legacy evaluator values. Evidence: preserved historical contract,
   active banners, R90-75 row/state and unchanged script hashes.
3. Produce a dependency-ready 90-day queue: each item has status, dependencies,
   window, risk, measurable acceptance, required validation and stop condition.
   Evidence: one rolling roadmap with R90-193..197 rows/Definitions; no second queue.
4. Reconcile current README/status/runbook/development/performance/release prose,
   preserving old evidence and immutable notes. Old code still supports only
   staging/prod and 180ms; profile support is the first planned implementation.
5. Run docs, links/JSON/roadmap structure, knowledge and full ci-check before a
   single completed delivery; verify exact hosted SHA, sync exact Vault range,
   and retain post-push facts locally without another closure push.

Non-goals: no runtime/scanner/detection-rule/schema/Makefile change; no simulated
network installation or live acceptance traffic; no tags, releases, registry push,
new hardware, external test department or production compliance claim.
Risk: documentation could imply a lowered profile is already executable/passed;
make support status and evidence gaps explicit. Freeze thresholds before runs;
failed targets require an explicit new contract revision, never after-the-fact edits.
Stop upload on required validation failure, ambiguous remote state or scope drift.

## Local acceptance comparison

All policy/queue criteria are implemented across fifteen documentation paths.
The full local ci-check passed (82.2% coverage, all-package uncached race,
42 evidence and 33 knowledge tests, both ASan fuzz gates, E2E, archive and full
Docker content/health). 201 complete unique row/Definition pairs and prior
completed rows/Definitions were checked. Legacy runtime/report/profile and
Makefile hashes are unchanged. Only final evidence/status prose changed after
that gate and receives affected static/knowledge checks before commit.

R90-194 remains the next unstarted implementation. No local-v1 L1/L2 run has
executed; no production acceptance is asserted. Final exact-SHA hosted/Vault
facts are recorded locally after the one completed batch push, avoiding another
closure CI run. The skill refinement records evaluator/schema inspection before
claiming a revised contract executable; no other unplanned scope was introduced.
