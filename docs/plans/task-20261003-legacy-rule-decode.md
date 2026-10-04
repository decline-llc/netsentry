# R90-165: preserve malformed legacy rule decode failures

## Baseline and selection audit

Clean fetched main HEAD/origin/main/FETCH_HEAD agrees at
`6f974967581efbfd571d53076a204a9a1ecfead6`. R90-164 seven-path feature and
three-path closure exact Git/Vault scopes, resolved identifiers, index/MOC and
current stable closure facts verified. Sep 5–Oct 3 history has 109 commits:
departmental SLO tooling, patched release/toolchain and core correctness. No
new qualifying R90-75 evidence or missing prior delivery found. All 168 unique
roadmap rows and Definitions match; only independent departmental R90-75 remains
unfinished. Its full contract and Oct 3–Dec 31 horizon stay unchanged.
No AGENTS or pre-existing edits; engine module resolves pinned Go 1.26.8.
409 Vault Markdown hashes and fourteen complete stable backups captured.

Empty local ready queue restored from source: parseRules first decodes wrapped
rawRule, which includes legacy MITRE string fields. On a type error it can fall
back to rulesFile/model.Rule, which ignores those fields. Malformed metadata can
therefore disappear and publish, or a null entry returned by that fallback can
reach applyRuleDefaults and panic. Retain the original wrapped decode error when
that weaker fallback would otherwise succeed. Plan/state precede behavior/docs.

## Scope, risk and authority

Seven paths: `engine/internal/rule/loader.go`, new
`engine/internal/rule/loader_legacy_decode_test.go`, new
`engine/internal/api/rule_reload_decode_test.go`, `docs/architecture.md`, this
plan, `docs/tasks/task-state-20261003-legacy-rule-decode.json`, rolling roadmap.
Capture the first decoder error and return it before a successful weaker fallback
can discard malformed recognized fields. Preserve valid canonical/legacy wrapped
and array normalization, defaults, unknown-field tolerance, existing empty/null
container and lone null-entry parsing, save/replace logic and semantic validation.
Risk low: malformed legacy fields previously silently ignored now reject at load.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**, under standing Sep 25 instruction. Static and
compile evidence cannot prove runtime/file/API/race outcomes. No new dependencies,
toolchain, private input or external publication authority.

## Acceptance mapped to evidence

1. Only first-error capture and pre-fallback return guard plus comment in loader;
   other tracked engine paths remain unchanged. Preserve diagnostics outside the
   newly rejected successful-fallback branch and nil/empty container shapes.
2. Direct public LoadFromFile regressions: three legacy MITRE string fields,
   number/bool/object/array wrong JSON kinds, canonical/legacy configs, with/without
   an earlier null entry. All 48 cases return nil result and a wrapped
   json.UnmarshalTypeError naming the affected field, preserve entire file bytes,
   and leave an already-loaded real Engine Rules/RuleCount/Match unchanged.
   Include directories containing spaces. No panic swallowing.
3. Positive canonical/legacy wrapped/array controls preserve defaults, normalized
   config/MITRE tuple, enabled matching, input bytes and unknown-field tolerance.
   Preserve six empty-container forms and existing lone-null normalization plus
   downstream semantic rejection; no new strict-container policy. Malformed syntax
   retains parse error and nil results without modifying files.
4. Public API Handler with actual LoadFromFile/real Engine: three malformed fields
   with/without null prefix return existing 500 INTERNAL_ERROR envelope, retain
   snapshot/matching/file bytes, then explicit valid-file retry returns 200 with
   one reloaded rule and correct legacy tuple/match. No fake rule manager/seams.
5. Pinned Go 1.26.8 rule/API/CLI/pipeline compile-only chain, binaries unexecuted;
   static source/format/docs/183 JSON/169 complete unique roadmap pairs/prior
   Definitions/R90-75/history/horizon/links/fences/seven paths/diff/sensitive review.
6. Seven-path feature and one docs-only closure, non-force push/fresh fetch,
   full-SHA exact Vault ranges/note/index/MOC, fourteen stable prose reconciliations,
   archived prior current prose/topic preservation, immutable note hashes and
   identical replay. Refresh future queue without starting another increment.

## Non-goals and stop conditions

No unknown/duplicate-field rejection, missing/null wrapper policy, direct lone-null
loader rejection, MITRE catalog/default/config precedence change, new semantic
validation at load/save, mutation/auth/API status policy, serialization/replacement
algorithm, runtime or panic/privacy/SLO claim, suites, private inputs, IPv6, release
or publication. Stop for competing edits, ambiguous static/compile/Git/Vault or
new product/private/external authority. Existing skills cover rejection conditions
and preservation; refine only for a repeatable new lesson. One increment only.


## R90-165 Implementation and Static/Compile Checkpoint (2026-10-03)

Runtime changes only first wrapped-error capture and a return guard with comment
inside successful simpler-model fallback. The fallback error has a distinct name
so it cannot shadow the retained original error. All other 70 tracked engine paths
unchanged. Valid normalization/defaults and permissive empty/null/unknown-field
parsing remain; no loader/save semantic validation or API runtime/status change.

Three public loader regression functions are authored: 48 wrong legacy-field/
JSON-kind/config/null-prefix cases assert original json.UnmarshalTypeError field,
string target/kind, nil result and full file-byte preservation; prior real Engine
Rules/RuleCount/Match retained. Four positive canonical/legacy wrapped/array cases
check default priority, entire normalized tuple, actual matching and original
bytes. Six empty-container shapes assert exact nil/non-nil forms, two lone-null
cases retain default normalization/downstream id diagnostic, three malformed
syntax cases retain parse rejection/bytes. Directories containing spaces included.
One public API regression has six real Handler/LoadFromFile/Engine rejection and
valid-repair retry cases. Canonical config makes the old weaker fallback candidate
otherwise compilable, while legacy metadata error cannot disappear. Existing 500
load envelope/request ID/content type/details, old snapshot/match/bad bytes and
200 one-rule retry with correct tuple/match/good bytes/input are asserted. Null
prefix reaches the formerly unsafe defaults branch. No private seams/fake rule
manager/sleeps/skips/panic swallowing; all assertions authored/compiled only.

Final complete pinned Go 1.26.8 rule/API/CLI/pipeline compile-only chain passed;
binaries unexecuted. Static source/70 other tracked engine paths/format/docs/183
JSON/169 unique roadmap pairs/168 prior Definitions/R90-75/history/horizon/links/
fences/seven paths/diff/sensitive review passed before staging. All behavioral/
race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution **not run;
delegated by user**. No runtime/file/API/race/SLO pass inferred.

Reusable decoder-fallback lesson added to local netsentry-next skill instruction
17: once input is identified as a supported format, a weaker compatibility decoder
must not discard a recognized-field failure; cover an input it would otherwise
accept. Markdown structure verified; separate from repository feature commit.
No next increment started; feature and single docs-only closure remain.
