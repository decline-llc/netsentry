# R90-137: Bind retained bundle decoding to captured inventory

## Selection and authority

Fresh fetched clean HEAD/origin/main/FETCH_HEAD is
`269b391f07b6b76bffaacb9516379f864bc1e9f6`. R90-136 audit
`fb5e3702deebbf0bf58ba9c3db208de549b4cf1a` and closure exact ranges have
verified iteration/index/MOC. The 353-file Vault snapshot JSON hash reproduces
`fa9ab7f875f0761feac0d256be5cefb4eec3e47fc1f2512485edc021da3737c1`.
The 53-commit Sep 4–Oct 2 phase audit distinguishes contracts/SLO tooling,
reconstruction/replay/runbook, collector/sender/reporter admission, queue repairs,
historical candidate publication and main toolchain metadata. R90-75 and ready
R90-137 alone remain unfinished, with complete forward contracts. All internal
dependencies R90-136/R90-118/R90-119/R90-122/R90-126/R90-135 are complete; no new
qualifying departmental measurements appear. Select exactly R90-137 within the
current Oct 2–Dec 30 horizon, with forecast dates not acting as eligibility gates.

Behavioral/CLI/traffic/acceptance/scanner/knowledge execution remains not run,
delegated by the user's standing instruction. Static AST/source/docs and exact
Git/Vault delivery remain local responsibilities. Historical published candidate
validation does not establish current-main behavior or SLO acceptance.

## Scope and design before edits

Exactly six paths: `scripts/slo_bundle.py`, `docs/slo-bundle.md`,
`docs/slo-runbook.md`, rolling roadmap, this plan and matching task state. Persist
plan/state before source/documentation edits. Change only `_Bundle.decode` runtime
logic; no other reader, source snapshot, consumer, schema, dependency, toolchain,
workflow, CLI option, public metadata field or publication change.

Require a complete captured inventory entry matching key, validate nonnegative
signed-64-bit byte count (excluding bool) and existing lowercase SHA-256 shape,
and capture expected bytes/digest locally. Require available nonzero O_NOFOLLOW/
O_NONBLOCK, open retained source once read-only, close raw descriptor on fdopen
failure and use context-managed stream ownership thereafter. Require regular
file and integer dev/inode/size/mtime_ns/ctime_ns metadata; compare admitted size
to captured bytes before reading, rejecting negative/mismatched sizes.

For JSONL, bound each readline by min(256 KiB, remaining captured bytes)+1.
Keep the original row-limit error first for unchanged overlong rows; reject an
inventory overrun before parsing its probe. Accumulate raw bytes/digest locally,
then keep the complete original strict UTF-8/JSON/object/submission semantic
block and row counting order. At EOF compare exact bytes/digest to inventory and
required before/after descriptor metadata. Close before assigning the row count.
Empty JSONL retains its existing zero-row success if inventory matches.

For metadata JSON, replace following unbounded read_text with a same-handle read
of captured bytes+1, reject overrun and short read/hash/metadata mismatch, then
close. Decode exact raw UTF-8 only after acquisition succeeds. Preserve universal-
newline normalization (CRLF and CR become LF), matching prior read_text semantics
while hashing original bytes; retain the original json.loads hooks/finite-float
and object validation. Assign documents only after successful close/parse.
Empty/malformed/BOM/invalid UTF-8 retain original decode rejection for stable
inventory-bound input. Parsing errors cannot publish a new trusted result.

The strict public inventory fields/formats stay intact. Inventory complete is the
existing snapshot-copy flag; decode success is separately represented by completed
checks and decoded state. Failure does not add a document, row count or completed
decode check. Already retained bytes and prior state are not repaired or deleted;
this does not promise invalidation of earlier successful decode history or a
whole-operation rollback. `_Bundle.check` keeps EvidenceError/ValueError as
mismatch; OSError propagates to the existing error/exit/partial-output boundary.

Review ordinary bundle reconcile, pair original-manifest retention, adapter and
sender reconstruction (four _Bundle constructors) and their wrapper catches,
strict inventory/receipt/status/source-digest handling. Bundle source digest
naturally changes; no old/new tool identity equality or schema relaxation follows.
Metadata/hash binding observes this decode acquisition only: no authenticity,
continuous writer exclusion, parent-traversal security or unrelated later reopen
protection. Source snapshot, summary reopens and reconstruction row readers remain
unchanged; no perpetual whole-bundle integrity claim.

## Acceptance and evidence map

| Acceptance | Planned evidence |
| --- | --- |
| Inventory-bound admission | Static entry validation, flag/descriptor ownership, regular-file/required metadata and known-size check before read |
| Bounded exact decode bytes | JSON captured-size+1 and JSONL remaining/row-limit+1 flow, exact consumed size/digest and before/after metadata at EOF |
| Close-before-trusted-state | Every rows/documents write outside stream context; unchanged check marks success only after action returns; no writes on rejection |
| Format/parsing/consumer compatibility | AST equality for all other original definitions/constants/methods/signature, entire JSONL validation and JSON hooks; newline compatibility review, four unchanged consumers/strict schema/source-digest semantics |
| Reviewable delivery | Python AST/docs/155 task JSON/full 141 unique roadmap multisets/history/link/fence/diff/sensitive scope, exact Git/Vault ranges and stable-note reconciliation/replay |

Departmental direct cases remain unexecuted: ordinary/space retained paths;
missing/directory/FIFO/symlink; unavailable flags/metadata/negative size; missing/
incomplete/mismatched inventory entry, invalid byte type/bool/negative/overflow
and digest shape; empty/exact/over captured bytes, short read/same-size different
hash, growth/truncation/mutation/immediate replacement; JSONL row boundaries,
exact raw bytes/hash/rows; JSON LF/CRLF/CR, BOM, Unicode, malformed/deep/duplicate/
nonfinite/finite-float-overflow and submission semantic rejection; open/fdopen/
fstat/read/close failure, no new trusted state/completed check, preservation via
independent read-only handles, all four consumers' strict schema/status/source-
digest/partial-artifact compatibility. Each rejection must reach this decoder;
nearby snapshot tests cannot substitute. No execution or scale result is claimed.

## Risk, non-goals and stop conditions

Risk medium: a shared decoder adds intentional acquisition errors before metadata
JSON parsing and bundle tool identity changes. Parsing memory still scales with
bounded metadata (up to existing 512 MiB report cap) and temporary decoded copies;
no resource/scale measurement is implied. Stop on format migration, incompatible
shared error/state ownership, unsafe unsupported primitives, ambiguous review or
new external/private/product authority. No tests, CLI invocation, traffic,
acceptance, authenticity, new budget/schema/reader or next increment. Existing
skills cover these boundaries; no skill update is warranted. Deliver one feature
and one docs-only closure, refresh queue without starting subsequent work.

## Implementation and static checkpoint

Implementation matches the decoder-local plan. Inventory and flags are validated,
local digest/counters initialized before open; fdopen failure closes the raw
handle and the stream immediately enters context ownership. Both lanes enforce
captured-byte/probe bounds, JSONL row limits and EOF exact raw digest/metadata.
No shared state assignment occurs within the stream context. Metadata JSON
parse follows close, with explicit original universal-newline behavior; the whole
original JSONL validation/counting sequence, JSON hooks/object check, all other
original methods/helpers/constants and the decode signature are AST-equal.
Source snapshot/check/receipt/reconcile/publication logic is unchanged, including
strict output schemas/statuses and success marking only after decode returns.

Four shared constructors and unchanged comparator/reconstruction/sender/report/
collector code were reviewed; existing wrappers retain ValueError/mismatch and
OSError/partial-output boundaries, never asserting successful operation after
close failure. Python AST/docs, 155 JSON parses, 141 full unique matching roadmap
row/Definition multisets, ordered history/link/fence/six-path scope/diff/sensitive
additions pass. All 306 immutable prior iteration notes and the 353-file Vault
baseline are captured. All direct departmental regressions remain unrun; no
observed decoder correctness or preservation result is claimed. The metadata
newline normalization and early admission error order are planned compatibility
choices; no scope/review deviation or skill change is warranted. Exact feature
Git/Vault delivery and stable current knowledge reconciliation remain.

## Delivery and acceptance closeout

Feature `90ed9bba8a8f431a1906b0923110c0e3fa8de25f` contains exactly the
six planned paths. Push and fresh fetch verified matching clean refs. Exact
range `269b391f07b6b76bffaacb9516379f864bc1e9f6..90ed9bba8a8f431a1906b0923110c0e3fa8de25f`
has verified note/index/MOC. Twelve stable current notes describe inventory-bound
decoding and all retained authority/compatibility limits; all 306 pre-existing
immutable iteration-directory notes are unchanged. Identical-range replay
preserves the 354-file snapshot JSON hash
`1ac717acfccf6729d515ce92dc3cf50166ab7eb33d66f146d6f822043fe2ff5f`.

Each acceptance criterion maps to planned direct source/AST and delivery evidence:
inventory/flags/ownership/regular metadata/size admission, bounded reads and EOF
hash/metadata, context close-before-state, complete original JSONL semantics,
JSON hooks/newline behavior and unchanged other methods/constants/interfaces/
strict schemas/status/snapshot/publication plus four consumers. No source scope
deviation occurred. All direct departmental regression cases remain unexecuted;
static equality does not establish observed rejection/preservation/compatibility.
Final Python/docs/155 JSON/141 full roadmap pairs/chronology/link/fence/scope/diff/
sensitive review passes; no skill edit is warranted.

One docs-only closure records verified feature facts. Resolve its SHA from Git,
verify its own push/fetch/exact Vault range; do not add another closure merely to
embed its self-referential hash. R90-137 implementation is complete, tests delegated.
No additional local ready increment is queued; next trigger verifies closure
knowledge and audits fresh evidence. R90-75 and all execution suites remain
departmental. Do not repeat completed commits/push/sync/publication or start
another increment in this trigger.
