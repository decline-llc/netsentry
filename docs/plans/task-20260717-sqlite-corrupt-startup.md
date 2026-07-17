# Task Plan: R90-09 fail closed on corrupt SQLite startup

## Metadata

- Timestamp: 2026-07-17T01:22:06-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `c049470a985300da1a4ee18c02043cff4c669e0e`

## Goal

Reject an existing corrupt or truncated SQLite alerts database through an
explicit read-only startup integrity preflight before journal-mode or schema
initialization can modify the file, and document a data-preserving operator
recovery path.

## Scope

- Detect whether the primary database existed and was non-empty before
  `sql.Open`.
- Run SQLite `PRAGMA quick_check` before initialization for that existing file.
- Return a stable integrity sentinel and clear non-modification message when
  the check errors or reports corruption.
- Add separate direct regressions for arbitrary corrupt bytes and a truncated
  formerly valid database, including byte-for-byte preservation assertions.
- Reconcile storage/testing documentation, changelog, roadmap, and task state.

## Non-Goals

- Do not repair, delete, rename, quarantine, copy, or overwrite an operator
  database automatically.
- Do not validate every historical daily shard or redesign runtime shard
  handling in this increment.
- Do not change the SQLite schema, recovery-log format, or emergency-mode
  semantics.
- Do not create a release tag or publish artifacts.

## Risks

- A validation query that runs after mutable initialization can alter evidence
  before corruption is reported.
- Treating an empty new database as corruption can break first startup.
- A weak regression can assert an error while missing file mutation.

## Validation

- Direct corrupt-byte and truncated-database startup rejection tests.
- Repeated focused alert-store tests under the race detector.
- Full native C/Go race suite, E2E smoke, documentation and knowledge checks,
  JSON parsing, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- Existing non-empty databases receive a read-only integrity check before
  journal/schema initialization.
- Corrupt arbitrary bytes and a truncated valid database both return the
  stable integrity error with a clear statement that the file was not modified.
- Each rejected database remains byte-for-byte identical to its pre-open state.
- New and existing healthy databases continue to open normally.
- Recovery guidance tells operators to stop, preserve/copy the database and
  sidecars, inspect a copy with SQLite tooling, and point NetSentry at a new
  path rather than deleting or repairing the original automatically.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires automatic repair/deletion, operator data,
privileged external services, a broader storage redesign, or tag/publication
authority.
