# Task Plan: R90-11 make historical daily-shard reads strictly read-only

## Metadata

- Timestamp: 2026-07-18T00:44:00-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `37cfec7976f11b3366f226ef5b7ff762db89d5f8`

## Goal

Open non-current daily SQLite shards through URL-safe read-only handles for
query and count operations, so malformed historical input cannot be modified
by a nominally read-only request.

## Scope

- Centralize construction of SQLite `mode=ro` handles already used by the
  integrity preflight.
- Route non-current daily-shard query and count opens through that helper.
- Add direct corrupt and truncated shard preservation regressions for both
  operations.
- Verify healthy cross-shard reads, including an active WAL-backed historical
  shard, retain their results.
- Reconcile storage/testing documentation, changelog, roadmap, and task state.

## Non-Goals

- Do not run an integrity preflight before every historical read.
- Do not change current-shard ownership, query/filter semantics, pagination,
  retention, or shard write behavior.
- Do not repair, delete, quarantine, snapshot, or replace malformed shards.
- Do not create a release tag or publish artifacts.

## Risks

- Incorrect SQLite URI encoding can fail for valid filesystem paths.
- Read-only mode can expose WAL/SHM access assumptions hidden by writable opens.
- Error-only regressions can miss mutation of malformed shard bytes.

## Validation

- Direct corrupt/truncated query and count failures with byte comparisons.
- Active WAL-backed and existing healthy cross-shard compatibility tests.
- Repeated focused alert-store tests under the race detector.
- Full native C/Go race suite, E2E smoke, documentation and knowledge checks,
  JSON parsing, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- Query and count use SQLite `mode=ro` for every non-current daily shard.
- Corrupt and truncated historical shards fail both operations without changing
  their bytes.
- Healthy cross-shard query/count results and active WAL visibility remain
  unchanged.
- The current shard continues to use its store-owned connection.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe read-only access requires repair, snapshots, operator data, a
broader storage/query redesign, or tag/publication authority.
