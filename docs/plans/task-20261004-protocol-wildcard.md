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
