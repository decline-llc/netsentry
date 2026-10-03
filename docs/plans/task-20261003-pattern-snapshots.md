# R90-157: isolate Aho-Corasick pattern snapshots

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`1bafc66e775d0bb22c6e22fa45b4449dfcb9af22`. R90-156 exact seven-path feature
and three-path closure Git/Vault note/scope/index/MOC verified, generated short
identifiers uniquely resolve to full endpoints. Sep 5–Oct 3 phase: 93 commits,
SLO tooling then core correctness repairs, no new qualifying R90-75 acceptance.
R90-75 sole unfinished departmental item; full independent asynchronous contract
and Oct 3–Dec 31 horizon unchanged. Empty local ready queue reconciled inside
source-grounded matcher ownership repair. Vault 393 Markdown hashes/14 complete
stable notes backed up outside Git. Pinned owning engine Go 1.26.8 preflighted.

## Scope, risk and authority

Eight paths: `engine/internal/rule/ahocorasick/ahocorasick.go`, new
`engine/internal/rule/ahocorasick/pattern_snapshot_test.go`,
`engine/internal/rule/engine.go`, new `engine/internal/rule/candidate_set_test.go`,
`docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-pattern-snapshots.json`, rolling roadmap.
Persist plan/state before source/docs edits. Patterns currently returns the owned
slice, allowing metadata edits that disagree with the already-compiled trie.
Return a fresh copy retaining normalized order, duplicates and non-nil empty
shape. NewMatcher already owns its input. Engine currently reads Patterns per
hit to store keyword text that it never consumes; use a rule-ID set instead so
copying exported metadata does not add a full pattern copy per engine hit.
Risk medium: public getter allocates a copy for nonempty patterns; no measured
allocation/throughput claim. Candidate membership, index guards, per-rule keyword
selection, normalization/trie/failure links/match ordering and engine publication
remain unchanged. No nil matcher/receiver or new empty-pattern match semantics.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Static checks and compilation only; no binary run.

## Acceptance mapped to evidence

1. Exact source diff: Patterns defensive make/copy return and comment only;
   Engine.Match candidate map becomes rule-ID set without Patterns read. Preserve
   every other tracked engine source and Match trie/normalization/index guards/
   per-rule selection/priority/early exit. No getter copy in engine packet path.
2. Direct public NewMatcher/Patterns/Match table: sensitive/insensitive normalization,
   duplicate/failure-suffix indices, constructor input and successive getter
   snapshots independently owned; edit returned entries, reslice and append;
   exact expected pattern text/order and fixed hit indices remain unchanged.
   Nil and empty constructor input retain non-nil zero-length getter and no hits.
3. Public concurrent getter callers edit only their own returned slices while
   reading Patterns and Match; synchronized start, joined completion before fixed
   metadata/hit assertions. No internal fields or timing sleeps; no race-pass claim.
4. Public Engine.Reload/Match fixed cases for duplicate keywords, shared rule
   keywords, original per-rule keyword selection versus first AC hit, mixed case,
   hit/miss windows, disabled rules, protocol/port rejection and critical early
   exit. Exact full alert fields/order/count and unchanged packet fixture; actual
   compiled engine, no injected matcher. No broadened matching policy.
5. Final pinned Go 1.26.8 Aho-Corasick/rule/API/pipeline complete compile-only chain;
   static exact source/direct boundaries/format/docs/175 JSON/161 complete unique
   roadmap pairs/prior Definitions/R90-75/history/horizon/links/fences/eight paths/
   diff/sensitive; existing benchmarks compile, execution remains delegated.
6. Feature and one docs-only closure commit/push/fresh fetch/full exact range Vault
   scope/note/index/MOC and generated identifier resolution; 14 current stable
   records reconciled with entire prior prose archived/topic tails/immutable
   hashes retained; identical replay; refresh queue without starting another item.

## Non-goals and stop conditions

No trie/Match algorithm/normalization/empty-pattern semantics/nil receiver,
per-rule filters/priority/early exit/reload/ownership, API schemas/metrics/storage,
benchmark policy/performance guarantees, dependencies/toolchain/suites/private
inputs/IPv6/publication. Stop for competing edits, ambiguous static/compile/Git/
Vault, new product/private/external authority or second increment. Existing skills
suffice; no generic outcome-only edit. Next selection on a separate trigger only.


## Implementation and validation checkpoint

Runtime only Patterns getter make/copy plus Engine candidate map value changed
from unused keyword string to rule-ID set. Final review retained existing
first-hit guard and index checks. No getter in packet path; trie/normalization/
Match/per-rule keyword/filter/priority/early exit/reload/publication and every
other tracked engine source preserved exactly. Getter keeps normalized order,
duplicates/non-nil empty result; public nonempty copies are deliberate, without
measured throughput/allocation, new empty-pattern or nil-receiver guarantees.

Three direct public regression functions authored. Sensitive/insensitive getter
cases verify fixed duplicate/suffix indices, constructor input independence,
successive snapshot/entry/reslice/append edits and nil/empty no-hit shape. Four
start-synchronized callers each perform 100 owned snapshot edits and fixed
Match checks; joined callers precede final metadata/hit assertions. Eight actual
Engine.Reload/Match cases cover duplicate/shared/original-keyword selection,
mixed case, window hit/miss, disabled shared keyword, protocol/port rejection,
critical early exit; exact full alert count/fields/order and packet values.
No internal-field writes, injected matcher, sleeps or weakened skips. All
assertions authored/compiled only; no runtime/race/SLO outcome inferred.

Pinned Go 1.26.8 final Aho-Corasick/rule/API/pipeline complete compile-only chain
passed after duplicate-guard review; binaries/benchmarks unexecuted. Exact source/
direct boundaries/Go-format/docs/175 JSON/161 complete unique roadmap pairs/prior
Definitions/R90-75/history/horizon/links/fences/eight paths/diff/sensitive passed.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. No unresolved implementation/compile/static
ambiguity; 393 baseline Vault hashes unchanged. Existing skills cover this
boundary and generated identifiers; no redundant generic edit. Feature plus one
docs-only closure remains; no following implementation or publication started.


## Delivery and queue closeout

Feature `bb98065a2a232459c3e9cfd19787cda803b3654e` contains exactly eight planned paths. Isolated
fix/r90-157-pattern-snapshots fast-forwarded freshly verified main; push/fresh
fetch verified clean HEAD/origin/main/FETCH_HEAD at that full SHA. Exact range
`1bafc66e775d0bb22c6e22fa45b4449dfcb9af22..bb98065a2a232459c3e9cfd19787cda803b3654e` note `04-开发迭代记录/2026-10-03-bb98065a2a-CI知识同步.md`, eight-path scope, full index and MOC verified.
Generated ten-character identifiers uniquely resolve through Git to recorded
full endpoints. Fourteen current stable notes reconciled; entire previous current
substantive prose archived under explicit R90-156 historical headings. Original
topic tails and all 346 baseline immutable iteration hashes retained; only
bounded documented generated MOC entries refreshed. Identical replay preserves
394 Markdown files; snapshot JSON SHA-256
`d7c369b9f0d1c1de067a33ef957148fd9f60ad25818e55f0c1cab86018b4653d`. Existing unique sibling local Vault selected
explicitly; no second empty or remote Vault.

Acceptance matches persisted plan: Patterns getter make/copy and Engine candidate
value changed from unused keyword text to rule-ID set. Existing duplicate-hit
and index guards retained; no getter/copy per engine hit. Trie/normalization/
matching/per-rule keyword/filter/priority/early-exit/reload/publication and all
other tracked engine source preserved exactly. Getter retains normalized order,
duplicates/non-nil empty result; public nonempty copy deliberate, without measured
throughput/allocation guarantee, nil-receiver or empty-pattern policy expansion.

Three authored direct public functions reach promised boundaries: sensitive/
insensitive getter normalization, duplicate/suffix fixed indices, constructor
input edits and independently mutable successive snapshots/entry/reslice/append,
empty no-hit shape; four synchronized callers each 100 snapshots/edits/fixed
hits, joined before final metadata/match assertions. Eight actual Engine.Reload/
Match cases cover duplicate/shared/original-keyword-versus-first-hit selection,
mixed case/windows/disabled/protocol-port/critical early exit, exact full alerts/
count/order and packet preservation. No private fields, injected matcher, sleeps
or weakened skip paths. Assertions authored/compiled only; no runtime/race/SLO
outcome or throughput evidence inferred.

Final pinned Go 1.26.8 Aho-Corasick/rule/API/pipeline complete compile-only chain
passed after duplicate-guard review; binaries/benchmarks unexecuted. Exact source/
direct boundaries/Go-format/docs/175 JSON/161 complete unique roadmap pairs/prior
Definitions/R90-75/history/horizon/links/fences/eight paths/diff/sensitive passed.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. No unresolved implementation/compile/static/Git/
Vault ambiguity or topic loss. Existing skills suffice; no redundant generic
edit. Planning deviation: empty queue restored inside source-grounded ownership
repair.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; push/fresh-fetch and verify exact feature..closure Vault scope/
note/index/MOC/stable prose before reporting. No self-reference follow-up closure.
Refreshed queue has no defined local dependency-ready item. R90-75 full independent
asynchronous departmental contract and Oct 3–Dec 31 horizon unchanged. Next trigger
verifies fetched closure/Vault and audits fresh code/queue before a separate
eligible plan. No following increment started; do not repeat R90-156/R90-157
delivery or R90-59 publication. IPv6/external publication needs separate authority.
