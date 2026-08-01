# Task Plan: R90-60 operator-triggered storage recovery

## Metadata

- Timestamp: 2026-08-01T05:12:42-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `59904b79424f80d760d3a9aac9c9617ef1e975cb`

## Goal

Implement the R90-57 fail-closed, operator-triggered restart-free recovery
contract for alert storage, including serialized lifecycle ownership, durable
preflight, idempotent replay/probe, authenticated API control, observability,
and direct fault/concurrency regressions.

## Scope

- Add one context-cancellable store lifecycle gate shared by database reads,
  writes, retention, recovery replay, recovery, and close.
- Add an atomic recovery-state transition so exactly one caller owns
  `emergency → recovering` and duplicate callers fail immediately.
- Preflight the complete recovery log and every existing referenced database or
  shard through read-only handles before any recovery write.
- Replay pending records idempotently, or prove empty-log write capability in a
  rolled-back transaction, then truncate only after complete success.
- Preserve sticky emergency state, full recovery input, and explicit
  pre-/post-writable phase evidence after cancellation or failure.
- Add an authenticated `POST /api/storage/recovery` route, bounded responses,
  verbose health recovery fields, and redacted storage-target audit metadata.
- Reconcile architecture, API, development, changelog, roadmap, task state, and
  local stable knowledge after validation.

## Non-Goals

- Do not add background timers, polling, automatic retry, disk cleanup,
  database repair, evidence deletion, or recovery-log rotation.
- Do not change SQLite or recovery JSONL formats, configuration schema, capture
  protocol, rule behavior, or public release evidence.
- Do not create/move a version tag, publish a GitHub Release or GHCR image,
  dispatch a workflow, or begin R90-59.
- Do not access private operator databases, logs, sidecars, or packet captures.

## Risks

- A lifecycle gate can deadlock shutdown or starve ordinary operations.
- Publishing `recovering` before exclusive ownership can leave a cancelled
  caller stuck or admit a second owner.
- A read-only preflight that misses a referenced shard can cross the writable
  boundary before rejecting incompatible input.
- Cancellation after an earlier shard commit but before truncation must retain
  a replayable complete log without aggregate inflation.
- API authentication must remain mandatory even when ordinary loopback
  mutations are configured without authentication.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| One recovery owner | Direct duplicate-trigger and context-aware lifecycle ownership tests |
| Cancellation-safe lifecycle | Tests cancel before exclusive readiness and during recovery, then prove later operations proceed |
| Preservation boundary | Malformed log/incompatible database or shard tests compare bytes before and after rejection through independent reads |
| Pending/empty success | Direct replay/truncation and rolled-back nonce probe tests prove healthy transition without a durable probe row |
| Failure and retry safety | Writable failure retains the log and emergency state; explicit retry is idempotent |
| Daily/encoded paths | Direct daily-shard and directory-with-spaces recovery regressions |
| Authenticated operator surface | API tests cover disabled auth, invalid token, healthy/in-progress conflict, failure phase, success, method, and audit redaction |
| Observable state | Verbose health exposes recovering phase/start/result while observation never triggers recovery |
| Repository delivery | Focused repeated race, full native, E2E, docs, knowledge, diff, remote, and exact Vault evidence pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `59904b79424f80d760d3a9aac9c9617ef1e975cb`.
- Verified the complete R90-57 feature/closure Git chain, both Vault notes, full
  index, MOC links, and stable SQLite storage note.
- Reconciled 129 commits since Jul 14 at phase level with no missing delivery
  record or unresolved validation deviation.
- Verified all 62 existing roadmap rows matched Definitions; R90-59 was the
  only unfinished row and remained blocked on actual publication authority.
- Identified a material queue gap: R90-57 explicitly deferred runtime recovery
  to another increment, but no implementation row existed.
- Added and selected R90-60 as the highest-priority dependency-ready
  correctness increment; the active horizon now reaches Oct 30.

## Validation

- Focused `go test -race` for `./internal/alert` and `./internal/api` from the
  `engine` module, including twenty uncached recovery-focused repetitions.
- Complete native `make test` and `make e2e-smoke`.
- `make docs-check`, `make evidence-check`, and `make knowledge-check`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- `git diff --check`, staged-diff review, and sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, synchronize both exact full-SHA ranges, and verify their
  Vault notes, full index, MOC links, and stable storage knowledge.

## Validated Evidence

- `go test -race ./internal/alert ./internal/api` passed after the final
  lifecycle/shutdown implementation change.
- Recovery-focused alert-store race tests passed 20 uncached repetitions in
  16.055 seconds; recovery/API/health/audit race tests passed 20 uncached
  repetitions in 1.126 seconds.
- `make test` passed the C parser/sender tests and every Go package under the
  race detector; `make e2e-smoke` passed with 6 packets processed, 5 alerts
  generated, and 8 rules loaded.
- `make docs-check`, `make evidence-check` (16 tests), and `make
  knowledge-check` (33 tests) passed.
- All 75 task-state JSON files parsed, all 63 roadmap rows matched exactly one
  Definition, and `git diff --check` passed.

## Execution Deviations

- The first repeated store-recovery run exposed a test expectation that treated
  `recovery_phase` as terminal history. The implementation intentionally clears
  the active phase on the next healthy operation while retaining
  `last_recovery_result`; the expectation was corrected and the complete
  focused sequence was rerun from the beginning.
- Final review found that shutdown needed to cancel an already claimed recovery
  while it waited for lifecycle exclusion. `Close` now marks closing, cancels
  that owner, waits for release, and has a direct race regression. Recovery
  failure health text was also bounded to a phase summary so private paths are
  not copied into the new recovery result. All validation above ran after these
  changes.

## Authority Boundaries

The rolling trigger authorizes the dependency-ready R90-60 repository change
and normal commit/push/Vault workflow. It does not authorize private data,
automatic cleanup, destructive recovery, tags, release or image publication,
or workflow dispatch.

## Stop Conditions

Stop on ambiguous ownership or preservation behavior, a need for background
retry/cleanup or format migration, failed or flaky required validation, private
data, external publication, or scope beyond this one runtime recovery control.
