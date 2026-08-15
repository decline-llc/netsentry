# Task Plan: R90-96 private UDS listener cancellation

## Metadata

- Timestamp: 2026-08-15T03:55:05-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `da317004c5ea655cda0ef19388d36c90029428ca`

## Goal

Reject receiver startup when its context is canceled after the private Unix
listener has been created but before that listener is published at the public
pathname or recorded as receiver-owned state.

## Scope

- Pass the startup context through private listener creation and check it at
  the existing synchronized post-creation, pre-publication boundary.
- Preserve context-sentinel matching while cleaning the detached listener,
  private socket path, and staging directory.
- Add one direct synchronized regression proving that the private listener
  exists before cancellation and that rejection publishes no receiver state or
  public/private artifact.
- Retain already-canceled startup, existing-socket probe cancellation,
  ordinary live startup, configured mode/ownership, and shutdown cleanup.
- Reconcile lifecycle documentation, roadmap, task state, delivery, and local
  Vault authority.

## Non-Goals

- Do not promise interruption during `MkdirTemp`, `ListenUnix`, `Chmod`,
  `Lstat`, `Link`, or any arbitrary filesystem operation.
- Do not change the UDS protocol, configuration schema, public API, stale/active
  peer policy, socket mode, connection handling, or shutdown ownership rules.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize performance policy, tags,
  releases, images, registry publication, workflow dispatch, or later work.

## Risks

- Returning success after post-creation cancellation can publish a listener
  that a newly launched cancellation goroutine immediately removes, hiding the
  failed pre-readiness contract from callers.
- Cleanup must close the detached listener without publishing `r.ln`,
  `r.socket`, or private-path ownership and without leaving a staging artifact.
- A regression that cancels before `Start`, during the existing-socket probe,
  or after publication would not reach the promised boundary.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Cancellation reaches the promised boundary | The existing receiver-local seam blocks after `ListenUnix`; the test observes the private socket before canceling and releasing startup |
| Startup returns the context cause | Direct assertion that the returned error matches `context.Canceled` through `errors.Is` |
| No ownership is published | Direct assertions that `r.ln`, `r.socket`, `r.privateSocketDir`, and `r.privateSocketPath` remain unset |
| No filesystem artifact survives | Direct assertions that the configured public path is absent and no `.netsentry-uds-*` staging entry remains |
| Existing lifecycle remains compatible | Already-canceled, probe-cancellation, ordinary mode/ownership, live startup, and shutdown cleanup tests plus complete receiver race coverage |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run the direct post-private-creation cancellation regression and the focused
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

This trigger authorizes only R90-96 context-aware private-listener creation,
its direct regression, compatibility documentation, task-state/roadmap
evidence, a focused feature commit, push of `main` without force or tags, and
local Vault synchronization. It does not authorize protocol/configuration/
public API changes, private or external input, performance policy, tags,
releases, images, registry publication, workflow dispatch, or later work.

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, rolling roadmap,
  knowledge contract, latest R90-95 plan/state, current receiver source/tests,
  and current stable Vault authority.
- Fetched `origin/main` and verified a clean
  `HEAD == origin/main == da317004c5ea655cda0ef19388d36c90029428ca`
  baseline.
- Verified the exact R90-95 feature/closure three-path commits, completed state,
  both immutable Vault notes, full-index rows, MOC links, and stable R90-96-
  ready authority; the 33-test knowledge gate passed.
- Reviewed 163 Jul 20 through Aug 15 commits across four phases: 59 commits Jul
  20-25, 40 Jul 26-Aug 2, 46 Aug 3-9, and 18 Aug 10-15. The only additions to
  the prior audit are the R90-95 feature/closure; no missing record, stale
  stable claim, or unresolved local validation result changes priority.
- Reconciled the R90-96 row from stale `Planned` to `Ready`; R90-59 and R90-75
  retain explicit external blockers and every unfinished item retains its
  dependency, window, risk, acceptance, validation, and stop fields.
- Confirmed `Start` passes no context into `createUnixListener`; the existing
  `afterListenerCreated` seam runs after `ListenUnix` and before mode,
  publication, ownership assignment, or receiver goroutine startup.

## Implementation Checkpoint

- `Start` now passes its context into private listener creation. Immediately
  after the existing listener-created seam, cancellation rejects startup with
  a wrapped context cause and uses the established detached-listener cleanup
  without assigning receiver listener, ownership, or private-path fields.
- The direct regression inspects the staging socket at the synchronized seam,
  cancels before releasing startup, and proves `context.Canceled` matching,
  nil published state, absent public pathname, and complete staging cleanup.
- The first direct run timed out waiting for the seam because the long
  test-derived temporary path plus private staging components exceeded the
  Unix-address limit. The fixture now uses a deliberately short temporary base
  and selects on early startup return while waiting for the seam; the corrected
  direct and focused compatibility sequence passes.
- The seven-test acceptance/compatibility set passes twenty times uncached
  under race, followed by a clean complete receiver-package race run. Complete
  repository validation remains the delivery boundary.

## Validated Evidence

- The engine module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU
  Make 4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six
  packets processed, five alerts generated, and eight rules loaded,
  documentation checks, and all 33 knowledge tests.
- All 111 task-state JSON files parse; all 100 roadmap rows match exactly one
  of 100 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers.
- The direct regression reaches the post-`ListenUnix`, pre-mode/publication
  seam, observes the private socket, and proves the context cause, nil receiver
  ownership, absent public path, and complete private cleanup. The focused
  compatibility set covers the distinct pre-canceled, existing-socket probe,
  ordinary startup/mode, live traffic, and post-readiness shutdown boundaries.
- `gofmt`, `git diff --check`, exact eight-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration schema, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.
- The long Unix-address fixture failure is fully recorded and corrected. The
  direct, repeated race, complete receiver, and complete repository sequences
  all passed after correction; no unresolved validation deviation remains.
  R90-96 awaits only exact staged review, feature delivery, fetched remote
  verification, and exact-range Vault synchronization.

## Delivery Results

- Feature commit:
  `21303ded81714c096851116027b842f4055bff1a` (`fix: reject cancellation
  before UDS publication`). It contains exactly the eight validated source,
  test, compatibility-documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == 21303ded81714c096851116027b842f4055bff1a`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `da317004c5ea655cda0ef19388d36c90029428ca..21303ded81714c096851116027b842f4055bff1a`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the completed R90-96 cancellation
  and artifact-cleanup boundary without rewriting immutable iteration notes.
  Replaying the identical range preserved Vault content hash
  `ba7f350f817a14febdea488c958bb6d061f6e8fe0aaf2d63b651f937f387991e`.
- R90-96 is complete. No dependency-ready local increment remains; R90-59 and
  R90-75 retain their external blockers and were not started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, an exported/public seam,
interruptible-filesystem guarantees, a dependency, protocol/configuration/
public API authority, private data, performance policy, publication authority,
or if validation remains ambiguous.
