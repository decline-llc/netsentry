# Task Plan: R90-91 post-probe delivery audit

## Metadata

- Timestamp: 2026-08-12T07:30:36-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `c0b1eb2dae8dd90eda745eacc87b0a6ece01a450`

## Goal

Reconcile the completed R90-90 feature and docs-only closure against current
Git, task-state, fetched remote, Vault, recent delivery history, and receiver
pathname-cleanup authority; restore at most one directly evidenced local
follow-on without changing runtime or tests.

## Scope

- Verify the exact R90-90 feature/closure parent chain, intended paths,
  completed plan/state, fetched remote, and both Vault records.
- Review the Jul 20 through Aug 12 delivery phases and retain only material
  trends, validation deviations, stale authority, missing records, and risks.
- Trace receiver-owned socket capture and shutdown cleanup through current
  source, direct tests, public documentation, and prior pathname guarantees.
- Separate pathname-generation identity from protocol, configuration, public
  API, peer policy, and broader receiver lifecycle changes.
- Define at most one bounded follow-on, then refresh the roadmap and task state
  without starting it.

## Non-Goals

- Do not change receiver source/tests, configuration, protocol, public API,
  workflows, release gates, benchmark/evidence artifacts, or immutable Vault
  iteration notes.
- Do not implement a stronger shutdown identity check, add a production seam,
  or change startup stale/active classification.
- Do not authorize comparable-environment performance scope, remote tag push,
  GitHub Release, GHCR publication, workflow dispatch, or private data access.

## Risks

- Device/inode reuse is a filesystem race boundary, not proof of an observed
  production failure; the audit must keep the claim evidence-based.
- Strengthening cleanup must fail closed when a generation signal is missing
  and must not turn shutdown into replacement-path deletion.
- A replacement regression that creates a new inode without proving immediate
  reuse would not directly reach the promised boundary.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-90 is exactly recoverable | Feature/closure SHAs and parents, intended paths, completed plan/state, fetched `FETCH_HEAD == HEAD == origin/main`, and two exact Vault notes/index/MOC links |
| Recent history has no concealed blocker | Four dated phase counts plus material validation, stale-authority, missing-record, and unresolved-risk review |
| Shutdown pathname ownership is stated precisely | Current captured `os.FileInfo`, `removeOwnedSocket` device/inode comparison, startup generation comparison, replacement tests, and public lifecycle prose |
| Any restored local work is direct and bounded | One fail-closed generation-aware cleanup outcome with a direct immediate-inode-reuse regression and protocol/configuration/API/publication non-goals |
| The forward queue is complete | R90-59 and R90-75 retain exact blockers; R90-92 has status, dependency, window, risk, acceptance, validation, and stop condition |
| No runtime or follow-on work is started | Exact documentation-only diff and scoped source/test/config/workflow/release review |
| Repository and knowledge records remain valid | Task-state JSON parsing, complete row/Definition multiset comparison, docs, knowledge, diff, staged-scope, and sensitive-information checks |

## Validation

- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, including raw counts, duplicate absence, and asymmetric IDs.
- Run `make docs-check`, `make knowledge-check`, and `git diff --check`.
- Review exact documentation-only scope plus credential, sensitive-path,
  source/test/config/workflow/generated-evidence, release, and publication
  boundaries.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun the knowledge gate, then
  synchronize and verify the exact full-SHA Vault range and current stable
  authority.

## Authority Boundaries

This trigger authorizes only R90-91's delivery-evidence audit, forward queue
and task-state reconciliation, documentation validation, commit/push, and local
Vault synchronization. It does not authorize receiver runtime/test changes,
R90-92 implementation, protocol/configuration/public API changes, private or
external input, performance policy, tag/release/image/registry publication, or
workflow dispatch.

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `FETCH_HEAD == HEAD == origin/main == c0b1eb2dae8dd90eda745eacc87b0a6ece01a450`.
- Verified the exact R90-90 feature/closure parent chain and exact eight/three
  paths, completed plan/state, both immutable Vault notes, full-index rows,
  MOC links, and current stable MOC/UDS authority.
- Parsed all 105 prior task-state JSON files and verified all 94 prior roadmap
  rows and Definitions matched as complete multisets without duplicates or
  asymmetry.
- Reviewed 153 Jul 20 through Aug 12 commits across four dated phases: 61
  behavior-like changes, 78 `docs: record R90-*` closures, and 14 other
  documentation changes. The two commits since the prior audit are exactly the
  R90-90 feature and closure. The R90-90 closure-placement deviation was
  corrected before delivery; no missing record, stale stable authority, or
  unresolved validation result changes priority.
- Confirmed R90-59 and R90-75 retain their external blockers and no local row
  was ready, selecting R90-91 as the documentation-only smallest safe
  unblocker.
- Confirmed startup pathname classification already compares device, inode,
  and change timestamp after an immediate-reuse regression exposed the weaker
  identity check, while shutdown `removeOwnedSocket` still uses only Unix mode
  and `os.SameFile`. Existing shutdown replacement tests cover a regular file
  and symlink, not an immediately rebound Unix listener with inode reuse.

## Audit Checkpoint

- Reconciled R90-90 feature
  `c17870eb7f829b7451ab866b00fead4ef6b72e92` and closure
  `c0b1eb2dae8dd90eda745eacc87b0a6ece01a450` with their exact parent chain,
  intended paths, completed task state, fetched remote, both exact Vault
  notes, full-index rows, MOC links, and current stable authority.
- Counted four delivery phases: Jul 20–26 has 61 commits (30 behavior-like and
  31 closures); Jul 27–Aug 2 has 38 (15 behavior-like, 19 closures, four other
  docs); Aug 3–9 has 46 (14 behavior-like, 24 closures, eight other docs); and
  Aug 10–12 has eight (two behavior-like, four closures, two other docs).
  Resolved validation/delivery deviations remain recorded in their source
  plans; none is an unresolved blocker or missing closure.
- Current public lifecycle prose promises identity-bound shutdown preservation.
  Source captures non-following socket metadata after listener creation but
  shutdown cleanup accepts only socket mode plus `os.SameFile`. The existing
  regular-file and symlink displacement tests cannot expose immediate Unix
  socket inode reuse; the already-delivered startup classification regression
  proves that race occurs on the local filesystem and requires change-time
  generation identity.
- Restored only R90-92 behind this audit: reuse the established non-following
  generation comparison for shutdown cleanup, fail closed when it cannot prove
  ownership, and directly preserve a serviceable immediate replacement
  listener while retaining ordinary cleanup and existing replacements.
  Protocol, peer policy, configuration, public API, runtime/test work, and
  publication remain outside R90-91.
- Persisted the dated audit record with exact phase counts, R90-90
  feature/closure/Vault reconciliation, source/test/public-doc mapping, and the
  complete four-item forward queue.
- The first roadmap-history edit matched an older generic no-ready marker and
  placed the Aug 12 checkpoint before prior R90-82 history. Delivery remained
  blocked while the unchanged paragraph was moved after the exact R90-90
  completion tail; ordered positions must pass before validation.

## Validation Checkpoint

- All 106 task-state JSON files parse and all 96 roadmap rows match exactly one
  of 96 Definitions with equal raw counts, no duplicates, and no asymmetric
  identifiers.
- R90-59, R90-75, R90-91, and R90-92 retain complete status, dependency,
  forecast window, risk, acceptance, required-validation, and stop or unblock
  fields.
- The ordered roadmap positions prove R90-90 completion precedes R90-91
  selection and R90-92 planning. `make docs-check`, all 33
  `make knowledge-check` tests, and `git diff --check` pass.
- Every acceptance criterion maps to the completed evidence. Exact three-path
  documentation scope plus credential, sensitive-path, source/test,
  configuration/workflow, generated-evidence, release, and publication review
  passes; ordinary roadmap text containing the word `token` is not credential
  material. No runtime, test, configuration, workflow, generated evidence,
  release artifact, private-data access, or external mutation was added.
- The only execution deviation was the corrected roadmap-history placement
  recorded above. It did not change audit facts or validation scope, and no
  unresolved validation deviation remains. The existing skill already names
  the required increment-specific history anchor, so no reusable skill change
  is warranted.
- R90-91 satisfies its local acceptance evidence and awaits only documentation
  delivery, fetched remote verification, and exact-range Vault synchronization.
  R90-92 remains planned and unstarted.

## Delivery Results

- Documentation feature commit:
  `972a6714caf91e089a220eeec88c16944a47757d` (`docs: audit post-probe
  delivery`). It contains exactly the three validated roadmap, plan, and
  task-state paths.
- The push produced no output or exit evidence during bounded polling and was
  interrupted. No retry occurred. A fresh non-mutating fetch proved the remote
  had advanced, resolving the ambiguity as successful delivery.
- The fetch verified
  `FETCH_HEAD == HEAD == origin/main == 972a6714caf91e089a220eeec88c16944a47757d`
  with fast-forward ancestry from the recorded baseline. The post-fetch
  33-test knowledge gate passed.
- Exact range
  `c0b1eb2dae8dd90eda745eacc87b0a6ece01a450..972a6714caf91e089a220eeec88c16944a47757d`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the completed R90-91 audit and
  ready/unstarted R90-92 boundary without rewriting immutable iteration notes.
  Replaying the identical range preserved reconciled Vault content hash
  `49af611ecaa37348a699ae392715c157025c9abc51355c8eecea114ae447d2a2`.
- R90-91 is complete. R90-92 is the next ready local increment and remains
  unstarted; R90-59 and R90-75 retain their external blockers.

## Stop Conditions

Stop if R90-90 evidence is missing or contradictory, the pathname-generation
gap cannot be bounded without a public API or product decision, validation
remains ambiguous, or completion would start runtime/test, private-data,
performance-policy, or publication work.
