# R90-126: Standalone sender-source reconstruction

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD baseline is
`99796842209d6909d66e0b6295284ca867963967`. R90-125 feature/closure
notes, full index and MOC links exist in the sole local Vault; stable prose
correctly identifies this ready increment. September 2–30 history contains 28
commits across queue recovery, SLO contract/report, adapter/runtime/ingress,
bundle/context/reconstruction and runbook/queue delivery. No missing closure or
new acceptance evidence changes priority. R90-59 authority and R90-75 departmental
acceptance remain separate. All 142 prior task states parse; 130 unique roadmap
rows/Definitions match. The Sep 30–Dec 28 horizon remains current.

The Sep 25 user instruction delegates tests, knowledge suites and acceptance.
Use static AST/source/schema/docs/JSON/diff and sensitive-information review plus
Git/remote/Vault verification. Do not run tests, imports/CLI smoke, traffic or
measurement acquisition. Source review establishes implementation coverage only.

## Frozen interface and output contract

`scripts/slo_sender_reconstruct.py` exposes
`reconstruct(output: Path, *, sender: Path | None = None,
max_bytes: int = DEFAULT_MAX_BYTES) -> dict` and CLI `--sender`, required
`--output-dir`, optional `--max-bytes`. Default budget is 64 GiB, minimum one byte.
Create a new private output directory in an existing parent; never overwrite or
repair supplied files. Snapshot fixed sender/submission.json, fixture.jsonl,
offered.jsonl and submissions.jsonl with existing bundle retention policy.
Omitted/missing/nonregular inputs are gaps; budget-exceeded snapshots retain
prefixes; changed/malformed/inconsistent sources are mismatches. Metadata has a
1 MiB cap and JSONL rows a 256 KiB cap. Fresh replay reads only those snapshots,
with one aligned row per ledger in memory and no packet-count-sized collection.

Publish schema_version=1, artifact_kind=netsentry_slo_sender_reconstruction to
sender-reconstruction.json. Status/exit: review_required/0 only after all checks,
mismatch/1 for inconsistent evidence, error/2 for I/O/resource failures,
incomplete/3 for absent or partial prerequisites. I/O/publication/interruption
can leave partial output without a final manifest; process exit and artifacts
must be reviewed together. Always retain successful prefix counts and the first
failure boundary/row, input byte/hash/count inventory, original sender digest,
current reconstruction/ingress/bundle/collector/reporter digests, run/link/origin
when validated, attempted/complete and nullable sender_records_match.
Execution, facts, compliance and comparability flags stay false; departmental
review stays true. No rebuilt packets or per-packet output are persisted.

## Acceptance to source and departmental evidence map

| Acceptance | Implementation/static review | Departmental validation (not run) |
| --- | --- | --- |
| Receipt and inventory | bundle snapshot/_sender plus origin year [2000,2100), fixed inventory and readiness gate | metadata/schema/flags/link/origin/digest/count/empty drift, incomplete receipts |
| Fixture contract | bounded strict JSON, exact fields, integer offsets monotonic within seven days, strict Base64 and unique bounded rules | malformed/oversized/duplicate members, payload types, offsets zero/equal/limit/out of range |
| Packet/oracle/frame | aligned pkt-N, ordered expected rule/event objects, ingress.frame only, exact length/hash | missing/extra/reordered/duplicate rows, oracle identity/order/type, marker/MTU/padding/checksum changes |
| Timing relationships | scheduled=offset, offered=scheduled+lateness, offset <= offered <= start <= return; integer contracts | each timing/type mismatch and bounds |
| Retained-source binding | replay hash/byte/row count of each complete ledger equals snapshot inventory; stop at first invalid row | post-snapshot mutation, changed digest/count and partial diagnostics |
| Offline bounded failures | no live sender, clocks/waits, network, discovery or services; O(one bounded row) memory | no side effects, large ledgers, missing/nonregular/mutating input, no-overwrite, budget, read/write/fsync/close/interruption |
| Delivery | syntax/docs/JSON/unique roadmap multiset/diff/sensitive review; verified push and exact-range Vault stable prose/replay | behavioral and knowledge suites remain delegated |

## Scope and boundaries

Seven intended paths: new script and its guide, Makefile syntax list,
docs/slo-runbook.md discovery link, roadmap, this plan and task state.
Reuse pure ingress.frame without changing sender behavior. No bundle/pair wiring,
new protocol, rule-engine oracle, physical authenticity, thresholds, dependencies,
CI changes, private input, release action or additional increment. Stop for an
undocumented format or new private/external/product authority. Runtime/cost remains
unverified. Record deviations and preserve partial evidence; refresh future queue
without starting it. Commit/push/fetch/sync feature, then one docs-only closure.

## Implementation/static review checkpoint

The seven-path implementation retains fixed sources through bundle snapshot
policy, validates the original sender receipt plus its origin-year boundary,
correlates bounded aligned rows using only ingress.frame, and verifies fresh
ledger inventories before success. Diagnostics distinguish snapshot prerequisites,
correlation row/boundary and I/O errors; partial output is preserved. The guide
specifies status/exit precedence and the possibility of a published manifest with
an unsuccessful final durability operation. Existing sender/bundle/pair behavior
is unchanged. R90-127 is scoped as a separate planned consumer integration with
complete dependency/window/risk/review/stop conditions, without implementation.

Static review maps every acceptance row above to source and explicit unexecuted
departmental cases. No direct regression is claimed or run under the user testing
split. AST/python-check, docs-check, task JSONs, unique complete roadmap multisets,
ordered history, local links/fences, exact seven-path and sensitive-information
review pass: 143 task JSONs, 131 unique roadmap pairs and seven intended paths. No scope deviation or reusable skill change
was identified. Behavioral correctness and resource costs remain unverified.

## Verified implementation delivery and next boundary

Feature `8378a4fc89b07fbcf0db851c00d1c7430cdde15b` contains exactly the seven intended paths.
Push without force/tags followed by a fresh fetch verified clean
HEAD/origin/main/FETCH_HEAD equality. Exact range
`99796842209d6909d66e0b6295284ca867963967..8378a4fc89b07fbcf0db851c00d1c7430cdde15b`
was synchronized to the sole existing local Vault. Note/index/MOC and nine
current stable notes are reconciled; pre-existing immutable iteration bytes
remain unchanged. Identical-range replay preserved Vault Markdown hash
`ae46ca1b9f34b9f76ef3addb62dc2f6a22586b684e79ce90235fb3e92a65c608`.

Every acceptance criterion maps to implemented source and the published
unexecuted departmental validation matrix. Static checks pass; no behavioral
regression, runtime/scale proof or SLO acceptance is claimed. No scope deviation
or skill update was necessary. R90-126 is complete as an implementation delivery
under the standing user testing split. R90-127 is next ready and unstarted;
its explicit consumer API/schema plan must be persisted on the next trigger.
R90-59 candidate/tag authority and R90-75 departmental acceptance remain separate.

This single docs-only closure receives its own verified push/fetch/exact Vault
range. Future sessions should verify that closure against fetched origin/main,
without repeating completed R90-126 feature commit/push/sync or running tests.
