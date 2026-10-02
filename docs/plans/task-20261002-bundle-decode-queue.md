# R90-136: Reconcile reporter delivery and scope inventory-bound bundle decoding

## Selection and authority

Fresh fetched clean HEAD/origin/main/FETCH_HEAD is
`4ea9cc6111e41c6302a175d0446dbb270e265363`. R90-135 feature
`ac44f270c73938e650902c170af1e26564c6f801` and closure exact ranges have
verified note/index/MOC. The 351-file Vault snapshot JSON hash reproduces
`318b028feae906df6fc1e0aeba56040825cbd8206365a4ee12d1c3bf22d02928`.
The 51-commit Sep 4–Oct 2 review separates contracts/SLO tooling, reconstruction/
replay integration/runbook, collector/sender/reporter admission, queue repair,
historical candidate publication and main toolchain metadata. R90-135 is complete
implementation with execution delegated. Only R90-75 remains unfinished, with
its full independent departmental contract and no new qualifying measurements.
No local ready item exists. Select exactly R90-136, the smallest documentation-
only queue repair; Oct 2–Dec 30 horizon remains current.

Behavioral/CLI/traffic/acceptance/scanner/knowledge suites remain not run under
the user's standing departmental split. Static source/docs/structure and exact
Git/Vault review remain local responsibilities. Published candidate evidence
cannot prove current main behavior or SLO acceptance. Existing skills already
require this evidence discipline; no skill edit is warranted.

## Scope and concrete source evidence

Exactly three repository paths: this plan, its matching task-state JSON and
rolling roadmap. Persist plan/state before roadmap edits. Preserve completed
history and all runtime behavior. Define R90-137 as planned until R90-136 delivery;
its distinct implementation plan must be persisted on the next trigger.

`slo_bundle._Bundle.snapshot` already opens supplied sources non-following and
nonblocking, checks regular-file/observed source metadata and bounded copy,
then records retained `bytes`/`sha256`/`complete`. Its success path invokes
`self.check(... lambda: self.decode(key))`. `_Bundle.decode` subsequently reopens
retained paths using ordinary `Path.open` for JSONL or unbounded `read_text` for
JSON. JSONL has the existing 256 KiB per-row cap, but neither lane binds the
bytes it actually decodes to the recorded inventory or captures descriptor
stability. A changed/replaced/grown retained artifact can be read differently
from its inventory; this is a source observation, not an executed exploit or
failure. R90-135 intentionally changed only reporter observation admission.

The private output directory and safe initial source copy reduce exposure, but
cannot substitute for inventory-bound decode. Direct `_Bundle` consumers are
ordinary bundle reconciliation, pair original-manifest snapshots, adapter
reconstruction and sender reconstruction. `_Bundle.check` classifies ValueError/
EvidenceError as mismatch; I/O errors propagate to existing wrappers. Review all
four consumers and direct snapshot/decode flow during implementation. Other
later readers and continuously writable output are outside this bounded change;
do not claim the entire bundle is authenticated or immutable forever.

## R90-137 proposed boundary

Keep `_Bundle.decode(key)` interface and strict public inventory/schema/status
formats. Admit one read-only non-following/nonblocking regular retained-file
handle with required integer device/inode/size/mtime_ns/ctime_ns; fail closed
on unavailable primitives/metadata. Decode only a complete captured inventory
entry. Check known file size against the captured byte count before read, bound
all reads by that count plus a rejection probe, retain the 256 KiB JSONL row
limit, and compare consumed bytes/SHA-256 plus descriptor metadata at EOF.
Close before committing decoded metadata or row count to in-memory bundle state.
JSON lane must avoid unbounded following `read_text`; row lane must retain its
existing strict JSON/object/submission validation. Preserve existing finite-
float/duplicate/nonfinite/UTF-8/semantic diagnostics for unchanged admitted
bytes and each existing format's parsing semantics. Do not repair/delete input
or partial output. Admission/decode/inventory failure must never mark that decode
check completed or populate a trusted document/row count; existing mismatch/
error/partial-artifact behavior continues, without whole-operation rollback.

Exact hash/count checks bind decoded bytes to the captured inventory at this
read boundary. They do not authenticate the original source, freeze writers,
secure parent traversal or protect unrelated later reopen operations. Required
error-handling adjustments must remain confined to this boundary and compatible
with all consumers; no new externally stored metadata fields or format migration.

## Acceptance and evidence map

| Acceptance | Planned evidence |
| --- | --- |
| Prior delivery/phase authority | Clean fetched refs, both R90-135 ranges note/index/MOC and 351-file snapshot; phase-level history |
| Smallest source-grounded next increment | Snapshot inventory versus decode reopen/limits/hash/call flow; four consumer sites and error ownership |
| Complete forward contracts | Unchanged R90-75 and planned/unstarted R90-137 status/dependencies/window/risk/acceptance/validation/stop |
| Reviewable audit delivery | Docs/154 task JSON/full 141 unique roadmap row/Definition multisets/history/link/fence/diff/sensitive scope; focused Git/Vault ranges |
| Current handoff | Reconcile 12 stable notes; preserve immutable iteration notes; identical-range complete snapshot replay |

R90-137 departmental direct cases remain unexecuted: ordinary/space retained
paths, missing/directory/FIFO/symlink/replacement, unavailable flags/metadata or
negative size, incomplete/missing inventory, empty/exact/over captured bytes,
short reads, same-size different digest, growth/truncation/mutation/immediate
replacement, JSONL row boundaries, exact bytes/hash/rows, invalid UTF-8/malformed/
deep/duplicate/nonfinite/finite-float-overflow JSON, existing submission semantic
rejections, open/fdopen/read/fstat/close errors, no trusted state/check completion
on failure, independent source/output preservation, and all four consumers'
strict schema/status/partial-artifact/source-digest compatibility. Every distinct
rejection must directly reach `_Bundle.decode`; nearby source snapshot checks
cannot substitute. This audit claims no runtime or regression result.

## Risks, non-goals and stop conditions

Audit risk low; future shared decoder change medium. No runtime, tests, traffic,
acceptance, toolchain, dependency, workflow, release or new protocol change in
R90-136. R90-137 excludes other readers, quotas, authenticity and perpetual output
integrity. Stop on contradictory delivery, ambiguous Vault discovery, required
schema migration, incompatible shared consumer error handling, ambiguous review
or new private/external/product authority. Complete this audit and one docs-only
closure; refresh the queue without starting R90-137.

## Queue and static review checkpoint

R90-136/R90-137 have complete definitions and chronologically appended selection
history; R90-75 retains its existing department contract. The distinct inventory
versus decode gap and all four `_Bundle` constructors are directly reviewed;
all SLO runtime files equal the selected fetched baseline. Docs, 154 task JSON
parses, 141 full unique matching roadmap row/Definition multisets, ordered
history, local links/fences, exact three-path scope, diff and sensitive additions
pass. All 304 immutable iteration-directory notes and the 351-file current Vault
baseline are captured. No review failure, scope deviation, runtime or execution
result occurred. R90-137 remains planned/unstarted. Exact Git/Vault delivery and
current stable handoff reconciliation remain; no skill change is needed.

## Delivery and acceptance closeout

Audit `fb5e3702deebbf0bf58ba9c3db208de549b4cf1a` contains exactly the
three planned paths. Push and fresh fetch verified matching clean refs. Exact
range `4ea9cc6111e41c6302a175d0446dbb270e265363..fb5e3702deebbf0bf58ba9c3db208de549b4cf1a`
has verified note/index/MOC. All 12 current stable notes record reporter closure
and the new decode handoff. All 304 pre-existing immutable iteration-directory
notes are unchanged. Identical-range replay preserves the 352-file snapshot
JSON hash `f2f6d0dee3441a07fe701b9c7f23822d46ec9c8b9a8fcca3343674f0d0d4cb0b`.

Each audit acceptance maps to planned direct source, structural and Git/Vault
evidence. All SLO runtime files remain unchanged; this audit executes no decoder
or test and does not claim behavioral correctness. Future R90-137 direct cases
remain departmental and unrun. No scope/review deviation occurred. Final docs,
154 task JSON, 141 full roadmap pairs, source/callers, chronology/link/fence/
scope/diff/sensitive review passes. Existing skills already require the workflow;
no skill edit is needed.

One docs-only closure records verified audit facts. Resolve its SHA from Git and
verify its own push/fetch/exact Vault range; do not create another closure merely
to embed its self-referential hash. R90-136 is complete, R90-137 ready/unstarted.
Next trigger verifies latest fetched tip and closure knowledge before persisting
the separate implementation plan. R90-75 and all execution suites remain
departmental. Do not repeat old deliveries/publication or start R90-137 here.
