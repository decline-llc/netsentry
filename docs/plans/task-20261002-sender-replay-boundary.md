# R90-139: Bound retained sender replay against captured inventory

## Selection and authority

Clean fresh HEAD/origin/main/FETCH_HEAD is
`d05bcefb85476bbaf9b451484a203874290d2cf2`. Both R90-138 audit/closure
exact Git ranges have verified ancestry/generated scope/note/index/MOC. The
357-file Vault snapshot JSON SHA-256 reproduces
`8b7e29acbf657ae304b98799385db22077cd9d3f6b1b297dd5914660c50e8fff`.
Twelve current stable notes close R90-138 and identify ready/unstarted R90-139;
310 immutable iteration-directory notes are captured for preservation.

The 57-commit Sep 4–Oct 2 phase review distinguishes contracts/tooling/replay,
collector/sender/reporter/decoder admission, queue repair, historical candidate
publication and main toolchain metadata. No missing delivery or qualifying new
R90-75 measurement appears. R90-139 dependencies R90-138/R90-126/R90-127/R90-137
are complete. The independent R90-75 departmental contract retains its full
status/dependency/window/risk/acceptance/validation/stop boundary. Horizon remains
Oct 2–Dec 30; dates are forecasts. Select exactly this implementation increment.

The user's standing test-department split applies. All behavioral/CLI/traffic/
acceptance/scanner/knowledge execution remains **not run; delegated by user**.
Static source/AST/docs/structure and Git/Vault evidence remain local duties.
No runtime correctness or SLO acceptance will be inferred from static review.

## Scope persisted before implementation

Exactly six repository paths: `scripts/slo_sender_reconstruct.py`,
`docs/slo-sender-reconstruct.md`, `docs/slo-runbook.md`, rolling roadmap, this plan
and matching task-state JSON. Persist plan/state before source/docs edits.
Runtime changes are confined to `_Rows` and `_replay`, plus required `stat`
import. Preserve `_correlate`, receipt/schema checks, reconstruct/integrate/main,
public API/schema/status/exit/partial-artifact behavior and source-digest binding.
Do not change bundle decoder/snapshot, adapter/report readers, budgets/options,
dependencies/toolchain/workflows or release artifacts. No next increment starts.

## Behavior and ownership

Validate all three complete matching-key inventories before any replay read:
nonnegative signed-64-bit bytes/rows with bool rejected and lowercase SHA-256.
Capture the validated primitive values for the replay. Require available nonzero
integer O_NOFOLLOW/O_NONBLOCK. Admit each ledger once with read-only flags;
close raw descriptor if wrapping fails, register each wrapped handle with
ExitStack before metadata/reader initialization and before acquiring the next
handle. Require a regular descriptor and integer dev/inode/size/mtime_ns/ctime_ns;
reject known size disagreement before that ledger read. Admit all three before
any correlation. Stack cleanup must attempt all registered closes even when a
later admission/read/verification/close raises.

Each `_Rows.read` uses min(256 KiB, remaining captured bytes)+1. Reject excess
bytes and captured row count before decoding/correlating an extra row; preserve
original UTF-8/strict JSON/object parsing sequence for unchanged admitted bytes.
Maintain at most one bounded row per ledger and existing matched-prefix progress.
At aligned EOF require stable descriptor metadata and exact consumed bytes/rows/
SHA-256 for each. Close all handles before clearing success progress or returning
to reconstruction completion/completed-check publication. Empty ledgers retain
the existing rejection. Semantic/shape failure remains mismatch; OSError remains
error/partial evidence. Earlier admission can intentionally supersede parsing
errors for changed input; no whole-operation rollback or input repair is promised.

Receipt and all fixture/oracle/frame/timing/alignment rules stay unchanged.
Standalone reconstruct -> integrate -> bundle -> pair consumers retain formats,
mode selection, completeness/error precedence and four-source binding. The
reconstruction source digest changes naturally and never waives comparability.
Read/hash/metadata checks bind these handles only; they do not authenticate
sources, secure parent traversal, freeze writers or protect unrelated reopens.

## Acceptance and direct evidence map

| Acceptance | Planned local evidence | Departmental direct cases, unrun |
| --- | --- | --- |
| Correct ready selection | Fresh clean refs, both R90-138 ranges, complete forward contracts and 57-commit phase review | Independent R90-75 acceptance |
| Inventory and admission before correlation | Source/AST order for all inventories, flags, opens, stack registration, regular/integer metadata and known size | Each ledger missing/directory/FIFO/symlink/space path; incomplete/missing/mis-keyed inventory; bool/negative/out-of-range bytes/rows/hash; missing flags/metadata and size mismatch before read |
| Bounded replay and exact verification | Read cap/probe, byte/row rejection before extra decode/correlation, original JSON sequence, EOF metadata/bytes/rows/hash | Empty/exact/over bytes/rows, short reads, row boundaries, same-size changed digest, growth/truncation/mutation/replacement; malformed/deep/duplicate/nonfinite/overflow/UTF-8 JSON |
| Cleanup and success ordering | Wrapping cleanup and stack ownership; all admissions before loop; close before cleared progress/return and caller complete/check | First/second/third open/fdopen/fstat/read/close faults; cleanup of earlier handles; no completion/check on failure; independent source byte preservation |
| Correlation and consumers preserved | AST equality for all other definitions/constants; direct reconstruct/integrate/bundle/pair error/schema/status/digest review | Every original fixture/oracle/frame/timing/alignment rejection; standalone/integrated/bundle/pair compatibility and partial evidence |
| Reviewable delivery and knowledge | Python/docs/static JSON/full unique roadmap/history/link/fence/scope/diff/sensitive review; exact commits/push/fetch/Vault ranges; 12 stable notes/310 immutable notes/replay snapshot | Behavioral, CLI, traffic, acceptance, scanner and knowledge suites remain delegated |

Every rejection must directly reach replay, not a nearby snapshot/decoder test.
The department owns these unexecuted regressions; static evidence proves source
structure only. Do not claim a test pass, runtime rejection or measured scale.

## Risk, non-goals and stop conditions

Medium risk: internal row-reader changes may alter diagnostic order or handle
ownership. Mitigate with captured primitive inventory, stack registration before
initialization, bounded local counters and source/AST review. No schema migration,
receipt reopen/snapshot refactor, adapter-reader hardening, new quota/CLI option,
authenticity/continuous integrity or external publication is in scope. Stop on
contradictory Git/Vault, ambiguous discovery, incompatible error/progress ownership,
required format migration, ambiguous static review or new private/external/product
authority. Close this increment with one docs-only delivery record; refresh but
do not begin any subsequent increment. Existing skills need no change so far.

## Implementation and static review checkpoint

Implemented the planned boundary in `_Rows`/`_replay` with one stat import.
All inventory primitives and flags are validated before replay opens; stack
registration precedes reader initialization and later acquisitions. Regular/
integer metadata and known-size checks occur before reads. Byte/row probes
precede new row decoding/correlation, and aligned EOF checks metadata then exact
bytes/count/hash. Captured metadata/inventory are primitive tuples. Success
progress clearing and caller complete/check publication follow stack exit.
Local standard-library ExitStack source confirms it continues registered cleanup
callbacks when a prior callback raises, then propagates failure. This is source
inspection, not fault injection or execution evidence.

AST comparison confirms all other module definitions/imports/constants (except
new stat import), `_replay` signature, full original correlation loop and JSON
parsing/counter sequence unchanged. Direct review covers standalone reconstruct,
integrate's four-source binding, bundle shared error precedence, pair strict
modes/qualifying-side logic and naturally changed reconstruction source digest;
consumer files are unchanged. Python/docs checks, 157 task JSON parses, 143 full
unique matching roadmap row/Definition multisets, ordered history/links/fences,
exact six-path scope, diff and sensitive additions pass. No static-review failure
or scope deviation occurred. Earlier admission/rejection boundaries and new source
digest are planned effects. Every named direct behavioral regression remains
unrun under departmental ownership; static review cannot establish runtime
correctness, cleanup under injected failure, preservation, performance or SLO
acceptance. Existing skill instructions already cover the workflow; no edit needed.

Feature Git/Vault delivery, current stable reconciliation and one docs-only
closure remain. Do not begin a subsequent increment.
