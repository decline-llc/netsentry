# R90-183: redact escaped JSON credential names

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD:
`4291faa24f816f51787a9c70fc757c3f29ad8843`. Prior R90-182 feature and sole closure
Git scopes, exact Vault notes/index/versioned bounded MOC and fourteen current
stable notes agree with this fetched baseline. Four-week review covers 146
commits, 186 unique complete roadmap row/Definition pairs, 446 Vault Markdown
files and 398 immutable iteration records. Recent SLO/toolchain/correctness
execution debt remains delegated; no missing delivery or qualifying R90-75
outcome was found. Sole unfinished R90-75 has its full independent contract.

The local ready queue is empty. Source evidence: sensitiveJSONRe recognizes
literal password/token keys only; the existing compatibility test explicitly
leaves pass\u0077ord unchanged. JSON decoding gives that spelling the same
password identity. Extend existing field recognition as the safest bounded
default under the roadmap prerequisite policy, superseding the prior escaped-key
non-goal without adding field names. This is source-supported evidence, not an
executed incident. [Go encoding/json](https://pkg.go.dev/encoding/json#Unmarshal)
and the pinned local decoder source establish JSON string decoding authority.
Select exactly R90-183, dependent on verified R90-182 and R90-148/164 redaction.
Refresh unfinished forecasts to Oct 7–Jan 4; completed history remains intact.

## Scope, risks, non-goals and stop condition

Nine paths: engine/internal/alert/redactor.go; existing redactor_json_test.go;
new redactor_json_keys_test.go; engine/internal/pipeline/redaction_json_keys_test.go;
docs/architecture.md; docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261007-json-key-redaction.json; rolling roadmap.

Low risk: broaden the lexical string-key capture and decode only a complete key,
retaining its exact raw spelling, separators and real value closing quote.
Compare decoded password/token names case-insensitively; redact complete and
preview-end string values using the existing escape-aware value boundary.
Skip undecodable and unrelated keys. Keep headers/pairs, opt-in pre-write
placement, marker and metadata behavior. Existing escaped-key exclusion is
replaced by explicit positive coverage; all other historical source remains.

Non-goals: whole-document parsing, non-string values, additional fields,
malformed multiline values/keys, truncated keys, RawPayload, dependencies,
private input, department contact, release/publication or R90-75 acceptance.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user** under the standing split. Stop for ambiguous compile/static/Git/Vault,
unintended compatibility changes, competing edits, new authority or another
increment. No following increment starts in this trigger.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Equivalent credential names redact complete and cut string values | Public scalar table covers partial/full Unicode escapes, encoded case variants, whitespace, nested/repeated keys, escaped value quotes/backslashes and preview-end dangling/partial escapes; independent expected bytes, decoded key/value identity, valid complete output and idempotence |
| Formatting, noncredential values and batch metadata survive | Public controls cover literal backslash-u keys, escaped quote/slash and malformed escapes, prefix/suffix names, non-string values and existing header/pair behavior; batch checks nil/order/pointers/full Alert metadata and RawPayload |
| Redaction reaches actual preview/write boundary | Real Engine/Worker/base64 payload fixtures for complete JSON and valid payloads longer than 200 bytes cut inside escaped-name values; enabled/disabled/write-failure modes, independent full writer-entry Alert expectations, packet preservation and terminal Stats accounting |
| Reviewable delivery and honest evidence | Pinned Go 1.26.8 alert/pipeline/API/cmd compile-only; formatting/docs/JSON/complete roadmap multisets/prior history/full R90-75/frozen handoff/links/fences/nine-path/sensitive/diff; non-force push/fresh exact refs/full-SHA Vault stable reconciliation/immutable preservation/replay; one docs-only closure |

Named regressions are authored and compile-reviewed, not executed. Compare each
assertion to actual constructor/matcher/preview/write behavior at closeout.
Snapshot local Vault hashes and stable prose before edits. Preserve all prior
topic prose and immutable notes when reconciling current status; replay each
exact pushed range and compare all Markdown hashes. Refresh the queue without
starting a next item; no other local ready increment is presently defined.

## Checkpoints

Plan and state persisted before runtime or documentation edits. Existing skill
guidance covers escape boundaries, direct fixture review and exact delivery;
update a skill only for a new repeatable lesson.


## R90-183 Compile and Static Checkpoint (2026-10-07)

Complete lexical string-key capture now decodes only the key before comparing
case-insensitive password/token identity. Raw key/separator bytes and existing
complete/cut value boundaries survive. The prior escaped-key unchanged assertion
is removed and replaced by explicit positive coverage; all other existing
redaction source is preserved. No dependencies or whole-document parser added.

Four direct authored declarations: per-character/full escapes and encoded case
variants with independent scalar bytes/decoded identities/valid complete JSON;
nested/repeated/whitespace/escaped values/cut tails/idempotence; unsupported and
unrelated key/non-string controls plus nil/order/pointer/full metadata/RawPayload;
real Engine/base64/Worker writer-entry complete and valid larger-than-200-byte
cut fixtures across enabled/disabled/write-failure modes, full independent Alert
expectations, packet preservation and terminal accounting. Source tracing confirms
those constructors and assertions reach their promised resources and boundaries.
The original handoff is preserved byte-for-byte with an appended supplement.

Pinned Go 1.26.8 compile-only passed alert/pipeline/API/cmd; binaries are local
outside the repository and were not executed. Static review passed formatting,
docs-check, 201 task JSON, 187 complete unique roadmap multisets, all prior
history/Definitions/full R90-75/testing split, Oct 7–Jan 4 unfinished forecasts,
links/fences/nine-path scope/sensitive review and git diff --check. All 446
selection Vault Markdown hashes remain unchanged. No unresolved failure or scope
expansion; earlier escaped-key exclusion is explicitly superseded. Existing
skills cover this workflow, so no new repeatable lesson or skill edit warranted.

Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Compilation/source review does not establish executed redaction,
preservation, persistence, race, performance, release or SLO outcomes. Feature
delivery/Vault and one docs-only closure remain pending; no next increment started.


## R90-183 Completion and Forward Queue Refresh (2026-10-07)

Feature `64d2495ef01684cc8abe9b76fb0e5a4c00ba2737` contains exactly nine intended
paths. Non-force push succeeded and immediate fresh fetch verified clean
HEAD/origin/main/FETCH_HEAD at that full SHA. Local Vault exact full-SHA range
`4291faa24f816f51787a9c70fc757c3f29ad8843..64d2495ef01684cc8abe9b76fb0e5a4c00ba2737`
verified iteration `04-开发迭代记录/2026-10-07-64d2495ef0-CI知识同步.md`, exact changed
paths, full commit index and versioned bounded MOC links, resolving abbreviated
metadata through Git. Fourteen current stable notes reconcile escaped-name and
delivery/queue authority; all previous non-generated body prose is preserved
under explicit historical headings, and 398 prior immutable iterations remain
unchanged. Identical range replay preserved all 447 Markdown hashes; snapshot
JSON SHA-256 `39f98f660ab46fbe15b529b401d0fec55fdf32fe5e88775a33f655941aeefb34`.

The initial local preservation verifier wrongly required the original title and
old status body to remain contiguous across the newly inserted current section.
An exact reconstructed-body comparison proved no prose loss; corrected only the
local verifier, then reran the complete exact-range/scope/index/MOC/stable/
immutable/hash review and identical replay successfully. No repository or Vault
prose was changed to satisfy that failed assertion. Existing skill instructions
cover comparison boundaries; no new generic lesson or skill edit warranted.

Acceptance comparison confirms same-field decoded identity, independently
expected scalar bytes/decoded values, complete JSON validity, nested/repeated/
whitespace/case/escape/cut boundaries, idempotence and noncredential/undecodable/
non-string controls. Batch assertions compare order/pointers/nil/full Alert
metadata and RawPayload. Actual Engine base64 matching and 200-byte truncation
feed Worker writer-entry snapshots; valid larger inputs, count and complete
content, enabled/disabled/write-failure modes, packet preservation and terminal
accounting reach the planned boundary on source review. The one obsolete
escaped-key exclusion is replaced by explicit positives; every other existing
redaction assertion and all prior handoff bytes are preserved. No new fields,
whole-document parser, dependencies or performance promise.

Pinned four-package Go 1.26.8 compile-only and complete formatting/docs/201 JSON/
187 unique complete roadmap multisets/history/full R90-75/split/handoff/horizon/
links/fences/nine-path/sensitive/diff checks passed. Only the recorded local
verifier deviation occurred; no unresolved compile/static/Git/Vault result or
scope expansion remains. Behavioral/race/full/scanner/knowledge/traffic/
acceptance **not run; delegated by user**. Delivered implementation and authored
regressions do not establish executed correctness, preservation, persistence,
race, performance, release or SLO passes.

### R90-183 Single Closure and Resume Authority

This three-path docs-only delivery record closes the same increment. Resolve
its full SHA from fresh Git refs after commit, then verify non-force push/fresh
clean HEAD/origin/main/FETCH_HEAD, exact feature..closure three-path Vault note/
index/versioned bounded MOC, fourteen current stable notes, prior topic/immutable
preservation and identical replay. Feature SHA above is immutable historical
evidence. Do not repeat verified feature/closure delivery or create a
self-reference closure.

Forward queue refreshed without starting another increment: no other defined
local ready item. R90-75 remains the sole unfinished row with its full independent
outstanding departmental contract and Oct 7–Jan 4 forecast. Next trigger audits
fresh Git/Vault/history/source/queue and persists a separate eligible plan before
editing; repair missing delivery evidence only. No delegated test execution,
private input, department contact, publication or SLO acceptance is authorized.
