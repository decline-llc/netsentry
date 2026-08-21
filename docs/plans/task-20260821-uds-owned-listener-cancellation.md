# Task Plan: R90-102 owned UDS listener cancellation

## Metadata

- Timestamp: 2026-08-21T00:39:32-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `81482afa283a8b5f21e6afa74a43527cb438e9f6`

## Goal

Reject receiver startup when its context is canceled after listener/path
ownership and connection capacity are assigned to the receiver but before
lifecycle goroutines launch or `Start` returns readiness.

## Scope

- Add one receiver-local, unexported synchronization seam after listener,
  pathname, and capacity ownership initialization and before lifecycle launch.
- Check the startup context at that boundary and preserve context-sentinel
  matching through identity-bound public/private cleanup.
- Clear receiver listener, pathname, and capacity ownership before rejected
  startup returns.
- Add direct synchronized regressions for ordinary cancellation and for a
  replacement pathname installed after receiver ownership assignment.
- Prove no accept lifecycle starts at the seam, rejected startup removes only
  its own public/private artifacts, and a replacement listener retains its
  identity, mode, and service.
- Retain every earlier cancellation seam, live startup, configured ownership,
  replacement preservation, and post-readiness shutdown behavior.
- Reconcile lifecycle documentation, roadmap, task state, delivery, and local
  Vault authority.

## Non-Goals

- Do not promise interruption during listener creation, filesystem operations,
  lifecycle goroutine scheduling, or any interval outside the synchronized
  post-ownership/pre-goroutine boundary.
- Do not change the UDS protocol, configuration schema, public API,
  stale/active peer policy, socket mode, connection handling, or shutdown
  ownership rules.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize performance policy, tags,
  releases, images, registry publication, workflow dispatch, or later work.

## Risks

- Receiver fields can expose listener and pathname ownership before any
  cancellation goroutine exists, so rollback must clear internal state as well
  as close and remove created artifacts.
- Cleanup that follows only the configured pathname can remove a replacement
  occupant rather than the receiver's captured socket identity.
- A regression that cancels before ownership assignment or after lifecycle
  launch does not reach the promised boundary.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Cancellation reaches the promised boundary | A receiver-local seam runs only after `r.ln`, captured socket/private-path ownership, and the bounded capacity channel are initialized and before either lifecycle goroutine; tests directly inspect each owned field and capacity at that seam |
| Startup returns the context cause | Direct `errors.Is(err, context.Canceled)` assertions after the post-ownership seam releases |
| Internal ownership is cleared | Direct assertions that `r.ln`, `r.socket`, `r.privateSocketDir`, `r.privateSocketPath`, and `r.slots` are reset after rejection |
| No accept lifecycle launches | The seam directly calls `Wait` before cancellation and proves it returns while `Start` is still blocked before the accept wait-group increment |
| Owned artifacts are removed | Ordinary cancellation proves the configured public path and private staging artifacts are absent after rejection |
| Replacement state is preserved | A direct replacement-listener regression displaces the owned public path before cancellation and proves identity, mode, and service survive while only the owned private artifact is removed |
| Existing lifecycle remains compatible | All earlier cancellation seams, ordinary mode/ownership, listener replacement, live startup, and post-readiness shutdown tests plus complete receiver race coverage |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run both direct post-ownership cancellation regressions and the focused
  adjacent startup cancellation, mode, replacement, and cleanup compatibility
  set.
- Run the acceptance set twenty times uncached under the race detector, then
  the complete receiver package uncached under race.
- Preflight every repository-pinned tool, then run the fail-fast native
  repository checks, E2E smoke, documentation, and knowledge gates applicable
  to the changed surface.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets with equal raw counts and no duplicates or asymmetry.
- Run `git diff --check`; review exact intended scope and anchored credential or
  sensitive-path matches before commit.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun `make knowledge-check`, and
  synchronize the exact full-SHA range to the sole local Vault.

## Authority Boundaries

This trigger authorizes only R90-102 post-ownership/pre-goroutine context
rejection, its direct regressions, compatibility documentation, task-state and
roadmap evidence, a focused feature commit, push of `main` without force or
tags, and local Vault synchronization. It does not authorize protocol,
configuration or public API changes, private/external input, performance
policy, tags, releases, images, registry publication, workflow dispatch, or
later work.

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills, active
  rolling roadmap, knowledge contract, latest R90-101 plan/state, current
  receiver source/tests, and current stable Vault authority.
- Fetched `origin/main` and verified a clean
  `FETCH_HEAD == HEAD == origin/main == 81482afa283a8b5f21e6afa74a43527cb438e9f6`
  baseline.
- Verified the exact R90-101 feature/closure parent chain and three-path scopes,
  completed state, both immutable Vault notes, full-index rows, MOC links, and
  current stable R90-102-ready authority; the 33-test knowledge gate passed.
- Reviewed 175 Jul 20 through Aug 21 commits across four phases: 59 commits Jul
  20-25, 40 Jul 26-Aug 2, 46 Aug 3-9, and 30 Aug 10-21. The only additions to
  the prior audit are the R90-101 feature/closure; no missing record, stale
  stable authority, or unresolved local validation result changes priority.
- Parsed all 116 prior task states and verified all 106 roadmap rows and
  Definitions match as complete multisets without duplicate or asymmetric
  identifiers.
- Confirmed R90-59 and R90-75 retain explicit external blockers and selected
  R90-102 as the sole dependency-ready local increment.
- Confirmed `Start` assigns listener, captured public/private pathname
  ownership, and bounded connection capacity after its final context check,
  then launches cancellation and accept goroutines without another check.

## Checkpoints

- The plan, task state, evidence map, non-goals, authority boundaries, and stop
  conditions are persisted before receiver behavior or compatibility
  documentation changes.
- `Start` now checks the context after assigning its listener, captured public
  and private pathname ownership, and bounded connection capacity but before
  launching either lifecycle goroutine. Rejection first clears every receiver
  ownership field, then closes and removes only the captured listener identity
  and private anchor.
- One direct table regression independently observes complete receiver
  ownership and capacity. It proves the accept wait group returns and no
  context watcher calls `Done` at the seam, then verifies context-sentinel
  matching, complete internal/artifact rollback, and replacement listener
  identity, mode, and service preservation.
- The first formatting/focused command stopped before any test because
  repository-root paths were passed from the Go module. The corrected complete
  formatting, direct regression, and adjacent lifecycle compatibility sequence
  pass; no test result from the stopped setup was counted.
- The acceptance test passes twenty times uncached under race, and the complete
  receiver package passes uncached under race. The module selected pinned Go
  1.25.12; GCC 13.3.0, GNU Make 4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0,
  timeout 9.4, jq 1.7, pkg-config 1.8.1, and libpcap 1.10.4 were available
  before complete validation.
- The complete fail-fast repository chain passes both C tests, every Go package
  uncached under race, E2E smoke with six packets processed, five alerts
  generated, and eight rules loaded, documentation, and all 33 knowledge
  tests.
- All 117 task-state JSON files parse; all 106 roadmap rows match exactly one
  of 106 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers. Ordered history places R90-101 completion, R90-102
  selection, implementation, and validation correctly.
- Every acceptance criterion reaches its direct promised boundary. Formatting,
  exact eight-path scope, and anchored credential, sensitive-path, dependency,
  configuration, protocol, public API, release, and publication reviews pass.
  No dependency, configuration schema, protocol, public API, workflow, release
  artifact, private-data access, or external mutation was added.
- The setup-path deviation is fully resolved by every corrected clean sequence.
  It reiterates the skill's existing module-relative command rule and does not
  warrant a redundant skill change. R90-102 satisfies its local acceptance
  evidence and awaits only delivery; no later increment is started.

## Delivery Results

- Feature commit:
  `2e88d00144e3642c99c7603dc53984cac66b620c` (`fix: reject cancellation
  after UDS ownership`). It contains exactly the eight validated source, test,
  compatibility-documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == 2e88d00144e3642c99c7603dc53984cac66b620c`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `81482afa283a8b5f21e6afa74a43527cb438e9f6..2e88d00144e3642c99c7603dc53984cac66b620c`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the completed R90-102 behavior
  without rewriting immutable iteration notes. Replaying the identical range
  preserved Vault content hash
  `79a8fc2654e00fc5bd650312c60b195b4c63e7cfd5d30176db70971534e4d6d5`.
- R90-102 is complete. No dependency-ready local increment remains; R90-59 and
  R90-75 retain their external blockers and were not started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, an exported/public seam,
interruptible-filesystem guarantees, a dependency, protocol/configuration/
public API authority, private data, performance policy, publication authority,
or if validation remains ambiguous.
