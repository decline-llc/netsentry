# Task Plan: R90-12 preserve malformed recovery logs during startup replay

## Metadata

- Timestamp: 2026-07-19T01:19:23-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `92e58e7df1b0cbbc533bc945dcc7c0c73d7e1c54`

## Goal

Reject malformed or truncated JSONL recovery input before startup replay can
persist a partial prefix or truncate the operator's only recovery artifact.

## Scope

- Give malformed recovery-log input a stable integrity error.
- Treat a non-empty final JSONL record without its terminating newline as
  truncated, even when its JSON payload is otherwise valid.
- Add direct malformed-record and truncated-record startup regressions with
  byte-for-byte preservation assertions.
- Prove valid logs truncate only after successful SQLite persistence and remain
  intact when persistence fails.
- Reconcile storage/testing documentation, changelog, roadmap, and task state.

## Non-Goals

- Do not repair, rewrite, quarantine, or partially accept malformed logs.
- Do not change the recovery JSON schema, event-id idempotency, or daily-shard
  routing.
- Do not redesign SQLite recovery or emergency-mode behavior.
- Do not create a release tag or publish artifacts.

## Risks

- Accepting an unterminated final record can mistake an interrupted append for
  complete durable input.
- Replaying a valid prefix before discovering a later malformed record can
  create ambiguous partial recovery.
- Truncating before persistence succeeds can discard the only recoverable copy.

## Validation

- Direct malformed and unterminated recovery-log startup failures with
  independent byte comparisons.
- Direct valid replay success and persistence-failure preservation coverage.
- Repeated focused alert-store tests under the race detector.
- Full native C/Go race suite, E2E smoke, documentation and knowledge checks,
  JSON parsing, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- Malformed JSON and a missing final JSONL newline return a stable recovery-log
  integrity error during startup.
- Rejected recovery files remain byte-for-byte identical to their pre-startup
  state, and no valid prefix is partially persisted.
- Valid newline-terminated logs replay idempotently and truncate only after
  successful persistence.
- A failed persistence attempt leaves the valid recovery log unchanged.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires automatic log repair, partial-record
acceptance, operator data, a replay-format redesign, or tag/publication
authority.
