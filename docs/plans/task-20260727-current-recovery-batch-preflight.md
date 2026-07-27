# Task Plan: R90-45 preflight the current recovery batch

## Metadata

- Timestamp: 2026-07-27T05:06:06-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `11371d520abd6004d6b1991eef5544dcffc48c8f`

## Goal

Validate every newly normalized alert against the complete durable recovery
contract before any current-batch JSONL append or SQLite write.

## Scope

- Reuse the complete recovery-record validator for the current normalized
  batch after the existing recovery log passes preflight and before append.
- Validate the complete current batch before writing any record so a valid
  prefix cannot be partially appended.
- Reject reachable invalid current identity, required text, severity, MITRE,
  protocol, and IPv4 fields with a clear current-record number.
- Preserve an existing valid pending recovery log and the target SQLite
  database byte-for-byte when the current batch is rejected.
- Preserve the existing behavior that a valid pending log and valid current
  batch persist together and truncate only after SQLite success.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the recovery JSONL format or SQLite schema.
- Do not validate optional payload or matched-keyword text.
- Do not change receiver, rule-loader, query-filter, or aggregation semantics.
- Do not repair, delete, truncate, or normalize rejected operator input.
- Do not add cross-process recovery-log locking.
- Do not create a release tag or publish artifacts.

## Risks

- Validating only during encoding could partially append a multi-record batch
  before a later record fails.
- Marking invalid current input as a storage fault could incorrectly degrade
  an otherwise healthy store.
- Moving current validation ahead of existing-log preflight could obscure
  already-durable operator corruption.

## Validation

- Direct valid-prefix current-batch rejection for blank required rule, network,
  and protocol text; a caller-supplied event-identity mismatch; unsupported
  severity; an invalid MITRE tuple; a noncanonical protocol; and invalid source
  or destination IPv4.
- Every rejection reports current record two, preserves a pre-existing valid
  pending log byte-for-byte, leaves SQLite byte-for-byte unchanged, and
  persists no alert.
- Existing valid pending-log plus current-batch compatibility regression.
- Existing exhaustive startup/runtime recovery-contract regressions.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Every newly normalized alert passes `validateRecoveryAlert` before any
  current-batch record is appended.
- The entire current batch is validated before append; no valid prefix is
  written when a later record is invalid.
- Invalid current input does not mark a healthy store degraded.
- Existing valid pending records remain untouched when the current batch is
  rejected.
- Valid pending and current records retain existing persistence, aggregation,
  idempotency, and successful-truncation behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires a recovery-format or SQLite-schema change,
cross-process locking, changing public alert semantics, automatic repair or
deletion, operator data, tag creation, or publication authority.

## Delivery Evidence

- `WriteBatch` now validates the complete normalized current batch with
  `validateRecoveryAlert` after existing-log preflight and before append.
- Direct valid-prefix regressions cover every planned reachable required-text,
  event-identity, severity, MITRE, protocol, and IPv4 rejection category.
- Every direct rejection preserves the existing pending log and SQLite bytes,
  persists no alert, reports current record two, and leaves storage health
  `ok`.
- The existing valid pending/current persistence and successful-truncation
  regression remains compatible.
- Twenty uncached focused alert-store race runs, the complete native race
  suite, E2E smoke, documentation, config, knowledge, JSON, and diff checks
  passed.
- Commit, push, fetched-remote verification, post-fetch knowledge validation,
  and exact-range Vault synchronization remain pending.
