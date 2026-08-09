# Task Plan: R90-81 post-management-plane delivery audit

## Metadata

- Timestamp: 2026-08-09T06:59:20-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `49ae9eb95c6ff500e3c525bff30d7a13a43b6938`

## Goal

Reconcile the completed R90-80 feature and docs-only closure against current
Git, task-state, fetched remote, Vault, recent delivery history, and direct test
evidence, then restore only the smallest evidence-grounded local reliability
queue without changing runtime or tests.

## Scope

- Verify the exact R90-80 feature/closure chain, intended paths, completed plan
  and state, fetched remote, and both Vault records.
- Review the Jul 20 through Aug 9 delivery phases and record only material
  trends, recurring validation deviations, stale authority, missing records,
  and unresolved risks.
- Trace the recurring receiver timing family from three delivery records to the
  exact Aug 6 idle-capacity failure and current direct test boundary without
  treating less-specific historical records as the same exact failure.
- Classify the observation without claiming a proven production defect, then
  restore one bounded follow-on reliability increment if its acceptance and
  stop boundaries are directly supportable.
- Refresh the active roadmap and task state without starting R90-82.

## Non-Goals

- Do not change runtime, tests, configuration, schema, workflows, release
  gates, benchmark/evidence artifacts, or generated Vault iteration notes.
- Do not diagnose a production timeout defect from isolated test failures or
  select legacy schema, migration, protocol, deployment, or performance policy.
- Do not infer changelog/artifact approval, tag-push authority, GitHub Release
  authority, GHCR publication authority, or comparable-environment evidence.
- Do not access private/external data or perform tag, release, registry,
  workflow, or other publication mutation.

## Risks

- Focused clean reruns do not erase a repeated full-gate reliability signal,
  but they also do not prove production behavior is wrong.
- The current test observes the latest process-wide session after replacement;
  it does not directly expose the handler slot whose release it promises.
- A speculative future queue could reopen completed work or cross product,
  compatibility, performance, or publication boundaries.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-80 is exactly recoverable | Feature/closure SHAs and parents, intended paths, completed plan/state, fetched `HEAD == origin/main`, and two exact Vault notes/index/MOC links |
| Recent history has no concealed blocker | Dated phase-level review of material trends, recurring deviations, unresolved validation, stale authority, and missing delivery records |
| The reliability gap is direct and bounded | Jul 23 generic receiver timing, Jul 29 idle-timeout/broken-pipe, exact Aug 6 idle-capacity failure, and current source/test review preserve evidence specificity, distinguish handler-slot release from shared-session polling, and avoid a runtime-defect claim |
| The forward queue is complete | R90-59 and R90-75 retain exact blockers; R90-82 has status, dependency, window, risk, acceptance, validation, and stop condition |
| No runtime or next increment is started | Exact documentation-only diff and scoped source/test/config/workflow/release review |
| Repository and knowledge records remain valid | All task-state JSON parsing, exact roadmap row/Definition coverage, docs, knowledge, diff, staged-scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 49ae9eb95c6ff500e3c525bff30d7a13a43b6938`.
- Verified the latest R90-80 feature and docs-only closure paths plus both Vault
  notes, full-index rows, MOC links, and current stable management-plane prose.
- Reviewed the Jul 20 through Aug 9 phase history and found no unresolved
  behavior validation result or missing delivery record that changes priority.
- Confirmed three independent full gates recorded receiver timing-family
  deviations followed by focused clean reruns; only the Aug 6 record names the
  current idle-capacity test exactly.
- Audited every unfinished item. R90-59 and R90-75 retain their external
  blockers, so selected R90-81 as the documentation-only smallest safe queue
  unblocker and left R90-82 unstarted.

## Validation

- Parse every task-state JSON and verify exact roadmap row/Definition coverage
  plus complete unfinished-item fields.
- Run `make docs-check` and `make knowledge-check` fail-fast.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, source/test/config/workflow/generated-evidence/release/
  publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range.

## Audit Results

- Verified the exact R90-80 feature/closure parent chain, six-path feature,
  three-path closure, completed plan/state, fetched baseline, both Vault notes,
  both full-index rows, both MOC links, and current stable management-plane
  authority.
- Counted 133 commits from Jul 20 through Aug 9: 56 behavior-like changes, 68
  R90 delivery closures, and 9 other documentation changes across recovery,
  persisted-row/schema contracts, fuzz/storage faults, performance evidence,
  the local-only tag boundary, and management-plane durability.
- Found no missing delivery record, stale current stable authority, or
  unresolved behavioral validation result that changes priority.
- Found completed historical row R90-04a lacked a dedicated Definition despite
  prior exact-coverage claims; restored its quality-only, non-traffic,
  non-release contract from the completed plan/state without rewriting history.
- Preserved historical evidence specificity: Jul 23 names only a receiver
  timing boundary, Jul 29 names the idle-timeout family plus broken-pipe
  symptom, and Aug 6 names the exact idle-capacity test and replacement-session
  observation failure. Each owning increment recorded focused clean reruns.
- Verified the current direct test observes receiver-side connection closure
  before replacement but infers replacement acceptance through a 10-millisecond
  poll of the process-wide latest-session snapshot rather than synchronizing on
  the promised handler-slot release boundary.
- Restored only R90-82 as a planned bounded test-evidence increment. Left
  R90-59 and R90-75 externally blocked and did not start R90-82.
- Wrote the dated R90-81 audit and reconciled the roadmap, plan, and task state
  without changing runtime, tests, configuration, workflows, release evidence,
  or generated Vault notes.

## Validation Checkpoint

- Python 3.12.3, jq 1.7, and GNU Make 4.3 were available before the fail-fast
  documentation validation sequence.
- The first structural check found 86 roadmap rows but 85 Definitions because
  completed historical R90-04a lacked a dedicated Definition. Delivery remained
  blocked until its completed plan/state contract was restored; the complete
  structural check was then rerun successfully.
- All 96 task-state JSON files parse; all 86 roadmap rows have exactly one
  Definition; R90-59, R90-75, R90-81, and R90-82 retain complete unfinished-item
  status, dependency, window, risk, acceptance, validation, and stop fields.
- `make docs-check`, the 33-test `make knowledge-check`, and
  `git diff --check` pass on the complete documentation-only surface.
- The exact four-path scope contains only the dated audit, roadmap, plan, and
  task state. It contains no runtime, test, config, workflow, generated
  evidence, release artifact, or publication path.
- Scoped credential-prefix and operator-sensitive absolute-path review passed.
  No dependency, schema, configuration, tag, release, image, registry, or
  workflow mutation is present.
- A reusable evidence-specificity safeguard was added to the local
  `netsentry-next` skill separately from the repository increment: historical
  package/family deviations cannot be promoted to an identical exact-test claim
  without direct records naming it.

## Authority Boundaries

This trigger authorizes only the R90-81 documentation audit, forward roadmap
and task-state reconciliation, validation, commit/push of those records, and
local Vault synchronization. It does not authorize runtime/test/configuration
changes, a product or compatibility decision, external/private input,
performance policy, tag/release/image publication, or workflow dispatch.

## Stop Conditions

Stop if exact R90-80 evidence is missing or contradictory, the recurring test
evidence cannot be bounded honestly, validation remains ambiguous, or completion
requires runtime/test changes, private/external input, product/performance
policy, publication authority, or starting R90-82.
