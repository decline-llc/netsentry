# Task Plan: R90-71 Go alert-store microbenchmarks

## Metadata

- Timestamp: 2026-08-04T05:33:50-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `e853f8e22d10c98cc9363356272c6d847421514b`

## Goal

Make the Go half of `make bench` exercise bounded primary SQLite alert-store
write and indexed filtered-query paths with production recovery durability,
deterministic cardinality, and allocation reporting, without changing storage
behavior or claiming portable throughput.

## Scope

- Add primary-store single-alert and fixed-size batched `WriteBatch`
  benchmarks.
- Keep the real recovery append, sync, SQLite transaction, and recovery-log
  clear path enabled during every timed write.
- Give every timed alert a deterministic unique event identity while clearing
  benchmark rows outside timed regions so fixture cardinality remains bounded.
- Add fixed-cardinality filtered `Query` benchmarks for exact rule and
  timestamp-range predicates backed by current SQLite indexes.
- Create databases, seed query rows, build inputs, inspect query plans, verify
  row/event/aggregate counts, and clean up outside timed regions.
- Report allocations and execute the new cases directly from the owning Go
  module and through the existing root `make bench` target.
- Update current performance, development, architecture, changelog, roadmap,
  and task-state documentation.

## Non-Goals

- Do not change production SQLite schema, queries, recovery durability,
  aggregation, event identity, lifecycle, sharding, or storage semantics.
- Do not disable or replace the recovery log, write directly to SQLite inside
  timed write cases, or retain duplicate event IDs across timed operations.
- Do not benchmark database creation, schema initialization, cleanup,
  correctness assertions, daily shards, recovery control, API, receiver,
  pipeline, or end-to-end throughput.
- Do not add dependencies, use operator data, allow unbounded fixture growth,
  or publish a host-independent threshold or production throughput claim.
- Do not start R90-72 or publication-blocked R90-59.

## Risks

- Reusing event IDs would benchmark idempotent no-op writes instead of the
  production insert/aggregation path.
- Leaving setup, row cleanup, cardinality checks, or query-plan inspection in
  timed regions would contaminate storage measurements.
- Growing the event ledger or aggregate table with benchmark iteration count
  would make later iterations incomparable and consume unbounded disk.
- Seeding queries through direct SQL could bypass the storage contract the
  benchmark is intended to represent.
- A filtered query can return correct results while SQLite performs a full
  scan, so the selected index must be verified directly before timing.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Primary writes cover one alert and a bounded batch | `BenchmarkStoreWriteBatch/single_alert` and `/batch_32_alerts` call the public `WriteBatch` path and verify event, aggregate-row, and aggregate-count cardinality outside timing |
| Every timed write exercises real work without unbounded rows | Inputs use distinct deterministic timestamps/event identities; rows are cleared outside timing and bounded-cardinality assertions run after every operation |
| Production recovery durability remains enabled | The benchmark store uses the default real recovery log; every successful operation verifies the durable log is empty only after the normal append/sync/commit/clear lifecycle |
| Filtered queries use bounded production fixtures | Query setup writes a fixed 512-alert corpus through `WriteBatch`; exact-rule and timestamp-range cases verify stable totals and page sizes outside timing |
| Query cases are index-backed | Direct `EXPLAIN QUERY PLAN` assertions require the planned rule and timestamp indexes before benchmark timing starts |
| Timings exclude setup and correctness work | Database open/close, seed writes, alert construction, plan inspection, cardinality checks, and cleanup occur while timers are stopped |
| Allocations and entry points are visible | Every case calls `ReportAllocs`; focused module discovery/execution and root `make bench` both execute the named cases |
| Existing behavior and claims remain compatible | Focused alert tests, full native tests, E2E smoke, documentation, and knowledge checks pass; docs retain local-synthetic and non-portable evidence labels |

## Trigger Audit

- Read both delivery skills and the complete rolling roadmap before selection.
- Fetched `origin/main` and verified clean local/remote equality at
  `e853f8e22d10c98cc9363356272c6d847421514b`.
- Verified R90-70 feature `388487da7205e98dd257ee54a1428673141c7457`
  and docs-only closure `e853f8e22d10c98cc9363356272c6d847421514b`,
  including intended paths and current task-state resume instructions.
- Verified both exact R90-70 Vault notes, full-index rows, MOC links, and
  current Makefile, matcher, rule-engine, testing, and MOC authority in the
  sole discovered local Vault.
- Reviewed 151 first-parent commits from Jul 14 through Aug 4 in three phases
  (47/56/48). The R90-70 feature and closure are the only additions since its
  149-commit trigger audit; no missing delivery record, stale stable prose, or
  unresolved validation deviation changes priority.
- Parsed all 85 task-state JSON files and verified all 75 roadmap rows match
  exactly one Definition.
- Audited every unfinished item: R90-71 is the sole dependency-ready local
  increment; R90-72 remains planned behind it, and R90-59 remains blocked on
  exact publication authority. Each retains status, dependency, window, risk,
  acceptance criteria, required validation, and stop condition.
- Inspected the production store open/write/query paths, recovery lifecycle,
  schema indexes, direct query-plan tests, existing benchmark conventions,
  root benchmark orchestration, and current performance claims. The bounded
  benchmark surface requires only a package-owned `_test.go` file and current
  documentation.

## Validation

- Preflight the module-selected Go toolchain plus GCC, Make, Bash, and Python
  versions before the complete fail-fast chain.
- Run `gofmt`, focused uncached alert tests, and focused alert race tests.
- From `engine`, discover and execute only the new store benchmark families
  with ordinary tests disabled, a bounded iteration count, and `-benchmem`.
- Run root `make bench` and verify all storage plus existing C/rule benchmark
  names execute without ordinary Go tests.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, generic-keyword, production-source, dependency, and evidence
  classification review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Added single-alert and 32-alert `WriteBatch` cases that time the public
  primary-store path with its real recovery append/sync/commit/clear sequence.
- Every timed write uses deterministic unique event identity. Direct event,
  aggregate-row, aggregate-count, and empty-log assertions run after timing;
  row cleanup is committed with the timer stopped, bounding the fixture.
- Added exact-rule and timestamp-range `Query` cases over 512 deterministic
  rows seeded once through `WriteBatch`. Each case requires its expected
  SQLite index via `EXPLAIN QUERY PLAN` and checks total/page contents outside
  timing.
- Database creation, fixture construction, seed writes, plan inspection,
  cleanup, and correctness diagnostics remain outside measured regions; only
  a package-owned benchmark `_test.go` file changes executable code.
- A one-iteration direct run discovered and passed all four storage cases with
  allocation output.

## Validation Deviations

- The first uncached complete alert-package run failed the existing
  `TestStorePrimaryWriteActiveCancellationRetainsRecoveryLogForIdempotentRetry`
  five-second return boundary after cancellation.
- The new benchmark functions are excluded from that ordinary-test command;
  no production storage source changed. Delivery remains blocked while the
  exact regression is repeated uncached enough to assess reproducibility and
  until both the alert package and complete native suite rerun cleanly.
- A subsequent 20-count uncached race command reproduced the same timeout.
  Inspection found that the fixture gave SQLite a 5-second busy timeout and
  used an equal 5-second outer return bound; the observed driver return near
  5.1 seconds made those independent deadlines race.
- The test-only fixture now uses a 1-second SQLite busy timeout while retaining
  the 5-second outer assertion, observable post-append/in-use synchronization,
  `context.Canceled` classification, exact log and database checks, and
  idempotent retry. Production code and timeout defaults are unchanged.
- The corrected exact regression passed 20 uncached race executions, followed
  by clean uncached complete alert-package runs both normally and under the
  race detector. Full repository validation remains required before delivery.

## Validated Evidence

- The owning module selected Go 1.25.12 on linux/amd64; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, jq 1.7, and curl 8.5.0 were available before
  the complete validation chain.
- A direct ten-iteration command with ordinary tests disabled discovered and
  passed `BenchmarkStoreWriteBatch/{single_alert,batch_32_alerts}` and
  `BenchmarkStoreQuery/{exact_rule,timestamp_range}` with allocation output.
- Root `make bench` passed all C cases, the four new SQLite store cases, and
  the four existing ten-second rule/matcher cases. The new cases reported
  allocations plus explicit 1/32 alerts per operation, and no ordinary Go
  package tests ran in the benchmark target. The complete root target passed
  again after the final metric placement correction.
- `make test` passed ordinary C tests and every Go package under uncached race
  after the focused cancellation correction. `make e2e-smoke` passed with 6
  packets processed, 5 alerts generated, and 8 rules loaded.
- `make docs-check` and the 33-test `make knowledge-check` pass. All 86
  task-state JSON files parse, all 75 roadmap rows match exactly one
  Definition, and formatting, whitespace, exact nine-path scope, credential,
  sensitive-path, generic-keyword, production-source, dependency, and evidence
  classification reviews pass.
- Source review confirms that timed writes use the real recovery log and
  public `WriteBatch`; event identity changes every operation; rows stay
  bounded by out-of-timer cleanup; query seeds use `WriteBatch`; and plan,
  cardinality, recovery, health, cleanup, and result checks remain outside
  measured regions. No production source, schema, dependency, generated
  evidence, threshold, or operator data changed.

## Delivery Results

- Feature commit:
  `9f29bf32cc3bbc446d03bd2185900c3dae4a84ef` (`test: add Go alert store
  benchmarks`).
- The exact nine-path feature was pushed without force, fetched, and verified
  as both `HEAD` and `origin/main`; the post-fetch 33-test knowledge gate
  passed.
- Exact range
  `e853f8e22d10c98cc9363356272c6d847421514b..9f29bf32cc3bbc446d03bd2185900c3dae4a84ef`
  was synchronized repeatedly with identical hashes to the sole local Vault.
  The iteration note, full index, MOC link, and current SQLite-storage,
  Makefile, testing, and MOC authority are verified; historical iteration
  notes were not rewritten.
- R90-72 is ready but was not started. R90-59 publication remains blocked.

## Authority Boundaries

This trigger authorizes only R90-71 benchmark test code, current public
benchmark documentation, roadmap/task-state reconciliation, repository
validation, commit/push, and local Vault synchronization. It does not authorize
production storage changes, schema/query/recovery optimization, operator data,
performance budgets, publication, tags, releases, images, registries, or
workflow dispatch.

## Stop Conditions

Stop if correct benchmark construction requires disabling recovery durability,
changing production storage behavior or schema, unbounded fixtures, operator
data, a dependency, a host-independent threshold, a production throughput
claim, publication authority, or an ambiguous benchmark/test result that
remains after focused uncached review.
