# NetSentry Post-Pathname Delivery Audit — 2026-08-09

## Audit Baseline

- Repository baseline: `79f6250de30c3128ecaec31e81ae19eecc9109d8`.
- Branch and remote: clean `main`; fetched `origin/main` matched the baseline.
- Audit period: 2026-07-20 through 2026-08-09.
- Delivery authority: current source and direct tests, Git commits, completed
  task plans/states, fetched remote refs, public documentation, and exact-range
  local Vault records.
- External boundary: R90-59 and R90-75 remain blocked; this audit authorizes no
  runtime/test change, socket compatibility decision, publication action,
  performance policy, workflow dispatch, or external evidence access.

## Method

1. Reconcile the R90-84 feature and docs-only closure with their parents,
   intended paths, completed plan/state, fetched branch, and two Vault records.
2. Review the previous three delivery weeks at phase level and retain material
   trends, validation deviations, missing records, stale authority, and risks.
3. Compare roadmap delivery-history order with the exact Git parent chain and
   move only mutable prose needed to restore chronology.
4. Trace already-canceled receiver startup through current source, checked
   callers, direct tests, and the asynchronous-lifecycle evidence rule.
5. Separate a deterministic pre-readiness cancellation boundary from
   active/stale peer, cross-process, protocol, and product policy.
6. Audit every unfinished item for status, dependency, forecast window, risk,
   acceptance criteria, validation, blocker, and stop condition.
7. Restore only one bounded local follow-on and do not start it.

## Delivery History Review

| Period | Commits | Main delivery theme |
| --- | ---: | --- |
| Jul 20–26 | 61 | recovery and persisted-row contracts, schema compatibility, and read-only preservation |
| Jul 27–Aug 2 | 38 | recovery structure/encoding, candidate refresh, operator recovery, and committed-prefix retry |
| Aug 3–6 | 22 | sustained fuzz closure, storage fault lifecycle, and benchmark evidence/baseline |
| Aug 7–9 | 20 | local-only tag boundary, management-plane durability, receiver lifecycle, and delivery audits |
| **Total** | **141** | one dependency-ordered correctness, recovery, evidence, release-boundary, management-plane, receiver, and audit sequence |

The period contains 58 behavior-like commits, 72 `docs: record R90-*`
closures, and 11 other documentation changes. Previously recorded storage,
benchmark, SSH, Vault, roadmap-structure, receiver-timing, and pathname-cleanup
deviations were resolved within their owning increments. No failed or ambiguous
behavioral validation remains open. The only stale repository authority found
is mutable chronology: correct R90-84 completion facts were inserted before
R90-82 completion while the chronological tail retained R90-84's pre-delivery
checkpoint.

## R90-84 Delivery Reconciliation

| Evidence | Audit result |
| --- | --- |
| Git chain | Feature `8fc16241921dcb2817e2c138e59e36a6ab774b02` is the child of R90-83 closure `5c4253d18283c80ec27b7c2c1f383616eac2a89e`; docs-only closure `79f6250de30c3128ecaec31e81ae19eecc9109d8` is the feature's child and fetched `origin/main`. |
| Intended paths | The feature changes exactly eight source, test, documentation, roadmap, plan, and state paths. The closure changes only the three R90-84 delivery-record paths. |
| Plan and state | The completed R90-84 plan/state records direct regular-file, symlink-to-active-socket, stale-socket, owned-cleanup, and replacement-path evidence; repeated/full validation; fetched remote; stable-prose reconciliation; and idempotent Vault replay without stale commit/push/sync instructions. |
| Vault | Both iteration notes exist with exact ranges; both full-index rows and MOC links exist. Stable MOC and UDS prose records the delivered pathname identity boundary, feature/closure SHAs, external blockers, and cross-process/peer-policy limits. |

R90-84 is exactly recoverable. The roadmap completion paragraph retained the
same correct commit, validation, range, Vault, and blocker facts; moving it
after the R90-84 validation checkpoint repairs only mutable ordering and does
not rewrite Git, task-state, or immutable Vault evidence.

## Cancellation-Before-Readiness Evidence

`Receiver.Start` currently classifies/removes the configured pathname, creates
and publishes the listener/socket identity, initializes capacity, applies mode,
and only then launches a goroutine waiting for `ctx.Done()`. An already-canceled
context can therefore cause filesystem/listener work before asynchronous
shutdown observes cancellation.

All current receiver cancellation regressions call `Start` with a live context
and cancel only after the method returns, with or without active connections.
The checked production and full-engine callers also supply live contexts and do
not rely on starting with an already-canceled context. No direct test calls
`Start` after cancellation or proves that an absent pathname remains absent and
a pre-existing Unix-socket identity remains unchanged at that boundary.

This supports one bounded follow-on: reject an already-canceled context with an
error matching `context.Canceled` before any pathname inspection, removal, or
listener creation; directly preserve both an absent path and a pre-existing
Unix-socket identity; and retain ordinary live-context startup plus existing
post-readiness cancellation behavior. It does not decide active/stale peer
status, add path locking, change the UDS protocol/configuration, or diagnose a
production failure.

## Forward Queue

| ID | Status | Dependency | Remaining condition |
| --- | --- | --- | --- |
| R90-59 | Blocked on remote-publication authority | R90-59a plus explicit tag-push, GitHub Release, and GHCR authority | Resolve the recorded changelog/artifact boundary and explicitly authorize the external publication actions. |
| R90-75 | Blocked / pending evidence; non-blocking | R90-74 plus comparable-environment evidence and explicit budget scope | Supply matched evidence and choose portable, same-host-only, or observation-only product/SLO scope. |
| R90-86 | Planned behind this audit | R90-85 | After R90-85 delivery, reject already-canceled receiver startup before filesystem mutation with direct absent-path and existing-socket preservation evidence. |

Every queue row has exactly one Definition and every unfinished item retains a
forecast window, risk, acceptance criteria, required validation, and stop or
unblock condition. R90-86 is the only bounded local follow-on and remains
unstarted in this trigger.

## Audit Conclusion

R90-84's feature and docs-only closure are complete, fetched, synchronized,
and consistent with current stable knowledge. The recent delivery sequence has
no missing record or unresolved behavioral validation failure. R90-85 repairs
the mutable completion ordering and restores R90-86 as a direct
cancellation-before-readiness preservation increment without claiming a
production defect or changing runtime/tests. R90-59 and R90-75 retain their
exact external blockers.
