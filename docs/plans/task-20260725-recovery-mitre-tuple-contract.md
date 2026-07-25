# Task Plan: R90-42 validate recovery MITRE tuples

## Metadata

- Timestamp: 2026-07-25T11:32:23-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `da25754b70df111746853b8bfc817db6e44cb4e5`

## Goal

Reject durable recovery records whose MITRE tactic, technique ID, and technique
name are not either all empty or all nonblank before startup replay or runtime
append can modify the recovery log or SQLite state.

## Scope

- Accept a recovery MITRE tuple when all three fields are exactly empty.
- Accept a recovery MITRE tuple when each field contains a non-whitespace
  character.
- Reject every partial empty/populated tuple shape.
- Reject a tuple containing a whitespace-only tactic, technique ID, or
  technique name.
- Apply validation through the shared startup and runtime recovery preflight.
- Preserve the complete log and missing/existing database state on rejection.
- Preserve valid complete tuple text exactly without normalization.
- Keep stored-row MITRE tuple validation compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not validate recovery tuples against the current MITRE catalog.
- Do not normalize case, trim text, or change alert filter semantics.
- Do not change the recovery format or stored-row contract.
- Do not migrate, repair, delete, or rewrite recovery or SQLite data.
- Do not change the SQLite schema.
- Do not create a release tag or publish artifacts.

## Risks

- Partial recovery tuples can be persisted successfully and fail only during a
  later row read.
- Catalog revalidation would make current code decide historical recovery tuple
  validity.
- Validation must remain in shared preflight so startup and runtime behavior
  cannot diverge.

## Validation

- Direct rejection for all six non-empty proper subsets of the three MITRE
  fields with a valid recovery prefix.
- Direct whitespace-only rejection for each tuple member.
- Missing and compatible existing database startup preservation.
- Runtime recovery-log and database byte preservation before append.
- Direct all-empty and fully populated replay compatibility without
  normalization.
- Existing stored-row MITRE tuple compatibility regressions.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Recovery MITRE tuples are either exactly all empty or contain non-whitespace
  text in all three fields.
- Every partial or whitespace-only tuple fails before startup or runtime state
  changes.
- Startup rejection preserves the complete log and both missing and compatible
  existing database state.
- Runtime rejection leaves the log and database unchanged without appending the
  current batch.
- Complete tuple values replay unchanged; all-empty tuples remain valid.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires current-catalog revalidation, text normalization,
a recovery-format change, stored-row migration, automatic repair or deletion,
operator data, tag creation, or publication authority.

## Delivery Evidence

- Feature commit: `4780f02688fabf75e89b954bc4f3f0982c0d1f6a`
- Shared recovery MITRE tuple validation implemented with all six partial
  shapes, each whitespace-only field, missing/existing database preservation,
  runtime preservation, all-empty compatibility, and complete unnormalized
  replay regressions.
- Twenty uncached focused alert-store race runs, the complete native race
  suite, E2E smoke, documentation, knowledge, JSON, diff, and
  sensitive-information checks passed.
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range:
  `da25754b70df111746853b8bfc817db6e44cb4e5..4780f02688fabf75e89b954bc4f3f0982c0d1f6a`
- Vault iteration note, full commit index, `00-MOC` link, and reusable stable
  storage note: verified.
