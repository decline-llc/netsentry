# R90-168: preserve relative file-prefixed SQLite filenames

## Selection audit and source evidence

Clean fetched main HEAD/origin/main/FETCH_HEAD:
`4cf106a67ac142d478f6f7a0294872c26c6ce2f6`. R90-167 six-path feature/three-path
closure exact Git and Vault note/index/MOC verified; fourteen stable current
records agree with fetched baseline. Sep 5–Oct 3 phase audit covers 115 commits:
SLO evidence tooling, patched release/toolchain, core correctness. No qualifying
R90-75 acceptance or missing delivery. All 171 unique roadmap pairs agree;
R90-75 full independent departmental contract and Oct 3–Dec 31 horizon unchanged.
Empty local ready queue restored with one bounded source-grounded follow-up.

Pinned modernc.org/sqlite v1.34.5 newConn sets SQLITE_OPEN_URI. Its bundled Linux
SQLite sqlite3ParseUri compares exact five-byte `file:` prefix and strips scheme.
Ordinary writableDatabaseDSN passes such relative filesystem paths unchanged
when they lack question marks; writable open can target a different file from
filesystem preflight, recovery and Store.Path. R90-167 question-mark handling
and historical non-goals are preserved. Resolve actual owning engine module and
alert/API/cmd/netsentry/pipeline package directories before compile-only chain.

## Scope, risk and authority

Six paths: engine/internal/alert/store.go, new
engine/internal/alert/store_file_prefix_test.go, docs/architecture.md, this plan,
docs/tasks/task-state-20261003-store-file-prefix.json, rolling roadmap.
Extend only existing ordinary raw-path guard to exclude exact `file:` prefixes;
those paths use existing absolute encoded file-URI builder without driver query.
Retain question-mark repair, durable FULL pragma, other raw ordinary paths,
Store.Path, recovery names and read-only helper. No dependencies. Low risk:
file-prefixed relative paths become literal filesystem names, consistent with
preflight; no URI-input API is introduced. All behavioral/race/CLI/full-suite/
scanner/knowledge/traffic/acceptance execution **not run; delegated by user**.
Compile/static checks and Git/Vault verification only; no runtime pass inferred.

## Acceptance and planned evidence

1. One guard condition and relevant comment only; static exact source transform
   and other engine path preservation, pinned driver/SQLite URI-prefix review.
2. Public primary controls in isolated temporary current directories via t.Chdir:
   relative file-prefixed basename/directory, spaces/percent/hash encoded forms,
   ordinary and uppercase-prefix controls; ordinary DELETE and durable WAL modes.
   Create/write/Query/List/Count/Close/reopen/repeated aggregation, exact Store.Path,
   input preservation, independent absolute encoded read-only logical observation
   and exact expected non-sidecar filenames/cleared recovery logs. For paths whose
   prefix SQLite would strip, preseed its alternate target as a healthy decoy,
   open/query an independent read-only observer before actual writer, reuse it
   after close/reopen, and assert decoy rows/aggregate/file bytes/modes unchanged.
3. Public daily controls with ordinary/file-prefixed/encoded relative directories
   and both modes: current/historical writes and existing historical preflight on
   retry, cross-shard Query/List/Count and reopen, exact path/file membership,
   independent absolute read-only per-shard rows/aggregates and unchanged inputs.
   Observe actual historical openShard boundary, no fake/private seams/sleeps/skips.
4. Public rejection controls on relative encoded file-prefixed path and both modes:
   malformed recovery and corrupt existing database return established respective
   integrity sentinel/nil Store, preserving complete tree membership/bytes/modes
   before writable open. No writable reopen of rejected persistent artifacts.
5. Owning Go 1.26.8 complete alert/API/CLI/pipeline compile-only chain with package
   directory preflight and unexecuted binaries. Static format/docs/JSON/unique
   roadmap multisets/prior Definitions/R90-75/history/horizon/links/fences/six paths/
   diff/sensitive review. Authored cases remain unexecuted; delegated gates explicit.
6. Six-path feature and one docs-only closure, non-force push/immediate fetch
   equality, exact full-SHA Vault ranges/note/index/MOC. Reconcile fourteen stable
   current notes, archive all prior substantive prose, preserve immutable notes,
   replay-identical hashes. Refresh queue without next implementation.

## Non-goals and stop conditions

No URI/in-memory/symlink policy, absolute path redesign, read-only DSN changes,
new path/API validation, schema/recovery/lifecycle/retention/cancellation/journal/
durability/CLI/API algorithms, suites/private inputs/toolchain/dependencies,
publication or SLO pass. Preserve established exact case-sensitive SQLite prefix
semantics; uppercase FILE: remains ordinary. Stop for competing edits, ambiguous
compile/static/Git/Vault or new authority. Existing URI/observer/preflight rules
cover the change; no redundant skill update. Exactly one increment this trigger.


## R90-168 Implementation and Compile/Static Checkpoint (2026-10-03)

Runtime diff is one raw-path guard condition and one relevant comment only:
ordinary exact file-prefixed paths use existing absolute encoded URI builder;
question-mark handling, durable FULL pragma and other ordinary paths retained.
Pinned SQLite's five-byte case-sensitive file: recognition confirmed from bundled
parser and string table. No read-only, recovery, Store.Path or other runtime edit.

Three external public regression functions authored: fourteen primary controls
(seven relative path shapes, ordinary DELETE/durable WAL), six daily controls
(three relative directories/both modes), four recovery/database rejection controls.
Temporary t.Chdir preserves relative file: public input; an absolute fixture would
miss this branch. Primary cases include exact `file:` filename, uppercase FILE:
and ordinary controls, prefixed basename/directory and spaces/percent/hash. Eight
prefixed primary cases preseed the stripped healthy alternate target, establish/
query an independent encoded absolute mode=ro observer before writer, reuse it
after Close/reopen/repeated aggregation, and compare complete decoy DB/recovery
bytes and modes. Exact-prefix-only file: has no persistent alternate target.
All positive cases assert exact Store.Path, Query count/content/timestamps/
aggregates, List equality/Count, input preservation, independent per-file logical
reads and exact expected non-sidecar files/cleared recovery logs. Only legitimate
sidecars beside expected databases allowed. Daily historical writes and second
historical retry reach actual openShard and existing read-only preflight. Four
rejections assert established integrity sentinel/nil Store and full tree bytes/
modes/membership preservation without writable reopen. No fake/private seam/
sleep/skip/panic swallowing; assertions authored/compiled, unexecuted.

Owning engine Go 1.26.8 and alert/API/cmd/netsentry/pipeline directories preflighted;
complete fail-fast compile-only chain passed after final boundary fixture added,
binaries unexecuted. Static source/format/docs/JSON/roadmap/history/R90-75/scope/
sensitive review required before staging. Behavioral/race/CLI/full-suite/scanner/
knowledge/traffic/acceptance execution **not run; delegated by user**. No runtime/
SQLite/durability/race/SLO pass inferred; no compile-chain failure.

Reusable local netsentry-next instruction 11 refined: preserve relative/reserved-
prefix form in the public call so fixtures cannot bypass the affected branch;
observe artifacts via an independent encoded absolute path. Markdown checked;
separate from repository commit. No next increment started.


## R90-168 Plan Amendment: Read-Only Relative URI Prerequisite (2026-10-03)

Static review of pinned Go 1.26.8 net/url.URL.String and bundled SQLite URI parser
found a necessary prerequisite before delivery: relative resolved paths passed as
URL.Path serialize as file://relative-component/... with the first component in
URI authority position. SQLite rejects that authority. Thus a writable-only guard
cannot fulfill the promised public relative reopen/historical-read contract.
This is source evidence, not a behavioral test result; no failing runtime claimed.

Amend this same increment before prerequisite edit. The six-path scope is unchanged.
In readOnlyDatabaseDSN, make the EvalSymlinks result absolute before sidecar lookup
and URI serialization, using existing resolve-path error wording. Preserve mode=ro,
readonly_shm, symlink resolution, sidecar classification and rejection diagnostics.
Runtime scope is now two exact helpers: writable guard/comment and read-only
absolute-path normalization; every other runtime path retained. Primary reopen,
daily historical query/count/retry and uppercase/ordinary relative controls
already directly reach this prerequisite. Corrupt-input regressions additionally
require the real SQLite not-a-database diagnostic, so an invalid-authority error
cannot satisfy their integrity-sentinel assertion. Recompile complete owning
chain after final source/assertions and redo static review. No separate increment,
URI API, dependency, publication or testing authority is needed.

Earlier read-only-change non-goal is superseded only by this necessary path
normalization. Other non-goals, delegated execution, R90-75 and horizon unchanged.


## R90-168 Final Amended-Scope Validation (2026-10-03)

Final runtime is confined to readOnlyDatabaseDSN absolute normalization of the
EvalSymlinks result, and writableDatabaseDSN exact file: exclusion/comment. Existing
mode=ro/readonly_shm/sidecar/symlink/error classification and question-mark/durable
query behavior retained. All other 74 tracked engine paths unchanged. The pinned
Go URL serializer/SQLite parser comparison explains why both helpers are required.

Fourteen primary, six daily and four rejection controls directly cover the amended
scope. Inputs remain relative via isolated t.Chdir; exact file: and uppercase FILE:
controls included. Eight seeded healthy alternate targets use a preopened/query-
warmed independent encoded absolute mode=ro observer reused after actual writes
and reopen; their full DB/recovery bytes/modes and one-row aggregate are asserted.
Daily first/second historical writes reach non-current openShard, existing-file
preflight and historical Query/List/Count; independent reads use absolute URIs.
Corrupt database rejection additionally requires actual SQLite not-a-database
text, excluding a weaker invalid-authority rejection. Full expected file sets,
recovery clearing, inputs and rejected trees remain asserted. All unexecuted.

Final complete preflighted Go 1.26.8 alert/API/cmd/netsentry/pipeline compile-only
chain passed after both source helpers and corrupt-diagnostic assertion settled;
binaries unexecuted. Static exact two-helper transform/pinned parser/74 other
paths/format/docs/186 JSON/172 unique roadmap pairs/171 prior Definitions/R90-75/
ordered history/horizon/links/fences/six paths/diff/added-content sensitive review
passed. Temporary static-review script quoting syntax was repaired and the entire
static review rerun successfully; no partial validation retained. All 415 Vault
baseline Markdown hashes unchanged. Behavioral/race/CLI/full-suite/scanner/knowledge/
traffic/acceptance execution **not run; delegated by user**. No runtime/SQLite/
durability/race/SLO pass inferred. Only deviation is the source-evidenced necessary
read-only prerequisite within the same six paths, plus repaired temporary tooling.
Local skill instruction 11 fixture refinement is separate; no next increment.


## R90-168 Completion and Forward Queue Refresh (2026-10-03)

Feature `35ff8c067acfd58be17697952ab3e11c3bc4dfa7` contains exactly the six intended paths. Recorded old remote tip,
non-force push and immediate fresh fetch verified clean main HEAD/origin/main/
FETCH_HEAD equality. Full-SHA range `4cf106a67ac142d478f6f7a0294872c26c6ce2f6..35ff8c067acfd58be17697952ab3e11c3bc4dfa7` synchronized to
`04-开发迭代记录/2026-10-03-35ff8c067a-CI知识同步.md`; exact six-path scope,
Git-resolved short identifiers, full index and MOC links verified. Fourteen stable
current notes reconciled; all prior substantive current/topic/history prose
archived exactly outside documented bounded generated MOC regions. All 368
baseline immutable iteration notes unchanged. Identical feature replay preserves
416 Markdown hashes; snapshot JSON SHA-256 `4bbf186e6973c5c42614fc25656803bb4fb269739a36e6c2f8d1f70c739f38dd`. Existing unique sibling
local Vault supplied explicitly; no second/remote Vault.

Acceptance matches amended plan: writable exact file: guard/comment and read-only
absolute normalization only; 74 other tracked engine paths unchanged. Ordinary
question-mark handling, durable FULL query, other ordinary writes and mode=ro/
readonly_shm/symlink/sidecar/error classification retained. Pinned Go URL serializer
and SQLite parser showed the read-only relative authority prerequisite; same
six-path plan was amended before editing. This necessary helper normalization
completes promised relative reopen/historical reads; no new URI/in-memory API.

Fourteen primary cases span seven relative path shapes/both modes; exact file:
filename, uppercase FILE:, ordinary and prefixed encoded forms. Eight primary
cases preseed stripped alternate target and preopen/query-warm an independent
encoded absolute mode=ro observer reused after writes/reopen, assert one-row
aggregate and whole decoy DB/recovery bytes/modes unchanged. Six daily cases span
three relative directories/two modes; actual current/historical writable opens,
existing historical preflight/retry, cross-shard Query/List/Count and reopen.
Positive controls assert exact Store.Path, totals/content/timestamps/aggregates,
unchanged inputs, per-file independent reads and exact expected non-sidecar files/
cleared recovery logs. Only legitimate sidecars beside expected DBs allowed.
Four malformed-recovery/corrupt-DB rejections prove sentinel/nil Store/complete
tree bytes/modes/membership; corrupt cases require actual SQLite diagnostic to
exclude weaker URI-authority rejection. Public boundaries directly reached by
authored assertions; no fake/private seams/sleeps/skips/panic swallowing or
writable reopen of rejected artifacts. All unexecuted, no runtime pass inferred.

Final complete preflighted Go 1.26.8 alert/API/cmd/netsentry/pipeline compile-only
chain and static exact two-helper transform/pinned parser/74 other paths/format/
docs/186 JSON/172 unique roadmap pairs/171 prior Definitions/R90-75/history/horizon/
links/fences/six paths/diff/sensitive passed. Binaries unexecuted. Temporary
static-review quoting syntax repaired; complete review rerun successfully.
Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. No SQLite/runtime/durability/race/SLO pass inferred.
Local skill instruction 11 refined for relative/reserved-prefix public fixtures
and independent encoded absolute observations; Markdown checked, separate from
repository commit. No unresolved compile/static ambiguity or next implementation.

This single three-path docs-only record closes the same increment. Resolve full
SHA from Git; push/fresh-fetch/exact feature..closure Vault verification before
reporting. Do not repeat verified feature/closure delivery or add self-reference
closure. Forward queue refreshed: no other defined local ready increment. R90-75
full independent asynchronous departmental contract and Oct 3–Dec 31 horizon
unchanged. Next trigger verifies fetched closure/Vault, audits fresh code/queue
and persists a separate eligible plan before edits. No next increment started.
