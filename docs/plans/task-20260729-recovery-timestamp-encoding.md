# Task Plan: R90-48 validate recovery timestamp encoding

## Metadata

- Timestamp: 2026-07-29T00:21:43-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `f164f048f95ec326577996bf614ab3a4c3e66bbc`

## Goal

Require every durable recovery timestamp to use the exact canonical UTC
RFC3339Nano string emitted by the JSON writer before replay or runtime append
can modify the recovery log or SQLite state.

## Scope

- Inspect the original JSON string for `timestamp`, `first_seen`, `last_seen`,
  and `window_start` while decoding each recovery record.
- Reject explicit UTC offsets, equivalent non-UTC offsets, and nonminimal
  fractional forms even when Go parses them to the expected instant.
- Preserve the complete recovery log and missing or compatible existing
  database on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output and replay behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the recovery JSON structure, timestamp values, aggregation
  window, event identity, SQLite schema, or public time semantics.
- Do not rewrite, normalize, truncate, or automatically repair rejected input.
- Do not broaden this increment to duplicate JSON keys or unrelated durable
  field contracts.
- Do not create a release tag or publish artifacts.

## Risks

- Decoding directly into `time.Time` discards the original offset and
  fractional spelling needed for exact validation.
- A second raw-field inspection must agree with the model decoder and report a
  field-specific integrity error without accepting writer-incompatible text.
- Test fixtures must retain consistent instants and identities so timestamp
  encoding failures are not obscured by later normalized-record validation.

## Validation

- Direct startup regressions cover all four timestamp fields with explicit UTC
  offsets, equivalent non-UTC offsets, and nonminimal fractional forms.
- Every startup rejection preserves the complete two-record log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime preflight regressions cover the same twelve field/form
  combinations and preserve the existing log plus SQLite bytes.
- Canonical writer output replays and truncates normally.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- The four durable timestamp strings must each equal the UTC RFC3339Nano value
  that NetSentry's writer emits for the decoded instant.
- Every planned alternate offset and fractional form fails through
  `ErrRecoveryLogIntegrity` with record and field context before state changes.
- Startup and runtime preservation boundaries pass direct byte checks.
- Canonical recovery records remain compatible.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Validation Evidence

- All twelve field/form combinations reject through the intended canonical
  timestamp error during startup and runtime preflight.
- Missing-database startup fixtures remain absent; compatible existing
  databases with a deliberately absent optional index remain byte-for-byte
  unchanged and uninitialized.
- Runtime fixtures preserve the complete recovery log and SQLite bytes without
  appending or persisting the current batch.
- Canonical writer strings for all four fields replay and truncate normally.
- Twenty uncached focused alert-store race runs passed.
- The complete native race suite passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge, JSON, diff, and
  sensitive-information checks passed.

## Delivery Evidence

- Feature commit: `6df3d8f45b2c581cf49c3b40e00198ba59dbc20e`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `f164f048f95ec326577996bf614ab3a4c3e66bbc..6df3d8f45b2c581cf49c3b40e00198ba59dbc20e`
- Vault feature note, full commit index, MOC link, and reusable stable storage
  note: verified.
- No later engineering increment was selected; the next `$netsentry-next`
  trigger will refresh the rolling roadmap.

## Stop Conditions

Stop if completion requires a recovery-format migration, accepting a timestamp
the current writer cannot emit, automatic repair, access to operator data, tag
creation, or publication.
