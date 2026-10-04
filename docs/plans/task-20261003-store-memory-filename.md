# R90-169: preserve the literal SQLite :memory: filename

## Selection and authority

Clean fresh-fetched main baseline: `fa299d7493486b0c660b652c2973069e801c249a`. R90-168 feature and closure
are delivered; exact Git/Vault range notes, full index, MOC and fourteen stable
current references verified. Sep 12–Oct 3 phase audit covers SLO adapters and
inventory admission, patched release/toolchain, and bounded core correctness.
No new R90-75 acceptance or missing delivery found. All 172 prior unique roadmap
row/Definition pairs agree; Oct 3–Dec 31 horizon and independent asynchronous
R90-75 full acceptance contract retained. Empty ready queue restored with one
source-grounded local repair; no next increment starts.

Pinned modernc.org/sqlite v1.34.5 opens with SQLITE_OPEN_URI and bundled SQLite
sqlite3BtreeOpen recognizes exact :memory: as ephemeral. NetSentry Options.Path
feeds filesystem preflight, Store.Path and recovery naming; no repository use or
contract promises an in-memory API. Ordinary writableDatabaseDSN returns exact
:memory: raw, while durable mode already uses an absolute URI. Infer from this
source that ordinary close loses rows and preflight may inspect a different file;
no behavioral failure observed. SQLite authoritative documentation confirms exact
sentinel versus ./ prefixed disk paths: https://www.sqlite.org/inmemorydb.html.

## Scope, risk, non-goals and stop condition

Six intended paths: engine/internal/alert/store.go; new
engine/internal/alert/store_memory_filename_test.go; docs/architecture.md;
this plan; docs/tasks/task-state-20261003-store-memory-filename.json;
docs/plans/rolling-90-day-roadmap.md. Add only exact :memory: exclusion to
ordinary raw-path guard and clarify comment; existing absolute URI builder
handles it. Low risk: this filename becomes persistent in ordinary mode, matching
durable mode and filesystem preflight. No dependencies, in-memory/URI-input API,
daily-shard algorithm, read-only helper, schema/recovery/lifecycle/retention,
CLI/API, journal or durable-pragma changes. Daily resolved basenames cannot equal
the exact sentinel, so no new daily coverage claimed. Stop on competing edits,
ambiguous compile/static/Git/Vault evidence or need for new authority.

User-directed department testing split remains active: behavioral/race/CLI/full
suite/scanner/knowledge/traffic/acceptance **not run; delegated by user**.
Compile/static and Git/Vault verification only; binaries unexecuted and no
runtime/SQLite/durability/race/SLO pass inferred.

## Acceptance mapped to evidence

1. Guard/comment-only runtime diff. Review pinned exact sentinel recognition,
   existing URI output and unchanged other runtime paths; format/diff checks.
2. Direct public primary cases in isolated t.Chdir with a space-bearing base,
   keeping exact :memory: input relative. Both ordinary DELETE and durable WAL;
   controls `./:memory:`, uppercase `:MEMORY:`, and a spaced
   directory component. Assert exact Store.Path, create/write/Query/List/Count,
   unchanged input, Close/reopen and repeated aggregation; independent encoded
   absolute mode=ro row/aggregate observations and exact files/cleared logs.
3. Existing compatible literal :memory: file seeded via public absolute path;
   preopen/query independent read-only observer before relative writer, reuse
   after write/Close/reopen; prove original rows plus repeated aggregate persist.
   Both modes. No mocks/private seams/sleeps/skips/panic swallowing.
4. Exact relative sentinel malformed recovery and corrupt database controls in
   both modes. Established respective integrity sentinel, nil Store and whole
   tree byte/mode/membership preservation before writable open. Corrupt case
   requires actual SQLite not-a-database diagnostic; no writable rejected reopen.
5. Preflight Go 1.26.8 and owning engine alert/API/cmd/netsentry/pipeline package
   roots; complete fail-fast compile-only chain. Static docs/JSON/roadmap raw
   counts and multisets/prior Definition and R90-75 preservation/ordered history/
   horizon/local links/fences/exact six-path scope/sensitive review.
6. Focused feature plus one docs-only delivery record if needed. Non-force push,
   immediate fetch equality, exact full-SHA Vault note/index/MOC, fourteen stable
   current prose reconciled with prior prose archived exactly, immutable notes
   preserved, identical-range replay hashes unchanged. Refresh future queue
   without starting next increment; resume verifies closure instead of repeating.

## Checkpoints

Plan/state persisted before runtime or architecture edits. All direct assertions
will be authored and compiled only, with delegated execution explicit.


## Final compile and static checkpoint (2026-10-03)

Runtime matches exactly the planned guard condition/comment transform; all other
tracked engine paths unchanged. Fourteen direct cases authored in three public
regression functions: eight fresh-path controls (four forms/two modes), two
compatible-existing cases and four integrity rejection cases. Exact :memory:
and ./ prefixed input remain relative at public Open; temporary base includes
spaces/percent/hash. Positive cases assert returned paths, Query totals/content/
aggregation/timestamps, List/Count, unchanged alerts, close/reopen and independent
absolute read-only logical rows plus exact files/cleared recovery logs. Existing
cases seed absolute literal files in the selected mode, preopen/query a separate
read-only observer before relative writer, reuse it after write/close/reopen and
assert durable row aggregation. Rejections assert sentinel/nil Store/full tree
bytes/modes/membership; corrupt fixture requires actual SQLite diagnostic. No
private seams, fake DBs, sleeps, skips or writable rejected-artifact reopen.

Preflighted Go 1.26.8 and four owning package directories; complete fail-fast
alert/API/cmd/netsentry/pipeline compile-only chain passed, binaries unexecuted.
Static exact source transform, Go format, make docs-check, 187 task JSON files,
173 unique row/Definition multisets, all 172 prior Definitions including R90-75,
testing split/horizon/history order, local links/fences/exact six paths/diff/
sensitive review passed. All 417 baseline Vault hashes unchanged. Behavioral/
race/CLI/full-suite/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Recent implementation throughput still adds unexecuted regression debt;
compile success is no runtime, SQLite, durability, race or SLO evidence. No
validation failure or scope deviation. Existing skill filename/observer rules
cover the repair; no reusable new workflow lesson warrants a redundant update.


## R90-169 Completion and Forward Queue Refresh (2026-10-03)

Feature `5df8af0637bb64dbc1de9d12ef29e9fa9822b819` contains exactly the six planned paths. Non-force
push and immediate fresh fetch verified clean main HEAD/origin/main/FETCH_HEAD
equality. Exact full-SHA range `fa299d7493486b0c660b652c2973069e801c249a..5df8af0637bb64dbc1de9d12ef29e9fa9822b819`
synchronized to `04-开发迭代记录/2026-10-03-5df8af0637-CI知识同步.md`;
Git-resolved identifiers, note/index/MOC and six-path scope verified. Fourteen
stable current notes reconciled; all prior substantive current/topic/history
prose retained exactly outside documented bounded generated MOC regions. All
370 baseline immutable iteration notes unchanged. Identical range replay
preserves all 418 Markdown hashes; snapshot JSON SHA-256
`f79b1d12d0f311d68d935865115e0925da09c2485b056d5fd3218c44c49e25d0`.
Existing unique sibling local Vault supplied explicitly; no second/remote Vault.
Phase audit confirms note/index coverage of all 117 Sep 12–Oct 3 baseline
commits; no new qualifying R90-75 acceptance. Unexecuted regression debt remains.

Acceptance matches plan: exact ordinary :memory: guard condition/comment only,
all other runtime retained. Other ordinary paths/durable FULL unchanged; no
in-memory API/daily algorithm/read-only helper change. Fourteen authored direct
cases reach exact relative public input, prefixed/uppercase/spaced controls,
create/write/Query/List/Count/close/reopen/repeated aggregation/unchanged alerts,
independent absolute read-only observation and exact files/cleared logs. Two
compatible-existing controls seed literal file via absolute public Open, preopen/
query-warm independent observer before relative writer and reuse it across both
opens. Four rejection controls preserve whole tree bytes/modes/membership with
established sentinel/nil Store; corrupt cases require actual SQLite diagnostic.
No fake/private seam/sleeps/skips/writable rejected reopen. All unexecuted.

Go 1.26.8 four-package preflight/full compile-only and static source/format/docs/
187 JSON/173 unique pairs/172 prior Definitions/R90-75/testing split/history/
horizon/links/fences/six paths/diff/sensitive passed; binaries unexecuted.
Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance **not run;
delegated by user**. No runtime/SQLite/durability/race/SLO pass inferred. No
validation failure or scope deviation; existing skill rules cover this repair.

This single three-path docs record closes the same increment. Resolve its full
SHA from Git and verify fetched remote plus exact feature..closure Vault note/
index/MOC; repair missing evidence only, without duplicate delivery or another
self-reference closure. Forward queue refreshed: no other defined local ready
increment. R90-75 independent asynchronous full contract and Oct 3–Dec 31 horizon
retained. Next trigger verifies completed closure, audits fresh history/code/
queue and persists a separate eligible plan. No next implementation started.
