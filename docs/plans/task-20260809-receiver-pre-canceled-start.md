# Task Plan: R90-86 receiver pre-canceled startup

## Metadata

- Timestamp: 2026-08-09T12:49:10-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `ab63ee3ef53fdb7a764ca0863dac36580d0318fa`

## Goal

Make an already-canceled receiver startup fail before any configured Unix
pathname mutation or listener creation while preserving established live
startup and post-readiness cancellation behavior.

## Scope

- Check the supplied context before receiver startup touches the configured
  pathname or creates a listener.
- Return a wrapped cancellation error that remains discoverable with
  `errors.Is(err, context.Canceled)`.
- Directly prove that an absent pathname stays absent and a pre-existing Unix
  socket retains the same filesystem identity after rejected startup.
- Retain ordinary absent-path startup, stale-socket reclamation, and
  post-readiness cancellation behavior.
- Reconcile receiver lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority for this increment.

## Non-Goals

- Do not distinguish active from stale Unix sockets, dial an existing peer,
  add peer authentication, or change stale-socket reclamation.
- Do not add cross-process pathname locking or change post-readiness cleanup,
  listener ownership, idle timeout, connection limits, or shutdown ordering.
- Do not change protocol, configuration, public API, packet handling, or
  non-Unix platform policy.
- Do not access operator/private data or perform tag, release, registry,
  workflow, performance-policy, or other publication mutation.

## Risks

- A cancellation check after pathname classification or stale-socket removal
  can report cancellation only after destroying the prior socket identity.
- Reusing the post-readiness watcher or cleanup path for pre-readiness
  cancellation can create and briefly publish a listener despite rejection.
- Broadening the change beyond the initial check can regress the established
  asynchronous cancellation and identity-bound cleanup behavior.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Already-canceled startup preserves an absent path | Direct regression starts with a canceled context, checks `errors.Is`, proves no listener is installed, and confirms the pathname remains absent |
| Already-canceled startup preserves an existing Unix socket | Direct regression captures the socket `FileInfo`, checks `errors.Is`, proves no receiver listener is installed, and compares the same pathname identity afterward |
| Cancellation is checked before all filesystem/listener work | `Start` checks `ctx.Err()` before `removeExistingSocket` and `net.Listen` |
| Live startup remains compatible | Existing absent-path and pre-existing Unix-socket startup regressions pass unchanged |
| Post-readiness cancellation remains compatible | Existing active cancellation, wait, owned-cleanup, and replacement-path regressions pass unchanged |
| Delivery evidence is repository-complete | Focused ordinary and twenty-count uncached race tests, full receiver race, native, E2E, docs, knowledge, JSON/Definition, diff, scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == ab63ee3ef53fdb7a764ca0863dac36580d0318fa`.
- Verified the R90-85 feature and docs-only closure records, completed task
  state, both exact Vault iteration notes, full-index rows, MOC links, and
  current stable MOC/UDS authority.
- Parsed all 100 prior task-state JSON files and verified all 90 roadmap rows
  match exactly one Definition with no duplicate or asymmetric identifiers.
- Reviewed the recent delivery history and found no missing closure, stale
  stable authority, or unresolved validation result that changes priority.
- Confirmed R90-59 and R90-75 retain their exact external blockers and R90-86
  is the sole dependency-ready increment with complete dependency, window,
  risk, acceptance, validation, and stop records.
- Direct source and test review confirms `Receiver.Start` removes an existing
  socket and creates a listener before its cancellation watcher begins, while
  every direct cancellation regression cancels only after successful startup.

## Validation

- Run the two direct R90-86 regressions normally and at least twenty times
  uncached under the race detector from the `engine` module.
- Run the complete receiver package uncached under the race detector.
- Run the complete fail-fast chain: `make test`, `make e2e-smoke`,
  `make docs-check`, and `make knowledge-check` after preflighting required
  local tools and the module-selected Go toolchain.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, duplicate absence, and complete unfinished-item fields.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, configuration, protocol, public API,
  release, and publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range and current stable UDS authority.

## Authority Boundaries

This trigger authorizes only R90-86's pre-filesystem cancellation check, two
direct receiver regressions, compatibility validation, current documentation
and task-state reconciliation, commit/push of the completed increment, and
local Vault synchronization. It does not authorize active/stale peer policy,
cross-process locking, post-readiness lifecycle changes, protocol/config/API
changes, private input, performance policy, tag/release/image/registry
publication, or workflow dispatch.

## Implementation Checkpoint

- `Receiver.Start` now checks `ctx.Err()` before pathname classification,
  stale-socket removal, or listener creation and wraps the returned sentinel
  with the configured listener path for operational context.
- The absent-path regression proves an already-canceled call returns an error
  matching `context.Canceled`, installs no receiver listener, and leaves the
  configured pathname absent.
- The existing-socket regression captures the original Unix socket's
  `FileInfo`, rejects already-canceled startup with the same sentinel, installs
  no receiver listener, and verifies the same socket identity remains at the
  configured pathname.
- The first focused command used repository-relative Go source paths after
  changing into the `engine` module, so `gofmt` stopped on two nonexistent
  paths before tests ran. The corrected module-relative command passed; the
  setup error changed no file and supplies no acceptance evidence.
- Both direct regressions pass once normally and twenty times uncached under
  the race detector. The complete receiver package also passes uncached under
  race, covering live absent/stale startup and established post-readiness
  cancellation and cleanup behavior.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes: both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation
  checks, and all 33 knowledge tests.
- All 101 task-state JSON files parse; all 90 roadmap rows match exactly one of
  90 Definitions with no duplicate or asymmetric identifiers, and every
  unfinished item retains its risk, validation, and stop records.
- Every planned local acceptance criterion is satisfied. The only validation
  deviation was the recorded non-mutating working-directory setup error before
  the focused tests; the complete corrected focused and repository sequences
  passed.
- `gofmt`, `git diff --check`, exact eight-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration, protocol,
  workflow, release artifact, private-data access, or external mutation was
  added.

## Delivery Results

- Feature commit:
  `97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3` (`fix: reject pre-canceled
  receiver startup`). It contains exactly the eight validated source, test,
  documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `HEAD == origin/main == 97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `ab63ee3ef53fdb7a764ca0863dac36580d0318fa..97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the delivered pre-readiness
  cancellation boundary without rewriting immutable iteration notes.
  Replaying the identical range preserved exact Vault content hash
  `41b42418edb5033763e8aa923f9f000765f6bac6cce27270f3c039c1884bc639`.
- R90-86 is complete. R90-59 and R90-75 retain their exact external blockers;
  no dependency-ready local increment remains and no later work was started.

## Stop Conditions

Stop if safe completion requires active/stale peer classification,
cross-process pathname locking, changing post-readiness cleanup or cancellation
ordering, protocol/configuration/public API changes, private data, ambiguous
repeated/full validation, or publication authority.
