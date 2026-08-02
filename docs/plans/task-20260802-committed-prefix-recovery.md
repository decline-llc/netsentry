# Task Plan: R90-62 committed-prefix multi-shard recovery

## Metadata

- Timestamp: 2026-08-02T01:21:31-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `89806508802fd8d8165f9606995d19bba0ef6da0`

## Goal

Directly prove that daily-shard operator recovery remains idempotent when an
earlier shard commits but a later shard fails or the active recovery context is
cancelled, retaining the complete recovery log and sticky emergency state until
an explicit retry completes every event exactly once.

## Scope

- Make daily-shard recovery replay order deterministic without changing alert
  grouping, schema, event identity, or public API behavior.
- Add one direct later-shard failure regression using a real compatible SQLite
  shard held by an independent write transaction after an earlier shard commits.
- Add one direct active-replay cancellation regression synchronized on an
  independently observed earlier-shard commit while the later shard is blocked.
- For both failure paths, compare the complete recovery-log bytes through a
  separate read, require sticky emergency/replay failure state, release the
  real fault, explicitly retry, and prove one event plus aggregate count one per
  input.
- Reconcile the storage test contract, changelog, roadmap, task state, and local
  stable SQLite/testing knowledge after validation.

## Non-Goals

- Do not change the recovery JSONL or SQLite schema, event/aggregation identity,
  API route/status contract, authentication, health fields, or configuration.
- Do not add background retry, automatic cleanup, rollback-by-copy, shard
  transactions spanning databases, or cross-process recovery ownership.
- Do not add a test-only production hook when a real SQLite lock and observable
  committed row can provide deterministic orchestration.
- Do not start R90-63 formatter fuzzing or R90-59 publication work.

## Risks

- Go map iteration can replay shards in an unstable order and make the boundary
  unprovable or flaky.
- A lock that blocks read-only preflight would test the wrong rejection phase;
  the chosen SQLite transaction must allow preflight and block only writable
  replay.
- Cancellation observed only before lifecycle readiness would repeat existing
  coverage instead of reaching active replay after a committed prefix.
- Retry assertions that inspect only total rows could miss aggregate inflation
  or a duplicate event ledger entry.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Deterministic serial replay | Direct code review and repeated focused race runs prove sorted shard paths while retaining per-shard input order |
| Later-shard failure reaches the committed-prefix boundary | A real reserved SQLite writer permits read-only preflight, the earlier shard commits, and the later write returns a replay-phase writable failure |
| Active cancellation reaches the same boundary | The test observes the earlier committed row through an independent read-only handle, then cancels while the later shard remains write-blocked |
| Failure preserves retry authority | Both cases compare complete log bytes after failure and require `emergency`, failed result, replay phase, and no truncation |
| Explicit retry is idempotent | After releasing the lock, retry succeeds; each event ledger ID occurs once and each recovered alert has aggregate count one |
| Existing behavior remains compatible | Focused alert-store race, twenty uncached committed-prefix repetitions, full native tests, E2E smoke, docs, and knowledge gates pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `89806508802fd8d8165f9606995d19bba0ef6da0`.
- Verified the R90-61 feature/closure commits, completed task state, both exact
  Vault notes, full index, MOC links, and stable testing knowledge.
- Reconciled 126 commits since Jul 14 across three delivery phases with no new
  missing delivery record, stale authority, or unresolved validation deviation.
- Parsed all 76 task-state JSON files and verified all 67 roadmap rows match
  exactly one Definition.
- Confirmed R90-62 is the highest-priority ready correctness increment; R90-63
  and R90-64 remain planned, while R90-59 remains authority-blocked.

## Validation

- Focused `go test -race` from the `engine` module for the alert store recovery
  tests, including twenty uncached committed-prefix repetitions.
- Complete native `make test` and `make e2e-smoke`.
- `make docs-check`, `make evidence-check`, and `make knowledge-check`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, staged-diff review, and a scoped sensitive-information
  review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, synchronize both exact full-SHA ranges, and verify their
  Vault notes, full index, MOC links, idempotent replay, and stable knowledge.

## Validated Evidence

- Both direct regressions pass under the race detector and reach writable
  `replay` after an earlier shard event is independently observable.
- A real reserved SQLite writer allows preservation-safe preflight, blocks the
  later writable shard, and requires no production test hook.
- Both fault paths retain the exact complete recovery-log bytes and report
  sticky emergency/replay failure; after lock release, explicit retry leaves
  one ledger event and aggregate count one for each input, then truncates the
  log and returns healthy.
- The two committed-prefix race tests passed five exploratory repetitions and
  twenty final uncached command-level repetitions.
- The complete `internal/alert` race package passed in 44.176 seconds.
- `make test` passed the C parser/sender tests and every Go package under the
  race detector.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded.
- `make docs-check`, `make evidence-check` (16 tests), and `make
  knowledge-check` (33 tests) passed.
- All 77 task-state JSON files parse and all 67 roadmap rows match exactly one
  Definition; `git diff --check` passed.
- Scoped credential-prefix and sensitive-absolute-path scans found no sensitive
  content. The manual generic-keyword matches were JSON decoder/test variable
  names only.

## Execution Deviations

- The first active-cancellation run did not observe the earlier shard within
  its five-second bound while the helper repeatedly opened and closed read-only
  handles. No production recovery error was captured, and the later-shard
  failure case passed.
- The observer was redesigned to pre-create both compatible shards and open one
  independent read-only handle before recovery. This removes repeated-reader
  interference while retaining a real observable commit boundary; the complete
  focused sequence was rerun successfully as recorded above.

## Delivery Results

- Feature commit:
  `981cb1e3a0041301f42629522cff844e04764c6f` (`fix: stabilize
  committed-prefix recovery replay`).
- The eight-path commit was pushed without force, fetched, and verified equal
  to `origin/main`; the post-fetch 33-test knowledge gate passed.
- Exact range
  `89806508802fd8d8165f9606995d19bba0ef6da0..981cb1e3a0041301f42629522cff844e04764c6f`
  was synchronized twice with identical iteration-note, full-index, and MOC
  hashes to the single existing local Vault.
- The generated note, full commit index, MOC link, and manually reconciled
  stable SQLite storage and testing/release notes identify the sorted replay
  and committed-prefix direct evidence.
- R90-63 is ready but was not started; R90-59 publication, tags, releases,
  images, and workflow dispatch remain outside authority.

## Authority Boundaries

The rolling trigger authorizes the bounded R90-62 recovery correctness and test
change plus its documentation, commit/push, and local Vault workflow. It does
not authorize private operator data, destructive repair, automatic recovery,
tags, releases, images, workflow dispatch, or later roadmap increments.

## Stop Conditions

Stop if the reserved SQLite lock rejects read-only preflight, the committed
prefix cannot be observed deterministically, retry inflates or loses data,
required validation is flaky or ambiguous, or completion requires a format,
API, product, private-data, or publication decision.
