# Task Plan: R90-16 reject semantically invalid recovery records

## Metadata

- Timestamp: 2026-07-20T05:28:00-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `246ab7baf878d815608e407d148698c827c5a774`

## Goal

Reject syntactically valid JSONL records that do not satisfy the durable
normalized-alert contract before replay can persist any valid prefix.

## Scope

- Validate every decoded recovery record before returning the complete replay
  set.
- Require the identity, timestamps/window/count, and network tuple that the
  versioned recovery writer always emits.
- Reject JSON `null`, empty objects, missing durable identity, and missing
  network fields with `ErrRecoveryLogIntegrity`.
- Preserve the complete rejected log and prove no valid prefix is persisted.
- Retain valid replay and event-ID idempotency behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not repair, rewrite, quarantine, or partially accept invalid records.
- Do not require optional MITRE, payload, or matched-keyword metadata.
- Do not change alert normalization, the recovery JSON schema, SQLite layout,
  daily-shard routing, or event-ID semantics.
- Do not create a release tag or publish artifacts.

## Risks

- An incomplete validator can persist alerts with empty durable identities.
- An over-strict validator can reject legitimate current recovery records.
- Validating records during replay instead of during the full read can allow
  partial-prefix persistence.

## Validation

- Direct null, empty-object, missing-identity, and missing-network regressions
  with independent byte comparisons and no-prefix-persistence assertions.
- Valid replay and idempotency compatibility coverage.
- Repeated focused alert-store tests under the race detector.
- Full native C/Go race suite, E2E smoke, documentation and knowledge checks,
  JSON parsing, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- Every syntactically valid but semantically invalid record returns
  `ErrRecoveryLogIntegrity` with its record number and field condition.
- The complete rejected recovery log remains byte-for-byte unchanged, and no
  valid prefix is persisted.
- Records emitted by the current normalized recovery writer still replay and
  truncate only after successful persistence.
- Duplicate valid replay remains idempotent.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if the durable semantic contract is ambiguous, compatibility requires
accepting records the current writer cannot generate, automatic repair or
operator data is required, or work reaches tag/publication authority.

## Local Validation

- JSON `null`, an empty object, and every missing required identity,
  timestamp/window/count, and network field fail at record 2 through
  `ErrRecoveryLogIntegrity`.
- Every direct rejection preserves the complete recovery log byte-for-byte and
  leaves the valid prefix unpersisted.
- Twenty uncached focused race runs, the full alert-store race package, the
  full native C/Go race suite, E2E smoke, documentation, JSON, and diff checks
  passed.
- Valid replay, post-commit truncation, and duplicate event-ID idempotency
  remain covered and passing.
- Feature commit: `40b58c2c5160262efc42e3d8d7e5e588cd71fcc6`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `246ab7baf878d815608e407d148698c827c5a774..40b58c2c5160262efc42e3d8d7e5e588cd71fcc6`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
