# R90-175: reject an explicit suppression scope with no compiled rule IDs

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD baseline:
`e4cf2c78a316b22f9439604f88d52d7a26c0c911`. R90-174 feature and closure exact
Git/Vault ranges, note/index/MOC verified. The four-week phase has 129 commits,
all covered by iteration notes and the full index; 178 unique roadmap row and
Definition pairs agree. Captured 429 Vault Markdown hashes, 382 immutable
iteration hashes and complete stable-note backups. No missing delivery or
supplied independent R90-75 acceptance result. The local ready queue is empty.

Source inference: NewSuppressor skips empty rule-ID entries, and matchesRuleID
treats an empty compiled map as unrestricted. Thus an enabled suppression with
`rule_ids: [""]` and a valid IP range suppresses unrelated rule IDs. Select the
bounded fail-closed default: a nonempty list must compile at least one nonempty
ID. Nil/empty lists still explicitly mean all rules. This is not an executed
failure. Forecast remains Oct 4–Jan 1; no date eligibility gate.

## Scope, risk, non-goals and stop condition

Nine intended paths: engine/internal/alert/suppressor.go; new
engine/internal/alert/suppression_rule_scope_test.go; new
engine/internal/api/suppression_rule_scope_test.go; docs/architecture.md;
docs/api-reference.md; docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261004-suppression-rule-scope.json; rolling roadmap.

Low risk: reject only enabled lists made entirely of empty strings, after the
existing prefix checks. Preserve diagnostic precedence, mixed empty/valid and
duplicate IDs, exact identity including surrounding whitespace, disabled skip,
nil/empty unrestricted scopes, IP/direction filters, API schemas/status mapping,
raw file loader/saver structural validation and atomic manager persistence.
No trimming, rule-existence checks, new validation of disabled rules, wildcard
syntax, private input, department messages, dependency/toolchain changes,
release/publication or SLO claims. Stop for ambiguous static/compile/Git/Vault
evidence, competing edits, new authority or a second increment.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Explicit scope cannot silently become unrestricted | Exact NewSuppressor guard after prefix validation; direct public constructor and manager constructor rejection for one/multiple empty IDs; mixed candidate set returns no partial suppressor |
| Existing scope semantics remain | Direct nil/empty all-rule controls; mixed valid/empty/duplicate exact scopes, whitespace-bearing literal IDs, disabled empty-only rule; source/input and alert/slice preservation; source/destination/any range controls |
| Rejected manager operations preserve prior state and allow retry | Real file-backed Add/Update/Reload rejection; complete file bytes/mode/directory membership and List/Filter/input preservation; valid same-operation retry, loaded canonical file and published filter; raw structural loader acceptance distinguished from compiler rejection |
| HTTP reaches the real manager boundary | Real file-backed manager via Handler; POST/PUT 400 VALIDATION_ERROR and reload 500 INTERNAL_ERROR, unchanged list/filter/file; valid retry succeeds with existing status/envelope semantics |
| Earlier diagnostics remain authoritative | Direct invalid source/destination/any prefix and missing-prefix cases with empty-only IDs assert their prior exact diagnostic |
| Reviewable delivery | Pinned Go 1.26.8 owning-module alert/API/pipeline/cmd compile-only, binaries unexecuted; static exact source/format/docs/JSON/roadmap multisets/history/authority/links/fences/nine-path/sensitive/diff; exact non-force push/fetch/Vault replay and stable prose preservation |

Behavioral/race/full/scanner/knowledge/traffic/acceptance checks are
**not run; delegated by user** under the persistent department split. Authored
regressions require departmental execution; compile/static evidence establishes
neither actual boundary coverage nor runtime, race, HTTP or SLO success.

## Delivery and forward queue

Complete exactly R90-175. Commit only the nine intended paths, push without
force, freshly verify HEAD/origin/main/FETCH_HEAD, then synchronize the exact
full-SHA range to the unique existing local sibling Vault. Reconcile current
stable prose while preserving substantive historical/topic prose and immutable
iteration notes; replay the identical range and compare all Markdown hashes.
Record verified facts in one three-path docs-only closure of the same increment,
then push/fetch/sync that exact second range. Do not add a self-reference closure.
No other local ready item is defined. R90-75 retains its full independent
departmental acceptance contract; next trigger audits fresh source/history/queue
before selecting and persisting a separate eligible plan.


## Compile and Static Checkpoint (2026-10-04)

Runtime source differs only in the public constructor comment and a three-line
check after all existing prefix validation. Nil/empty unrestricted lists,
nonempty literal IDs (including whitespace), mixed empties, duplicates and
disabled skip retain their established source branches. No manager/file/API
runtime implementation changes. Current repository suppression config
contains an empty suppression set and is unaffected by the new guard.

Five direct declarations in two files are authored. Constructors assert nil
results for one/multiple empty IDs across source/destination/any ranges with a
valid earlier candidate; caller rules remain intact. Compatibility asserts
literal exact IDs, nil/empty all-rule scopes, duplicates/empties, disabled skip,
both source/destination probes, range misses, Filter contents and alert/input
identity. Diagnostic cases reach missing/invalid source/destination/any prefixes
before the new guard. Real file-backed Add/Update/Reload reject, compare complete
bytes/mode/directory membership/List/Filter/input, then retry the same operation
with a valid scoped candidate and observe canonical file plus published filter.
HTTP Handler uses the real file-backed manager, exact details/request ID,
POST/PUT 400 and reload 500 semantics, file/list/filter preservation, valid retry
statuses/responses and loaded canonical persistence. All declarations unexecuted.

Pinned Go 1.26.8 preflight and the complete fail-fast alert/API/pipeline/cmd
compile-only chain passed; binaries remain outside the repository and unexecuted.
docs-check and static exact transform/Go format/193 JSON/179 unique roadmap
pairs/all 178 prior Definitions and full roadmap history/unchanged split/R90-75/
horizon/historical handoff/local links/fences/nine-path/sensitive/diff pass.
All 429 baseline Vault Markdown hashes remain unchanged. The first local static
checker trimmed Git porcelain leading spaces and misread one path; its parser
was corrected and the entire static chain rerun successfully. No competing edit,
code/compile failure, scope or acceptance deviation occurred. Existing skill
instructions cover this work; no reusable skill update is warranted.

Behavioral/race/full/scanner/knowledge/traffic/acceptance checks are **not run;
delegated by user**. No runtime, HTTP, race or SLO pass is inferred. Recent phase
progress spans supplied-evidence SLO tooling, patched candidate/toolchain and
native correctness repairs; execution debt remains delegated rather than cleared
by implementation volume. No next increment started; feature delivery pending.


## Completion and Forward Queue Refresh (2026-10-04)

Feature `1378776219ee97a1f6074c0d12b0b016d6d1eb81` contains exactly the nine planned paths. Non-force push and
immediate fresh fetch verify clean HEAD/origin/main/FETCH_HEAD equality. Exact
full-SHA range `e4cf2c78a316b22f9439604f88d52d7a26c0c911..1378776219ee97a1f6074c0d12b0b016d6d1eb81` synchronized to
`04-开发迭代记录/2026-10-04-1378776219-CI知识同步.md`; Git-resolved range, nine-path note,
full index and MOC verified. Fourteen current stable notes reconciled; all prior
substantive topic prose retained exactly in an explicit R90-174 history section,
excluding only the versioned bounded generated CI MOC region in comparisons.
All 382 baseline immutable iteration hashes preserved. Identical feature replay
preserves all 430 Markdown hashes; snapshot JSON SHA-256:
`4c9c80854b0f4917b371f7a224a4a1d8cd0d6a97dba786a3ab668f39b9691b8b`.
Existing unique sibling local Vault supplied explicitly; no second or remote
Vault created. No transport or synchronization failure.

Acceptance comparison confirms the exact three-line guard follows every prior
prefix check and precedes candidate append/publication/persistence. Five public
regression declarations reach the constructor, real file-backed manager and
HTTP Handler boundaries promised by the plan. One/multiple empty IDs and all
three range fields are represented; compatibility preserves explicit nil/empty
all-rule scopes, mixed/duplicate/literal whitespace IDs and disabled skip. Earlier
missing/invalid-prefix diagnostics are asserted. Rejection checks actual file
bytes/mode/directory membership and published list/filter; retries observe loaded
canonical persistence and matching. HTTP assertions retain existing status/code/
details/request ID and retry envelopes. Source/input preservation and no partial
constructor output are asserted. Every declaration remains unexecuted.

Pinned Go 1.26.8 four-package compile-only and docs/static source/format/193 JSON/
179 unique pairs/178 prior Definitions and complete roadmap history/authority/
historical handoff/links/fences/nine-path/sensitive/diff review passed. Repository
config was directly verified as an empty suppression set; its checkpoint prose
was corrected before feature push and the complete static review rerun. The
local static parser correction is recorded above. No scope, behavioral acceptance
or compilation deviation. Existing skill instructions suffice; no skill edit.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. No runtime, HTTP, race, release or SLO pass inferred.

This single three-path docs-only delivery record closes the same increment.
Resolve its full SHA from Git and verify non-force push/fresh fetched equality
plus exact feature..closure Vault note/index/MOC before reporting. Do not add a
self-reference closure or repeat verified feature delivery. Queue refreshed
without starting a next increment: no other defined local ready item. R90-75
retains its full independent departmental contract; no qualifying outcome was
supplied. Oct 4–Jan 1 horizon unchanged. Next trigger verifies closure/Vault,
audits fresh source/history/queue and persists a separate eligible plan before
editing. Repair only missing delivery evidence.
