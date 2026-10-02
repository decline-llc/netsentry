# Supplied SLO evidence bundle

R90-118 adds `scripts/slo_bundle.py`, a standard-library API/CLI for the
[native UDP measurement lane](slo-ingress.md), [engine export](slo-runtime.md),
[ledger adapter](slo-collect.md) and [reporter](slo-report.md). It snapshots
supplied artifacts and checks their consistency. It does not run traffic or
services, execute tests, or certify either proposed SLO profile. Implementation
has received static review only; behavioral validation is delegated by the user.

## Department invocation

After stopping writers and retaining process exit statuses, supply existing
files. The following is an unexecuted invocation template, not run evidence:

```bash
python3 scripts/slo_bundle.py \
  --sender sender-run \
  --capture capture-summary.json \
  --engine engine-run \
  --adapter retained-run \
  --summary acceptance_staging_YYYYMMDD_HHMMSS.json \
  --output-dir review-bundle
```

The public API is `reconcile(output, *, sender=None, capture=None, engine=None,
adapter=None, summary=None, max_bytes=64*1024**3, reconstruct_ledgers=False,
max_reconstruction_bytes=64*1024**3, scratch_dir=None, reconstruct_sender=False,
max_sender_reconstruction_bytes=64*1024**3)`. Directory/path arguments are
`pathlib.Path` values. Each input is optional so a partial handoff can publish
explicit gaps; omitted inputs never produce a completed bundle. Output must be
a new directory in an existing parent. No existing file/directory is replaced
and no run evidence is deleted. This tool does not change the acceptance report
path `~/benchmarks/r90‑75/acceptance_{staging|prod}_{YYYYMMDD_HHMMSS}.json`.

## Retained sources and checks

| Input | Fixed retained paths | Checks |
| --- | --- | --- |
| Sender directory | `sender/submission.json`, `fixture.jsonl`, `offered.jsonl`, `submissions.jsonl` under `sender/` | Schema, completed submission declaration, source digest syntax, link/port types, exact inventory, hashes/bytes/rows, nonempty equal ledger row counts; each submission declares success and ordered scheduling/send times |
| Capture file | `capture/summary.json` | Closed component, inbound host/microsecond timestamp declaration, port and counters; pcap stats validity retained |
| Engine directory | `engine/close.json`, `engine/events.jsonl` | Closed export, origin, timestamp/durable/event-ID boundary declarations, raw event digest/bytes/rows |
| Adapter directory | `adapter/receipt.json`, `manifest.json`, `offered.jsonl`, `events.jsonl`, `observations.json` under `adapter/` | Receipt schema, adapter completion, inventory, hashes/bytes/rows, manifest/observation schemas and matching non-derived metadata |
| Report file | `report/report.json` | Exact embedded observation bytes/digest; complete summary recomputed with the current reporter, including missing-alert counts, p99, deadline violations and measurement status |

All valid component run IDs must agree, as must sender/engine origins and the
manifest/observation start time. Capture has no origin field. Sender destination
port must match capture. Sender offers must be byte-identical to adapter offers;
engine events must be byte-identical to adapter events. In this UDP bundle lane,
`provenance.fixture_sha256` must identify the exact retained sender `fixture.jsonl`
bytes. Freeze that meaning before collection; a digest of a different fixture
representation requires a separately specified format, not silent substitution.
Report run ID/origin/profile/provenance are checked by full summary comparison.

Only documented fixed basenames are opened; receipt-supplied paths are never
followed. Duplicate JSON members, nonfinite constants, malformed JSON/JSONL,
non-object rows, wrong receipt fields/types, repeated or unknown inventory names,
invalid hashes and divergent content are mismatches. JSON equality distinguishes
booleans, integers and floating-point representations. Source-code SHA-256 values
are syntax-checked declarations; matching them to retained source commits requires
external review. Rules/configuration files are not included by this tool.

## Status and completion semantics

| Status / exit | Meaning |
| --- | --- |
| `review_required` / 0 | Every required snapshot and implemented cross-check completed consistently; department review remains mandatory |
| `mismatch` / 1 | At least one malformed, changed or inconsistent input, including a false/invalid component completion assertion; gaps are retained too |
| error / 2 | Invalid CLI/API arguments or output I/O failure; retain partial files and inspect process outcome |
| `incomplete` / 3 | No detected mismatch, but inputs are absent/unreadable, byte limits prevent full snapshots, or required cross-checks cannot run |

Mismatches take precedence over gaps. `bundle.json` contains both lists, checked
and skipped boundaries, declared run IDs/origins, report status and retained
inventory with exact byte hashes and the bundle/reporter source-code digests. `bundle_complete` and `cross_checks_complete`
are true only for `review_required`. `snapshots_complete` means all source bytes
were copied, not that they were valid. `reported_measurement_status` is the supplied
report claim; trust it only after the report comparison appears in completed checks.
A consistent report can itself say `inconclusive` or `measurement_failed`; bundle
status never upgrades that result. Every bundle keeps `execution_complete: false`,
`slo_compliance_asserted: false` and `departmental_review_required: true`.

Sender submission, capture close and engine export close are component boundaries.
The adapter manifest's run/drain completion remains a departmental assertion,
validated for schema and consistency only. Component receipts do not record all
process exit/fsync outcomes; a late error can leave a visible receipt. Retain and
review exit statuses separately. Capture drops and unavailable pcap stats remain
visible diagnostics, not a replacement denominator or an automatic packet-loss
verdict. Missing expected alerts remain in the report's failure and latency counts.

The default mode does not replay the adapter's packet identity join. No mode regenerates frames or
verifies the oracle against rules, authenticates jointly rewritten artifacts, establishes
physical arrival/clock accuracy, or proves storage persistence/hardware allocation.
Receipts and hashes show consistency of supplied bytes, not independent truth.
Opt-in reconstruction is described below; execution and physical qualification remain departmental.

## Bounds and persistence

The default total retained-input budget is 64 GiB, including duplicate source
copies. Increase `--max-bytes` explicitly for larger runs after allocating storage.
Metadata receipts/capture summaries are capped at 1 MiB each; manifest and
observations at 64 MiB each; report at 512 MiB to allow embedded JSON escaping and
summary expansion. These are implementation bounds, not demonstrated scale.
JSONL parsing streams rows capped at 256 KiB including newline. Metadata parsing
and report recomputation use memory proportional to bounded JSON sizes, with
multiple copies in memory. All raw input copies count against the byte budget.

Inputs must be regular files; direct file symlinks and special files are not read.
Source size/mtime/ctime/identity changes detected during copying produce a mismatch.
This is not an adversarial filesystem/authenticity guarantee; supply quiescent,
trusted paths. A byte-limit/read error retains the copied prefix, marked incomplete,
and continues inventorying other inputs. Malformed complete files remain retained.
Output directories use mode 0700 and snapshots 0600. Files and child directories
are synced before `bundle.json` is exclusively published last, then the output
parent is synced. Publication is not an atomic directory transaction: disk/fsync,
interruption or late parent-sync errors may leave partial files or a visible manifest.
Inspect nonzero process exits and use a new output directory for retries.

## Inventory-bound retained decoding (R90-137)

After a complete source snapshot, `_Bundle.decode` independently admits its
retained file through one read-only `O_NOFOLLOW | O_NONBLOCK` handle. A complete
captured inventory entry with matching key, nonnegative signed-64-bit byte count
and lowercase SHA-256 is required. Missing/unavailable admission flags, nonregular
files or unavailable integer device/inode/size/mtime/ctime metadata are errors.
Known retained size must equal captured bytes before any read. No source or
retained file is repaired, overwritten or deleted by decoding.

JSON reads are bounded by captured bytes plus one rejection byte. JSONL reads
are bounded by both remaining captured bytes and the existing 256 KiB row limit,
plus a rejection byte. EOF requires exact consumed byte count/SHA-256 and unchanged
descriptor metadata. The handle closes before a row count or decoded document
is committed. Metadata JSON retains the original universal-newline parsing
semantics (CRLF/CR become LF) while the inventory digest covers exact raw bytes.
JSONL keeps its strict JSON/object/submission validation and row-counting order.

Rejected admission, mismatched inventory, parse or close failure adds no new
trusted decoded state or completed decode check. Validation failures retain the
existing mismatch classification; I/O errors retain the existing non-success
and partial-output behavior. Inventory `complete` remains the source-copy flag;
it does not alone prove successful decoding. Earlier successful state/history
is not repaired or erased, and no whole-operation rollback is promised. Bundle,
pair metadata and both reconstruction consumers share this decoder without
public schema/inventory/status changes. Bundle source digests naturally change;
the existing tool identity and comparability review still applies.

This binds bytes used by this decode to the captured inventory at its read
boundary. It does not authenticate the original source, continuously freeze
writers, secure parent traversal or protect unrelated later reopen operations.
Metadata parsing still needs memory proportional to the existing bounded JSON
sizes and temporary copies; no measured scale or whole-bundle perpetual integrity
is claimed. Implementation has static source/AST review only; direct departmental
regressions remain unexecuted.

## Inventory-bound report-source read (R90-141)

`_summary` separately admits retained `adapter/observations.json` after the
original report-source fields/hash/string checks. Its complete captured inventory
must match that key, with nonnegative signed-64-bit bytes (bool rejected), the
existing 64 MiB observations ceiling and lowercase SHA-256. Required nonzero
integer `O_NOFOLLOW`/`O_NONBLOCK` flags must be available.

One read-only non-following/nonblocking regular descriptor supplies the raw bytes.
Integer device/inode/size/mtime_ns/ctime_ns metadata and known size equal to
captured bytes are required before read. The read is bounded to captured bytes+1;
the extra byte rejects. At EOF, metadata must remain unchanged and exact consumed
bytes/SHA-256 must match inventory. Wrapping failure closes the raw descriptor;
all wrapped admission/read/verification paths use context-managed close.

Close precedes the original embedded UTF-8/raw-byte/hash equality and complete
summary recomputation. Raw bytes are neither parsed nor newline-normalized here;
original source diagnostics and semantic comparison remain. Failure cannot newly
complete this check or qualify optional adapter/sender replay. Missing decoded
documents still skip; Value/shape failures remain mismatch and OSError retains the
existing non-success/partial-output boundary. Prior state/history and partial
files are preserved without a whole-operation rollback guarantee.

Bundle and fresh per-side pair reconciliation share this base check across
default/context/adapter/sender/combined modes; public schemas/status/exits/budgets
remain. The bundle source digest changes without waiving comparability. This
binds this read only, without authenticity, continuous writer exclusion,
parent-traversal security or protection of unrelated later reopens. The
[R90-141 plan](plans/task-20261002-report-binding-boundary.md) maps direct inventory/
admission/limit/EOF/cleanup/close/consumer cases; those regressions remain **not
run; delegated by user**. Nearby snapshot/decoder/replay cases cannot substitute.

## Required departmental validation — not executed

| Boundary | Required cases |
| --- | --- |
| Complete chain | Real retained sender/capture/engine/adapter/report, exact identities/origins/fixture/port; both profiles; failed/inconclusive report remains failed/inconclusive |
| Missing/partial | Every omitted/missing/unreadable artifact, absent receipts, false completion, mixed gaps and mismatches, invalid process exits after visible receipt |
| Encoding/schema | Duplicate members, nonfinite numbers, wrong types/unknown fields, invalid or repeated inventory paths, blank/oversized/truncated JSONL, UTF-8 errors, boolean/numeric substitutions |
| Tampering | Each raw hash/byte/row count, divergent offered/events copies, fixture/provenance mismatch, run/origin/port drift, embedded report bytes and every recomputed summary field |
| Filesystem | Source mutation, symlink/FIFO/directory, exact and exceeded per-file/total budgets, empty files, interrupted reads, disk-full/write/fsync/close/parent-sync faults, no-overwrite and mode preservation |
| Retained decoder admission | Direct `_Bundle.decode` normal/space, missing/directory/FIFO/symlink, unavailable flags/metadata and negative size, missing/incomplete/wrong-key inventory, invalid byte type/bool/negative/overflow or digest shape |
| Retained decoder binding | Empty/exact/over captured bytes, short read, same-size different hash, growth/truncation/mutation/immediate replacement, JSONL row boundaries, exact raw bytes/hash/rows, LF/CRLF/CR/BOM/Unicode, malformed/deep/duplicate/nonfinite/finite-float-overflow and submission diagnostics |
| Retained decoder failure state | Open/fdopen/fstat/read/close faults; no new document/row count/completed check on rejection; independent source/output-byte preservation; strict schema/status/source-digest/partial-output compatibility across all four consumers; nearby source snapshot checks cannot substitute |
| Semantics/scale | Submission failures/timestamp order, equal-count but different identities, complete packet join replay, truthful oracle/clocks/durability, generator overhead, long-run disk/memory behavior |

No test or actual acceptance bundle was generated by the agent. Both profiles
still lack qualifying evidence under the [formal SLO contract](performance-slo.md).

## Comparing retained runs

The [declared-comparability tool](slo-compare.md) retains and reconciles two
bundles before comparing their available profile/workload/resource/tooling
declarations. Same declarations do not establish real hardware independence,
qualified workload or acceptance. Testing remains departmental.

## Offline observation reconstruction (R90-122)

The [raw-ledger reconstruction checker](slo-reconstruct.md) snapshots a supplied
adapter package, validates its original receipt and sources, reruns the adapter
on retained copies, and compares every observation field. Expected alerts compare
by event identity; missing values remain present. This separately diagnoses the
raw-ledger derivation gap without live traffic or a compliance claim. Bundle and
pair comparison now offer explicit fresh reconstruction mode under R90-123.
Implementation review is static only; departmental behavior/acceptance tests remain
outstanding.

## Fresh reconstruction mode (R90-123)

Add `--reconstruct-ledgers` to the bundle invocation above. This unexecuted option
requires a fresh replay of this bundle's retained adapter files. It does not accept
an old reconstruction receipt as proof. Default calls still emit schema v1; enabled
adapter-only calls emit schema v2 with `reconstruction_policy: retained_adapter_replay_v1`.

All existing source/report checks must complete without gaps, mismatches or skipped
checks before reconstruction begins. Otherwise the reconstruction summary records
incomplete evidence and a gap, while original mismatches remain. Eligible bundles
create `reconstruction/`, containing R90-122's retained `adapter/`, `rebuilt/` and
`reconstruction.json`. All five nested adapter inventory entries must match this
bundle's original snapshots, including exact bytes, hashes and ledger row counts.
The nested tool also binds run/metadata and rebuilt observations. An unchanged old
receipt cannot hide altered sources or replace current replay.

The v2 bundle adds `reconstruction`, `errors`, `max_reconstruction_bytes` and
`reconstruction_source_sha256`. The summary has these exact fields:

- `status`: review_required, mismatch, incomplete or error.
- `reconstruction_attempted`, `reconstruction_complete`: the nested adapter's
  attempt/completion flags; replay completion alone does not mean agreement.
- `source_binding_complete`: all five nested source entries match this bundle.
- `observations_match`: normalized observation equality, or null before completion.
- `manifest`: relative file/bytes/SHA-256 for the new nested receipt, or null when
  skipped or an operation failed before a reference could be retained.

A complete v2 bundle requires a complete, source-bound replay with observation
agreement. Nested mismatch becomes bundle mismatch; incomplete replay becomes
a bundle gap. Local replay I/O/SQLite errors produce `status: error`, exit 2 and
nonempty `errors`; snapshot/publication errors can still prevent a final receipt.
Precedence is error > mismatch > incomplete > review_required. Diagnostics and
partial files remain retained. With reconstruction gaps, `snapshots_complete` is
false even if all original files were copied; inspect individual inventory flags.
No measurement outcome, compliance flag or physical qualification is upgraded.

`--max-reconstruction-bytes` defaults to a separate 64 GiB input budget for the
nested replay. This is additional to `--max-bytes` and excludes rebuilt copies,
generated metadata and temporary SQLite storage. `--scratch-dir` chooses an
existing directory for that temporary index. Both options are used only when
reconstruction is enabled; positive byte-limit arguments are validated regardless.
Allocate space for all original and nested copies; no total workspace quota or
performance claim is implied. Nested child entries are synced before the bundle
receipt is published. Existing partial-output/no-overwrite rules remain.

### Additional departmental validation — not executed

Cover v1 default compatibility; opt-in v2 success, base-check skip, nested gaps,
changed snapshot/receipt/rows/run identity, observation mismatch, byte-only
formatting changes, raw identity/time rejection, replay budgets, scratch/SQLite/
I/O failure and late publication errors. Verify source binding independently of
nested completion, mixed diagnostic precedence, old receipt substitution, retained
partial outputs and unchanged failed/inconclusive measurement outcomes.

## Fresh sender reconstruction (R90-127)

Add `--reconstruct-sender` to require [sender-source replay](slo-sender-reconstruct.md).
It can be selected alone or with `--reconstruct-ledgers`. These are unexecuted
handoff options. Both selected replays require the original bundle checks to
complete; after that gate, a recorded failure in one does not suppress the other.

The sender operation consumes only this bundle's four retained `sender/` files.
It creates `sender-reconstruction/sender/` copies and
`sender-reconstruction/sender-reconstruction.json`. All four fresh inventory
entries must equal the enclosing bundle's entries, including file name,
bytes/SHA-256/completeness and JSONL row counts. No supplied standalone receipt
or prior nested replay is accepted as proof of current execution.

Sender selection emits bundle schema v3 with these additions to the base fields:

| Field | Contract |
| --- | --- |
| `sender_reconstruction_policy` | `retained_sender_replay_v1` |
| `sender_reconstruction` | Status, attempted/complete/binding flags, `sender_records_match` and manifest reference |
| `adapter_reconstruction_required` | Boolean; true includes all existing adapter-replay fields and requires that replay too |
| `max_sender_reconstruction_bytes` | Positive retained-input budget for nested sender replay |
| `sender_reconstruction_source_sha256` | Current sender reconstruction source digest |
| `errors` | Combined recorded operation errors from all selected replays |

The sender summary uses `status`, `reconstruction_attempted`,
`reconstruction_complete`, `source_binding_complete`, `sender_records_match`
(true after completed correlation, otherwise null), and `manifest` (relative
file/bytes/SHA-256, or null). A complete bundle requires every selected replay to
complete, bind its inputs and agree. Partial binding is never sufficient.
An operation error takes precedence over mismatch and incomplete evidence; all
diagnostic lists remain available. One replay can succeed while the other fails.
Original component failure or missing input is never upgraded by replay.

`--max-sender-reconstruction-bytes` defaults to 64 GiB, separate from bundle and
adapter budgets. Pair review applies it separately to each side. It covers four
additional retained input copies, excluding generated metadata; no SQLite scratch
is needed for sender replay. All budget arguments must be positive even when a
mode is disabled. This is not a total workspace quota. Preserve partial output
and inspect process exit after late write/fsync/close or interruption failures.

Without sender selection, v1 default and v2 adapter-only structures remain.
Agreement establishes supplied-byte derivation, never actual sending, rule-oracle
truth, physical arrival/durability, clock accuracy, comparability or SLO compliance.

### Additional departmental validation — not executed

Cover each sender inventory field/missing/extra/partial file; old receipt
substitution; each nested mismatch, gap and I/O error; sender-only, adapter-only,
both and neither; simultaneous differing replay outcomes, budgets/no-overwrite,
late publication and interrupted retention. The
[R90-127 plan](plans/task-20260930-slo-sender-integration.md) maps static source
review to the full unexecuted handoff. No behavioral or knowledge suite ran.
