# Task Plan: R90-85 post-pathname delivery audit

## Metadata

- Timestamp: 2026-08-09T12:36:37-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `79f6250de30c3128ecaec31e81ae19eecc9109d8`

## Goal

Reconcile the completed R90-84 feature and docs-only closure against current
Git, task-state, fetched remote, Vault, recent delivery history, and receiver
cancellation authority; repair the mutable roadmap chronology and restore only
one directly evidenced local follow-on without changing runtime or tests.

## Scope

- Verify the exact R90-84 feature/closure parent chain, intended paths,
  completed plan/state, fetched remote, and both Vault records.
- Review the Jul 20 through Aug 9 delivery phases and retain only material
  trends, validation deviations, stale authority, missing records, and risks.
- Move the misplaced R90-84 completion paragraph to its chronological delivery
  position after the R90-84 validation checkpoints; preserve its facts.
- Trace cancellation before receiver readiness from current `Start` ordering,
  checked callers, direct tests, documentation, and the established
  asynchronous-lifecycle evidence rule.
- Define one bounded follow-on only if the existing gap can be stated without
  choosing active/stale peer, cross-process, protocol, or product policy.
- Refresh the roadmap and task state without starting the follow-on.

## Non-Goals

- Do not change receiver source/tests, configuration, protocol, public API,
  workflows, release gates, benchmark/evidence artifacts, or immutable Vault
  iteration notes.
- Do not reopen R90-84 behavior, choose active/stale socket or peer policy, add
  cross-process locking, or infer a production defect from an untested edge.
- Do not authorize comparable-environment performance scope, remote tag push,
  GitHub Release, GHCR publication, workflow dispatch, or private data access.

## Risks

- Leaving delivery completion before its prerequisite history makes the roadmap
  materially misleading even when its commit facts are correct.
- A missing cancellation regression is not itself proof of a runtime defect;
  the audit must separate direct source ordering from unobserved production
  behavior.
- A follow-on that specifies active/stale socket handling or concurrent path
  replacement would cross boundaries preserved by R90-84.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-84 is exactly recoverable | Feature/closure SHAs and parents, intended paths, completed plan/state, fetched `HEAD == origin/main`, and two exact Vault notes/index/MOC links |
| Recent history has no concealed blocker | Dated phase counts plus material validation, stale-authority, missing-record, and unresolved-risk review |
| Roadmap delivery history is chronological | R90-84 completion paragraph appears after its selection, implementation, deviation, and successful validation checkpoints without fact changes |
| Any restored local work is direct and bounded | Current `Start` context-watcher ordering, all checked callers, absence of a pre-canceled direct regression, and explicit non-goals establish at most one cancellation-before-readiness increment |
| The forward queue is complete | R90-59 and R90-75 retain exact blockers; any follow-on has status, dependency, window, risk, acceptance, validation, and stop condition |
| No runtime or follow-on work is started | Exact documentation-only diff and scoped source/test/config/workflow/release review |
| Repository and knowledge records remain valid | Task-state JSON parsing, complete row/Definition multiset comparison, docs, knowledge, diff, staged-scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 79f6250de30c3128ecaec31e81ae19eecc9109d8`.
- Verified the exact R90-84 feature/closure parents and paths, completed state,
  both Vault notes, full-index rows, MOC links, and current stable UDS/MOC
  authority.
- Parsed all 99 task-state JSON files and verified all 88 roadmap rows match
  exactly one Definition without duplicate identifiers.
- Confirmed R90-59 and R90-75 retain their complete external blockers and no
  dependency-ready local row exists, selecting R90-85 as the smallest safe
  documentation-only unblocker.
- Found the R90-84 completion paragraph at line 2973 before the R90-82
  completion and the file tail still ending at R90-84's pre-delivery checkpoint;
  the facts are correct but the mutable delivery chronology is stale.
- Confirmed current `Receiver.Start` installs the listener and only then starts
  its context watcher; every direct receiver cancellation test cancels after
  successful `Start`, and the checked production/integration callers provide a
  live context without relying on already-canceled startup.

## Validation

- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, including raw counts, duplicate absence, and asymmetric IDs.
- Run `make docs-check`, `make knowledge-check`, and `git diff --check`.
- Review exact documentation-only scope plus credential, sensitive-path,
  source/test/config/workflow/generated-evidence, release, and publication
  boundaries.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range and current stable authority.

## Authority Boundaries

This trigger authorizes only R90-85's delivery-evidence audit, mutable roadmap
chronology repair, forward queue and task-state reconciliation, documentation
validation, commit/push, and local Vault synchronization. It does not authorize
receiver runtime/test changes, active/stale peer policy, cross-process locking,
private/external input, performance policy, tag/release/image/registry
publication, workflow dispatch, or starting the follow-on.

## Audit Checkpoint

- Verified the exact R90-84 feature/closure parent chain, intended eight/three
  paths, completed plan/state, fetched remote equality, both exact Vault notes,
  both full-index rows and MOC links, and current stable MOC/UDS authority.
- Counted 141 Jul 20 through Aug 9 commits across four delivery phases: 58
  behavior-like changes, 72 `docs: record R90-*` closures, and 11 other
  documentation changes. No unresolved behavioral validation result or missing
  delivery record changes priority.
- Moved the unchanged R90-84 completion paragraph from before R90-82 completion
  to directly after R90-84 successful validation. No commit, validation, range,
  Vault, blocker, or immutable evidence fact changed.
- Confirmed current `Start` publishes listener/path state before launching its
  context watcher, every direct receiver cancellation regression cancels only
  after successful startup, and checked callers do not rely on pre-canceled
  startup.
- Restored R90-86 as the sole bounded local follow-on: an already-canceled
  context must fail before pathname mutation, directly preserving an absent
  path and a pre-existing Unix-socket identity. Active/stale peer policy,
  cross-process locking, post-readiness behavior, and runtime/test work remain
  outside this audit.

## Validation Checkpoint

- All 100 task-state JSON files parse and all 90 roadmap rows match exactly one
  of 90 Definitions with equal raw counts, no duplicates, and no asymmetric
  identifiers.
- Direct marker-order validation proves R90-82 completion, R90-83 completion,
  R90-84 selection, R90-84 completion, and R90-85 selection now occur in that
  chronological order.
- R90-59, R90-75, R90-85, and R90-86 retain complete status, dependency,
  forecast window, risk, acceptance, required-validation, and stop or unblock
  fields.
- `make docs-check`, all 33 `make knowledge-check` tests, and
  `git diff --check` pass.
- Every acceptance criterion maps to the completed evidence with no deviation.
  Exact four-path staged scope plus credential, sensitive-path, source/test,
  config/workflow/generated-evidence, release, and publication review passes;
  no runtime, test, configuration, workflow, generated evidence, release
  artifact, private-data access, or external mutation was added.

## Delivery Results

- Documentation feature commit:
  `73a5f03ab685408d802084ce5864d33cfa3bf03b` (`docs: audit post-pathname
  delivery`). It contains exactly the four validated audit, roadmap, plan, and
  task-state paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `HEAD == origin/main == 73a5f03ab685408d802084ce5864d33cfa3bf03b`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `79f6250de30c3128ecaec31e81ae19eecc9109d8..73a5f03ab685408d802084ce5864d33cfa3bf03b`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to the corrected delivery chronology
  and planned/unstarted R90-86 boundary without rewriting immutable iteration
  notes. Replaying the identical range preserved exact Vault content hash
  `8b9cb2e84a8335f7f9f025ca63234577a36580639b18c3459fb77c3fae09dda8`.
- R90-85 is complete. R90-86 is the next ready local increment and remains
  unstarted; R90-59 and R90-75 retain their external blockers.

## Stop Conditions

Stop if R90-84 evidence is missing or contradictory, the chronology cannot be
repaired without altering immutable evidence, cancellation-before-readiness
requires product or compatibility policy, validation remains ambiguous, or
completion would start runtime/test, private-data, performance-policy, or
publication work.
