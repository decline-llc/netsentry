# Task Plan: R90-82 receiver idle-capacity evidence

## Metadata

- Timestamp: 2026-08-09T07:32:58-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `9541d44db18b9c13e521b83be8aae79a9e5068be`

## Goal

Make the receiver idle-timeout capacity-release regression synchronize on the
actual handler-slot boundary and prove replacement acceptance without polling
the process-wide latest-session snapshot, while preserving production timeout,
protocol, configuration, and shutdown behavior.

## Scope

- Retain the bounded connection limiter as an internal receiver implementation
  detail while making its available-slot token directly observable to package
  tests.
- Update the idle-timeout regression to prove the first handler processed a
  frame, exited because the idle deadline closed its connection, released its
  exact slot, and admitted a replacement that delivers a packet.
- Apply the same direct slot-release and replacement-delivery boundary to the
  existing ordinary-disconnect and protocol-violation capacity-reuse tests.
- Record focused repeated race, complete native, E2E, documentation, knowledge,
  and delivery evidence for this one increment.

## Non-Goals

- Do not add a public runtime API, metric, configuration option, protocol frame,
  production hook, fixed sleep, or broad retry loop.
- Do not change the configured connection limit, read timeout, deadline refresh,
  overload, decode-error, reconnect, or shutdown semantics.
- Do not diagnose a production timeout defect from historical validation
  deviations or change unrelated receiver tests and runtime behavior.
- Do not access private/external traffic or perform tag, release, registry,
  workflow, or other publication mutation.

## Risks

- Refactoring the limiter from occupied tokens to available tokens could alter
  overload or slot-release behavior if acquisition and return are not exactly
  paired.
- A test that observes connection close without the limiter token can still race
  the handler defer and overstate capacity release.
- A replacement hello observed through shared state can be overwritten and does
  not directly prove the replacement handler accepted a data frame.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| The handler-capacity boundary is observable without a public seam | Internal available-token limiter initialized before the accept loop; package test claims and returns the exact token after each terminating condition |
| Idle timeout exits the first handler and releases its slot | First connection delivers a packet, receiver-side deadline closes it without a decode error, then the test directly claims the released limiter token |
| A replacement is accepted without shared-session polling | Replacement sends hello plus a distinct packet and `WaitForPacket` observes that packet after the released token is returned |
| Existing capacity reuse remains direct | Ordinary disconnect and protocol violation regressions claim the released token and prove distinct replacement packet delivery; overload and decode-error assertions remain intact |
| Production semantics remain unchanged | No exported API/config/protocol/default change; limiter preserves nonblocking excess rejection and one return for every acquired slot |
| The evidence is repeatable and repository-complete | Focused normal run, at least twenty uncached focused race executions, complete `make test`, E2E, docs, knowledge, JSON/Definition, diff, staged-scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 9541d44db18b9c13e521b83be8aae79a9e5068be`.
- Verified the R90-81 docs-only closure iteration note, full-index row, MOC
  link, and current stable receiver test authority in the sole local Vault.
- Reviewed the Jul 20 through Aug 9 delivery phases and found no newer missing
  record, stale stable authority, or unresolved validation result that changes
  priority.
- Audited the forward queue: R90-59 and R90-75 retain exact external blockers;
  R90-82 is the sole dependency-ready increment and every unfinished record has
  status, dependency, window, risk, acceptance, validation, and stop fields.

## Validation

- Run the three direct capacity-reuse regressions normally and at least twenty
  times uncached under the race detector from the owning `engine` module.
- Run the complete uncached native race suite with `make test`, then
  `make e2e-smoke`, `make docs-check`, and `make knowledge-check` fail-fast.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage
  plus complete unfinished-item fields.
- Run `gofmt`, `git diff --check`, staged-diff review, and scoped credential,
  sensitive-path, configuration/protocol/runtime/public-API, release, and
  publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range.

## Implementation Checkpoint

- The receiver now initializes its existing bounded limiter as an internal
  available-token channel before starting the accept loop. Each accepted
  handler claims exactly one token and returns exactly one token after
  `handleConn` exits; each accept loop retains its own initialized channel while
  package tests observe the same channel through the receiver, and the
  nonblocking overload branch is unchanged.
- The protocol-violation, ordinary-disconnect, and idle-timeout capacity tests
  each claim and return the actual released token before dialing a replacement.
  The ordinary and idle first handlers prove readiness by delivering a packet,
  and every replacement proves acceptance by delivering a distinct packet.
- The idle test still requires receiver-side close before the client deadline
  and still asserts zero decode errors. The protocol test still asserts exactly
  one decode error, and the ordinary test still directly exercises excess
  rejection.
- The prior shared latest-session polling and retry loops were removed from all
  three direct capacity-reuse regressions. No exported API, configuration,
  protocol, default, timeout, or shutdown contract changed.

## Validation Checkpoint

- The first focused command combined repository-relative source paths with the
  `engine` module working directory and stopped at `gofmt` before any formatting
  or test ran. The corrected complete focused sequence used module-relative
  paths and passed.
- The three direct regressions pass once normally and twenty times uncached
  under the race detector. Complete repository validation remains the delivery
  boundary.
- The first exact roadmap structural gate found 86 queue rows but 87 Definition
  headings because R90-81 added a complete R90-04a Definition without removing
  the older policy-section definition. This directly contradicted the prior
  closure's exact-one claim. Delivery remained blocked while Git history proved
  the duplicate's origin; the newer complete definition remains active and the
  redundant older heading/prose was removed. The completed R90-04a row and
  immutable R90-81 records were not rewritten. All structural and repository
  validation must be rerun after this documentation correction.
- The corrected structural gate proves all 86 roadmap rows match exactly one of
  86 Definitions, R90-59/R90-75/R90-82 retain every required unfinished-item
  field, all 97 task-state JSON files parse, and `git diff --check` passes.
- The complete fail-fast repository chain then passed again: both C test
  binaries; every Go package uncached under the race detector; E2E smoke with
  six packets processed, five alerts generated, and eight rules loaded;
  `make docs-check`; and all 33 `make knowledge-check` tests.
- Every planned acceptance criterion is satisfied. The only deviations were the
  stopped pre-test module-path command and the evidence-grounded duplicate
  Definition reconciliation; neither changed the R90-82 behavior boundary or
  required new authority.
- Staged review tightened the limiter ownership so each accept loop receives
  its own initialized channel rather than dynamically reading the receiver
  field. The same channel remains observable to package tests. The focused
  normal run, twenty uncached race runs, and complete repository chain all
  passed again after that final source change.
- Final staged review contains exactly the intended roadmap, plan, task-state,
  receiver source, and receiver test paths. Credential-prefix,
  operator-sensitive-path, private-traffic, configuration, dependency,
  workflow, generated-evidence, release-artifact, and publication-mutation
  review found no out-of-scope addition.

## Authority Boundaries

This trigger authorizes only R90-82's internal limiter observability, direct
receiver capacity-reuse regressions, required documentation/task-state
reconciliation, validation, commit/push of the completed increment, and local
Vault synchronization. It does not authorize public API/configuration/protocol
changes, production-traffic claims, private input, performance policy, tag or
release publication, image/registry mutation, or workflow dispatch.

## Delivery Results

- Feature commit:
  `6118a0fb628a2a0ae0527c0783f436f96314a353` (`test: stabilize receiver
  capacity release evidence`). It contains exactly the five validated roadmap,
  plan, task-state, receiver source, and receiver test paths.
- `main` was pushed without force or tags. A fresh fetch verified
  `HEAD == origin/main == 6118a0fb628a2a0ae0527c0783f436f96314a353`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `9541d44db18b9c13e521b83be8aae79a9e5068be..6118a0fb628a2a0ae0527c0783f436f96314a353`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC, test-gate, and UDS protocol prose was reconciled to the delivered
  available-token capacity boundary and the corrected R90-04a duplicate
  Definition authority without rewriting immutable iteration notes. Replaying
  the identical range preserved exact Vault content hash
  `bdacddb75b810a02a1d87373646989491f0c47e4327b351ca9908d4c3e442a00`.
- R90-82 is complete. R90-59 and R90-75 retain their exact external blockers;
  no dependency-ready local increment remains and no later work was started.

## Stop Conditions

Stop if deterministic evidence requires a public runtime seam, changes timeout
or overload semantics, cannot directly distinguish handler exit from slot
release, leaves repeated/full validation ambiguous, needs private/external
services, or crosses publication authority.
