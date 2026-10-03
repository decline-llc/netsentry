# R90-161: make zero-value Stats alert observation safe

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`de68278a8886eaf07a44d576e5ffe98afb6a0d08`. R90-160 exact six-path feature/
three-path closure Git/Vault scope/note/index/MOC verified; generated identifiers
uniquely resolve through Git to full endpoints. 101-commit Sep 5–Oct 3 phase moves
from SLO tooling to core correctness; no new qualifying R90-75 evidence. R90-75
sole unfinished full independent asynchronous departmental contract and Oct 3–
Dec 31 horizon unchanged. Empty queue reconciled inside source-grounded zero-map
panic repair. Unique existing local Vault's 401 Markdown hashes/14 entire stable
notes backed up outside Git; no AGENTS. Owning engine pinned Go 1.26.8 preflighted.

## Scope, risk and authority

Seven paths: `engine/internal/stats/stats.go`, new
`engine/internal/stats/zero_alerts_test.go`, new
`engine/internal/pipeline/zero_stats_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-zero-stats-alerts.json`, rolling roadmap. Plan/state
persisted before source/other docs. Exported Stats' other atomic methods/Snapshot
work on a zero value, but ObserveAlerts writes a nil severity map at first non-nil
alert. Pipeline catches the panic after successful write and before completion.
Allocate severity map lazily at first non-nil alert while holding existing mu;
retain every other line. Risk low: enables zero-value observation, with first-use
allocation only. Preserve nil/empty/all-nil no-op behavior, per-entry counts,
blank-to-low/dynamic/repeated labels, snapshots/exposition, New's four stable
labels/start time and existing Worker write/export/error/completion gates.
Zero Stats keeps zero StartedAt and only observed labels; it gains no automatic
start time or constructor-equivalent zero labels. No live transactional snapshot,
race, allocation or throughput guarantee. All behavioral/race/CLI/full-suite/
scanner/knowledge/traffic/acceptance execution **not run; delegated by user**.
Compile/static evidence is no Worker/SQLite/filesystem/runtime/SLO pass.

## Acceptance mapped to evidence

1. Exact three-line runtime diff: nil map initialization inside existing locked
   non-nil entry loop; all other tracked engine source/New/render/Snapshot/Worker
   byte-identical. Nil/empty/all-nil bypass map initialization; no schema changes.
2. External direct public zero Stats and New table covers nil/empty/all-nil,
   single/mixed/default/all-severities/repeated-pointer/dynamic-label batches,
   two observations per case. Compare full snapshots to seeded unrelated
   counters/durations/queue/buckets/start time, exact total/severity map and
   Prometheus lines; preserve alert pointers/values and prove snapshot map copy
   isolation. Typed nil Stats remains safe with exact zero snapshot.
3. Four start-synchronized zero Stats writers each 100 mixed nil/high/blank/custom
   observations plus all-nil batches; join before exact aggregate/exposition and
   input-preservation assertions. Counts 1200, high/low/custom 400 each; StartedAt
   remains zero, unobserved stable labels not invented. No timing sleeps/race pass.
4. Public actual Worker.Run with zero Stats and actual SQLite Store writes two
   valid alerts in mixed nil batch, keeps packet unchanged, yields exact generated/
   severity/processed/completed/write/panic counters and expected healthy store
   rows/count/timestamps. Encoded temporary path; actual writer, small matcher
   fixture; no private state. No test execution or SQLite durability claim.
5. Public Worker no-alert and injected writer-failure controls keep no write or
   failed-write generated/completed/error metrics and empty severity map; zero
   Stats does not change gates. No private fields/observer semantics changes.
6. Pinned Go 1.26.8 Stats/pipeline/API complete compile-only chain; static exact
   source/direct boundaries/Go-format/docs/179 JSON/165 complete unique roadmap
   pairs/prior Definitions/R90-75/history/horizon/links/fences/seven paths/diff/
   sensitive. All binaries/benchmarks unexecuted.
7. Feature plus one docs-only closure exact full-SHA commit/push/fetch/Vault scope/
   note/index/MOC/identifier resolution; 14 stable notes reconcile current authority
   with entire prior prose archived/topic tails/immutable hashes retained;
   identical replay, refresh queue and stop.

## Non-goals and stop conditions

No zero Stats start-time/default-label initialization, New/Snapshot/renderer/API
changes, Worker runtime changes, write/export/terminal policy, live cross-counter
transaction, overflow guarantees, storage/rule/config/schema/dependencies/toolchain/
suites/private inputs/IPv6/publication. Stop for competing edits, ambiguous static/
compile/Git/Vault, new private/product/external authority or following increment.
Existing skills cover lazy ownership/locking/joined invariants; no redundant update.
Next selection only on a separate trigger.


## Implementation and validation checkpoint

Runtime diff adds only three-line nil severity-map allocation inside existing
locked loop after nil-entry skip. All other tracked engine/New/Snapshot/renderer/
Worker source byte-identical. Zero StartedAt and observed-only labels remain;
New retains four initialized labels and nonzero start time. Nil/empty/all-nil,
blank-to-low/dynamic/repeated counts, input ownership and write/export/terminal
policies unchanged. No throughput/allocation or transactional-live-snapshot claim.

Four external public regression functions authored/compiled only. Zero/New nine
batch cases each twice compare entire snapshots after seeding unrelated counters/
durations/buckets/queue/start time, exact initial label shapes, total/severity
sum/Prometheus lines, pointers/values and returned map-copy isolation. Typed nil
Stats remains exact zero snapshot. Four start-synchronized zero Stats writers each
100 mixed nil/high/blank/custom plus all-nil observations join before full snapshot/
exposition/input assertions: total 1200, high/low/custom each 400, zero start time.
Actual public Worker.Run with zero Stats and real SQLite writer records two mixed
nil-batch rows with expected count/timestamps/healthy store, processed/completed/
write/generated/severity/panic counters and unchanged packet/matcher calls.
Separate public no-alert and injected writer-failure controls preserve completion/
write-error/empty-label gates. Encoded SQLite path exercised; small matcher fixture
explicitly used, no claim of actual rule matching. No private state, sleeps,
panic-catching test code or weakened skip paths. No runtime/SQLite/race pass.

Final pinned Go 1.26.8 Stats/pipeline/API complete compile-only chain passed after
adding explicit constructor initial-label assertions; binaries/benchmarks
unexecuted. Static exact source/direct boundaries/Go-format/docs/179 JSON/165
complete unique roadmap pairs/prior Definitions/R90-75/history/horizon/links/
fences/seven paths/diff/sensitive passed. No unresolved validation failure or
ambiguity. All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
execution **not run; delegated by user**. 401 baseline Vault hashes unchanged.
Existing skills cover locked first use, direct boundaries and joined invariants;
no redundant update. Feature plus one docs-only closure remains; no next increment
or publication started.
