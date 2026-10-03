# R90-150: overflow-safe daily-shard query bounds

## Scope and source-grounded selection

Ready after completed R90-149; existing public Store.Query daily-shard path.
Current horizon Oct 3–Dec 31, independent R90-75 departmental acceptance.
Fresh clean fetched main HEAD/origin/main/FETCH_HEAD: `8cd3b430384e3727d91ecb8e1783a4bb34998fc0`.
79-commit Sep 5–Oct 3 phase audit shows supplied-evidence SLO tooling followed
by core snapshot/concurrency/completion/privacy/pagination repairs. R90-149
exact seven-path feature and three-path closure Git/Vault note/index/MOC are
verified; no missing delivery or new qualifying acceptance evidence. Only R90-75
is unfinished. Reconcile empty local ready queue inside this correctness repair,
without an audit-only increment. Vault: 379 Markdown files, 14 stable full-content
backups, 332 immutable iteration hashes; snapshot JSON SHA-256
`704c24031f30d9245b1756d2a8b746f3de015aff4b6693d492b1f4c7788b37d2`.

Six paths: `engine/internal/alert/store.go`, new
`engine/internal/alert/query_bounds_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-shard-query-bounds.json` and
`docs/plans/rolling-90-day-roadmap.md`. Persist plan/state before runtime,
architecture or roadmap edits. Runtime change is only sliceBounds.

queryDailyShards normalizes limit/offset, then sliceBounds adds offset+limit
before clamping. A programmatic Query with limit=MaxInt, offset=1 and three
merged rows can overflow and panic while the primary SQL backend accepts it.
R90-149 HTTP per_page cap/offset validation does not cover public programmatic
Store.Query limits. Clamp limit to length-offset before adding, retaining
existing query normalization, sorting/filter/count/lifecycle/SQL semantics.
Risk medium: arithmetic bounds and primary/daily behavioral compatibility.

## Acceptance and evidence map

1. For normalized nonnegative offset and positive limit, return empty at/past
   length; otherwise compute length-offset, clamp limit to remaining length,
   then add. Direct helper cases cover empty, ordinary/exact/past end, small
   arrays with MaxInt limit/nonzero offset and near-MaxInt length without any
   huge allocation. No new arbitrary cap or invalid-input policy.
2. Author real public Store.Query regressions for primary and daily modes in
   an encoded temporary path with three rows across current/historical days.
   Verify two actual daily files and merged fixture sorting, baseline count
   and complete Alert values. Direct queries exercise MaxInt limits at offsets
   0/1/2, filtered large-limit nonzero-offset queries, exact/past/max-int offset,
   normal limit, default limit and negative offset. Check hardcoded expected
   selected indices/full copied baseline Alert values, total count, healthy
   state, Count and unchanged full result after all queries. No fake Query,
   post-hoc pointer observation, recovery or skipped panic regression.
3. Complete pinned Go 1.26.8 alert/API compile-only chain; Go parse-format,
   docs-check, task JSON, full unique roadmap row/Definition multisets/full
   contract, prior Definitions/R90-75/history/horizon/links/fences/six-path/diff/
   anchored sensitive/source review. All behavioral/race/CLI/full-suite/scanner/
   knowledge/acceptance execution **not run; delegated by user**. Static and
   compile evidence is not runtime/SQL preservation evidence.
4. Deliver one feature plus one three-path docs-only closure, exact full SHA
   push/fresh-fetch refs, Vault scope/note/index/MOC and 14 stable sections;
   archive the entire previous current-section prose, retain topic tails and
   332 immutable hashes, verify identical-range all-hash replay. Refresh queue
   without starting another increment or creating self-reference closures.

## Non-goals and stop conditions

No API/router/pagination/filter/SQL/schema/lifecycle/recovery/writer changes,
no reconciliation of existing negative-limit modes or count overflow, no data
migration, dependency/toolchain changes, test execution, IPv6 or publication.
Stop for competing changes, ambiguous validation/Git/Vault, new product decision,
private input or external authority. Only this increment per trigger. Existing
skill workflow suffices; do not invent a skill edit merely to narrate delivery.


## Implementation and validation checkpoint

Runtime is exactly one sliceBounds transform: clamp positive normalized limit to
length-offset before adding. All other tracked engine files and surrounding
query normalization/filter/sort/count/SQL/HTTP/lifecycle source match baseline.
Two direct functions author 11 helper bounds and 10 public real SQLite Query
cases per primary/daily mode. Source reaches two actual shard files in an
encoded directory, copied complete Alert baseline values, large nonzero-offset
limits, filters/count/order/health/default/negative-offset/past-end and final
logical row preservation. No fake Query/recover/skip substitutes this boundary.
Pinned Go 1.26.8 full alert/API compile-only chain passed on final source; output
binaries outside repository never invoked. Exact transform/direct-boundary/
Go-format/docs/168 JSON/154 full unique roadmap/full contracts/prior Definitions/
R90-75/history/horizon/links/fences/six-path/diff/sensitive checks pass.
All execution suites remain not run; delegated by user. No runtime/SQL outcome
is claimed. No validation, compile or implementation failure occurred. Prior
Definition separators were reused, avoiding historical-section formatting
changes. Existing skills suffice; no generic edit warranted.
379-file Vault baseline is unchanged. Retain entire prior current-section prose
under explicit historical heading and original topic tails/332 immutable hashes
when refreshing current authority. Exact feature and one three-path docs-only
closure delivery remain; no subsequent increment starts here.


## Delivery and queue closeout

Feature `da7b5cc8ecf33ce60ad4a7a28a63ea2ab5aa6a6a` contains exactly six planned paths. Isolated
`fix/r90-150-shard-query-bounds` fast-forwarded freshly verified main;
push/fresh-fetch verified clean HEAD/origin/main/FETCH_HEAD at that SHA.
Exact range `8cd3b430384e3727d91ecb8e1783a4bb34998fc0..da7b5cc8ecf33ce60ad4a7a28a63ea2ab5aa6a6a` six-path generated scope, iteration note
`04-开发迭代记录/2026-10-03-da7b5cc8ec-CI知识同步.md`, full index and MOC are verified. Fourteen stable
current sections are reconciled; their entire prior R90-149 current prose is
retained under explicit historical headings and original topic tails remain.
All 332 baseline immutable iteration hashes are unchanged; identical replay
preserves the 380-file Markdown snapshot, JSON SHA-256 `43fc60de4e70798afe2904e0a40ca7592a4e88e3c2507d4365a68d774e26694a`.
Unique existing local sibling Vault selected explicitly; no extra empty Vault.

Acceptance matches the persisted plan: runtime changes only sliceBounds to
clamp limit to length-offset before adding. All other tracked engine files and
surrounding normalization/sort/filter/count/SQL/HTTP/lifecycle/recovery/writer
source match baseline. At/past-end remains empty, no arbitrary cap or new
negative-limit mode policy. Two authored direct functions contain 11 helper
bounds and 10 real public Store.Query cases per primary/daily mode in an encoded
path. Daily fixture requires two actual files. Full Alert values are copied
before boundary queries, then checked against hardcoded indices; large limits
at nonzero/filtered offsets, ordinary/default/negative-offset, empty cases,
count/order/health and final logical rows are directly asserted. No fake Query,
recover/skip or shared-pointer observation weakens the regression boundary.

Pinned Go 1.26.8 final alert/API complete compile-only chain and source/direct-
boundary/Go-format/docs/168 JSON/154 full unique roadmap/full contracts/all prior
Definitions/R90-75/history/horizon/links/fences/six-path/diff/sensitive checks pass.
No test binary or behavioral/race/CLI/full-suite/scanner/knowledge/acceptance
execution; all remain not run, delegated by user. No runtime/SQL/file-byte/race/
SLO pass is claimed. No implementation/validation/compile/delivery failure or
Vault topic loss occurred. Existing skills sufficient; no generic edit. Prior
heading separators preserved. The only planning deviation is empty ready-queue
repair within this source-grounded increment; current Oct 3–Dec 31 horizon and
completed history unchanged.

One three-path docs-only record closes the same increment: resolve full SHA
from Git, push/fresh-fetch and exact feature-tip..closure-tip Vault verification
before reporting, without another self-reference closure. Queue refresh: no
currently defined dependency-ready local item. R90-75 full independent
asynchronous departmental contract is unchanged and does not block development.
Next trigger verifies fetched closure/Vault, audits fresh core-code evidence/
queue and persists a separate eligible plan before editing. No subsequent
implementation starts here. IPv6/external publication require separate authority;
do not repeat R90-149/R90-150 delivery or R90-59 publication.
