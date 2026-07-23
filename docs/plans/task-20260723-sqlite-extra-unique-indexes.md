# Task Plan: R90-22 reject write-blocking SQLite uniqueness extensions

## Metadata

- Timestamp: 2026-07-23T04:31:25-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `bd0f4d04689513ac9806a69991c97993895bd9e9`

## Goal

Reject existing schemas with extra unique indexes that can block NetSentry's
valid fixed-column alert or event writes, before writable SQLite initialization.

## Scope

- Inspect every unique index on `alerts` and `alert_events` during read-only
  schema preflight.
- Accept uniqueness containing an existing binary-collated safe write identity:
  `alerts.id`, the complete alert aggregation key, or
  `alert_events.event_id`.
- Reject subset, unrelated, or expression-only uniqueness that can collide
  across otherwise valid NetSentry writes.
- Apply the same boundary to primary databases and historical shard writes.
- Keep ordinary non-unique indexes and redundant identity-containing unique
  indexes compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not evaluate arbitrary index expressions or prove application-specific
  operator constraints.
- Do not reject non-unique query indexes.
- Do not migrate, drop, rename, or rewrite operator indexes or tables.
- Do not change NetSentry's canonical schema or SQL write statements.
- Do not create a release tag or publish artifacts.

## Risks

- Rejecting all operator indexes would break compatible query extensions.
- Accepting uniqueness over only part of a write identity can reject valid
  batches after the recovery log has already been appended.
- A primary-only fix can leave historical shard writes exposed.

## Validation

- Direct primary rejection for `alerts(rule_id)`,
  `alert_events(created_at)`, expression-only, partial-subset, and
  non-binary-collated identity unique indexes, with field-specific errors and
  byte preservation.
- Direct historical-shard unique-index rejection with byte preservation.
- Non-unique and binary-identity-containing unique-index compatibility with
  successful multi-alert writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every unique index lacking a complete binary-collated canonical safe identity
  returns `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Ordinary indexes and uniqueness containing an existing safe identity remain
  compatible and do not block NetSentry writes.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires evaluating arbitrary index expressions, schema
migration, rewriting operator indexes, operator data, or publication authority.

## Local Validation

- Primary subset, timestamp-only, expression-only, partial-subset, and
  non-binary-collated identity unique indexes fail with
  `ErrDatabaseIntegrity` and byte preservation.
- A historical shard with a write-blocking unique index fails before writable
  open and remains byte-for-byte unchanged.
- Ordinary indexes plus binary identity-containing unique indexes reopen and
  accept distinct valid writes.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, knowledge, JSON, and diff checks pass.
- Feature commit: `b62cbff41ec3f72adfa07030dcba17058a3e239e`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `bd0f4d04689513ac9806a69991c97993895bd9e9..b62cbff41ec3f72adfa07030dcba17058a3e239e`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
