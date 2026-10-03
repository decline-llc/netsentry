# R90-147: Preserve concurrent hello and heartbeat state updates

## Selection, baseline and authority

Fresh fetched clean main/HEAD/origin/main/FETCH_HEAD is
`410104eb05bfaf03d4234802fc836ee64ca6ae09`. R90-146 feature/closure exact
Git/Vault ranges, seven/three-path generated scope, notes, full index and MOC are
verified. The 373-file snapshot JSON SHA-256 is
`f83f7e6ad487d703ca1a78b85e032ac0e56a9e027d0d006808008a5cbd43880e`;
14 current stable notes have full-content backups and 326 immutable iteration
notes are captured. The 73-commit Sep 4–Oct 2 phase audit covers measurement/
reconstruction, bounded evidence inputs, queue reconciliation, historical
candidate/toolchain delivery and core rule/pipeline/suppression repairs. No
missing delivery or new qualifying R90-75 acceptance evidence appears.

The defined local queue is empty. Source shows heartbeatState.SetHello and
SetHeartbeat independently Snapshot then Store, so an interleaved writer can
overwrite the other's unrelated fields with a stale State. Concurrent receiver
connection handlers invoke these setters; atomic.Value makes each published
State safe to read but does not serialize compound updates. Register R90-147 as
bounded core correctness repair after R90-146 and the existing receiver control
contract, within this increment rather than a standalone queue audit. Persist
plan/state before runtime or roadmap edits. Horizon remains Oct 2–Dec 30;
R90-75 retains its full independent asynchronous departmental contract.

The user-directed test split applies: behavioral/race/CLI/full-suite/scanner/
knowledge/acceptance execution is **not run; delegated by user**. Author direct
regression source and perform static/compile-only review, without executing test
binaries. Engine module preflight resolves pinned Go 1.26.8.

## Exact scope and behavior

Exactly six paths: `engine/internal/receiver/heartbeat.go`, new
`engine/internal/receiver/heartbeat_test.go`, `docs/architecture.md`, rolling
roadmap, this plan and matching task state. Isolated local branch:
`fix/r90-147-control-state-serialization`. Add one private writer mutex and take
it before Snapshot in both setters, holding it through the same field updates
and single Store. Snapshot remains an unchanged atomic.Value Load, without a
reader lock. Constructor initialization, all State/frame fields, timestamp UTC
generation and receiver/API source remain unchanged.

SetHello updates SessionID/Hello and preserves Heartbeat/LastHeartbeatAt;
SetHeartbeat updates SessionID/Heartbeat/LastHeartbeatAt and preserves Hello.
Writer acquisition order defines publication order; no fairness or wire-arrival
ordering is promised. Global State continues to contain the latest stored hello
and latest stored heartbeat, potentially from different capture sessions; top-
level SessionID belongs to the last setter. Connection-local hello/heartbeat
validation is unchanged. This repair does not create per-session state, select
an active capture, reset heartbeat on hello or enforce heartbeat sequence order.

## Acceptance and direct evidence

| Acceptance | Local static/compile evidence | Direct departmental regressions, unrun |
| --- | --- | --- |
| Baseline and eligibility | Fetched refs, prior exact ranges/Vault, phase/queue audit, persisted contract | R90-75 independent |
| Serialize complete updates | Mutex Lock/defer before Snapshot in both setters; unchanged fields and single Store; atomic Snapshot unchanged | Direct setters with synchronized concurrent hello/heartbeat starts and joined final complete fields; intermediate snapshots contain whole frames and heartbeat/time pair |
| Preserve unrelated fields and metadata | Exact source transform; SessionID last setter and UTC update behavior unchanged | Both sequential orders, repeated hello/heartbeat replacement, initial zero state; hello preserves exact heartbeat timestamp |
| Receiver callers reach repair | Unchanged real handleLine/connectionSession/State path | Concurrent valid hello on one connection and heartbeat on another established connection, final Receiver.State retains both, frame/control counts and decode/packet counters unchanged |
| Independent value snapshots | State contains only value fields; no mutable slice/reference introduced | Modifying returned State/Hello/Heartbeat fields cannot alter published state |
| Scope and delivery | Go parse/format; pinned receiver/API/pipeline compile-only; docs/165 JSON/151 full unique roadmap/unchanged R90-75/ordered history/links/fences/six-path/diff/sensitive review; exact feature/closure Git/Vault/stable prose/immutable hashes/replay | Behavioral/race/broader suites delegated |

Concurrent regression source uses channel start gates/readiness and joins, without
fixed sleeps. Repeated synchronized starts exercise overlaps but do not guarantee
the old implementation loses a particular update on every scheduling run. Final
assertions require both complete frame values; static source review establishes
the whole transaction lock boundary. Tests reach setters and real receiver frame
handling, not an invented helper. Compilation is not runtime or race evidence.

## Risks, non-goals and stop conditions

Low-to-medium correctness risk: writer contention and accidental cross-session
claims. Lock only infrequent control-state writers; retain atomic readers and
all frame validation/fields. Exclude per-session model/API/metrics changes,
sequence/active-capture policy, timestamp injection, new production test seams,
packet queue/listener/startup/shutdown changes, matcher/storage/suppression work,
IPv6/dependencies/toolchain/suite execution/external publication. Stop for
required protocol or product decisions, non-value State fields, contradictory
evidence, competing edits or new external authority.

Complete exactly this increment; on review success commit six paths, fresh-
verify baseline, fast-forward main, push/fetch and synchronize exact range.
Reconcile stable prose preserving original topic tails and immutable history.
One three-path docs-only closure records verified facts and receives its own
exact Git/Vault verification. Refresh the queue without starting more work.
Existing skills cover compound-update ownership, observable synchronization and
topic-boundary preservation; no generic skill edit is warranted at selection.


## Implementation and final static checkpoint

Exact source comparison permits only sync import, one writer mutex and Lock/
defer Unlock before Snapshot in both existing setters. All frame/State fields,
constructor, single stores, UTC clock update, atomic Snapshot and real receiver/
API/Stats/pipeline/suppression/rule/exporter/config source are unchanged. State
contains only value fields; mutex lives in the private pointer-owned writer state.

Five direct functions cover both sequential orders/zero/replacement/timestamp,
256 synchronized concurrent setter pairs, 1024 hello and heartbeat writes each
with a reader that validates whole frames/time, 128 concurrent real handleLine
hello/heartbeat pairs on independent connection sessions with exact counters,
and modifying a returned State copy. The reader observes both first frames
while writers are paused before subsequent replacements, proving an actual
nonzero read before continuing; channel cleanup releases writers/readers.
All final field and independent-counter assertions occur after writers join.
No fixed sleeps, production seam, same-connection concurrent handler misuse,
FIFO requirement or deterministic old-race reproduction claim is introduced.

Pinned Go 1.26.8 complete receiver → API → pipeline compile-only chain passed
and passed again after final reader synchronization refinement. Binaries remain
outside the repository and unexecuted. Exact source/direct-boundary/value-graph
review, Go parse/format/docs, 165 task JSON, 151 full unique roadmap pairs, full
forward contract, unchanged R90-75, chronological history, links/fences, six-path
scope, diff and sensitive additions pass. No runtime/race/SLO outcome is proven;
all execution suites remain delegated and unrun. Full 373-file Vault baseline
is unchanged; full stable-content backups support topic preservation at sync.

A reusable lesson refined the local skill's existing concurrency instruction:
atomic Load/Store safety does not serialize a whole read-modify-publish transaction.
Require synchronization over that complete ownership boundary. Generic Markdown
frontmatter/numbering/fences validated; the edit stays outside feature delivery.
The sole planning deviation is source-grounded empty-queue repair; no scope or
implementation/static/compilation failure. Complete feature plus one docs-only
closure delivery; no subsequent increment starts here.


## Delivery and queue closeout

Feature `8b8b7703f447391e2296813dc7b95023afd143e6` contains exactly six planned paths. Isolated local
implementation branch `fix/r90-147-control-state-serialization` was fast-forwarded
to freshly verified main; push/fresh-fetch verified clean HEAD/origin/main/
FETCH_HEAD at that SHA. Exact range `410104eb05bfaf03d4234802fc836ee64ca6ae09..8b8b7703f447391e2296813dc7b95023afd143e6` has six-path
generated scope, iteration note `04-开发迭代记录/2026-10-02-8b8b7703f4-CI知识同步.md`,
full index and MOC verified. Fourteen current stable notes were reconciled with
every original substantive topic tail retained exactly. All 326 baseline
immutable iteration hashes are unchanged. Identical range replay preserves the
374-file Markdown snapshot; snapshot JSON SHA-256 is `6054b8ec198a04ff967f4146c4cfbdcd4c35aa1bdec58f0e14ac1445f7f504a3`.
The unique existing local sibling Vault was selected explicitly.

Acceptance matches the plan: one private mutex encloses both complete setter
read-modify-publish transactions, while atomic Snapshot, State/frame fields,
constructor, UTC clock update and single Store remain. Receiver/API/Stats and
other core source are unchanged. Hello preserves heartbeat/time, heartbeat
preserves hello. Latest-per-frame global aggregate may contain different sessions;
last setter determines SessionID. No frame/session policy, ordering, active-
capture, listener/queue/shutdown/API/metrics or production test seam was added.

Five direct regression functions cover sequential zero/order/replacements/time,
256 joined setter pairs, 1024 updates per writer with a reader that must observe
both first frames before writers continue, 128 real receiver hello/heartbeat
pairs on independent valid connections with final State/session/counter checks,
and independent returned value mutation. Final complete pinned Go 1.26.8
receiver/API/pipeline compile-only chain and static source/direct-boundary/value/
Go-format/docs/165 JSON/151 full unique roadmap/full forward contract/unchanged
R90-75/history/links/fences/six-path/diff/sensitive review pass. No binary or
behavioral/race/CLI/full-suite/scanner/knowledge/acceptance suite was executed;
all execution is user-delegated, without a runtime/race/SLO outcome claim.

The sole planning deviation registered this source-grounded core correctness
repair within the empty queue; there was no scope or implementation/static/
compile/delivery failure or Vault topic loss. A generic lesson refined the local
skill concurrency rule to review synchronization over the whole snapshot
read-modify-publish transaction; atomic load/store safety alone can still lose
updates. Markdown frontmatter/numbering/fences pass; skill edit is local-only
and separate from the repository commits.

One three-path docs-only record closes this increment. Resolve its full SHA from
Git, push/fresh-fetch and verify exact feature-tip..closure-tip Vault before
reporting; no extra closure merely for self-reference. Queue refresh: R90-147
implementation is complete; no further dependency-ready local item is defined.
R90-75 retains its full independent asynchronous departmental acceptance contract
and does not block development. Next trigger verifies fetched closure/Vault,
audits fresh core-code evidence and forward queue, then persists a separate
eligible plan before edits. No subsequent implementation starts here. Oct 2–Dec 30
horizon remains; IPv6 and external publication remain separate authority. Do not
repeat completed R90-146/R90-147 delivery or R90-59 publication.
