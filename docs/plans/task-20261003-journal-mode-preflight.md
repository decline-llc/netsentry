# R90-172: validate SQLite journal mode before startup side effects

## Selection and authority

Clean fresh-fetched main baseline: `65dda10dbea44262ce21509281f186fb261be0a8`. R90-171 feature/closure
Git ranges, exact note/index/MOC and fourteen stable current references verified.
Sep 12–Oct 3 phase audit: 123 commits have immutable note/index coverage;
latest ranges have bounded MOC links. Patched release/toolchain, SLO adapters
and bounded correctness history retained. No missing delivery or supplied
R90-75 outcome. All 175 unique row/Definition pairs agree. Independent R90-75
full departmental contract, testing split and Oct 3–Dec 31 horizon retained.
Captured 423 Markdown hashes, 376 immutable notes and fourteen stable
prose backups. Empty ready queue restored from direct source: Open currently
reads recovery input, creates parent directories and inspects existing databases
before init rejects an unsupported journal mode. This is source inference,
not an executed failure. Move the existing allowlist before these operations.

## Scope, risk, non-goals and stop condition

Six paths: engine/internal/alert/store.go;
engine/internal/alert/store_journal_preflight_test.go; docs/architecture.md;
this plan; docs/tasks/task-state-20261003-journal-mode-preflight.json;
rolling roadmap. Low risk: unsupported ordinary-mode options reject earlier
with the existing raw-value diagnostic. Preserve supported modes, ordinary
empty/whitespace WAL defaults, normalization, context/durable-WAL/busy-bound
precedence and caller options. No config contract change, dependencies,
new supported modes, schema/recovery/retention/locking changes, SLO acceptance
or publication. Stop on competing edits, ambiguous compile/static/Git/Vault
evidence or new authority. All execution (behavioral/race/CLI/full/scanner/
knowledge/traffic/acceptance) **not run; delegated by user**; compile only.

## Acceptance mapped to evidence

1. Move only existing allowlist validation from init to Open after context,
   durable-WAL and busy-bound checks, before path/clock/recovery/filesystem/DB
   operations. Preserve raw invalid-option diagnostic and accepted normalization.
2. Direct public Open rejects invalid token, near-match, padded invalid token
   and SQL-like text across primary/daily stores and absent, healthy database,
   corrupt DB/WAL/SHM, malformed recovery and regular-file parent fixtures.
   Assert nil Store, exact diagnostic, no clock call, caller options unchanged,
   full tree bytes/modes/membership. Healthy fixture uses independent encoded
   absolute mode=ro handle opened/warmed before rejection and reused afterward.
   Public paths include spaces and reserved URI characters; no writable reopen
   after rejected input. No runtime preservation claim until department executes.
3. Direct precedence controls cover canceled/deadline contexts, durable invalid
   and blank modes, supported non-WAL durable rejection, and combined busy
   overflow plus invalid journal on native64. Accepted empty/whitespace defaults,
   all six modes, lower-case/padded normalization in primary/daily stores and
   durable empty/padded-WAL controls observe real live PRAGMA journal_mode, public
   write/query/close/reopen and unchanged caller alert/options. No driver mock,
   private synchronization seam, sleeps or timing/durability claims.
4. Preflight pinned Go 1.26.8; fail-fast compile-only alert/API/cmd/pipeline.
   Static exact runtime diff/format/docs/JSON/176 roadmap multisets/175 prior
   Definitions/R90-75/split/horizon/history/links/fences/six paths/sensitive/diff
   and unchanged Vault hashes. Binaries remain unexecuted.
5. Focused feature and at most one docs-only delivery record; non-force push
   and immediate fetch equality, exact full-SHA Vault note/index/MOC, fourteen
   stable current-authority updates with prior substantive prose and immutable
   notes preserved, identical replay hashes. Refresh queue; no next increment.

## Checkpoints

Plan/state persisted before runtime/architecture/queue edits. Existing skill
input-boundary/preservation rules suffice; no redundant skill update. Tests
are authored evidence pending departmental execution, not behavioral passes.


## R90-172 Compile and Static Checkpoint (2026-10-03)

Exact runtime diff moves the existing six-mode allowlist from init to Open,
after unchanged context/durable-WAL/busy-bound guards and before path/clock/
recovery/filesystem/DB work. Supported normalization/defaults and raw invalid
mode diagnostic retained; all other runtime unchanged. Three direct functions
authored: forty primary/daily invalid-mode cases across four strings and five
fixtures; six earlier-diagnostic controls on native64 (unrepresentable busy
positive overflow omitted on native32); thirty-two accepted mode/store cases
with two opens each, including all six modes, blank/default and lower/padded
spellings plus durable empty/padded-WAL. Healthy rejection observes retained
rows through an independently encoded mode=ro handle opened/warmed before
rejection and reused afterward. Full tree bytes/modes/membership, no clock,
caller options and public alert preservation asserted. Positive controls reach
actual live-connection PRAGMA and public write/query/close/reopen. Reused real
fixture/snapshot helpers; no driver injection, private synchronization seam,
sleep or timing claim. All assertions unexecuted.

Preflighted exact Go 1.26.8 and owning packages; fail-fast alert/API/cmd/pipeline
compile-only passed, binaries unexecuted. Static exact source/format/docs/190
JSON/176 unique roadmap pairs/175 preserved Definitions/R90-75/split/history/
horizon/links/fences/six-path scope/diff/sensitive review passed. All 423 Vault
baseline Markdown hashes unchanged. Behavioral/race/CLI/full/scanner/knowledge/
traffic/acceptance **not run; delegated by user**. No runtime/persistence/
durability/race/SLO pass inferred; regression debt remains departmental.
No failure or scope deviation; existing skill instructions suffice. Initial
audit count refined to include every historical iteration note, not only CI
notes: 376 immutable notes captured/preserved. No next increment started.
