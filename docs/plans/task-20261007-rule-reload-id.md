# R90-184: reject the reserved rule reload ID on HTTP creation

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD:
`9ee44097bb33547964b682626fa553793c66b7b4`. R90-183 feature/sole closure Git
scopes and exact Vault notes/index/versioned bounded MOC verified. Four-week
audit covers 148 commits, 187 unique complete roadmap pairs, 448 Vault Markdown
files, 400 immutable iterations and fourteen current stable notes agreeing with
fetched authority. Recent SLO/toolchain/correctness execution remains delegated;
no missing delivery or qualifying R90-75 outcome found. Sole unfinished R90-75
retains its full independent contract. No local ready increment is defined.

Source proof: validateRuleBasics accepts ID reload. Handler registers the exact
/api/rules/reload endpoint separately from /api/rules/. The former accepts POST
only, so a created rule with that exact ID cannot use its documented PUT/DELETE
URL. The pinned net/http ServeMux implementation and existing API registrations
establish route precedence. Reject that reserved identity on HTTP creation as
the smallest safe contract repair under the existing safest-default policy.
Select only R90-184, dependent on verified R90-183 and existing rule management.
Keep the current Oct 7–Jan 4 forecast and every completed history record.

## Scope, risk, non-goals and stop condition

Seven paths: engine/internal/api/router.go;
engine/internal/api/rule_reload_id_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261007-rule-reload-id.json; rolling roadmap.

Low risk: one exact-name validation branch after existing required/forbidden
character diagnostics. Reject POST /api/rules with ID reload using HTTP 400
VALIDATION_ERROR, message Invalid rule request, detail
`id "reload" is reserved for the rules reload endpoint`. Auth, configured-file
and decode checks still precede it; duplicate, engine validation and persistence
follow it. Case-sensitive IDs such as Reload, RELOAD and reload-extra remain
addressable and valid. File loader/core Engine compatibility remains unchanged.

Non-goals: new ID syntax, case normalization, route changes, reserved-ID migration,
suppression rules, whole-schema tightening, private input, department contact,
publication or R90-75 acceptance. Existing file-loaded reload rules may still
load/match/reload; their existing PUT/DELETE routing remains unchanged.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Stop for ambiguous compile/static/Git/Vault, competing edits, unexpected
diagnostic or compatibility change, new authority or another increment.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Reserved HTTP creation fails before persistence/publication | Public Server.Handler + real Engine fixtures for absent/healthy/existing-reserved/file-parent seed paths and auth modes; exact error/status/request ID; complete bytes/modes/tree snapshots, Rules/count/matching/caller preservation; earlier auth/file/decode and later duplicate/config boundaries distinguished |
| Addressable neighboring IDs remain usable | Public create/get/update/delete/reload controls for Reload, RELOAD and reload-extra; exact response/full rule snapshots, file loader/rebuilt Engine and matching observations, ID-omitted update, nil dependency nonuse and authenticated routes |
| Existing file-loaded reserved IDs and reload routing remain compatible | Public LoadFromFile/Engine.Reload accepts reload; actual POST reload response/count/match/file preservation; PUT/DELETE reserved endpoint keeps 405/Allow POST; request bodies do not mutate seed/snapshot |
| Reviewable delivery and honest evidence | Pinned Go 1.26.8 API/rule/cmd compile-only; format/docs/202 JSON/188 complete unique roadmap multisets/history/full R90-75/split/handoff/links/fences/seven-path/sensitive/diff; non-force push/fresh exact refs/full-SHA Vault stable reconciliation/immutable preservation/replay and one docs-only closure |

Direct regressions are authored and compile-reviewed, not executed. Trace the
public handler, real engine and persistence path before claiming each boundary.
Local baseline Vault hashes/stable bodies and prior roadmap/handoff are backed
up. Reconcile only current stable authority, preserve full previous topic prose
and immutable notes, and replay each exact range with Markdown hash comparison.

## Checkpoints

Plan/state and acceptance map persisted before source or documentation edits.
Complete exactly this increment. Refresh the queue without starting a next item;
no other local ready increment is currently defined. Change a skill only for a
new reusable lesson; existing direct-boundary and delivery guidance applies.


## R90-184 Compile and Static Checkpoint (2026-10-07)

Runtime delta is exactly the three-line exact reload-ID branch after existing
required/forbidden-character checks. Auth/configured-file/decode ordering,
route registrations, persistence and file-loader/core policy are unchanged.
The API reference distinguishes reserved rejection from other duplicate IDs;
the entire prior handoff is retained with an appended four-declaration supplement.

Direct source covers eight rejected-create fixture/auth combinations for absent,
healthy, existing-reserved and file-parent paths with spaces. Complete envelope,
request ID/tree bytes/modes/membership/Rules/count/matching/input assertions reach
the real public handler and Engine before persistence. Seven diagnostic controls
cover auth/file/decode/required-ID/forbidden-character precedence and reserved
identity before name/config validation. Case-sensitive Reload/RELOAD/reload-extra
controls reach create/list/ID-omitted update/delete/POST reload, full rule values,
persisted loader/rebuilt Engine and independently expected alerts. Existing
file-loaded reload identities exercise POST reload and unchanged 405/Allow POST
PUT/DELETE on literal and percent-encoded route paths, with complete seed
preservation. Source review traces constructors, routing and actual files;
these are authored boundaries, not executed coverage.

Pinned Go 1.26.8 API/rule/cmd compile-only passed; no binaries executed. Complete
static formatting/docs/202 JSON/188 unique complete roadmap multisets/prior
history/full R90-75/testing split/horizon/handoff/links/fences/seven-path/sensitive/
diff checks passed. All 448 baseline Vault Markdown hashes remain unchanged.
No unresolved compile/static result or scope expansion.

A separate local netsentry-next skill refinement adds actual-router creation and
subsequent management review for route identifiers, including reserved and
encoded names. This reusable lesson prevents accepted-but-unaddressable resources
from being mistaken for a valid identifier contract. Skill Markdown structure
validated; the local skill is outside the repository commit.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation/static source review does not establish executed rejection,
persistence/preservation, CRUD, routing, race, release or SLO outcomes. Feature
and one docs-only closure delivery/Vault remain pending; no next increment started.
