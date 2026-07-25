# Task Plan: R90-38 validate recovery event identity

## Metadata

- Timestamp: 2026-07-25T10:09:40-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `54b7a39618c616566bb69792d9d11e992fd3a2ef`

## Goal

Reject recovery records whose nonblank `event_id` differs from the
deterministic event identity used by normalized alert writes and replay
idempotency.

## Scope

- Compare every decoded recovery `event_id` with the deterministic identity
  derived from the record's event fields.
- Reject a mismatched identity before startup replay or runtime append can
  modify durable state.
- Preserve the complete recovery log and missing or existing SQLite state on
  startup rejection.
- Preserve recovery-log and SQLite bytes when runtime preflight rejects an
  existing mismatched record.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change event-ID derivation or the recovery JSONL format.
- Do not validate or reconcile aggregated SQLite row `event_id` values.
- Do not scan or repair the `alert_events` ledger.
- Do not repair, truncate, delete, or rewrite invalid recovery input.
- Do not create a release tag or publish artifacts.

## Risks

- An altered event identity can bypass replay deduplication or collide with an
  unrelated event.
- Validation must precede writable SQLite initialization and runtime append.
- The comparison must use the writer's existing deterministic derivation
  rather than a duplicate implementation.

## Validation

- Direct nonblank event-identity mismatch after a valid recovery prefix.
- Missing and existing database startup preservation.
- Runtime preflight rejection with full log and database byte preservation.
- Valid startup replay and idempotency compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Every recovery `event_id` equals the deterministic identity derived from its
  event fields.
- A mismatch fails with `ErrRecoveryLogIntegrity` before SQLite modification.
- Startup rejection preserves the complete log and missing/existing database
  state.
- Runtime rejection appends nothing and preserves recovery-log and database
  bytes.
- Valid replay and duplicate-event idempotency retain current behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires changing event-ID derivation, rewriting the
recovery format, stored-row or event-ledger reconciliation, automatic repair,
operator data, tag creation, or publication authority.

## Delivery Evidence

- Feature commit: `99b081bf941af5ab4a257900c3d08cfd339c5dc2`
- Implementation and twenty uncached focused alert-store race runs pass.
- Complete native race suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks pass.
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `54b7a39618c616566bb69792d9d11e992fd3a2ef..99b081bf941af5ab4a257900c3d08cfd339c5dc2`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
