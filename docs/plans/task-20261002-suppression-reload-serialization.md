# R90-146: Serialize suppression reload reads and publication

## Selection, evidence and authority

Freshly fetched clean main/HEAD/origin/main/FETCH_HEAD is
`b2086d9769d2b253e443daf5fee28b4f7c4079a7`. Both R90-145 feature/closure
exact ranges, ten/three-path generated scope, iteration notes, index and MOC
are verified. The 371-file Vault snapshot JSON SHA-256 is
`6fc1f38a24b167ef66587ccf450b8e32d0b4a1a6525a3d30f7ee078753551ce8`;
14 current stable notes and 324 immutable iteration-directory notes are captured.
The 71-commit Sep 4–Oct 2 phase audit distinguishes measurement/reconstruction,
bounded input admission, repeated queue reconciliation, historical candidate
publication/main toolchain, rule ownership and packet completion. No missing
delivery or new qualifying R90-75 evidence appears. Previous Vault topic-boundary
repair is recorded; explicit topic headings now separate stable knowledge.

The local queue is empty. Core-source review finds ReloadFromFile reads and
validates the canonical suppression file before taking the manager mutation
lock. Add/Update/Delete can commit while that read is in progress, then reload
can publish stale rules over the successful mutation. Unlike the rule API's
complete transaction lock, suppression reload does not serialize its entire
read-to-publication transaction. Register bounded correctness repair R90-146 as
ready after R90-145 and existing R90-79 suppression persistence; persist this
plan/state before runtime or roadmap edits. No standalone audit increment.

R90-75 retains its complete independent asynchronous departmental contract and
does not block development. Horizon stays Oct 2–Dec 30. Behavioral/race/CLI/full-
suite/scanner/knowledge/acceptance execution is **not run; delegated by user**.
Author direct regression source, perform static/compile-only review and never
execute test binaries. Engine module preflight resolves pinned Go 1.26.8.

## Scope and implementation

Exactly seven paths: `engine/internal/alert/suppressor.go`, new
`engine/internal/alert/suppression_reload_test.go`, `docs/architecture.md`,
`docs/api-reference.md`, rolling roadmap, this plan and matching task state.
Local branch: `fix/r90-146-suppression-reload-serialization`. Keep nil-manager and
unconfigured-path guards; take the existing exclusive manager lock before the
authoritative load and hold it through validation/compilation/publication.
Use a private manager-local loader function defaulting to the unchanged public
LoadSuppressionsFromFile, following existing instance-local filesystem injection
patterns, so direct public reload tests observe the actual read boundary.
No exported API, JSON, schema, filesystem algorithm or persistence classification
changes. Existing public loader and replaceLocked bodies remain unchanged.

Reload errors preserve the prior active rules/filter and release the lock.
Missing files still clear the active set; reload never writes canonical bytes.
List/Filter retain their existing read lock and snapshot behavior; reload I/O
now holds the mutation lock longer and can delay these read-side operations.
No promise of lock-free suppression, cancellation or bounded reload I/O is added.
Only operations through the same manager serialize; standalone Save/Load calls
and external writers remain outside the guarantee. Cross-process coordination,
new atomic filter design, retries, migration, pipeline/rule/store/metrics,
IPv6, dependencies/toolchains, suite execution and publication are non-goals.

## Acceptance and evidence map

| Acceptance | Local static/compile evidence | Direct departmental regressions, unrun |
| --- | --- | --- |
| Baseline and queue authority | Clean fetched refs; exact prior Git/Vault; phase review; full unfinished contracts | R90-75 independent |
| Read-to-publication serialization | Guards then Lock/defer Unlock then real loader then unchanged replaceLocked; instance-local default loader | Public ReloadFromFile observes exclusive ownership at authoritative read; channel-pause after real load, start Add/Update/Delete, join both, compare exact canonical bytes, List and Filter |
| Successful mutation survives overlapping reload | All mutations retain same lock, complete persisted candidate semantics unchanged | Three separate add/update/delete cases compare expected final disk/list/filter, including stale prior and replacement addresses |
| Reload errors preserve state and unlock | Existing loader diagnostics and compiler failures unchanged; deferred Unlock covers all returns | Read sentinel, malformed JSON, invalid set, invalid enabled CIDR; canonical bytes unchanged, old List/Filter retained, subsequent successful mutation |
| Empty and defensive boundaries | Nil/unconfigured guards before lock/load; missing-file loader and defensive copies unchanged | Public nil/unconfigured reload; missing-file clearing without creation; returned List nested slices mutated without changing active filter |
| Delivery and scope | Go parse/format; pinned alert/API/pipeline compile-only, docs/JSON/full unique roadmap/history/links/fences/seven-path/diff/sensitive review; exact feature/closure Git/Vault; stable knowledge/immutable preservation/idempotence | All execution suites delegated |

Synchronization uses observable loader readiness and channels with outer deadlines,
without sleeps or scheduling assumptions. Lock observation is inside the loader
invoked by public ReloadFromFile and does not substitute for final file/list/filter
assertions. Static and compilation evidence does not prove runtime/race success.

## Risk, stops and delivery

Medium correctness risk: stale publication and extended lock duration. Preserve
all public errors and committed/uncommitted persistence boundaries. Stop for
required exported configuration changes, incompatible loader diagnostics,
cross-process writer guarantees, ambiguous ordering, competing user edits or new
external/product authority. Finish exactly this increment. On review success,
commit seven paths, freshly verify baseline, fast-forward main, push/fetch, sync
exact range and reconcile stable prose. One three-path docs-only closure records
verified facts, receives exact Git/Vault verification and refreshes the queue
without beginning another implementation. Existing skill instructions cover the
observed transaction and knowledge boundaries; no skill edit is warranted yet.


## Implementation and final static checkpoint

Exact source comparison permits only the default-real private loader field/
constructor initialization, reload comment and existing Lock/defer relocation
before the loader. Nil/unconfigured guards precede the lock; all errors release
it. Public LoadSuppressionsFromFile, replaceLocked, Add/Update/Delete, List/Filter,
checked durability code and all API/store/rule/pipeline/Stats/exporter/config
source remain unchanged. Every manager constructor supplies the real loader.

Five direct regression functions reach public ReloadFromFile. Read callbacks
verify exclusive ownership before real LoadSuppressionsFromFile. Three separate
channel-held real-read cases start Add/Update/Delete, join both operations and
compare exact canonical bytes, full List and source/destination Filter results
against expected final candidates. The lock assertion exposes the original
missing-lock boundary independently of goroutine scheduling; the started signal
is not a claim of FIFO lock acquisition or proof that a contender is queued.
Four distinct read/parse/set/compile failures check sentinel/diagnostic preservation,
unlocked state, old rules/filter, unchanged input bytes and a later successful
mutation. Nil/unconfigured and missing-file clearing/absence plus all four nested
List slices retain defensive behavior. No fixed sleeps or helper-only substitute.

Pinned Go 1.26.8 complete alert → API → pipeline compile-only chain passed;
binaries remain outside the repository and unexecuted. Go parse/format/docs,
164 task JSON, 150 complete unique roadmap row/Definition pairs, full forward
contract, unchanged R90-75, ordered history, links/fences, exact seven paths,
diff and sensitive additions pass. Compilation/static review does not prove
runtime/race behavior; all execution suites remain delegated and unrun.

The complete 371-file Vault baseline is unchanged. Full content backups of the
14 stable notes and their explicit topic boundaries were captured so status
updates can prove original topical prose retention; 324 immutable history hashes
will be checked. Existing skills already cover readiness/transaction and topic
preservation; no local skill edit is needed. No implementation/static/compile
failure or scope deviation occurred beyond the recorded empty-queue repair.
Complete exact feature delivery then one three-path docs-only closure; do not
begin a subsequent increment.
