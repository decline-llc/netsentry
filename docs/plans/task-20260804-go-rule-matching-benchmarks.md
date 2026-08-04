# Task Plan: R90-70 Go rule-matching microbenchmarks

## Metadata

- Timestamp: 2026-08-04T04:48:56-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `fffea8c7d030b84f836137fb22e94ae552a8e677`

## Goal

Make the existing Go half of `make bench` execute deterministic
Aho-Corasick and full rule-engine matching hot paths for no-hit and multi-hit
payloads, with meaningful allocation reporting and without changing production
matching behavior or claiming a portable performance threshold.

## Scope

- Add package-owned Aho-Corasick `Match` benchmarks for deterministic no-hit
  and multi-hit byte payloads.
- Add full rule-engine `Match` benchmarks for deterministic no-hit and
  multi-hit packets across payload, IP, and port rules.
- Build matchers, rules, immutable engine state, packets, and Base64 payloads
  outside timed regions.
- Verify expected pattern indices and alert rule IDs outside timed regions.
- Keep benchmark results live without package-global mutable sinks and report
  allocations and processed payload bytes.
- Prove direct Go-module benchmark discovery and execution as well as execution
  through the root `make bench` target.
- Make the Go benchmark invocation skip ordinary tests so long-running
  all-package benchmarks cannot interfere with unrelated timing-sensitive
  package tests; `make test` remains the correctness/race gate.
- Update current performance, development, architecture, changelog, roadmap,
  and task-state documentation.

## Non-Goals

- Do not change Aho-Corasick, rule-engine, rule-validation, Base64-decoding, or
  alert-building production behavior.
- Do not weaken or skip any test in `make test`; isolate only the dedicated
  benchmark target's Go invocation from ordinary test execution.
- Do not benchmark rule construction/reload, mutable shared output, SQLite,
  receiver/pipeline work, or end-to-end packet throughput in this increment.
- Do not add dependencies, external corpora, generated evidence artifacts, or
  host-independent numeric thresholds.
- Do not describe one local synthetic run as production throughput, a release
  gate, or a cross-host regression budget.
- Do not start R90-71, R90-72, or publication-blocked R90-59.

## Risks

- Building matchers, JSON rule config, engine state, packet fixtures, or Base64
  payloads inside benchmark loops would contaminate matching measurements.
- A package-global result sink would create mutable shared output and make
  future parallel execution unsafe; an unused local result could instead be
  optimized away.
- A multi-hit payload can reach the Aho-Corasick candidate stage but fail
  protocol, port, direction, offset, or depth rechecks, so alert IDs must be
  verified through the complete engine path.
- Equal rule priorities or duplicate pattern hits could make order-based
  assertions brittle; fixture correctness must compare the complete expected
  set without depending on incidental ordering.
- `make bench` uses a ten-second Go benchtime for every sub-benchmark, so the
  full required run is deliberately more expensive than focused discovery.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Aho-Corasick exposes stable no-hit and multi-hit hot paths | `BenchmarkMatcherMatch/no_hit` and `/multi_hit` use fixed patterns/payloads and verify the exact returned index set outside timed regions |
| The complete rule engine exposes stable no-hit and multi-hit hot paths | `BenchmarkEngineMatch/no_hit` and `/multi_hit` use one immutable mixed rule set, pre-encoded packets, and verify the exact alert rule-ID set outside timed regions |
| Timings exclude fixture construction and correctness work | Matcher creation, JSON config, `Reload`, Base64 preparation, initial/final assertions, and all diagnostics occur outside `ResetTimer`/after `StopTimer` |
| Results remain observable without shared mutable state | Each benchmark retains only a local final result and passes it to `runtime.KeepAlive`; no package-global sink or parallel mutation is added |
| Allocations and work size are visible | Every sub-benchmark calls `ReportAllocs` and `SetBytes` before the timed loop |
| Both supported entry points execute the benchmarks | Focused `go test -run '^$' -bench ... -benchmem` from `engine`, followed by root `make bench` using the same explicit ordinary-test exclusion |
| Existing behavior remains compatible | Focused rule tests, full native tests, E2E smoke, documentation, and knowledge checks pass |
| Claims remain evidence-bounded | Public docs identify local synthetic benchmark coverage without publishing host-independent thresholds or production throughput claims |

## Trigger Audit

- Read both delivery skills and the complete rolling roadmap before selection.
- Fetched `origin/main` and verified clean local/remote equality at
  `fffea8c7d030b84f836137fb22e94ae552a8e677`.
- Verified the R90-69 feature `1a612273dd49a216710441dc2eae9e0e2b4d16f7`
  and docs-only closure `fffea8c7d030b84f836137fb22e94ae552a8e677`,
  including their exact parents and intended paths.
- Verified both exact R90-69 Vault notes, full-index rows, MOC links, and current
  stable Makefile/testing/MOC authority in the single local Vault.
- Reviewed 149 commits from Jul 14 through Aug 4 in three phases (47/56/46).
  The two commits added since the prior audit are the expected R90-69 feature
  and closure; no missing record, stale authority, or unresolved validation
  deviation changes priority.
- Parsed all 84 task-state JSON files and verified all 75 roadmap rows match
  exactly one Definition.
- Audited the unfinished queue: R90-70 is the sole dependency-ready increment;
  R90-59 remains blocked on exact publication authority, and R90-71/R90-72
  retain unfinished internal dependencies.
- Inspected the root benchmark target, Aho-Corasick matcher, full rule-engine
  matcher and validation, packet/rule models, existing focused tests, and
  current public performance claims. No `Benchmark*` function exists at the
  selected baseline, and no production-code change is needed to add the
  bounded benchmark surface.

## Validation

- Preflight repository-required Go, GCC, Make, Bash, and Python versions before
  the complete fail-fast chain.
- Run `gofmt` and focused uncached tests for `./internal/rule/...`, then repeat
  that focused package set under the race detector.
- From `engine`, explicitly discover and execute the two benchmark families
  only, with tests disabled, a bounded iteration count, and allocation output.
- From the repository root, run the dedicated `make bench` target and verify
  all four Go sub-benchmark names execute after the C benchmarks without
  selecting ordinary tests.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, generic-keyword, and production-source-change review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Added `BenchmarkMatcherMatch/no_hit` and `/multi_hit` with one immutable
  case-insensitive automaton, fixed local payloads, exact index-set checks, byte
  reporting, allocation reporting, and a benchmark-local retained result.
- Added `BenchmarkEngineMatch/no_hit` and `/multi_hit` with one immutable mixed
  payload/IP/port rule state. Packet fields and Base64 payload previews are
  prepared before each timed region, and exact alert rule-ID sets are checked
  before and after timing without depending on priority order.
- Matcher construction, JSON config encoding, engine reload, fixture setup,
  correctness assertions, and diagnostics remain outside timed loops; no
  production source or dependency changed.
- Focused uncached rule tests and a ten-iteration direct benchmark run passed,
  discovering all four sub-benchmark names with allocation and byte output.
- Current performance/development/architecture/changelog guidance now records
  the implemented local synthetic boundary while reserving SQLite work for
  R90-71 and any portable budget decision for R90-72.

## Validation Deviations

- The first full `make bench` run passed. Two later exact runs executed all
  four new benchmark cases but failed because the existing
  `TestStorePrimaryWriteActiveCancellationRetainsRecoveryLogForIdempotentRetry`
  timed out while ordinary tests were running concurrently with long benchmark
  packages.
- The exact storage test passed immediately uncached and then passed 20
  uncached race-detector repetitions, so matcher behavior and the storage test
  itself did not reproduce a standalone failure.
- Repetition under `make bench` showed that treating the second timeout as a
  transient would be unsafe. The root target needs `-run '^$'` on its Go
  benchmark command so it measures benchmarks only; the already required full
  `make test` run continues to own ordinary and race correctness validation.
- This correction changes only benchmark orchestration. It does not change or
  weaken production code, the storage regression, or the full native test
  gate.

## Validated Evidence

- The owning module selected Go 1.25.12; GCC 13.3.0, GNU Make 4.3, Bash
  5.2.21, and Python 3.12.3 were available before the complete validation
  chain.
- Both rule packages passed uncached focused tests and uncached race tests. A
  direct 100-iteration command with ordinary tests disabled discovered and
  executed `BenchmarkEngineMatch/{no_hit,multi_hit}` and
  `BenchmarkMatcherMatch/{no_hit,multi_hit}` with byte and allocation output.
- After adding the same ordinary-test exclusion to the root target, the final
  `make bench` run passed all C cases and all four ten-second Go sub-benchmarks.
  Non-benchmark Go packages exited without running their ordinary tests.
- `make test` then passed ordinary C tests and every Go package under uncached
  race on the final source. `make e2e-smoke` passed with 6 packets processed,
  5 alerts generated, and 8 rules loaded.
- `make docs-check` and the 33-test `make knowledge-check` pass. All 85
  task-state JSON files parse, and all 75 roadmap rows match exactly one
  Definition.
- `gofmt`, `git diff --check`, exact ten-path scope review, staged diff review,
  and scoped credential, sensitive-path, generic-keyword, and production-source
  reviews pass. The only Go additions are benchmark `_test.go` files, and the
  Makefile change affects only benchmark orchestration.
- The source proves that construction, JSON encoding, rule reload, packet and
  Base64 setup, assertions, diagnostics, and `runtime.KeepAlive` stay outside
  timed regions. No dependency, generated evidence, threshold, or production
  performance claim was added.

## Delivery Results

- Feature commit:
  `388487da7205e98dd257ee54a1428673141c7457` (`test: add Go rule matching
  benchmarks`).
- The exact ten-path feature was pushed without force, fetched, and verified as
  both `HEAD` and `origin/main`; the post-fetch 33-test knowledge gate passed.
- The helper's first exact-range call failed before writing because its
  documented default Vault path was stale. Repeating the unchanged range with
  the sole discovered Vault supplied explicitly succeeded.
- Exact range
  `fffea8c7d030b84f836137fb22e94ae552a8e677..388487da7205e98dd257ee54a1428673141c7457`
  was synchronized repeatedly with identical hashes. The iteration note, full
  index, MOC link, and stable Makefile, Aho-Corasick, rule-engine, testing, and
  MOC authority are verified; historical iteration notes were not rewritten.
- R90-71 is ready but was not started. R90-72 remains planned, and R90-59
  publication remains blocked.

## Authority Boundaries

This trigger authorizes only R90-70 benchmark test files, current public
benchmark documentation, roadmap/task-state reconciliation, repository
validation, commit/push, and local Vault synchronization. It does not authorize
production matcher optimization, semantic changes, SQLite benchmarks,
external/private input, performance budgets, publication, tags, releases,
images, registries, or workflow dispatch.

## Stop Conditions

Stop if correct benchmark construction requires a production semantic change,
a dependency, shared mutable output, an external corpus, SQLite/store work, a
host-independent threshold, a production throughput claim, publication
authority, or an ambiguous benchmark/test result that remains after focused
uncached review.
