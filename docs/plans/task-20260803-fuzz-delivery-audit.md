# Task Plan: R90-65 fuzz delivery and local hardening audit

## Metadata

- Timestamp: 2026-08-03T04:41:51-07:00
- Branch: main
- Risk Level: Low
- Remote baseline: `23983e1ac696b923a4595e7b97f0e7e1d935dc97`

## Goal

Reconcile the completed dual-harness fuzz delivery with current public gap
claims, code and direct tests, fetched Git evidence, and the exact local Vault,
then replace the broad local storage-fault gap with a bounded dependency-ready
queue through the active roadmap horizon.

## Scope

- Audit the R90-64 feature and closure commits, completed task state, evidence
  record, fetched `origin/main`, exact Vault notes/index/MOC links, and stable
  fuzz/testing knowledge.
- Review the Jul 14 through Aug 3 delivery chain at phase level and record only
  material delivery, authority, or evidence deviations.
- Compare every current public remaining-gap claim with checked-in code and
  direct tests, distinguishing local work from external-input work.
- Add a dated audit and a dependency-ordered local storage-fault queue with a
  complete window, dependency, risk, acceptance, validation, and stop
  definition for every increment.
- Reconcile current architecture/development wording, the roadmap checkpoint,
  and this task state without implementing the later queue.

## Non-Goals

- Do not change runtime, test, fuzz, evidence-generation, release-gate, or
  storage behavior.
- Do not run another sustained campaign, access a private/external corpus, or
  rewrite immutable R90-04 or R90-64 evidence.
- Do not claim the local synthetic fuzz baseline is external, realistic,
  production-derived, or release evidence.
- Do not create a tag, GitHub Release, image, registry mutation, or workflow
  dispatch, and do not start R90-59 or any newly planned increment.
- Do not reopen product-scale protocol, schema, migration, or automatic cleanup
  decisions from historical audit documents.

## Risks

- Historical plans and audit reports contain intentionally frozen gap language;
  treating it as current authority would rewrite evidence or duplicate work.
- The public external fuzz and realistic-traffic gaps require approved input;
  placing them in the ready local queue would create a false unblock condition.
- Broad SQLite fault language can hide already-covered corruption, sidecar,
  recovery, and emergency boundaries; future work must name a direct untested
  failure boundary and its preservation evidence.
- The exact R90-64 Vault notes and links are present, but current stable Vault
  prose still contains superseded release-blocker language; closeout must
  reconcile that local knowledge without changing historical iteration notes.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| R90-64 delivery is exact and recoverable | Full feature/closure SHAs, fetched ref equality, completed state, committed evidence, exact Vault notes, full-index rows, MOC links, and stable-note review |
| Recent delivery history has no concealed deviation | Dated phase counts plus material trend, delivery-record, authority, and unresolved-risk review |
| Every public current gap is correctly classified | Side-by-side architecture/development claims, R90-04/R90-64 evidence, current code, and direct test names/bodies |
| External work is not presented as local-ready | Audit records external fuzz corpus and broader realistic traffic as input-dependent optional evidence, not R90-59 or release-gate prerequisites |
| The local queue is complete and bounded | Every added row and Definition names dependency, window, risk, acceptance criteria, required validation, and stop condition through 2026-10-31 |
| No later work is started | Diff contains only R90-65 plan/state/audit/current-doc/roadmap reconciliation and no source or test behavior change |
| Repository and knowledge records remain valid | Task-state JSON parsing, exact row/Definition coverage, docs, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `23983e1ac696b923a4595e7b97f0e7e1d935dc97`.
- Verified the exact R90-64 feature
  `73ab39ef88245b01b3d3418f0d9aeb0f6db1d546` and closure
  `23983e1ac696b923a4595e7b97f0e7e1d935dc97`, completed task state, and
  committed one-million-iteration dual-harness evidence.
- Verified both exact Vault iteration notes, their full-index entries, MOC
  links, and the R90-64 stable fuzz/testing updates in the single existing
  local Vault discovered under the repository knowledge contract.
- Counted 139 commits from Jul 14 through Aug 3 across three phases (47, 56,
  and 36); no missing recent feature/closure record or remote deviation changes
  priority.
- Parsed all 79 task-state JSON files and verified all 68 roadmap rows match
  exactly one Definition.
- Audited both unfinished rows: R90-59 remains blocked on exact publication
  authority; dependency-complete R90-65 is the only ready increment.

## Planned Local Queue

1. Prove ordinary primary writes remain replay-safe when SQLite contention or
   active cancellation occurs after the recovery append but before commit.
2. Add deterministic recovery-log append fault coverage for open, write, sync,
   and close boundaries without permitting database mutation or loss of the
   pre-existing valid log prefix.
3. Prove post-commit recovery-log clearing faults cannot lose alerts or inflate
   aggregates and remain recoverable through an explicit bounded retry.

The roadmap will assign these increments dependencies and forecast windows
through Oct 31. External corpus campaigns and broader realistic-traffic tuning
remain separately input-dependent and are not inserted as ready work.

## Validation

- Run `make docs-check` and `make knowledge-check`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Verify every unfinished increment contains status, dependency, window, risk,
  acceptance criteria, required validation, and stop condition.
- Run `git diff --check`, intended staged-diff review, and a scoped
  sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Validated Evidence

- Fetched `origin/main` equals clean starting baseline
  `23983e1ac696b923a4595e7b97f0e7e1d935dc97`.
- Exact R90-64 feature/closure commits, completed state, committed evidence,
  two Vault notes, full-index entries, MOC links, and stable fuzz/testing notes
  were reviewed directly.
- The dated audit reconciles 139 commits across three phases, three current
  public gap claims, the R90-04 public real-traffic result, and the R90-64 local
  synthetic result without treating commit volume as completion evidence.
- All 80 task-state JSON files parse and all 71 roadmap rows match exactly one
  Definition; the five unfinished rows each have a dependency, window, risk,
  acceptance criteria, required validation, and stop condition.
- `make docs-check`, the 33-test `make knowledge-check`, and the initial
  `git diff --check` passed.

## Authority Boundaries

This trigger authorizes the bounded R90-65 documentation audit, roadmap
refresh, repository validation, commit/push, and local Vault reconciliation.
It does not authorize private or external input, runtime/test implementation,
historical evidence rewrite, release acceptance changes, tags, releases,
images, registry publication, or workflow dispatch.

## Stop Conditions

Stop without later implementation if exact R90-64 delivery evidence is
ambiguous, current public claims cannot be reconciled without rewriting
history, the next queue needs private/external input or a product/release
decision, validation is ambiguous, or work would start a later increment.
