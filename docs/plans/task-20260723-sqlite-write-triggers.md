# Task Plan: R90-23 reject write-affecting SQLite triggers

## Metadata

- Timestamp: 2026-07-23T04:43:03-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `572038a5e081697f5b4a522687222a510dd1f0f8`

## Goal

Reject existing schemas with triggers attached to NetSentry's write-critical
tables before writable SQLite initialization.

## Scope

- Inspect trigger ownership during read-only schema preflight.
- Reject every trigger attached to `alerts` or `alert_events`, regardless of
  timing or event, because arbitrary trigger SQL can abort or alter fixed
  NetSentry write semantics.
- Apply the same boundary to primary databases and historical shard writes.
- Keep triggers confined to unrelated operator tables compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not parse, interpret, execute, classify, drop, or rewrite trigger bodies.
- Do not reject triggers attached only to unrelated operator tables.
- Do not migrate or rewrite operator schemas.
- Do not change NetSentry's canonical schema or SQL write statements.
- Do not create a release tag or publish artifacts.

## Risks

- Allowing even an apparently benign trigger on a write-critical table leaves
  arbitrary SQL in NetSentry's write path.
- Rejecting triggers globally would break isolated operator extensions.
- A primary-only fix can leave historical shard writes exposed.

## Validation

- Direct primary rejection for `alerts` `BEFORE INSERT`, `alerts`
  `AFTER UPDATE`, `alert_events`, and case-variant table-name triggers, with
  named errors and byte preservation.
- Direct historical-shard trigger rejection with byte preservation.
- Unrelated operator-table trigger compatibility with successful multi-alert
  writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every trigger attached to `alerts` or `alert_events` returns
  `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Triggers confined to unrelated operator tables remain compatible and do not
  block NetSentry writes.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires interpreting or rewriting trigger SQL, schema
migration, operator data, or publication authority.

## Local Validation

- Primary `alerts` `BEFORE INSERT`, `alerts` `AFTER UPDATE`,
  `alert_events`, and case-variant table-name triggers fail with
  `ErrDatabaseIntegrity` and byte preservation.
- A historical shard with a write-critical trigger fails before writable open
  and remains byte-for-byte unchanged.
- A trigger confined to an unrelated operator table remains active while
  NetSentry accepts distinct valid writes.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, knowledge, JSON, and diff checks pass.
- Tag/publication actions: none.
