# R90-138: Reconcile bundle decode delivery and scope bounded sender replay

## Selection and authority

Clean fresh HEAD/origin/main/FETCH_HEAD is
`45ff1af40d7ff77b1a8b3a5a2d5680f9f83c9a75`. R90-137 feature
`90ed9bba8a8f431a1906b0923110c0e3fa8de25f` and closure have verified
exact Git ranges and generated iteration scope/index/MOC evidence. The 355-file
Vault snapshot JSON SHA-256 is
`572fec4b4cbe57bbf38dce55e2da60cd4611450ed6ce48c19e21ad502b61cbb7`.
Twelve current stable notes correctly close decoder implementation; all 308
pre-existing iteration-directory notes are captured for preservation.

The 55-commit Sep 4–Oct 2 phase review separates contracts/tooling/replay,
collector/sender/reporter/decoder admission, queue repair, historical candidate
publication and main toolchain metadata. No missing delivery record or qualifying
new departmental measurement appears. R90-75 remains the sole unfinished item,
with its full departmental dependency/window/risk/acceptance/validation/stop
contract. No local ready item exists. Select exactly R90-138, the smallest
source-grounded documentation-only queue repair. Horizon remains Oct 2–Dec 30.

Behavioral/CLI/traffic/acceptance/scanner/knowledge suites remain **not run;
delegated by user**, under the roadmap's standing development/test split.
Static source/docs/structure and exact Git/Vault review remain local duties.
Candidate publication evidence does not establish current-main behavior or SLO
acceptance. No skill change is warranted by this bounded audit.

## Scope and direct source evidence

Exactly three repository paths: this plan, matching task-state JSON and rolling
roadmap. Persist plan/state before roadmap edits. Preserve completed history and
all runtime source. Define R90-139 as planned until R90-138 delivery, then ready;
its separate implementation plan belongs to the next trigger.

`slo_sender_reconstruct._replay` opens fixture/offered/submissions ledgers using
ordinary following `Path.open("rb")` through `ExitStack`. `_Rows.read` limits
individual rows to 256 KiB plus a rejection probe; it has no cumulative captured-
byte bound or captured-row limit. `_Rows.verify` compares consumed bytes/count/
hash to inventory only after aligned EOF and correlation. There is no replay-
handle regular-file admission or before/after descriptor metadata comparison.
A retained pathname changed between decode and replay can therefore follow a
symlink, block on a FIFO, or consume additional individually bounded rows before
EOF mismatch. This is source evidence, not an executed failure or exploit.

R90-137 binds `_Bundle.decode` at its own read boundary. That handle is closed
before `_replay` reopens these three files. Private output directories and initial
source snapshots reduce exposure but do not supply replay admission or total
read bounds. Receipt validation and exact EOF digest verification already exist;
do not describe this gap as absent hashing or missing source-copy protection.

Direct flow: standalone `reconstruct` calls `_replay` only after receipt and
snapshot checks; `integrate` invokes fresh `reconstruct` and binds all four sender
inventories to the enclosing bundle. Bundle `reconcile` selects integration;
pair `compare` selects it through each side's fresh bundle. Value/shape failures
remain mismatches, OSError remains error/partial evidence, and completion/check
publication follows replay return. One matched prefix is diagnostic progress,
not a successful reconstruction. Inspect all these boundaries in R90-139.

## R90-139 proposed boundary

Keep standalone/reconstruction/integration APIs, public schema/status/exit
formats, original receipt validation and correlation rules. Before replay reads,
validate all three complete matching-key captured ledger inventories, nonnegative
bounded integer bytes/rows (reject bool), lowercase SHA-256 and available
non-following/nonblocking primitives. Open each retained ledger once read-only,
non-following and nonblocking; require regular descriptors and integer
Device/inode/size/mtime_ns/ctime_ns metadata. Reject known size disagreement
before consuming that ledger; acquire all three handles before correlation.
Fail closed on unavailable flags or metadata. Close every acquired descriptor
if a later admission, wrapping, read, verification or close fails.

Bound each read by remaining captured bytes plus a rejection probe and retain
the 256 KiB row cap. Reject over-inventory bytes or rows before correlating the
extra row; finish with exact consumed bytes/rows/SHA-256 and unchanged descriptor
metadata for each ledger. Keep one bounded row per ledger, strict UTF-8/JSON/
duplicate/nonfinite/finite-float handling and unchanged fixture/oracle/frame/
timing diagnostics for unchanged admitted bytes. Close all three inputs before
clearing success progress, returning successful replay, setting reconstruction
complete or recording the completed replay check. Preserve existing errors,
mismatches, partial artifacts and matched-prefix diagnostics without resetting
prior history or claiming whole-operation rollback. Source digests change
naturally and do not waive enclosing-bundle/pair comparability.

This binds only these replay reads to captured inventory; it cannot authenticate
sources, guarantee continuous writer exclusion, secure parent traversal or freeze
other readers. Earlier admission may intentionally supersede parsing diagnostics
for changed inputs. No schema migration, new budget option, source snapshot
refactor, receipt reopen, adapter-reader or publication change is authorized.

## Acceptance and evidence map

| Acceptance | Planned evidence |
| --- | --- |
| Prior delivery and phase authority | Clean fetched refs; R90-137 feature/closure ancestry and exact generated scope/index/MOC; 355-file baseline and 55-commit phase review |
| Distinct smallest follow-up | Direct `_Rows.read/verify`, `_replay` admission/EOF/close, reconstruct/integrate/bundle/pair caller and diagnostic review |
| Complete forward queue | R90-75 unchanged full contract; R90-138/R90-139 status/dependencies/window/risk/acceptance/required review/stop |
| Reviewable audit delivery | Docs check; all task JSON; full unique matching roadmap row/Definition multisets; ordered history/links/fences/scope/diff/sensitive additions |
| Current knowledge handoff | Reconcile all 12 stable current notes; preserve 308 immutable iteration notes; verify generated note/index/MOC and identical-range snapshot replay |

R90-139 departmental direct cases remain unexecuted: valid and space paths;
each ledger missing/directory/FIFO/symlink; missing/incomplete/mis-keyed inventory;
invalid bool/negative/out-of-range bytes or rows/hash; absent flags/metadata;
known-size mismatch before read; empty/exact/over captured bytes and rows;
short reads, same-size changed digest, growth/truncation/observed mutation and
pathname replacement; row boundary/UTF-8/malformed/deep/duplicate/nonfinite/
finite-float-overflow input; all existing fixture/oracle/frame/timing/alignment
rejections; first/second/third open/fdopen/fstat/read/close faults with previously
acquired handle cleanup; exact EOF inventory and stable metadata; no completion/
completed check on admission/read/verify/close failure; independent input byte
preservation and standalone/integrated/bundle/pair strict schema/status/partial-
artifact/source-digest compatibility. Each rejection must directly reach replay,
not a nearby snapshot or decoder test. No runtime result is claimed.

## Risks, non-goals and stop conditions

Audit risk low; future shared replay change medium. Stop on contradictory prior
delivery, ambiguous Vault discovery, schema migration, incompatible progress/error
ownership, ambiguous static review or new private/external/product authority.
Do not execute tests, traffic, acceptance, scanner or knowledge suites, change
runtime/dependencies/toolchain/workflows, publish releases or begin R90-139.
Close this audit with one docs-only delivery record and refresh the queue.

## Static review checkpoint

R90-138/R90-139 complete forward definitions and chronological selection are
persisted; R90-75 retains its independent departmental contract. Direct review
covers `_Rows` inventory timing, all three `_replay` opens/EOF/close, reconstruct
success/error publication, integrate's four-source binding and bundle/pair mode
propagation. All runtime files match the fetched baseline. Docs check, 156 task
JSON parses, 143 complete unique matching roadmap row/Definition multisets,
ordered history, links/fences, exact three-path scope, diff and sensitive additions
pass. The initial Vault verifier wrongly expected full hashes inside generated
note prose; inspection of the versioned generator established its abbreviated
range format. Review then verified full Git ancestry/endpoints, exact changed
scope and note/index/MOC. No delivery gap, runtime, test or ambiguous review
result occurred. This verification-only correction does not change scope.

Current stable handoff reconciliation and exact audit delivery remain. R90-139
is planned/unstarted until this audit is delivered; every direct behavioral case
remains unrun. Existing skills already require exact evidence and need no edit.

## Delivery and acceptance closeout

Audit `cf7376b5a3a15b17ee9d0b93a77b7c3e082cab0a` contains exactly the
three intended paths. Push/fresh fetch verified matching clean refs. Exact range
`45ff1af40d7ff77b1a8b3a5a2d5680f9f83c9a75..cf7376b5a3a15b17ee9d0b93a77b7c3e082cab0a`
has verified note/scope/index/MOC. All 12 current stable notes now describe
completed decoder authority and the distinct replay follow-up. All 308 prior
immutable iteration-directory notes remain unchanged. Identical-range replay
preserves the 356-file snapshot JSON SHA-256
`495de996ed7741ddd4e314e83b2c0ca179b439c58c82780077e1725375562399`.

Acceptance matches the evidence map: prior delivery/phase evidence, direct
replay and all consumer/diagnostic ownership review, complete forward contracts,
structural/static checks and current stable reconciliation. Every future direct
regression remains unexecuted under departmental ownership; no source review
establishes runtime rejection or measured SLO compliance. No runtime/source,
scope or ambiguous review deviation remains. Corrected generated-note verifier
assumption is recorded above; existing skills need no change.

One docs-only closure records verified audit facts. Resolve its SHA from Git,
verify push/fresh fetch and its exact Vault range before final reporting; avoid
another closure merely to persist its self-reference. R90-138 is complete;
R90-139 ready and unstarted. Next trigger verifies latest tip/closure knowledge
and persists a distinct implementation plan. Preserve R90-75's independent
contract and all delegated execution; do not repeat old deliveries/publication.
