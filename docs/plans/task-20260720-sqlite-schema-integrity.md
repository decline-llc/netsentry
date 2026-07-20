# Task Plan: R90-15 reject incompatible existing SQLite schemas

## Metadata

- Timestamp: 2026-07-20T05:00:50-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `d848aa3c4ff01101d076889ac2afd329a1b1a2c6`

## Goal

Reject a structurally valid existing SQLite database that does not satisfy the
NetSentry alert-store schema before any writable initialization can modify it.

## Scope

- Extend the existing read-only `quick_check` preflight with read-only schema
  validation for required alert and event tables, columns, types, primary-key
  roles, non-null constraints, and the aggregation uniqueness contract.
- Apply the same preflight to the primary database and existing daily-shard
  write targets.
- Return the stable `ErrDatabaseIntegrity` boundary for incompatible schemas.
- Preserve rejected database bytes and retain compatible existing, empty, and
  missing database behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add automatic schema migration, repair, deletion, or data copying.
- Do not require optional query indexes before writable initialization; current
  initialization may create missing performance indexes.
- Do not change journal modes, recovery-log semantics, query behavior, or
  retention policy.
- Do not create a release tag or publish artifacts.

## Risks

- An over-strict preflight can reject a compatible database.
- An incomplete preflight can let schema initialization or PRAGMA changes
  modify unrelated operator data before failing.
- SQLite metadata quirks around primary keys and auto-indexes can produce false
  compatibility results.

## Validation

- Direct unrelated-schema, missing/incompatible required-column, and missing
  aggregation-uniqueness startup regressions with byte preservation.
- Direct incompatible historical daily-shard write rejection with byte
  preservation.
- Compatible existing, empty, missing, and healthy daily-shard regressions.
- Repeated focused alert-store race tests, full native tests, E2E smoke,
  documentation, knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Every existing non-empty database is checked through a separate read-only
  handle before writable open.
- Missing required tables or columns, incompatible column definitions, and a
  missing aggregation uniqueness contract return `ErrDatabaseIntegrity`.
- Every directly rejected database remains byte-for-byte unchanged.
- Compatible existing, empty, missing, and healthy daily-shard databases keep
  their current behavior.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires automatic schema migration, repair, deletion,
operator data, or a broader storage redesign.

## Local Validation

- Required-table/column, type, non-null, primary-key, and aggregation-uniqueness
  rejection cases passed 20 focused race runs with byte preservation.
- Compatible current databases, existing empty files, missing files, optional
  query-index recreation, healthy daily shards, and a path containing spaces
  retained their expected behavior.
- Incompatible historical daily-shard writes return `ErrDatabaseIntegrity` and
  preserve the rejected shard bytes.
- Full native C and Go race tests, E2E smoke, documentation, knowledge, JSON,
  and diff checks passed.
- Delivery verification is recorded below.

## Completion

- Feature commit: `8e3c2b05f27a1035d0bf3bee3a5e762964d84865`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `d848aa3c4ff01101d076889ac2afd329a1b1a2c6..8e3c2b05f27a1035d0bf3bee3a5e762964d84865`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
