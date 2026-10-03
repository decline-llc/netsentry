# R90-145: Count successful terminal packet processing

## Selection and authority

Clean freshly fetched main/HEAD/origin/main/FETCH_HEAD is
`36d2cd49761308b9884cdbe47dd7ed9f5c2b2b42`. Both R90-144 feature/closure exact Git/Vault ranges,
generated path scope, iteration notes, full index and MOC are verified. The
369-file baseline snapshot JSON SHA-256 reproduces
`52c2acd1c86030707bec2dbb3fee3a86870dc50bdbe0645d5c8a3a7d28ede3da`.
Fourteen current stable notes identify ready/unstarted R90-145; 322 immutable
iteration-directory notes are captured. The 69-commit Sep 4–Oct 2 phase audit
separates measurement/tooling/replay, admission contracts, queue audits,
historical candidate publication/main toolchain and the new core ownership
repair. No missing delivery or qualifying R90-75 measurement appears.

R90-145 is the sole ready local item; R90-144 and R90-116 dependencies are complete.
R90-75 retains its full independent asynchronous departmental contract; it does
not block development. Oct 2–Dec 30 horizon remains current. The standing user
split applies: behavioral/race/CLI/full-suite/scanner/knowledge/acceptance execution
is **not run; delegated by user**. Author direct regression source and perform
static/compile-only checks; never invoke generated test binaries. Preflight
confirms engine module resolves its pinned Go 1.26.8 toolchain.

## Scope persisted before behavior and documentation edits

Exactly ten paths: `engine/internal/pipeline/worker.go`,
`engine/internal/stats/stats.go`, new direct regression files
`engine/internal/pipeline/completion_test.go`,
`engine/internal/stats/completion_test.go`,
`engine/internal/api/completion_metrics_test.go`, `docs/architecture.md`,
`docs/api-reference.md`, rolling roadmap, this plan and matching task state.
Implement on `feat/r90-145-packet-completion-counter`; after completed review,
commit then fast-forward main from freshly verified baseline, push/fetch main
and synchronize exact range. One docs-only closure records delivered facts and
gets its own exact Git/Vault verification. No remote feature branch, force push,
tag, release, workflow dispatch, dependency or toolchain changes.

## Behavior and compatibility

Add atomic packetsCompleted, nil-safe IncPacketCompleted, Snapshot.PacketsCompleted
and `netsentry_packets_completed_total` with counter HELP/TYPE in existing
RenderPrometheus. Change only Worker.processed(): return on failed optional
Processed export, then increment once. Existing terminal calls cover no alerts,
full suppression and successfully persisted/exported alerts. Existing writer,
Arrival/Durable export errors and panic recovery exit before terminal success;
terminal Processed error or panic also prevents completion. Prior counters,
alert accounting, logging/error ownership, observer calls and SLO exporter
implementation/schemas remain unchanged.

processPacket keeps its original IncPacketProcessed before matching. Receiver
counting and old packet/alert rate gauges keep their current meaning. New count
measures successful Worker terminal actions in this process; writer success alone
is not a universal durable-storage proof. No-alert and fully suppressed packets
complete without storage writes. Enabled observer success includes exporter
acceptance of events, not final fsync/receipt publication. Neither legacy nor new
aggregate count is a per-packet offered-load/loss oracle or SLO acceptance gate.
Counts reset at process restart; Snapshot samples atomics independently while
workers are active, without a transactional cross-counter consistency promise.

API runtime source stays unchanged: /api/metrics uses RenderPrometheus and exposes
the additive counter automatically. /api/health verbose JSON keeps all existing
throughput keys and values; no new JSON field or completion rate is introduced.
Run still skips nil packets and preserves its select/cancellation/shutdown
semantics. Pre-cancelled empty input begins no packet and increments nothing;
this does not strengthen cancellation races when input is also ready. No retry,
batching, dropping, drain or observer-failure behavior is changed.

## Acceptance and direct evidence map

| Acceptance | Planned local evidence | Departmental direct execution, unrun |
| --- | --- | --- |
| Prior delivery and eligibility | Fresh clean refs; exact previous ranges/scope/note/index/MOC; full 369-file snapshot/phase/forward contract review | R90-75 stays independent |
| Count only successful terminal actions | Exact Worker helper diff; all three original calls preserved; observer success precedes sole increment | Public Worker.Run: no-alert/full-suppression/persisted success, each with/without observer, exactly once |
| Failure ownership and no completion | Unchanged processPacket, error/panic/cancellation/observer paths; terminal short-circuit | Run: writer failure; Arrival/Durable/Processed errors; terminal export failure in every success branch; matcher/redactor/observer panic; nil input and pre-cancelled empty input; no new completed count, legacy counters/alerts/call order preserved |
| Observable terminal seam/concurrency | Direct source review of synchronized observer return and shared Stats atomics | Hold Processed before return and read completed=0; release then completed=1; multiple workers counted after join without intermediate cross-counter invariant |
| Stats and public exports | Atomic/nil-safe/snapshot/Prometheus source; unchanged API/JSON projection and gauge source | Stats zero/nil/independent/concurrent increments/exact HELP TYPE value; HTTP /api/metrics positive and nil cases, existing counters/gauges/verbose-health JSON unchanged |
| Documentation and delivery | Go parse/format; pinned compile-only pipeline/stats/API; docs/JSON/full unique roadmap/history/links/fences/ten-path scope/diff/sensitive additions; feature/closure exact main/Git/Vault/stable notes/immutable preservation/identical replay | All suites remain delegated and unrun |

Every completion/failure case reaches public Worker.Run and terminal/error
boundary; manually incrementing Stats is not Worker evidence. Export cases reach
real public HTTP handlers. The terminal seam uses channel readiness, no fixed
sleep. Compile-only review checks test source but proves no runtime outcome.

## Risk, non-goals and stop conditions

Medium risk: conflating started work, write success, exported evidence and completed
processing. Keep legacy metrics and terminal semantics explicit; use atomic new
counter and direct source/regression review. Exclude health JSON/API migrations,
legacy metric rename/removal, SLO schema/policy changes, loss/throughput claims,
retries/batching/draining/shutdown, IPv6, dependencies/toolchains, test execution,
private traffic or external publication. Stop for ambiguous terminal ownership,
breaking public metric compatibility, required migration, competing edits or new
product/private/external authority. Complete exactly this increment; refresh
future queue without starting another feature. Initial review found no needed skill change; the concurrency evidence lesson is recorded below.


## Implementation, compile-only and static review checkpoint

Runtime source permits only the Worker terminal helper change and additive
Stats atomic/method/Snapshot/Prometheus fields, plus the accurate independent-
sampling comment. Exact source comparison proves all original Worker callers,
Run/processPacket/error/panic/observer/logging, legacy Stats/gauges, API/health
JSON and exporter paths unchanged. The new completion method is nil-safe;
observer success precedes the sole pipeline increment.

Eight direct regression functions cover 17 Run terminal/error/panic scenarios,
including each success branch with/without export and each branch's terminal
export error. The channel-synchronized Processed seam inspects zero before
return and one after success; no fixed sleeps. Nil input and cancelled empty
input begin no packet. Concurrent worker and independent Stats counts compare
after join. HTTP tests reach both real metrics and verbose-health handlers;
Stats tests require exact HELP/TYPE/value and unchanged independent old counts.
No helper-only test is counted as terminal completion evidence.

Pinned Go 1.26.8 compile-only pipeline/stats/API succeeds; binaries remain outside
the repository and were not executed. Go parse/format/docs, 163 task JSON,
149 complete unique roadmap pairs, complete unfinished contracts, unchanged
R90-75 Definition, ordered history, links/fences, ten-path scope, diff and
sensitive-addition review pass. Final comment edit gets the complete compile/static
chain again before commit. All execution suites remain delegated and unrun;
static/compilation does not establish runtime success or SLO acceptance.

Reusable workflow lesson: independent atomic loads can describe different
instants. The local netsentry-next skill's existing concurrency instruction now
requires synchronized quiescence before comparing cross-counter invariants;
Markdown/frontmatter/numbering/fences were checked. This generic skill edit is
outside the feature commit, adds no repository/private data and avoids false
concurrency failures. No scope or validation-ownership deviation. Feature and
one docs-only delivery closure remain; no subsequent increment starts here.


## Final pre-commit checkpoint

After the final Stats sampling comment, the complete pinned pipeline → stats →
API compile-only chain passed again; no binary was executed. The complete static
source/regression/format/docs/163 JSON/149 unique roadmap/unchanged R90-75/history/
links/fences/ten-path/diff/sensitive review also passed. Complete 369-file Vault
baseline is unchanged. This is development review only; all execution suites
remain delegated and unrun. No implementation/static/compilation failure occurred.
Commit exact ten paths, fast-forward fresh main and verify feature delivery;
then create only one docs-only closure. Skills contain the separately validated
generic concurrency improvement; no further skill change is needed.


## Delivery, deviation and queue closeout

Feature `1027b5a2ca7553e62c037c3e76c191ad140bf31f` contains exactly the planned ten paths. After isolated branch
implementation, fresh baseline verification and fast-forward main, push/fresh
fetch verified clean HEAD/origin/main/FETCH_HEAD at that full SHA. Exact range
`36d2cd49761308b9884cdbe47dd7ed9f5c2b2b42..1027b5a2ca7553e62c037c3e76c191ad140bf31f` has its generated ten-path scope, iteration note
`04-开发迭代记录/2026-10-02-1027b5a2ca-CI知识同步.md`, full index and MOC verified.
Fourteen current stable notes are reconciled; all 322 baseline immutable
iteration-directory notes retain their hashes. The resulting 370-file snapshot
JSON SHA-256 is `b9eff975bc08a86170716773c68f3178c286f852c2b99a691b9fa537175bd9d9`. Identical feature-range replay preserves
every Markdown hash. Local-only Vault discovery used the unique existing sibling
explicitly; no second Vault or remote artifact was created.

The sole reconciliation deviation was two stable notes whose unheaded substantive
topic prose shared the current-status replacement boundary. Source-grounded rule
ownership and configuration/management-transaction explanations were reconstructed
under explicit topic headings before closure; this is not a byte-for-byte recovery
claim. Immutable history was unaffected. The local skill's existing Vault
instruction now preserves topic prose and establishes explicit boundaries before
status replacement. This and the atomic-quiescence refinement were validated as
generic local-only Markdown changes, separate from the feature commit.

The runtime and direct evidence meet the planned boundaries: eight regression
functions cover 17 Run terminal/error/panic cases plus observer-return readiness,
nil/cancelled empty input, joined workers, Stats and HTTP compatibility. Final
pinned Go 1.26.8 pipeline/stats/API compile-only chain and full static review
passed; no generated binary or behavioral/race/CLI/full-suite/scanner/knowledge/
acceptance suite was executed. All such execution remains user-delegated.
Completion is process-local, not a loss oracle, durable-export receipt or SLO
gate; independently sampled atomics require quiescence for cross-counter checks.

One three-path docs-only delivery record closes this same increment. Its final
SHA is resolved from Git after commit, then push/fresh-fetch and exact feature-tip
to closure-tip Vault synchronization are verified before reporting. No second
closure is created merely to embed its own SHA.

Queue refresh: R90-145 implementation is complete; no further dependency-ready
local increment is currently defined. R90-75 retains its complete independent
asynchronous departmental acceptance contract and does not block development.
Next trigger verifies the fetched closure and Vault, audits fresh core-code
evidence and the forward queue, and persists a separate plan for any eligible
work before editing. No subsequent implementation begins here. Oct 2–Dec 30
horizon remains current; IPv6 remains separate product/protocol scope. Do not
repeat completed R90-144/R90-145 delivery or R90-59 publication.
