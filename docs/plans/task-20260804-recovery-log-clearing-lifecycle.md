# Task Plan: R90-68 recovery-log clearing lifecycle faults

## Metadata

- Timestamp: 2026-08-04T01:23:22-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `cac3178512a84356364f82261f2b7dffdfdf8e58`

## Goal

Make post-commit recovery-log clearing durable and directly injectable, then
prove open/truncate, sync, and close failures cannot lose an alert or inflate
an aggregate and that one explicit operator recovery returns every retained-log
or already-cleared outcome to healthy state.

## Scope

- Add a file `Sync` between recovery-log truncation and close.
- Add one store-local, unexported clear-file opener that defaults to the current
  `os.OpenFile` flags and exposes only `Sync` and `Close`.
- Add direct open/truncate, sync, and close failure regressions after an alert
  commit is independently observable.
- Exercise every failure phase for an ordinary primary database and a
  pre-existing historical daily shard under a directory requiring URL encoding.
- Open and reuse one independent read-only SQLite observer before each writer
  to prove the commit boundary and post-recovery cardinality.
- Classify exact retained versus already-cleared log bytes, phase-specific
  sticky emergency health, and one explicit `Recover` result.
- Update current storage/testing guidance, changelog, roadmap, and task state.

## Non-Goals

- Do not change append open/write/sync/close behavior, the recovery format,
  SQLite schema, public API, event identity, aggregation, or retry ownership.
- Do not roll back a committed SQLite transaction or delete an uncommitted
  recovery log.
- Do not add automatic retry, background cleanup, filesystem mounts, privileged
  fault infrastructure, dependencies, or cross-process ownership.
- Do not access private/external data or perform tag, release, image, registry,
  or workflow mutations.
- Do not start the publication-blocked R90-59 or invent a later increment.

## Risks

- `O_TRUNC` clears bytes during open, so sync and close errors can report
  failure after the file is observably empty. Tests must not call that retained
  evidence or infer crash durability from an in-process read.
- A retained complete log after a committed transaction will replay the same
  event; event-ledger idempotency must prove no aggregate inflation.
- An empty log after a committed transaction requires a writable probe rather
  than replay. The probe must roll back and preserve the committed alert.
- Daily-shard proof can accidentally test only the current shard or bypass URL
  encoding. The target must be a pre-existing non-current shard under a path
  containing spaces.
- A package-global seam would race unrelated tests; injection must remain
  owned by one Store instance.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Clearing has an explicit durability boundary | Production clear path calls `Sync` after successful `O_TRUNC` and before close; existing success paths remain compatible |
| Open/truncate failure retains evidence after commit | Injected `EIO` opener failure; independently observed event/aggregate count one; exact canonical one-record log retained; emergency health names `truncate` |
| Sync failure classifies already-cleared state | Real `O_TRUNC`, injected sync `EIO`, exact empty log, committed event/aggregate count one, emergency health names `sync` |
| Close failure classifies already-cleared state | Real truncate and sync followed by real close plus injected `EIO`; exact empty log, committed event/aggregate count one, emergency health names `close` |
| One operator retry is lossless and idempotent | With the seam removed, one `Recover` replays retained evidence or probes an empty log, preserves one event and aggregate count one, leaves an empty log, and returns healthy |
| Primary and encoded daily-shard paths are covered | Every phase runs against an ordinary primary database and a pre-existing non-current shard beneath a directory containing spaces, with observers opened before writes |
| Existing behavior remains compatible | Focused tests, twenty uncached focused race runs, full alert/native tests, E2E, docs, and knowledge checks pass |

## Trigger Audit

- Fetched `origin/main` and verified a clean local/remote equality at
  `cac3178512a84356364f82261f2b7dffdfdf8e58`.
- Verified the exact R90-67 feature and docs-only closure commits, completed
  task state, both Vault notes, full index, MOC links, and current stable
  SQLite/testing/MOC authority.
- Reviewed 145 commits from Jul 14 through Aug 4 across governance/release,
  storage-contract/recovery, and fault/fuzz-hardening phases. No missing recent
  delivery record, stale authority, or unresolved validation deviation changes
  the dependency order.
- Parsed all 82 task-state JSON files and verified all 71 roadmap rows match
  exactly one Definition.
- Audited the unfinished queue: R90-68 is the sole dependency-ready increment;
  R90-59 remains blocked on exact publication authority.
- Inspected `WriteBatch`, `ReplayRecoveryLog`, `Recover`,
  `truncateRecoveryLog`, health classification, daily-shard path selection,
  and existing committed-prefix/primary interruption tests. The clear path
  currently opens with `O_TRUNC` and closes without an explicit `Sync`, and its
  three post-commit failure phases lack direct deterministic regressions.

## Validation

- Preflight repository-required Go, GCC, Make, Bash, and Python tools before
  the complete fail-fast chain.
- Run all new clear-lifecycle regressions under `go test -race -count=1`.
- Run the exact focused set twenty times uncached under the race detector.
- Run the complete `internal/alert` package under `go test -race -count=1`.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast in the safe ordinary-build order.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and a scoped
  sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Split the existing append file contract into a shared `Sync`/`Close`
  durability boundary plus append-only `Write`, then added one unexported
  per-Store clear opener with the production `os.OpenFile` path as its nil
  default.
- `truncateRecoveryLog` now calls file `Sync` after successful `O_TRUNC` and
  before close; sync failure attempts close and retains the sync diagnostic.
- Six table-driven direct cases inject open/truncate, sync, and close `EIO`
  after primary and pre-existing non-current daily-shard commits. The daily
  target directory contains spaces and is reached through the URL-safe SQLite
  path.
- Every case opens one independent observer before the writer, proves one event
  and one aggregate with count one at the failure boundary, verifies exact
  retained or empty log bytes plus phase-specific emergency, and proves one
  explicit `Recover` returns healthy without inflation.
- The first focused uncached race execution and the existing replay/probe
  compatibility regression passed.

## Validated Evidence

- Repository-required Go 1.25.12, GCC 13.3.0, GNU Make 4.3, Bash 5.2.21,
  and Python 3.12.3 were available before the complete validation chain.
- The exact six-case regression passed twenty final uncached race repetitions.
  Every iteration used `-count=1`.
- The complete `internal/alert` package passed under uncached race in 61.850
  seconds. `make test` passed ordinary C tests and every Go package under
  uncached race.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded. `make docs-check` and the 33-test knowledge gate passed.
- All 83 task-state JSON files parse and all 71 roadmap rows match exactly one
  Definition. Formatting, whitespace, intended-path, and scoped
  sensitive-information reviews pass.
- Every acceptance item reaches its direct boundary in both modes: open failure
  occurs after the observer sees the commit but before `O_TRUNC`; sync failure
  occurs after real truncation; close failure occurs after real sync and close;
  exact bytes distinguish retained from already-cleared state; and one
  `Recover` preserves one event, one row, and aggregate count one.

## Authority Boundaries

This trigger authorizes only R90-68's store-local clear fault seam, explicit
file sync, direct primary and encoded daily-shard clear lifecycle regressions,
the smallest correction required by those tests, supporting documentation,
repository validation, commit/push, and local Vault synchronization. It does
not authorize uncommitted-log deletion, automatic retry/cleanup, private input,
publication, tags, releases, images, registries, or workflow dispatch.

## Stop Conditions

Stop if completion requires deleting an uncommitted log, rolling back a
committed SQLite transaction, weakening sticky emergency semantics, privileged
filesystem infrastructure, a package-global mutable seam, a format/schema
migration, publication authority, or an ambiguous result that remains after
focused uncached review.
