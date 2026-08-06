# Task Plan: R90-73 versioned local benchmark evidence capture

## Metadata

- Timestamp: 2026-08-06T00:47:16-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `b20845a8b7b4584e9cfa49aadc5ee663c17a2fe2`

## Goal

Add one versioned, directly tested command that executes the complete existing C
and Go microbenchmark surface and writes a machine-readable, path-redacted local
evidence envelope containing exact Git/tree state, environment and toolchain
fingerprints, command parameters, raw output, parsed metrics, and an explicit
local-synthetic classification without changing any measured benchmark.

## Scope

- Add a Python API and CLI that runs the existing C `capture` benchmark target
  and Go `testing.B` benchmark command without changing their fixtures or timed
  boundaries.
- Require all six named C cases and all eight named Go sub-benchmarks; reject
  duplicate, malformed, unknown, failed, or partial output.
- Record the full HEAD and tree SHAs, branch, clean/dirty state, OS/kernel/
  architecture, Go/GCC/Make/Python versions, explicit benchmark parameters,
  exit status, elapsed time, redacted raw output, and parsed numeric metrics.
- Redact repository, home, and temporary-directory paths by default in commands,
  output, and environment strings.
- Add a root Make target, fixture-driven parser/contract tests, and public
  performance/evidence documentation.
- Perform one bounded direct complete-surface capture into a temporary path as
  validation; do not commit its numeric result as the R90-74 baseline.

## Non-Goals

- Do not modify C or Go benchmark fixtures, timing regions, runtime behavior,
  storage semantics, configuration, dependencies, or release gates.
- Do not apply a threshold, compare samples, aggregate repeated results, or
  claim production throughput, cross-host comparability, or an SLO.
- Do not access an external corpus or private operator data and do not include
  hostnames, usernames, absolute paths, credentials, or raw packet data.
- Do not start R90-74, R90-75, R90-59, a tag, release, image, registry change,
  or workflow dispatch.

## Risks

- A permissive parser could accept a successful exit code while omitting a
  benchmark or accepting duplicate/unknown cases.
- Go benchmark suffixes and optional metrics require strict enough parsing to
  prove completeness without coupling to one CPU count or metric ordering.
- Build and test output, Git status, or tool versions can embed sensitive local
  paths even when the benchmark result lines do not.
- A long default Go benchtime would make direct validation needlessly costly;
  the validation run must remain bounded while exercising the identical cases.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| One versioned command captures the existing complete surface | Root Make target and direct bounded run execute the unchanged C target plus Go `-run '^$' -bench=.` path |
| Every established case is required | Fixture tests cover all six C and eight Go cases, duplicates, unknowns, malformed metrics, failed packages, and missing cases |
| Git and environment context is exact and safe | Unit tests assert full commit/tree SHAs, clean/dirty state, OS/kernel/architecture/toolchain fields, explicit parameters, and default path redaction |
| Raw and parsed evidence is reviewable | JSON retains redacted raw stdout/stderr, command status/timing, and typed per-case metrics with a schema version |
| Claims remain bounded | Evidence records `local_synthetic`, denies production derivation, thresholds, release, tag, and publication authority, and docs repeat those limits |
| Existing behavior remains compatible | Focused Python tests, bounded direct capture, shell/Python/docs/evidence/knowledge gates, complete native tests, and diff review pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, repository knowledge
  contract, complete active roadmap, latest plan/state, and R90-73 definition.
- Fetched `origin/main` and verified a clean local/remote baseline at
  `b20845a8b7b4584e9cfa49aadc5ee663c17a2fe2`.
- Verified both R90-72 feature/closure Vault notes, full-index rows, MOC links,
  and current stable MOC/Makefile/testing authority in the sole discovered
  local Vault.
- Reviewed the Jul 14 through Aug 5 delivery phases and found no missing
  delivery record, stale release authority, or unresolved validation result;
  the earlier benchmark orchestration and Vault-path deviations remain closed.
- Audited every unfinished roadmap item. R90-73 is the sole highest-priority
  dependency-ready increment; R90-74 remains planned, while R90-75 and R90-59
  remain blocked on their recorded external inputs or authority.
- Inspected the existing root/capture Make targets, all C/Go benchmark bodies,
  current performance audit, evidence contract, and analogous versioned Python
  evidence tooling before selecting the implementation shape.

## Validation

- Focused `python3 -m unittest scripts.test_benchmark_evidence`.
- One bounded complete capture with reduced explicit C iteration count and Go
  benchtime, written outside the repository and validated through the same API.
- `make shell-check`, `make python-check`, `make docs-check`, `make
  evidence-check`, `make knowledge-check`, and complete native `make test`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, staged-scope review, and anchored credential-prefix,
  hostname/path, private-input, and authority-claim scans.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify each exact full-SHA Vault range.

## Implementation Results

- `scripts/benchmark_evidence.py` provides versioned capture and validation
  APIs plus a CLI. The root `make benchmark-evidence` target invokes it with
  explicit C iterations and Go benchtime while preserving the existing C
  target and Go `-run '^$' -bench=. -benchmem` execution boundary.
- The evidence contract requires all six C parser/formatter/socket cases and
  all eight Go matcher/rule/store sub-benchmarks. It rejects missing,
  duplicate, unknown, malformed, failed, iteration-mismatched, or write-error
  output and requires typed metrics for each exact case.
- The envelope records full commit/tree SHAs, branch, clean/dirty state and
  opaque status/diff hashes, environment/toolchain fields and their recomputed
  fingerprint, explicit command parameters, redacted raw stdout/stderr,
  command timing/status, and parsed metrics.
- Validation reparses the retained raw results and requires them to equal the
  stored metrics, binds both command argument vectors to the recorded
  parameters, and rejects unredacted home or temporary paths.
- The schema marks every accepted result as local synthetic microbenchmark
  evidence and explicitly denies production derivation, numeric thresholds,
  and release/publication authority.

## Validation Checkpoint

- All 14 focused fixture/contract tests pass. They exercise every named C/Go
  case, incomplete surfaces, duplicates, unknown/malformed/failed output,
  write errors, iteration mismatch, raw/parsed disagreement, exact command
  parameters, path redaction, and direct clean/dirty Git-state capture.
- A bounded real root Make capture with 10,000 C iterations and `10x` Go
  benchtime passed all 6 C and 8 Go cases, then passed independent validation.
  The temporary artifact recorded the intentionally dirty implementation tree
  with eight status entries and contained no `/home`, `/tmp`, or `/var/tmp`
  path. Its numeric output was not committed or treated as R90-74 evidence.
- `make shell-check`, `make python-check`, `make docs-check`, the 36-test
  `make evidence-check`, the 33-test `make knowledge-check`, complete native
  `make test` with uncached Go race execution, and `git diff --check` pass.
- All 88 task-state JSON files parse; all 78 roadmap rows match exactly one
  Definition; every unfinished Definition retains goal, risk, validation, and
  stop fields.
- No benchmark fixture, timed region, runtime, storage, config, dependency,
  release-gate, threshold, external input, publication surface, or generated
  numeric evidence is changed. R90-74, R90-75, and R90-59 remain unstarted.

## Execution Deviations

- The first implementation checkpoint validated the complete surface but only
  checked that parsed case names were present in a loaded artifact. Review
  identified that an edited parsed result could then disagree with retained
  raw output. Before delivery, validation was tightened to reparse and compare
  both surfaces, recompute the environment fingerprint, and bind command argv
  to parameters; focused and complete validation were rerun successfully.
- No benchmark execution, delivery, authority, or evidence-classification
  deviation remains unresolved.

## Authority Boundaries

This trigger authorizes only R90-73 benchmark-evidence tooling, tests,
documentation, task/roadmap records, validation, commit/push, and the local
Vault workflow. It does not authorize private/external input, benchmark or
runtime semantic changes, numeric thresholds, portable/production claims,
R90-74 execution, publication, tags, releases, images, registries, or workflow
dispatch.

## Stop Conditions

Stop if the established surface cannot be identified unambiguously, path
redaction cannot preserve reviewable evidence, a complete bounded run is
partial or ambiguous, required validation is flaky, or completion requires a
benchmark/runtime semantic change, private input, threshold/product decision,
publication authority, or later increment.
