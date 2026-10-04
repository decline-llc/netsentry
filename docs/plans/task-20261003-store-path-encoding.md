# R90-167: preserve question marks in writable SQLite filenames

## Selection and authority

Fetched clean main HEAD/origin/main/FETCH_HEAD:
`ffc308e71cd4cb14d6191108547f6d5742321389`. R90-166 feature and closure
exact Git scopes, Vault notes/index/MOC and fourteen stable current records agree.
Sep 5–Oct 3 phase audit covers 113 commits: SLO evidence tools, patched release
and toolchain, followed by core correctness. No new departmental R90-75 acceptance
or missing delivery. All 170 roadmap rows/Definitions match uniquely. Only R90-75
remains unfinished, with its full independent departmental contract unchanged.
Restore one source-grounded ready repair; Oct 3–Dec 31 forecast remains.

Pinned modernc.org/sqlite v1.34.5 newConn splits a non-file DSN at its first `?`
and processes the suffix as options. writableDatabaseDSN currently returns every
ordinary filename raw. A legal question mark can therefore select a truncated
filename or driver options instead of the pathname that was preflighted.

## Scope and risk

Six paths: engine/internal/alert/store.go, new
engine/internal/alert/store_path_encoding_test.go, docs/architecture.md, this plan,
docs/tasks/task-state-20261003-store-path-encoding.json, rolling roadmap.
Use the existing absolute file-URI builder for ordinary paths containing `?`;
set synchronous(FULL) only for durable mode. Preserve raw ordinary paths without
question marks, existing durable URI options/diagnostic, Store.Path and recovery
paths. The same helper covers primary and non-current daily-shard writable opens.
Low risk: filesystem names containing `?` become literal filenames; interpreting
such names as driver query options is no longer supported. No dependency change.

## Acceptance and planned evidence

1. Runtime helper-only diff; URI escaping protects the complete filesystem path,
   ordinary mode adds no driver query and durable mode retains its exact pragma.
   Review against pinned driver source and compile owning Go 1.26.8 module.
2. Direct public primary Open/WriteBatch/Query/List/Count/Close/reopen regressions
   for ordinary controls, question marks in directory and basename, combined
   spaces/percent/hash/question marks, and a query-looking pragma suffix. Cover
   ordinary DELETE and durable WAL modes. Assert exact Store.Path, expected rows,
   input preservation, exact file tree and independently read-only observed rows.
3. Direct daily public controls with ordinary/encoded directory names and both
   modes write current plus historical timestamps, reaching openShard and
   read-only historical query/count, close/reopen and observe both actual files.
   Exact query totals/content and logical per-file rows; no alternate truncated
   DB, private fields, mocks, sleeps or skips.
4. Direct malformed recovery on encoded primary path rejects with integrity
   sentinel/nil Store and unchanged whole tree before a writable open; ordinary
   and durable controls. Reuse existing non-following tree snapshot helper.
5. Authored assertions and pinned Go 1.26.8 alert/API/CLI/pipeline compile-only
   binaries, unexecuted. Static Go formatting, docs, JSON, complete unique roadmap
   multisets, prior Definitions/history/R90-75, intended scope, diff and sensitive
   review. Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
   execution **not run; delegated by user** under standing Sep 25 authority.
6. One six-path feature and one docs-only closure: non-force push, immediate fetch
   equality, full-SHA Vault ranges/note/index/MOC; affected stable current prose
   reconciled with prior substantive prose archived, immutable iteration notes
   preserved and exact replay idempotence. Refresh queue; no next implementation.

## Non-goals and stop conditions

No SQLite URI input API, new in-memory/path/symlink policy, changes to other
ordinary filenames, read-only DSNs, recovery algorithms, schema, lifecycle,
retention, cancellation, journal/default/durability policy, API/runtime behavior,
private inputs, dependencies/toolchain, release/publication or SLO acceptance.
Stop for competing edits, ambiguous compile/static/Git/Vault or new authority.
No test execution or runtime pass inferred. Existing URI regression instruction
covers this repair; no redundant skill update. Exactly one increment.


## R90-167 Implementation and Compile/Static Checkpoint (2026-10-03)

Runtime changes only writableDatabaseDSN: ordinary filenames with literal `?`
join the existing absolute encoded URI path; ordinary mode has no query options,
durable mode retains synchronous(FULL). All other engine runtime paths unchanged.
Store.Path and recovery names retain the caller's actual filename.

Three external public regressions authored: twelve primary controls spanning six
path shapes and ordinary DELETE/durable WAL; six daily controls spanning three
directory shapes/both modes, current/historical writes and existing historical
preflight on retry; two malformed-recovery preservation controls. Primary and
daily cases assert exact Store.Path, Query count/content/aggregation/timestamps,
List equality/Count, unchanged input, Close/reopen, independent encoded mode=ro
observations of rows/aggregates, complete expected non-sidecar file set and cleared
recovery logs. Only sidecars beside exact expected database names are allowed;
read-only observers can legitimately leave WAL sidecars. Query-looking pragma
suffix stays literal. Malformed recovery asserts integrity sentinel/nil Store
and complete tree membership/bytes/modes preservation. No private seams/mocks/
sleeps/skips/panic swallowing; all assertions authored/compiled, unexecuted.

Initial compile chain stopped at an incorrect CLI package path after alert/API
compilation. Resolved actual engine/cmd/netsentry module path and reran the complete
fail-fast Go 1.26.8 alert/API/CLI/pipeline compile-only chain successfully; no
partial-chain evidence retained. Binaries unexecuted. Static format/docs/JSON,
unique roadmap multisets/prior Definitions/R90-75/history/horizon/scope/diff/
sensitive review required before staging. Behavioral/race/CLI/full-suite/scanner/
knowledge/traffic/acceptance execution **not run; delegated by user**. No filesystem/
SQLite/runtime/durability/race/SLO pass inferred. Existing skill preflight and URI
rules cover the workflow; no redundant update. No next increment started.


Final static review passed: exact helper-only transform, 73 other tracked engine
paths unchanged, Go formatting, docs-check, 185 task JSON files, 171 complete
unique roadmap row/Definition pairs, all 170 prior Definitions/R90-75 contract,
ordered history/active horizon/local links/fences, six paths, diff and added-content
sensitive review. All 413 Vault baseline hashes remain unchanged. Existing
historical documented paths were excluded from new-content sensitive review;
no new sensitive paths or credential matches. Initial CLI path deviation fully
resolved by complete rerun. Feature delivery and single docs-only closure remain.
