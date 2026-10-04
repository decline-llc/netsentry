# R90-171: reject SQLite busy timeout overflow before startup

## Selection and authority

Clean fresh-fetched main baseline: `5fb9da8806f787501421846413b06f6438dddf7a`.
R90-170 feature/closure exact Git/Vault notes/index/MOC and fourteen current
stable references verified. Sep 12–Oct 3 phase audit covers SLO adapters and
inventory admission, patched release/toolchain and bounded core correctness.
All 121 phase commits have note/index coverage; latest exact ranges have MOC
links. No missing delivery or new qualifying R90-75 outcome. 174 unique prior
roadmap row/Definition pairs; independent R90-75 full departmental contract,
testing split and Oct 3–Dec 31 horizon retained. Captured 421 Vault hashes,
374 immutable iteration notes and fourteen stable prose backups.

Empty local ready queue restored from pinned modernc.org/sqlite v1.34.5 source:
PRAGMA busy_timeout calls _sqlite3Atoi, whose _sqlite3GetInt32 rejects values
above 2147483647 and leaves zero; Xsqlite3_busy_timeout takes int32 milliseconds.
NetSentry currently interpolates an unrestricted positive Go int, so a larger
positive setting disables the busy handler. This is source inference, not an
executed failure. Authoritative [SQLite busy timeout API](https://www.sqlite.org/c3ref/busy_timeout.html)
and [PRAGMA documentation](https://www.sqlite.org/pragma.html#pragma_busy_timeout)
confirm the millisecond argument and nonpositive disabled-handler semantics;
the exact overflow behavior is established from the repository-pinned parser.

## Scope, risk, non-goals and stop condition

Eight paths: engine/internal/config/config.go;
engine/internal/config/busy_timeout_bounds_test.go;
engine/internal/alert/store.go; engine/internal/alert/store_busy_timeout_bounds_test.go;
docs/architecture.md; this plan;
docs/tasks/task-state-20261003-busy-timeout-bounds.json; rolling roadmap.
Add matching signed-32-bit positive upper bounds in config validation and Open
before recovery/filesystem/SQLite work. Keep config negative/zero values and
Store nonpositive default 5000, already-done context precedence and existing
durable-journal validation precedence. Low risk: only previously unrepresentable
positive values reject. No dependencies, new operational timeout cap, changed
locking/retry/driver cleanup, pragma connection lifecycle, schema/recovery,
retention, journal semantics, SLO acceptance or publication. Stop on competing
edits, ambiguous compile/static/Git/Vault evidence or new authority.

All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Regression binaries compile only, unexecuted.

## Acceptance mapped to evidence

1. Source-only boundary checks preserve all other runtime behavior. Direct
   public config Load accepts defaults/nonpositive/1/5000/2147483647, rejects
   larger native integers with a named diagnostic, preserves YAML bytes and
   applies the same boundary after environment expansion. Values outside the
   native Go int domain retain decoder errors. Combined diagnostics retain order.
2. Public Open rejects larger native positive ints with nil Store and exact
   option diagnostic before absent directories, recovery input or existing
   database changes. Primary/daily and ordinary/durable cases include compatible
   databases, corrupt DB/sidecars, malformed recovery and invalid parent
   occupants. Compare full tree bytes/modes/membership; compatible fixtures use
   a preopened/query-warmed independently encoded read-only observer before
   rejection, reused afterward. No writable reopen of rejected artifacts.
3. Already-canceled/deadline contexts and durable-journal validation preserve
   precedence over the new option error. Representable positives and
   nonpositive defaults in primary/daily ordinary/durable stores retain actual
   effective PRAGMA value on the live SQLite connection, public write/query and
   reopen behavior. PRAGMA inspection is read-only access to the real connection,
   not an injected driver or private synchronization seam. No lock contention or
   elapsed-time claims. Direct overflow cases apply only when a native int can
   represent them; native decoder/domain behavior is explicit.
4. Preflight exact Go 1.26.8 and owning engine packages; fail-fast compile-only
   config/alert/API/cmd/netsentry/pipeline chain. Static exact source/format/docs/
   JSON/roadmap multisets/174 prior Definitions/R90-75/split/history/horizon/
   links/fences/eight-path scope/diff/sensitive review and unchanged Vault hashes.
5. Focused feature and one docs-only delivery record when needed. Non-force
   push/immediate fetch equality; full-SHA Vault note/index/MOC, stable current
   authority reconciliation, all prior substantive prose and immutable notes
   preserved, identical-range replay hashes stable. Refresh queue without
   starting another increment or duplicating completed delivery.

## Checkpoints

Plan/state persisted before runtime/architecture/queue edits. No runtime,
contention, durability, race or SLO pass follows from compile/static evidence.
Existing skill input-boundary and preservation instructions cover this repair;
no redundant skill update planned.


## R90-171 Compile and Static Checkpoint (2026-10-03)

Runtime diff exactly matches plan: matching signed-32-bit maximum constants and
one positive upper-bound guard each in configuration validation and Store.Open.
Existing nonpositive values/default 5000 and context/durable-journal precedence
retained; all other runtime unchanged. Six direct regression functions authored:
config boundary/default/whole-config/YAML-preservation, environment expansion,
combined diagnostics/native-domain parsing; sixty Linux-amd64 rejection cases
(three overflows, both store modes/both journal modes/five filesystem fixtures),
three precedence cases, twenty-four accepted-value/mode cases with two actual
opens each. Native 32-bit direct positive overflows are unrepresentable; those
branches explicitly omitted, with config decoder assertions retained.

Rejections reach public Open and assert exact error/nil Store before clock/path
resolution, complete bytes/modes/membership preservation and compatible-row
retention through an independent encoded mode=ro observer opened/query-warmed
before rejection and reused afterward. Corrupt DB/WAL/SHM, malformed recovery,
absent directory and regular-file parent controls included. Positive controls
inspect actual live-connection PRAGMA and stored effective value, then public
write/query/caller preservation and close/reopen. No driver injection, private
synchronization seam, sleeps or contention/timing claims. Assertions unexecuted.

Preflighted Go 1.26.8 and five owning packages; complete fail-fast config/alert/
API/cmd/netsentry/pipeline compile-only chain passed, binaries unexecuted. Static
exact source/format/docs/189 JSON/175 unique roadmap pairs/174 prior Definitions/
R90-75/testing split/history/horizon/links/fences/eight-path scope/diff/sensitive
review passed. All 421 baseline Vault Markdown hashes unchanged. Behavioral/
race/CLI/full-suite/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Authored regression debt remains departmental; no runtime, pragma,
persistence, contention, race or SLO pass inferred. No validation failure or
scope deviation; generic skill rules suffice. No next increment started.


## R90-171 Completion and Forward Queue Refresh (2026-10-03)

Feature `3a11443ca30e426699ec0e85de08b6b0bb4e5bb8` contains exactly the eight planned paths. Non-force push
and immediate fresh fetch verified clean main HEAD/origin/main/FETCH_HEAD
equality. Exact full-SHA range `5fb9da8806f787501421846413b06f6438dddf7a..3a11443ca30e426699ec0e85de08b6b0bb4e5bb8` synchronized to
`04-开发迭代记录/2026-10-03-3a11443ca3-CI知识同步.md`; Git-resolved identifiers, note/index/MOC and eight-path scope verified.
Fourteen stable current notes reconciled; all prior substantive current/topic/
history prose archived exactly outside only the actual generated CI MOC region
resolved from versioned constants. Legacy MOC content unchanged. All 374
baseline immutable iteration notes preserved. Identical feature replay preserves
all 422 Markdown hashes; snapshot JSON SHA-256 `04640a5d32dbca4f6c351b4bee076b1f1351c8c3e626aa298e79de601a9c6490`. Existing unique
sibling local Vault explicitly supplied; no second/remote Vault.

Acceptance matches plan: two matching maximum constants and one positive-bound
guard per config/Open, before persistent side effects. All other runtime
unchanged; nonpositive defaulting/context/durable-journal precedence retained.
Six authored functions directly reach config/YAML/default/env/diagnostic/native
int boundaries, sixty Linux-amd64 primary/daily ordinary/durable Open rejection
cases across three values and five fixtures, compatible independent read-only
row observation reused after rejection, full tree bytes/modes/membership,
three precedence cases, twenty-four accepted-value/mode cases with two opens
and actual live-connection PRAGMA/stored value/public write/query/input/reopen.
No weaker external-connection pragma observation, driver injection, private
synchronization seams, sleeps or contention/timing claims. Native32 impossible
positive-overflow cases explicitly omitted; decoder assertions authored. All
assertions unexecuted; no boundary upgraded to a behavioral pass.

Exact Go 1.26.8 five-package preflight/fail-fast compile-only and static exact
source/format/docs/189 JSON/175 unique pairs/174 prior Definitions/R90-75/split/
history/horizon/links/fences/eight paths/diff/sensitive review passed. Binaries
unexecuted. Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
**not run; delegated by user**. Growing authored regression debt remains
with department; no runtime/pragma/persistence/contention/race/SLO pass inferred.
No validation failure or scope deviation. Existing skill rules suffice.

This single three-path docs-only record closes the same increment. Derive its
full SHA from Git; verify non-force push/fresh fetched refs and exact
feature..closure Vault note/index/MOC before reporting. Do not repeat verified
feature/closure delivery or create a self-reference closure. Forward queue
refreshed: no other defined local ready increment. R90-75 independent full
asynchronous departmental contract and Oct 3–Dec 31 horizon retained. Next
trigger verifies closure/Vault, audits fresh history/code/queue and persists a
separate eligible plan before editing. No next implementation started.
