# Task Plan: R90-26 reject write-affecting SQLite foreign keys

## Metadata

- Timestamp: 2026-07-24T00:40:09-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `b1c60bec99f7d3f367d531836b6d2f26b93d556c`

## Goal

Reject existing schemas with foreign-key relationships involving NetSentry's
write-critical tables before writable SQLite initialization.

## Scope

- Inspect SQLite foreign-key metadata through the existing read-only preflight
  handle.
- Reject outgoing foreign keys declared by `alerts` or `alert_events`.
- Reject incoming foreign keys from other tables that reference `alerts` or
  `alert_events`.
- Apply the same boundary to primary databases and historical shard writes.
- Keep foreign keys confined to unrelated operator tables compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not enable, disable, evaluate, remove, or rewrite foreign-key enforcement.
- Do not reject foreign keys whose source and target are both unrelated tables.
- Do not migrate or rewrite operator schemas.
- Do not change NetSentry's canonical schema or SQL write statements.
- Do not create a release tag or publish artifacts.

## Risks

- A source-only check can miss incoming relationships that affect retention
  deletes from a write-critical table.
- Case-sensitive table matching can miss relationships SQLite resolves
  case-insensitively.
- A database-wide foreign-key ban would unnecessarily reject unrelated
  operator schemas.

## Validation

- Direct outgoing relationships from `alerts` and `alert_events`, and an
  incoming relationship to a write-critical table, with named errors and byte
  preservation.
- Case-variant referenced-table coverage.
- Direct historical-shard rejection with byte preservation.
- Unrelated-table foreign-key compatibility with successful NetSentry writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every foreign key whose source or target is `alerts` or `alert_events`
  returns `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Foreign keys confined to unrelated operator tables remain compatible and do
  not block NetSentry writes.
- SQLite identifier case variants cannot bypass the boundary.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires enabling or evaluating foreign-key actions, schema
migration, rewriting operator constraints, operator data, or publication
authority.

## Local Validation

- Outgoing relationships from both write-critical tables and incoming
  relationships to them fail with `ErrDatabaseIntegrity` and byte preservation.
- Case-variant and implicit-primary-key references cannot bypass preflight.
- A historical shard with a write-critical relationship fails before writable
  open and remains byte-for-byte unchanged.
- Relationships confined to unrelated operator tables remain compatible while
  complete NetSentry alert/event writes succeed.
- Twenty uncached focused foreign-key race runs pass.
- The complete uncached native suite, E2E smoke, documentation, knowledge,
  JSON, and diff checks pass.
- Feature commit: `0ddba61bde65fe1bb5ca9757bc87d06123409251`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `b1c60bec99f7d3f367d531836b6d2f26b93d556c..0ddba61bde65fe1bb5ca9757bc87d06123409251`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
