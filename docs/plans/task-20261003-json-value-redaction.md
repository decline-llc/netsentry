# R90-148: Redact complete escaped JSON credential values

## Selection and authority

Fresh fetched clean main/HEAD/origin/main/FETCH_HEAD is
`2fb0cfd10e24c9927df060b386f9a1fa536b6a0e`. R90-147 feature/closure exact
Git/Vault ranges, six/three-path generated scope, notes, index and MOC are
verified. The 375-file snapshot JSON SHA-256 is
`ed6d777484b86d9e01be2ade47c339a767079b05acee365805e1ed18ab630ac1`;
14 stable full-content backups and 328 immutable iteration hashes are captured.
The 75-commit Sep 5–Oct 3 phase audit covers measurement/reconstruction, bounded
evidence inputs, repeated queue reconciliation, historical publication/toolchain
and the recent rule/pipeline/suppression/receiver repairs. No missing delivery
or new qualifying R90-75 measurement appears. Today advances the active 90-day
horizon to Oct 3–Dec 31; completed historical dates remain unchanged.

The defined local queue is empty. Source redactor's JSON string value pattern
excludes every quote without respecting escapes, so it can stop at an escaped
quote and leave a credential suffix in the alert preview. Select bounded privacy
correctness repair R90-148 after R90-147 and existing optional pipeline redaction.
Register it inside this implementation increment, not a separate queue audit.
Persist plan/state before runtime or roadmap edits.

R90-75 keeps its full independent asynchronous departmental acceptance contract.
Behavioral/race/CLI/full-suite/scanner/knowledge/acceptance execution remains
**not run; delegated by user**. Author direct regressions and perform static/
compile-only checks; never execute test binaries. Engine module preflight resolves
pinned Go 1.26.8. Existing skills cover direct boundary evidence and knowledge
preservation; no skill edit is warranted at selection.

## Scope and behavior

Exactly eight paths: `engine/internal/alert/redactor.go`, new
`engine/internal/alert/redactor_json_test.go`, new
`engine/internal/pipeline/redaction_json_test.go`, `docs/architecture.md`,
`docs/api-reference.md`, rolling roadmap, this plan and matching task state.
Isolated local branch: `fix/r90-148-json-value-redaction`.

Change only sensitiveJSONRe's value matcher: a backslash plus any non-CR/LF
character forms one escape pair; ordinary value characters exclude unescaped
quote, backslash and literal CR/LF. The closing quote is therefore unescaped.
Keep the two existing capture groups and replacement expression so the original
key, whitespace, colon and surrounding quotes remain. Literal password/token
keys retain existing case-insensitive matching. Empty strings and repeated/nested
occurrences remain supported without decoding or reserializing the preview.
Header/pair expressions, replacement order, batch nil/empty behavior, `[REDACTED]`
marker, pipeline/config/API/schema and all other runtime code are unchanged.

This is best-effort lexical redaction of complete recognized string values, not
a JSON validator or exhaustive sanitizer. Unknown escaped characters are treated
as lexical pairs; escaped property names, extra sensitive keys, non-string values,
truncated/unterminated fields and literal-newline malformed values gain no new
guarantee. Matching a complete object is unnecessary: a complete recognized value
inside a payload preview can be redacted. Existing header and pair stages retain
their behavior, including their interaction with surrounding payload text.

## Acceptance and direct evidence map

| Acceptance | Local static/compile evidence | Direct departmental regressions, unrun |
| --- | --- | --- |
| Baseline, selection and horizon | Fresh refs, exact prior Git/Vault, phase/queue audit; new Oct 3–Dec 31 horizon | R90-75 independent |
| Entire escaped string value removed | Single regex substitution; unchanged capture/replacement and all other source | Scalar public RedactSensitivePayload: escaped quotes/backslashes, quote parity, encoded controls/Unicode/slash, empty/simple values and case-insensitive keys; exact output, valid JSON and idempotence |
| Surrounding formatting and multiple values | Prefix/closing captures and ReplaceAllString unchanged | Nested/repeated properties and adjacent public fields remain exact; complete recognized values in request/body fragments redacted |
| Batch behavior | RedactSensitivePayloads unchanged | Public batch with nil/empty alerts, multiple escaped credential previews and preserved alert metadata |
| Before writer boundary, switch unchanged | Worker redactor remains before WriteBatch; Worker/config source unchanged | Public Worker.Run with real RedactSensitivePayloads, writer copies payloads at invocation: enabled success/failure redacted before writer, disabled success unchanged; original packet/alert metadata and counters preserved |
| Non-goals and compatibility | Header/pair expressions/order unchanged; no whole-preview serialization | Ordinary header/pair redaction and unrelated/non-string fields preserved in bounded fixtures; unterminated/literal-newline values have no claimed redaction guarantee |
| Delivery and scope | Go parse/format; pinned alert/pipeline/API compile-only; docs/166 JSON/152 full unique roadmap/full forward contract/R90-75/ordered history/links/fences/eight-path/diff/sensitive review; exact feature/closure Git/Vault/stable topics/immutable/replay | All execution suites delegated |

Tests use synthetic fixture strings only. Scalar tests call the public redactor,
batch tests reach its public wrapper, and pipeline tests invoke the real redactor
through Worker.Run and capture immutable strings during WriteBatch. Holding alert
pointers until after Run is not proof of pre-write redaction. Static and compilation
evidence does not prove behavioral, leakage, race or acceptance outcomes.

## Risk and stops

Medium privacy/correctness risk: accidentally consuming an adjacent property or
overstating malformed-preview coverage. Preserve exact prefix/suffix formatting
and demonstrate quote/backslash parity with direct fixtures. Exclude key decoding,
new sensitive field lists, header/pair redesign, full JSON parsing/reserialization,
malformed/truncated fail-closed policy, switch/schema/API/storage/matcher/metrics/
shutdown changes, IPv6/dependency/toolchain/test execution/external publication.
Stop for required product/privacy policy, incompatible formatting/replacement,
ambiguous lexical scope, competing edits or new external authority.

Complete exactly R90-148. On successful review, commit eight paths, fresh-verify
baseline, fast-forward main, push/fetch and sync the exact range. Preserve stable
topic prose and immutable iteration history. One three-path docs-only closure
records verified delivery and receives exact Git/Vault verification; refresh
the queue without starting another implementation.


## Implementation, static deviation and final checkpoint

Exact source comparison permits only sensitiveJSONRe's string-value expression
replacement. Escape pairs consume backslash plus non-CRLF; ordinary characters
exclude quote/backslash/CRLF, so the terminator is unescaped. Prefix and closing
capture groups plus replacement remain; every public redactor/header/pair/batch,
Worker/config/main/API/Stats/store/suppression/receiver/rule/exporter body is
unchanged. No whole-preview decoding or serialization is introduced.

Five direct regression functions contain 32 key/value combinations, nine raw
lexical parity/Unicode/slash/nested/repeated/formatting/fragment cases, batch nil/
empty/metadata preservation, header/pair compatibility and bounded non-goal cases.
Three Worker.Run cases use the real batch redactor and copy model.Alert values
inside WriteBatch before return: enabled success/error redact before writer
entry, disabled success retains original preview. All alert metadata, original
packet value and original success/error/completion accounting are asserted.
No post-return pointer observation substitutes for the pre-write boundary.

Pinned Go 1.26.8 complete alert → pipeline → API compile-only chain passed;
binaries remain outside the repository and unexecuted. No behavioral/race/CLI/
full-suite/scanner/knowledge/acceptance suite was invoked, and no runtime leakage
or SLO claim follows from compilation/static review.

The first completed-Definition equality check falsely included the adjacent
historical level-three subsection in the last prior Definition. Adding the new
Definition appeared to change old authority even though that subsection was
unchanged. The external static checker now scopes each body at its structurally
external level-two/three heading boundary. Full static/docs/diff rerun confirms
all prior Definition bodies, including R90-75, unchanged. This is a recorded
validation-tool deviation, not an implementation/compile failure; no runtime
source edit was needed to resolve it. The generic local skill structural-review
instruction now confirms comparison boundaries exclude neighboring historical
subsections before claiming a section changed. Markdown/frontmatter/numbering/
fences pass; skill edit stays separate from repository feature delivery.

Final Go parse/format/docs, 166 task JSON, 152 full unique roadmap pairs, full
forward contracts, unchanged prior Definitions/R90-75, ordered history, active
Oct 3–Dec 31 horizon, links/fences, eight-path scope, diff and sensitive review
pass. The full 375-file Vault baseline is unchanged. Full stable-content backups
support topical preservation and 328 immutable hashes remain captured. Feature
plus one docs-only closure delivery remain; no subsequent increment starts here.


## Delivery and queue closeout

Feature `d5addcb22226404e2d6a43c3e0df800fde3a646b` contains exactly eight planned paths. Isolated branch
`fix/r90-148-json-value-redaction` was fast-forwarded to freshly verified main;
push/fresh-fetch verified clean HEAD/origin/main/FETCH_HEAD at that SHA.
Exact range `2fb0cfd10e24c9927df060b386f9a1fa536b6a0e..d5addcb22226404e2d6a43c3e0df800fde3a646b` has eight-path generated scope, iteration
note `04-开发迭代记录/2026-10-03-d5addcb222-CI知识同步.md`, full index and MOC
verified. Fourteen stable status sections are current for Oct 3; every original
substantive topic tail is retained exactly, and all 328 baseline immutable
iteration hashes are unchanged. Identical replay preserves the 376-file Markdown
snapshot; snapshot JSON SHA-256 is `d362bbd1c5bf7a73e1e82382c5ade51c5b7d2c6939b4421797684a58e3c37ff9`. The unique existing
local sibling Vault was selected explicitly.

Acceptance matches the plan: only sensitiveJSONRe's string-value expression
changes, consuming escape pairs until an unescaped closing quote. The same two
captures/replacement preserve keys/whitespace/colon/quotes. Header/pair/batch,
Worker/config/main/API/schema/store/receiver/rule/Stats/exporter source remains.
Literal case-insensitive password/token keys, optional pre-write invocation and
marker remain. Best-effort preview handling does not add key decoding, a full
JSON sanitizer or malformed/truncated-value fail-closed policy.

Five direct regression functions contain 32 generated key/value combinations,
nine raw lexical/parity/format/nested/repeated/fragment cases, exact output/JSON/
idempotence/canaries, batch nil/empty/order/metadata, header/pair/non-goal fixtures,
and three public Worker.Run cases using the real redactor. Writer copies complete
Alert values at entry before returning success/error; enabled paths are redacted
and disabled path remains original, with packet/metadata/counters preserved.
Pinned Go 1.26.8 complete alert/pipeline/API compile-only chain and final static
source/direct-boundary/Go-format/docs/166 JSON/152 full unique roadmap/full forward
contract/all prior Definitions/R90-75/history/horizon/links/fences/eight-path/
diff/sensitive checks pass. No binary or behavioral/race/CLI/full-suite/scanner/
knowledge/acceptance suite was executed; all remain user-delegated, without a
runtime leakage, race or SLO outcome claim.

The first static Definition comparison falsely included an adjacent historical
level-three subsection. Checker boundaries were corrected; complete static/docs/
diff rerun proved prior Definition bodies unchanged. This recorded validation-
tool deviation required no runtime source edit; there was no implementation or
compile failure. Generic local skill structural review now verifies boundaries
exclude neighboring historical subsections before claiming section changes;
Markdown frontmatter/numbering/fences pass and skill edit is separate from Git.
The other planning adjustments are source-grounded empty-queue repair inside
this increment and active 90-day horizon advancing to Oct 3–Dec 31; completed
history remains. No delivery failure or Vault topic loss occurred.

One three-path docs-only record closes this same increment. Resolve its full SHA
from Git, push/fresh-fetch and verify exact feature-tip..closure-tip Vault before
reporting; no extra self-reference closure. Queue refresh: R90-148 implementation
is complete and no further dependency-ready local item is currently defined.
R90-75 retains its full independent asynchronous departmental contract and does
not block development. Next trigger verifies fetched closure/Vault, audits fresh
core-code evidence and forward queue, then persists a separate eligible plan.
No subsequent implementation starts here. Oct 3–Dec 31 remains current; IPv6/
external publication remain separate authority. Do not repeat completed
R90-147/R90-148 delivery or R90-59 publication.
