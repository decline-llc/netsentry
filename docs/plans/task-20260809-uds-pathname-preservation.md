# Task Plan: R90-84 UDS pathname preservation

## Metadata

- Timestamp: 2026-08-09T12:16:37-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `5c4253d18283c80ec27b7c2c1f383616eac2a89e`

## Goal

Keep receiver startup and shutdown within the Unix-socket filesystem identity
created by that receiver, preserving unrelated regular files and symlinks
without changing the existing stale- or active-socket reclamation policy.

## Scope

- Inspect the configured pathname without following symlinks before startup.
- Continue removing a pre-existing Unix socket before listening, preserving the
  established stale/active socket policy.
- Reject any pre-existing non-socket or symlink before mutation and leave its
  content or link target unchanged.
- Disable the Go Unix listener's automatic pathname unlink and retain the
  created socket's filesystem identity for explicit receiver-owned cleanup.
- On shutdown, remove the pathname only when it still identifies the socket
  created by that receiver; preserve a replacement regular file or symlink.
- Reconcile receiver lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority for this increment.

## Non-Goals

- Do not distinguish an active Unix socket from a stale one, dial an existing
  peer, add peer authentication, or change reconnect/session protocol.
- Do not add cross-process locking or claim an atomic ownership guarantee
  against a process replacing the path between inspection and removal.
- Do not change socket mode, connection limits, idle timeout, packet handling,
  configuration schema, public API, or non-Unix platform policy.
- Do not access operator/private data or perform tag, release, registry,
  workflow, performance-policy, or other publication mutation.

## Risks

- `net.UnixListener` normally unlinks its path during close, so an explicit
  post-close identity check alone would still delete a replacement occupant.
- Following a symlink during classification can mutate its target or misclassify
  the pathname as an owned socket.
- Comparing only file type at shutdown can delete a different socket or other
  occupant that replaced the listener pathname.
- Startup inspection and removal cannot become cross-process atomic without a
  broader locking or platform-specific design that is outside this increment.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Pre-existing regular files are preserved | Direct startup regression compares exact bytes and mode after rejection and proves no listener was installed |
| Pre-existing symlinks are preserved without following | Direct startup regression points the link at an active Unix socket, compares the exact link target and target socket identity after rejection, and proves no listener was installed |
| Absent-path and stale-socket startup remain compatible | Existing absent-path receiver tests plus a direct pre-existing Unix-socket restart regression |
| Ordinary owned socket cleanup remains intact | Existing cancellation/shutdown regressions and a direct pathname-absence assertion after receiver wait |
| Replacement regular files survive shutdown | Direct regression unlinks the owned socket pathname, writes exact replacement bytes, cancels, waits, and compares the replacement |
| Replacement symlinks survive shutdown | Direct regression displaces the owned socket with a link, cancels, waits, and verifies exact link target and target bytes |
| Cleanup is identity-bound | Listener auto-unlink is disabled; shutdown compares the current `Lstat` identity and socket type with the captured created-socket identity before removal |
| Delivery evidence is repository-complete | Focused ordinary and repeated uncached race tests, full native, E2E, docs, knowledge, JSON/Definition, diff, scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 5c4253d18283c80ec27b7c2c1f383616eac2a89e`.
- Verified the R90-83 feature and docs-only closure chain, completed state,
  exact closure Vault note, full-index row, MOC link, and current stable UDS
  pathname authority.
- Reviewed the Jul 20 through Aug 9 delivery phases and found no newer missing
  delivery record, stale current stable authority, or unresolved validation
  result that changes priority.
- Parsed all 98 prior task-state JSON files and verified all 88 roadmap rows
  match exactly one Definition with no duplicate identifiers.
- Audited the unfinished queue: R90-59 and R90-75 retain their external
  blockers; R90-84 is the sole dependency-ready increment with complete
  dependency, window, risk, acceptance, validation, and stop records.
- Direct source/test review confirms startup and shutdown unconditionally call
  `os.Remove`, the Unix listener retains automatic close-time unlink behavior,
  and no direct non-socket, symlink, stale-socket, or replacement-path
  regression exists.
- Two initial structural-audit commands failed only in setup: one unpacked a
  one-group regular-expression result as a tuple, and one ran from the Go
  module while using repository-relative documentation paths. The corrected
  repository-root multiset audit passed before selection; neither failed
  command changed files or supplied acceptance evidence.

## Validation

- Run all direct R90-84 receiver pathname regressions normally and at least
  twenty times uncached under the race detector from the `engine` module.
- Run the complete fail-fast chain: `make test`, `make e2e-smoke`,
  `make docs-check`, and `make knowledge-check` after preflighting required
  local tools and the module-selected Go toolchain.
- Parse every task-state JSON and verify the complete roadmap row/Definition
  multisets, duplicate absence, and complete unfinished-item fields.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, configuration, protocol, public API,
  release, and publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range and current stable UDS authority.

## Authority Boundaries

This trigger authorizes only R90-84's bounded UDS pathname classification,
created-socket identity cleanup, direct receiver regressions, current
documentation/task-state reconciliation, validation, commit/push of the
completed increment, and local Vault synchronization. It does not authorize
active/stale peer classification, cross-process locking, protocol/config/API
changes, private input, performance policy, tag/release/image/registry
publication, or workflow dispatch.

## Implementation Checkpoint

- Startup now classifies the configured path with `Lstat`: absence proceeds,
  a Unix socket retains the established removal behavior, and every non-socket
  or symlink returns a clear error without installing a listener or mutating
  the occupant.
- The started `net.UnixListener` disables automatic unlink and the receiver
  captures the created socket's `FileInfo`. Shutdown compares the current
  `Lstat` result with that captured identity and socket type before removal, so
  a displaced pathname is preserved.
- Direct regressions cover regular-file and symlink startup rejection, stale
  Unix-socket restart, ordinary owned cleanup, and replacement regular-file and
  symlink preservation with exact content or link-target comparisons.
- The first twenty-count race run exposed that `Wait` could return after the
  accept loop observed listener close but before the following explicit path
  removal completed. Shutdown now performs identity-bound unlink immediately
  before listener close, keeping removal within the synchronous `Stop` path and
  ensuring accept-loop completion cannot precede owned cleanup. The complete
  focused sequence must be rerun before accepting evidence.

## Validated Evidence

- The corrected five-test pathname sequence passes once normally and twenty
  times uncached under the race detector. The complete receiver package also
  passes uncached under race.
- The symlink-startup regression points at an active Unix socket, so following
  the link would incorrectly reclaim it; the direct rejection preserves the
  exact link and target socket identity.
- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config, and libpcap 1.10.4 were available before complete validation.
- The complete fail-fast repository chain passes after the final direct test:
  both C test binaries, every Go package uncached under the race detector,
  E2E smoke with six packets processed, five alerts generated, and eight rules
  loaded, documentation checks, and all 33 knowledge tests.
- All 99 task-state JSON files parse; all 88 roadmap rows match exactly one of
  88 Definitions with no duplicate identifiers, and every unfinished item
  retains its complete selection fields.
- Every planned local acceptance criterion is satisfied. The only behavioral
  validation deviation was the recorded first repeated-race cleanup-order
  failure, which was corrected and followed by the complete focused and
  repository reruns. The two earlier structural commands were setup-only and
  produced no evidence or mutation.
- `gofmt`, `git diff --check`, exact eight-path scope, staged diff, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration, protocol,
  workflow, release artifact, private-data access, or external mutation was
  added.

## Delivery Results

- Feature commit:
  `8fc16241921dcb2817e2c138e59e36a6ab774b02` (`fix: preserve UDS pathname
  occupants`). It contains exactly the eight validated source, test,
  documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `HEAD == origin/main == 8fc16241921dcb2817e2c138e59e36a6ab774b02`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `5c4253d18283c80ec27b7c2c1f383616eac2a89e..8fc16241921dcb2817e2c138e59e36a6ab774b02`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the delivered startup and shutdown
  pathname-identity boundary without rewriting immutable iteration notes.
  Replaying the identical range preserved exact Vault content hash
  `2eea69c66524a2c9664f896036e9e24fdd8d0269878bffe8a8f7162f2b7fe4a1`.
- R90-84 is complete. R90-59 and R90-75 retain their exact external blockers;
  no dependency-ready local increment remains and no later work was started.

## Stop Conditions

Stop if safe completion requires changing active/stale socket reclamation,
dialing or authenticating an existing peer, cross-process locking,
platform-specific ownership promises, operator data, ambiguous repeated/full
validation, or tag/publication authority.
