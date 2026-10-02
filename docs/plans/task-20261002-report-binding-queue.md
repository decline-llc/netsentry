# R90-140: Reconcile sender replay delivery and scope report-source binding

## Selection and authority

Fresh clean HEAD/origin/main/FETCH_HEAD is
`a09297eb98031871466ac050482a83f69e89450f`. Both R90-139 feature/closure
exact Git/Vault ranges have verified ancestry/generated scope/note/index/MOC.
The 359-file snapshot JSON SHA-256 reproduces
`3226d11b7474915ee2170802a0e881b1f7d8df22fda9aaf3461814859fa5679c`.
Twelve current stable notes correctly close replay implementation; 312 immutable
iteration-directory notes are captured for preservation.

The 59-commit Sep 4–Oct 2 phase review distinguishes contracts/tooling/replay,
collector/sender/reporter/decoder/replay boundaries, queue repair, historical
candidate publication and main toolchain metadata. No missing delivery or new
qualifying R90-75 measurement appears. The full independent departmental R90-75
status/dependency/window/risk/acceptance/validation/stop contract remains intact.
No local ready increment exists. Select exactly R90-140, the smallest source-
grounded documentation-only queue repair; horizon remains Oct 2–Dec 30.

All behavioral/CLI/traffic/acceptance/scanner/knowledge suites remain **not run;
delegated by user** under the standing test-department split. Static source/docs/
structure and exact Git/Vault review remain local duties. Prior candidate release
evidence does not establish current-main runtime behavior or SLO acceptance.

## Scope persisted before roadmap editing

Exactly three repository paths: this plan, matching task-state JSON and rolling
roadmap. Preserve completed history and runtime source. Define R90-141 as planned
until audit delivery, then ready; its separate implementation plan belongs to the
next trigger. Do not implement its runtime change or execute tests here.

## Direct source evidence and next boundary

`slo_bundle._summary` validates report source fields/hash/string, then reopens
`adapter/observations.json` with ordinary unbounded following `Path.read_bytes`.
It compares exact raw bytes to the report's embedded UTF-8 source and its hash,
then recomputes the supplied decoded observations' summary. These existing hash/
raw-byte/semantic checks are real authority and must remain. That later read does
not compare consumed bytes to captured inventory or inspect descriptor stability,
and it can follow a replaced symlink, block on nonregular input or read additional
bytes before comparison rejects. This is static source evidence, not an executed
failure or exploit.

R90-137 binds `_Bundle.decode` at its own closed handle. R90-139 separately binds
sender replay ledgers. Neither protects this later observations read. Initial
bounded source copies and private output reduce exposure but cannot substitute
for its own bounded inventory admission. The observation snapshot has the existing
reporter 64 MiB ceiling. Other pair metadata/receipt/proof readers are outside
this proposed single read-boundary change.

Only direct `_summary` call is bundle `reconcile` via `_Bundle.check`, requiring
decoded report and observations documents. This base check precedes either optional
adapter or sender replay. Pair `compare` calls fresh `reconcile` for each side;
its default/context/adapter/sender/combined modes therefore share this boundary.
Value/shape failures remain mismatch; OSError propagates to existing non-success/
partial-output wrappers. Missing documents skip the check; a failed action is not
added to checks completed. Preserve these semantics without promising rollback.

## R90-141 proposed contract

Preserve `_summary` signature, existing report-source field/hash/string validation
and its diagnostic order before acquisition. Require a complete matching-key
captured observations entry, nonnegative signed-64-bit bytes (reject bool),
lowercase SHA-256 and captured size within the existing 64 MiB observations
ceiling. Require available nonzero integer non-following/nonblocking flags.
Admit one read-only regular-file handle, integer dev/inode/size/mtime_ns/ctime_ns
metadata and known-size equality before read; fail closed if primitives or metadata
are unavailable. Close raw descriptor on wrapping failure and wrapped handle on
all other paths. Bound raw read to captured bytes plus one rejection byte.

At EOF require exact consumed bytes/hash and unchanged descriptor metadata;
close before accepting embedded source equality, recomputing observations or
returning a successful check. Keep exact UTF-8 source/raw-byte/hash comparison and
complete original summary recomputation sequence after acquisition. No new JSON
parse, newline normalization or decoded-state mutation. A failed acquisition/read/
inventory/close cannot complete the report-source check or newly qualify optional
replay. Preserve captured source/retained artifacts, mismatch/OSError/partial-output
boundaries and public schemas/status/exit formats. Earlier admission rejection for
altered input is intentional; do not erase prior state/history or infer whole-
operation rollback. Inspect ordinary bundle, pair and both replay-option paths.

Only `_summary` runtime acquisition is in scope; use existing imports/helpers
and retain snapshot/decoder/readers/consumers/report validation/recompute logic.
No new budgets/options, schemas/dependencies/toolchains/workflows/publication.
Bundle source digest changes naturally and does not waive exact comparability.
Binding this read is not authenticity, continuous writer exclusion, parent-
traversal security or protection of every later reopen.

## Acceptance and direct evidence map

| Acceptance | Planned evidence |
| --- | --- |
| Prior delivery and phase authority | Fresh clean refs; both R90-139 exact Git/Vault ranges, 359-file baseline and 59-commit phase review |
| Distinct smallest follow-up | `_summary` validation/read/compare/recompute order versus decoder/replay; bundle check/replay eligibility and pair callers/errors |
| Complete forward queue | Unchanged R90-75 full contract; R90-140/R90-141 status/dependencies/window/risk/acceptance/required review/stop |
| Reviewable documentation delivery | Docs, 158 task JSON parses, 145 full unique matching roadmap row/Definition multisets, ordered history/links/fences/three-path scope/diff/sensitive additions and unchanged runtime |
| Current knowledge handoff | Reconcile 12 stable notes; preserve 312 immutable iteration notes; exact generated scope/note/index/MOC and identical-range snapshot replay |

R90-141 departmental direct regressions remain unrun: ordinary/space path,
missing/directory/FIFO/symlink/replacement, missing/incomplete/mis-keyed inventory,
invalid bool/negative/out-of-range/over-ceiling bytes or hash, absent flags/metadata,
known size mismatch before read, empty/exact/over captured bytes, short read,
same-size changed digest, growth/truncation/observed mutation, open/fdopen/fstat/
read/close faults and raw/wrapped cleanup, no completed check/replay qualification
on failure, independent byte preservation, existing missing/source-schema/hash/
string diagnostics, exact embedded UTF-8/raw/hash equality and mismatches, unchanged
summary recomputation and bundle/pair default/context/adapter/sender/combined
schema/status/partial-artifact/source-digest behavior. Each rejection must directly
reach `_summary`; nearby snapshot/decoder/replay tests cannot substitute. This
helper does not parse JSON, so those readers retain their own separate test duties.

## Risk, non-goals and stop conditions

Audit risk low; future shared report-binding change medium. Stop on contradictory
prior delivery, ambiguous Vault discovery, incompatible error/eligibility ownership,
required schema migration, ambiguous static review or new private/external/product
authority. Do not run suites, modify runtime, broaden reader hardening, change
release authority or begin R90-141. Complete this audit and one docs-only closure;
refresh queue without subsequent implementation. Existing skills need no edit.

## Static review checkpoint

R90-140/R90-141 complete contracts and chronologically appended selection are
persisted; R90-75 retains its independent departmental contract. Direct source
review covers `_summary` source validation/unbounded read/raw+hash comparison/
summary recomputation, its only check call, decoded-document prerequisites,
ValueError versus OSError ownership, both replay integration gating and pair's
fresh per-side reconciliation/mode/status propagation. Existing source hashing
and semantic checks are retained as authority; decoder/replay protection is
explicitly limited to its own handles. No JSON parser is proposed here.

Docs check, 158 task JSON parses, 145 complete unique matching roadmap row/
Definition multisets, ordered history/links/fences/exact three-path scope/diff/
sensitive additions pass. All runtime files match fetched baseline and the full
359-file Vault snapshot remains unchanged. No static-review or scope deviation
occurred. Every future direct acquisition/consumer regression remains unrun under
departmental ownership; no runtime rejection or SLO result is claimed. Exact
audit Git/Vault delivery and current stable reconciliation remain; no skill edit
is warranted. Do not implement R90-141 in this trigger.
