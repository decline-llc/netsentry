# Task Plan: R90-72 performance evidence and portable-budget audit

## Metadata

- Timestamp: 2026-08-05T01:30:02-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `323be1f38fca456a0d17a7801e18bc50c5212075`

## Goal

Reconcile the complete local C and Go microbenchmark surface, repeat-pcap and
corpus-pressure tooling, public performance claims, recent delivery history,
and exact Git/task-state/remote/Vault evidence, then restore the smallest
evidence-supported performance baseline or budget queue through the active
roadmap horizon without inventing portable or production thresholds.

## Scope

- Audit the R90-70 and R90-71 feature plus docs-only closure pairs, completed
  plans/task states, named benchmark bodies, root/module Make execution, and
  exact pushed ranges.
- Verify the single local Vault contains all four iteration notes, full-index
  entries, MOC links, and current stable Makefile/rule/storage/testing
  authority.
- Review the Jul 14 through Aug 5 delivery chain at phase level and record only
  material trends, deviations, stale authority, missing records, and
  unresolved risks.
- Compare `docs/performance.md`, architecture/development/README/audit claims,
  C and Go benchmark harnesses, metrics, repeat-pcap pressure, corpus-pressure,
  and checked-in evidence.
- Add a dated audit that classifies each measurement surface by execution
  boundary, reproducibility, comparability, evidence class, and supported
  claim.
- Refresh the dependency-ordered roadmap with only the smallest bounded local
  evidence work that the audit supports; record external-environment or
  product/SLO dependencies explicitly instead of inventing a threshold.
- Reconcile current performance guidance, roadmap checkpoint, and task state
  without implementing the later queue.

## Non-Goals

- Do not change runtime, tests, benchmark harnesses, evidence generators,
  performance thresholds, release gates, or configuration behavior.
- Do not run or commit a new benchmark baseline, pressure campaign, external
  corpus result, or production-derived traffic evidence in this increment.
- Do not rewrite historical plans, task evidence, audit records, local
  benchmark samples, or generated Vault iteration notes.
- Do not claim that one host, synthetic repeat-pcap traffic, or unmatched local
  samples establish production capacity or a cross-host regression budget.
- Do not create a tag, GitHub Release, image, registry mutation, or workflow
  dispatch, and do not start R90-59 or any newly planned increment.

## Risks

- C benchmarks report custom per-packet or per-operation results while Go uses
  `testing.B`; combining them without a common evidence envelope can imply
  comparability that the raw harnesses do not provide.
- The historical June C and pressure samples predate the Go benchmark surface
  and were not collected as repeated matched-environment observations.
- Repeat-pcap and corpus-pressure rates include different pipeline, corpus,
  rule-set, and machine effects; treating them as microbenchmark thresholds
  would create false regression claims.
- Generated Vault notes preserve historical wording while stable notes carry
  current authority; checking only one can reopen a closed gap or miss a stale
  public claim.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-70 and R90-71 are exactly recoverable | Full feature/closure SHAs and parents, intended-path diffs, completed plans/states, and fetched `HEAD == origin/main` |
| Every benchmark claim reaches its named execution boundary | Direct C parser/formatter/socket, Go Aho-Corasick/rule/store benchmark bodies and root/module Make target review |
| Vault evidence is complete and current | Four exact chained iteration notes, four full-index rows, four MOC links, and stable Makefile/rule/storage/testing review in the single discovered Vault |
| Recent history has no concealed delivery deviation | Dated phase-level commit counts plus material validation, authority, delivery-record, and unresolved-risk review |
| Current performance claims are correctly bounded | Side-by-side benchmark, metrics, repeat-pcap, corpus-pressure, evidence, and public-doc classification |
| The forward queue is complete and supportable | Every new unfinished row names status, dependency, window, risk, acceptance criteria, required validation, and stop condition; unsupported portable thresholds remain explicitly blocked |
| No later work is started | Diff contains only the R90-72 plan/state/audit/current-doc/roadmap reconciliation and no source, test, benchmark, evidence, threshold, or runtime behavior change |
| Repository and knowledge records remain valid | All task-state JSON parses; exact row/Definition coverage; docs, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the complete
  rolling roadmap before selection.
- Fetched `origin/main` and verified clean local/remote equality at
  `323be1f38fca456a0d17a7801e18bc50c5212075`.
- Verified the R90-71 feature and docs-only closure are represented by exact
  Vault notes, full-index rows, MOC links, and current stable authority in the
  single discovered Vault.
- Reviewed the recent dependency-ordered delivery history and the complete
  unfinished queue. R90-72 is the sole dependency-ready item; R90-59 remains
  blocked on exact version/SHA publication authority.
- Read the prior performance-queue audit, current performance guide, root and C
  Make targets, and direct Go benchmark inventory before persisting this plan.

## Validation

- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Verify every unfinished row has status, dependency, forecast window, risk,
  acceptance criteria, required validation, and stop condition.
- Run `make docs-check` and `make knowledge-check` fail-fast.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, generic-keyword, generated-evidence, threshold, and
  source/test-path review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Audit Results

- The R90-70/R90-71 feature and closure commits form the expected uninterrupted
  chain from `fffea8c7d030b84f836137fb22e94ae552a8e677` through the fetched
  baseline. Their intended paths match their completed plans and states.
- All four generated Vault notes carry the expected chained ranges and are
  reachable from the full index and MOC. Current stable Aho-Corasick, rule,
  SQLite, Makefile, testing, and MOC prose records both delivered boundaries.
- Direct source review confirms the C parser, C formatter, local-drain UDS,
  Go Aho-Corasick, full-rule, durable-write, and indexed-query harnesses reach
  their documented boundaries while excluding the planned setup and
  correctness work from timing.
- Repeat-pcap pressure covers the deterministic synthetic full pipeline;
  corpus pressure is specific to supplied sanitized input; runtime lifetime
  rates and histograms are observation signals. None is interchangeable with
  a standalone microbenchmark or portable capacity result.
- The 153-commit phase review found no missing delivery pair or unresolved
  benchmark validation deviation. The prior R90-70 Vault default-path failure
  was recovered with explicit single-Vault selection and left no missing note.
- Current Go numeric results are not versioned; the complete surface has no
  repeated matched-environment sample set; and the three historical synthetic
  pressure samples range from 552 to 1,402 pps. The prior 10% pps/20% RSS
  figures remain planning assumptions rather than active gates.
- R90-73 through R90-75 now separate deterministic evidence capture, a
  repeated same-host observation baseline, and a budget decision blocked on
  comparable-environment evidence plus explicit product/SLO scope. No later
  increment was started.

## Validation Checkpoint

- All 87 task-state JSON files parse, all 78 roadmap rows match exactly one
  Definition, and every unfinished Definition contains goal, risk, required
  validation, and stop fields.
- `make docs-check`, the 33-test `make knowledge-check`, and
  `git diff --check` pass on the complete documentation surface.
- The exact eight-path staged diff contains only the audit, current public
  guidance, roadmap, plan, and task state. It contains no source, test,
  benchmark, evidence-generator, config, workflow, or generated-evidence path.
- Credential-prefix and operator-absolute-path scans have no matches. The sole
  generic sensitive-word match is the plan's own `credential` review label;
  manual review confirms it contains no sensitive value.
- No workflow deviation or reusable skill correction was identified; the
  existing trigger-audit, evidence-classification, and two-commit closeout
  rules matched the work before delivery. During delivery, the first
  post-fetch command incorrectly inferred a full SHA from abbreviated commit
  output and stopped before the knowledge gate. The complete sequence was
  rerun successfully with `git rev-parse HEAD`; the generic skill now requires
  resolving rather than reconstructing full SHAs.

## Delivery Results

- Feature commit:
  `13b259f3779840a8a410803dfd209f19bbb71649` (`docs: audit performance
  evidence`).
- The exact eight-path documentation feature was pushed without force,
  fetched, and verified as both `HEAD` and `origin/main`; the complete
  post-fetch 33-test knowledge gate passed on the authoritative full SHA.
- Exact range
  `323be1f38fca456a0d17a7801e18bc50c5212075..13b259f3779840a8a410803dfd209f19bbb71649`
  was synchronized twice with identical Vault tree hashes. The generated
  iteration note, full index, and MOC link are verified.
- Stable MOC, Makefile/build, and testing/release knowledge now records the
  completed R90-72 evidence boundary and R90-73/R90-74/R90-75 queue while
  preserving historical iteration notes.
- R90-73 is ready but was not started. R90-75 remains blocked on comparable
  environment evidence plus product/SLO scope, and R90-59 publication remains
  blocked.

## Authority Boundaries

This trigger authorizes only the R90-72 documentation audit, current public
performance-claim reconciliation, forward roadmap queue, repository
validation, commit/push, and local Vault workflow. It does not authorize
benchmark or runtime implementation, benchmark threshold enforcement,
external/private inputs, production claims, publication, tags, releases,
images, registries, or workflow dispatch.

## Stop Conditions

Stop if completion requires changing runtime, tests, benchmark or evidence
behavior, selecting a product/SLO threshold, accessing private/external input,
inventing cross-host comparability, rewriting historical records, publication
authority, or starting a later increment.
