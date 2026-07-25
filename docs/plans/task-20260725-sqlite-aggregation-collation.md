# Task Plan: R90-27 require binary SQLite aggregation collation

## Metadata

- Timestamp: 2026-07-25T03:02:04-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `75b160aa268331e37883bbb1ac3019346ce494f0`

## Goal

Require the canonical alert aggregation uniqueness key to use SQLite binary
collation before writable initialization.

## Scope

- Inspect the full key-column metadata for candidate aggregation indexes.
- Accept the canonical column order only when every key column uses binary
  collation.
- Reject inline and explicit non-binary aggregation uniqueness in primary and
  historical databases without modifying their bytes.
- Preserve compatible binary-collated schemas and distinct case-sensitive
  alert identities.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the canonical aggregation identity or UPSERT statement.
- Do not migrate, rebuild, or rewrite operator indexes.
- Do not alter the R90-22 policy for additional compatible indexes.
- Do not create a release tag or publish artifacts.

## Risks

- Column-only index inspection can accept non-binary collation and merge
  distinct rule, source, destination, or window identities.
- A broad index-policy change could reject compatible operator extensions.
- Rejected primary and historical databases must remain byte-for-byte
  unchanged.

## Validation

- Direct inline and explicit non-binary aggregation-key rejection tests with
  byte preservation.
- Direct historical-shard rejection with byte preservation.
- Binary-collated canonical-key compatibility with writes whose rule IDs
  differ only by case.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- A canonical aggregation uniqueness key is accepted only when all key columns
  use SQLite binary collation.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Compatible binary-collated schemas preserve distinct case-sensitive alert
  identities and remain writable.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

- Stop if completion requires changing the aggregation identity, schema
  migration, rewriting operator indexes, operator data, tag creation, or
  publication authority.

## Delivery Evidence

- Feature commit: `6a40a0aaf9b21d5d8a9ce08b7939d5b7b4ec8241`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `75b160aa268331e37883bbb1ac3019346ce494f0..6a40a0aaf9b21d5d8a9ce08b7939d5b7b4ec8241`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
