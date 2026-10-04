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
