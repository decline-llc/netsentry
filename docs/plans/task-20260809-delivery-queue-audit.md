# Task Plan: R90-76 post-tag delivery and forward-queue audit

## Metadata

- Timestamp: 2026-08-09T04:28:45-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc`

## Goal

Reconcile the completed local v0.1.1 tag delivery and the preceding two to four
weeks of correctness, fuzz, storage-fault, and performance work against current
Git, task-state, roadmap, test, remote, and Vault evidence, then restore the
smallest evidence-grounded local queue through October 31 without exercising
the blocked publication or performance-budget authorities.

## Scope

- Verify the R90-59a feature and docs-only closure commits, exact local tag,
  absent remote tag, completed plan/state, and both pushed Vault ranges.
- Review recent delivery at phase level and record only material deviations,
  stale authority, missing delivery records, and unresolved risks.
- Audit current source, direct tests, Make targets, public limitations, and
  checked-in evidence for the next bounded local correctness or reliability
  work.
- Reconcile the active task-state baseline with fetched `origin/main`.
- Add a dated audit and refresh the dependency-ordered roadmap with complete
  definitions for only evidence-supported future increments.
- Deliver documentation, task-state, roadmap, and Vault reconciliation without
  starting any later increment.

## Non-Goals

- Do not change runtime, tests, benchmark or evidence behavior, configuration,
  release gates, workflows, or the historical release candidate.
- Do not move, recreate, or push `v0.1.1`; do not create a GitHub Release,
  publish GHCR, or dispatch a workflow.
- Do not invent changelog approval, artifact equivalence, comparable-host
  benchmark evidence, a product SLO, or a performance threshold.
- Do not access private traffic or external operator data, rewrite immutable
  iteration notes, or start an increment introduced by this audit.

## Risks

- The latest active task state records the R90-59a feature SHA while the fetched
  branch includes its later docs-only closure; treating the feature SHA as the
  current baseline would leave stale resume authority.
- A local signed tag can be mistaken for publication readiness even though the
  remote tag is absent and current tag workflows publish two unauthorized
  external outputs.
- Historical gap prose can reopen completed work or promote broad product-scale
  limitations into an unbounded ready increment.
- Generated Vault iteration notes are immutable history; only current stable
  prose should be reconciled when authority changed.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-59a is exactly recoverable | Feature/closure SHAs and parents, intended paths, signed local tag object/target, completed plan/state, and fetched `HEAD == origin/main` |
| Publication authority remains fail-closed | Direct remote-tag absence, workflow trigger review, and unchanged GitHub Release/GHCR prohibition |
| Vault evidence is complete and current | Both exact iteration notes, full-index rows, MOC links, and stable release authority in the sole discovered Vault |
| Recent history has no concealed delivery deviation | Dated phase-level commit and feature/closure review with material validation and delivery deviations recorded |
| The stale active baseline is reconciled | Current task-state and roadmap checkpoint identify fetched closure SHA `5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc` |
| The forward queue is complete and bounded | Every unfinished item has status, dependency, window, risk, acceptance criteria, required validation, and stop condition; externally blocked work stays blocked |
| No later work is started | Diff is documentation/task-state only and contains no source, test, config, workflow, generated evidence, tag, or external mutation |
| Repository and knowledge records remain valid | All task-state JSON parses; exact row/Definition coverage; docs, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the rolling
  roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc`.
- Verified the signed local `v0.1.1` tag still targets exact candidate
  `78cd78574e03c8f73ff68248eed2c409d6bca406` while the remote tag is absent.
- Verified both R90-59a Vault notes, full-index rows, MOC links, and current
  stable local-only release authority in the sole discovered Vault.
- Found R90-59 and R90-75 still blocked on their recorded external conditions;
  no dependency-ready row remained.
- Selected this documentation-only queue audit as the smallest safe unblocker;
  no runtime, publication, or performance-budget work is authorized.

## Validation

- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Verify every unfinished definition names status, dependency, forecast window,
  risk, acceptance criteria, required validation, and stop condition.
- Run `make docs-check` and `make knowledge-check` fail-fast.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, source/test/config/workflow, generated-evidence, tag, and
  external-mutation review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Audit Results

- The R90-59a feature and closure form the expected exact parent chain and each
  changes only its roadmap, plan, and task-state paths. Both Vault notes,
  full-index rows, MOC links, and stable release authority are present.
- Fresh local tag verification accepts object
  `f1a38ecb82b9c63e8411f3df040bdea84e985dd8`, exact peeled candidate
  `78cd78574e03c8f73ff68248eed2c409d6bca406`, and the recorded ED25519 signer.
  The remote tag and GitHub Release remain absent; GHCR inspection reports no
  `v0.1.1` image; both workflows still publish from a pushed `v*` tag.
- The Jul 14 through Aug 7 history contains 161 commits across four recorded
  phases: 68 behavior/refactor/test commits, 73 `docs: record R90-*` closures,
  and 20 other documentation commits. Every material validation, transport,
  Vault-path, benchmark-orchestration, and planning deviation was resolved or
  remains explicitly bounded; no missing recent delivery record changes the
  next priority.
- Direct source and test review found that rule CRUD and explicit reload each
  replace the full seed file and active immutable snapshot without a shared
  transaction lock. Existing direct tests cover individual operations, not
  concurrent interleavings, so successful operations can derive from the same
  old snapshot and lose an earlier change.
- Rule and suppression temporary-file replacement both lack explicit short-
  write rejection, file sync, containing-directory sync, and direct lifecycle-
  fault evidence. Suppression mutation itself is already serialized, so its
  durability boundary remains separate from rule transaction concurrency.
- R90-77 through R90-80 now cover rule transaction serialization, rule-file
  durability, suppression-file durability, and a final management-plane audit.
  Product-scale protocols, legacy-schema removal, R90-59 publication, and
  R90-75 performance policy remain outside the local ready queue.
- Aggregate Go coverage observed during the audit is 80.2%. It informed direct
  test review but is not treated as proof of a defect or a numeric acceptance
  threshold.

## Validation Checkpoint

- All 91 task-state JSON files parse; all 84 roadmap rows match exactly one
  Definition; all seven unfinished items contain status, dependency, window,
  risk, acceptance criteria, required validation, and stop condition.
- `make docs-check`, the 33-test `make knowledge-check`, and
  `git diff --check` pass on the complete documentation surface.
- The exact six-path diff contains only architecture/development guidance, the
  dated audit, roadmap, plan, and task state. It contains no source, test,
  benchmark, evidence, config, workflow, generated Vault note, or release
  artifact path.
- Credential-prefix and operator-sensitive absolute-path scans have no matches.
  Direct tag-state checks retain the exact local object/candidate and remote
  absence.
- No reusable workflow deviation was identified. The existing preflight,
  audit, one-increment, validation, delivery, and two-commit closeout rules
  matched the work, so the `netsentry-next` skill is unchanged.

## Authority Boundaries

This trigger authorizes only the R90-76 documentation audit, active-state and
roadmap reconciliation, repository validation, commit/push of those records,
and the local Vault workflow. It does not authorize runtime or test changes,
private/external input, a performance-budget decision, tag push, GitHub
Release, GHCR publication, workflow dispatch, or any later increment.

## Stop Conditions

Stop if completion requires source/test behavior changes, private or external
input, a product or compatibility decision, changelog/artifact approval,
performance-budget authority, tag or publication mutation, rewriting immutable
evidence, or starting a later increment.
