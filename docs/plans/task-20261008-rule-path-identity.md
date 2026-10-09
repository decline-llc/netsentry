# R90-187: preserve exact decoded rule identity during HTTP management

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD:
`e05a812dc955b7975c9b0ba2e9ce77504c4ad53d`. R90-186 feature/sole closure
Git/Vault scopes, full index, bounded MOC, fourteen stable notes and prior topic
reconstruction verified. Four-week audit: 154 commits, 190 complete unique
roadmap pairs, 454 Vault Markdown files and 407 immutable historical records
excluding the generated index. Recent correctness work adds compiled regressions;
execution and independent R90-75 acceptance remain departmental. No missing
latest delivery or new acceptance evidence.

Rule creation already rejects slash IDs, but by-ID management trims leading and
trailing slashes before its guard. An absent-ID PUT or DELETE for decoded /prior/
can select prior. Core validateRuleSet and file loading allow legacy slash IDs.
Select only R90-187, dependent on verified R90-186 and existing rule routes/engine.
Restore the empty ready queue under the safest-default policy; preserve history
and the Oct 8–Jan 5 forecast.

## Scope, risk, non-goals and stop condition

Seven paths: engine/internal/api/router.go;
engine/internal/api/rule_path_identity_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261008-rule-path-identity.json; rolling roadmap.

Low risk: remove only the rule by-ID slash trim. The existing empty/slash guard
then checks the exact decoded ID and returns 404 NOT_FOUND / Rule not found
before auth, decoding, lookup or mutation. Slash-free IDs retain exact case and
percent-escape text through correctly encoded routes. Creation syntax, reload
registration, later diagnostics and engine/file validation retain their contracts.
Legacy slash IDs remain usable through file editing, loader, Engine.Reload and
matching, and the collection reload endpoint.

Non-goals: ServeMux normalization/registrations/raw redirects, double decoding,
other ID syntax, suppression behavior, core/file restrictions or migration,
dependencies/toolchain, private input, department contact, publication or formal
SLO acceptance. Pinned Go 1.26.8 source establishes the separate raw-path 307
cleanup boundary before handler entry. Behavioral/race/full/scanner/knowledge/
traffic/acceptance **not run; delegated by user**. Stop for ambiguous compile/
static/Git/Vault, competing edits, broader router policy or new authority.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Decoded slash paths cannot select neighbors | Actual router PUT/DELETE; eight literal/encoded paths, two auth modes, absent/healthy/pre-existing-slash/file-parent artifacts with spaces; truly absent update ID; exact 404/request ID; complete tree modes/bytes/membership, full Rules/count/independently expected matching and callers preserved |
| Routing/diagnostic boundaries stay explicit | Empty-ID and missing-auth slash rejection before auth; ordinary auth/file/decode/body-ID mismatch/unknown-ID/method controls; two raw doubled-slash requests assert pinned 307/Location separately from handler 404; rejected state preserved |
| Encoded slash-free identities retain exact CRUD | Actual create/list/absent-ID update/delete/reload for ordinary, literal percent-escape and case neighbors; full responses/target/neighbor snapshots, independent seed load/rebuilt engine/matching |
| Legacy file/engine identities remain usable | Coexisting prior and /prior/ save/load/Engine.Reload/matching, collection POST reload and file edit/reload; persisted and rebuilt observations, no migration |
| Honest reviewable delivery | Pinned Go 1.26.8 API/rule/cmd compile-only without execution; format/docs/JSON/complete unique roadmap multisets/history/full R90-75/testing split/handoff/horizon/links/fences/seven-path/sensitive/diff; non-force push/fresh refs/exact Vault scope/stable/topic/immutable/hash replay; one docs-only closure |

Four direct regression declarations are planned. Trace actual engine, seed path
and pinned router before reusing versioned helpers. New update fixtures delete
the id member and assert wire absence. Baseline Vault/stable/history/handoff
snapshots are retained locally.

## Checkpoints

Plan/state/acceptance map persisted before runtime and public-document edits.
Review every direct boundary; compilation does not prove execution. Existing
local skill guidance covers serialized omission/exact route/neighbor preservation
and pinned redirects; avoid redundant updates. Finish exactly this increment;
refresh the queue without starting another.


## R90-187 Compile and Static Checkpoint (2026-10-08)

Runtime delta is exactly one removed rule by-ID slash-trim line. The existing
empty/slash guard now checks the exact decoded ID before authentication,
decoding, lookup or mutation. Creation syntax, route registration, raw redirects,
suppression behavior and core/file validation retain their contracts.

Four authored direct declarations reach the actual public router, configured
seed artifacts and real Engine. Eight literal/encoded paths, two methods/auth
modes and four fixture kinds give 128 rejection cases, including coexisting
prior and /prior/. Truly absent-ID bodies remove the member and assert encoded
absence. Complete tree membership/modes/bytes, Rules/count, independently
expected matching and packet/request callers are preserved. Ten controls cover
empty-ID/slash-before-auth, ordinary auth/file/malformed/unknown-field/mismatch/
unknown-ID/method diagnostics; two raw doubled-slash controls separate pinned
307/Location from handler 404. Five ordinary/percent-escape/case identities
exercise exact encoded create/list/update/delete/reload with full responses,
neighboring prior preservation, persisted load/rebuilt engine/matching. Legacy
save/load/Engine.Reload/matching and file-edit/collection-reload remain usable,
with read-only reload and caller preservation. Fixture wrapper and reused
helpers were traced through actual save/load/Reload and configured file paths.
Every acceptance maps to a direct authored boundary, not executed evidence.

Pinned Go 1.26.8 API/rule/cmd compile-only passed; no binaries executed. Complete
fail-fast formatting/compile/static review passed. Formatting/docs/218 docs
JSON (205 states)/191 complete unique roadmap multisets/history/full R90-75/
testing split/handoff/horizon/links/fences/seven-path/sensitive/diff checks passed.
All 454 baseline Vault Markdown hashes remain unchanged. Source-proven empty
queue restoration is recorded; no unresolved result, competing edit or scope
expansion. Existing local skill serialized-member/exact-route/neighbor/pinned
runtime guidance applied and Markdown checked; no redundant update required.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation and static review establish no executed rejection,
preservation, CRUD, routing, race, release or SLO pass. Legacy slash identities
require seed-file editing/reload; raw redirects precede the repaired handler.
Feature delivery/Vault and the sole docs-only closure remain pending. No next
increment started; full independent R90-75 acceptance remains outstanding.


## R90-187 Completion and Forward Queue Refresh (2026-10-08)

Feature `0abe531590a48768a87a2b45b5ef35794d415118` contains exactly seven intended paths. Non-force push
and immediate fresh fetch verified clean HEAD/origin/main/FETCH_HEAD at that
full SHA. Exact local Vault full-SHA range `e05a812dc955b7975c9b0ba2e9ce77504c4ad53d..0abe531590a48768a87a2b45b5ef35794d415118` verified iteration
`04-开发迭代记录/2026-10-08-0abe531590-CI知识同步.md`, seven-path scope,
all 618 full-index commits and versioned bounded MOC, resolving abbreviated
metadata through Git. Fourteen stable current notes reconcile exact decoded
rule identity and delivery/queue authority. Complete prior non-generated topic
prose is preserved under explicit historical headings; all 407 baseline
historical records excluding generated index remain unchanged. Identical range
replay preserved all 455 Markdown hashes; snapshot JSON SHA-256 `fbd35bf6b7ca86e4179300bf16e5f908d9108d38b7afa7fafe19dc2b51cd22aa`.

Acceptance comparison confirms exactly one removed trim line and the existing
empty/slash guard before auth, decoding, lookup or mutation. Four direct authored
router/real Engine/file declarations reach 128 decoded path cases with complete
preservation, ten diagnostic controls, two pinned raw 307/Location controls,
five exact encoded CRUD/list/reload identities and legacy seed-edit/reload.
Truly absent-ID fixtures check wire absence. Actual configured seed artifacts,
constructor/helper defaults, loader/Reload, independent matching and router
normalization were traced. Existing creation/suppression/core/file and raw
redirect contracts retain their scope. No unresolved result or scope expansion.

Pinned Go 1.26.8 API/rule/cmd compile-only and full fail-fast format/compile/static
review passed; no binaries executed. Docs/218 JSON (205 states)/191 complete
unique roadmap multisets/history/full R90-75/testing split/handoff/horizon/links/
fences/seven-path/sensitive/diff passed. Existing local skill guidance applied
and Markdown checked without redundant changes. Behavioral/race/full/scanner/
knowledge/traffic/acceptance **not run; delegated by user**. No executed
rejection/preservation/CRUD/routing/race/release/SLO result is inferred. Legacy
slash IDs require file editing/reload; raw redirects precede the handler.

### R90-187 Single Closure and Resume Authority

This three-path docs-only delivery record is the sole closure of the increment.
Resolve its full SHA through fresh Git after commit; verify non-force push/fresh
clean HEAD/origin/main/FETCH_HEAD, exact feature..closure three-path Vault note/
index/versioned bounded MOC, fourteen stable current notes, complete prior topic
and immutable/feature-iteration preservation, and identical hash replay. The
feature SHA above is historical evidence. Do not repeat verified delivery or
create a self-reference closure.

Forward queue refreshed without starting another increment: no other defined
local ready item. R90-75 remains the sole unfinished row with its full independent
departmental acceptance outstanding; Oct 8–Jan 5 forecast. The next trigger
audits fresh Git/Vault/history/source/queue and persists a separate eligible
plan before edits. Repair missing evidence only; no delegated execution/private
input/department contact/publication/SLO acceptance authority added.
