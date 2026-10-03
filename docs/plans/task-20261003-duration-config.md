# R90-160: reject unrepresentable duration settings

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`88bb30dcac46385b3f354b77aaf636ef5c005f68`. R90-159 exact six-path feature/
three-path closure Git/Vault scope/note/index/MOC verified; generated identifiers
uniquely resolve through Git to full endpoints. 99-commit Sep 5–Oct 3 phase moves
from SLO tooling to core correctness; no new qualifying R90-75 evidence. R90-75
sole unfinished full independent asynchronous departmental contract and Oct 3–
Dec 31 horizon unchanged. Empty queue reconciled inside source-grounded duration
representability repair. Unique existing local Vault's 399 Markdown hashes/14
entire stable notes backed up outside Git; no AGENTS. Owning engine pinned Go
1.26.8 preflighted before compile-only checks.

## Scope, risk and authority

Six paths: `engine/internal/config/config.go`, new
`engine/internal/config/duration_bounds_test.go`, `docs/architecture.md`, this
plan, `docs/tasks/task-state-20261003-duration-config.json`, rolling roadmap.
Plan/state persisted before source/other docs edits. main converts aggregation
and health freshness integer seconds to time.Duration and multiplies by
time.Second; values outside [-9223372036, 9223372036] seconds cannot be represented.
At the first out-of-range values sign flips, incorrectly invoking downstream
nonpositive defaults or turning negative/default requests into positive windows.
Require representability for engine.alert_aggregation_window and
engine.health_freshness_limit_seconds during public config.Load validation.
Derive signed symmetric whole-second bounds from duration representation; cast
native int to int64 for portable comparisons. Keep defaults, representable
positive/zero/negative inputs, YAML/env behavior, other validators and downstream
conversions/fallback policies unchanged. No tighter operational upper bound.
Risk low: formerly overflowing numeric settings now fail startup with named
validation diagnostics; operators must correct them. Native int YAML parsing
bounds remain in force on 32-bit builds; no new target policy.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Compile/static evidence is no startup/runtime pass.

## Acceptance mapped to evidence

1. Exact config runtime diff: time import, whole-second bound constant and two
   named range checks only. Every other tracked engine source, defaults/main
   conversions/fallback/other validation/env/strict decoder stays byte-identical.
2. Public file-backed Load for each setting: accepted minimum whole seconds,
   -1/0/1/maximum retain exact integer values and complete config/default fields;
   duration round-trip and sign prove representable values. First positive and
   negative overflow, int64 maximum/minimum reject with nil config and exact
   field/bounds diagnostic. Native 32-bit out-of-int fixtures retain parse errors;
   no skipped or weakened assertions. Every input file remains byte-identical.
3. Omitted settings retain defaults 60/30 via public Load. Both overflow fields
   combined with invalid API port retain all three ordered validation diagnostics
   on native 64-bit; native 32-bit still fails parsing. No unrelated error loss.
4. Numeric environment expansion for both settings covers accepted 60 and first
   positive/negative overflow, with exact values or named error/nil config; native
   32-bit out-of-int errors remain parse errors. Input files remain byte-identical.
5. Pinned Go 1.26.8 config/owning CLI/alert/API complete compile-only chain;
   static exact source/direct boundaries/format/docs/178 JSON/164 complete unique
   roadmap pairs/prior Definitions/R90-75/history/horizon/links/fences/six paths/
   diff/sensitive. All test binaries/benchmarks unexecuted.
6. Feature plus one docs-only closure exact full-SHA commit/push/fetch/Vault scope/
   note/index/MOC/identifier resolution; 14 stable notes reconcile current authority
   with entire prior prose archived/topic tails/immutable hashes retained; identical
   replay; refresh queue and stop.

## Non-goals and stop conditions

No operational max-duration policy, positive-only rule, fallback/default/main
conversion changes, programmatic Store/API option validation, retention/busy
timeout/other numeric settings, schema/matching/metrics/dependency/toolchain/
suites/private inputs/IPv6/publication. Stop for competing edits, ambiguous static/
compile/Git/Vault, new private/product/external authority or second increment.
Existing skills cover representability and every rejection boundary; no redundant
update. Next selection only on a separate trigger.


## Implementation and validation checkpoint

Runtime diff adds time import, derived whole-second bound and two named signed
range checks only. Static arithmetic confirms [-9223372036, 9223372036] whole
seconds fit signed nanoseconds; first positive/negative values outside flip sign
under wrapping conversion. Every other tracked engine source/default/main
conversion/fallback/validator/env/strict-decoder byte-identical. Representable
negative/zero/positive values retained; formerly overflowing config now rejects
before startup with named bounds. No operational cap or target policy introduced.

Three direct external public Load regression functions authored/compiled only.
Both settings cover signed endpoints/-1/0/1, first signed overflow and int64
extremes; complete config including unrelated defaults/path remains equal for
accepted values, with duration round-trip/sign assertions. Overflow rejects nil
config with exact field/bounds diagnostics. Omitted defaults stay 60/30. Combined
both overflows plus invalid API port retain three exact ordered diagnostics.
Numeric env expansion covers accepted 60 and first signed overflows for both
fields. Native 32-bit out-of-int fixtures require parse errors rather than skips;
all Load input bytes compared before/after. No private validator/default access,
sleeps or weakened tests. No startup/runtime outcome inferred.

Pinned Go 1.26.8 config/owning CLI/alert/API complete compile-only chain passed;
binaries and benchmarks unexecuted. Exact source/direct boundaries/arithmetic/
Go-format/docs/178 JSON/164 complete unique roadmap pairs/prior Definitions/
R90-75/history/horizon/links/fences/six paths/diff/sensitive passed. No unresolved
validation failure or deviation; all behavioral/race/CLI/full-suite/scanner/
knowledge/traffic/acceptance execution **not run; delegated by user**. No startup/
SQLite/filesystem/runtime/SLO pass. 399 baseline Vault hashes unchanged. Existing
skills cover rejection boundaries/portable comparisons; no redundant update.
Feature plus one docs-only closure remains; no following increment/publication.
