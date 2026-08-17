# Task Plan: R90-98 published UDS listener cancellation

## Metadata

- Timestamp: 2026-08-17T01:02:46-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `1a42f0401c49b8ecc25fea361aa846cb6c36c13b`

## Goal

Reject receiver startup when its context is canceled after the created Unix
listener is published at the configured pathname but before `Start` publishes
receiver ownership or reports readiness.

## Scope

- Add one receiver-local, unexported synchronization seam after public/private
  listener identity validation and before ownership reaches `Receiver`.
- Check the startup context at that post-publication boundary and retain
  context-sentinel matching through identity-bound public/private cleanup.
- Add one direct synchronized regression that observes the public and private
  paths as the same Unix-socket identity before cancellation.
- Prove rejection leaves no receiver listener/ownership, configured pathname,
  private socket, or staging directory.
- Retain already-canceled startup, existing-socket probe cancellation,
  post-private-creation cancellation, ordinary mode/ownership, live startup,
  and post-readiness shutdown behavior.
- Reconcile lifecycle documentation, roadmap, task state, delivery, and local
  Vault authority.

## Non-Goals

- Do not promise interruption during `MkdirTemp`, `ListenUnix`, `Chmod`,
  `Lstat`, `Link`, or any arbitrary filesystem operation.
- Do not change the UDS protocol, configuration schema, public API,
  stale/active peer policy, socket mode, connection handling, or shutdown
  ownership rules.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize performance policy, tags,
  releases, images, registry publication, workflow dispatch, or later work.

## Risks

- Returning success after post-publication cancellation can expose receiver
  ownership that the newly launched cancellation goroutine immediately tears
  down, concealing a failed pre-readiness contract from callers.
- Cleanup must use the captured created identity so cancellation cannot remove
  a pathname occupant that replaced the published socket.
- A regression that cancels before publication or after `Start` returns does
  not reach the promised boundary.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Cancellation reaches the promised boundary | A receiver-local seam blocks only after `Lstat` proves the public and private paths are the same created Unix socket; the test independently observes both identities before cancellation |
| Startup returns the context cause | Direct assertion that the returned error matches `context.Canceled` through `errors.Is` |
| No ownership is published | Direct assertions that `r.ln`, `r.socket`, `r.privateSocketDir`, and `r.privateSocketPath` remain unset |
| Only created artifacts are removed | Direct assertions that the configured public path and private staging artifacts are absent after rejection, using the established identity-bound cleanup path |
| Existing lifecycle remains compatible | Already-canceled, probe-cancellation, post-private-creation cancellation, ordinary mode/ownership, live startup, and shutdown tests plus complete receiver race coverage |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run the direct post-publication cancellation regression and the focused
  startup cancellation/mode/cleanup compatibility set.
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

This trigger authorizes only R90-98 post-publication/pre-readiness context
rejection, its direct regression, compatibility documentation, task-state and
roadmap evidence, a focused feature commit, push of `main` without force or
tags, and local Vault synchronization. It does not authorize protocol,
configuration or public API changes, private/external input, performance
policy, tags, releases, images, registry publication, workflow dispatch, or
later work.

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, rolling roadmap,
  knowledge contract, latest R90-97 plan/state, current receiver source/tests,
  and current stable Vault authority.
- Fetched `origin/main` through the documented SSH-over-443 fallback and
  verified a clean
  `FETCH_HEAD == HEAD == origin/main == 1a42f0401c49b8ecc25fea361aa846cb6c36c13b`
  baseline.
- Verified the exact R90-97 feature/closure three-path commits, completed
  state, both immutable Vault notes, full-index rows, MOC links, and stable
  R90-98-ready authority; the 33-test knowledge gate passed.
- Reviewed 167 Jul 20 through Aug 17 commits across four phases: 59 commits Jul
  20-25, 40 Jul 26-Aug 2, 46 Aug 3-9, and 22 Aug 10-17. The only additions to
  the prior audit are the R90-97 feature/closure; no missing record, stale
  stable claim, or unresolved local validation result changes priority.
- Parsed all 112 prior task states and verified all 102 prior roadmap rows and
  Definitions match as complete multisets without duplicate or asymmetric
  identifiers.
- Confirmed R90-59 and R90-75 retain explicit external blockers and selected
  R90-98 as the sole dependency-ready local increment.
- Confirmed `createUnixListener` validates the published public/private socket
  identity and then returns it to `Start`; neither function checks the context
  again before `Start` assigns receiver ownership, launches lifecycle
  goroutines, and returns success.

## Implementation Checkpoint

- `Receiver` now has one unexported post-publication seam. It runs only after
  the configured public pathname and private staging path have passed the
  existing non-following Unix-socket identity check.
- Immediately after that seam, active cancellation returns a wrapped context
  cause through the established failure path. That path removes the public
  hard link only while it still matches the created identity, then closes the
  detached listener and removes the private socket and staging directory;
  `Start` receives no listener or ownership to publish.
- The direct regression uses a deliberately short temporary base, independently
  observes that both paths exist as the same Unix-socket identity, cancels and
  releases the seam, then proves `context.Canceled` matching, nil receiver
  ownership, an absent public pathname, and no private staging artifact.
- The first focused command stopped during `gofmt` because repository-relative
  paths were passed from the `engine` module directory. It changed nothing and
  ran no test. The complete sequence was rerun with module-relative paths and
  passed.
- The corrected focused acceptance/compatibility set passes normally and
  twenty times uncached under race; the complete receiver package also passes
  uncached under race. Complete repository validation remains the delivery
  boundary.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The first complete fail-fast chain stopped in the unchanged R90-92
  `TestRemoveOwnedSocketPreservesImmediateReplacementUnixListenerWithReusedInode`
  fixture because that run did not induce immediate inode reuse. The exact test
  then passed twenty uncached race executions, so the failure did not
  reproduce and no unrelated source or test changed.
- A complete restart of the fail-fast chain passes both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six
  packets processed, five alerts generated, and eight rules loaded,
  documentation checks, and all 33 knowledge tests.
- All 113 task-state JSON files parse; all 102 roadmap rows match exactly one
  of 102 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers. R90-59 and R90-75 retain their exact external
  blockers.
- The direct regression reaches the post-publication, pre-ownership seam,
  observes the public/private listener identities, and proves context-cause
  matching, nil receiver ownership, absent public path, and complete private
  cleanup. The focused compatibility set covers every adjacent cancellation,
  ordinary mode/ownership, live-startup, and post-readiness boundary named by
  R90-98.
- `gofmt`, `git diff --check`, exact eight-path scope, and anchored credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration schema, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.
- The setup-only module-path error and the nonreproducing R90-92 fixture failure
  remain recorded at their exact strengths. Every required sequence was rerun
  successfully after its deviation; no unresolved validation result remains.
  R90-98 awaits only exact staged review, feature delivery, fetched remote
  verification, and exact-range Vault synchronization.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, an exported/public seam,
interruptible-filesystem guarantees, a dependency, protocol/configuration/
public API authority, private data, performance policy, publication authority,
or if validation remains ambiguous.
