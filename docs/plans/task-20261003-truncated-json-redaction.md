# R90-164: redact JSON credential values truncated at preview end

## Baseline and selection audit

Clean fetched main HEAD/origin/main/FETCH_HEAD agrees at
`18aa324a89c211a9f95d37655541ea17b4fd8a96`. Prior R90-163 feature/closure
exact Git paths, Vault range notes, resolved short identifiers, index and MOC
are verified. Sep 5–Oct 3 phase history has 107 commits spanning departmental
SLO tooling, patched release/toolchain delivery and core correctness repairs.
No new qualifying R90-75 outcome or missing prior delivery was found. The
167 unique roadmap rows and Definitions match. Only R90-75 is unfinished;
its independent departmental contract and Oct 3–Dec 31 horizon remain.
Restore one ready increment from the source gap: Engine.Match caps payload
previews at 200 bytes, while the JSON credential regex requires a closing
quote. Secret prefixes survive when that quote lies outside the preview.
No AGENTS or pre-existing edits. Engine module resolves Go 1.26.8. Baseline
407 Vault Markdown hashes and fourteen complete stable notes captured locally.
Plan/state persisted before runtime or other documentation changes.

## Scope, risk and authority

Seven paths: `engine/internal/alert/redactor.go`, existing
`engine/internal/alert/redactor_json_test.go`, new
`engine/internal/pipeline/redaction_truncation_test.go`, `docs/architecture.md`,
this plan, `docs/tasks/task-state-20261003-truncated-json-redaction.json`, roadmap.
Extend only the credential value terminator to accept preview end, optionally
consuming one dangling escape backslash. Preserve opening key/spacing and any
actual closing quote; do not fabricate JSON closure. Existing escape-pair
handling prevents escaped quotes from ending values. Risk low: open-ended
password/token string fragments now redact their visible value suffix.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
is **not run; delegated by user**, under standing Sep 25 authority. Compilation
and static checks do not establish runtime or privacy acceptance results.

## Acceptance and evidence map

1. Source diff changes only JSON value termination plus an explanatory comment.
   Earlier header/pair stages, marker, key matching and pipeline placement stay.
   All other tracked engine paths except the existing regression file unchanged.
2. Direct scalar tests cover empty/plain/escaped-quote/backslash/dangling escape/
   partial Unicode escape/multibyte/space-containing preview-end values, mixed
   key casing, surrounding text, earlier complete values and idempotence. Adjust
   two prior out-of-scope controls explicitly; retain unrelated/malformed line,
   encoded-key and nonstring controls. Batch checks retain nil entries/order/
   metadata and change only PayloadPreview.
3. Real rule.Engine.Match to Worker.Run to writer-entry snapshots: canonical
   complete JSON payloads larger than 200 bytes yield exactly one 200-byte
   preview with no closing quote, including cuts immediately after a backslash
   and escaped quote. Verify count/content at matching and writer boundaries,
   redaction enabled/disabled and writer-failure accounting, packet preservation,
   no panic. No fake matcher for the promised truncation boundary.
4. Pinned Go 1.26.8 alert/pipeline/API compile-only chain; binaries unexecuted.
   Static format/docs/JSON/complete roadmap multisets/prior Definitions/R90-75/
   chronology/horizon/links/fences/intended paths/diff/sensitive review.
5. Seven-path feature and one docs-only closure: non-force push/fresh fetch,
   full exact ranges, Vault note/index/MOC, fourteen stable notes reconciled,
   substantive prior prose archived, immutable notes preserved, exact replay
   preserves Markdown hashes. Refresh future queue without next implementation.

## Non-goals and stop conditions

No key decoding, additional credential names, whole-JSON parsing/validation,
multiline malformed-value policy, RawPayload redaction, preview cap change,
matching/storage/API/pipeline runtime algorithm change, toolchain/dependencies,
private inputs, suites, release/tag/publication or SLO guarantee. This repair
supersedes only R90-148's historical truncated-value non-goal. Stop for competing
edits, ambiguous static/compile/Git/Vault evidence or new product/private/external
authority. Existing skill rules already cover larger-than-boundary fixtures and
returned counts; no redundant skill update. Next increment needs a new trigger.


## R90-164 Implementation and Static/Compile Checkpoint (2026-10-03)

Only the JSON value terminator changes, with one explanatory comment: accept an
actual closing quote or preview end with an optional dangling escape backslash.
Replacement keeps an actual quote and adds none for a truncated value. Earlier
escape pairs/header/pair/marker/key formatting and optional Worker placement
remain. All other tracked engine paths except the existing redactor regression
file are unchanged. Two prior truncated-value non-goal controls move into direct
redaction cases; encoded-key/nonstring/raw-linebreak controls remain.

Two direct scalar/batch regression functions are authored: 32 key/value cases
cover empty/plain/escaped quote/backslash/dangling/partial Unicode/multibyte/space-
containing preview-end values, prior complete token, exact surrounding text and
idempotence; batch checks preserve four entries/order/nil/metadata across repeats.
One real Engine-to-Worker regression has 30 cases: two keys, five cuts, three
redaction/write modes. Complete valid source payloads exceed 200 bytes, actual
Engine.Match yields exactly one exact 200-byte preview, and writer-entry snapshots
check full metadata/content/count and accounting. Boundary cuts include a dangling
backslash, escaped quote and partial Unicode escape. No fake matcher/private seam,
sleeps, skips or panic swallowing; all assertions authored/compiled only.

Pinned Go 1.26.8 alert/pipeline/API complete compile-only chain passed; binaries
unexecuted. Static source/68 other tracked engine paths/format/docs/182 task JSON/168 complete
unique roadmap pairs/167 prior Definitions/R90-75/chronology/horizon/links/fences/
seven paths/diff/sensitive review passed. Existing
skill guidance already covers actual truncation count/content; no skill change.
Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Feature and one docs-only closure remain; no next
increment started.


## R90-164 Completion and Forward Queue Refresh (2026-10-03)

Feature `e4563171ae8795a226f202c0187d49ddd69d06a6` contains exactly the seven planned paths. The
isolated implementation branch fast-forwarded freshly checked main; non-force
push and fresh fetch verified clean HEAD/origin/main/FETCH_HEAD equality.
Exact full range `18aa324a89c211a9f95d37655541ea17b4fd8a96..e4563171ae8795a226f202c0187d49ddd69d06a6` synchronized to
`04-开发迭代记录/2026-10-03-e4563171ae-CI知识同步.md`; note scope, short identifiers
resolved through Git, full index/MOC verified. Fourteen current stable notes
reconciled. Entire prior current prose archived and original topic/historical
tails preserved exactly, excluding only documented bounded generated MOC regions.
All 360 baseline immutable iteration notes preserved. Identical feature replay
preserves 408 Markdown hashes; snapshot JSON SHA-256
`5f56c27c29e3bf34ef601ead3e87ac52bf07e0df9e81fa9025798dde1e8026d5`.
Existing unique sibling local Vault supplied explicitly; no second/remote Vault.

Acceptance matches plan: only JSON terminator plus comment; 68 other tracked
engine paths unchanged. Two prior truncated-value non-goal controls move to new
redaction assertions. 32 scalar cases and batch metadata/nil/order/idempotence;
30 real Engine-to-Worker cases span two keys/five cuts/three write-redaction modes.
Complete valid sources exceed 200 bytes; actual Match count/exact 200-byte content
and writer-entry count/full-alert content prove the promised boundary in authored
source, including dangling backslash/escaped quote/partial Unicode cuts. Existing
complete-value, unrelated-key/nonstring/raw-linebreak controls remain. No runtime
execution/pass inferred. No fake matcher/private seam/sleeps/skips/panic swallowing.

Pinned Go 1.26.8 alert/pipeline/API compile-only chain and complete static review
passed: exact source/68 other paths/format/docs/182 JSON/168 unique roadmap pairs/
167 prior Definitions/R90-75/history/horizon/links/fences/seven paths/diff/sensitive.
Binaries unexecuted. Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/
acceptance execution **not run; delegated by user**. No validation deviations or
unresolved ambiguity. Existing skill rules cover truncation count/content and
historical boundaries; no redundant skill update.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; push/fresh-fetch/exact feature..closure Vault verification
before reporting, without another self-reference closure. After verification,
do not repeat feature/closure commit/push/sync. Queue refresh: no other defined
local ready increment. R90-75 remains independent asynchronous departmental
acceptance with its full contract and Oct 3–Dec 31 horizon unchanged. Next trigger
verifies fetched closure/Vault, audits fresh code/queue and persists a separate
eligible plan before edits. No next increment started.
