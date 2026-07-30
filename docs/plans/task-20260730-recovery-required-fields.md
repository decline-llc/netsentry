# Task Plan: R90-51 require complete recovery JSON records

## Metadata

- Timestamp: 2026-07-30T04:16:05-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `9e93170ec61e0bdfd8928d1ab7a4c4ed6061cf87`

## Goal

Require every top-level recovery JSON field that the current writer always
emits to be present before model decoding can substitute a Go zero value.

## Scope

- Extend the complete top-level member scan before model decoding.
- Require every current writer field except optional `raw_payload`.
- Report the first missing field in deterministic writer order.
- Preserve duplicate-field, unsupported-name, and malformed-JSON diagnostic
  precedence.
- Preserve complete recovery logs and missing or compatible existing database
  state on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output with omitted and populated `raw_payload`.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the JSON value-type or `null` acceptance policy.
- Do not require `raw_payload`, recursively inspect JSON values, or change JSON
  field ordering.
- Do not change normalized alert semantics, SQLite schema, or public API
  behavior.
- Do not add a versioned recovery migration, rewrite rejected input, or create
  a release tag or public artifact.

## Risks

- A hardcoded required-field list can drift from the writer's non-`omitempty`
  tags.
- Returning a missing-field error too early can obscure a duplicate,
  unsupported-name, or malformed-record diagnostic.
- Tests that cover only fields whose zero values already fail semantic
  validation would miss the `dst_port` and optional-text cases that currently
  replay silently when absent.

## Validation

- Direct startup regressions remove each non-`omitempty` writer field in turn
  and require a deterministic field-specific integrity error.
- Every startup rejection preserves a complete valid-prefix log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime regressions remove each required field in turn and preserve
  the existing log plus SQLite bytes.
- Records that are also duplicate, unsupported, or malformed retain the
  established diagnostic precedence.
- Canonical writer records replay with empty and populated `raw_payload`, and
  the test derives the required-field expectation from every emitted field.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Every syntactically valid recovery object contains each field emitted
  unconditionally by the current `model.Alert` JSON writer.
- Missing fields fail through `ErrRecoveryLogIntegrity` with record and field
  context before durable state changes.
- `raw_payload` remains optional and populated writer output remains
  compatible.
- Duplicate, unsupported-name, and malformed diagnostics retain precedence.
- Startup and runtime preservation boundaries pass direct byte checks.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Validation Evidence

- All 19 non-`omitempty` writer fields reject when removed during startup and
  runtime preflight with record- and field-specific integrity errors.
- Missing-database startup fixtures remain absent; compatible existing
  databases with a deliberately absent optional index remain byte-for-byte
  unchanged and uninitialized.
- Runtime fixtures preserve the complete recovery log and SQLite bytes without
  appending or persisting the current batch.
- Duplicate and unsupported-name diagnostics retain precedence over missing
  fields; structurally malformed records retain their original decode error.
- Canonical writer records replay with omitted and populated `raw_payload`, and
  a direct writer-alignment check covers every unconditionally emitted field.
- Twenty uncached focused alert-store race runs passed.
- The complete native race suite passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge, JSON, and diff checks passed.

## Delivery Evidence

- Feature commit: `4a27cece77f0f94b18982677c7562fac1e754b93`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `9e93170ec61e0bdfd8928d1ab7a4c4ed6061cf87..4a27cece77f0f94b18982677c7562fac1e754b93`
- Vault feature note, full commit index, MOC link, and reusable stable storage
  note: verified.
- No later engineering increment was selected; the next `$netsentry-next`
  trigger will refresh the rolling roadmap.

## Stop Conditions

Stop if completion requires a versioned recovery migration, making
`raw_payload` mandatory, changing JSON value/null semantics, changing the
recovery format, automatic repair, operator data, tag creation, or
publication.
