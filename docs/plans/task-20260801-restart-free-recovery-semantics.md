# Task Plan: R90-57 restart-free emergency recovery semantics

## Metadata

- Timestamp: 2026-08-01T04:57:43-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `46bbf8a0535c30e707b7dfbaefee9cab27a81d84`

## Goal

Define a fail-closed, operator-triggered state machine for recovering alert
storage from sticky emergency mode without restarting NetSentry, before any
runtime or API implementation begins.

## Selected Product Policy

The Aug 1 user instruction cancels hard schedule gates and additional
prerequisite review. R90-57 therefore selects the safest bounded default:

- recovery is initiated only by an authenticated operator action;
- NetSentry never retries recovery in the background;
- recovery never deletes, repairs, rotates, or truncates operator evidence as
  cleanup;
- one serialized probe owns recovery while normal writes remain fail-closed;
- only a complete probe and idempotent replay may return storage to healthy.

## Scope

- Document the states, events, guards, actions, and externally observable
  outcomes for operator-triggered recovery.
- Define write ownership, probe serialization, cancellation, retry, concurrent
  operator request, shutdown, and recovery-log preservation invariants.
- Define the implementation test plan, including fault, replay, concurrency,
  cancellation, and evidence-preservation cases.
- Reconcile architecture, API, development guidance, changelog gap wording,
  roadmap status, and task state.

## Non-Goals

- Do not add a recovery API, CLI, configuration field, goroutine, timer, or
  background retry loop.
- Do not change `Store`, SQLite, recovery-log, health, metrics, or runtime
  behavior.
- Do not automatically delete, repair, replace, truncate, or quarantine a
  database, WAL, SHM, or recovery log.
- Do not tag or publish v0.1.1, dispatch workflows, or access operator data.
- Do not begin R90-59 or a later implementation increment.

## Risks

- A probe that races `WriteBatch` can duplicate or reorder recovery records.
- Replaying through a second writable owner can conflict with the live SQLite
  handle or daily-shard handles.
- Treating a read-only health check as recovery proof can clear emergency mode
  while writes still fail.
- Cancellation or shutdown during replay can leave a committed prefix and an
  untruncated log; the design must retain idempotent replay safety.
- Automatic cleanup can destroy the only database or sidecar evidence.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Complete state machine | Architecture table defines states, transitions, guards, serialized actions, and observable results |
| Fail-closed write ownership | Concurrency invariants keep normal writes blocked and permit one probe owner |
| Evidence preservation | Design forbids cleanup and defines byte-preservation checks for rejected probes |
| Safe retry/cancellation | Development test plan covers repeated triggers, cancellation, shutdown, partial commit, and idempotent replay |
| Operator boundary | API guidance defines an authenticated explicit action and stable conflict/failure outcomes without implementing it |
| Accurate scope | Changelog and roadmap distinguish the completed design from unimplemented runtime recovery |
| Repository delivery | Documentation, knowledge, JSON, diff, remote, and exact-range Vault evidence pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `46bbf8a0535c30e707b7dfbaefee9cab27a81d84`.
- Verified the R90-58 feature and closure chain in Git and the local Vault note,
  full index, and MOC.
- Reconciled 251 July/August commits and 127 commits since Jul 14 at phase
  level; the expected implementation/closure pattern has no missing delivery
  record.
- Verified all 62 roadmap rows have Definitions and both unfinished items have
  status, dependency, forecast window, risk, acceptance criteria, validation,
  and stop conditions.
- Confirmed the Jul 16 global schedule waiver already makes dates forecasts;
  the Aug 1 instruction removes R90-57's additional product-review gate.
- Selected R90-57 as the highest-priority dependency-ready increment. R90-59
  still requires explicit authority for actual tag/publication mutations.

## Validation

- Review the state machine against current `Store.WriteBatch`, emergency health,
  recovery append/replay/truncate ordering, and existing emergency tests.
- Verify the test plan directly covers every promised state and failure edge.
- `make docs-check`
- `make evidence-check`
- `make knowledge-check`
- Parse every task-state JSON file and verify roadmap row/Definition coverage.
- `git diff --check`, staged-diff review, and sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, synchronize each exact full-SHA range, and verify its Vault
  note, full index, MOC link, and stable storage knowledge.

## Authority Boundaries

The user authorized removing schedule and additional product-review gates for
roadmap selection. The design may choose the safest bounded default. The user
did not authorize runtime implementation, automatic cleanup, private-data use,
tag creation, GitHub Release or GHCR publication, or workflow dispatch.

## Stop Conditions

Stop on ambiguous current behavior, a design that needs two writable owners or
automatic cleanup, any requirement to change runtime code/API behavior, failed
or ambiguous required validation, private data, publication action, or work
beyond this single increment.

## Validation Evidence

- Reviewed the design against current `Store.WriteBatch` ordering: existing-log
  preflight, current-batch validation, durable append, emergency short-circuit,
  SQLite write, and truncate-after-success.
- Reviewed emergency classification and sticky health behavior plus the direct
  restart-replay regression in `engine/internal/alert/store_test.go`.
- The state table covers healthy, degraded, emergency, recovering, and closed;
  trigger, duplicate trigger, ordinary operations, preflight rejection,
  writable failure, cancellation, shutdown, and success all have explicit
  transitions.
- The implementation plan covers one-owner concurrency, lifecycle exclusion,
  pre-writable byte preservation, post-writable disclosure, idempotent partial
  replay, empty-log write proof, daily shards, encoded paths, authentication,
  audit redaction, cancellation, and shutdown.
- `make docs-check`: pass.
- `make evidence-check`: pass, 16 tests.
- `make knowledge-check`: pass, 33 tests.
- Every task-state JSON file parsed and all 62 roadmap rows matched exactly one
  Definition.
- `git diff --check`: pass.

No runtime code, API route, configuration, schema, tag, publication action, or
private evidence was introduced. Repository delivery and exact-range Vault
evidence remain pending.
