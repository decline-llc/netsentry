# Task Plan: R90-20 bound recovery-record encoding and replay

## Metadata

- Timestamp: 2026-07-22T02:14:56-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `7f5ad8ecd1f2850c7641be4132aa44d864628097`

## Goal

Use one explicit 4 MiB durable-record limit for recovery writing and reading so
valid large writer output remains replayable and oversized output cannot
partially mutate the recovery log.

## Scope

- Encode and size-check the complete batch before opening the recovery log.
- Accept individual valid JSONL records through 4 MiB during startup and
  runtime preflight.
- Reject above-limit writer records before append with a clear durable-size
  error and preserve the existing log and SQLite database.
- Retain structural, semantic, normalized-invariant, persistence, truncation,
  and idempotency behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not accept unbounded recovery records or change the JSONL format.
- Do not add cross-process locking, compression, or automatic repair.
- Do not change rule, API-body, UDS-frame, or SQLite row limits.
- Do not create a release tag or publish artifacts.

## Risks

- Scanner capacity must include the line delimiter without weakening the
  payload limit.
- Streaming size checks could append a valid prefix before a later oversized
  batch record fails.
- Large-record tests can hide cache reuse or create excessive test memory.

## Validation

- Direct above-64-KiB runtime persistence and startup replay regressions.
- Exact 4-MiB record acceptance and above-limit rejection.
- Above-limit direct append and `WriteBatch` preservation cases, including a
  valid first record followed by an oversized record.
- Existing malformed/truncated and valid pending-log regressions.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Writer-generated records larger than 64 KiB and no larger than 4 MiB persist
  and replay successfully.
- Every record in a batch is encoded and checked before the file is opened.
- An above-limit record returns the durable-size sentinel without changing the
  existing recovery log or SQLite database.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires an on-disk format migration, unbounded record
acceptance, automatic repair, operator data, or tag/publication authority.

## Local Validation

- A 70 KiB writer-generated record persists through runtime `WriteBatch` and
  replays at startup without triggering the former scanner ceiling.
- A record whose encoded JSON is exactly 4 MiB appends and reads successfully.
- A 4 MiB plus one byte record fails direct append and runtime `WriteBatch`
  with `ErrRecoveryRecordTooLarge`; a preceding valid batch record is not
  appended, and existing log and SQLite bytes remain unchanged.
- Existing malformed and truncated input preservation remains compatible.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, JSON, and diff checks pass.
- Tag/publication actions: none.
