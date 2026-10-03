# R90-155: reject negative duration observations

## Selection and baseline audit

Clean freshly fetched main HEAD/origin/main/FETCH_HEAD:
`f55595c21d55e3933ae2f9048fe48a8d44297705`. R90-154 exact seven-path feature
and three-path closure Git/Vault note/index/MOC verified. Sep 5–Oct 3 phase:
89 commits of SLO evidence tooling and core correctness repairs; no new qualifying
R90-75 acceptance. Full departmental asynchronous contract and Oct 3–Dec 31
horizon unchanged. Empty local ready queue reconciled inside this source-grounded
repair. Vault 389 Markdown hashes and 14 full stable notes backed up outside Git.
Pinned owning engine module Go 1.26.8 preflighted; no new tool or dependency.

## Scope, risk and authority

Six paths: `engine/internal/stats/stats.go`, new
`engine/internal/stats/negative_duration_test.go`, `docs/architecture.md`, this
plan, `docs/tasks/task-state-20261003-negative-durations.json`, rolling roadmap.
Persist plan/state before source/docs edits. Public match/write duration observers
currently cast negative time.Duration to uint64, generating huge unsigned sums
while incrementing operation counts and every finite histogram bucket. Reject
negative samples before any counter update, following existing negative queue
observation policy. Zero/positive samples remain accepted. No observed incident.
Risk medium: malformed samples cease contributing to exported counters. Preserve
metric names/types/labels, bucket boundaries, renderer, all other tracked engine
source and independently sampled snapshot semantics. No general overflow policy
for accumulated positive durations, no new runtime telemetry/error channel.
Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**; compilation/static checks only, no binary invoked.

## Acceptance mapped to evidence

1. Source diff: only add d < 0 to existing nil guards of both public duration
   observers. Reject before count/sum/buckets; nil receiver still safe. Preserve
   zero/positive observation logic and every other tracked engine file exactly.
2. Direct public Stats/Snapshot/RenderPrometheus regressions for each observer:
   negative-only on fresh instance, zero then negative, seeded positive then
   negative, signed minimum duration, -1 ns and -1 second; exact full snapshot
   unchanged after rejection, exact output unchanged and every histogram line.
3. Positive boundary regressions for each observer: zero, 1 ns, every existing
   finite bucket boundary and +1 ns immediately above, plus > largest bucket.
   Hardcoded independent boundaries and sums/counts; exact full bucket list and
   metric count/sum/total/+Inf lines. Isolation of the other observer and unrelated
   counters. Nil receivers with negative/zero/positive samples return safely.
4. Concurrent public observations of both APIs: start synchronization and joined
   writers before explicit hardcoded count/sum/all-bucket/renderer assertions;
   include rejected negative and accepted zero/positive samples. No live aggregate
   transactional or race-pass claim.
5. Final pinned Go 1.26.8 stats/pipeline/API complete compile-only chain;
   source/direct-boundary/Go-format/docs/173 JSON/159 complete unique roadmap
   rows/Definitions/prior history/R90-75/horizon/links/fences/six paths/diff/
   sensitive review. Authored assertions remain unexecuted.
6. Feature plus one docs-only closure: commit/push/fresh fetch exact refs; full
   exact-range Vault note/index/MOC and scope, reconcile 14 stable current notes
   archiving entire prior current prose and preserving topic tails/immutable
   hashes; identical replay; refresh queue and stop before another increment.

## Non-goals and stop conditions

No API/metric schema/labels/buckets/rate formula, Worker/matcher/store behavior,
positive accumulation overflow or snapshot transaction policy, dependencies,
toolchain, suites, private input, IPv6 or external publication changes. Stop for
ambiguous static/compile/Git/Vault evidence, competing work, new product/private/
external authority or a second increment. Existing skills suffice; no outcome-only
skill edit. New increment selected only on a separate subsequent trigger.


## Implementation and validation checkpoint

Runtime diff adds d < 0 only to both existing nil guards. Negative samples return
before count/sum/bucket updates; zero/positive/nil-receiver semantics, histogram
bounds/renderer and every other tracked engine file preserved exactly.
Three direct public regression functions authored: both observers with fresh/
zero/250 ms seeds reject -1 ns, -1 second and signed minimum; full snapshot and
entire metrics text unchanged, exact histogram/counter lines, nil receiver.
58 accepted boundary subcases cover zero/1 ns/each of 13 exact bounds and +1 ns/
above largest bound; complete snapshot checks isolate other and unrelated counters.
Four start-synchronized writers each perform 100 rounds on both APIs including
negative/zero/250 ms. Joined writers precede full snapshot/renderer assertions:
800 counts/100-second sums per observer; first eight finite buckets 400, final
five 800. No live transactional snapshot or race-pass claim. All unexecuted.

Pinned Go 1.26.8 stats/pipeline/API complete compile-only chain passed; external
binaries unexecuted. Static two-guard transform/all other tracked engine source/
direct boundaries/Go format/docs/173 JSON/159 complete unique roadmap pairs/prior
Definitions/R90-75/history/horizon/links/fences/six paths/diff/sensitive passed.
Initial historical preservation check caught one extra blank line at new
Definition insertion; exact previous whitespace restored, complete static/docs
chain rerun successfully. No behavior/compile failure; validation deviation
resolved. All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
execution **not run; delegated by user**. Positive cumulative overflow outside
scope; no runtime/SLO outcome inferred. 389 baseline Vault hashes unchanged.
Existing skills cover this boundary; no redundant generic edit. Feature plus one
docs-only closure remains, without another increment or publication action.


## Delivery and queue closeout

Feature `5c28c131c6432136e8453da169ffee7add9c4299` contains exactly six planned paths. Isolated
fix/r90-155-negative-durations fast-forwarded freshly verified main; push/fresh
fetch verified clean HEAD/origin/main/FETCH_HEAD at that full SHA. Exact range
`f55595c21d55e3933ae2f9048fe48a8d44297705..5c28c131c6432136e8453da169ffee7add9c4299` note `04-开发迭代记录/2026-10-03-5c28c131c6-CI知识同步.md`, six-path scope, full index and MOC verified.
Fourteen current stable notes reconciled; entire previous substantive current
prose archived under explicit R90-154 historical headings. Original topic tails
and all 342 baseline immutable iteration hashes retained; documented generated
MOC entries refreshed. Identical exact-range replay preserves 390 Markdown files;
snapshot JSON SHA-256 `3efd82adca5630392ece79e45fc1238604879788844f70957edc48ebe8db6ef4`. Existing unique sibling local Vault
selected explicitly; no second empty or remote Vault.

Acceptance matches persisted plan: runtime only d < 0 in both existing nil
observer guards, before count/sum/bucket changes. Zero/positive/nil semantics,
bounds/renderer and every other tracked engine file remain exactly unchanged.
Three authored direct public functions cover both observers with fresh/zero/
250 ms seeds and -1 ns/-1 second/signed minimum rejection, whole snapshot and
exposition preservation, exact full metric/bucket lines and nil receivers.
58 accepted boundary subcases cover zero/1 ns/all 13 finite bounds/+1 ns/above
largest; full snapshot isolates other/unrelated counters. Four synchronized
writers join before both observer counts 800/sums 100 seconds and full bucket/
renderer checks (first eight finite buckets 400, last five 800). No fixture
rewrites internal atomics or replaces the public snapshot/renderer boundary.
All assertions authored/compiled only; no runtime/race/SLO acceptance inferred.

Pinned Go 1.26.8 stats/pipeline/API complete compile-only chain and exact source/
direct boundaries/format/docs/173 JSON/159 complete unique roadmap pairs/prior
Definitions/R90-75/history/horizon/links/fences/six paths/diff/sensitive passed.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**; binaries unexecuted. Initial historical whitespace
check caught one insertion blank line; restored exact prior section, reran complete
static/docs chain successfully. No unresolved validation ambiguity or topic loss.
Positive accumulated sum overflow and transactional live snapshots outside scope.
Existing skills sufficient; no redundant generic edit. Planning deviation: empty
ready queue reconciled inside source-grounded signed-duration counter repair.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; push/fresh-fetch and verify exact feature..closure Vault scope/
note/index/MOC/stable prose before reporting. No self-reference follow-up closure.
Refreshed queue has no defined local dependency-ready item. R90-75 full independent
asynchronous departmental contract and Oct 3–Dec 31 horizon remain unchanged.
Next trigger verifies fetched closure/Vault and audits fresh code/queue before
persisting a separate eligible plan. No following increment started; do not repeat
R90-154/R90-155 delivery or R90-59 publication. IPv6/external publication needs
separate authority.
