# R90-156: count an unpublished rule snapshot as empty

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`f0ab2b5d1f6cbc33c52795ad687c19787922bc39`. R90-155 exact six-path feature
and three-path closure note/scope/index/MOC verified. Generated ten-character
identifiers uniquely resolve to supplied full ranges; prior extra audit assertion
about generated full-SHA rendering resolved against actual format. Sep 5–Oct 3:
91 commits of SLO tooling and core correctness repairs; no new qualifying R90-75
acceptance. R90-75 sole unfinished independent departmental item; its full
contract and Oct 3–Dec 31 horizon unchanged. Empty local ready queue reconciled
inside source-grounded core rule repair. 391 Markdown hashes and 14 complete
stable notes backed up outside Git. Owning engine module Go 1.26.8 preflighted.

## Scope, risk and authority

Seven paths: `engine/internal/rule/engine.go`, new
`engine/internal/rule/zero_count_test.go`, new
`engine/internal/api/zero_rule_count_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-zero-rule-count.json`, rolling roadmap.
Persist plan/state before source/docs edits. A non-nil zero-value Engine safely
handles Rules/Match/Reload, while RuleCount dereferences its nil initial state.
API health/metrics call RuleCount and inherit that panic. Match existing empty
snapshot policy: load once, return zero if nil, otherwise count published rules.
No observed incident; no typed nil Engine receiver guarantee. Risk low: only
unpublished state gains safe behavior; initialized state and atomic publication
retain existing behavior. No dependencies/toolchain or general API redesign.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Static checks/compilation only; binaries unexecuted.

## Acceptance mapped to evidence

1. Source diff: RuleCount only, single atomic load/nil guard/zero fallback, then
   unchanged length of published allByPriority. Preserve all other tracked engine
   source, immutable ownership, reload validation/order and initialized counting.
2. Direct public zero-value Engine RuleCount/Rules/Match before reload, valid
   Reload then exact count/rules/alert identity, invalid Reload with published
   state retained, empty Reload clearing state. Compare constructor-created
   empty engine; explicitly count enabled and disabled loaded rules as before.
   Fixed synthetic fixture/independent expected values; no nil-receiver claim.
3. API Handler with actual zero-value rule.Engine (no rule-manager mock): health,
   verbose health, metrics and rules listing before load, after valid load, after
   rejected reload and after empty reload. Exact status/content types, rule count,
   zero metrics/unrelated queue/storage values and rule JSON identity; existing
   fakeStore/fakeQueue isolate unrelated storage/queue, no SQL or wire claim.
4. Final pinned Go 1.26.8 rule/API/pipeline complete compile-only chain; static
   exact source/direct public boundaries/Go-format/docs/174 JSON/160 complete
   unique roadmap pairs/prior Definitions/R90-75/history/horizon/links/fences/
   seven paths/diff/sensitive. Authored assertions remain unexecuted.
5. Feature plus one docs-only closure: commit/push/fresh fetch full exact refs;
   exact full-range Vault scope/note/index/MOC with generated short identifiers
   resolved through Git; reconcile 14 stable current notes archiving entire prior
   current prose, preserving topic tails/immutable hashes; identical replay;
   refresh forward queue and stop without a following increment.

## Non-goals and stop conditions

No nil Engine receiver support, reload/match semantics, rule defaults/priority/
validation, API schemas/routes/auth/metrics labels, storage/queue, snapshot
synchronization, dependencies/toolchain, suites/private inputs/IPv6/publication.
Stop for competing edits, ambiguous static/compile/Git/Vault, new private/product/
external authority or a second increment. Existing skills cover recovery and
scope; no generic outcome-only edit. Next selection only on a separate trigger.


## Implementation and validation checkpoint

Runtime only RuleCount: load atomic snapshot once, return zero if unpublished,
otherwise retain length of allByPriority. Every other tracked engine file, Rules/
Match/Reload/immutable ownership/priority/validation/API/renderer/store source
preserved exactly; no typed nil receiver support or new publication semantics.
Two direct public regression functions authored. Engine begins with count call
before any Reload, compares constructor empty behavior, then loads two rules
(including disabled), verifies exact priority order and full expected alert,
rejects null-rule Reload while retaining count/rules/match, then clears via nil
Reload. Actual rule.Engine backs API Handler across unpublished/loaded/rejected/
cleared phases: all four health/verbose-health/metrics/rules endpoints assert
status/content type, expected JSON rule identity/count/shape and zero metrics/
queue/storage values. Existing fakeStore/fakeQueue isolate unrelated dependencies;
no rule mock, SQL or wire evidence claim. All assertions remain unexecuted.

Pinned Go 1.26.8 rule/API/pipeline complete compile-only chain passed, binaries
not invoked. Exact runtime/source/direct-boundary/Go-format/docs/174 JSON/160
complete unique roadmap pairs/prior Definitions/R90-75/history/horizon/links/
fences/seven paths/diff/sensitive review passed. No implementation/compilation/
static deviation. All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/
acceptance execution **not run; delegated by user**; no runtime/race/SLO outcome.
391 baseline Vault hashes unchanged. Separate local netsentry-next skill
refinement: resolve generated abbreviated metadata through Git before comparing
full recorded endpoints, while synchronization input still requires full SHAs.
Markdown structure checked; no repository feature path or authority expansion.
Feature plus one docs-only closure remains; no subsequent implementation started.
