# Task Plan: R90-17 preflight recovery logs before writable SQLite initialization

## Metadata

- Timestamp: 2026-07-20T05:34:39-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `4e13a25882c386c82680cc3219f13926799e5d58`

## Goal

Reject invalid durable recovery input before startup can create or modify any
SQLite database.

## Scope

- Read and fully validate recovery input before database directory creation,
  writable SQLite open, PRAGMA changes, or schema initialization.
- Replay the validated in-memory snapshot after successful store initialization
  without rereading the file.
- Preserve missing-database absence and compatible existing database bytes when
  malformed or semantically invalid recovery input is rejected.
- Retain valid recovery replay, truncation, idempotency, and database
  initialization behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add cross-process recovery-log locking or concurrent-writer support.
- Do not repair, rewrite, quarantine, or partially accept invalid logs.
- Do not change the recovery format, semantic record contract, SQLite schema,
  or daily-shard routing.
- Do not create a release tag or publish artifacts.

## Risks

- Reading the log twice can validate one snapshot and replay another.
- A stale snapshot can hide an unexpected append between preflight and replay.
- Refactoring replay can accidentally change daily-shard grouping or event-ID
  idempotency.

## Validation

- Direct malformed and semantic-invalid startup failures with missing database
  absence and recovery-log byte preservation.
- Direct invalid startup against a compatible database missing an optional
  index, proving database bytes and missing-index state remain unchanged.
- Valid missing/existing database replay, truncation, and idempotency coverage.
- Repeated focused alert-store race tests, full native tests, E2E smoke,
  documentation, knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Invalid recovery input is fully rejected before any writable database open.
- A missing target database remains absent after rejection.
- A compatible existing target and its optional-index state remain unchanged
  after rejection.
- The exact preflight snapshot, not a second read, is used for startup replay.
- Valid replay and initialization behavior remains compatible.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires cross-process recovery locking, automatic log
repair, a recovery-format change, operator data, or tag/publication authority.

## Local Validation

- Malformed, truncated, and semantic-invalid recovery input leaves a missing
  target database absent and preserves the complete recovery log.
- Malformed and semantic-invalid input leaves a compatible existing database
  byte-for-byte unchanged; a separate read-only handle proves its deliberately
  missing optional index was not recreated.
- Valid recovery input still initializes and replays into missing and compatible
  existing databases, recreates optional indexes, truncates after persistence,
  and remains event-ID idempotent.
- Twenty uncached focused race runs, the full alert-store race package, full
  native C/Go race suite, E2E smoke, documentation, JSON, and diff checks passed.
- Feature commit: `9a13283c3124ca270f39ca9ec63573e94283438c`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `4e13a25882c386c82680cc3219f13926799e5d58..9a13283c3124ca270f39ca9ec63573e94283438c`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
