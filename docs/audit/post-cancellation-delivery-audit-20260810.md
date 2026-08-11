# NetSentry Post-Cancellation Delivery Audit — 2026-08-10

## Audit Baseline

- Repository baseline: `6ea917e976d71432a4beb72967f73f2abf5c908b`.
- Branch and remote: clean `main`; fetched `origin/main` matched the baseline
  after the documented SSH-over-443 retry path.
- Audit period: 2026-07-20 through 2026-08-10.
- Delivery authority: current source and direct tests, Git commits, completed
  task plans/states, fetched remote refs, public documentation, and exact-range
  local Vault records.
- External boundary: R90-59 and R90-75 remain blocked; this audit authorizes no
  runtime/test change, peer-trust or protocol decision, publication action,
  performance policy, workflow dispatch, or external evidence access.

## Method

1. Reconcile the R90-86 feature and docs-only closure with their parents,
   intended paths, completed plan/state, fetched branch, and two Vault records.
2. Review the previous three delivery weeks at phase level and retain material
   trends, validation deviations, missing records, stale authority, and risks.
3. Trace pre-existing Unix-socket startup through current source, checked
   callers, direct tests, and public lifecycle documentation.
4. Separate directly evidenced pathname preservation from peer trust,
   authentication, protocol, and cross-process ownership policy.
5. Audit every unfinished item for status, dependency, forecast window, risk,
   acceptance criteria, validation, blocker, and stop condition.
6. Refresh the rolling horizon, restore only one bounded local follow-on, and
   do not start it.

## Delivery History Review

| Period | Commits | Main delivery theme |
| --- | ---: | --- |
| Jul 20–26 | 61 | recovery and persisted-row contracts, schema compatibility, and read-only preservation |
| Jul 27–Aug 2 | 38 | recovery structure/encoding, candidate refresh, operator recovery, and committed-prefix retry |
| Aug 3–6 | 22 | sustained fuzz closure, storage fault lifecycle, and benchmark evidence/baseline |
| Aug 7–10 | 24 | local-only tag boundary, management-plane durability, receiver lifecycle, and delivery audits |
| **Total** | **145** | one dependency-ordered correctness, recovery, evidence, release-boundary, management-plane, receiver, and audit sequence |

The period contains 59 behavior-like commits, 74 `docs: record R90-*`
closures, and 12 other documentation changes. Previously recorded storage,
benchmark, SSH, Vault, roadmap-structure, receiver-timing, pathname-cleanup,
and command-working-directory deviations were resolved within their owning
increments. This trigger's ordinary port-22 fetch failed and its first
authenticated SSH-over-443 fetch ended with a transient broken pipe; the
keepalive retry fetched `origin/main` successfully. No failed or ambiguous
behavioral validation, missing delivery record, or stale stable authority
remains open.

## R90-86 Delivery Reconciliation

| Evidence | Audit result |
| --- | --- |
| Git chain | Feature `97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3` is the child of R90-85 closure `ab63ee3ef53fdb7a764ca0863dac36580d0318fa`; docs-only closure `6ea917e976d71432a4beb72967f73f2abf5c908b` is the feature's child and fetched `origin/main`. |
| Intended paths | The feature changes exactly eight source, test, documentation, roadmap, plan, and state paths. The closure changes only the three R90-86 delivery-record paths. |
| Plan and state | The completed R90-86 plan/state records both direct pre-canceled preservation regressions, live/post-readiness compatibility, repeated/full validation, the corrected module-relative command, fetched remote, stable-prose reconciliation, and idempotent Vault replay without stale commit/push/sync instructions. |
| Vault | Both iteration notes exist with exact ranges; both full-index rows and MOC links exist. Stable MOC and UDS prose records the pre-filesystem cancellation boundary, exact feature/closure references, external blockers, and unchanged active/stale-peer policy. |

R90-86 is exactly recoverable. Its one non-mutating focused-command setup
error was corrected before tests supplied evidence; the complete focused and
repository chains passed and no unresolved deviation carries into selection.

## Pre-Existing Unix-Socket Evidence

`Receiver.Start` calls `removeExistingSocket` before `net.Listen`.
`removeExistingSocket` uses non-following `Lstat`, rejects regular files and
symlinks, but removes every pathname whose mode is a Unix socket without
checking whether an active listener still owns it. The subsequent listener
therefore replaces either a stale socket or the published pathname of a live
peer.

The direct `TestStartReclaimsPreExistingUnixSocket` regression explicitly
disables auto-unlink and closes its listener before calling `Start`; it proves
stale pathname reclamation only. Searches across receiver tests and checked
production/integration callers found no direct startup over a live Unix
listener and no assertion that its pathname identity and service remain
available. Public architecture/development prose accurately says that
pre-existing Unix-socket reclamation remains established behavior and does not
claim active-owner detection.

This supports one bounded follow-on: reject startup while the existing Unix
socket is currently connectable, preserve its pathname identity and continued
service, retain stale-socket reclamation, and make any eventual stale removal
conditional on the classified pathname identity still being current. A
successful connection proves reachability, not peer identity or trust. The
follow-on therefore does not authenticate the peer, change the hello/frame
protocol or capture sender, introduce general cross-process locking, or claim
a production incident.

## Forward Queue

| ID | Status | Dependency | Remaining condition |
| --- | --- | --- | --- |
| R90-59 | Blocked on remote-publication authority | R90-59a plus explicit tag-push, GitHub Release, and GHCR authority | Resolve the recorded changelog/artifact boundary and explicitly authorize the external publication actions. |
| R90-75 | Blocked / pending evidence; non-blocking | R90-74 plus comparable-environment evidence and explicit budget scope | Supply matched evidence and choose portable, same-host-only, or observation-only product/SLO scope. |
| R90-88 | Planned behind this audit | R90-87 | After R90-87 delivery, preserve a connectable existing listener and identity-bound pathname replacement while retaining stale-socket reclamation. |

Every queue row has exactly one Definition and every unfinished item retains a
forecast window, risk, acceptance criteria, required validation, and stop or
unblock condition. The horizon now runs through Nov 8. R90-88 is the only
bounded local follow-on and remains unstarted in this trigger.

## Audit Conclusion

R90-86's feature and docs-only closure are complete, fetched, synchronized,
and consistent with current stable knowledge. The recent delivery sequence has
no missing record or unresolved behavioral validation failure. R90-87 restores
R90-88 as a direct active-listener pathname-preservation increment while
keeping reachability distinct from trust and leaving runtime/tests unchanged.
R90-59 and R90-75 retain their exact external blockers.
