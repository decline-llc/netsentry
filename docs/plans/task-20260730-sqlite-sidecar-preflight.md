# Task Plan: R90-56 preserve SQLite sidecars during read-only preflight

## Metadata

- Timestamp: 2026-07-30T10:30:01-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `cac88a4320dc820d9def98c8f5af775a0af5dfa2`

## Goal

Make every NetSentry-owned read-only SQLite handle preserve an existing
database, WAL, and SHM byte-for-byte while retaining active-WAL visibility and
clear rejection of deterministic corrupt or inconsistent sidecar fixtures.

## Scope

- Resolve an existing database symlink before inspecting whether either SQLite
  sidecar exists and constructing the read-only URI, matching the Unix VFS
  sidecar location.
- Keep the current URL-safe `mode=ro` URI for databases without sidecars.
- Add SQLite's `readonly_shm=1` URI control whenever a WAL or SHM sidecar is
  present so the default Unix VFS cannot open the SHM read/write, create it,
  truncate it, rebuild it in place, or update reader marks.
- Reuse that helper for primary startup, non-current historical writes, and
  historical query/count handles.
- Build deterministic WAL and SHM fault fixtures only from temporary,
  independently owned databases.
- Use a separate helper process for active-WAL SHM fixtures so the read-only
  connection cannot reuse an in-process read/write SHM node.
- Corrupt structurally recognized WAL/SHM metadata while retaining valid
  checksums where necessary, then snapshot the database and both sidecars
  after fault injection and compare every byte after rejection.
- Preserve healthy active-WAL query/count behavior and clean databases without
  sidecars.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not repair, checkpoint, delete, recreate, quarantine, or rewrite a
  database or either sidecar.
- Do not introduce a custom SQLite VFS or a NetSentry parser for every WAL or
  wal-index structure.
- Do not promise an atomic snapshot against an unrelated process that creates
  a sidecar between filesystem inspection and SQLite connection establishment.
- Do not reject every stale SHM that SQLite can safely reconstruct in private
  heap memory; this increment closes the mutation boundary and verifies
  deterministic rejection while an active owner prevents reconstruction.
- Do not change schemas, alert rows, query semantics, retention, recovery-log
  behavior, journal defaults, or public APIs.
- Do not start R90-57 or R90-58, create a release tag, publish artifacts, or
  change release authority.

## Risks

- SQLite `mode=ro` alone still opens a present SHM read/write and may update
  read marks or rebuild a malformed wal-index.
- Applying `readonly_shm=1` when no SHM exists can make an otherwise healthy
  clean database fail to open, so the URI option must be conditional.
- An in-process fixture can reuse the writer's read/write SHM node and conceal
  whether the read-only URI actually prevents writes.
- SQLite ignores many malformed WAL prefixes or tails as non-authoritative;
  fault metadata must remain structurally recognizable so rejection is
  deterministic.
- A helper process can checkpoint or remove sidecars during graceful close;
  fixture teardown must occur only after preservation assertions.
- Filesystem presence checks and SQLite's lazy connection establishment have a
  bounded time-of-check/time-of-use window that this increment does not claim
  to eliminate.
- Inspecting sidecars beside an unresolved database symlink misses the Unix
  VFS target files and can silently restore writable SHM access.

## Validation

- Direct URI regression verifies URL encoding and conditional
  `readonly_shm=1` selection.
- A no-sidecar database still passes primary read-only preflight.
- A detached active-WAL database with a supported healthy WAL remains readable
  and leaves the database, WAL, and SHM unchanged.
- A WAL header with an unsupported format version and a recomputed valid
  header checksum fails primary startup with `ErrDatabaseIntegrity`; separate
  file reads prove all three files are unchanged.
- An independently owned active-WAL SHM with matching duplicate headers,
  recomputed valid header checksums, and an unsupported wal-index version fails
  the encoded-path historical read/write boundary clearly; separate file reads
  prove all three files are unchanged.
- Historical query and count retain healthy active-WAL visibility through
  separate read-only handles.
- A database-file symlink resolves to its active-WAL target before sidecar
  selection; validation and a second read-only query retain the WAL-only row
  without changing the target database, WAL, or SHM.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, formatting, diff, staged-diff, and sensitive-information
  checks.

## Acceptance Criteria

- Read-only handles never request writable SHM access when a WAL or SHM exists.
- Database symlinks cannot redirect sidecar inspection away from the Unix VFS
  target used by the read-only handle.
- Deterministic corrupt WAL and active-owner inconsistent SHM fixtures fail
  before NetSentry obtains any writable database handle.
- Rejected primary and encoded-path historical fixtures preserve the database,
  WAL, and SHM byte-for-byte.
- Healthy active-WAL historical query/count results remain visible and
  sidecar bytes remain unchanged through the separate read-only handles.
- Healthy databases without sidecars retain their current startup behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the single local Vault.

## Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| No writable SHM request | URL-safe read-only DSN conditionally contains `readonly_shm=1` whenever either sidecar exists |
| Symlink-safe sidecar lookup | Unix database-file symlink resolves to its active-WAL target before URI construction; validation and query preserve all target bytes |
| Corrupt WAL rejection | Checksummed unsupported WAL-version primary fixture returns `ErrDatabaseIntegrity` |
| Inconsistent SHM rejection | Checksummed unsupported wal-index version under an independent active owner fails encoded-path historical access |
| Three-file preservation | Independent pre/post `os.ReadFile` comparisons for `.db`, `-wal`, and `-shm` after each rejected operation |
| Active-WAL compatibility | Independent writer plus historical query/count regression sees WAL-only rows with unchanged bytes |
| Clean compatibility | Existing nonempty database without sidecars passes startup preflight |
| Delivery | Fetched remote equality, post-fetch knowledge gate, exact-range Vault note/index/MOC, and stable storage note |

## Plan Audit

- The trigger fetched `origin/main` and verified clean local and remote equality
  at `cac88a4320dc820d9def98c8f5af775a0af5dfa2`.
- The prior R90-55 feature and docs-only closure commits, task state, both exact
  Vault notes, full index, MOC links, idempotent replay, and stable storage note
  are complete.
- The recent-history audit reconciled all 247 July commits. The six commits
  after the dated R90-53 audit are exactly the feature and closure pairs for
  R90-53, R90-54, and R90-55; no unrecorded delivery deviation remains.
- All 62 roadmap rows have matching Definitions. Every unfinished increment
  retains a dependency, delivery window, risk, acceptance criteria, required
  validation, and stop condition.
- R90-56 is the only highest-priority dependency-ready increment. R90-57
  remains blocked on a product decision, R90-58 remains planned behind this
  increment, and R90-59 remains blocked on explicit publication authorization.
- Direct modernc SQLite experiments invalidated the assumption that `mode=ro`
  preserves SHM bytes and confirmed the conditional `readonly_shm=1`,
  cross-process fixture, and clean-no-sidecar boundaries recorded above.

## Execution Deviation

- **Observed:** The interrupted execution briefly left a competing
  team-generated plan/state pair plus an unvalidated WAL/SHM parser draft and
  tests that referenced missing helper-process functions.
- **Impact:** None of that draft was accepted as implementation or validation
  evidence. It conflicted with this persisted plan's bounded no-parser scope
  and could have rejected recoverable SQLite states.
- **Resolution:** The duplicate records and parser draft were removed. The
  implementation now contains only conditional `readonly_shm=1` URI selection,
  and the deterministic WAL/SHM faults use complete independent-process test
  helpers.
- **Observed:** Read-only code review then found that inspecting sidecars beside
  an unresolved database symlink could miss the files SQLite opens beside the
  real target. The same review found that fixture readiness and shutdown waits
  were initially unbounded.
- **Impact:** The first focused/full test results were invalidated and are not
  accepted as final evidence.
- **Resolution:** The shared URI helper now resolves database symlinks before
  sidecar inspection and URI construction; a Unix active-WAL symlink regression
  proves target-byte preservation and WAL visibility. Helper processes use a
  15-second context deadline. The complete validation matrix was rerun only
  after these corrections.

## Validation Evidence

- The final seven-test sidecar matrix passed 20/20 uncached race runs after the
  detached-WAL, bounded-helper, and database-symlink corrections.
- `make test` passed the capture tests and every Go package with race detection;
  the complete alert package passed in 38.680 seconds.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded.
- `make docs-check`, `make config-check`, and `make knowledge-check` passed;
  the knowledge gate ran 33 tests.
- Every task-state JSON document parsed successfully. `gofmt -l` and
  `git diff --check` produced no findings.
- Independent read-only code review closed the healthy-detached-WAL,
  helper-timeout, and symlink findings, then reported no remaining correctness
  or scope blocker after three additional uncached race runs of all seven new
  sidecar tests.

Every behavior and preservation criterion maps to passing direct evidence.
Remote equality and exact-range Vault evidence were then verified after the
feature commit; this docs-only record closes the increment without starting
later work.

## Delivery Evidence

- Feature commit:
  `e97e7ceeaa1acd877e773278f6992add7baa22a4`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `cac88a4320dc820d9def98c8f5af775a0af5dfa2..e97e7ceeaa1acd877e773278f6992add7baa22a4`
- Vault feature note, full commit index, MOC link, idempotent identical-range
  replay, and the stable SQLite storage note at feature version
  `e97e7ceeaa`: verified.
- R90-58 is ready but was not started. R90-57 remains blocked on its product
  decision, and R90-59 remains blocked on explicit publication authorization.

## Final Plan Audit

- All R90-56 acceptance criteria have direct regression, full-suite, E2E,
  documentation, review, remote, and local Vault evidence.
- Every validation result affected by a code or fixture correction was
  explicitly invalidated and rerun on the final implementation.
- The delivered file set remains bounded to the selected R90-56 increment.
  No R90-57/R90-58 implementation, release tag, artifact publication, or
  release-authority change occurred.
- All 62 roadmap rows still have matching Definitions; the forward queue now
  exposes R90-58 as the only dependency-ready increment while preserving the
  explicit R90-57 and R90-59 blocks.

## Stop Conditions

Stop if deterministic rejection would modify a fixture after the fault
snapshot, require operator data or privileged storage faults, perform automatic
repair, require a custom VFS or broad WAL parser, change healthy active-WAL
visibility, depend on an in-process read/write SHM node, expand into R90-57 or
R90-58, or require tagging or publication authority.
