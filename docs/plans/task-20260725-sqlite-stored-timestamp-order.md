# Task Plan: R90-32 validate stored SQLite timestamp order

## Metadata

- Timestamp: 2026-07-25T07:13:02-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `5fb3e1b8920746a65e4f1eb4bdd90870c98d2b04`

## Goal

Reject syntactically valid persisted alert timestamps whose ordering cannot be
produced by the aggregation writer.

## Scope

- Reject rows whose `first_seen` is after `last_seen`.
- Reject rows whose `window_start` is after `first_seen`.
- Cover primary list/query decoding and historical cross-shard reads.
- Prove a rejected historical shard remains byte-for-byte unchanged.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not infer or require the current aggregation-window duration for historical
  rows.
- Do not normalize time zones, repair, delete, migrate, or rewrite invalid rows.
- Do not change timestamp syntax parsing, writer normalization, the SQLite
  schema, or the recovery-log format.
- Do not create a release tag or publish artifacts.

## Risks

- Reversed first/last timestamps can break ordering and aggregation semantics.
- A window start after the first event cannot describe that aggregate.
- Historical rejection must stay on the existing read-only query path.

## Validation

- Direct first-after-last and window-after-first rejection regressions across
  list and query decoding.
- Direct historical-shard rejection with byte preservation.
- Healthy primary and cross-shard timestamp compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Every decoded row satisfies `first_seen <= last_seen`.
- Every decoded row satisfies `window_start <= first_seen`.
- Valid rows retain current list, query, ordering, pagination, and cross-shard
  behavior.
- A rejected historical shard remains byte-for-byte unchanged.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires automatic row repair, deletion, schema migration,
assuming a historical aggregation-window duration, operator data, tag creation,
or publication authority.

## Delivery Evidence

- Feature commit: `5d8eb60015c977e5f371846faff85ab015002615`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `5fb3e1b8920746a65e4f1eb4bdd90870c98d2b04..5d8eb60015c977e5f371846faff85ab015002615`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
