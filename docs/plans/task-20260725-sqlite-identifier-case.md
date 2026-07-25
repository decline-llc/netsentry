# Task Plan: R90-28 honor SQLite identifier case semantics

## Metadata

- Timestamp: 2026-07-25T06:31:02-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `544ca7498c2623daaaf25381b04ef6793de49e5c`

## Goal

Treat required SQLite table columns and unique-index key columns with SQLite's
case-insensitive identifier semantics during read-only schema preflight.

## Scope

- Match required columns case-insensitively while preserving their definition
  checks and unknown-column policy.
- Match canonical aggregation and safe unique-index identities
  case-insensitively while retaining binary-collation requirements.
- Prove primary and historical schemas with case-variant required identifiers
  remain writable.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the canonical schema, aggregation identity, or write SQL.
- Do not weaken type, nullability, primary-key, generated-column, collation,
  uniqueness, trigger, check-constraint, or foreign-key validation.
- Do not migrate or rewrite operator schemas.
- Do not create a release tag or publish artifacts.

## Risks

- Normalizing only required-column lookup could still reject index metadata
  that preserves the declared casing.
- Over-broad normalization could conceal a genuinely unknown required column
  or weaken binary-collation enforcement.
- Historical compatibility must not bypass the existing read-only preflight.

## Validation

- Direct primary-database regression with case-variant required tables,
  columns, canonical aggregation key, and compatible identity uniqueness.
- Direct historical-shard case-variant compatibility regression.
- Existing incompatible-schema and binary-collation rejection regressions.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Preflight matches required and indexed identifiers with SQLite's
  case-insensitive semantics.
- A compatible primary database with case-variant required identifiers opens
  and accepts writes.
- A compatible historical shard with case-variant required identifiers accepts
  writes through the existing preflight path.
- All existing rejection boundaries remain enforced.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires schema migration, identifier rewriting, weakening a
write-safety constraint, operator data, tag creation, or publication authority.
