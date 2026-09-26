# R90-116: Opt-in engine lifecycle export

## Authority and selected boundary

User again explicitly delegates all tests to the specialist department and asks
for implementation progress. Do not run tests, knowledge tests, benchmarks,
traffic or CLI smoke checks. Source/format/static structure review and compilation
without execution are allowed. Baseline `1dd5dac2923fee51d7af4b43ae1fd0695c33a88d`
is freshly fetched equal to HEAD/origin/main/FETCH_HEAD and clean. R90-115 feature
and closure are present in Vault index/MOC. September history now progresses
from contract to report to raw-ledger adapter. R90-59 publication and R90-75 real
acceptance remain separate; missing hardware does not block development.

Implement R90-116 at the Go engine boundary: optional packet measurement metadata
(run ID, oracle packet ID, live-arrival Unix nanoseconds), propagated through
UDS/queue to worker. CLI flags explicitly enable a new private export directory
and freeze run origin. Export arrival, successful durable-write return and
terminal processing rows compatible with R90-115. Derive correlation event IDs
from packet/rule identities; this is distinct from storage aggregation IDs.
No runtime export is enabled by default. Native C capture currently cannot supply
the oracle identity; native capture/generator integration is the next increment,
not a reason to defer this usable instrumented engine boundary.

For enabled measurement, require WAL and configure synchronous=FULL for every
SQLite connection via escaped DSN, including daily shards; verify runtime pragmas.
The durable timestamp is a conservative upper bound at successful WriteBatch
return. Hardware/VM persistence and clocks require departmental validation.
Export errors are sticky, cancel the measurement engine and prevent a successful
close receipt; partial files remain. Healthy shutdown syncs raw events and emits
a checksum-bound close receipt, which is not run completeness or SLO compliance.
Panic and storage failure must never invent durable or terminal success.
Suppression can complete processing without a durable alert; independent expected
events still remain missing in the oracle. Cancellation does not imply drain.

## Scope and acceptance evidence

- Model/receiver: optional validated metadata; ordinary frames remain usable.
- Pipeline: optional observer before matching, after successful WriteBatch and
  after terminal processing; no success marker after failed/panicked processing.
- New `engine/internal/measurement/export.go`: concurrent JSONL export, run/clock
  validation, stable correlation IDs, exclusive output, sticky errors, raw hash
  and synchronized close receipt. No unbounded per-packet identity map.
- Storage: opt-in full-synchronous WAL DSN on primary and shard connections.
- Main: flags, lifecycle wiring, cancellation and close after workers stop.
- Docs: `docs/slo-runtime.md`, adapter/SLO/architecture links, roadmap and active
  R90-75 state; this plan and `task-state-20260926-slo-runtime.json`.

Manual source review, gofmt, compile-only Go build, JSON/roadmap/link/diff and
sensitive-data checks map to structural acceptance. No behavioral pass is
claimed. Document a departmental matrix for enabled/disabled paths, invalid
metadata, concurrent workers, write/sync/close faults, panic/storage failures,
clock changes, suppression/missing alerts and shard/reconnect durability.

## Risks, non-goals and delivery

Export overhead is unmeasured. Arrival metadata is supplied, not physical proof;
wall clock jumps require detection/review. Packet duplicates are caught by the
adapter, not an in-memory runtime set. A collector close receipt does not prove
all offered packets were processed. No traffic generation, C parser changes,
release/tag/image action, CI weakening or performance compliance assertion.
Persist this plan/state before implementation. Deliver one focused feature,
verify push/fetch and exact Vault range, reconcile stable notes and replay, then
one docs-only closure. Queue the remaining native ingress integration without
starting it. Stop only for ambiguous delivery or a new authority requirement.

## Implementation and review checkpoint

Implemented 15 intended paths: seven engine source files and eight documentation/
state files. The opt-in exporter writes correlated arrival/durable/processed
JSONL, retains a hash/count-bound close receipt and cancels on sticky export
errors. Pipeline instrumentation preserves missing observations after failures;
no-match/suppression completion does not create durable events. Native ingress
identity remains queued separately. Measurement storage uses escaped FULL pragma
DSNs for primary and daily-shard/replacement connections, with initialization
verification; the locally installed pinned driver's documented `_pragma` behavior
was inspected directly. No new dependency or CI configuration changed.

`go build -o /tmp/netsentry-r90-116 ./cmd/netsentry` completed successfully from
the engine module. The binary was not executed. gofmt is clean. Manual source
review covered enabled/disabled paths, observer lifecycle ordering, concurrent
writes, fail/cancel/close behavior, JSONL shape and correlation IDs, wall-clock
limits, storage DSN propagation and no-overwrite retention. Static checks parsed
133 task-state JSON files and matched all 121 unique roadmap row/Definition
pairs, resolved changed-document links and confirmed ordered history/diff hygiene.

Tests, test compilation/execution, CLI smoke, benchmarks, acceptance traffic and
knowledge tests were not run, as the user again instructed. Compile/static
success is not behavioral evidence. Departmental validation is explicitly listed
in `docs/slo-runtime.md`. No SLO/profile capacity or independent hardware claim
is made. No generic skill change is needed: user authority already overrides
skill-local test gates. R90-117 remains unstarted.

## Delivery results

Feature `9b7262f88c1bfde8d5f984b0e159f6821eb5d130` contains the 15 intended paths.
Go 1.25.14 compile-only build, gofmt, manual source and static JSON/roadmap/link/
diff/sensitive-data review passed. No binary execution or tests occurred.
Behavioral, knowledge, benchmark and acceptance validation remain delegated.

Push without force/tags succeeded and fresh fetch verified clean
HEAD/origin/main/FETCH_HEAD equality. Exact range
`1dd5dac2923fee51d7af4b43ae1fd0695c33a88d..9b7262f88c1bfde8d5f984b0e159f6821eb5d130`
was synchronized to the local Vault. Note/index/MOC and seven reconciled stable
notes were verified; iteration history was preserved. Replay retained Markdown
SHA-256 `1a8c7be0e874faf59eba7a8fa8e828be65a1e1276be55a85166987769cd6a7db`.

R90-116 implementation is complete with tests delegated. The engine boundary
consumes supplied live-arrival metadata; R90-117 native ingress/oracle wiring is
ready for a separate persisted plan, not started. This engine implementation is
not full live-ingress integration or acceptance evidence. R90-75 remains the
department's measurement outcome, not a development prerequisite. This single
docs-only closure receives static review and its own verified Git/Vault range.
