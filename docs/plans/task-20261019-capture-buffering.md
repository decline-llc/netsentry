# R90-198: bound measurement capture buffering and idle shutdown

Baseline: clean fetched HEAD/origin/main/FETCH_HEAD
7d01b1abed5f0b02b2acd97b50adfb147ed68473. R90-195 exact hosted run
39192719403 completed/success; local delivery record and both Vault iteration
ranges/index/MOC/stable note agree. Four-week phase audit shows bounded
correctness repairs, then complete quality gates and local-v1 tooling. The
hosted unbuilt-binary and allocator prerequisites were repaired with a clean
full gate; preserve their failed history. Current stable status is authoritative;
older current-status blocks still contain superseded R90-195 resume prose and
will be reconciled without losing topic content. No user changes were present.

Forward queue: R90-195 complete, R90-198 selected, R90-196 depends on this
repair, R90-197 depends on accepted L1, R90-75 deferred. Each has dependency,
window, risk, acceptance, required validation and stop condition. Refresh horizon
to Oct19-Jan16; forecasts do not gate work.

## Acceptance and planned evidence

1. Explain historical ~692/694ms samples and idle timeout from retained raw
   evidence and libpcap buffering/read behavior. Preserve previous raw files;
   a packet-buffer timeout does not bound an idle read or per-packet latency.
2. Measurement-only immediate delivery, nonblocking reads and bounded poll
   make buffer delivery and idle signal handling independent of wake traffic.
   Keep inbound Ethernet, host microsecond timestamps, required filter and
   exclusive durable summary. Direct native tests exercise setup ordering and
   each new failure (immediate/nonblock/selectable fd/poll), interrupted poll,
   cancellation before readiness and during work; ordinary/offline calls retain
   their paths. Include tests in normal and sanitizer C gates.
3. Remove the fixture control wake dependency; record graceful stop elapsed
   times. Repeat independently identified owned-veth low-rate smoke runs and
   retain exact oracle/offer/submission/arrival/terminal/WAL/SQLite/API accounting,
   raw latency and clock/resources. Exercise an idle-before-first-packet signal
   stop using the real capture binary in an isolated namespace with a local UDS
   receiver. Smoke is tooling evidence only, never L1/L2.
4. Preflight exact Go/tools, engine module, Docker daemon/frontend/base manifest;
   focused checks, full make ci-check, knowledge/static/sensitive checks. Batch
   one non-force upload, fetch exact SHA, exact hosted completed/success and
   full-SHA Vault sync/stable reconciliation/replay. Post-push facts local only.

## Scope, risks and stop conditions

Own capture main/new direct native tests and build targets, fixture teardown/test,
current measurement docs, roadmap/plan/state/evidence. No detector/storage,
legacy evaluators, workload/threshold changes, L1/L2 run, release/tag/registry,
public interface, sudo or external coordination. CPU cost may rise with immediate
mode; retain resource data and leave capacity to L1. A selectable descriptor is
required for measurement mode; unavailable support fails closed. The poll bound
is a scheduling interval, not a hard end-to-end latency or shutdown guarantee
under UDS backpressure. Stop upload on failed/ambiguous required checks; stop
measurement on missing identity/timestamp/direction/accounting/clock proof.


## Final acceptance comparison

Each planned boundary reaches the direct native scenario, real idle probe or
real pipeline smoke named in the evidence table. All 15 native scenarios pass
ordinarily and under ASan; normal outputs were cleaned/rebuilt afterward. The
28 Python methods and two independent exact-accounting low-rate smokes pass;
four alert samples range 4.747–6.768ms, capture stops 32.64/33.73ms without wake
traffic. Real zero-capture SIGINT/SIGTERM stop in 33.53/16.98ms before first input.
These are tooling observations, not an L1 percentile or acceptance result.
All 226 historical raw artifacts rehash unchanged. Source/executable provenance
matches both smoke invocations and the final gate.

Complete ci-check passed with pinned tools, all-package uncached race/coverage
(82.3%), scanner, C/ASan fuzz, 28/42/33 tooling/evidence/knowledge methods, E2E,
archive and full Docker content/health. Only closeout prose follows, with affected
static/knowledge checks. Initial native enum-name compile collision was corrected
before the successful focused chain; no required gate failed or was bypassed.
The generic skill was refined for idle-before-input cancellation without wake
traffic. No scope/threshold change, L1/L2, next implementation or publication.

Resolve the feature commit containing this plan, then use the local delivery
record and fresh refs to verify exact hosted success and Vault/stable/replay.
Do not repeat a verified push/sync or send a closure-only upload. R90-196 becomes
ready only after these facts; this trigger ends after R90-198 delivery.
