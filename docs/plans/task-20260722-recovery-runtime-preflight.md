# Task Plan: R90-19 preflight recovery logs before runtime append

## Metadata

- Timestamp: 2026-07-22T01:57:23-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `53c90ef4cefff4d22391196b6c8be61e3c8cc058`

## Goal

Reject an already malformed or semantically invalid recovery log before a
normal runtime alert write appends new durable records or modifies SQLite.

## Scope

- Validate the complete existing recovery log inside the serialized write
  critical section before appending the current normalized batch.
- Preserve rejected recovery-log and database bytes for structural, semantic,
  and normalized-invariant failures.
- Retain valid pending-log persistence, truncation, aggregation, and
  idempotency behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not add cross-process recovery-log locking or filesystem watch behavior.
- Do not change the JSONL format or add automatic repair/partial acceptance.
- Do not redesign SQLite transactions or daily-shard recovery atomicity.
- Do not create a release tag or publish artifacts.

## Risks

- Reading before append can accidentally discard valid pending records.
- Moving the existing read can weaken the append-to-persistence sequence.
- External writers remain outside the process-local `writeMu` boundary.

## Validation

- Direct malformed, truncated, semantic-invalid, and normalized-invariant
  runtime rejection cases with full-log and database byte preservation.
- Valid pending-log plus current-batch persistence compatibility.
- Repeated focused alert-store race tests, full native tests, E2E smoke,
  documentation, knowledge, JSON, diff, and sensitive-information checks.

## Acceptance Criteria

- `WriteBatch` reports `ErrRecoveryLogIntegrity` before appending to an invalid
  existing recovery log.
- Rejected input and the SQLite database remain byte-for-byte unchanged.
- Valid pending records and the current batch persist exactly once before the
  recovery log is truncated.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires cross-process locking, a recovery-format
change, automatic repair, operator data, or tag/publication authority.

## Local Validation

- Malformed, truncated, semantic-invalid, and normalized-invariant recovery
  logs each fail runtime `WriteBatch` before append.
- Every direct rejection preserves the complete log and SQLite bytes; a
  separate read-only handle confirms that no alert row was persisted.
- Valid pending recovery records still persist with the current batch and the
  log truncates only after successful persistence.
- Twenty uncached focused race runs, the full native C/Go race suite, E2E
  smoke, documentation, JSON, and diff checks pass.
- Feature commit: `9c93c8f82dfad07e17fcf57e4ba0818136b02710`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `53c90ef4cefff4d22391196b6c8be61e3c8cc058..9c93c8f82dfad07e17fcf57e4ba0818136b02710`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
