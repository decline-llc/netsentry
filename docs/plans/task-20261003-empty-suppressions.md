# R90-163: reject enabled suppressions with no compiled prefixes

## Selection and baseline audit

Clean main was fetched and HEAD/origin/main/FETCH_HEAD agree at
`df55f0501a5bc53334aae9ce14543b56a42d0728`. R90-162 feature and single
three-path closure, exact local Vault notes/index/MOC and fourteen current stable
notes are verified. The Sep 5–Oct 3 history has 105 commits: blocked-queue
reconciliation, departmental SLO tooling, reviewed toolchain/input boundaries,
then core correctness repairs. Implementation delivery does not establish the
department's unexecuted runtime or R90-75 acceptance results. No missing prior
delivery or new qualifying SLO evidence was found. All 166 roadmap rows and
Definitions match without duplicates; the sole unfinished item is independent
R90-75 departmental acceptance. Its full contract and Oct 3–Dec 31 horizon stay
unchanged. Restore one local ready increment from the source gap below; no
separate queue. No AGENTS or pre-existing edits; owning engine Go 1.26.8 is
available. Plan/state precede source and other documentation edits.

## Scope, risk and authority

Six intended paths: `engine/internal/alert/suppressor.go`, new
`engine/internal/alert/empty_suppressions_test.go`, `docs/architecture.md`, this
plan, `docs/tasks/task-state-20261003-empty-suppressions.json`, rolling roadmap.
The structural validator checks raw CIDR list lengths; compilePrefixes skips
exact empty strings. An enabled empty-only list can therefore publish a filter
with no source, destination or any prefixes. Add a post-compilation guard in
NewSuppressor before append, reusing `suppression %q must include at least one
CIDR`. Risk low: previously accepted enabled empty-only lists now reject.
Preserve exact empty elements within valid lists, disabled-rule skipping, nil/
empty rule sets, prefix parsing and masking, rule-ID scoping, parse-error order,
and transactional manager publication/persistence. File load/save remain
structurally validating APIs; actual compilation rejects empty filters.

Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution is
**not run; delegated by user**, under the standing Sep 25 instruction. Static
and compile-only checks cannot prove runtime or SLO outcomes. No dependencies,
toolchain, external publication or private inputs are authorized by this repair.

## Acceptance mapped to evidence

1. Source diff adds only the compiled src/dst/any emptiness guard after all three
   parsers and before publication. Static comparison proves all other tracked
   engine source unchanged and prior diagnostics/order retained.
2. Direct public NewSuppressor and manager-constructor regressions cover nil/
   empty/empty-only source, destination, any and combined lists, exact diagnostic,
   nil result and input preservation. Disabled invalid/empty lists and empty
   rule sets remain accepted. Invalid nonempty prefixes keep existing diagnostics.
3. Public positive controls cover mixed empty/exact/CIDR lists, each direction,
   masking, rule-ID scoping, disabled filters, negative alerts and unchanged
   caller slices/alerts. Preserve existing IPv6 behavior without expanding scope.
4. Actual file-backed manager Add/Update/Reload regressions reject empty-only
   candidates before publication, preserve prior List/filter and complete file
   bytes, and permit valid retry. Reload uses a real canonical JSON file and
   actual LoadSuppressionsFromFile. Loader/serializer themselves gain no compiled
   validation guarantee. Include a directory containing spaces.
5. Pinned Go 1.26.8 alert/API/pipeline compile-only chain; binaries unexecuted.
   Static format/docs/JSON/complete roadmap multisets/prior Definitions/R90-75/
   history/horizon/local links/fences/intended paths/diff/sensitive review.
6. Focused feature and one docs-only closure: full-SHA push, fresh fetch and exact
   Vault ranges/note/index/MOC; reconcile fourteen current stable notes, preserve
   substantive prior prose and immutable notes, replay exact ranges to verify
   preservation; refresh the queue and stop after this increment.

## Non-goals and stop conditions

No whitespace trimming, rejection of empty elements in a valid list, changed
disabled validation, rule-ID normalization, new IP-family policy, parser/save/API
runtime or persistence algorithm changes, live traffic, runtime/race/performance/
SLO guarantee, tag or publication. Stop for competing edits, ambiguous static/
compile/Git/Vault evidence or new product/private/external authority. Existing
skills cover compiled emptiness and preservation; update only for a repeatable
new lesson. The next increment requires another trigger.


## Implementation and validation checkpoint

The only runtime change is a three-line compiled src/dst/any emptiness guard after
all parsers and before append. All other 68 tracked engine paths are unchanged.
Disabled skipping, mixed lists, prefix masking/scoping, existing IP-family behavior
and earlier parse/structural diagnostics remain. Five external public regression
functions are authored: nine rejection shapes across three constructors with nil
results/input/file preservation; six empty/disabled controls; twelve direction/IP-
family matching controls plus unscoped and mixed exact/CIDR controls; nine ordered
invalid/whitespace diagnostics; twenty-seven actual file-backed Add/Update/Reload
rejection/preservation/valid-retry cases. Twenty-one nonempty empty-entry mutation
cases reach the new compiled guard; six nil/empty raw-list cases retain the earlier
structural path. Reload's two raw-empty cases preserve the loader error wrapper.
Static review corrected those two assertions before compilation; no executed test
failure or ambiguous result occurred. No private seams, sleeps, panic swallowing
or test skips. Regressions remain unexecuted.

Pinned Go 1.26.8 alert/API/pipeline complete compile-only chain passed; binaries
unexecuted. Static exact source/format/docs/181 task JSON/167 unique roadmap row
and Definition pairs/all 166 prior Definitions/R90-75/history/horizon/local links/
fences/six paths/diff/sensitive review passed. Baseline Vault hashes are unchanged.
Existing skills already cover direct boundaries and preservation; no redundant
skill update. Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance
execution **not run; delegated by user**. Feature and one docs-only delivery
closure remain; no following increment started.

Initial sensitive scanning included unchanged historical prose and matched a prior
Vault path; manual review confirmed it was outside this increment. The corrected
review checks changed additions and new files; the complete static chain is rerun.


## Delivery and queue closeout

Feature `6f3781a36b5997f2c11a35e5d7f2ee379627c080` contains exactly the six intended paths. The
isolated branch fast-forwarded freshly verified main; push/fresh-fetch proved
clean HEAD/origin/main/FETCH_HEAD equality. Exact full-SHA range
`df55f0501a5bc53334aae9ce14543b56a42d0728..6f3781a36b5997f2c11a35e5d7f2ee379627c080` is synchronized to
`04-开发迭代记录/2026-10-03-6f3781a36b-CI知识同步.md`. Note scope, full index/MOC and short
identifiers uniquely resolved through Git are verified. Fourteen current stable
notes are reconciled; complete prior current prose is archived and historical/
topic tails are preserved. One current SLO topic sentence is corrected from
undeveloped collection adaptation to delivered native ingress/lifecycle/collector/
reconstruction tooling and still-pending departmental qualifying measurements.
All 358 prior immutable notes remain unchanged. Identical range replay preserves
406 Markdown hashes; snapshot JSON SHA-256
`453f82b968015f6556c8e252aed3e74313fd4c905c2f63b876da88099feec6ee`. The existing unique
sibling local Vault was passed explicitly; no second or remote Vault was created.

Acceptance matches the persisted plan: three-line guard after all parsers and
before append; 68 other tracked engine paths unchanged. Public constructors
cover nine rejection shapes and six empty/disabled controls; twelve direction/
IPv4/IPv6 matching controls plus unscoped and exact/CIDR-in-one-list checks; nine
ordered parse/whitespace diagnostics. Twenty-seven real file-backed mutation
cases preserve List/filter/whole bytes and allow valid retry; 21 nonempty empty-
entry cases reach the new compiled guard, six retain structural validation. Two
raw-empty Reload cases retain the exact loader wrapper. All regression assertions
are authored/compiled only, with no runtime/file/race/SLO pass implied. No loader/
save validation expansion, whitespace normalization or disabled/IP-family policy
change. Review corrected the two wrapped-error assertions before compilation.

Pinned Go 1.26.8 alert/API/pipeline compile-only chain and complete static review
passed. The initial sensitive scan matched only an unchanged historical path;
manual review, changed-additions/new-files scanning and complete clean static
rerun resolved it. Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/
acceptance execution **not run; delegated by user**. No unresolved validation
ambiguity. Existing skills cover the lessons; no redundant skill change.

This single three-path docs-only delivery record closes the same increment.
Resolve its full SHA from Git; push/fresh-fetch/exact feature..closure Vault
verification before reporting, without an additional self-reference closure.
After delivery, do not repeat feature/closure commit/push/sync. Queue refresh:
no defined local ready increment; R90-75 remains independent asynchronous
departmental acceptance with its full contract and Oct 3–Dec 31 horizon unchanged.
The next trigger verifies fetched closure/Vault, audits fresh code/queue and
persists a separate eligible plan. No following increment started.
