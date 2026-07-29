# Task Plan: R90-50 enforce canonical recovery JSON field names

## Metadata

- Timestamp: 2026-07-29T01:14:16-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `6c981dff0757aa8f05d09f3454c735ed4ae05ea4`

## Goal

Require each top-level durable recovery JSON name to equal one exact field tag
the current writer can emit, rather than relying on Go's unknown-field ignore
and case-insensitive matching behavior.

## Scope

- Extend the complete top-level member scan before model decoding.
- Reject a single unknown scalar or nested top-level field.
- Reject a single ASCII case-variant alias of a supported durable field and
  report its canonical writer name.
- Preserve duplicate-field precedence when a record contains both duplicate
  and unsupported names.
- Preserve malformed JSON diagnostics by returning vocabulary errors only
  after a complete structural parse.
- Preserve complete recovery logs and missing or compatible existing database
  state on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output with omitted and populated optional
  `raw_payload`.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not recursively inspect the names inside JSON values.
- Do not change JSON field ordering, recovery encoding, normalized alert
  semantics, SQLite schema, or public API behavior.
- Do not add a versioned recovery migration or compatibility mode for another
  writer schema.
- Do not rewrite, normalize, truncate, or automatically repair rejected input.
- Do not create a release tag or publish artifacts.

## Risks

- A hardcoded vocabulary can drift from the writer's model tags.
- Treating every lowercase name as supported would retain unknown-field
  ambiguity, while case-insensitive acceptance would retain aliases the writer
  cannot emit.
- Field-name rejection must not hide the more specific duplicate or malformed
  record error.

## Validation

- Direct startup regressions reject one unknown scalar, one unknown nested
  object, and one case-variant supported name.
- Every startup rejection preserves a complete valid-prefix log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime regressions cover the same three rejection classes and
  preserve the existing log plus SQLite bytes.
- A record with duplicate and unknown names retains duplicate error precedence;
  malformed input retains its JSON decode error.
- Canonical writer records replay with empty and populated `raw_payload`, and
  every emitted top-level name is recognized by the vocabulary.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- Every syntactically valid top-level recovery JSON name equals one exact
  current writer field tag.
- Unknown and noncanonical case aliases fail through `ErrRecoveryLogIntegrity`
  with record and field context before durable state changes.
- Duplicate and malformed diagnostics retain their planned precedence.
- Startup and runtime preservation boundaries pass direct byte checks.
- Canonical records, including optional raw payload output, remain compatible.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires a versioned recovery migration, accepting a field
the current writer cannot emit, recursively constraining JSON values, changing
the recovery format, automatic repair, access to operator data, tag creation,
or publication.
