# R90-166: reject already-canceled store startup before side effects

## Baseline and selection audit

Clean fetched main HEAD/origin/main/FETCH_HEAD agrees at
`6bf857cdef5490167df4ad658af2a483631a9654`. R90-165 seven-path feature and
three-path closure exact Git/Vault scopes, resolved identifiers, index/MOC and
stable closure facts verified. Sep 5–Oct 3 history has 111 commits covering
departmental SLO tooling, patched release/toolchain and core correctness. No new
qualifying R90-75 evidence or missing prior delivery found. All 169 unique roadmap
rows and Definitions match; only independent departmental R90-75 is unfinished.
Its full contract and Oct 3–Dec 31 horizon remain. No AGENTS or pre-existing edits;
engine module resolves Go 1.26.8. 411 Vault Markdown hashes/fourteen complete
stable backups captured. Source gap: Store.Open takes context but reads recovery
input and creates target directories before any context cancellation check.
Restore one bounded ready increment; persist this plan/state before other edits.

## Scope, risk and authority

Six paths: `engine/internal/alert/store.go`, new
`engine/internal/alert/store_open_cancellation_test.go`, `docs/architecture.md`,
this plan, `docs/tasks/task-state-20261003-store-open-cancellation.json`, roadmap.
Add only an entry ctx.Err guard before option defaulting/validation, recovery
preflight, pathname resolution, directory creation or database initialization.
Return nil Store and the context error unchanged. Low risk: only an already-done
context gains earlier failure/error precedence; live startup remains unchanged.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**, under standing Sep 25 instruction. Compile/static
review cannot establish filesystem/SQLite/runtime outcomes. No private input,
dependency/toolchain or publication authority is required.

## Acceptance mapped to evidence

1. Three-line entry guard only; all other tracked engine paths unchanged. Existing
   live defaults, integrity/schema/recovery/pruning and diagnostic logic retained.
   Canceled and already-expired contexts return their exact sentinel with nil Store
   before option/filesystem/recovery diagnostics.
2. Public Open regressions for two cancellation causes, ordinary and space-containing
   directory paths, and five fixtures: absent nested target; healthy populated
   existing DB; corrupt DB plus WAL/SHM/recovery artifacts; malformed recovery with
   missing DB; regular-file parent occupant. Twenty cases assert errors.Is/exact
   sentinel/nil Store and complete tree membership, file bytes and modes unchanged.
   Healthy fixture uses a separate URI-encoded read-only observer established before
   the rejected Open, proving logical row preservation without a writable reopen.
   Invalid durable-mode option controls prove cancellation precedes policy checks.
3. Public live startup controls with background and live cancelable contexts across
   ordinary/space paths create/open a real store, write and query/count one row,
   close successfully and confirm logical state through a read-only observer.
   No private fields, mock drivers, injected seams, sleeps or panic swallowing.
4. Pinned Go 1.26.8 alert/API/CLI/pipeline compile-only chain, binaries unexecuted.
   Static exact source/format/docs/184 JSON/170 complete unique roadmap pairs/prior
   Definitions/R90-75/history/horizon/links/fences/six paths/diff/sensitive review.
5. Six-path feature and one docs-only closure, non-force push/fresh fetch, exact
   full-SHA Vault ranges/note/index/MOC, fourteen current stable reconciliations,
   archived prior current prose/topic preservation, immutable notes and identical
   replay hashes. Refresh future queue without beginning another increment.

## Non-goals and stop conditions

No active-cancellation/startup interruption guarantee, nil-context support, new
option validation, context-aware recovery/preflight redesign, lifecycle/close/
retention/SQLite/driver/API/recovery algorithm or error precedence change for live
contexts, private inputs, suites, dependencies/toolchain, IPv6/release/publication
or runtime/SLO claim. Stop for competing edits, ambiguous static/compile/Git/Vault
or new product/private/external authority. Existing skill rules already cover
entry cancellation, preservation and encoded read-only URIs; no redundant update.
Finish only this increment; another selection requires another trigger.


## R90-166 Implementation and Static/Compile Checkpoint (2026-10-03)

Three-line ctx.Err guard at public Open entry only; all other 72 tracked engine
paths unchanged. Already-canceled and already-expired callers return unchanged
context sentinel/nil Store before option validation, path resolution, recovery
reads, directory creation or DB initialization; live startup code remains intact.
No active-cancellation or nil-context guarantee is introduced.

Three public external regression functions are authored: twenty cases span two
cancellation causes/two directory shapes/five fixtures (absent nested target,
healthy populated DB, corrupt DB/WAL/SHM/recovery artifacts, malformed recovery
with absent DB, regular-file parent occupant). Exact sentinel/errors.Is/nil Store
and complete tree membership/file bytes/modes are asserted. Healthy DB uses DELETE
journal mode during seeding, then an independent URI-encoded mode=ro observer
queries before snapshot/rejection and reuses that handle afterward, without a
writable reopen. Two invalid durable-mode controls assert cancellation precedes
policy validation and preserve the tree. Four background/live-cancelable controls
create real stores, WriteBatch/List/Count one fixture row, Close and verify one
row through read-only observation. No private fields/fake drivers/sleeps/skips/
panic swallowing; all assertions authored and compiled only, not executed.

Pinned owning Go 1.26.8 alert/API/CLI/pipeline complete compile-only chain passed;
binaries unexecuted. Static exact source/72 other engine paths/format/docs/184
JSON/170 complete unique roadmap pairs/169 prior Definitions/R90-75/history/
horizon/links/fences/six paths/diff/sensitive review passed before staging.
Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. No filesystem/SQLite/runtime/race/SLO pass inferred.
Existing skill instructions cover entry cancellation/read-only preservation/URI
encoding; no redundant skill update. Feature/single closure remain; no next work.


## R90-166 Completion and Forward Queue Refresh (2026-10-03)

Feature `2dcd6d46bb38d5807ff41dff94a2e9086aae4f34` contains exactly the six intended paths. Isolated branch
fast-forwarded freshly verified main; recorded old remote tip, non-force push
and immediate fresh fetch verified clean HEAD/origin/main/FETCH_HEAD equality.
Full-SHA range `6bf857cdef5490167df4ad658af2a483631a9654..2dcd6d46bb38d5807ff41dff94a2e9086aae4f34` synchronized to
`04-开发迭代记录/2026-10-03-2dcd6d46bb-CI知识同步.md`; exact six-path note scope,
Git-resolved short identifiers, full index/MOC verified. Fourteen current stable
notes reconciled; entire prior current prose archived and original topic/history
tails preserved exactly outside bounded generated MOC regions. All 364 baseline
immutable iteration notes preserved. Identical feature range replay preserves
412 Markdown hashes; snapshot JSON SHA-256
`f47d7a1f338e8c23a0bc610802b69f159cb85b931ce0652f900acdfd4b5c66b2`.
Existing unique sibling local Vault passed explicitly; no second/remote Vault.

Acceptance matches plan: three-line entry guard only; 72 other tracked engine
paths unchanged. Twenty public preservation cases cover canceled/expired contexts,
ordinary/space directory paths and all five named persistent/absent fixtures.
Exact sentinel/errors.Is/nil Store and complete membership/bytes/modes asserted;
healthy independent URI-encoded read-only observer established/query before
snapshot/rejection and reused afterward. Two invalid durable-mode policy precedence
controls and four background/live-cancelable real create/write/List/Count/Close/
read-only controls directly reach promised public boundaries. No private seams,
fake drivers/sleeps/skips/panic swallowing. All assertions authored/compiled only;
no active-startup cancellation/nil-context or filesystem/SQLite/runtime claim.

Complete pinned Go 1.26.8 alert/API/CLI/pipeline compile-only chain and static
source/direct boundaries/72 other engine paths/format/docs/184 JSON/170 unique
roadmap pairs/169 prior Definitions/R90-75/history/horizon/links/fences/six paths/
diff/sensitive review passed. Binaries unexecuted. Behavioral/race/CLI/full-suite/
scanner/knowledge/traffic/acceptance execution **not run; delegated by user**.
Resolved verification deviation: initial topic-tail comparison included the CI
MOC generator's bounded region. Source-defined generated markers were verified;
exclude only both bounded generated regions and compare all remaining prose
exactly. No topic loss or unresolved ambiguity. Existing skill instructions cover
entry cancellation/preservation/encoded URI/generated-region boundaries; no
redundant skill update.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; push/fresh-fetch/exact feature..closure Vault verification before
reporting, without another self-reference closure. After verification, do not
repeat feature/closure commit/push/sync. Future queue refreshed: no other defined
local ready increment. R90-75 remains independent asynchronous departmental
acceptance with full contract and Oct 3–Dec 31 horizon unchanged. Next trigger
verifies fetched closure/Vault, audits fresh code/queue and persists a separate
eligible plan before edits. No next increment started.
