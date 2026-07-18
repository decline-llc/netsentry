# Task Plan: R90-10 preserve corrupt historical daily shards on write

## Metadata

- Timestamp: 2026-07-18T00:31:14-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `14b31698cf42209796f9e9334b983722288a98f4`

## Goal

Reject writes to an existing corrupt or truncated non-current daily SQLite
shard through the same read-only integrity preflight used at primary-database
startup, before shard journal or schema initialization can modify the file.

## Scope

- Inspect an existing non-current shard before opening its writable handle.
- Reuse the R90-09 read-only `PRAGMA quick_check` and stable
  `ErrDatabaseIntegrity` contract.
- Add separate direct write regressions for arbitrary corrupt bytes and a
  truncated formerly valid shard, including byte-for-byte preservation.
- Reconcile storage/testing documentation, changelog, roadmap, and task state.

## Non-Goals

- Do not repair, delete, rename, quarantine, copy, or overwrite a shard.
- Do not redesign shard rotation, recovery-log replay, querying, or retention.
- Do not add eager validation of every historical shard at process startup.
- Do not create a release tag or publish artifacts.

## Risks

- Writable shard initialization can create a journal or alter corrupt input
  before returning an error.
- An empty or missing historical shard must remain a valid first-write target.
- A regression that checks only the returned error can miss file mutation.

## Validation

- Direct corrupt-byte and truncated-valid historical-shard write rejection
  tests with independent byte comparisons.
- Focused alert-store tests under the race detector.
- Full native C/Go race suite, E2E smoke, documentation and knowledge checks,
  JSON parsing, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- An existing non-empty non-current shard receives a read-only integrity check
  before writable open and journal/schema initialization.
- Corrupt arbitrary bytes and a truncated valid shard both reject the write
  with `ErrDatabaseIntegrity` and the preservation message.
- Each rejected shard remains byte-for-byte identical to its pre-write state.
- Missing, empty, and healthy non-current shards continue to accept writes.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires automatic repair/deletion, operator data, a
broader cross-shard storage redesign, or tag/publication authority.

## Completion

- Feature commit: `d08702d2c3a1e425c27b4bba5238039915603c97`
- Fetched `origin/main`: verified at the feature commit
- Focused race validation: 20 consecutive passes
- Full native C/Go race suite and `make e2e-smoke`: passed
- Documentation, knowledge, JSON, diff, and sensitive-information checks:
  passed
- Vault range:
  `14b31698cf42209796f9e9334b983722288a98f4..d08702d2c3a1e425c27b4bba5238039915603c97`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
