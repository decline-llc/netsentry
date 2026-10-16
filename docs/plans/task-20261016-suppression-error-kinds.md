# R90-192: classify suppression errors by cause

## Selection and authority

Clean freshly fetched HEAD/origin/main/FETCH_HEAD:
`585b1048ed764b464f2f3d4c9c76f2c45bcb3049`. R90-191 feature/sole closure
notes and three-path closure scope, 627-commit index, versioned bounded MOC,
fourteen current stable notes and identical 464-file closure replay verified.
Four-week phase audit: 164 commits, retained-input SLO tooling then bounded
correctness repairs and departmental inventories. No missing latest delivery
or new qualifying R90-75 acceptance artifact found. Standing delegated tests
remain active; Oct 16, 2026–Jan 13, 2027 forecast is unchanged.

The only unfinished row is R90-75 with its full external-evidence contract.
Restore a ready R90-192 after R90-191 and the suppression manager/API. Source
uses substring matching for duplicate, missing and unconfigured-file errors.
IDs or filesystem paths containing those phrases can misclassify validation,
persistence, durability or reload failures. Use sentinel identity without
changing existing diagnostic strings or accepted suppression identities.

## Scope, risk and boundaries

Ten paths: engine/internal/alert/suppressor.go;
engine/internal/alert/suppression_error_kind_test.go;
engine/internal/api/router.go;
engine/internal/api/suppression_error_kind_test.go;
engine/internal/api/suppression_error_kind_internal_test.go;
docs/api-reference.md; docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261016-suppression-error-kinds.json;
docs/plans/rolling-90-day-roadmap.md.

Low risk: export missing/duplicate/unconfigured-file sentinel causes, wrap them
without changing Error() text, and use errors.Is at the three API classifiers.
True duplicate remains 409 SUPPRESSION_ALREADY_EXISTS; missing update/delete
remains 404 SUPPRESSION_NOT_FOUND; unconfigured reload remains 409
SUPPRESSIONS_RELOAD_UNAVAILABLE. Validation remains 400, pre-rename persistence
500 INTERNAL_ERROR, post-rename uncertainty 500 SUPPRESSIONS_DURABILITY_UNCERTAIN,
and configured reload failure 500 INTERNAL_ERROR regardless of incidental text.
Preserve prior auth/routing/null/decoding order and manager/file/filter behavior.

Non-goals: new ID restrictions, rule errors, file schema, durability lifecycle,
locking, broad error taxonomy, dependencies, private input, department contact,
publication or SLO acceptance. Behavioral/race/full/scanner/knowledge/traffic/
acceptance **not run; delegated by user**. Compile-only is permitted; execute no
test binary. Stop for competing edits, ambiguous compile/static/Git/Vault,
broader policy or new authority.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Causes survive wrapping while diagnostics remain exact | Public manager/constructor/save errors for duplicate, missing Update/Delete and unconfigured reload; errors.Is on raw/wrapped/joined errors and negative same-text impostors; unchanged caller/list/filter/artifacts. |
| Incidental ID phrases do not select unrelated HTTP statuses | Actual Handler/manager/file for not-found/already-exists/not-configured IDs: invalid create/update stays 400; duplicate stays 409; missing update/delete stays 404; configured invalid-file reload stays 500; healthy object CRUD/reload retains exact IDs and neighbors. |
| Persistence and durability retain their classes | Real non-directory parent paths containing misleading phrases for POST/PUT/DELETE: exact 500 envelope and tree/list/filter/caller preservation, then repaired-path successful retry; internal response-writer seam checks committed marker versus identical plain-text errors and wrapped/joined sentinel causes. Synthetic seam proves classification only, not filesystem durability. |
| Reviewable delivery | Pinned Go 1.26.8 alert/API/cmd compile-only; format/docs/JSON/complete unique roadmap pairs/history/R90-75/testing split/prior handoff/horizon/links/fences/ten-path/sensitive/diff review; exact push/fetch/Vault note/index/bounded MOC/stable/topic/immutable/hash replay; one docs-only closure. |

## Checkpoints

Plan/state/evidence map persisted before behavior or public-document edits.
Trace actual constructors, path selection and error chains before treating
regressions as direct authored coverage. Preserve text and test that misleading
text cannot substitute for cause identity. No executed runtime/SLO outcome is
inferred. Finish this increment only and refresh the queue without starting next.


## R90-192 Compile and Static Checkpoint (2026-10-16)

Three documented sentinels identify missing, duplicate and unconfigured-file
causes. Four manager/validation returns retain exact diagnostic strings; three
API substitutions use errors.Is instead of substring matching. Auth/routing/
null/decoding order, ID acceptance, manager locks, persistence and file schema
are unchanged. Existing committed marker and persistence-prefix handling remain.

Five authored declarations in three files map to every planned boundary:
public manager/constructor/save exact messages and raw/wrapped/joined causes,
negative same-text errors and immutable caller snapshots; eighteen actual-router
ID/error cases plus unconfigured reload; nine real non-directory-parent failures
and repaired-path retries; three phrase IDs through exact public CRUD/list/
encoded absent-ID update/reload with independent file/rebuilt-manager/filter
and neighboring-state observations; five explicitly synthetic response-writer
cases for wrapped/joined missing, committed, persistence and plain-text causes.
Full artifact membership/modes/bytes, active list/filter and caller probes are
observed at preservation boundaries. Synthetic classification is not filesystem
durability evidence. Constructor/path resolution and pinned ENOTDIR versus
ErrNotExist behavior were traced; every actual request targets its promised
resource. No runtime outcome is inferred from declarations.

Pinned Go 1.26.8 alert/API/cmd compile-only passed; no test binary executed.
Format/docs/223 JSON (210 states)/196 complete unique roadmap pairs/prior
completed history/full R90-75/testing split/prior handoff bytes/90-day horizon/
links/fences/ten-path/sensitive/diff review passed. All 464 baseline Vault
Markdown hashes remain unchanged. Behavioral/race/full/scanner/knowledge/
traffic/acceptance **not run; delegated by user**. No ambiguous result or scope
expansion remains. Existing error-cause, fixture-resource and serialized-omission
guidance suffices without a redundant skill edit. Feature delivery/Vault and
sole docs-only closure remain pending; no next increment started.
