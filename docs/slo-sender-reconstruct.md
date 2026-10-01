# Offline sender-source reconstruction

`scripts/slo_sender_reconstruct.py` checks whether supplied reference-sender
ledgers derive from their retained fixture and run/link configuration. It uses
only pure frame construction from [the sender](slo-ingress.md); it does not send
packets, wait for a schedule, discover resources or start services. Behavioral
validation remains **not run; delegated by user**. Agreement is not SLO acceptance.

## Interface and retained artifacts

Unexecuted departmental command template (the parent must exist):

```bash
python3 scripts/slo_sender_reconstruct.py --sender sender-run \
  --output-dir sender-review --max-bytes 68719476736
```

Python API: `reconstruct(output: Path, *, sender: Path | None = None,
max_bytes: int = DEFAULT_MAX_BYTES) -> dict`. Omitting `sender` records incomplete
evidence. The new output directory uses mode 0700 and retained files use 0600.
Existing output is rejected; source bytes are neither repaired nor deleted.
Only these fixed sources are opened and copied under `sender/`:

- `submission.json`: original schema-v1 completion receipt, run/origin/link,
  byte boundary, inventory and original sender source hash.
- `fixture.jsonl`: scheduled offsets, Base64 payloads and expected rule IDs.
- `offered.jsonl`: packet identity, offer time, frame length and expected events.
- `submissions.jsonl`: successful submission claims, times and frame hash.

Retention reuses the [bundle policy](slo-bundle.md): non-following regular-file
opens, per-file metadata comparison, exact hashes/counts, a cumulative input
budget (default 64 GiB), 1 MiB receipt limit and 256 KiB JSONL row limit. Budget
exhaustion retains a prefix and prevents replay. Parent directory traversal is
not an authenticated filesystem boundary. Metadata/source hashes do not prove
acquisition authenticity. Generated metadata is outside the input budget.

## Correlation contract

The original receipt must satisfy the sender schema and exact three-ledger
inventory; its origin is canonical whole-second UTC in `[2000,2100)`. Nonempty
ledger counts must agree. Strict JSON decoding rejects duplicate members and
nonfinite values. Missing or failed prerequisites prevent correlation.

Replay keeps one bounded row per ledger, consuming the retained files in order:

1. Fixture fields are exact; offsets are nonnegative integers, nondecreasing and
   at most seven days. Payload uses strict Base64 decoding. Rule IDs are bounded
   ASCII identifiers, unique within each row; an empty rule list is permitted.
2. Row N must use `pkt-N` in both offer and submission. Expected events must equal
   the fixture's ordered rules with `slo_` plus SHA-256 of UTF-8 packet ID, NUL,
   and UTF-8 rule ID. Reordering expected events is a mismatch because the sender
   preserves fixture order. No rule matching is performed to certify the oracle.
3. `slo_ingress.frame` rebuilds the marked Ethernet/IPv4/UDP frame, including MTU
   constraints, checksums and padding. Offered bytes and submission frame hash
   must equal the rebuilt length and SHA-256. Frames are not saved or submitted.
4. Scheduled offset equals the fixture offset. `offered_ns` equals scheduled
   offset plus lateness and satisfies `offset <= offered <= send_start <=
   send_return`; timing values are nonnegative bounded integers. These checks
   compare claims and cannot establish actual scheduling or clock accuracy.
5. Unequal lengths, missing/extra/reordered/duplicate packet identities and
   inconsistent fields fail. Fresh replay byte/hash/row inventories must match
   the original snapshots before reconstruction can complete.

The first correlation failure stops replay and retains the matched prefix count,
one-based current row (when available) and fixed diagnostic boundary without
copying untrusted values into messages. Snapshot/schema failures are also listed
by source/check; they need not reach a correlation row. Memory is bounded by row
and metadata sizes, not packet count; runtime and disk costs are unmeasured.

## Result and failure handling

`sender-reconstruction.json` has `schema_version: 1` and
`artifact_kind: netsentry_slo_sender_reconstruction`. It includes:

- Status, attempted/complete flags, `sender_records_match` (true only after full
  correlation, otherwise null), progress and completed/skipped check names.
- Retained inventory with bytes/hash/completeness and decoded counts where
  available, plus gaps, mismatches and I/O errors.
- Validated run/origin/link and original sender source hash when available;
  current reconstruction, ingress, bundle, collector and reporter source hashes.
- Permanent `execution_complete`, `facts_verified`, `slo_compliance_asserted`
  and `comparability_established` false; `departmental_review_required` true.

| Status | Exit | Meaning |
| --- | --- | --- |
| `review_required` | 0 | All byte-derivation checks completed; departmental review still required |
| `mismatch` | 1 | Malformed or inconsistent supplied evidence |
| `error` | 2 | Captured I/O failure; inspect partial artifacts |
| `incomplete` | 3 | Missing/partial prerequisites prevented completion |

Errors take precedence over mismatches, which take precedence over incomplete
inputs. Setup, source-hash reads, final publication/fsync/close or interruption
may terminate with exit 2 and no complete manifest. Process termination can leave
other exit codes. Inspect both process outcome and artifacts: even a published
manifest is insufficient if the operation exits unsuccessfully. Preserve the
entire directory and retry into a fresh path. A matching prefix is not success.

## Departmental handoff and scope

Unexecuted validation cases are enumerated in the
[R90-126 plan](plans/task-20260930-slo-sender-reconstruct.md): each identity,
sequence, oracle, timing and inventory mismatch; malformed payloads and size
boundaries; missing/nonregular/mutating input, budgets, no-overwrite, I/O/fsync/
close/interruption and scale. No behavioral, acceptance or knowledge suite ran.

This is a standalone review tool. [Bundle](slo-bundle.md) and
[pair](slo-compare.md) review do not consume this receipt or run sender replay;
their existing reconstruction option rebuilds adapter observations only.
Fresh consumer integration is a separate queued increment. Actual kernel/NIC
submission, physical timing/durability, authentic builds and oracle correctness
remain departmental evidence obligations in the [runbook](slo-runbook.md).
