# Task Plan: R90-104 UDS lifecycle-launch cancellation

## Metadata

- Timestamp: 2026-08-22T05:46:28-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `7b7821678c1b09336ea8b8bcce990dfd9de84f01`

## Goal

Reject receiver startup when its context is canceled after the cancellation
watcher and accept loop have both launched but before `Start` returns
readiness, and complete lifecycle termination plus ownership rollback before
returning the rejection.

## Scope

- Add one receiver-local, unexported synchronization seam after both lifecycle
  goroutines have observably started and before readiness return.
- Check the startup context at that boundary and preserve context-sentinel
  matching through bounded lifecycle shutdown and identity-bound cleanup.
- Track the cancellation watcher inside startup rollback and retain the
  receiver wait group for the accept loop and handlers so rejected startup can
  wait for both without changing public `Stop(); Wait()` semantics.
- Add direct synchronized regressions for ordinary cancellation and for a
  replacement pathname installed after lifecycle launch.
- Prove both lifecycle goroutines are live at the seam, both terminate before
  rejection returns, owned public/private artifacts are absent, and a
  replacement listener retains identity, mode, and service.
- Retain every earlier cancellation seam, live startup, configured ownership,
  replacement preservation, connection handling, and post-readiness shutdown
  behavior.
- Reconcile compatibility documentation, roadmap, task state, delivery, and
  local Vault authority.

## Non-Goals

- Do not promise cancellation at arbitrary goroutine-scheduling or filesystem
  intervals outside the synchronized post-launch/pre-return boundary.
- Do not change the UDS protocol, configuration schema, public API,
  stale/active peer policy, socket mode, connection limits, frame handling, or
  shutdown pathname-ownership rules.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize performance policy, tags,
  releases, images, registry publication, workflow dispatch, or later work.

## Risks

- Returning while either goroutine can still read receiver ownership can race
  field clearing or defer cleanup until after rejected startup appears done.
- Closing the owned listener before pathname identity is checked can remove a
  replacement occupant or allow listener auto-unlink to escape the ownership
  boundary.
- A regression that cancels before both lifecycle goroutines start or after
  `Start` returns does not reach the promised boundary.
- Waiting on lifecycle work without first making the accept loop interruptible
  can deadlock rejected startup.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Cancellation reaches the promised boundary | A receiver-local seam runs only after explicit observable-start synchronization for the cancellation watcher and accept loop and before readiness return; the direct test blocks `Start` at that seam |
| Startup returns the context cause | Direct `errors.Is(err, context.Canceled)` assertions after post-launch cancellation and seam release |
| Both launched goroutines terminate before rejection | Separate watcher accounting plus the existing receiver wait group are both joined before rollback; the direct regression reaches that path and proves receiver `Wait` has returned by the time rejected `Start` returns |
| Internal ownership is cleared safely | Direct assertions after lifecycle termination that `r.ln`, `r.socket`, `r.privateSocketDir`, `r.privateSocketPath`, and `r.slots` are reset |
| Owned artifacts are removed | Ordinary cancellation proves the configured public path and private staging artifacts are absent after rejection |
| Replacement state is preserved | A direct replacement-listener regression displaces the owned public path after lifecycle launch and proves identity, mode, and service survive while only owned artifacts are removed |
| Existing lifecycle remains compatible | All earlier cancellation seams, ordinary mode/ownership, listener replacement, live connection handling, and post-readiness shutdown tests plus complete receiver race coverage |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run both direct post-lifecycle-launch cancellation regressions and the
  focused adjacent startup cancellation, mode, replacement, and cleanup set.
- Run the acceptance test twenty times uncached under the race detector, then
  the complete receiver package uncached under race.
- Preflight every repository-pinned tool, then run the fail-fast native
  repository tests, E2E smoke, documentation, and knowledge gates applicable
  to the changed surface.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets with equal raw counts and no duplicates or asymmetry.
- Run `git diff --check`; review exact intended scope and anchored credential or
  sensitive-path matches before commit.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun `make knowledge-check`, and
  synchronize the exact full-SHA range to the sole local Vault.

## Authority Boundaries

This trigger authorizes only R90-104 post-lifecycle-launch/pre-readiness
context rejection, its direct regressions, compatibility documentation,
task-state and roadmap evidence, a focused feature commit, push of `main`
without force or tags, and local Vault synchronization. It does not authorize
protocol, configuration or public API changes, private/external input,
performance policy, tags, releases, images, registry publication, workflow
dispatch, or later work.

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills, active
  rolling roadmap, knowledge contract, latest R90-103 and R90-102 plan/state,
  current receiver source/tests, and current stable Vault authority.
- Fetched `origin/main` and verified a clean
  `FETCH_HEAD == HEAD == origin/main == 7b7821678c1b09336ea8b8bcce990dfd9de84f01`
  baseline.
- Verified the exact R90-103 feature/closure parent chain and three-path
  scopes, completed state, both immutable Vault notes, full-index rows, MOC
  links, and current stable R90-104-ready authority; the 33-test knowledge gate
  passed.
- Reviewed 179 Jul 20 through Aug 22 commits across four phases: 59 commits Jul
  20-25, 40 Jul 26-Aug 2, 46 Aug 3-9, and 34 Aug 10-22. The only additions to
  the prior audit are the R90-103 feature/closure; no missing record, stale
  stable authority, or unresolved local validation result changes priority.
- Parsed all 118 prior task states and verified all 108 roadmap rows and
  Definitions match as complete multisets without duplicate or asymmetric
  identifiers.
- Confirmed R90-59 and R90-75 retain explicit external blockers and selected
  R90-104 as the sole dependency-ready local increment.
- Confirmed `Start` launches an untracked context watcher and a wait-group-
  tracked accept loop after its final context check, then returns readiness
  without a later check or direct post-launch/pre-return regression.

## Checkpoints

- The plan, task state, evidence map, non-goals, authority boundaries, and stop
  conditions are persisted before receiver behavior or compatibility
  documentation changes.
- `Start` now waits for observable entry into both lifecycle goroutines, then
  checks cancellation before readiness return. Rejection cancels the derived
  lifecycle, joins the watcher separately and the accept/handler wait group,
  and clears ownership only after both boundaries terminate. The separate
  watcher accounting preserves existing public `Stop(); Wait()` behavior.
- One direct synchronized table regression independently observes the complete
  receiver ownership and captured public/private listener identity after both
  lifecycle starts. It verifies context-sentinel matching, bounded lifecycle
  termination before rejection returns, complete internal/artifact rollback,
  and replacement listener identity, mode, and service preservation.
- Focused normal and adjacent lifecycle race validation pass. The acceptance
  regression passes twenty times uncached under race, followed by a clean
  complete receiver-package race run. No implementation or focused-validation
  deviation occurred.
- The module selected pinned Go 1.25.12; GCC 13.3.0, GNU Make 4.3, Bash
  5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7, pkg-config 1.8.1,
  and libpcap 1.10.4 were available before complete validation.
- The complete fail-fast repository chain passes both C tests, every Go
  package uncached under race, E2E smoke with six packets processed, five
  alerts generated, and eight rules loaded, documentation, and all 33
  knowledge tests.
- All 119 task-state JSON files parse; all 108 roadmap rows match exactly one
  of 108 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers. Ordered history places R90-103 completion, R90-104
  selection, implementation, and validation correctly.
- Every acceptance criterion reaches its direct promised boundary. Formatting,
  exact eight-path scope, and anchored credential, sensitive-path, dependency,
  configuration, protocol, public API, release, and publication reviews pass.
  No dependency, configuration schema, protocol, public API, workflow, release
  artifact, private-data access, or external mutation was added.
- The first pre-commit chronology command used a repeated generic complete-
  validation sentence and stopped after resolving it to the older R90-102
  checkpoint. No later output from that fail-fast sequence was counted. The
  corrected complete sequence uses the unique R90-104 acceptance marker and
  must pass before delivery.
- That setup deviation reiterates the skill's existing increment-specific
  history-anchor rule and does not warrant a redundant skill change. R90-104
  otherwise satisfies its local acceptance evidence and awaits only delivery;
  no later increment is started.

## Delivery Results

- Feature commit:
  `513da95a0819b0c3886654dc78122c5be50a8dea` (`fix: reject cancellation
  after UDS lifecycle launch`). It contains exactly the eight validated source,
  test, compatibility-documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == 513da95a0819b0c3886654dc78122c5be50a8dea`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `7b7821678c1b09336ea8b8bcce990dfd9de84f01..513da95a0819b0c3886654dc78122c5be50a8dea`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the completed R90-104 behavior
  without rewriting immutable iteration notes. Replaying the identical range
  preserved Vault content hash
  `594fc6c4a408c9c76bef547c3ea3028e5d6747c49badcce7cf6da078cf64aa52`.
- R90-104 is complete. No dependency-ready local increment remains; R90-59 and
  R90-75 retain their external blockers and were not started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, an exported/public seam,
interruptible-filesystem guarantees, a dependency, protocol/configuration/
public API authority, private data, performance policy, publication authority,
or if validation remains ambiguous.
