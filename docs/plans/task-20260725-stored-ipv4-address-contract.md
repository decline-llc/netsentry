# Task Plan: R90-37 validate stored SQLite IPv4 addresses

## Metadata

- Timestamp: 2026-07-25T09:56:17-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `9b5e4299d18e774a955ecc7b0561d6bbd5aed977`

## Goal

Reject persisted alert rows whose source or destination address is not strict
IPv4 before list or query reads can expose the row or derive its aggregation
identity.

## Scope

- Require strict IPv4 `src_ip` and `dst_ip` values in the shared stored-row
  decoder.
- Reject malformed, ordinary IPv6, and IPv4-mapped IPv6 address text in both
  address fields.
- Apply the same behavior to primary and historical list/query paths.
- Preserve historical shard bytes when a read rejects an invalid row.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add IPv6 support or normalize address text.
- Do not scan all stored rows during startup.
- Do not validate or reconcile the event ledger.
- Do not repair, delete, migrate, or rewrite persisted rows.
- Do not change the SQLite schema or aggregation identity.
- Do not create a release tag or publish artifacts.

## Risks

- Address validation must precede aggregation-identity derivation so failures
  identify the corrupt address rather than a dependent ID mismatch.
- Validation must remain in the shared decoder so primary and historical reads
  cannot diverge.
- Historical rejection must retain the existing read-only preservation
  boundary.

## Validation

- Direct malformed, ordinary IPv6, and IPv4-mapped IPv6 rejection for both
  `src_ip` and `dst_ip` across list and query decoding.
- Historical read-only rejection through an encoded filesystem path with
  byte-for-byte shard preservation.
- Valid primary and historical IPv4 compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Persisted rows accept only strict IPv4 source and destination addresses.
- Every invalid address fails before aggregation-identity derivation.
- Primary and historical list/query paths share the same rejection behavior.
- Historical rejection leaves the shard byte-for-byte unchanged.
- Valid IPv4 rows retain current behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires IPv6 product support, address normalization,
automatic row repair or deletion, schema migration, a full-table startup scan,
event-ledger reconciliation, operator data, tag creation, or publication
authority.

## Delivery Evidence

- Implementation and twenty uncached focused alert-store race runs pass.
- The first full native suite exposed an R90-29 collation fixture whose
  source/destination cases seed case-variant IPv6 below recovery. R90-37
  correctly rejects those stored rows.
- The fixture now uses distinct valid IPv4 rows under its compatible `NOCASE`
  schema. It passed twenty uncached affected-test race runs, followed by clean
  combined focused and complete native race suites.
- E2E smoke, documentation, knowledge, JSON, diff, and
  sensitive-information checks pass.
- Pending commit, push, fetched-remote verification, and exact-range Vault
  synchronization.
