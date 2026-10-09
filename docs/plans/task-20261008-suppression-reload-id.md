# R90-185: reject the reserved suppression reload ID on HTTP creation

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD:
`19f09d64db9748bd9a460e495dc1c43bf54c1e5f`. Prior R90-184 feature and sole
closure scopes, exact Vault ranges, index and versioned bounded MOC verified.
The four-week phase audit covers 150 commits and 188 complete unique roadmap
row/Definition pairs. Current local Vault has 450 Markdown files, 403 historical
records excluding the generated full index, and fourteen current stable notes
agreeing with fetched authority. SLO/toolchain/correctness history retains
departmental execution debt; no missing latest delivery or R90-75 acceptance
found. Preserve all completed history and the independent R90-75 contract.

Source proof: suppression POST decodes then calls Add without reserving reload.
Manager structural validation accepts that ID, including disabled suppressions.
The public router registers /api/suppressions/reload separately; PUT/DELETE use
its POST-only handler, and the generic by-ID handler also rejects reload.
The created resource cannot use its documented management URL. Restore the
empty local ready queue with this exact-name HTTP repair under the standing
safest-default policy. Select only R90-185, dependent on verified R90-184 and
existing suppression management. Refresh unfinished forecasts to Oct 8–Jan 5.

## Scope, risk, non-goals and stop condition

Seven paths: engine/internal/api/router.go;
engine/internal/api/suppression_reload_id_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261008-suppression-reload-id.json; rolling roadmap.

Low risk: one exact-ID check after auth, manager availability and JSON decoding,
before Add. POST /api/suppressions with ID reload returns 400 VALIDATION_ERROR,
message Invalid suppression request, detail
`id "reload" is reserved for the suppressions reload endpoint`.
This applies to enabled and disabled suppressions and takes precedence over
duplicate/set/compiler/persistence errors. Reload, RELOAD and reload-extra
remain exact, manageable IDs. Direct manager Add/Update/Delete, structural file
load/save, constructor and reload compatibility remain unchanged.

Non-goals: route changes, other ID syntax, case normalization, legacy migration,
rule policy, suppression semantics, dependency/toolchain changes, private input,
department contact, publication or R90-75 acceptance. Existing file-loaded reload
IDs retain their current PUT/DELETE routing limitation. Behavioral/race/full/
scanner/knowledge/traffic/acceptance **not run; delegated by user**. Stop for
ambiguous compile/static/Git/Vault evidence, competing edits, incompatible
diagnostics, new authority or work outside this increment.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Reserved creation rejects before Add effects | Public Server.Handler plus real file-backed SuppressionManager for absent/healthy/existing-reserved/file-parent paths with spaces; enabled/disabled and auth modes; exact error/request ID, full tree bytes/modes/membership, List/Filter and input preservation |
| Earlier diagnostics remain authoritative | Actual handler auth/manager/decode controls, including malformed and unknown-field bodies; reserved-ID priority over duplicate and CIDR/compiler failures; complete state preservation |
| Neighboring identities remain manageable | Authenticated public create/list/ID-omitted update/delete/reload for Reload/RELOAD/reload-extra with encoded path variants; full responses, real manager filters, persisted loader/rebuilt-manager controls and deletion body |
| Legacy and direct-manager contracts remain usable | Load/save/constructor/filter/POST reload for file-loaded reload, unchanged literal/encoded PUT/DELETE 405 and Allow POST, plus direct Add/Update/Delete controls; full tree/snapshot preservation at rejected HTTP routes |
| Reviewable delivery with honest evidence | Pinned Go 1.26.8 API/alert/cmd compile-only without binary execution; format/docs/JSON/full unique roadmap multisets/history/R90-75/split/handoff/links/fences/seven-path/sensitive/diff; non-force push/fresh exact refs, full-SHA Vault stable reconciliation/preservation/replay and one docs-only closure |

Regressions are authored and compile-reviewed, not executed. Trace constructors,
resource paths and actual routing before claiming each authored boundary. Local
Vault hashes, stable prose and historical roadmap/handoff are backed up. Preserve
all prior non-generated prose and immutable records; replay each exact pushed
range and compare every Markdown hash. No new reusable skill lesson is needed;
the existing routed-identifier and delivery guidance covers this repair.

## Checkpoints

Plan/state and acceptance map persisted before runtime or public-doc edits.
Complete exactly this increment; refresh the queue without starting another.


## R90-185 Compile and Static Checkpoint (2026-10-08)

Runtime delta is exactly the four-line HTTP creation guard after authentication,
manager availability and JSON decoding, before Add. No route, core manager,
structural file policy, suppression matching or legacy migration changed.
The API reference records both the exact-name contract and legacy limitation;
the entire prior handoff remains with a four-declaration supplement.

Direct authored source covers sixteen fixture/auth/enabled combinations for
absent/healthy/existing-reserved/file-parent paths with spaces. Exact envelopes,
request IDs, complete tree bytes/modes/membership, List, independently expected
Filter results and caller data are compared. Seven diagnostic controls establish
earlier auth/manager/malformed/unknown-field and reserved-ID priority over
duplicate/CIDR/compiler errors. Reload/RELOAD/reload-extra controls reach actual
public create/list/ID-omitted update on literal/encoded paths, POST reload and
encoded deletion, with full response/state and independent persisted loader/
rebuilt-manager observations. Legacy reload identities retain literal/encoded
POST reload and PUT/DELETE 405/Allow POST with complete tree preservation; direct
Delete/Add/Update controls retain public manager/file compatibility. Fixture
constructors, selected paths, actual router and manager transaction boundaries
were traced against every planned acceptance criterion. These are authored
boundaries, not executed coverage.

Pinned Go 1.26.8 API/alert/cmd compile-only passed; no binaries executed. Complete
format/docs/216 documentation JSON files (203 task states)/189 complete unique
roadmap multisets/prior history/full R90-75/testing split/horizon/handoff/links/
fences/seven-path/sensitive/diff review passed. All 450 baseline Vault Markdown
hashes remain unchanged. No unresolved compile/static result or scope expansion.
The server restart preserved source, temporary compiled binaries and baseline
backups; clean expected Git scopes and saved state were rechecked before resume.
No repeated delivery action occurred. Existing skill guidance covers the repair
without a new reusable refinement.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation/static review does not establish executed rejection,
persistence/preservation, CRUD, routing, race, release or SLO outcomes. Feature
and sole docs-only closure delivery/Vault remain pending; no next increment
started.


## R90-185 Completion and Forward Queue Refresh (2026-10-08)

Feature `7f1ff9bbde65474f7f5ae8b205315fb7a6dd4b5a` contains exactly seven intended paths. Non-force push
succeeded and immediate fresh fetch verified clean HEAD/origin/main/FETCH_HEAD
at that full SHA. Local Vault exact full-SHA range
`19f09d64db9748bd9a460e495dc1c43bf54c1e5f..7f1ff9bbde65474f7f5ae8b205315fb7a6dd4b5a` verified iteration
`04-开发迭代记录/2026-10-08-7f1ff9bbde-CI知识同步.md`, exact changed paths,
all 614 full-index commits and versioned bounded MOC links, resolving abbreviated
metadata through Git. Fourteen current stable notes reconcile the suppression
HTTP identity contract and delivery/queue authority. Their complete prior
non-generated body prose remains under explicit historical headings; all 403
prior historical records excluding the generated full index are unchanged.
Identical range replay preserved all 451 Markdown hashes; snapshot JSON SHA-256
`99eb37844e94811b7d397451b596f2bd66ebb82bd61ca2db811bb41df3aa13a6`.

Acceptance comparison confirms only the four-line HTTP creation guard changed
runtime behavior. It runs after auth/manager/decode and before Add. Sixteen
auth/enabled/fixture combinations cover absent/healthy/existing-reserved/file-
parent paths with spaces; full error/request ID/tree bytes/modes/membership/
List/independent Filter/caller observations target the actual resources. Seven
controls distinguish earlier diagnostics and reserved-ID priority over duplicate/
CIDR/compiler validation. Neighboring exact case variants reach the real public
create/list/ID-omitted literal and encoded updates/reload/encoded delete, full
response/state and independently loaded/rebuilt filters. Legacy reserved IDs
reach file helpers, constructor, literal/encoded POST reload and unchanged
PUT/DELETE 405/Allow POST with tree preservation; direct Delete/Add/Update retains
manager policy. These boundaries were traced through the actual constructor,
router, file path and Add/replaceLocked transaction. They are direct authored
assertions, not executed outcomes.

Pinned Go 1.26.8 API/alert/cmd compile-only and complete formatting/docs/216 docs
JSON (203 states)/189 complete unique roadmap multisets/history/full R90-75/
testing split/handoff/horizon/links/fences/seven-path/sensitive/diff passed.
No binaries executed. Only plan deviation is the source-grounded empty-queue
restoration and refreshed unfinished forecast; the server restart was recovered
from verified saved state before further action. No scope expansion, competing
edit or unresolved compile/static/Git/Vault result. Existing generic skill
instructions already cover this routed-identifier repair; no skill change.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation and authored assertions establish no executed rejection,
persistence/preservation, CRUD, routing, race, release or SLO pass. Legacy
file-loaded reload IDs retain their current HTTP PUT/DELETE route limitation.

### R90-185 Single Closure and Resume Authority

This three-path docs-only delivery record closes the same increment. Resolve
its full SHA through fresh Git after commit; verify non-force push/fresh clean
HEAD/origin/main/FETCH_HEAD, exact feature..closure three-path Vault note/index/
versioned bounded MOC, fourteen stable current notes, complete prior topic and
immutable preservation and identical replay. Feature SHA is historical evidence;
do not repeat verified delivery or create a self-reference closure.

Forward queue refreshed without starting another increment: no other defined
local ready item. R90-75 remains the sole unfinished row with its full independent
departmental acceptance outstanding; Oct 8–Jan 5 forecast. The next trigger audits
fresh Git/Vault/history/source/queue and persists a separate eligible plan before
editing. Repair missing evidence only; no delegated execution/private input/
department contact/publication/SLO acceptance authority is added.
