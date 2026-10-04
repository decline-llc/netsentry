# R90-176: retain explicit any protocol wildcard semantics

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD baseline:
`d89966e53ed45903f6b101cfebea19aea25a3984`. R90-175 feature and single closure
Git/Vault ranges, note/index/MOC verified. All 131 four-week phase commits have
iteration and full-index coverage. The 179 roadmap row/Definition multisets
agree without duplicates. Captured 431 Vault Markdown hashes and fourteen
stable-note backups. No missing delivery or supplied R90-75 acceptance outcome.
Recent delivery spans supplied-evidence SLO tooling, candidate/toolchain work
and native correctness repairs; authored regression debt remains delegated.

The local ready queue is empty. Source inference: parseProtocol accepts explicit
`any`, represented by an empty protocol map when used alone. Compilation skips
that marker, so combining it with named protocols incorrectly narrows the map.
Select the bounded default that a union containing explicit `any` permits every
protocol. Blank entries remain ignored; no new protocol names are accepted.
This is source evidence, not an executed failure. Oct 4–Jan 1 forecast retained.

## Scope, risk, non-goals and stop condition

Nine intended paths: engine/internal/rule/engine.go; new
engine/internal/rule/protocol_wildcard_test.go; new
engine/internal/api/protocol_wildcard_test.go; docs/architecture.md;
docs/api-reference.md; docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261004-protocol-wildcard.json; rolling roadmap.

Low risk: mixed explicit wildcard lists will generate detections for additional
protocols as their wildcard requests. Validate every entry before clearing the
compiled map; an invalid entry still rejects regardless of wildcard position.
Reuse the existing addProtocols helper for payload compilation. Preserve case
and surrounding-space handling, blank-entry behavior, named-only unions,
direction/port/IP/window/keyword gates, disabled rules, original config bytes,
priority/early exit, snapshots and HTTP/file schemas and status mapping.
No numeric protocols, wildcard syntax extension, loader/saver validation change,
private input, department contact, dependency/toolchain change, release/SLO
claim or publication. Stop for ambiguous compile/static/Git/Vault evidence,
competing edits, new authority or a second increment.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Explicit any dominates named protocol filters in every rule type | Public Engine.Reload/Match table across payload/IP/port rules, wildcard orders/case/space/duplicates and protocols 0/1/6/17/255; assert complete alerts and input/snapshot preservation |
| Non-wildcard semantics and other gates remain | Nil/empty/blank-only unrestricted controls; blanks mixed with TCP remain TCP-only; named unions/duplicates; direct direction, port, IP, offset/depth, case and disabled controls |
| Wildcard cannot discard malformed entries | All three types, enabled/disabled, invalid-before/after/between wildcard entries; exact earlier error, unchanged count/rules/matching/config and valid retry; earlier direction/window diagnostics preserved |
| File-backed HTTP reaches public behavior | Real Engine with seed file through POST/PUT/reload for all three types; canonical persisted protocol list and active matching; invalid wildcard mixtures preserve bytes/mode/membership/snapshot and existing status/code/details, valid same-operation retry |
| Reviewable delivery | Pinned Go 1.26.8 owning-module rule/API/pipeline/cmd compile-only, binaries unexecuted; static exact transform/format/docs/JSON/roadmap multisets/prior history/authority/links/fences/nine-path/sensitive/diff; exact non-force push/fetch/Vault replay with prior topic prose preserved |

Behavioral/race/full/scanner/knowledge/traffic/acceptance checks are
**not run; delegated by user** under the persistent department split.
Compilation and authored assertions do not establish runtime or HTTP passes.
R90-75 retains its full independent departmental acceptance contract.

## Delivery and forward queue

Complete exactly R90-176. Commit only the nine intended paths, non-force push,
freshly verify clean HEAD/origin/main/FETCH_HEAD, then synchronize the exact
full-SHA range to the unique existing sibling local Vault. Reconcile current
stable prose, archive substantive prior topic prose exactly and preserve
immutable iteration notes. Replay the identical range and compare all Markdown
hashes. Record verified facts in one three-path docs-only closure of this same
increment, then push/fetch/sync that second exact range; no self-reference
closure. No other local ready increment is defined. Next trigger audits fresh
source/history/queue before selecting and persisting a separate eligible plan.


## Compile and Static Checkpoint (2026-10-04)

Runtime change is limited to payload compilation reusing addProtocols and the
shared helper retaining an explicit any marker until all entries validate,
then clearing its owned map. No match, ordering, publication, loader/saver,
API or schema code changed. Blanks retain ignored behavior; named unions and
all earlier direction/window/protocol/port diagnostics retain their source
order. The repository seed has no explicit any protocol marker.

Four direct declarations in two files are authored. Public Engine cases cover
all three types, wildcard orders/case/spaces/duplicates, five representative
protocols including unnamed values, complete alert contents, original config
and packet/snapshot preservation. Compatibility includes nil/empty/blank-only,
blank-plus-TCP, named unions, destination/port/IP gates, payload case/window/
keyword rejection and disabled rules. Invalid entries before/after/between
wildcards reject enabled and disabled candidates with exact error, retaining
rules/count/matching/config, followed by valid retry; direction and negative
window diagnostics retain precedence. Real file-backed HTTP create/update/
reload reaches every type, invalid-before/after ordering, existing 400
VALIDATION_ERROR and exact operation message/details/request ID, file bytes/
mode/membership and published matching/count/rules, valid retry status/response,
loaded canonical original protocol list and every protocol/alert content.
Reload leaves the supplied file intact. All declarations remain unexecuted.

Pinned Go 1.26.8 preflight and complete fail-fast rule/API/pipeline/cmd
compile-only passed after the final test edit; binaries remain outside the
repository and unexecuted. Static exact runtime transform/Go format/docs-check/
194 task JSON/180 unique roadmap pairs/all 179 prior Definitions and complete
roadmap history/authority/handoff/local links/fences/nine-path/sensitive/diff
review passed. All 431 baseline Vault Markdown hashes remain unchanged.
An initial temporary static-harness generation quoting error was corrected;
the complete static chain was rerun successfully. Static source review also
corrected the authored reload assertion to its existing 400 VALIDATION_ERROR
boundary before final compilation. No code/compile or unresolved static failure.
Existing skill instructions suffice; no redundant skill update is warranted.

Behavioral/race/full/scanner/knowledge/traffic/acceptance checks are **not run;
delegated by user**. No runtime, HTTP, race, release or SLO pass is inferred.
R90-75 remains independent; no next increment started. Feature delivery pending.


## Completion and Forward Queue Refresh (2026-10-04)

Feature `e5cc1f8ee9a1e85c7b6610ddcc6fbc378b76cd78` contains exactly the nine planned paths. Non-force push and
immediate fresh fetch verify clean HEAD/origin/main/FETCH_HEAD equality. Exact
full-SHA range `d89966e53ed45903f6b101cfebea19aea25a3984..e5cc1f8ee9a1e85c7b6610ddcc6fbc378b76cd78` synchronized to
`04-开发迭代记录/2026-10-04-e5cc1f8ee9-CI知识同步.md`; Git-resolved endpoints,
nine-path note, full index and bounded generated MOC verified. Fourteen current
stable notes reconciled; every prior substantive body paragraph retained exactly
under explicit R90-175 history, excluding only the actual versioned generated
CI MOC region from prose comparisons. Every baseline immutable iteration note
preserved. Identical feature replay preserves all 432 Markdown hashes; snapshot
JSON SHA-256: `f9567ed711a8c793daa7e42302e682bc4763791c4f094a6cddd61524fd16dd61`. Existing unique sibling local Vault explicitly supplied;
no second/remote Vault created. No push/fetch/synchronization failure.

Acceptance comparison confirms all three compilers reach the shared helper;
every entry validates before its wildcard clears the owned map. Payload compiler
reuse preserves the exact former protocol error and prior direction/window plus
later port checks. Matching/publication/file/API runtime code is unchanged.
Four public declarations reach every planned boundary: complete Engine alerts
for five representative protocols and wildcard/blank/named controls; other gates,
disabled behavior and original inputs/config; invalid-before/after/between,
enabled/disabled retained snapshot and valid retry, earlier diagnostics;
real file-backed create/update/reload for every type and both invalid orders,
400 validation envelopes and complete bytes/mode/membership/count/rules/matching
preservation, canonical original protocol lists and valid retry responses and
alert contents. All declarations remain unexecuted. The repository seed has no
explicit any protocol marker. No unsupported protocol is newly accepted.

Pinned Go 1.26.8 complete rule/API/pipeline/cmd compile-only and docs/static exact
transform/format/194 JSON/180 unique pairs/179 prior Definitions plus full roadmap
history/authority/handoff/links/fences/nine-path/sensitive/diff passed. Temporary
static-harness generation and authored HTTP assertion corrections are recorded
above; complete checks passed after correction. No scope, compile or unresolved
static deviation. Existing skill instructions suffice; no redundant skill edit.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. No runtime, HTTP, race, release or SLO pass inferred.

This single three-path docs-only delivery record closes the same increment.
Resolve its full SHA from Git and verify non-force push/fresh fetched equality
plus exact feature..closure Vault note/index/MOC before reporting; do not add a
self-reference closure or repeat completed feature delivery. Forward queue
refreshed without starting a next increment: no other defined local ready item.
R90-75 retains its independent full departmental acceptance contract; no
qualifying outcome supplied. Oct 4–Jan 1 forecast unchanged. Next trigger verifies
closure/Vault and audits fresh source/history/queue before persisting a separate
eligible plan. Repair only missing delivery evidence.
