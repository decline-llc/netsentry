# Task Plan: R90-69 storage-fault delivery and forward-queue audit

## Metadata

- Timestamp: 2026-08-04T04:31:11-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `159fcf92122b387b3b80ecc5853150a6de1450d0`

## Goal

Reconcile the completed R90-66 through R90-68 local storage-fault sequence
against current code, direct tests, Git/task-state evidence, fetched
`origin/main`, the exact local Vault, and public remaining-gap claims, then
restore a bounded dependency-ready queue through the active roadmap horizon.

## Scope

- Audit all three feature and docs-only closure pairs, their completed task
  states, named direct test bodies, and the exact pushed ranges.
- Verify the single local Vault contains all six iteration notes, full-index
  entries, MOC links, and current stable SQLite/testing authority.
- Review the Jul 14 through Aug 4 delivery chain at phase level and record only
  material deviations, stale authority, missing records, and unresolved risks.
- Compare current architecture, development, performance, README, and audit
  gap claims with checked-in source, tests, and Make targets.
- Add a dated audit and a dependency-ordered local Go benchmark queue with a
  complete window, dependency, risk, acceptance, validation, and stop
  definition for every increment through Oct 31.
- Reconcile current architecture/development guidance, roadmap checkpoint, and
  task state without implementing the later queue.

## Non-Goals

- Do not change runtime, tests, benchmark harnesses, evidence generators,
  release gates, storage behavior, or configuration behavior.
- Do not rerun historical storage fault campaigns or rewrite completed plans,
  task evidence, audit records, or generated Vault iteration notes.
- Do not access private/external corpora or claim local synthetic benchmarks
  are realistic traffic, production throughput, or release evidence.
- Do not select IPv6, pcapng/DLT, TLS decryption, stream/fragment reassembly,
  schema migration, or automatic cleanup without their separate product and
  compatibility decisions.
- Do not create a tag, GitHub Release, image, registry mutation, or workflow
  dispatch, and do not start R90-59 or any newly planned increment.

## Risks

- Historical plans and audits contain intentionally frozen gap language;
  treating them as current authority could reopen completed storage work or
  rewrite valid evidence.
- Feature task states record the feature SHA while their later docs-only
  closures are separate commits; checking only one source can falsely report a
  missing or repeated delivery step.
- Broad performance-roadmap prose can imply host-independent budgets or
  production evidence that the existing local synthetic tools cannot support.
- Adding every long-term product goal to the ready queue would create an
  unbounded plan; each local item must correspond to a direct checked-in gap.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-66 through R90-68 are exactly recoverable | Full feature/closure SHAs and parents, intended-path diffs, completed states/plans, and fetched `HEAD == origin/main` |
| Every promised storage boundary has direct evidence | Named primary contention/cancellation, append lifecycle, and primary/encoded-shard clear lifecycle test bodies plus their production call sites |
| Vault evidence is complete and current | Six exact chained iteration notes, six full-index rows, six MOC links, and stable SQLite/testing/MOC review in the single discovered Vault |
| Recent delivery history has no concealed deviation | Dated three-phase commit counts plus material delivery, validation, authority, and unresolved-risk review |
| Current public gaps are correctly classified | Side-by-side architecture/development/performance/README/AUDIT claims, `make bench`, absence of Go `Benchmark*` functions, and external/product boundary review |
| The local queue is complete and bounded | R90-70 through R90-72 each name dependency, window, risk, acceptance, required validation, and stop condition through 2026-10-31 |
| No later work is started | Diff contains only R90-69 plan/state/audit/current-doc/roadmap reconciliation and no source, test, or benchmark behavior change |
| Repository and knowledge records remain valid | All task-state JSON parses; exact row/Definition coverage; docs, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `159fcf92122b387b3b80ecc5853150a6de1450d0`.
- Verified the exact feature/closure chain from R90-66 feature
  `260d53d6b5804ca37dc83b083486d429a5e9c983` through R90-68 closure
  `159fcf92122b387b3b80ecc5853150a6de1450d0`, including commit parents,
  intended paths, completed plans, and task states.
- Inspected the named direct test bodies and their production call sites. The
  primary interruption, append open/short-write/sync/close, and post-commit
  primary/encoded-shard clear open/sync/close boundaries match their plans.
- Verified all six exact Vault notes, full-index entries, MOC links, and the
  current stable SQLite/testing/MOC account in the single existing local Vault.
- Reviewed 147 commits from Jul 14 through Aug 4 across governance/release,
  contract/recovery, and fault/fuzz-hardening phases. No missing recent
  delivery record, stale authority, or unresolved validation deviation changes
  priority.
- Parsed all 83 task-state JSON files and verified all 72 roadmap rows match
  exactly one Definition.
- Audited the unfinished queue: R90-69 is the sole dependency-ready item and
  R90-59 remains blocked on exact version/SHA publication authority.
- Compared current public gaps with source and Make targets. The completed
  storage sequence should not be reopened; external fuzz/traffic remains
  input-dependent; product-scale protocol work remains outside this trigger;
  and `make bench` runs `go test -bench=.` although no Go `Benchmark*`
  function exists.

## Validation

- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Verify every new unfinished row has status, dependency, forecast window,
  risk, acceptance criteria, required validation, and stop condition.
- Run `make docs-check` and `make knowledge-check` fail-fast.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, generic-keyword, and source/test-path review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Audit Results

- The six R90-66 through R90-68 commits form the expected uninterrupted
  feature/closure chain from the R90-65 closure to the fetched baseline. Each
  commit contains only its planned feature or three-path closeout scope.
- The three named direct regression groups reach the exact promised primary
  interruption, append lifecycle, and post-commit primary/encoded-shard clear
  boundaries. No weaker neighboring test was substituted.
- All six generated Vault notes carry the expected chained ranges and are
  reachable from the full index and MOC. Current stable SQLite/testing/MOC
  prose closes the sequence while preserving historical checkpoint wording.
- The 147-commit review found no missing recent delivery pair, unresolved
  validation deviation, or stale release authority. R90-59 remains blocked
  only on explicit exact-version and exact-commit publication authority.
- External fuzz and realistic-traffic work remains input-dependent; broad
  protocol, migration, and cleanup goals remain outside this trigger. The
  direct local gap is the absence of Go benchmarks behind the existing
  `make bench` command.
- R90-70 through R90-72 now provide one bounded sequence through Oct 31:
  matcher benchmarks, SQLite alert-store benchmarks, and a documentation audit
  before any performance budget is proposed.

## Validated Evidence

- All 84 task-state JSON files parse and all 75 roadmap rows match exactly one
  Definition. Every unfinished row has a status, dependency, window, risk,
  acceptance criteria, required validation, and stop condition.
- `make docs-check` and the 33-test `make knowledge-check` pass.
- The exact six-path documentation diff contains no source, test, benchmark,
  evidence-generator, release-gate, or runtime behavior change.
- Formatting, whitespace, intended-path, credential-prefix, sensitive-path,
  generic-keyword, and source/test-path reviews pass.
- No workflow deviation or reusable skill correction was identified; the
  existing trigger-audit and two-commit delivery rules matched the work.

## Authority Boundaries

This trigger authorizes only the R90-69 documentation audit, current public
gap reconciliation, forward roadmap queue, repository validation,
commit/push, and local Vault workflow. It does not authorize benchmark or
runtime implementation, external/private inputs, production claims,
publication, tags, releases, images, registries, or workflow dispatch.

## Stop Conditions

Stop if completion requires changing runtime or test behavior, choosing a
product-scale protocol/schema/migration design, inventing host-independent
performance thresholds without evidence, rewriting historical records,
accessing private/external input, publication authority, or starting R90-70.
