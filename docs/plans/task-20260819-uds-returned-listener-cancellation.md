# Task Plan: R90-100 returned UDS listener cancellation

## Metadata

- Timestamp: 2026-08-19T06:19:17-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `039cd60a04b0e682a282f9d0f22c130f7cfedcfc`

## Goal

Reject receiver startup when its context is canceled after
`createUnixListener` returns a live public/private listener identity but before
`Start` publishes that listener and pathname ownership on the receiver.

## Scope

- Add one receiver-local, unexported synchronization seam immediately after
  successful listener creation returns and before receiver ownership assignment.
- Check the startup context at that post-return boundary and retain
  context-sentinel matching through identity-bound public/private cleanup.
- Add direct synchronized regressions for ordinary cancellation and for a
  replacement pathname installed after listener return.
- Prove rejected startup publishes no receiver listener or pathname ownership,
  removes its own public/private artifacts, and preserves a replacement listener
  identity and service.
- Retain entry, probe, private-creation, publication, ordinary mode/ownership,
  live-startup, replacement-preservation, and post-readiness shutdown behavior.
- Reconcile lifecycle documentation, roadmap, task state, delivery, and local
  Vault authority.

## Non-Goals

- Do not promise interruption during listener creation, filesystem operations,
  or any interval other than the synchronized post-return/pre-ownership boundary.
- Do not change the UDS protocol, configuration schema, public API,
  stale/active peer policy, socket mode, connection handling, or shutdown
  ownership rules.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize performance policy, tags,
  releases, images, registry publication, workflow dispatch, or later work.

## Risks

- A live returned listener has public and private filesystem artifacts but no
  receiver owner or cancellation goroutine, so returning success after
  cancellation violates the pre-readiness contract.
- Cleanup that follows only the configured pathname can remove a replacement
  occupant rather than the returned listener's captured identity.
- A regression that cancels inside `createUnixListener` or after ownership is
  assigned does not reach the promised boundary.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Cancellation reaches the promised boundary | A receiver-local seam runs only after `createUnixListener` returns its listener, captured socket identity, and private paths and before any of them are assigned to `Receiver`; tests independently observe those artifacts before cancellation |
| Startup returns the context cause | Direct `errors.Is(err, context.Canceled)` assertions after the post-return seam releases |
| No ownership is published | Direct assertions that `r.ln`, `r.socket`, `r.privateSocketDir`, and `r.privateSocketPath` remain unset |
| Owned artifacts are removed | Ordinary cancellation proves the configured public path and private staging artifacts are absent after rejection |
| Replacement state is preserved | A direct replacement-listener regression displaces the returned public path before cancellation and proves identity, mode, and service survive while only the returned private artifact is removed |
| Existing lifecycle remains compatible | Entry/probe/private/publication cancellation, ordinary mode/ownership, listener replacement, live startup, and post-readiness shutdown tests plus complete receiver race coverage |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run both direct post-return cancellation regressions and the focused adjacent
  startup cancellation, mode, replacement, and cleanup compatibility set.
- Run the acceptance set twenty times uncached under the race detector, then
  the complete receiver package uncached under race.
- Preflight repository-pinned tools, then run the fail-fast native repository
  checks, E2E smoke, documentation, and knowledge gates applicable to the
  changed surface.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets with equal raw counts and no duplicates or asymmetry.
- Run `git diff --check`; review exact intended scope and anchored credential or
  sensitive-path matches before commit.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun `make knowledge-check`, and
  synchronize the exact full-SHA range to the sole local Vault.

## Authority Boundaries

This trigger authorizes only R90-100 post-return/pre-ownership context
rejection, its direct regressions, compatibility documentation, task-state and
roadmap evidence, a focused feature commit, push of `main` without force or
tags, and local Vault synchronization. It does not authorize protocol,
configuration or public API changes, private/external input, performance
policy, tags, releases, images, registry publication, workflow dispatch, or
later work.

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, complete rolling
  roadmap, knowledge contract, latest R90-99 plan/state, current receiver
  source/tests, and current stable Vault authority.
- Fetched `origin/main` and verified a clean
  `FETCH_HEAD == HEAD == origin/main == 039cd60a04b0e682a282f9d0f22c130f7cfedcfc`
  baseline.
- Verified the exact R90-99 feature/closure parent chain and three-path scopes,
  completed state, both immutable Vault notes, full-index rows, MOC links, and
  current stable R90-100-ready authority; the 33-test knowledge gate passed.
- Reviewed 171 Jul 20 through Aug 19 commits across four phases: 59 commits Jul
  20-25, 40 Jul 26-Aug 2, 46 Aug 3-9, and 26 Aug 10-19. The only additions to
  the prior audit are the R90-99 feature/closure; no missing record, stale
  stable authority, or unresolved local validation result changes priority.
- Parsed all 114 prior task states and verified all 104 roadmap rows and
  Definitions match as complete multisets without duplicate or asymmetric
  identifiers.
- Confirmed R90-59 and R90-75 retain explicit external blockers and selected
  R90-100 as the sole dependency-ready local increment.
- Confirmed `createUnixListener` returns a live listener, captured public socket
  identity, and private path anchor to `Start`; `Start` assigns all four to the
  receiver and launches lifecycle goroutines without another context check.

## Checkpoints

- The plan, task state, evidence map, non-goals, authority boundaries, and stop
  conditions are persisted before receiver behavior or compatibility
  documentation changes.
- `Start` now checks the context after `createUnixListener` returns and before
  assigning any listener or pathname state to `Receiver`. Rejection uses the
  captured socket identity to remove the public path, closes the detached
  listener, and removes its private identity anchor and staging directory.
- One direct table regression independently observes the returned public and
  private socket identity. Its ordinary case proves complete owned-artifact
  cleanup; its replacement case displaces the public path with a live listener
  and proves identity, mode, and service preservation after cancellation.
- The initial formatting command stopped on a missing closing test brace before
  any test ran. After correction, the direct test, complete focused compatibility
  set, twenty uncached race repetitions, and the complete receiver race package
  pass.
- The module selected the pinned Go 1.25.12 toolchain. GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation, and
  all 33 knowledge tests.
- All 115 task-state JSON files parse; all 104 roadmap rows match exactly one of
  104 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers. Ordered history places R90-98 completion, R90-99
  selection/completion, R90-100 selection, and R90-100 validation correctly.
- Every acceptance criterion reaches its direct promised boundary. Formatting,
  exact eight-path scope, and anchored credential, sensitive-path, dependency,
  configuration, protocol, public API, release, and publication reviews pass.
  No dependency, configuration schema, protocol, public API, workflow, release
  artifact, private-data access, or external mutation was added.
- R90-100 satisfies its local acceptance evidence and awaits only exact staged
  review, feature delivery, fetched remote verification, and exact-range Vault
  synchronization. No later increment is started.

## Delivery Results

- Feature commit:
  `286531d3748c27edff8172d9b78f0f54a070937a` (`fix: reject cancellation
  after UDS listener return`). It contains exactly the eight validated source,
  test, compatibility-documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == 286531d3748c27edff8172d9b78f0f54a070937a`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `039cd60a04b0e682a282f9d0f22c130f7cfedcfc..286531d3748c27edff8172d9b78f0f54a070937a`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the completed R90-100 behavior
  without rewriting immutable iteration notes. Replaying the identical range
  preserved Vault content hash
  `be26d72f5208ced6198e0c92ce39d450eb8a5a4b491509b9523f4ca99c3db599`.
- R90-100 is complete. No dependency-ready local increment remains; R90-59
  and R90-75 retain their external blockers and were not started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, an exported/public seam,
interruptible-filesystem guarantees, a dependency, protocol/configuration/
public API authority, private data, performance policy, publication authority,
or if validation remains ambiguous.
