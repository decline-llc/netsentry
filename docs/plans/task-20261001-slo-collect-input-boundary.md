# R90-129: Bound standalone SLO adapter input admission and retention

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD baseline is
`a88476ced2b9ab2409a5a89aa17aa47d56dafd09`. R90-128 feature/closure exact
Vault notes/index/MOC and nine stable notes are verified; closure-range replay
preserved Vault Markdown hash
`6df06b8293443cb19e4f1fb497c59548f7cb4450c4cfce96104fbb7b8e5214f4`.
The Oct 1 closure has no unresolved discrepancy. Its 32-commit phase audit,
144 prior task states and 131 roadmap row/Definition pairs were verified.
Current post-audit state has 145 task JSONs and 133 unique roadmap pairs.
Horizon is Oct 1–Dec 29. R90-59 release authority and R90-75 departmental
acceptance remain independent. User testing delegation is still active.

## Frozen API and behavior

Add `DEFAULT_MAX_INPUT_BYTES = 64 * 1024**3` and
`collect(manifest, offered, events, output, scratch_dir=None, *,
max_input_bytes=DEFAULT_MAX_INPUT_BYTES)`. CLI adds `--max-input-bytes` with
that default. Require a positive integer using the existing report integer
contract even for empty/small sources. Existing positional parameters, receipt
schema, observation schema, and successful parse semantics remain unchanged.
The one budget covers the exact retained source bytes of manifest.json,
offered.jsonl, and events.jsonl in aggregate; it excludes generated observations,
receipt metadata, retained copies' filesystem overhead and SQLite temporary data.
Manifest remains capped at 64 MiB; each JSONL row at 256 KiB.

Open all three source files before creating output or reading input with
`O_RDONLY | O_NOFOLLOW | O_NONBLOCK`; use `fstat` on each descriptor and admit
only regular files. Keep and read those same handles. Missing, symlink, directory,
FIFO, device and other nonregular inputs fail before any source read/output
creation. Existing finalized-file caller contract stands; no new path-parent,
mount, authenticity or concurrent-hostile-mutation guarantee is claimed.

Account exact bytes once as manifest bytes are read and ledger rows are consumed.
Never consume/store bytes beyond the remaining budget: bound each read to the
minimum of the row ceiling and remaining budget plus one detection byte; if that
byte exists, fail before retaining it. At EOF compare descriptor device, inode,
size, mtime_ns and ctime_ns with the pre-read snapshot; changed or unavailable
metadata fails closed. Ledger receipts are appended only after full EOF and
metadata verification. The public completion receipt remains the last artifact
published. On any failure retain already-written partial files, publish no
completed receipt, leave sources unchanged, and preserve current no-overwrite
output behavior.

Keep the existing `_rows` helper unchanged for `slo_ingress.send_fixture`.
Implement adapter-specific streamed row reading over already admitted handles.
`slo_reconstruct.reconstruct` must pass its explicit `max_bytes` through as the
collector input budget, including explicitly larger values; no hidden second
64 GiB cap. Bundle callers keep using their existing outer source snapshot and
reconstruction budget. No global reporter-reader policy changes.

## Acceptance evidence map

| Acceptance | Source/static review | Departmental validation, not run |
| --- | --- | --- |
| Admission | Three descriptors opened non-following/nonblocking and verified regular before output/read | ordinary paths, spaces, missing/directory/FIFO/symlink/device inputs; no source read or output on failed admission |
| Aggregate limit | One counter charges manifest and both JSONL streams exactly once; bounded plus-one read detects overage before retain | zero/invalid, exact, one-over, manifest-near-limit and both cross-file budget exhaustion positions |
| Stream/schema compatibility | Existing duplicate/nonfinite/UTF-8/object/256 KiB and row-specific validators remain; legacy `_rows` caller is untouched | successful legacy receipt/observation bytes, malformed and oversized rows, sender fixture compatibility |
| Stable source | Same opened descriptor retained; before/after dev/inode/size/mtime/ctime compared; source never writable | changed size/content/timestamp, metadata-unavailable behavior and exact source-byte preservation |
| Completion/partial state | Per-ledger receipt only after EOF/stat; overall receipt last; no-overwrite retained | parse/I/O/short-write/sync/close/interruption failures and partial output with no completion receipt |
| Caller budgets | Reconstruction forwards explicit configured max; bundle outer budget behavior unchanged | standalone defaults, custom smaller/larger limits and standalone/bundle/pair replay propagation |
| Delivery | AST/python-check/docs-check/JSON/roadmap/link/scope/diff/sensitive review plus verified Git/Vault | Behavioral tests and knowledge suites remain delegated and unrun |

## Scope and stop conditions

Eight intended paths: `scripts/slo_collect.py`, `scripts/slo_reconstruct.py`,
`docs/slo-collect.md`, `docs/slo-reconstruct.md`,
`docs/slo-runbook.md`, roadmap, this plan and task state. No live sender changes,
traffic, test execution, dependency, protocol, global reader policy, worktree-wide
quota, performance claim, CI, private input or release action. No package imports
or CLI smoke under user direction.

Stop if strict regular-file admission conflicts with a documented supported input,
source metadata cannot be obtained for the promised stability check, receipt or
caller compatibility requires a new product choice, or explicit larger replay
budgets cannot propagate without changing an external format. Preserve partial
outputs, state diagnostics and do not claim scale/runtime evidence. No following
increment is started in this trigger.
