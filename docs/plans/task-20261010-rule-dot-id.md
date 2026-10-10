# R90-188: reject dot-segment rule IDs at the HTTP boundary

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD: `728375b4ec4b7b234cb7edf90597716835e2c50d`.
R90-187 feature and sole closure notes/scopes, full 619-commit index, bounded
MOC and fourteen stable current notes verified. Four-week phase audit covers
156 commits: SLO evidence tooling followed by bounded correctness deliveries;
behavioral execution remains departmental and R90-75 has no new acceptance.
Baseline snapshots retain 456 Vault Markdown files and 409 immutable iteration
records excluding the generated index. Completed history and full R90-75
contract remain authoritative.

The empty ready queue is restored with one source-proven correctness repair.
HTTP rule creation accepts exact `.` and `..`; literal management paths are
cleaned away by pinned Go 1.26.8 ServeMux before handler entry. Encoded dots
reach the current by-ID handler. Rule file/core validation allows legacy dots.
Select R90-188 after verified R90-187. Forecast: Oct 10–Jan 7 (90 days).

## Scope, risk and boundaries

Seven paths: engine/internal/api/router.go;
engine/internal/api/rule_dot_id_test.go; docs/api-reference.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261010-rule-dot-id.json; rolling roadmap.

Low risk: validateRuleBasics rejects exact `.` and `..` with 400
VALIDATION_ERROR / Invalid rule request / `id cannot be . or ..`. By-ID handling
rejects those exact decoded IDs with 404 NOT_FOUND / Rule not found before
authentication, decoding, lookup or mutation. Authentication/file/decode
precedence at creation remains. Other dots, percent text and case identities
remain exact. Legacy IDs retain file loading, Engine.Reload, matching and
collection reload; edit files to manage them. Raw-path redirects remain pinned
307 with exact Location, separate from encoded-path handler rejection.

Non-goals: suppression policy, ServeMux registrations/normalization, double
decoding, core/file migration or ID restrictions, dependencies/toolchain,
private input, department contact, publication and SLO acceptance.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated
by user**. Compilation executes no test binary. Stop for ambiguous compile,
static, Git or Vault evidence, competing edits, broader routing policy or new
authority. No subagents or test-department contact is authorized.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| HTTP creation rejects each dot ID without side effects | Public router + actual Engine; both IDs, enabled/disabled, auth on/off and absent/healthy/legacy-dot/file-parent artifacts; 32 cases; full tree modes/bytes/membership, Rules/count, independent matching and caller preservation |
| Decoded dot management rejects before mutation | Four case-variant encoded dot paths, PUT/DELETE and auth on/off including missing auth/malformed body; coexisting dot/neighbor engine/file snapshots preserved; raw . and .. 307/Location controls |
| Ordinary diagnostics retain precedence | Auth, missing seed, malformed, unknown field, slash and reload-ID creation controls with exact envelopes and preserved state |
| Other dot/percent identities retain exact CRUD | Actual encoded creation/list/absent-ID update/delete/collection reload for ..., .prior, prior., prior..id, %2e, %2E; full response/neighbor/persisted load/rebuilt matching; absent body ID asserts wire omission |
| Legacy dots retain file/core compatibility | Save/load/Engine.Reload/matching and file edit/collection reload with exact . and .. IDs and neighbor preservation |
| Reviewable honest delivery | Pinned Go 1.26.8 API/rule/cmd compile-only, format/docs/JSON/complete unique roadmap row-Definition multisets/history/full R90-75/split/handoff/horizon/links/fences/seven-path/sensitive/diff; non-force push/fetch exact refs; exact local Vault note/index/bounded MOC/stable/topic/immutable/hash replay; one docs-only closure |

## Checkpoints

Plan/state/evidence map and baseline snapshots persisted before behavior or
public documentation changes. Existing skill exact-route/serialized-omission/
neighbor/pinned-runtime guidance applies; avoid redundant additions.
Compare each authored direct boundary against the plan at closeout. Refresh the
queue without starting a second increment.


## R90-188 Compile and Static Checkpoint (2026-10-10)

Runtime delta: the rule by-ID guard rejects exact decoded . and .., and
validateRuleBasics rejects both with detail `id cannot be . or ..` before
transaction/persistence/publication. Existing creation auth/file/decode
precedence and other diagnostics retain their contracts. No normalization,
registration, suppression, core/file restriction or migration change.

Five direct declarations reach the actual router/Engine/configured seed file.
32 creation cases cover both IDs, enabled/disabled, auth on/off and four
artifacts with spaces. 24 encoded management cases include missing auth and
malformed bodies, plus four pinned raw 307/Location controls. Six ordinary
diagnostic controls preserve full state. Six other-dot/percent IDs exercise
exact create/list/absent-ID update/reload/delete with full responses, neighboring
prior, independent persisted load/rebuilt engine/count/matching. Legacy . and
.. retain identity through save/load/Engine.Reload and collection reload before
and after file edits. Reused fixture constructors, selected modes, actual seed
paths, serialized omission and helper matching defaults traced. Full tree
membership/modes/bytes, Rules/count and callers are checked where promised.
Every acceptance maps to authored direct boundaries, not executed outcomes.

Pinned Go 1.26.8 API/rule/cmd compile-only passed; no test binary executed.
Formatting/docs/219 JSON (206 states)/192 complete unique roadmap multisets/
complete history/full R90-75/testing split/prior handoff/90-day horizon/links/
fences/seven intended paths/sensitive/diff checks passed. All 456 baseline Vault
Markdown files remain unchanged. No unresolved result or scope expansion.
Existing skill exact-route/neighbor/serialized-omission/pinned-runtime guidance
applied and Markdown checked; no redundant skill edit.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated
by user**. Compilation/static review proves no executed rejection, preservation,
CRUD, routing, race, release or SLO pass. Legacy dots require seed edits/reload;
raw redirects precede the guard. Feature delivery/Vault and the sole docs-only
closure remain pending. No next increment started.
