# R90-196: two fresh L1 runs after metadata repair

Selected 2026-10-30 from fetched origin/main
`2bdff2d3bf2d6da80eca3bafc0ed9e9f910d5399`. Exact hosted run 41172644641
is completed/success; local R90-200 delivery and 479-file Vault replay agree.
Four-week audit shows correctness, fixture/capture, rejected-probe retention and
metadata handoff repairs with full gates. Prior failed attempts and both dirty
worktrees remain preserved, not part of this delivery. All 204 complete roadmap
identifiers pair uniquely; unfinished Definitions contain all required fields.

## Scope, frozen parameters and authority

Execute only R90-196 with new IDs r90-196-l1-05 and r90-196-l1-06. Each gets new
namespace, port, UDS, database and output. Run/review the first before the second.
100 rules, 512-byte frames, 600s at 100pps with 500pps during seconds 180–239 and
480–539, 108000 scheduled packets, 600 independent oracle events, 30s observation.
Actual per-minute offered load +/-5%; each sustained/burst loss <=1%; zero missing,
extra or duplicate alerts; clock bound <=10ms; full/sustained/each-burst nearest-rank
p99 plus clock bound <=500ms. SUT 4 CPUs/3GiB, generator 2 disjoint CPUs/1GiB,
no swap, >=8GiB free disk and cumulative evidence/copies/scratch/logs <=4GiB.
All thresholds freeze before transmission. Synthetic owned-veth, shared single VM;
no private real corpus is required. Production/physical NIC conclusions remain
unverified. No runtime/evaluator/workload/clock-service change, L2, tag or release.

## Acceptance mapped to evidence

| Criterion | Evidence |
| --- | --- |
| Metadata accepted before long runs | Actual frozen declaration and clock output through derive and bundle/reconstruction in an explicitly synthetic no-send fixture; no acceptance certificate |
| Stable clock and isolated resources | 60s retained 100ms probes, namespace-aware interface inventory, cgroup limits/affinities and pinned tool/module/Docker preflight |
| Two independent workloads | Separate frozen/oracle/config/source/binary hashes and all intent/offered/submission rows; independent full schedule/frame reconstruction |
| Correct alerts and storage | Exact lifecycle/oracle identities, independent URI-encoded read-only SQLite event mapping/integrity/per-rule counts, API and close receipts |
| Load/loss/latency and validity | Every event/packet, all cohorts and denominators, initial/active clock probes, actual sampling gaps, resource peaks/errors and graceful-stop times |
| Reviewed evidence | Original seals, retained adapter, fresh versioned reconstruction, explicit projection proof, local raw context and seven checks bound to exact input/bundle/context/manifest hashes |
| Complete delivery | Focused/knowledge checks, complete make ci-check after measurement, sanitized evidence/docs, one batch upload, exact hosted/Vault verification |

The legacy component envelope uses labelled post-run projections only: check
every null submission error before omission, preserve capture extensions in a
hash-bound projection proof, and keep original bytes. No legacy sender-frame
replay claim; full independent frame/oracle review plus versioned adapter replay
establish the stated boundary. Local context is explicit contract-relevant raw
facts, not a production hardware-envelope claim. Retain reviewer/source hashes.

## Risks and stop conditions

Shared-host scheduling, clock synchronization and fsync contention can invalidate
measurement. Stop on any numerical failure, invalid clock, missing artifact,
configuration drift, resource/budget overrun or unsupported contract error.
Retain every failure and original input. Do not retry to select favorable runs or
lower thresholds; scope any source-proven repair as a later increment. A stable
preflight cannot exclude later clock steps; disclose actual maximum probe gap.
Do not run expensive gates during acquisition. Raw artifacts stay local under
one owned root lock/reservation. Preserve original pending worktrees by hashes.

## Closeout

Compare every criterion to evidence, record deviations, refresh the queue without
starting L2. Full validation precedes one non-force batch; fetch exact SHA and
verify hosted completed/success and Vault note/index/MOC/stable/replay. Post-push
facts stay local without closure-only upload. No acceptance claim from partial
metrics, smoke, synthetic preflight or a reporter return code alone.

## Actual local result

Both l1-05/l1-06 independently passed complete raw, SQLite/API, resource/clock,
versioned reconstruction and hash-bound review: 108000/108000 packets and600/600
alerts per run, zero segmented loss/missing/extra/duplicates. Worst cohort p99
upper bounds are9.095959ms and8.987318ms; both well within500ms. No code, workload,
threshold or host clock change. The private record generator's requirement for
unique sequential namespace inode numbers was corrected against fresh-construction
source and independent run-state evidence; original failed script/diagnostic and
all artifacts are retained. This is a reviewer correction, not an SUT repair.
Acceptance mapping is complete for two local L1 runs; required full CI and exact
delivery still remain. L2 is neither executed nor accepted here.

## Full gate and delivery resume

Complete make ci-check passed exit0: nine freshly fetched locked assets, zero reachable vulnerabilities,34 fixture/local-SLO/42 evidence/33 knowledge methods, uncached race/coverage 82.3%, C/ASan fuzz, E2E/archive and full Docker build/content/runtime health.
Historical/raw and review bindings,complete roadmap multiset,paths/JSON/Markdown/
sensitive checks pass. Final prose is followed by affected checks. Resolve this
commit and common-Git R90-196-final record; exact fetched SHA hosted success and
Vault verification complete delivery. Never repeat05/06 or upload closure facts
when already verified. R90-197 L2 starts in a later trigger; no L2 result exists.
