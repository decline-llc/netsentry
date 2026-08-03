# Task Plan: R90-66 primary write interruption recovery

## Metadata

- Timestamp: 2026-08-03T05:04:12-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `667cedc72dec9ce58fc7c12aff3be2d37e9ab835`

## Goal

Directly prove that an ordinary primary-store `WriteBatch` interrupted by real
SQLite contention or active context cancellation after its durable recovery
append but before transaction commit leaves no partial database mutation,
retains the complete log, and succeeds exactly once on one explicit retry.

## Scope

- Add one direct real-SQLite contention regression for an ordinary primary
  write while an independent connection owns the write reservation.
- Add one direct active-cancellation regression synchronized on both the exact
  durable recovery record and an in-use primary SQLite connection.
- Open and reuse one independent read-only observer before each writer starts
  to prove event and aggregate state at the interrupted boundary.
- Prove one retry after releasing contention persists the event once, leaves
  aggregate count one, clears the recovery log, and restores healthy state.
- Update current storage-testing guidance, changelog, roadmap, and task state.

## Non-Goals

- Do not add a production failpoint, filesystem fault injection, new storage
  abstraction, dependency, recovery format, schema migration, or cross-process
  ownership contract.
- Do not change append open/write/sync/close behavior or post-commit log
  clearing; those remain R90-67 and R90-68.
- Do not weaken emergency semantics, automatically delete recovery evidence,
  or add background retry.
- Do not access private/external data, change release gates, create a tag,
  publish an artifact/image, or start R90-59 or R90-67.

## Risks

- Cancellation after observing only the file can still occur before SQLite
  owns a connection; the test must also observe the store connection in use.
- Reopening observers in a polling loop can perturb SQLite lock scheduling;
  each test must open one read-only observer before the writer and reuse it.
- A fixed sleep can turn scheduling luck into false evidence; readiness must be
  synchronized on observable state with a bounded diagnostic deadline only.
- Retrying the same alert appends a duplicate recovery record by design; event
  identity must prove the two records produce one event and aggregate count one.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Real contention occurs after durable append | Independent `BEGIN IMMEDIATE` owner, exact expected JSONL bytes, SQLite busy/locked error, and degraded health |
| Active cancellation reaches the promised boundary | Exact complete log plus `DB.Stats().InUse == 1` while the write owner is held, followed by `context.Canceled` |
| Neither interruption partially mutates SQLite | One pre-opened read-only observer per case reports zero matching event and alert rows before retry |
| The complete recovery log is retained | Byte-for-byte comparison with the exact canonical one-record JSONL after both failures |
| One explicit retry is idempotent | After releasing the lock, one retry returns nil, observer reports one event row and one aggregate row with count one, log is empty, and health is `ok` |
| Existing behavior remains compatible | Both focused tests pass twenty uncached race runs, the full alert package and native suite pass under race, E2E passes, and docs/knowledge checks pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `667cedc72dec9ce58fc7c12aff3be2d37e9ab835`.
- Verified the R90-65 feature and closure Git chain, completed state, exact
  Vault notes/index/MOC links, and corrected current stable authority.
- Reviewed 141 commits from Jul 14 through Aug 3 across three phases; no new
  delivery-record, remote, Vault, or release-authority deviation changes
  priority.
- Parsed all 80 task-state JSON files and verified all 71 roadmap rows match
  exactly one Definition.
- Audited all four unfinished rows: R90-59 remains blocked, R90-66 is the sole
  dependency-ready increment, and R90-67/R90-68 correctly depend on it.
- Inspected `WriteBatch`, `writeBatchToDB`, existing pre-cancel coverage, and
  R90-62 active multi-shard recovery tests. The direct ordinary primary
  after-append contention/cancellation boundary is not currently tested.

## Validation

- Preflight the repository-required Go, GCC, Make, Bash, and Python tools.
- Run both new primary interruption regressions under `go test -race -count=1`.
- Run the exact focused pair twenty times uncached under the race detector.
- Run the complete `internal/alert` package under `go test -race -count=1`.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast in the safe ordinary-build order.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and a scoped
  sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Execution Deviations

The first focused sequence invoked `gofmt` with a repository-root path while
the command was already running from the Go module. Formatting failed before
the tests, so the later results from that sequence were discarded. The entire
formatting-plus-focused sequence was rerun from module-relative paths.

The corrected sequence then exposed a real error-contract gap: active
cancellation while SQLite waited on the independent write reservation returned
the driver's `interrupted (9)` diagnostic without retaining
`context.Canceled` in the error chain. The bounded correction joins an active
context error with every transaction-stage driver error, preserving both
classifications. No schema, recovery format, retry policy, or ownership
behavior changed.

## Validated Evidence

- Repository-required Go 1.25.12, GCC 13.3.0, GNU Make 4.3, Bash 5.2.21,
  and Python 3.12.3 were available before the complete validation chain.
- The corrected focused pair passed under the race detector, followed by five
  exploratory and twenty final uncached race repetitions.
- Both direct cases used a real independent `BEGIN IMMEDIATE` owner, retained
  the exact canonical one-record JSONL, and exposed zero matching event and
  aggregate rows through one observer opened before the writer.
- Active cancellation waited for the exact synced log plus one in-use store
  connection and returned an error matching `context.Canceled` while retaining
  the SQLite diagnostic.
- One retry after lock release produced one event, one aggregate row with
  count one, an empty recovery log, and healthy state in both cases.
- The complete `internal/alert` package passed under uncached race in 39.818
  seconds. `make test` passed ordinary C tests and every Go package under
  uncached race.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded; documentation and the 33-test knowledge gate also passed.

## Delivery Results

- Feature commit:
  `260d53d6b5804ca37dc83b083486d429a5e9c983` (`fix: preserve primary write
  cancellation recovery`).
- The exact eight-path feature was pushed without force, fetched, and verified
  as both `HEAD` and `origin/main`; the post-fetch 33-test knowledge gate
  passed.
- Exact range
  `667cedc72dec9ce58fc7c12aff3be2d37e9ab835..260d53d6b5804ca37dc83b083486d429a5e9c983`
  was synchronized twice with identical hashes to the single local Vault.
- The generated iteration note, full index, MOC link, and updated stable
  SQLite/testing authority are verified; historical iteration notes were not
  rewritten.
- R90-67 is ready but was not started. R90-68 remains planned, and R90-59
  publication remains blocked.

## Authority Boundaries

This trigger authorizes only R90-66 direct primary write-interruption tests,
the smallest correction required if those direct tests expose an in-scope
defect, supporting documentation, repository validation, commit/push, and the
local Vault workflow. It does not authorize R90-67/R90-68 implementation,
private input, publication, tags, releases, images, registries, or workflow
dispatch.

## Stop Conditions

Stop if deterministic proof requires a production failpoint, fixed sleeps,
repeated observer opens, cross-process ownership support, automatic evidence
cleanup, a recovery/schema migration, destructive host faults, publication
authority, or an ambiguous result that remains after focused uncached review.
