# Task Plan: R90-67 recovery-log append lifecycle faults

## Metadata

- Timestamp: 2026-08-04T00:44:13-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `2f62acf9025969a50dd0295f3881ce7cd2784ec6`

## Goal

Make recovery-log append open, short-write, sync, and close failures directly
injectable, then prove each failure returns before SQLite mutation, preserves
the exact pre-existing valid prefix, reports its lifecycle phase, and leaves
complete or incomplete appended evidence fail-closed for explicit recovery.

## Scope

- Add one store-local, unexported append-file seam that defaults to the current
  `os.OpenFile` behavior and exposes only `Write`, `Sync`, and `Close`.
- Treat a nil-error short write as `io.ErrShortWrite` instead of accepting a
  truncated recovery record as a successful append.
- Add direct primary-store regressions for append open, short-write, sync, and
  close failures after seeding one canonical pending record.
- Use one independently opened read-only SQLite observer per case to prove no
  event or aggregate mutation at the rejected boundary.
- Prove complete appended records replay once after restart, while an
  incomplete suffix fails startup preflight and remains byte-for-byte intact.
- Update current storage/testing guidance, changelog, roadmap, and task state.

## Non-Goals

- Do not change the recovery JSON format, SQLite schema, public API, retry
  policy, emergency classification, cross-process ownership, or append order.
- Do not automatically truncate, repair, rename, or delete a partial recovery
  record or any other failed append evidence.
- Do not inject post-commit recovery-log truncate/sync/close faults; those
  remain R90-68.
- Do not add privileged filesystem simulation, mounts, dependencies, private
  input, tags, releases, images, registry mutations, or workflow dispatch.
- Do not start R90-68 or the publication-blocked R90-59.

## Risks

- A package-global fault seam would race unrelated tests; injection must be
  owned by one Store instance.
- A short write can append an incomplete JSONL suffix. Rolling it back would
  destroy fault evidence, so later reads must reject and preserve the file.
- A sync or close error leaves durability uncertain even when a complete line
  is visible. The complete record must remain replay-safe without claiming the
  failed operation was durable.
- A test that opens a fresh observer only after failure can miss the promised
  boundary or perturb SQLite scheduling; each observer must predate the write.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Open failure precedes log and SQLite mutation | Injected open sentinel; exact seeded prefix; pre-opened observer reports zero events and alerts; degraded health names `open` |
| Short write is rejected and evidence is retained | Real partial write through the narrow file wrapper; returned `io.ErrShortWrite`; exact prefix plus expected incomplete suffix; startup preflight returns `ErrRecoveryLogIntegrity` without changing log or SQLite |
| Sync failure precedes SQLite mutation | Full appended canonical line followed by injected sync sentinel; exact log bytes; zero database rows; degraded health names `sync` |
| Close failure precedes SQLite mutation | Successful write/sync and real close followed by injected close sentinel; exact log bytes; zero database rows; degraded health names `close` |
| Complete append evidence replays once | Restart after sync and close failures persists both distinct pending events and aggregates once, clears the log only after success, and remains healthy |
| Existing append behavior remains compatible | Direct success regression plus focused tests, twenty uncached focused race runs, full alert/native tests, E2E, docs, and knowledge checks pass |

## Trigger Audit

- Fetched `origin/main` and verified a clean local/remote equality at
  `2f62acf9025969a50dd0295f3881ce7cd2784ec6`.
- Verified the exact R90-66 feature and docs-only closure commits, completed
  task state, both Vault notes, full index, MOC links, and current stable
  SQLite/testing authority.
- Reviewed 143 commits from Jul 14 through Aug 3 across governance/release,
  storage-contract/recovery, and fault/fuzz-hardening phases. No missing recent
  delivery record, stale authority, or unresolved validation deviation changes
  the dependency order.
- Parsed all 81 task-state JSON files and verified all 71 roadmap rows match
  exactly one Definition.
- Audited the unfinished queue: R90-67 is the sole dependency-ready increment,
  R90-68 depends on it, and R90-59 remains blocked on exact publication
  authority.
- Inspected `WriteBatch`, `appendRecoveryLog`, restart replay, health
  classification, and existing primary interruption tests. The four append
  lifecycle phases have distinct production errors but no direct deterministic
  injection or preservation regressions.

## Validation

- Preflight repository-required Go, GCC, Make, Bash, and Python tools before
  the complete fail-fast chain.
- Run all new append-lifecycle regressions under `go test -race -count=1`.
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

- Added one unexported per-Store append opener returning the narrow
  `Write`/`Sync`/`Close` contract; a nil seam keeps the production
  `os.OpenFile` behavior.
- `appendRecoveryLog` now rejects nil-error partial writes with
  `io.ErrShortWrite` and preserves every byte already appended.
- One table-driven direct regression injects open, real partial-write, sync,
  and close failures. Every case observes zero SQLite rows through a read-only
  handle opened before the writer and verifies degraded phase-specific health.
- Open failure preserves the exact prior log and succeeds after the seam is
  removed. Sync and close failures preserve complete appended lines that
  restart replay once. Short-write startup fails with
  `ErrRecoveryLogIntegrity` and preserves the exact incomplete suffix.
- The first focused uncached race execution passed all four direct cases.

## Validated Evidence

- Repository-required Go 1.25.12, GCC 13.3.0, GNU Make 4.3, Bash 5.2.21,
  and Python 3.12.3 were available before the complete validation chain.
- The exact four-case regression passed twenty final uncached race
  repetitions. Every iteration used `-count=1`.
- The complete `internal/alert` package passed under uncached race in 64.759
  seconds. `make test` passed ordinary C tests and every Go package under
  uncached race.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded. `make docs-check` and the 33-test knowledge gate passed.
- All 82 task-state JSON files parse and all 71 roadmap rows match exactly one
  Definition. Formatting, whitespace, intended-path, and scoped
  sensitive-information reviews pass.
- Every acceptance item reaches its direct boundary: open is rejected before
  file mutation; short-write leaves the expected incomplete bytes; sync and
  close leave exact complete bytes; one pre-opened observer sees zero SQLite
  rows; and only complete records replay once after restart.

## Authority Boundaries

This trigger authorizes only R90-67's store-local append fault seam,
short-write rejection, direct append lifecycle regressions, the smallest
correction required by those tests, supporting documentation, repository
validation, commit/push, and local Vault synchronization. It does not authorize
R90-68, evidence repair, private input, publication, tags, releases, images,
registries, or workflow dispatch.

## Stop Conditions

Stop if deterministic coverage requires privileged mounts, destructive host
faults, a package-global mutable seam, automatic failed-evidence truncation, a
recovery-format/schema change, cross-process write ownership, publication
authority, or an ambiguous result that remains after focused uncached review.
