# Task Plan: R90-18 reject inconsistent normalized recovery records

## Metadata

- Timestamp: 2026-07-21T03:57:46-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `c1bed1e25d90e88b7d29aad0d294e4e2d137bcf8`

## Goal

Reject structurally valid recovery records whose durable identity, timestamps,
window, or aggregate count cannot be emitted by the normalized alert writer.

## Scope

- Derive recovery invariants from the same normalized-alert rules used before
  appending durable JSONL records.
- Reject mismatched durable ID, first-seen, last-seen, window-start, and
  aggregate-count values before SQLite initialization or replay.
- Preserve the complete recovery log, a missing database, and compatible
  existing database bytes for every rejection class.
- Retain valid recovery replay, truncation, and event-ID idempotency.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the recovery JSONL format or add automatic repair.
- Do not derive or replace caller-supplied event IDs.
- Do not add cross-process recovery locking or schema migration.
- Do not create a release tag or publish artifacts.

## Risks

- Duplicating writer logic in validation can drift and reject valid records.
- Comparing timestamps by representation rather than instant can reject
  semantically equivalent valid values.
- Replay normalization can conceal invalid durable fields unless rejection
  occurs during preflight.

## Validation

- Direct durable-ID, first-seen, last-seen, window-start, and aggregate-count
  rejection cases with full-log and missing-database preservation.
- Each rejection against a compatible database with byte preservation and no
  optional-index creation.
- Valid replay, truncation, and idempotency compatibility.
- Repeated focused alert-store race tests, full native tests, E2E smoke,
  documentation, knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Every inconsistent normalized field fails with `ErrRecoveryLogIntegrity`
  before writable SQLite initialization.
- Rejected input and target database state remain unchanged.
- Validation and durable writing share one normalized invariant definition.
- Valid replay behavior remains compatible.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if the writer contract is ambiguous, compatibility requires accepting
records the current writer cannot emit, automatic log repair is required,
operator data is needed, or work reaches tag/publication authority.

## Local Validation

- Durable-ID, first-seen, last-seen, window-start, and aggregate-count
  mismatches each fail before SQLite initialization against both missing and
  compatible existing databases.
- Every direct rejection preserves the complete recovery log; missing
  databases remain absent, and existing database bytes plus optional-index
  state remain unchanged.
- Valid replay, truncation, and event-ID idempotency regressions pass.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, knowledge, JSON, and diff checks pass.
- Feature commit: `cb2fd7d1889b33a01829226becb44260f1668651`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `c1bed1e25d90e88b7d29aad0d294e4e2d137bcf8..cb2fd7d1889b33a01829226becb44260f1668651`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
