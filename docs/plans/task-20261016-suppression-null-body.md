# R90-191: reject null suppression mutation bodies

## Selection and authority

Starting clean fetched HEAD/origin/main/FETCH_HEAD:
`b51b0f72932b4927a8e02e2c39f60af55eb34ff4`. R90-190 feature and sole
three-path closure note/index/bounded MOC and fourteen current notes verified;
462-file identical closure replay required before implementation. Four-week
history: 162 commits, retained-input SLO tooling followed by bounded correctness
repairs and departmental inventories. Behavioral execution remains delegated;
no qualifying R90-75 independent acceptance evidence is present.

The only unfinished row is departmental R90-75. Restore a ready R90-191 after
R90-190 and the existing suppression mutation API. Source shows that JSON null
decodes into a zero Suppression; PUT assigns the path ID, then manager Update
persists and publishes the disabled zero entry. Reject null without changing
valid object semantics. Forecast Oct 16, 2026–Jan 13, 2027; dates do not gate work.

## Scope, risk and boundaries

Seven paths: engine/internal/api/router.go;
engine/internal/api/suppression_null_body_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261016-suppression-null-body.json;
docs/plans/rolling-90-day-roadmap.md. Low risk: suppression decoder only.
Decode into a nullable pointer and reject nil after the established complete
JSON-document check. POST and PUT return 400 VALIDATION_ERROR / Invalid
suppression request / `suppression request body must be a JSON object`.
Authentication, manager availability, routing, malformed/trailing/unknown/size
errors keep precedence. Explicit disabled objects and optional/null object
members retain established meaning. Rejected requests preserve files and filter.

Non-goals: rule/shared decoder changes, route/identity changes, core manager or
file-format policy, mandatory enabled/member fields, migrations, dependencies,
private input, department contact, publication or R90-75 acceptance. Behavioral/
race/full/scanner/knowledge/traffic/acceptance **not run; delegated by user**.
Compile-only is permitted; execute no test binary. Stop for competing changes,
ambiguous compilation/static/Git/Vault evidence, broader policy or new authority.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Null cannot create or overwrite a suppression | Actual Handler and real manager; POST/PUT, whitespace variants, auth on/off, absent/healthy/file-parent configured paths with spaces and in-memory mode; existing active prior and independent probes; full tree/list/filter/caller preservation; healthy file independently loaded. |
| Earlier diagnostics remain authoritative | Actual-router auth and missing-manager controls; malformed, second-document, malformed trailing, unknown-field, oversized trailing, reserved path and missing target checks with exact error envelope; no mutation. Null body must reject before missing-target lookup. |
| Object semantics and recovery remain usable | Explicit disabled create/update, empty update object, null optional members, enabled update with omitted ID; verify serialized presence/absence, exact response, file load/rebuilt manager/filter, neighboring suppression preservation and retry after null rejection. |
| Reviewable delivery with honest evidence | Pinned Go 1.26.8 API/cmd compile-only, gofmt/docs/JSON/complete unique roadmap multisets/history/R90-75/split/prior handoff/90-day horizon/links/fences/seven-path/sensitive/diff review; non-force push/fetch and exact Vault note/index/bounded MOC/stable/topic/immutable/hash replay. One docs-only closure if needed. |

## Checkpoints

Plan/state persisted before behavior or public-document edits. Trace the nullable
pointer, complete-document error order, fixture constructor and configured path
before counting authored regressions as direct coverage. No test execution or
runtime/SLO outcome may be inferred. Refresh the future queue without starting
another increment; existing skill guidance applies unless a new lesson emerges.


## R90-191 Compile and Static Checkpoint (2026-10-16)

The runtime change is confined to decodeSuppression: decode a nullable pointer,
retain complete-document diagnostics, then reject nil with the explicit object
error before Add/Update/persistence/filter publication. Shared/rule decoder,
core manager/file schema and accepted object behavior are unchanged.

Three authored declarations map to all planned boundaries: 48 null POST/PUT
rejections across three whitespace variants, auth on/off and four actual
file/in-memory modes; 20 exact diagnostic controls and two encoded path guards;
healthy-file and in-memory disabled create/update, empty PUT and enabled
optional-null PUT, each after a preserved null rejection. Tests assert serialized
ID absence/null membership, exact response/list/independent file/rebuilt filter,
neighboring prior and caller-probe preservation. Constructor file selection,
pointer/null behavior in pinned encoding/json source and full trailing decode
ordering were traced. Declarations are authored coverage, not executed outcomes.

Pinned Go 1.26.8 API and cmd compile-only passed; no test binary executed.
Formatting, docs-check, 222 JSON files (209 states), 195 complete unique roadmap
row/Definition pairs, prior completed history/full R90-75/testing split/prior
handoff bytes, 90-day horizon, links/fences, seven-path scope, sensitive review
and diff-check passed. All 462 baseline Vault Markdown hashes are unchanged.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated
by user**. No runtime/SLO pass or new authority claimed. Existing skill guidance
on null/omitted serialization, decoder diagnostics and actual resources suffices;
no redundant skill edit. Exact feature delivery/Vault and sole docs-only closure
remain; no next increment started.


## R90-191 Completion and Forward Queue Refresh (2026-10-16)

Feature `cdc53201ce9309811f0ba68dd785e6f969bd8b89` contains exactly seven intended paths.
Non-force push and immediate fresh fetch verified clean HEAD/origin/main/
FETCH_HEAD at that full SHA. Exact local Vault range
`b51b0f72932b4927a8e02e2c39f60af55eb34ff4..cdc53201ce9309811f0ba68dd785e6f969bd8b89`
verified note `04-开发迭代记录/2026-10-16-cdc53201ce-CI知识同步.md`,
seven-path scope, all 626 index commits and versioned bounded MOC links,
resolving abbreviated metadata endpoints through Git.

Fourteen current stable notes reconcile null suppression requests, object and
diagnostic boundaries, delegated execution and current delivery/queue authority.
Complete prior topic bodies are preserved beneath historical headings, excluding
only the versioned generated MOC region. All 447 prior non-stable/non-generated
notes, including 415 immutable iteration records, are unchanged. Identical exact
feature-range replay preserved all 463 Markdown hashes; sorted sha256sum-manifest
SHA-256: `39fd5fa17f1df614eba65e1819c098aa63945ba5ed153033b821caa841672da9`.

Acceptance review confirms the suppression-only nullable decode/complete-document
order and all three direct authored declarations: 48 null rejection cases,
22 diagnostic/path controls and two object/retry resource modes. Full artifacts,
active/list/filter/caller, independent file/rebuilt-manager and neighbor
observations reach the promised boundaries. Pinned Go 1.26.8 API/cmd compile-only
and complete formatting/docs/222 JSON (209 states)/195 unique complete roadmap
pairs/history/R90-75/testing split/prior handoff/horizon/links/fences/seven-path/
sensitive/diff review passed. No test binary executed; behavioral/race/full/
scanner/knowledge/traffic/acceptance **not run; delegated by user**. No runtime
or SLO pass inferred. Deviation: empty ready queue restored with the source-proven
null replacement repair and forecast refreshed; no scope expansion or ambiguous
result remains. Existing skill guidance sufficed without redundant edits.

### R90-191 Single Closure and Resume Authority

This three-path docs-only delivery record is the sole closure. Resolve its full
SHA through freshly fetched HEAD/origin/main/FETCH_HEAD, verify clean exact refs
and feature..closure note/index/bounded MOC, fourteen current notes, complete
prior topic/immutable/feature-note preservation and identical replay. The feature
SHA above is historical evidence. Do not repeat verified commit/push/sync or
create a self-reference closure.

Forward queue refreshed without starting another increment. R90-75 independent
departmental acceptance remains the sole unfinished row, Oct 16–Jan 13 forecast.
No other defined local ready increment. Next trigger audits fresh Git/Vault/
history/source/queue and persists a separate eligible plan before editing; repair
missing evidence only. No delegated test execution, private input, department
contact, publication or acceptance authority is added.
