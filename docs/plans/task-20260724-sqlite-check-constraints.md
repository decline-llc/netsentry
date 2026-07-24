# Task Plan: R90-25 reject write-affecting SQLite check constraints

## Metadata

- Timestamp: 2026-07-24T00:11:08-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `a1aeefeb4071a3844dd4a69d1c1e66fa10f679b2`

## Goal

Reject existing schemas with `CHECK` constraints on NetSentry's
write-critical tables before writable SQLite initialization.

## Scope

- Inspect the stored `CREATE TABLE` SQL for `alerts` and `alert_events` through
  the existing read-only preflight handle.
- Detect the `CHECK` keyword only at SQLite lexical boundaries outside quoted
  strings, quoted identifiers, and comments.
- Apply the same boundary to primary databases and historical shard writes.
- Keep constraints on unrelated operator tables compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not parse, evaluate, classify, remove, or rewrite constraint expressions.
- Do not reject constraints on unrelated tables.
- Do not migrate or rewrite operator schemas.
- Do not change NetSentry's canonical schema or SQL write statements.
- Do not create a release tag or publish artifacts.

## Risks

- A raw substring match can falsely reject `CHECK` inside strings, comments,
  or quoted identifiers.
- Evaluating arbitrary expressions outside SQLite can accept a write blocker
  or reject based on incomplete semantics.
- A primary-only fix can leave historical shard writes exposed.

## Validation

- Direct `alerts` table-level, `alerts` column-level, and `alert_events`
  constraint rejections with named errors and byte preservation.
- Case-variant keyword rejection and quoted/comment/string false-positive
  lexical coverage.
- Direct historical-shard constraint rejection with byte preservation.
- Unrelated-table constraint compatibility with successful NetSentry writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every `CHECK` constraint on `alerts` or `alert_events` returns
  `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Constraints confined to unrelated operator tables remain compatible and do
  not block NetSentry writes.
- Strings, comments, and quoted identifiers containing `CHECK` do not create
  false rejections.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires evaluating constraint expressions, schema
migration, rewriting operator constraints, operator data, or publication
authority.

## Local Validation

- Table-level and column-level `CHECK` constraints on both write-critical
  tables fail with `ErrDatabaseIntegrity` and byte preservation.
- A historical shard with a `CHECK` constraint fails before writable open and
  remains byte-for-byte unchanged.
- Quoted identifiers, strings, comments, and identifier substrings containing
  `CHECK` do not create false rejections.
- An unrelated-table constraint remains active while complete NetSentry
  alert/event writes succeed.
- Twenty uncached focused check-constraint race runs pass.
- The complete uncached native suite, E2E smoke, documentation, knowledge,
  JSON, diff, and sensitive-data checks pass.
- Feature commit: `1a4f565b1ef07b91a0c5ce80efc7cc78c382bb5b`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `a1aeefeb4071a3844dd4a69d1c1e66fa10f679b2..1a4f565b1ef07b91a0c5ce80efc7cc78c382bb5b`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
