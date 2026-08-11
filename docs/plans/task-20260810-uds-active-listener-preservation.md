# Task Plan: R90-88 active UDS listener preservation

## Metadata

- Timestamp: 2026-08-10T19:26:16-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `1dcda25ce728336a984892ae849dffeb1d01b4d6`

## Goal

Preserve a pre-existing connectable Unix listener and its pathname identity
during receiver startup while retaining stale-socket reclamation and making
that removal conditional on the originally classified identity still being
current.

## Scope

- Probe a pre-existing non-symlink Unix socket with a bounded local Unix
  connection attempt before removing it.
- Reject startup clearly when the probe connects, without installing a
  receiver listener, replacing the pathname, or disrupting later service.
- Treat only a connection-refused result as evidence that the captured socket
  is a stale-removal candidate; preserve the pathname on every ambiguous probe
  result.
- Re-inspect the pathname without following symlinks and require the same
  captured filesystem identity immediately before stale removal.
- Directly prove active-listener service/identity preservation, ambiguous-probe
  preservation, concurrent replacement-socket preservation, and retained stale
  reclamation.
- Reconcile receiver lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority for this increment.

## Non-Goals

- Do not authenticate or trust the existing peer, exchange hello/packet/
  heartbeat frames, or change the UDS protocol or capture sender.
- Do not add a configuration field, public API, long-lived peer connection,
  general cross-process lock, PID ownership check, or automatic retry.
- Do not change listener shutdown cleanup, connection limits, read timeout,
  packet handling, reconnect semantics, or non-Unix platform policy.
- Do not access operator/private data or perform tag, release, registry,
  workflow, performance-policy, or other publication mutation.

## Risks

- A successful probe connection can consume peer backlog or handler capacity;
  it must be closed immediately and tests must prove the peer remains usable.
- Treating timeout, permission, or another ambiguous error as stale could
  unlink a live peer that cannot currently accept the probe.
- A stale socket can be replaced between the probe result and removal; failing
  to compare identities can delete the replacement listener.
- A global mutable test seam could introduce race risk; probe injection must
  remain receiver-local and unexported.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Connectable existing listener is preserved | Direct `Start` regression captures the original `FileInfo`, observes the probe connection, checks the clear active-listener error and absent receiver listener, compares `os.SameFile`, then completes a second synchronized service round trip through the original listener |
| Stale Unix socket remains reclaimable | Existing direct stale-socket `Start` regression passes unchanged through the real bounded probe and receiver lifecycle |
| Concurrent replacement is preserved | Direct `Start` regression uses a receiver-local probe seam to replace the captured stale pathname with a live Unix listener before returning refusal, then checks rejection, replacement identity, and service |
| Ambiguous probe failure is fail-closed | Direct `Start` regression injects a non-refusal probe error and proves the original Unix-socket identity remains unchanged with no receiver listener installed |
| Removal is identity-bound | Source re-inspects with `Lstat`, requires Unix-socket mode plus the same device/inode and captured change timestamp, and removes only after connection refusal |
| Existing lifecycle behavior remains compatible | Already-canceled, regular-file, symlink, absent-path, reconnect, active cancellation, capacity, timeout, and owned-cleanup receiver regressions pass |
| Delivery evidence is repository-complete | Direct ordinary and twenty-count uncached race tests, complete receiver race, native, E2E, docs, knowledge, JSON/Definition, diff, scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- Fetched `origin/main` through the already established SSH-over-443 route and
  verified a clean worktree with
  `FETCH_HEAD == HEAD == origin/main == 1dcda25ce728336a984892ae849dffeb1d01b4d6`.
- Verified the R90-87 feature and docs-only closure parent chain and exact
  four/three paths, both immutable Vault notes, full-index rows, MOC links,
  current stable MOC/UDS authority, and reconciled Vault hash
  `1f14b6313b9692b419f9bc4a3c0ee4eb03b9cd0178e0f0032da11ee3e25335ef`.
- Parsed all 102 prior task-state JSON files and verified all 92 roadmap rows
  and Definitions match as complete multisets without duplicates or asymmetry.
- Reviewed 147 Jul 20 through Aug 10 commits across four delivery phases and
  found no missing closure, stale stable authority, or unresolved validation
  result that changes priority.
- Confirmed R90-59 and R90-75 retain their external blockers and R90-88 is the
  sole dependency-ready increment with complete queue fields.
- Confirmed `removeExistingSocket` currently removes every pre-existing Unix
  socket after one `Lstat`; the direct reclamation regression closes its
  listener first, and no direct test preserves a connectable peer or a pathname
  replaced during classification.

## Validation

- Run the active-listener, ambiguous-probe, replacement-identity, and stale
  reclamation regressions normally and at least twenty times uncached under the
  race detector from the `engine` module.
- Run the complete receiver package uncached under the race detector.
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

## Implementation Deviation

- The first focused run passed the active-listener, ambiguous-probe, and stale
  cases but failed the concurrent replacement regression because the local
  filesystem immediately reused the removed socket inode. `os.SameFile` alone
  therefore accepted the replacement, `Start` removed it, and incorrectly
  succeeded.
- Delivery remains blocked while the identity check is strengthened with the
  non-following `Stat_t` change timestamp captured by the original `Lstat`.
  The complete focused sequence must restart after that correction; the failed
  run supplies no acceptance evidence.

## Implementation Checkpoint

- `Receiver.Start` now uses a receiver-local unexported probe with a one-second
  Unix connect bound. A successful connect is closed immediately and returns a
  clear active-listener error; only `ECONNREFUSED` reaches stale removal.
- Missing paths may proceed, while permission, timeout, and every other
  ambiguous probe result returns a wrapped error without removal.
- A connection-refused candidate is re-inspected with `Lstat`; removal requires
  Unix-socket mode, the same device/inode, and the same captured `Stat_t.Ctim`.
  This preserves non-socket, symlink, active-socket, and inode-reusing socket
  replacements.
- The active-listener regression synchronizes on acceptance of the probe,
  verifies the same pathname identity, and then completes a second byte
  round-trip through the original listener. The ambiguous-probe regression
  preserves the original socket identity. The replacement regression uses the
  receiver-local seam to install a live socket before returning refusal, then
  proves its identity and service remain intact.
- The corrected four-test sequence passes once normally and twenty times
  uncached under the race detector. The complete receiver package passes
  uncached under race, covering regular/symlink preservation, pre-cancellation,
  absent/stale startup, reconnect, timeout/capacity, active cancellation, and
  owned cleanup.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes: both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation
  checks, and all 33 knowledge tests.
- All 103 task-state JSON files parse; all 92 roadmap rows match exactly one of
  92 Definitions with no duplicate or asymmetric identifiers, and every
  unfinished item retains its risk, validation, and stop records.
- Every planned acceptance criterion is satisfied. The one behavioral
  validation deviation was the immediate inode-reuse failure recorded above;
  the strengthened generation check and complete corrected focused/repository
  sequences passed.
- `gofmt`, `git diff --check`, exact eight-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.

## Authority Boundaries

This trigger authorizes only R90-88's bounded Unix-socket liveness probe,
identity-bound stale removal, receiver-local test seam, direct regressions,
compatibility validation, current documentation and task-state reconciliation,
commit/push of the completed increment, and local Vault synchronization. It
does not authorize peer authentication/trust, protocol/config/public API or
capture changes, general cross-process ownership, private input, performance
policy, tag/release/image/registry publication, workflow dispatch, or a later
roadmap increment.

## Stop Conditions

Stop if safe completion requires trusting or authenticating the peer, changing
the hello/frame protocol or capture sender, a public/configured probe policy,
general cross-process locking, ambiguous repeated/full validation, private
data, or publication authority.
