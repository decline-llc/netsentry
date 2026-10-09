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
