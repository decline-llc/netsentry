# Task Plan: R90-90 context-aware UDS probe cancellation

## Metadata

- Timestamp: 2026-08-11T03:14:35-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `22ba8ce639d79547875885f4ce107321273dd3b7`

## Goal

Make cancellation during the bounded existing-socket liveness probe terminate
receiver startup promptly before listener readiness while preserving the
captured pathname identity and established active/stale classification.

## Scope

- Pass `Start`'s context through `removeExistingSocket` and the receiver-local
  existing-socket probe seam.
- Replace the fixed standalone dial timeout with a bounded `net.Dialer` whose
  `DialContext` observes both the startup context and the existing one-second
  probe bound.
- Add one direct regression synchronized on probe entry, then cancel startup
  and prove a prompt error matching `context.Canceled`, unchanged Unix-socket
  identity, and no installed receiver listener.
- Preserve direct active-listener, ambiguous-probe, replacement-identity,
  stale-reclamation, already-canceled, ordinary startup, and post-readiness
  cancellation behavior.
- Reconcile current lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority for this increment.

## Non-Goals

- Do not change refusal-only stale classification, pathname generation checks,
  active-peer probing semantics, listener cleanup, retry, or timeout policy.
- Do not add a configuration field, public API, exported test seam, protocol
  frame, peer authentication/trust, cross-process lock, or capture change.
- Do not change connection limits, read deadlines, packet handling, reconnect,
  non-Unix behavior, or post-readiness shutdown ordering.
- Do not access private/external data or perform tag, release, registry,
  workflow, performance-policy, or other publication mutation.

## Risks

- Returning only the dialer's diagnostic can obscure the startup context
  sentinel and make `errors.Is` cancellation checks unreliable.
- Treating cancellation as connection refusal could incorrectly remove a
  pathname that must remain fail-closed.
- A timing-only regression can cancel before probe entry or pass after the full
  fixed timeout without proving prompt context observation.
- Changing a receiver-local seam without updating every direct probe fixture
  can weaken compatibility evidence or leave the package uncompilable.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| During-probe cancellation is prompt and preserves the sentinel | Direct regression runs `Start` asynchronously, blocks the receiver-local probe after signaling entry, cancels, and requires a bounded return satisfying `errors.Is(err, context.Canceled)` |
| Rejected startup preserves pathname identity | The same regression captures non-following metadata before startup and compares Unix-socket mode plus complete `sameUnixSocketIdentity` after return |
| No listener becomes ready | The direct regression proves `r.ln == nil` and the original pathname still occupies the configured path |
| Production probe observes context and remains bounded | Receiver-local probe accepts `context.Context`; production uses `net.Dialer{Timeout: existingSocketProbeTimeout}.DialContext` |
| Classification behavior is unchanged | Direct active-listener, ambiguous-probe, replacement-identity, and stale-reclamation regressions pass through the context-aware seam |
| Lifecycle compatibility is retained | Already-canceled, absent-path ordinary startup, reconnect, and active post-readiness cancellation regressions pass |
| Delivery evidence is repository-complete | Direct ordinary and twenty-count uncached race tests, complete receiver race, native, E2E, docs, knowledge, JSON/Definition, diff, scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `FETCH_HEAD == HEAD == origin/main == 22ba8ce639d79547875885f4ce107321273dd3b7`.
- Verified the exact R90-89 feature/closure parent chain and exact four/three
  paths, completed plan/state, both immutable Vault notes, full-index rows,
  MOC links, current stable MOC/UDS authority, and reconciled closure-range
  Vault hash.
- Parsed all 104 prior task-state JSON files and verified all 94 prior roadmap
  rows and Definitions match as complete multisets without duplicates or
  asymmetry.
- Reviewed 151 Jul 20 through Aug 11 commits: 60 behavior-like changes, 77
  `docs: record R90-*` closures, and 14 other documentation changes. The prior
  ambiguous R90-89 push was resolved by a fetched-behind result and one
  successful retry; no missing record, stale stable authority, or unresolved
  validation result changes priority.
- Confirmed R90-59 and R90-75 retain their external blockers and R90-90 is the
  sole dependency-ready increment.
- Confirmed production probing uses non-contextual `net.DialTimeout`, while
  current direct tests cover pre-canceled startup and post-readiness shutdown
  but not cancellation synchronized after liveness-probe entry.

## Validation

- Run the direct during-probe cancellation regression plus active-listener,
  ambiguous-probe, replacement-identity, stale-reclamation, already-canceled,
  ordinary startup, reconnect, and active-cancellation compatibility tests.
- Run that focused set at least twenty times uncached under the race detector
  from the `engine` Go module, then run the complete receiver package uncached
  under race.
- Preflight the module-selected Go toolchain and every repository-pinned local
  tool needed by the complete fail-fast chain.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, including raw counts, duplicate absence, and asymmetric IDs.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, configuration, protocol, public API,
  release, and publication review.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun the knowledge gate, then
  synchronize and verify the exact full-SHA Vault range and current stable UDS
  authority.

## Authority Boundaries

This trigger authorizes only R90-90's context propagation through the bounded
existing-socket probe, receiver-local seam update, direct synchronized
cancellation regression, compatibility validation, current documentation and
task-state reconciliation, commit/push, and local Vault synchronization. It
does not authorize active/stale classification changes, protocol/configuration/
public API or capture changes, peer trust/authentication, cross-process
ownership, private input, performance policy, tag/release/image/registry
publication, workflow dispatch, or a later roadmap increment.

## Implementation Checkpoint

- The receiver-local probe seam now accepts `context.Context`, and `Start`
  passes its startup context through `removeExistingSocket` without changing
  pathname classification or error wrapping.
- Production probing uses a `net.Dialer` with the established one-second
  timeout and calls `DialContext`, so the earlier of startup cancellation or
  the probe bound terminates the dial.
- The new direct regression creates and closes a Unix listener while retaining
  its pathname, captures non-following metadata, starts the receiver
  asynchronously, and waits on an explicit probe-entry channel before
  canceling. It then requires a prompt error matching `context.Canceled`,
  `r.ln == nil`, Unix-socket mode, and complete device/inode/change-time
  identity preservation.
- Existing active-listener, ambiguous-probe, replacement-identity,
  stale-reclamation, pre-canceled, ordinary startup, reconnect, and active
  post-readiness cancellation regressions pass through the updated seam.
- The complete focused set passes once normally and twenty times uncached under
  the race detector; the complete receiver package also passes uncached under
  race. No implementation or focused-validation deviation occurred.

## Validated Evidence

- The engine module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU
  Make 4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes: both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation
  checks, and all 33 knowledge tests.
- All 105 task-state JSON files parse; all 94 roadmap rows match exactly one of
  94 Definitions with no duplicate or asymmetric identifiers, and every
  unfinished item retains its risk, validation, and stop records.
- Every planned acceptance criterion is satisfied. The direct regression
  reaches probe entry before cancellation and checks the context sentinel,
  prompt return, complete pathname identity, and absent receiver listener;
  nearby pre-canceled or post-readiness tests are retained only as compatibility
  evidence. No implementation or validation deviation occurred.
- `gofmt`, `git diff --check`, exact eight-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.
- R90-90 satisfies its local acceptance evidence and awaits only exact staged
  review, feature delivery, fetched remote verification, and exact-range Vault
  synchronization. No later increment is started.

## Delivery Results

- The first closure edit anchored the completion paragraph on a repeated
  generic sentence and placed the correct facts inside older R90-59a history.
  Delivery stayed blocked while the paragraph was moved after the exact R90-90
  validation marker; the structural gate must prove selection, implementation,
  validation, and completion ordering before commit.

- Feature commit:
  `c17870eb7f829b7451ab866b00fead4ef6b72e92` (`fix: honor cancellation
  during UDS probe`). It contains exactly the eight validated source, test,
  documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == c17870eb7f829b7451ab866b00fead4ef6b72e92`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `22ba8ce639d79547875885f4ce107321273dd3b7..c17870eb7f829b7451ab866b00fead4ef6b72e92`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the delivered context-aware probe,
  direct synchronized cancellation evidence, and unchanged classification/
  public-contract boundary without rewriting immutable iteration notes.
  Replaying the identical range preserved exact Vault content hash
  `07ee79b395aab016fc0e7617999a4f3bca5a8e5b59c772e5a4efec703b3ba997`.
- R90-90 is complete. R90-59 and R90-75 retain their external blockers; no
  dependency-ready local increment remains and no later work was started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps or a public seam, if the
context sentinel cannot be preserved without weakening refusal-only stale
classification or pathname identity, if repeated/full validation remains
ambiguous, or if completion needs private data, product policy, protocol/
configuration/public API, or publication authority.
