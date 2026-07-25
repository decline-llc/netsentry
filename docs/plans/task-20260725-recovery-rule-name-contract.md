# Task Plan: R90-40 validate recovery rule names

## Metadata

- Timestamp: 2026-07-25T10:31:07-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `e40c15b2028c1a7aa4c7185a575553659c3c9ec8`

## Goal

Reject recovery records whose `rule_name` is missing or blank before startup
replay or runtime append can modify durable state.

## Scope

- Require `rule_name` to contain at least one non-whitespace character during
  recovery semantic validation.
- Reject missing, empty, and whitespace-only names without normalization.
- Preserve the complete recovery log and missing or existing SQLite state on
  startup rejection.
- Preserve recovery-log and SQLite bytes when runtime preflight rejects an
  existing blank-name record.
- Preserve nonblank padded rule names exactly during replay.
- Retain stored-row required-text behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not trim, normalize, rename, or otherwise rewrite rule names.
- Do not change rule schema, alert identity, or the recovery JSONL format.
- Do not scan or migrate already persisted SQLite rows.
- Do not repair, truncate, delete, or rewrite invalid recovery input.
- Do not create a release tag or publish artifacts.

## Risks

- Accepting a blank name defers corruption detection until a later row read.
- Validation must precede writable SQLite initialization and runtime append.
- A trim-and-store implementation would silently alter valid public text.

## Validation

- Direct missing, empty, and whitespace-only recovery rule-name rejection after
  a valid prefix.
- Missing and existing database startup preservation.
- Runtime preflight rejection with full log and database byte preservation.
- Direct nonblank padded-name replay without normalization.
- Existing stored-row required-text compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Recovery `rule_name` must contain a non-whitespace character.
- Missing, empty, and whitespace-only values fail with
  `ErrRecoveryLogIntegrity` before SQLite modification.
- Startup rejection preserves the complete log and missing/existing database
  state.
- Runtime rejection appends nothing and preserves recovery-log and database
  bytes.
- Valid padded names replay without substitution or trimming.
- Stored-row blank-name errors retain their current behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires normalizing rule names, changing rule schema or
alert identity, rewriting the recovery format, stored-row migration, automatic
repair, operator data, tag creation, or publication authority.

## Delivery Evidence

- Recovery rule-name validation implemented with direct missing, empty,
  whitespace-only, startup-preservation, runtime-preservation, padded-name, and
  stored-row compatibility regressions.
- Twenty uncached focused alert-store race runs passed.
- `make test`, `make e2e-smoke`, `make docs-check`, explicit-Vault
  `make knowledge-check`, task-state JSON parsing, `git diff --check`, and the
  sensitive-information review passed.
- Pending commit, push, fetched-remote verification, post-fetch knowledge
  validation, and exact-range Vault synchronization.
