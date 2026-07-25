# Task Plan: R90-39 validate recovery severity

## Metadata

- Timestamp: 2026-07-25T10:18:06-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `ad913f05c90d71b53205d66bc6cc969fd98b8121`

## Goal

Reject recovery records whose severity is not exactly one of the four public
values before startup replay or runtime append can modify durable state.

## Scope

- Accept only `low`, `medium`, `high`, and `critical` during recovery semantic
  validation.
- Reject empty, case-variant, and unsupported severity without normalization.
- Preserve the complete recovery log and missing or existing SQLite state on
  startup rejection.
- Preserve recovery-log and SQLite bytes when runtime preflight rejects an
  existing invalid-severity record.
- Share severity validation with stored-row decoding.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add, rename, normalize, or reclassify severity values.
- Do not change the recovery JSONL format.
- Do not scan or migrate already persisted SQLite rows.
- Do not repair, truncate, delete, or rewrite invalid recovery input.
- Do not create a release tag or publish artifacts.

## Risks

- Case normalization would hide invalid durable input and diverge from stored
  row behavior.
- Validation must precede writable SQLite initialization and runtime append.
- Shared validation must retain existing stored-row error behavior.

## Validation

- Direct empty, uppercase-known, and unsupported recovery severity rejection
  after a valid prefix.
- Missing and existing database startup preservation.
- Runtime preflight rejection with full log and database byte preservation.
- Direct valid replay compatibility for `low`, `medium`, `high`, and
  `critical`.
- Existing stored-row severity rejection compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Recovery severity accepts exactly the four public values.
- Empty, case-variant, and unsupported values fail with
  `ErrRecoveryLogIntegrity` before SQLite modification.
- Startup rejection preserves the complete log and missing/existing database
  state.
- Runtime rejection appends nothing and preserves recovery-log and database
  bytes.
- All four valid values replay successfully without substitution.
- Stored-row severity errors retain their current behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires changing or normalizing the public severity enum,
rewriting the recovery format, stored-row migration, automatic repair,
operator data, tag creation, or publication authority.

## Delivery Evidence

- Shared recovery/stored-row severity validation implemented with direct
  startup, runtime, valid-enum, and compatibility regressions.
- Twenty uncached focused alert-store race runs passed.
- The first complete native race suite exposed a related fixture that used the
  regular writer to manufacture an invalid stored severity. The fixture now
  seeds that intentionally invalid row directly for its collation-filter
  concern; the affected focused race run and complete native race suite then
  passed.
- `make test`, `make e2e-smoke`, `make docs-check`, explicit-Vault
  `make knowledge-check`, task-state JSON parsing, `git diff --check`, and the
  sensitive-information review passed.
- Pending commit, push, fetched-remote verification, post-fetch knowledge
  validation, and exact-range Vault synchronization.
