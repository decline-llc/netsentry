# Task Plan: R90-54 enforce recovery JSON numeric encoding

## Metadata

- Timestamp: 2026-07-30T09:37:28-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `e466c372e21fcc67e2661e701c6ca803b94d5d9f`

## Goal

Require recovery `dst_port` and `aggregated_count` values to use the canonical
unsigned base-10 integer JSON spelling emitted by the current normalized alert
writer before model decoding discards exponent, fractional, or sign details.

## Scope

- Extend top-level recovery value validation for both numeric model fields.
- Accept only `0` or a nonzero ASCII digit followed by ASCII digits at the raw
  JSON value boundary.
- Reject syntactically valid exponent, fractional, and negative-sign spellings
  before model decoding and representation-dependent semantic checks.
- Preserve the established malformed-JSON diagnostic for JSON-forbidden
  leading-zero spellings while proving the same no-mutation boundary.
- Report the first numeric-encoding error in deterministic record order.
- Preserve duplicate-field, unsupported-name, missing-field, wrong-kind, and
  malformed-JSON diagnostic precedence.
- Preserve complete recovery logs and missing or compatible existing database
  state on startup rejection.
- Apply the same rejection through runtime existing-log preflight before
  appending a current batch or touching SQLite.
- Preserve canonical writer output and reconcile storage documentation,
  roadmap status, and task state.

## Non-Goals

- Do not change destination-port or aggregate-count semantic ranges.
- Do not accept JSON syntax forbidden by the standard, including a leading
  plus sign or a multi-digit number beginning with zero.
- Do not generalize the independently maintained recovery field contract;
  R90-55 owns that follow-up.
- Do not change the recovery JSON structure, SQLite schema, public API,
  normalized alert behavior, or record-size policy.
- Do not add a recovery migration, rewrite rejected input, create a release
  tag, publish an artifact, or change release authority.

## Risks

- Valid JSON exponent and fractional spellings can decode to the same Go
  integer and bypass semantic checks unless raw bytes are validated first.
- A negative-zero spelling can decode to an otherwise valid zero port, so
  semantic range validation alone cannot enforce the writer representation.
- Returning a numeric error before the complete structural scan finishes can
  obscure a later duplicate, unsupported-name, or malformed-record diagnostic.
- Treating a leading-zero record as structurally valid would replace the
  established JSON decode error with a misleading contract error.

## Validation

- Direct startup regressions cover fractional, exponent, and negative-sign
  spellings for both `dst_port` and `aggregated_count`.
- Direct malformed startup regressions cover a leading-zero spelling for both
  fields and retain the established decode diagnostic.
- Every startup rejection preserves a complete valid-prefix log and leaves a
  missing database absent or a compatible existing database byte-for-byte
  unchanged.
- Direct runtime regressions cover every planned alternate spelling for both
  fields while preserving the complete existing log plus SQLite bytes.
- Duplicate, unsupported-name, missing-field, wrong-kind, and malformed
  diagnostics retain their established precedence.
- Canonical writer records prove both numeric fields use and satisfy the raw
  integer encoding contract.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- `dst_port` and `aggregated_count` accept only the canonical unsigned
  base-10 integer spelling emitted by the normalized writer.
- Valid JSON exponent, fractional, and negative-sign alternatives fail through
  `ErrRecoveryLogIntegrity` with record and field context before model
  decoding or durable state changes.
- JSON-forbidden leading-zero alternatives fail as malformed input before
  durable state changes without changing diagnostic precedence.
- Startup and runtime preservation boundaries pass direct byte checks for
  every planned representation class across both fields.
- Canonical writer output remains compatible.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the single local Vault.

## Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Canonical raw integer spelling | Unit checks against writer-produced `dst_port` and `aggregated_count` JSON values |
| Fraction/exponent/sign rejection | Startup and runtime table tests for both numeric fields with field-specific integrity errors |
| Leading-zero rejection | Startup and runtime malformed-record cases for both fields with decode-error precedence |
| No durable mutation | Complete log bytes, missing/existing database state, SQLite bytes, and alert counts |
| Compatibility | Canonical writer replay plus focused and complete suites |
| Delivery | Fetched remote equality, post-fetch knowledge gate, exact-range Vault note/index/MOC, and stable storage note |

## Validation Deviation

- **Observed:** The first combined formatting and focused-test command used
  repository-relative Go paths while already running from the `engine`
  directory. Formatting therefore reported missing files, but the shell
  continued to a passing focused test.
- **Impact:** Neither formatting nor that later test result was accepted as
  evidence because the multi-step command did not fail fast.
- **Resolution:** The complete sequence was rerun from the repository root
  with strict fail-fast shell settings; formatting and focused tests passed.
  The local `netsentry-next` skill now makes this evidence rule explicit.

## Validation Evidence

- Startup and runtime preflight reject fractional, exponent, and
  negative-sign spellings for both `dst_port` and `aggregated_count` with
  field-specific canonical-encoding errors before model decoding.
- JSON-forbidden leading-zero spellings for both fields retain the established
  malformed-record decode diagnostic.
- All startup cases preserve the complete valid-prefix log, leave a missing
  database absent, and preserve a compatible existing database byte-for-byte
  after an independent read-only inspection.
- All runtime cases preserve the complete existing recovery log and SQLite
  bytes without appending or persisting the current batch.
- The first numeric encoding error follows record order, and existing
  duplicate, unsupported-name, missing-field, wrong-kind, and malformed
  diagnostic precedence remains unchanged.
- Canonical writer output satisfies the raw integer encoding contract and
  replays successfully.
- Twenty uncached focused alert-store race runs passed.
- The complete C and Go native race suite passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge (33 tests), task-state JSON,
  formatting, and diff checks passed.

## Delivery Evidence

- Feature commit: `1e138805cdc133b87acd722f319fcc0cc624196f`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `e466c372e21fcc67e2661e701c6ca803b94d5d9f..1e138805cdc133b87acd722f319fcc0cc624196f`
- Vault feature note, full commit index, MOC link, and reusable stable storage
  note: verified.
- R90-55 is the next ready increment and was not started.

## Stop Conditions

Stop if completion requires accepting a writer-impossible numeric spelling,
changing numeric semantics or the recovery format, a versioned migration,
automatic repair, operator data, external publication authority, or starting
R90-55.
