# Task Plan: R90-36 enforce the recovery IPv4 address contract

## Metadata

- Timestamp: 2026-07-25T09:43:54-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `6306fce9b5bad932ca3d907fefaf7640eb5a318a`

## Goal

Reject recovery-log records whose source or destination address is not strict
IPv4 before startup replay or runtime append can modify durable state.

## Scope

- Require strict IPv4 `src_ip` and `dst_ip` values during recovery-record
  semantic validation.
- Reject malformed, ordinary IPv6, and IPv4-mapped IPv6 address text.
- Preserve the complete recovery log and missing or existing SQLite state on
  startup rejection.
- Preserve the recovery log and SQLite bytes when runtime preflight rejects an
  invalid pending record.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add IPv6 support or normalize address text.
- Do not change the recovery JSONL format or aggregation identity.
- Do not validate or migrate already persisted SQLite alert rows.
- Do not repair, truncate, delete, or rewrite invalid recovery input.
- Do not create a release tag or publish artifacts.

## Risks

- Address validation must precede identity derivation so failures identify the
  corrupt address rather than a dependent ID mismatch.
- A startup failure must not create a missing database or initialize indexes in
  an existing database.
- Runtime rejection must happen before appending a new record.

## Validation

- Direct malformed, ordinary IPv6, and IPv4-mapped IPv6 rejection for both
  `src_ip` and `dst_ip`.
- Missing and existing database startup preservation with a valid prefix.
- Runtime preflight rejection with full log and database byte preservation.
- Valid startup replay and runtime write compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Recovery records accept only strict IPv4 source and destination addresses.
- Every invalid address fails with `ErrRecoveryLogIntegrity` before SQLite
  modification.
- Startup rejection preserves the full log and missing/existing database state.
- Runtime rejection appends nothing and preserves both recovery log and
  database bytes.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires IPv6 support, a recovery-format change, address
normalization, stored-row migration, automatic log repair, operator data, tag
creation, or publication authority.

## Delivery Evidence

- Implementation and all required validation pass.
- The first full native suite exposed an R90-29 collation fixture that wrote
  IPv6 through recovery. Rule/severity cases now use valid IPv4 input, while
  source/destination collation-only rows seed SQLite below the recovery
  boundary.
- The adjusted fixture passed twenty uncached race runs; the combined focused
  suite and complete native race suite then passed.
- E2E smoke, documentation, knowledge, JSON, diff, and
  sensitive-information checks pass.
- Pending commit, push, fetched-remote verification, and exact-range Vault
  synchronization.
