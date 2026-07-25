# Task Plan: R90-30 validate stored SQLite alert numerics

## Metadata

- Timestamp: 2026-07-25T06:54:12-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `2aabb40cb0058e4c0ffbde428edfd25ed1d9ecf2`

## Goal

Reject persisted alert rows whose numeric values cannot satisfy the public alert
model instead of silently narrowing or returning them.

## Scope

- Reject stored destination ports below zero or above 65535 before converting
  them to `uint16`.
- Reject stored aggregate counts that are zero or negative.
- Cover primary list/query decoding and historical cross-shard reads.
- Prove a rejected historical shard remains byte-for-byte unchanged.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not repair, delete, clamp, migrate, or rewrite invalid rows.
- Do not change the SQLite schema, writer normalization, recovery-log format,
  or API validation.
- Do not add a full-table startup scan or broaden this increment to textual or
  timestamp invariants.
- Do not create a release tag or publish artifacts.

## Risks

- A direct integer conversion can hide negative or oversized ports through
  unsigned wrapping.
- Accepting non-positive aggregate counts can expose states the writer cannot
  produce.
- Historical rejection must stay on the existing read-only query path.

## Validation

- Direct negative and above-65535 destination-port rejection regressions.
- Direct zero and negative aggregate-count rejection regressions.
- Direct historical-shard rejection with byte preservation.
- Healthy primary and cross-shard query compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Every stored destination port outside `0..65535` returns a field-specific
  read error before conversion.
- Every stored aggregate count below one returns a field-specific read error.
- Valid rows retain current list, query, count, and cross-shard behavior.
- A rejected historical shard remains byte-for-byte unchanged.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires automatic row repair, deletion, schema migration, a
full-table startup scan, operator data, tag creation, or publication authority.

## Delivery Evidence

- Feature commit: `23679d6fbf6619315b6260e614dad62b2f3c2863`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `2aabb40cb0058e4c0ffbde428edfd25ed1d9ecf2..23679d6fbf6619315b6260e614dad62b2f3c2863`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
