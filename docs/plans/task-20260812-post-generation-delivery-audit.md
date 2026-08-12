# Task Plan: R90-93 post-generation delivery audit

## Metadata

- Timestamp: 2026-08-12T09:08:26-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `c59c3aca6a67b1975f178734d6b0f81a6bcab6b8`

## Goal

Reconcile the completed R90-92 feature and docs-only closure against current
Git, task-state, fetched remote, Vault, recent delivery history, and the
post-listen socket mode/ownership boundary; restore at most one directly
evidenced local follow-on without changing runtime or tests.

## Scope

- Verify the exact R90-92 feature/closure parent chain, intended paths,
  completed plan/state, fetched remote, and both Vault records.
- Review the Jul 20 through Aug 12 delivery phases and retain only material
  trends, validation deviations, stale authority, missing records, and risks.
- Trace the interval from `net.Listen` through mode application, non-following
  pathname inspection, ownership publication, and shutdown cleanup through
  current source, direct tests, Go's local API contract, and public docs.
- Separate created-listener ownership from protocol, configuration schema,
  public API, peer policy, and general cross-process ownership.
- Define at most one bounded follow-on, then refresh the roadmap and task state
  without starting it.

## Non-Goals

- Do not change receiver source/tests, configuration, protocol, public API,
  workflows, release gates, benchmark/evidence artifacts, or immutable Vault
  iteration notes.
- Do not implement descriptor-bound mode application, add a production seam,
  or alter startup/shutdown behavior.
- Do not authorize comparable-environment performance scope, remote tag push,
  GitHub Release, GHCR publication, workflow dispatch, or private data access.

## Risks

- `os.Chmod` following a symlink is a documented pathname behavior, but the
  audit must not claim an observed production exploit or generalize beyond the
  bounded post-listen interval.
- A weak follow-on could apply mode safely yet still capture a replacement
  socket as owned, or verify identity without proving replacement targets were
  not modified.
- Tests that replace a pathname before `Start` or during stale probing do not
  directly reach replacement after listener creation.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-92 is exactly recoverable | Feature/closure SHAs and parents, intended paths, completed plan/state, fetched `FETCH_HEAD == HEAD == origin/main`, and two exact Vault notes/index/MOC links |
| Recent history has no concealed blocker | Four dated phase counts plus material validation, stale-authority, missing-record, and unresolved-risk review |
| Post-listen ownership is stated precisely | Current `net.Listen`, pathname `os.Chmod`, later `Lstat`, ownership publication, shutdown cleanup, local `go doc os.Chmod`, direct tests, and public lifecycle prose |
| Any restored local work is direct and bounded | One created-listener-bound mode/identity outcome with direct regular-file, symlink-target, and replacement-listener preservation evidence |
| The forward queue is complete | R90-59 and R90-75 retain exact blockers; R90-94 has status, dependency, window, risk, acceptance, validation, and stop condition |
| No runtime or follow-on work is started | Exact documentation-only diff and scoped source/test/config/workflow/release review |
| Repository and knowledge records remain valid | Task-state JSON parsing, complete row/Definition multiset comparison, docs, knowledge, diff, staged-scope, and sensitive-information checks |

## Validation

- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, including raw counts, duplicate absence, and asymmetric IDs.
- Run `make docs-check`, `make knowledge-check`, and `git diff --check`.
- Verify ordered roadmap history from R90-92 completion through R90-93
  selection and R90-94 planning.
- Review exact documentation-only scope plus credential, sensitive-path,
  source/test/config/workflow/generated-evidence, release, and publication
  boundaries.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun the knowledge gate, then
  synchronize and verify the exact full-SHA Vault range and current stable
  authority.

## Authority Boundaries

This trigger authorizes only R90-93's delivery-evidence audit, forward queue
and task-state reconciliation, documentation validation, commit/push, and local
Vault synchronization. It does not authorize receiver runtime/test changes,
R90-94 implementation, protocol/configuration/public API changes, private or
external input, performance policy, tag/release/image/registry publication, or
workflow dispatch.

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- Fetched `origin/main` through the documented IPv4 SSH-over-443 transport and
  verified a clean worktree with
  `FETCH_HEAD == HEAD == origin/main == c59c3aca6a67b1975f178734d6b0f81a6bcab6b8`.
- Verified the exact R90-92 feature/closure parent chain and exact eight/three
  paths, completed plan/state, both immutable Vault notes, full-index rows,
  MOC links, and current stable MOC/UDS authority.
- Parsed all 107 prior task-state JSON files and verified all 96 prior roadmap
  rows and Definitions match as complete multisets without duplicates or
  asymmetry.
- Reviewed 157 Jul 20 through Aug 12 commits across four phases: 61 commits
  Jul 20-26, 38 Jul 27-Aug 2, 46 Aug 3-9, and 12 Aug 10-12. The last phase has
  three behavior-like changes, six delivery closures, and three other docs;
  only the R90-92 feature/closure follows the prior audit, with no missing
  record, stale stable authority, or unresolved validation result.
- Confirmed R90-59 and R90-75 retain their external blockers and no
  dependency-ready local row remains, selecting R90-93 as the documentation-
  only smallest safe unblocker.
- Confirmed `Start` calls pathname-based `os.Chmod` after `net.Listen` and
  before non-following `Lstat`; the local Go API contract states that
  pathname-based `Chmod` changes a symlink target. Existing direct tests do not
  replace the pathname during this post-listen interval.

## Audit Checkpoint

- Reconciled R90-92 feature
  `b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e` and closure
  `c59c3aca6a67b1975f178734d6b0f81a6bcab6b8` with their exact parent chain,
  intended paths, completed task state, fetched remote, both exact Vault
  notes, full-index rows, MOC links, and current stable authority.
- Counted four delivery phases: Jul 20-26 has 61 commits (30 behavior-like and
  31 closures); Jul 27-Aug 2 has 38 (15 behavior-like, 19 closures, four other
  docs); Aug 3-9 has 46 (14 behavior-like, 24 closures, eight other docs); and
  Aug 10-12 has 12 (three behavior-like, six closures, three other docs).
  Resolved validation and transport deviations remain recorded in their source
  plans; none is an unresolved blocker or missing delivery record.
- `Start` currently creates the listener, disables auto-unlink, calls
  pathname-based `os.Chmod`, then calls non-following `os.Lstat` and publishes
  that result as owned. The local Go API contract explicitly says pathname
  `Chmod` changes a symlink target. A regular-file or symlink replacement can
  therefore be mutated before rejection, while a replacement Unix listener can
  be captured as owned even though the active `net.Listener` refers to the
  unlinked original socket.
- Existing direct regressions cover a regular file or symlink present before
  `Start`, replacement during stale-socket probing, and regular-file, symlink,
  or immediate listener replacement during shutdown. None synchronizes after
  `net.Listen` and before mode/ownership handling, so those tests do not reach
  this boundary.
- Restored only R90-94 behind this audit: apply mode through the created
  listener identity, require the non-following pathname still matches before
  ownership publication, and directly preserve regular-file bytes/mode,
  symlink plus target mode, and replacement-listener identity/mode/service.
  Protocol, configuration schema, public API, peer policy, general cross-
  process ownership, runtime/test work, and publication remain outside R90-93.
- The created-object metadata lesson is reusable, so the separate local
  `netsentry-next` skill now prefers handle-bound metadata mutation before
  non-following ownership capture and forbids following pathname mutation while
  replacement is possible. No skill file is included in this repository audit.
- The first two roadmap-history edits matched older no-ready checkpoints and
  placed the new audit before Aug 9 history. Delivery remained blocked while
  the unchanged R90-93 paragraph was anchored after the exact R90-92 content
  hash and completion tail; ordered positions must pass before validation.

## Validation Checkpoint

- All 108 task-state JSON files parse and all 98 roadmap rows match exactly one
  of 98 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers.
- R90-59, R90-75, R90-93, and R90-94 retain complete status, dependency,
  forecast window, risk, acceptance, required-validation, and stop or unblock
  fields.
- Ordered roadmap positions prove R90-92 completion precedes R90-93 selection
  and R90-94 planning. `make docs-check`, all 33 `make knowledge-check` tests,
  and `git diff --check` pass.
- Every acceptance criterion maps to the completed evidence. Exact three-path
  documentation scope plus credential, sensitive-path, source/test,
  configuration/workflow, generated-evidence, release, and publication review
  passes. No runtime, test, configuration, workflow, generated evidence,
  release artifact, private-data access, or external mutation was added.
- The only execution deviation was the corrected two-attempt roadmap-history
  placement recorded above. It did not change audit facts or validation scope,
  and no unresolved validation deviation remains. The existing skill already
  requires increment-specific history anchors, so no additional skill change
  is warranted.
- R90-93 satisfies its local acceptance evidence and awaits only documentation
  delivery, fetched remote verification, and exact-range Vault synchronization.
  R90-94 remains planned and unstarted.

## Stop Conditions

Stop if R90-92 evidence is missing or contradictory, the post-listen ownership
gap cannot be bounded without an exported API or product decision, validation
remains ambiguous, or completion would start runtime/test, private-data,
performance-policy, or publication work.
