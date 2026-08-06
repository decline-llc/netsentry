# Task Plan: R90-74 repeated single-host benchmark baseline

## Metadata

- Timestamp: 2026-08-06T01:48:39-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`

## Goal

Record an observation-only baseline from at least five uncached complete C/Go
benchmark evidence captures taken sequentially from one exact clean commit and
unchanged local environment, retain every reviewed raw sample, and produce
recomputable median, inclusive interquartile-range, and coefficient-of-
variation summaries without applying a performance threshold.

## Scope

- Collect five default-parameter `make benchmark-evidence` samples in one
  isolated detached worktree at the clean fetched R90-73 closure commit.
- Require identical full commit/tree identity, clean status, environment and
  toolchain fingerprint, command parameters, complete 6-C/8-Go surface, and
  evidence classification across all samples.
- Extend the versioned benchmark evidence API and CLI to aggregate reviewed
  samples and independently validate the summary by recomputing it from every
  referenced raw artifact.
- Summarize every reported performance/allocation metric with count, minimum,
  maximum, median, inclusive Q1/Q3, IQR, arithmetic mean, sample standard
  deviation, coefficient of variation, and range as a percentage of median.
- Commit the five path-redacted raw JSON samples, their SHA-256-bound aggregate
  JSON, a concise Markdown interpretation, tests, docs, plan/state, and roadmap
  checkpoint.

## Non-Goals

- Do not change any benchmark fixture, timed region, runtime/storage behavior,
  dependency, configuration, release gate, or R90-73 capture schema.
- Do not cherry-pick implementation changes into the clean sampling worktree;
  aggregate only after all five immutable R90-73 samples exist.
- Do not treat descriptive variation as pass/fail, select a numeric budget,
  compare another host, claim production capacity, or infer an SLO.
- Do not use pressure/PCAP/corpus traffic, private operator input, hostnames,
  usernames, credentials, or sensitive absolute paths.
- Do not start R90-75 or R90-59 and do not tag, release, publish, dispatch a
  workflow, or mutate an image/registry.

## Risks

- A dirty or changing sampling tree, toolchain/environment drift, parameter
  drift, or Go test caching would invalidate comparability.
- Thermal/background-load noise can dominate five sequential local samples;
  the evidence must retain raw results and report rather than conceal spread.
- Aggregating missing or differently keyed metrics can silently compare unlike
  cases unless the API proves the complete metric surface is identical.
- Committed raw command output can leak paths or host metadata unless every
  sample passes the R90-73 contract plus a separate sensitive-data review.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Five uncached complete samples share one exact source | Detached clean worktree at `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`; every sample records identical full commit/tree, `clean=true`, environment fingerprint, parameters, and `-count=1` Go command |
| Every raw result remains reviewable | Five committed path-redacted R90-73 JSON artifacts retain stdout/stderr plus parsed 6-C/8-Go metrics and pass independent validation |
| Summary is deterministic and recomputable | Tested API binds sample basenames and SHA-256 values, computes defined descriptive statistics, and validation reproduces byte-equivalent semantic JSON from the raw artifacts |
| Metric surfaces cannot drift silently | Direct regressions reject fewer than five samples, dirty Git state, commit/tree/environment/parameter/classification mismatch, missing metrics, duplicate sample names, tampered digests, and edited aggregate values |
| Interpretation remains observation-only | JSON/Markdown explicitly deny thresholds, production derivation/capacity, cross-host portability, SLO, release, tag, and publication authority |
| Repository delivery remains compatible | Focused tests, direct aggregate validation, complete native, shell/Python/docs/evidence/knowledge checks, JSON/roadmap coverage, diff/scope, and sensitive-information review pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, complete active
  roadmap, repository knowledge contract, latest R90-73 plan/state, capture
  implementation/tests, and current performance guidance before selection.
- Fetched `origin/main` through GitHub SSH-over-443 and verified a clean
  `HEAD == origin/main` baseline at
  `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`.
- Verified exact R90-73 feature/closure commits, both Vault iteration notes,
  full-index rows, MOC links, and current stable MOC/Makefile/testing authority
  in the sole discovered local Vault.
- Reviewed the Jul 14 through Aug 6 delivery phases and found no missing
  delivery record, stale authority, or unresolved validation deviation; the
  R90-73 port-22 fetch interruption is closed by fetched SSH-over-443 evidence.
- Parsed all 88 task-state JSON files and verified all 78 roadmap rows match one
  Definition. Every unfinished item retains status, dependency, forecast
  window, risk, acceptance criteria, required validation, and stop condition.
- Selected R90-74 as the sole dependency-ready increment. R90-75 and R90-59
  remain blocked on their recorded external evidence/product or publication
  authority conditions.

## Validation

- Preflight Git, compiler, Go/module toolchain, Make, Python, disk space, and
  exact default benchmark parameters before starting the five-run sequence.
- Validate each raw capture immediately and compare immutable Git/tree,
  environment, parameter, command, surface, and classification fields before
  accepting the next sample.
- Focused aggregation and capture regressions through `python3 -m unittest
  scripts.test_benchmark_evidence` and direct CLI aggregate/revalidation.
- Complete `make shell-check`, `make python-check`, `make docs-check`, `make
  evidence-check`, `make knowledge-check`, and native `make test`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, staged-scope review, raw-artifact hash verification,
  and anchored credential, hostname, operator-path, private-input, threshold,
  production-claim, and publication-authority scans.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify each exact full-SHA Vault range.

## Sampling and Implementation Results

- Preflight created one isolated detached worktree at clean fetched commit
  `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`, tree
  `14aeceed857497b6c81a20b8cb06c9f187607a8f`, with Go 1.25.12,
  Linux/amd64, GCC 13.3.0, Make 4.3, Python 3.12.3, and sufficient temporary
  disk space. The worktree was removed after the reviewed artifacts were copied
  and verified.
- Five sequential default-parameter captures ran from 08:50:12Z through
  09:02:35Z with 100,000 C iterations and 10-second Go benchtime. Every sample
  immediately passed the R90-73 validator and matched the exact Git/tree,
  clean-state hashes, environment/toolchain fingerprint, parameters, command
  vectors, classification, and complete 6-C/8-Go surface before the next run.
- The five retained raw sample SHA-256 values are recorded in the aggregate
  JSON and Markdown. Direct byte comparison proved the checked-in artifacts
  equal the frozen accepted outputs; a separate scan found no home, temporary,
  hostname, username, credential, key, or private-input value.
- The versioned API now requires at least five individually valid samples,
  rejects duplicate basenames and every context mismatch, binds raw SHA-256
  values, and computes 43 metric series with minimum, maximum, median,
  inclusive Q1/Q3/IQR, mean, sample standard deviation, coefficient of
  variation, and range relative to median.
- Independent baseline validation reconstructs the complete semantic JSON from
  raw samples and rejects a changed digest or statistic. `make
  benchmark-baseline-check` exposes that gate and `make evidence-check` runs it
  before the evidence regression suite.
- The highest observed coefficient of variation is 11.252396% for
  `BenchmarkMatcherMatch/no_hit` `ns/op` (3,485–4,470 ns/op; median 3,561).
  No metric has an order-of-magnitude outlier. The record reports this spread
  without a pass/fail threshold or causal claim.

## Validation Checkpoint

- All 20 focused capture/aggregation tests pass, including five-sample
  descriptive statistics, fewer-than-five and duplicate rejection, dirty Git,
  commit/tree/environment/parameter/metric drift, raw-reparse enforcement,
  digest/statistic tampering, and path-safe observation-only Markdown.
- The checked-in `baseline.json` independently recomputes from all five
  checked-in raw samples and exposes exactly 43 metric series.
- `make shell-check`, `make python-check`, `make docs-check`, the 42-test
  `make evidence-check`, the 33-test `make knowledge-check`, complete native
  `make test` with uncached Go race execution, and `git diff --check` pass in
  one final fail-fast rerun.
- All 89 task-state JSON files parse, all 78 roadmap rows match exactly one
  Definition, and every unfinished Definition retains goal, risk, required
  validation, and stop fields.
- Final staged-scope and sensitive-information review remain the delivery
  boundary. No source, evidence, or documentation changed after the successful
  complete validation rerun except this validation record.

## Execution Deviations

- The first focused aggregation test run reported two fixture assertion
  failures because the dirty-state and changed-parameter fixtures violated the
  existing single-sample contract before reaching the intended cross-sample
  boundary. Production code and real samples were unaffected. The fixtures
  were corrected to remain individually valid while differing across samples,
  and the complete 20-test sequence passed.
- The five default captures each took roughly two and a half minutes, longer
  than the earlier `10x` validation run but consistent with the planned `10s`
  Go benchtime. Parameters and serial execution were preserved rather than
  shortened or parallelized.
- The first complete native gate failed in the unchanged receiver test
  `TestStartIdleTimeoutReleasesConnectionCapacity` because the replacement
  session was not observed within its existing bound. The exact test then
  passed 20 uncached race executions. No benchmark path or unrelated receiver
  source was changed; the entire fail-fast chain was restarted and the clean
  full native race suite passed.
- No sample, environment, metric, sensitive-data, evidence-classification, or
  authority deviation remains unresolved.

## Delivery Results

- Feature commit:
  `77e1ec005e077e1e66049a5a4eb809afd87fa23c` (`test: record
  single-host benchmark baseline`). Its exact 15 paths contain the aggregation
  API/tests, five reviewed raw samples, aggregate/report, Make/docs, roadmap,
  plan, and task state only.
- The feature was pushed without force through GitHub SSH-over-443, fetched,
  and verified as both `HEAD` and `origin/main`. The post-fetch 33-test
  knowledge gate and checked-in raw-sample baseline recomputation passed.
- Exact range
  `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f..77e1ec005e077e1e66049a5a4eb809afd87fa23c`
  was synchronized repeatedly with identical generated note/index/MOC hashes
  to the sole discovered local Vault.
- The generated iteration note, full index, MOC link, and manually reconciled
  stable MOC, Makefile/build, and testing/release notes identify the five-sample
  observation boundary, 43 recomputable metric series, highest observed CV,
  and absence of threshold or portable/production authority. Exact-range replay
  preserved those stable edits and hashes.
- R90-75 remains blocked on comparable-environment evidence plus explicit
  product/SLO budget scope, and R90-59 remains blocked on exact publication
  authority. No next increment is ready or started.

## Authority Boundaries

This trigger authorizes only R90-74 clean local benchmark sampling, versioned
descriptive aggregation/validation, reviewed evidence/docs/task/roadmap paths,
validation, commit/push, and the local Vault workflow. It does not authorize
private or external data, benchmark/runtime semantic changes, numeric gates,
cross-host/production/SLO claims, R90-75 decisions, R90-59 publication, tags,
releases, images, registries, or workflow dispatch.

## Stop Conditions

Stop if any sample is dirty, cached, incomplete, malformed, sensitive, or
differs in commit/tree/environment/toolchain/fixture/parameters; if the raw
metric surface cannot be aggregated exactly; if variation has an unexplained
order-of-magnitude outlier or ambiguous cause; if required validation is
flaky; or if completion needs private/external input, a threshold/product
decision, publication authority, or a later increment.
