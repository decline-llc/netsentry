# R90-135: Admit finalized reporter observations through one bounded handle

## Selection and authority

On Oct 2 fresh clean HEAD/origin/main/FETCH_HEAD is
`66cde661c06226e5953e0ba0ed936302c8f5ed05`. R90-134 audit
`17361f8c0df8797ea6a033d5e8b3c3e18b4c7016` and closure exact ranges have
verified iteration/index/MOC evidence. The 349-file Vault snapshot JSON hash is
`b96394b22638adc2a41f8112a62b9b86b8a1d2a3d44f16c7ab2ef6a90924dd75`.
The 49-commit Sep 4–Oct 2 phase audit distinguishes contracts/SLO tooling,
reconstruction/replay/runbook, collector/sender admission, queue repairs,
historical candidate publication and main toolchain metadata. Only R90-75
(department acceptance) and R90-135 (ready) remain unfinished; each has complete
risk/dependency/window/acceptance/validation/stop contracts. No new qualifying
execution evidence appears. Select exactly R90-135; all internal dependencies
R90-134/R90-114/R90-119/R90-122/R90-123 are complete. Oct 2–Dec 30 refreshes the
90-day horizon; forecast movement does not add an eligibility gate.

The user-directed test split persists. Behavioral/CLI/traffic/acceptance/scanner/
knowledge suites are not run and remain departmental. Static AST/source/docs
checks and exact Git/Vault delivery remain agent responsibilities. Historical
R90-59 candidate execution does not establish main behavior or SLO acceptance.

## Scope and implementation boundary

Exactly six repository paths: `scripts/slo_report.py`, `docs/slo-report.md`,
`docs/slo-runbook.md`, rolling roadmap, this plan and matching task state.
Persist this plan/state before source/documentation edits. No other shared reader,
consumer, schema, dependency, toolchain, workflow or publication change.

Keep `read_observations(path)` signature/tuple and `MAX_INPUT_BYTES = 64 MiB`.
Require nonzero available `O_NOFOLLOW`/`O_NONBLOCK` flags, open once read-only,
transfer descriptor ownership to the binary stream with close on wrapping error,
and use its context manager for all admission/read failures. Require regular
file and integer device/inode/size/mtime_ns/ctime_ns metadata; fail closed if
unavailable. Reject known over-cap size before reading with the existing cap
message. Read at most cap+1 bytes, retain existing over-cap diagnostic, compare
required same-descriptor metadata before/after read and consumed size to admitted
size. Close before the unchanged UTF-8/JSON decode block and return. No source
write or path reopen occurs. Standalone admission fails before summary or output
creation; preserve source and existing output. Known-size admission and observed
mutation rejection intentionally precede JSON diagnostics. Otherwise keep the
existing complete JSON parsing, duplicate/nonfinite diagnostics and semantic
validation order unchanged, including top-level semantic rejection at summarize.

Direct callers are `slo_report.main`, `slo_reconstruct.reconstruct` (rebuilt
observations) and `slo_compare._conditions` (retained observations). Their wrappers
already classify ValueError/EvidenceError and OSError as error/mismatch outcomes;
review those catches and retain their prior partial artifacts. Whole-operation
rollback for shared consumers is not promised. Their reporter source digest
changes naturally; no schema or digest binding relaxation is allowed.

Metadata observes differences at read boundaries. It does not authenticate
content, freeze writers, establish parent traversal safety or independently
certify measurement facts. Same-handle reads do not reopen a replacement path;
metadata-visible immediate replacement/change must fail closed. A successful
read does not promise perpetual source stability.

## Acceptance and evidence map

| Acceptance | Planned evidence |
| --- | --- |
| One bounded, regular-file acquisition | Static flag/open/ownership/regular metadata/known-size/read-limit/EOF-size/post-metadata/context-close flow review |
| Preserve reporter formats and diagnostics | AST compare every other original definition/constant, reader signature and complete original decode block; no schema/status/threshold/writer edits |
| Rejection does not publish standalone report | Unchanged main AST read-before-summary-before-write; source opens read-only; all acquisition/close failures precede decode/return |
| Shared caller compatibility | All three direct calls and reconstruction/comparison wrapper error classification and source-digest bindings inspected |
| Honest delivery and handoff | Python AST/docs/153 task JSON/full 139 unique roadmap multiset/history/link/fence/diff/sensitive scope checks; exact Git/Vault ranges and stable-note reconciliation/replay |

Departmental direct cases remain unexecuted: ordinary/space paths; missing,
directory, FIFO, symlink; unavailable flags/metadata and negative size; empty/exact-cap/known-over-
cap/read-over-cap; mutation/growth/truncation/immediate pathname replacement;
open/fdopen/read/fstat/close failure; exact source bytes/digest; UTF-8/malformed/
deep/duplicate/nonfinite JSON and existing semantic diagnostics; no-overwrite and
independent read-only source/output byte preservation; shared reconstructed/pair
inputs, source-digest identity and all error/status classifications. Each distinct
rejection needs a direct reader regression; nearby retained-snapshot coverage
cannot substitute. No executed rejection, compatibility or scale result claimed.

## Risks, non-goals and stop conditions

Risk medium: the shared reader adds intentional admission failures before decode,
and reporter source identities change. All existing data schemas and successful
summary/publication behavior remain intact. Do not promise authenticity,
continuous mutation exclusion or whole-operation rollback. Stop on required
schema migration, inconsistent shared caller contracts, ambiguous validation,
new external/private/product authority or unsupported primitives without safe
fail-closed handling. No tests, CLI invocation, traffic, acceptance, release,
new budget option or next increment. Existing skills already capture the required
boundary/evidence discipline; no skill edit warranted. Deliver one feature and
one docs-only closure for this same increment, then refresh without starting
another item.

## Implementation and static checkpoint

Source matches the planned reporter-local admission. Flags fail closed before
open; fdopen failures close the raw descriptor; context-managed regular metadata,
nonnegative/known-size admission, cap+1 read and EOF metadata/consumed-size checks
precede source close, unchanged decode and return. No source/output write is
performed by acquisition. The original main read/summary/writer order and catches
are unchanged. Reconstruction preserves mismatch for EvidenceError/ValueError,
error for OSError and partial rebuilt files; comparison's CLI catches either
without claiming a completed comparison, preserving existing retained files.
Successful schemas/status/source bindings remain intact; source digests change
naturally and are never equated to old tool identities.

Static AST equality proves all other original definitions/constants, reader
signature/tuple annotation and the complete decode block unchanged. The three
call sites and unchanged bundle/reconstruction/comparison source were reviewed.
Python AST, docs, 153 task JSON parses, 139 full unique row/Definition multisets,
ordered history, links/fences, six-path scope, diff and sensitive additions pass.
Initial read-only searches named absent test paths; corrected repository-file
search confirmed no existing SLO test suite to invoke. No code/runtime validation
sequence was affected and no behavior was executed. All departmental cases
remain unrun, including each added admission rejection, negative/unavailable
metadata, wrapping/close failures and shared consumer outcomes. No scope change
or skill improvement is warranted. Exact feature Git/Vault delivery remains.
