# Task Plan: R90-87 post-cancellation delivery audit

## Metadata

- Timestamp: 2026-08-10T18:57:58-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `6ea917e976d71432a4beb72967f73f2abf5c908b`

## Goal

Reconcile the completed R90-86 feature and docs-only closure against current
Git, task-state, fetched remote, Vault, recent delivery history, and receiver
pathname authority; restore only one directly evidenced local follow-on
without changing runtime or tests.

## Scope

- Verify the exact R90-86 feature/closure parent chain, intended paths,
  completed plan/state, fetched remote, and both Vault records.
- Review the Jul 20 through Aug 10 delivery phases and retain only material
  trends, validation deviations, stale authority, missing records, and risks.
- Trace pre-existing Unix-socket handling through current `Start` source,
  direct tests, callers, and public documentation.
- Separate the directly evidenced active-listener preservation gap from peer
  authentication, protocol, cross-process locking, and stale-path recovery
  implementation choices.
- Define at most one bounded follow-on, then refresh the roadmap and task state
  without starting it.

## Non-Goals

- Do not change receiver source/tests, configuration, protocol, public API,
  workflows, release gates, benchmark/evidence artifacts, or immutable Vault
  iteration notes.
- Do not implement active/stale socket classification, dial an existing peer,
  add peer authentication, add cross-process locking, or change cleanup.
- Do not authorize comparable-environment performance scope, remote tag push,
  GitHub Release, GHCR publication, workflow dispatch, or private data access.

## Risks

- An existing socket pathname is not proof that its listener is live; the
  audit must distinguish current unconditional removal from a chosen probe
  implementation.
- Treating every listener as trusted would overstate the evidence and could
  conceal denial-of-service or peer-authentication policy.
- A follow-on that changes the UDS frame contract, capture handshake, or
  multi-process ownership would exceed this documentation-only increment.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-86 is exactly recoverable | Feature/closure SHAs and parents, intended paths, completed plan/state, fetched `HEAD == origin/main`, and two exact Vault notes/index/MOC links |
| Recent history has no concealed blocker | Dated phase counts plus material validation, stale-authority, missing-record, and unresolved-risk review |
| Existing-socket behavior is stated precisely | Current `removeExistingSocket` ordering, checked callers, stale-socket regression setup, absence of an active-listener preservation regression, and public lifecycle prose |
| Any restored local work is direct and bounded | One preservation outcome with explicit peer-authentication, protocol, locking, and publication non-goals |
| The forward queue is complete | R90-59 and R90-75 retain exact blockers; any follow-on has status, dependency, window, risk, acceptance, validation, and stop condition |
| No runtime or follow-on work is started | Exact documentation-only diff and scoped source/test/config/workflow/release review |
| Repository and knowledge records remain valid | Task-state JSON parsing, complete row/Definition multiset comparison, docs, knowledge, diff, staged-scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- The ordinary SSH fetch failed on port 22; SSH-over-443 authentication passed,
  the first fetch ended in a transient broken pipe, and a keepalive retry
  fetched `origin/main` successfully.
- Verified a clean worktree with
  `HEAD == origin/main == 6ea917e976d71432a4beb72967f73f2abf5c908b`.
- Verified the exact R90-86 feature/closure parents and paths, completed state,
  both Vault notes, full-index rows, MOC links, and current stable UDS/MOC
  authority.
- Parsed all 101 prior task-state JSON files and verified all 90 roadmap rows
  match exactly one Definition without duplicate or asymmetric identifiers.
- Confirmed R90-59 and R90-75 retain their complete external blockers and no
  dependency-ready local row exists, selecting R90-87 as the smallest safe
  documentation-only unblocker.
- Confirmed `removeExistingSocket` removes every non-symlink Unix socket after
  `Lstat`; the sole reclamation regression closes its listener before startup,
  and no direct regression preserves a live listener or its pathname identity.

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

This trigger authorizes only R90-87's delivery-evidence audit, forward queue
and task-state reconciliation, documentation validation, commit/push, and
local Vault synchronization. It does not authorize receiver runtime/test
changes, active-peer probing or trust policy, cross-process locking,
private/external input, performance policy, tag/release/image/registry
publication, workflow dispatch, or starting the follow-on.

## Audit Checkpoint

- Verified the exact R90-86 feature/closure parent chain, intended eight/three
  paths, completed plan/state, fetched remote equality, both exact Vault notes,
  both full-index rows and MOC links, and current stable MOC/UDS authority.
- Counted 145 Jul 20 through Aug 10 commits across four delivery phases: 59
  behavior-like changes, 74 `docs: record R90-*` closures, and 12 other
  documentation changes. No unresolved behavioral validation result, stale
  stable authority, or missing delivery record changes priority.
- Recorded the port-22 failure, transient first SSH-over-443 broken pipe, and
  successful keepalive fetch retry without treating cached refs as delivery
  evidence.
- Confirmed current startup unconditionally removes an existing Unix socket,
  while the direct reclamation regression closes its listener first and proves
  only stale-path behavior.
- Refreshed the rolling horizon through Nov 8 and restored only R90-88 behind
  this audit: preserve a connectable listener's pathname/service and make
  stale removal identity-bound. Peer trust, authentication, protocol, locking,
  runtime/test work, and publication remain outside R90-87.

## Validation Checkpoint

- The first validation command was rejected before execution because its
  temporary-file cleanup form was disallowed. It changed nothing and supplies
  no evidence; the replacement process-substitution sequence ran from the
  beginning without temporary files.
- All 102 task-state JSON files parse and all 92 roadmap rows match exactly one
  of 92 Definitions with equal raw counts, no duplicates, and no asymmetric
  identifiers.
- Direct marker-order validation places R90-86 selection, implementation,
  validation, completion, R90-87 selection, and R90-87 validation in order.
- R90-59, R90-75, R90-87, and R90-88 retain complete status, dependency,
  forecast window, risk, acceptance, required-validation, and stop or unblock
  fields.
- `make docs-check`, all 33 `make knowledge-check` tests, and
  `git diff --check` pass after the chronology correction.
- Every acceptance criterion maps to the completed evidence with only the
  recorded non-mutating command-rejection deviation. Exact four-path scope
  plus credential, sensitive-path, source/test, config/workflow,
  generated-evidence, release, and publication review passes; no runtime,
  test, configuration, workflow, generated evidence, release artifact,
  private-data access, or external mutation was added.

## Stop Conditions

Stop if R90-86 evidence is missing or contradictory, active-listener
preservation cannot be bounded without a new product or compatibility decision,
validation remains ambiguous, or completion would start runtime/test,
private-data, performance-policy, or publication work.
