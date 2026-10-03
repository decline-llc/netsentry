# R90-154: exclude nil entries from generated-alert counts

## Selection and baseline audit

Clean fetched main HEAD/origin/main/FETCH_HEAD: `d57c634ad4ba4bbc3d6966d97e5e319f8ad43294`.
R90-153 exact six-path feature and three-path closure Git/Vault scope/note/index/
MOC verified. Sep 5–Oct 3 phase covers 87 commits: supplied SLO evidence tools
then core correctness repairs. No qualifying new R90-75 acceptance; full
independent asynchronous contract and Oct 3–Dec 31 horizon unchanged. Empty local
ready queue restored inside this source-grounded metrics repair, without separate
audit-only delivery. Vault 387-file hash snapshot and complete 14 current stable
notes backed up outside Git. Owning engine module resolves pinned Go 1.26.8.

## Scope, risk and authority

Seven paths: `engine/internal/stats/stats.go`, new
`engine/internal/stats/nil_alerts_test.go`, new
`engine/internal/pipeline/nil_alert_metrics_test.go`, `docs/architecture.md`,
this plan, `docs/tasks/task-state-20261003-nil-alert-count.json`, rolling roadmap.
Persist plan/state before source/docs changes. Runtime only ObserveAlerts:
count entries passing existing nil skip in same severity loop, then add that
count under existing lock. Preserve early nil-receiver/empty-batch handling,
empty-severity low fallback, repeated-pointer entry counting and dynamic labels.
Current total uses len(alerts), while severity loop and Store normalization skip
nil; Worker can therefore report phantom generated alerts for successful mixed
or all-nil batches. No observed incident or SLO acceptance claim.
Risk medium: counter changes affect derived rates; metrics names/labels/schema
and existing Worker success/error/export ordering remain unchanged. Snapshot
atomics remain individually sampled; no general transactional snapshot promise.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Static/docs checks and compilation only; no
binary execution or runtime/SQL/race evidence inferred from compilation.

## Acceptance mapped to evidence

1. Exact ObserveAlerts-only transform; count only non-nil entries once in existing
   severity loop and publish total inside lock. Preserve all other tracked engine
   source, worker/renderer/API/lifecycle/store/recovery/schema behavior. Source diff.
2. Author direct public Stats.ObserveAlerts/Snapshot/RenderPrometheus table for
   nil/empty/all-nil/mixed/full-severity/default/dynamic/repeated-pointer inputs,
   exact hardcoded totals and full severity maps/metric lines/type, nil-receiver
   compatibility and unchanged input pointers/full Alert values. Direct repeated
   observations accumulate correctly. No separate validation or dedup policy.
3. Author concurrent public observations with synchronized start and join before
   reading independently sampled counters; assert explicit totals/severity sum,
   map values and metric lines after quiescence. No live aggregate invariant claim.
4. Author actual public Worker.Run with real primary SQLite Store at encoded path
   for mixed and all-nil batches: query row identities/count/order plus exact
   metrics severity/total, packet completion/write/error counters, packet values.
   Separate Worker failure/export cases with existing injected writer/observer
   fixtures confirm unchanged publication gates and completion/error counters.
   Synthetic input; real-store assertions authored only, no SQL/durability pass.
5. Final Go 1.26.8 stats/pipeline/API/alert complete compile-only chain;
   source/direct-boundary/Go-format/docs/JSON/unique complete roadmap/prior
   Definitions/R90-75/history/horizon/links/fences/seven-path/diff/sensitive review.
6. Feature and one docs-only closure: commit/push/fresh fetch full exact refs,
   exact Vault note/index/MOC, affected stable current prose reconciled with prior
   substantive prose/topic tails/immutable hashes retained; identical replay;
   refresh queue and stop without following implementation.

## Non-goals and stop conditions

No counter names/labels/API schemas or rate formula, worker/matcher/suppressor/
redactor/export success semantics, storage/aggregation/SQL/lifecycle, general
snapshot synchronization, input validation/dedup, dependencies/toolchain, suites,
private inputs, IPv6 or publication changes. Stop for ambiguous static/compile/
Git/Vault, competing work, new product/private/external authority, or a second
increment. Existing skills suffice; no skill edit solely to narrate delivery.


## Implementation and validation checkpoint

Runtime changes only ObserveAlerts: remove total=len(alerts), count entries after
existing nil skip within severity loop, then add that count inside existing lock.
Nil-receiver/empty-batch guards, low fallback, repeated-entry/dynamic labels and
all other tracked engine/Worker/renderer/API/store/export/lifecycle/rate/schema
source remain unchanged. No extra traversal or allocation; no performance claim.

Four direct functions authored: nine public Stats/Snapshot/renderer cases,
observed twice, exact totals/full severity maps/one exact metric line per label
and type, nil receiver and original input pointers/full values. Four concurrent
writers each perform 100 mixed and all-nil observations, synchronized start and
join before asserting 800 total/400 high/400 low and matching metric lines.
Actual Worker.Run with encoded primary SQLite store covers mixed two-row and
all-nil zero-row success, fixed row IDs/order/severity/timestamp/count, health,
packet values and exact processing/completion/write/error/severity counters.
Four injected writer/export cases retain publication/completion gates and original
writer batch pointers. No fake storage replaces real SQL normal-path evidence;
failure fixtures explicitly remain injected boundary checks. All unexecuted.

Pinned Go 1.26.8 final stats/pipeline/API/alert complete compile-only chain passed;
external binaries unexecuted. Source/direct-boundary/Go-format/docs/172 JSON/
158 unique complete roadmap/prior Definitions/R90-75/history/horizon/links/fences/
seven-path/diff/sensitive review passed. All behavioral/race/CLI/full-suite/scanner/
knowledge/traffic/acceptance suites **not run; delegated by user**. No runtime/
SQL/durability/race/SLO pass inferred. No source, compilation or validation
failure; 387-file Vault baseline unchanged. Existing skills suffice; no generic
outcome-only skill edit. Feature and one docs-only closure remain, without a
following implementation or publication action.
