# R90-158: validate calendar dates before shard cleanup

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`7cd71694b25f99349bd1c8da1331280fe3ca85ef`. R90-157 exact eight-path
feature/three-path closure Git/Vault scope/note/index/MOC verified; generated
short identifiers uniquely resolve to full endpoints. 95-commit Sep 5–Oct 3
phase remains core correctness delivery after SLO tooling; no new qualifying
R90-75 acceptance. Sole unfinished R90-75 full independent asynchronous contract
and Oct 3–Dec 31 horizon unchanged. Empty ready queue reconciled inside this
source-grounded retention deletion guard. Existing unique local Vault's 395
Markdown hashes/14 entire stable notes backed up outside Git. No AGENTS present.
Owning engine pinned Go 1.26.8 preflighted before compile-only checks.

## Scope, risk and authority

Six paths: `engine/internal/alert/store.go`, new
`engine/internal/alert/shard_calendar_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-shard-calendar.json`, rolling roadmap. Plan/state
persisted before runtime/other docs edits. Existing cleanup recognizes a digit
shape but accepts impossible dates such as 2025-02-29, deleting arbitrary contents
and WAL/SHM sidecars under those names. Before deleting an expired candidate,
require time.Parse with the existing canonical YYYY-MM-DD layout. Keep lexical
cutoff equality, valid daily names including leap days, disabled retention,
context/lifecycle behavior, file counts and removal order unchanged. Risk low:
invalid-date filenames and sidecars remain; operators may inspect/remove them
separately. Discovery/query behavior is unchanged and can still report invalid
unrelated files as errors. No new calendar year range, filename or retention
policy; supported time.Parse year 0000 remains eligible if expired.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. No filesystem/SQL/runtime pass inferred.

## Acceptance mapped to evidence

1. Exact runtime diff adds calendar parse/error skip before removeShardSet and
   clarifies its comment; every other tracked engine source stays byte-identical.
2. Public Store.Open startup and public PruneExpiredShardFiles direct regressions
   both preserve exact arbitrary bytes for every malformed date's base/WAL/SHM:
   non-leap Feb 29, Feb 30, Apr 31, zero/13 month, zero/32 day. Noncanonical and
   unrelated names and sidecars, orphan sidecars and a date-shaped directory with
   nested file are also retained. Fixtures precede Open for startup and follow
   Open for direct invocation; no private field injection or sleeps.
3. Same cases delete exactly nine files from three expired valid shard sets:
   leap-day 2024-02-29, ordinary prior day 2026-06-19, supported year 0000-01-01.
   Retain cutoff 2026-06-20, fresh day and current primary; direct returns nine
   then zero on replay. Startup verified by file absence/preservation after Open,
   with zero further deletions; closing store precedes final retained-byte check.
4. Public disabled-retention and pre-canceled contexts preserve all fixture bytes;
   canceled error retains errors.Is(context.Canceled), zero files counted.
5. Pinned Go 1.26.8 alert/API/pipeline complete compile-only chain; static exact
   source/direct boundaries/Go-format/docs/176 JSON/162 unique complete roadmap
   pairs/prior Definitions/R90-75/history/horizon/links/fences/six paths/diff/
   sensitive. Test binaries and existing benchmarks unexecuted.
6. Feature plus one docs-only closure exact full-SHA commit/push/fetch/Vault scope/
   note/index/MOC/identifier resolution; fourteen current stable notes with entire
   prior prose archived and topic tails/immutable hashes retained; identical
   replay. Refresh queue and stop after this increment.

## Non-goals and stop conditions

No shard discovery/query/writes, SQLite schema/recovery, active-handle retention,
retention cutoff policy, new filename/year restrictions, metrics/API/dependencies/
toolchain/suites/private inputs/IPv6/publication. Stop for competing edits,
ambiguous static/compile/Git/Vault or new private/product/external authority.
Existing skills cover preservation and input rejection; no redundant skill edit.
Next selection only on a separate trigger.


## Implementation and validation checkpoint

Runtime diff is only time.Parse/error-skip before expired set deletion and comment
clarification. Every other tracked engine source remains byte-identical. Existing
regex/cutoff equality/removal ordering/count/context/lifecycle/year range and
shard discovery/query/schema/recovery unchanged. Seven impossible-calendar
fixtures preserve arbitrary base/WAL/SHM bytes; no broader filename policy.

Two direct public regression functions authored/compiled only. Startup fixtures
precede Open; direct fixtures follow Open. Seven invalid dates, noncanonical and
unrelated names/sidecars, orphan sidecars and directory/nested file retain bytes.
Three valid expired sets (leap day, ordinary day, supported year zero) remove nine
files; explicit cleanup counts nine then zero, startup leaves zero further
removals. Cutoff/fresh/current retained; close precedes final byte checks.
Disabled retention and pre-canceled public cleanup retain all bytes, count zero
and preserve errors.Is(context.Canceled). No private fields, sleeps or skips.

Pinned Go 1.26.8 alert/API/pipeline complete compile-only chain passed; binaries
and benchmarks unexecuted. Exact source/direct boundaries/Go-format/docs/176 JSON/
162 unique complete roadmap pairs/prior Definitions/R90-75/history/horizon/links/
fences/six paths/diff/sensitive passed. Initial temporary static-review adaptation
pointed its prior-completion assertion at current R90-158; corrected to unique
R90-157 completion and reran the entire static/docs/diff chain successfully. No
repository behavior change or unresolved validation ambiguity from that tool
error. All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
execution **not run; delegated by user**; no filesystem/SQL/runtime/SLO pass.
395 baseline Vault hashes unchanged. Existing skills already cover unique markers
and preservation; no redundant update. Feature and single docs-only closure
remain; no next increment or publication started.


## Delivery and queue closeout

Feature `f78e2ce1386d4f3c20c68bfd9eb5e0ea5a32941e` contains exactly six planned paths. Isolated
fix/r90-158-shard-calendar fast-forwarded freshly verified main; push/fresh-fetch
verified clean HEAD/origin/main/FETCH_HEAD at that full SHA. Exact range
`7cd71694b25f99349bd1c8da1331280fe3ca85ef..f78e2ce1386d4f3c20c68bfd9eb5e0ea5a32941e` note `04-开发迭代记录/2026-10-03-f78e2ce138-CI知识同步.md`, six-path scope/index/MOC verified. Generated
short identifiers uniquely resolve through Git to full endpoints. Fourteen
current stable notes reconciled; entire prior substantive current prose archived
under R90-157 historical headings. Original topic tails and all
348 baseline immutable iteration hashes retained, excluding only bounded
documented generated MOC regions. Identical replay preserves 396 Markdown
hashes; snapshot JSON SHA-256 `34eff76b4c473fa4be6dd84afab06ecc719232fc96b0908bc8fd76e992be5877`. Existing unique
sibling local Vault selected explicitly; no second empty or remote Vault.

Acceptance matches persisted plan. Runtime adds calendar time.Parse/error-skip
before expired set deletion plus comment; every other tracked engine source
byte-identical. Existing cutoff equality/supported year range/valid leap days/
removal order/count/context/lifecycle/discovery/query/schema/recovery unchanged.
Invalid-date files remain for operator inspection; discovery can still report
unrelated-file read errors. No new filename/year/retention/active-handle policy.

Two direct public functions authored/compiled only. Startup fixtures precede
Open; direct fixtures follow Open. Seven impossible calendar dates preserve
exact arbitrary base/WAL/SHM bytes; noncanonical/unrelated names/sidecars, orphan
sidecars and date-shaped directory/nested file retained. Three valid expired
sets (leap day, ordinary day, supported year zero) remove nine files; direct
count nine then zero, startup leaves zero further removals. Cutoff/fresh/current
retained; Close precedes final retained-byte assertions. Disabled retention and
pre-canceled cleanup preserve all bytes/count zero/errors.Is(context.Canceled).
No private-field injection, sleeps or skips; no filesystem/SQL/runtime/race/SLO
pass inferred from unexecuted assertions.

Pinned Go 1.26.8 alert/API/pipeline complete compile-only chain passed; binaries
and benchmarks unexecuted. Exact source/direct boundaries/Go-format/docs/176 JSON/
162 unique complete roadmap pairs/prior Definitions/R90-75/history/horizon/links/
fences/six paths/diff/sensitive passed. Initial temporary static-review marker
pointed to current completion before it existed; corrected unique prior R90-157
marker and entire static/docs/diff rerun passed. No unresolved source/compile/
static/Git/Vault ambiguity. Existing skills cover marker/preservation, no redundant
update. All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
execution **not run; delegated by user**. Planning deviation: empty ready queue
restored inside source-grounded deletion guard.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; push/fresh-fetch and verify exact feature..closure Vault scope/
note/index/MOC/current stable prose before reporting; no self-reference closure.
Refreshed queue has no defined local dependency-ready item. R90-75 full independent
asynchronous departmental contract and Oct 3–Dec 31 horizon unchanged. Next trigger
verifies fetched closure/Vault and audits fresh code/queue before separate
eligible plan. No following increment started; do not repeat R90-157/R90-158
delivery or R90-59 publication. IPv6/external publication needs separate authority.
