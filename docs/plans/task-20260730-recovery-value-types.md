# Task Plan: R90-52 enforce recovery JSON value types

## Metadata

- Timestamp: 2026-07-30T09:09:29-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `13cfba3979da7526ac62c0a022bbb0d16f60ce1e`

## Goal

Require every top-level recovery JSON value to use the same JSON kind and
non-null representation emitted by the current alert writer before model
decoding can turn `null` into an accepted Go zero value.

## Scope

- Extend the complete top-level member scan before model decoding.
- Require strings for text and timestamp fields and JSON numbers for
  `dst_port` and `aggregated_count`.
- Reject JSON `null` for required fields and for present optional
  `raw_payload`.
- Report the first value-contract error in deterministic record order.
- Preserve duplicate-field, unsupported-name, and malformed-JSON diagnostic
  precedence.
- Preserve complete recovery logs and missing or compatible existing database
  state on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output with omitted and populated `raw_payload`.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not require `raw_payload`, constrain the contents of string values, or
  recursively inspect arrays or objects.
- Do not add a separate canonical numeric-spelling policy beyond the current
  typed model decoder.
- Do not change normalized alert semantics, SQLite schema, or public API
  behavior.
- Do not add a versioned recovery migration, rewrite rejected input, or create
  a release tag or public artifact.

## Risks

- A hardcoded field-kind map can drift from the writer's model tags and Go
  types.
- Returning a value error before the structural scan finishes can obscure a
  later duplicate, unsupported-name, or malformed-record diagnostic.
- Tests that cover only values already rejected by semantic validation would
  miss `null` values that currently pass as valid port, optional text, or an
  all-empty MITRE tuple.

## Validation

- Direct startup regressions replace every writer field with JSON `null` and
  require a deterministic field-specific value-contract error.
- Direct startup type regressions exercise representative text, timestamp, and
  numeric fields with the wrong JSON kind.
- Every startup rejection preserves a complete valid-prefix log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime regressions cover the same null and type classes while
  preserving the existing log plus SQLite bytes.
- Records that are also duplicate, unsupported, or malformed retain the
  established diagnostic precedence.
- Canonical writer records replay with omitted and populated `raw_payload`,
  and a writer-alignment check covers every emitted field and expected kind.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Every present top-level recovery value has the JSON kind emitted for that
  field by the current `model.Alert` JSON writer.
- JSON `null` fails through `ErrRecoveryLogIntegrity` with record and field
  context before durable state changes.
- `raw_payload` remains optional but must be a string when present.
- Duplicate, unsupported-name, and malformed diagnostics retain precedence.
- Startup and runtime preservation boundaries pass direct byte checks.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Validation Evidence

- JSON `null` in all 20 writer fields rejects during startup and runtime
  preflight with record- and field-specific value-contract errors.
- Representative wrong-kind text, timestamp, port, and aggregate-count values
  reject through the same paths.
- Missing-database startup fixtures remain absent; compatible existing
  databases with a deliberately absent optional index remain byte-for-byte
  unchanged and uninitialized.
- Runtime fixtures preserve the complete recovery log and SQLite bytes without
  appending or persisting the current batch.
- Duplicate and unsupported-name diagnostics retain precedence over missing
  and invalid values; structurally malformed records retain their original
  decode error.
- Canonical writer records with omitted and populated `raw_payload` satisfy
  every configured field-kind contract and replay successfully.
- Twenty uncached focused alert-store race runs passed.
- The complete native race suite passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge, JSON, and diff checks passed.

## Delivery Evidence

- Feature commit: `f4985bb7fc3b6f50a5f90aa13d4d482cd712695c`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `13cfba3979da7526ac62c0a022bbb0d16f60ce1e..f4985bb7fc3b6f50a5f90aa13d4d482cd712695c`
- Vault feature note, full commit index, MOC link, and reusable stable storage
  note: verified.
- No later engineering increment was selected; the next `$netsentry-next`
  trigger will refresh the rolling roadmap.

## Stop Conditions

Stop if completion requires a versioned recovery migration, making
`raw_payload` mandatory, recursively constraining JSON values, changing the
recovery format, automatic repair, operator data, tag creation, or
publication.
