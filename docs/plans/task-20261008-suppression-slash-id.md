# R90-186: preserve suppression identity across slash-bearing HTTP paths

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD:
`5ebf12e2ed5a39b1938219ebd07ae0ea9941d364`. R90-185 feature/sole closure
scopes, exact Vault ranges/index/versioned bounded MOC, fourteen current stable
notes and prior-prose reconstruction verified. Four-week audit covers 152 commits,
189 complete unique roadmap pairs, 452 Vault Markdown files and 405 immutable
historical records excluding the generated index. SLO/toolchain/correctness
execution debt remains departmental; no missing latest delivery or qualifying
R90-75 acceptance found. Sole unfinished R90-75 retains its full contract.

Source proof: suppression creation rejects reload but permits slash-bearing IDs.
Management strips leading/trailing slashes from the decoded URL ID, then rejects
remaining internal slashes. A created /prior/ ID is inaccessible as itself and
an encoded management URL can select prior instead. Select only R90-186 as the
smallest coherent slash-identity repair, dependent on verified R90-185 and
existing routes/manager. Restore the empty ready queue under the safest-default
policy; preserve completed history and the Oct 8–Jan 5 forecast.

## Scope, risk, non-goals and stop condition

Seven paths: engine/internal/api/router.go;
engine/internal/api/suppression_slash_id_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261008-suppression-slash-id.json; rolling roadmap.

Low risk, two related boundaries: after the existing reserved-ID check and before
Add, reject any creation ID containing literal / with 400 VALIDATION_ERROR,
message Invalid suppression request, detail `id cannot contain /`, for enabled
and disabled inputs. Auth/manager/decode and reserved reload diagnostics retain
priority; duplicate/CIDR/compiler/persistence checks follow the new guard.
Remove the by-ID slash trimming so its existing empty/slash/reload rejection
checks the exact decoded ID and returns 404 NOT_FOUND before auth or mutation.
Normal IDs retain their manager/resource identity. Percent-escape text is a
literal ID when its percent sign is encoded; question/hash and case neighbors
remain usable through correctly encoded paths.

Non-goals: other ID syntax, double decoding, ServeMux registration/redirect policy,
global HTTP normalization, core manager/file restrictions, automatic legacy
migration, suppression semantics, dependencies/toolchain, private input,
department contact, publication or R90-75 acceptance. Standard raw-path cleanup
can redirect before the handler; only decoded IDs reaching the by-ID handler
get its 404 contract. Legacy slash-bearing file entries still load/filter/reload
and require file edits or direct manager methods for management. Behavioral/race/
full/scanner/knowledge/traffic/acceptance **not run; delegated by user**. Stop for
ambiguous compile/static/Git/Vault, broader router policy, competing edits or new
authority outside this increment.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Slash creation rejects before effects | Actual public handler plus real file-backed manager for /, /prior, prior/, /prior/, prior/child, /reload and reload/; absent/healthy/pre-existing-slash/file-parent paths with spaces; enabled/disabled and auth modes; exact envelope/request ID, full tree/List/independently expected Filter/caller preservation |
| Diagnostic ordering stays explicit | Actual auth/manager/malformed/unknown-field/required-ID/reserved-ID controls, plus slash priority over duplicate/CIDR/compiler; complete state preservation |
| Management cannot trim a decoded slash ID into a neighbor | Actual literal trailing and percent-encoded leading/trailing/internal/all-slash PUT/DELETE with ID-omitted update bodies; auth modes and exact 404; coexistence of prior and /prior/ with complete state/filter/tree preservation; raw doubled-slash redirect control distinguishes ServeMux from handler |
| Slash-free encoded IDs retain exact CRUD | Authenticated create/list/ID-omitted update/delete/reload for ordinary, literal percent-escape, question/hash and case neighbors; actual encoded routes, complete responses, exact requested and neighboring List/Filter state, persisted loader/rebuilt-manager observations |
| Legacy manager/file policy stays usable | Slash-bearing load/save/constructor/filter/POST reload and direct Update/Delete/Add; manager transaction/persistence and rebuilt filter assertions, no automatic migration |
| Reviewable honest delivery | Pinned Go 1.26.8 API/alert/cmd compile-only without execution; format/docs/JSON/complete unique roadmap multisets/history/full R90-75/split/handoff/horizon/links/fences/seven-path/sensitive/diff; exact non-force push/fetch/Vault stable preservation/replay and one docs-only closure |

Five regression declarations are planned, authored and compile-reviewed rather
than executed. Reuse versioned helpers in suppression_reload_id_test.go while
tracing actual constructors, file paths, filtering and router normalization.
Baseline Vault hashes/stable prose and roadmap/handoff are backed up. Preserve
all prior non-generated body prose and immutable records across exact-range replay.

## Checkpoints

Plan/state/acceptance map persisted before runtime or public-document edits.
Finish only this increment and refresh the queue without starting another.
Assess a small generic skill refinement for exact identity assertions through
router decoding/normalization; keep any skill update outside repository delivery.

## Source-review deviations before validation

Review found that earlier R90-184/185 ID-omitted prose describes a body whose
model still serializes an empty id member. Preserve those source/history records,
clarify their actual ID-empty boundary in the current handoff, and explicitly
delete the id member in the new absent-ID fixtures. This is a bounded evidence
correction within the existing seven-path scope, not an executed result.

The initial formatting step rejected an extra closing brace in the new regression
file. No compilation or later validation ran in that fail-fast sequence. Correct
the brace and rerun the complete formatting/compile/static chain before delivery.

Pinned Go 1.26.8 ServeMux source returns a 307 temporary redirect during raw-path
cleanup. Correct the initially authored historical 301 expectation to that exact
pinned-runtime boundary before delivery and rerun compilation/static review.


## R90-186 Compile and Static Checkpoint (2026-10-08)

Runtime delta is the four-line slash creation guard after the reserved-ID check,
plus removal of the one by-ID trim line. Authentication/availability/decode and
reserved creation diagnostics retain priority; management applies its existing
404 check to the exact decoded ID before auth or mutation. Registrations, raw
ServeMux normalization, core manager/file policy and filter semantics are intact.

Five authored declarations reach actual public handlers and real file-backed
managers. Seven slash forms across four fixture/auth/enabled variants total
112 rejected creates; nine diagnostic controls distinguish earlier checks and
slash priority over duplicate/CIDR/compiler failures. Prior and /prior/ coexist
for 32 literal/encoded management rejection cases across eight paths, methods
and auth modes, with truly absent body ID, exact 404/request ID and complete
tree/List/serialized caller/Filter preservation. Two doubled-slash raw controls
verify the pinned source's 307/Location boundary separately. Ordinary, literal
percent-escape, question/hash and case neighbors reach exact encoded CRUD and
reload with full response/state plus persisted loader/rebuilt-manager filters.
Legacy slash identities retain load/save/constructor/filter/reload and direct
Update/Delete/Add. Shared fixtures' actual path/default/manager resolution and
Add/replaceLocked order were traced; no ignored option or weaker resource stands
in for a planned boundary.

The current handoff clarifies prior R90-184/185 ID-omitted wording as serialized
ID-empty fixtures, preserving their source/history. New source explicitly
deletes the member and checks the encoded body for absence. Initial formatting
rejected an extra closing brace before compilation; it was corrected. Pinned
Go 1.26.8 ServeMux source corrected a historical 301 expectation to 307 before
delivery. A temporary verifier miscounted helper calls, was corrected without
source changes, and the complete fail-fast formatting/compile/static sequence
was rerun successfully. No unresolved result or scope expansion remains.

Pinned Go 1.26.8 API/alert/cmd compile-only passed; no binaries executed. Complete
format/docs/217 documentation JSON files (204 states)/190 complete unique roadmap
multisets/history/full R90-75/testing split/handoff/horizon/links/fences/seven-path/
sensitive/diff review passed. All 452 baseline Vault Markdown hashes remain
unchanged. Source counts and every acceptance boundary were compared to the plan.

Separate local netsentry-next guidance now checks actual serialized member
presence for omitted/null/empty fixtures, exact requested and neighboring
resource identity after router decoding/normalization, and pinned-runtime source
for redirect/handler-entry assertions. Markdown structure validated; this small
generic refinement remains outside repository feature delivery.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation/static review and authored declarations establish no executed
rejection, preservation, exact CRUD, routing, race, release or SLO pass. Feature
and sole docs-only closure delivery/Vault remain pending; no next increment
started. Legacy slash entries require file/direct-manager management; standard
raw-path redirects remain outside the decoded-handler rejection boundary.
