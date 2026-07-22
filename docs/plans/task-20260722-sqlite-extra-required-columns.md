# Task Plan: R90-21 reject write-blocking SQLite schema extensions

## Metadata

- Timestamp: 2026-07-22T02:25:08-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `17d30a38b12b14b51e050789974f76397bea9a02`

## Goal

Reject existing schemas with unknown mandatory columns that NetSentry's fixed
alert/event inserts cannot populate, before any writable SQLite initialization.

## Scope

- Retain SQLite default metadata while reading required table definitions.
- Reject unknown `NOT NULL` columns without usable non-NULL defaults in both
  `alerts` and `alert_events`.
- Apply the same boundary to primary databases and historical shard writes.
- Keep unknown nullable columns and `NOT NULL` columns with non-NULL defaults
  compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not reject all extra columns or evaluate arbitrary SQLite expressions.
- Do not migrate, drop, rename, populate, or rewrite operator schema.
- Do not change NetSentry's canonical schema or SQL insert columns.
- Do not create a release tag or publish artifacts.

## Risks

- Discarding `PRAGMA table_info` default metadata causes false rejection.
- Allowing an extra mandatory column without a default defers failure until
  after writable initialization and recovery append.
- A primary-only fix can leave historical shard writes exposed.

## Validation

- Direct primary startup rejection for extra mandatory columns in `alerts` and
  `alert_events`, including a literal NULL default, with byte preservation and
  field-specific errors.
- Direct historical-shard write rejection with byte preservation.
- Nullable and non-null defaulted extra columns reopen and accept alert writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every unknown `NOT NULL` column without a usable non-NULL default returns
  `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Unknown nullable and non-NULL-defaulted columns remain compatible and do not
  block NetSentry inserts.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires schema migration, evaluating arbitrary default
expressions, rewriting operator tables, operator data, or publication authority.

## Local Validation

- Unknown mandatory columns without defaults in `alerts` and `alert_events`
  fail primary startup with field-specific `ErrDatabaseIntegrity` errors.
- An unknown `NOT NULL DEFAULT NULL` column is rejected as lacking a usable
  default.
- A historical shard with an unknown mandatory column fails before writable
  open and remains byte-for-byte unchanged.
- Unknown nullable and `NOT NULL DEFAULT 'legacy'` columns reopen and accept a
  complete alert/event write; the defaulted value is verified.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, JSON, and diff checks pass.
- Feature commit: `352cf8fc96ab70a73a0b3f7e3da0cf4f32245160`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `17d30a38b12b14b51e050789974f76397bea9a02..352cf8fc96ab70a73a0b3f7e3da0cf4f32245160`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
