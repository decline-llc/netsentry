# R90-133: Snapshot bounded finalized fixtures before sender submission

## Selection and authority

The user explicitly starts R90-133 development. Clean fetched
`HEAD == origin/main == FETCH_HEAD` is
`e8d057963d3c655b4494ed1f3d74dc10a8303e6d`. R90-132 audit and closure exact
ranges have their verified iteration notes, index rows and MOC links; the
345-file Vault snapshot JSON hash is
`d7f644b9d643d9dba56f0479033b616a48f935d330f88bdfdbacb4c6b1945c48`.
The 45-commit Sep 3–Oct 1 phase audit separates SLO implementation/replay,
collector admission, queue repair, historical candidate publication and main
metadata delivery. All R90-133 dependencies are delivered; R90-75 remains
departmental acceptance. No new execution evidence has appeared.

The standing test-department split continues: behavioral/CLI/traffic/acceptance,
scanner and knowledge suites are not run, delegated by user. Static AST and
documentation checks and exact Git/Vault delivery remain agent responsibilities.

## Scope and behavior

Six intended repository paths: `scripts/slo_ingress.py`, `docs/slo-ingress.md`,
`docs/slo-runbook.md`, the rolling roadmap, this plan and the matching task state.
Keep `slo_collect._rows`, collector/bundle/replay implementations, schema-v1
receipt/inventory fields, runtime capture/engine, dependencies and workflows.

Add keyword-only `max_fixture_bytes` to `send_fixture`, default 64 GiB, and
`--max-fixture-bytes` to the CLI. Require an integer in `[1, 2^63-1]`, rejecting booleans
in the Python API, before output creation. Validate the CLI budget before socket
setup. The budget bounds only source fixture bytes, not generated ledgers or
workspace storage. Existing positional calls and successful receipt fields stay
compatible; existing output is never overwritten.

Open the supplied final pathname once with `O_NOFOLLOW | O_NONBLOCK`, require a
regular file and stable device/inode/size/mtime/ctime metadata. Reject a known
oversized file before retention. Copy raw JSONL rows to new mode-0600
`fixture.jsonl` within the total budget and 256 KiB row limit, checking short
writes. Compare metadata and consumed size, flush/fsync and close both source
and retained writer before creating offer/submission ledgers or invoking send.
Errors retain the available snapshot prefix and prevent completion and send.

Read retained bytes through one non-following regular-file handle, verify its
captured metadata before reading, parse each row with the existing strict JSON
diagnostics, and preserve the existing semantic/send loop. Bound this read by
the captured byte count; after EOF require exact bytes/rows/hash and unchanged
descriptor metadata before completion. The source inventory remains the existing
four fields (`file`, `bytes`, `rows`, `sha256`). Scheduled offsets/oracle order,
offer-before-send, callback full-frame semantics and failure receipts stay intact.

Filesystem metadata is an observation, not authenticity or continuous mutation
exclusion. Parent traversal and mutation of locally retained output are outside
the authenticated boundary. A later invalid semantic row or local replay change
may follow already submitted packets; preserve the entire incomplete run. Full
snapshot preparation adds latency before scheduling; it does not change origin
or offsets and cannot establish offered-load adequacy.

## Acceptance and evidence map

| Acceptance | Evidence planned |
| --- | --- |
| Bounded API/CLI admission | Static signature/default/positive-integer checks, CLI propagation and validation-before-socket order; source/row limits checked before writes |
| No sends on rejected acquisition | Data-flow review: same-handle regular-file and metadata/size checks, snapshot flush/fsync/close complete before ledger creation and semantic/send loop |
| Stable exact retained inventory | Bounded memory one raw row at a time, short-write checks, exact snapshot count/hash and retained replay comparison, descriptor metadata at both boundaries |
| Format and submission compatibility | Strict bundle/receipt and replay schemas compared with unchanged receipt fields and semantic loop; existing callers remain valid; no protocol, frame, oracle or schedule changes |
| Honest review and delivery | AST/docs/JSON/unique full roadmap multisets/history/links/fences/diff/sensitive-scope checks; focused commit, push/fetch and exact Vault reconciliation/replay |

Departmental cases remain unexecuted: ordinary/space paths, missing/directory/
FIFO/symlink, zero/bool/negative/noninteger/exact/over budgets and signed-64-bit
ceiling, 256 KiB boundary,
source mutation/growth/truncation and immediate pathname replacement, short
read/write/fsync/close/interruption, zero-send acquisition failures, 0600/0700,
exact bytes/rows/hash and strict downstream receipt replay, output replacement
and mutation, no-overwrite/preservation, empty and late malformed/semantic rows,
callback exception/short send, preparation latency/lateness and large inputs.
No executed regression or scale result is claimed.

## Risks, non-goals and stop conditions

Risk is medium: acquisition precedes a traffic-producing callback, snapshot
preparation affects lateness and changes partial fixture retention from a consumed
prefix to a complete bounded input. Strict schemas forbid silently adding receipt
fields. Stop on required format migration, inconsistent identities, necessary
dependency/global-reader changes or new private/external/product authority.
No actual traffic, test/acceptance execution, resource discovery, authentication,
global quotas, new protocols, runtime/toolchain/workflow/publication change or
next increment is in scope. Preserve partial evidence rather than deleting it.

## Implementation and static review checkpoint

The implementation matches the planned sender-local boundary. Snapshot metadata
is checked after the retained writer finishes fsync/close; source close also
precedes return into ledger creation and send. Retained replay checks its captured
metadata before reads and after EOF, with exact inventory equality. The existing
shared signed-64-bit integer validation constrains both new budget entrypoints.
Receipt fields are unchanged; one limitations string records the preparation
latency/storage boundary. The source reader and downstream implementations remain
untouched. No external input, traffic or test callback was executed.

AST comparison against the selected baseline confirms unchanged frame/checksum/
link/ledger functions, original positional arguments, complete semantic/timing/
oracle/send loop and receipt fields/values except descriptive limitations. The
strict bundle receipt schema still matches all produced fields. Static source
review covers admission flags, regular-file/metadata checks, budget/row checks
before writes, short-write handling, writer/source close before send, retained
inventory comparison and CLI propagation/validation-before-socket. Python/docs,
151 task JSON parses, 137 unique roadmap row/Definition pairs and ordered history
review pass. These are static evidence only; the departmental cases remain unrun.
An initial AST review command used a list where `ast.dump` requires a node; the
corrected complete static review passes and no runtime behavior was executed.
Existing skills cover finalized-input and exact-boundary review; no skill change
is warranted. Final diff/scope/sensitive review and Git/Vault delivery remain.

## Delivery and acceptance closeout

Feature `8a92cd99e16c599a1e9d99e614bba4cb5838e90e` contains exactly the six
planned paths. Push and fresh fetch verified matching clean HEAD/origin/main/
FETCH_HEAD; publication tag identities remained unchanged. Exact range
`e8d057963d3c655b4494ed1f3d74dc10a8303e6d..8a92cd99e16c599a1e9d99e614bba4cb5838e90e`
was synchronized and its iteration note, index row and MOC link verified. All
12 stable current notes describe the implementation and retained limitations;
all 298 pre-existing iteration-directory notes remain intact. Replay preserves
the 346-file snapshot JSON hash
`4d8b0b7305afe7e0bbb903ba2c76191fba3e0afeb4e28c4ee65b34ea26e46537`.

All implementation acceptance criteria map to the promised static source and
AST evidence. The unchanged complete semantic/send loop and strict field schemas
supply direct compatibility review; no nearby test is substituted for an executed
regression. The full departmental matrix is still explicitly unrun. Planned
snapshot preparation delay and complete fixture retention are documented; there
is no repository scope deviation. Final Python/docs/static compatibility review,
151 JSON parses, 137 unique matching row/Definition pairs, chronology, local
links/fences, exact scope, diff and sensitive additions review pass.

The one docs-only closure records the feature as complete implementation, with
execution evidence delegated. Resolve the closure SHA from Git, verify its own
push/fetch and exact Vault range before final reporting. The next trigger should
verify that closure and audit fresh evidence; no additional local ready increment
is queued and R90-75 remains departmental. No subsequent work starts here.
