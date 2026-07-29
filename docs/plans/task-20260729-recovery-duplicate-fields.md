# Task Plan: R90-49 reject duplicate recovery JSON fields

## Metadata

- Timestamp: 2026-07-29T00:44:46-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `e1434794348c5cc01b075f5e375c98027ec34a15`

## Goal

Reject ambiguous duplicate top-level recovery JSON names before Go's model
decoder silently keeps the last value and discards evidence of the earlier
member.

## Scope

- Inspect every syntactically valid recovery record's top-level JSON member
  names before model decoding.
- Reject identical duplicate durable names whether their values agree or
  conflict.
- Reject ASCII case-variant names that target the same supported model field
  under Go's case-insensitive matching behavior.
- Reject repeated unknown top-level names without otherwise tightening the
  accepted unknown-field contract.
- Preserve complete recovery logs and missing or compatible existing database
  state on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output and existing malformed/missing-field
  diagnostics.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not reject a single unknown field or a single case-variant supported name;
  those broader canonical-field-name decisions remain separate.
- Do not inspect duplicate names recursively inside unknown nested JSON values.
- Do not change JSON field ordering, recovery encoding, normalized alert
  semantics, SQLite schema, or public API behavior.
- Do not rewrite, normalize, truncate, or automatically repair rejected input.
- Do not create a release tag or publish artifacts.

## Risks

- Returning a duplicate error from a structurally malformed record could
  obscure the established JSON decode diagnostic.
- Exact-string tracking alone would miss aliases that Go maps to the same
  exported model field.
- A member scanner must consume complete values without applying new schema
  restrictions to otherwise ignored unknown fields.

## Validation

- Direct startup regressions reject agreeing duplicates, conflicting
  duplicates, case-variant durable aliases, and repeated unknown names.
- Every startup rejection preserves a complete valid-prefix log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime regressions cover the same four rejection classes and
  preserve the existing log plus SQLite bytes.
- Existing malformed and missing-field diagnostics remain unchanged.
- Canonical writer output replays and truncates normally.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- A syntactically valid recovery record cannot contain the same top-level JSON
  name twice.
- Two case variants that target one supported durable model field are treated
  as a duplicate.
- Every planned rejection reports the record and duplicate name through
  `ErrRecoveryLogIntegrity` before durable state changes.
- Startup and runtime preservation boundaries pass direct byte checks.
- Canonical records and established non-duplicate error behavior remain
  compatible.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Validation Evidence

- Agreeing and conflicting duplicate durable fields, a case-variant durable
  alias, and a repeated unknown top-level name reject through the intended
  record-scoped integrity error during startup and runtime preflight.
- Missing-database startup fixtures remain absent; compatible existing
  databases with a deliberately absent optional index remain byte-for-byte
  unchanged and uninitialized.
- Runtime fixtures preserve the complete recovery log and SQLite bytes without
  appending or persisting the current batch.
- Malformed input containing a duplicate retains its original JSON decode
  error; canonical records, a single unknown field, nested unknown duplicates,
  and a single case-variant supported name remain compatible.
- Twenty uncached focused alert-store race runs passed.
- The first complete native suite hit the existing receiver idle-timeout timing
  boundary; twenty uncached focused receiver race reruns and the required clean
  complete native rerun passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge, JSON, diff, and
  sensitive-information checks passed.

## Stop Conditions

Stop if completion requires recursively constraining unknown nested JSON,
rejecting all unknown or case-variant fields, changing the recovery format,
automatic repair, access to operator data, tag creation, or publication.
