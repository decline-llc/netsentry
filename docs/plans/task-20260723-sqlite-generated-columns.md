# Task Plan: R90-24 reject write-affecting SQLite generated columns

## Metadata

- Timestamp: 2026-07-23T04:50:34-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `fd1bbb476f340a159729052f2675b201678f1c68`

## Goal

Reject existing schemas with virtual or stored generated columns on
NetSentry's write-critical tables before writable SQLite initialization.

## Scope

- Replace the incomplete `PRAGMA table_info` view with `PRAGMA table_xinfo`
  during required-column preflight.
- Reject virtual-generated and stored-generated column extensions on `alerts`
  or `alert_events` because arbitrary expressions enter NetSentry's fixed
  write path.
- Apply the same boundary to primary databases and historical shard writes.
- Keep ordinary nullable and non-NULL-defaulted columns compatible.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not parse, evaluate, classify, remove, or rewrite generated expressions.
- Do not reject ordinary compatible column extensions.
- Do not migrate or rewrite operator schemas.
- Do not change NetSentry's canonical schema or SQL write statements.
- Do not create a release tag or publish artifacts.

## Risks

- Continuing to use `table_info` silently omits generated columns.
- Interpreting arbitrary expressions outside SQLite can accept a write blocker
  or reject based on incomplete semantics.
- A primary-only fix can leave historical shard writes exposed.

## Validation

- Direct primary rejection for virtual and stored generated columns across
  `alerts` and `alert_events`, with named errors and byte preservation.
- Direct historical-shard generated-column rejection with byte preservation.
- Existing nullable and non-NULL-defaulted ordinary-column compatibility with
  successful writes.
- Repeated uncached focused alert-store race tests, full native tests, E2E
  smoke, documentation, knowledge, JSON, diff, and sensitive-data checks.

## Acceptance Criteria

- Every virtual or stored generated column on `alerts` or `alert_events`
  returns `ErrDatabaseIntegrity` during read-only schema preflight.
- Rejected primary and historical databases remain byte-for-byte unchanged.
- Ordinary nullable and non-NULL-defaulted column extensions remain compatible
  and do not block NetSentry writes.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if completion requires parsing or evaluating generated expressions,
schema migration, rewriting operator columns, operator data, or publication
authority.

## Local Validation

- Virtual and stored generated columns on both `alerts` and `alert_events`
  fail with `ErrDatabaseIntegrity` and byte preservation.
- A historical shard with a generated column fails before writable open and
  remains byte-for-byte unchanged.
- Ordinary nullable and non-NULL-defaulted columns reopen and accept complete
  alert/event writes.
- Twenty uncached focused generated-column race runs pass.
- The first full native run hit an unrelated receiver timing boundary; twenty
  focused receiver race reruns and the complete uncached native rerun pass
  without a receiver change.
- E2E smoke, documentation, knowledge, JSON, and diff checks pass.
- Feature commit: `4b342ae65b10279448b438e43b1947f1cfb282fc`
- Fetched `origin/main`: verified at the feature commit
- Post-fetch knowledge validation: passed
- Vault range:
  `fd1bbb476f340a159729052f2675b201678f1c68..4b342ae65b10279448b438e43b1947f1cfb282fc`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none.
