# Task Plan: R90-77 rule-management transaction serialization

## Metadata

- Timestamp: 2026-08-09T05:04:52-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `40798847be8e7bb9270b5c5d7675c27f7addf7b1`

## Goal

Serialize each file-backed rule create, update, delete, and explicit reload as
one management transaction so concurrent successful requests cannot lose an
accepted mutation or leave the canonical seed file and active immutable rule
snapshot disagreeing, while packet matching remains lock-free.

## Scope

- Add one server-owned rule-management transaction lock shared by create,
  update, delete, and explicit reload.
- Hold the lock across each operation's authoritative state read, validation,
  canonical file replacement or reload read, and active snapshot publication.
- Add synchronized concurrent create/create, update/delete, and
  mutation/reload regressions that prove the second transaction waits at an
  observable boundary and that both successful outcomes survive.
- Add direct validation- and persistence-failure regressions proving that a
  rejected transaction neither changes the canonical file nor the active
  snapshot and does not prevent a later valid transaction.
- Update current architecture, development, API, roadmap, task-state, and
  changelog guidance to record the serialized management boundary.

## Non-Goals

- Do not change rule schemas, validation, ordering, response bodies/statuses,
  authentication, matching semantics, or the immutable snapshot design.
- Do not add cross-process file locking, seed-file migration, automatic retry,
  optimistic versioning, or a new dependency.
- Do not harden short write, file sync, close, rename, or parent-directory sync
  behavior reserved for R90-78.
- Do not change suppression management, start R90-78, activate a performance
  budget, or exercise tag/release/registry/workflow publication authority.

## Risks

- Locking only file replacement or snapshot publication would retain the
  read-modify-write lost-update window.
- Reusing the rule engine's snapshot mechanism as the transaction lock would
  block or couple packet matching to management requests.
- A test based on request timing or goroutine scheduling could pass without
  reaching the promised interleaving.
- A failed transaction could leak the lock, modify the file before rejection,
  or leave a later valid operation unable to proceed.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Concurrent creates cannot lose either successful mutation | A synchronized create/create regression blocks the first transaction at its authoritative state read, starts the second, proves it cannot cross the same boundary, then verifies both responses, canonical file, and active snapshot |
| Concurrent update/delete preserve both successful outcomes | A synchronized update/delete regression reaches the same boundary and verifies the updated survivor plus deleted target in both file and memory |
| Mutation and explicit reload share one transaction order | A synchronized mutation/reload regression proves reload cannot pass a blocked mutation and verifies both successful responses plus exact file/snapshot agreement |
| Rejected transactions preserve state and release serialization | Direct invalid-rule and unwritable-seed regressions compare canonical file and active snapshot before/after rejection, then complete a later valid transaction |
| Packet matching remains outside the management lock | The lock is owned only by the API server transaction path; the rule engine keeps its atomic immutable snapshot and focused engine race tests continue to pass |
| Public behavior and future boundaries are accurate | Current architecture, development, API, changelog, roadmap, and task state describe serialized in-process management without claiming cross-process or crash durability |
| Repository evidence is complete | Focused repeated API race tests, complete API/rule tests, full native tests, E2E smoke, docs, knowledge, JSON/roadmap, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Read `netsentry-next`, `netsentry-roadmap`, the complete rolling roadmap,
  repository knowledge contract, latest R90-76 plan/state, and relevant source,
  tests, public docs, and stable Vault notes before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 40798847be8e7bb9270b5c5d7675c27f7addf7b1`.
- Verified the exact R90-76 feature and closure commits, both Vault iteration
  notes, full-index rows, MOC links, and stable R90-77 rule/config/API authority.
- Reviewed the Jul 20 through Aug 9 delivery sequence at phase level. The only
  commits since the Aug 9 audit are its expected feature and closure; no new
  delivery deviation, stale authority, or missing record changes priority.
- Audited the unfinished queue: R90-77 is the only dependency-ready local
  increment; R90-59 and R90-75 remain externally blocked, and R90-78 through
  R90-80 retain unfinished internal dependencies plus complete definitions.
- Direct source review confirms create/update/delete derive full candidates
  from `RuleManager.Rules`, while explicit reload independently reads the seed
  file; none shares a transaction lock. Existing tests cover individual
  operations but no synchronized management interleaving.

## Validation

- Preflight the repository-selected Go toolchain plus GCC, Make, Bash, and
  Python before the complete validation chain.
- Run `gofmt` and focused uncached API/rule tests, then repeat the synchronized
  concurrency tests uncached under the race detector.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast on the final surface.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage
  plus complete fields for every unfinished increment.
- Run `git diff --check`, intended staged-diff review, and scoped credential,
  sensitive-path, dependency, schema, config, workflow, release, and
  publication reviews.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Added one `Server`-owned rule transaction mutex shared by create, update,
  delete, and explicit reload. Mutation locks begin before the authoritative
  active snapshot read; reload locks before the canonical file read; both hold
  through validation/file handling and active snapshot publication.
- The rule engine and packet-matching interfaces are unchanged, so `Match`
  continues to read immutable state through `atomic.Pointer` without acquiring
  the API management lock.
- Added channel-synchronized create/create, update/delete, and mutation/reload
  regressions. Each blocks the first transaction at its authoritative read,
  observes that the second cannot cross the corresponding rule-manager
  boundary, then verifies both HTTP outcomes plus exact canonical-file and
  active-snapshot agreement.
- Added direct validation and read-only-directory persistence failures that
  compare prior file bytes and active rules, then prove the same server accepts
  a later valid request after the deferred unlock.
- The first focused command found a missing test-only package import and did
  not produce behavioral evidence. After correction, the complete uncached API
  and rule package sequence passed; the four direct concurrency/failure tests
  then passed 20 uncached race-detector repetitions.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl, timeout, jq, pkg-config, and libpcap
  1.10.4 were available before the complete fail-fast chain.
- Final uncached focused API/rule tests and their race-detector variants pass.
  The four direct R90-77 regressions pass 20 uncached race repetitions, with
  each first transaction synchronized at an observable `RuleManager` boundary
  rather than a scheduling assumption.
- `make test` passes the C parser/sender tests and every Go package uncached
  under the race detector. `make e2e-smoke` passes with 6 packets processed, 5
  alerts generated, and 8 rules loaded.
- `make docs-check` and the 33-test `make knowledge-check` pass on the complete
  behavior/documentation surface.
- All task-state JSON files parse; every roadmap row has exactly one Definition
  and all unfinished items retain status, dependency, window, risk, acceptance,
  validation, and stop records.
- `gofmt`, `git diff --check`, exact nine-path scope, staged-diff, credential,
  sensitive-path, dependency, schema, config, workflow, release, and
  publication reviews pass. No dependency, rule schema, configuration,
  workflow, release artifact, or external mutation was added.

## Authority Boundaries

This trigger authorizes only R90-77 in-process rule-management serialization,
its direct tests, current documentation/task-state reconciliation, repository
validation, commit/push, and local Vault synchronization. It does not authorize
cross-process coordination, schema or compatibility policy, file-durability
work, suppression changes, private/external data, performance policy, tag or
release mutation, image/registry publication, or workflow dispatch.

## Stop Conditions

Stop if correct serialization requires public rule semantic/schema changes,
cross-process locking, migration policy, disabling hot reload, a new
dependency, private data, publication authority, or an ambiguous synchronized
regression or full-suite result that remains after focused uncached review.
