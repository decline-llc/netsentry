# R90-195: bounded local Ethernet fixture

Baseline: freshly fetched clean main fe4d2dbb22368325a6c636c8026e9c984a5e442f.
R90-194 hosted run 39107211112 completed/success; local delivery record,
Vault iteration/full index/MOC and stable local-v1 topic agree. Recent four-week
history progressed from bounded API/storage repairs through the superseding full
CI delivery rule to current-resource acceptance tooling. Historical test waivers
are inactive. R90-75 remains deferred; R90-195 is the highest dependency-ready
item; R90-196/197 depend on its tooling and then actual L1 evidence respectively.
All unfinished Definitions specify window, risk, acceptance, validation and stop.

## Acceptance mapped to evidence

1. Freeze exactly 100 distinct enabled payload rules, 512-byte marked Ethernet/
   IPv4/UDP frames, deterministic independent per-second oracle and full 600s
   100/500pps schedule (108000 frames, 600 positives). Unit tests independently
   parse checksums/headers/keywords/oracle/hashes and schedule totals. Dry-run
   writes the frozen contract without opening a socket or starting a SUT.
2. A rootless user/network namespace owns loopback and a new veth pair only;
   never accept an interface argument or enter the host network namespace.
   SUT and generator are separately affinity/quota/memory constrained through
   delegated cgroup v2. Verify the actual controls, inbound Ethernet readiness,
   API/rule count and capture startup before send. Direct cancellation/early
   failure/readiness tests and an owned-veth low-volume smoke establish this.
3. Require >=8GiB disk; account all files/copies/scratch under one explicit
   evidence root against cumulative 4GiB. Reserve bounded per-run capacity,
   constrain child file growth and count files/bytes periodically. Fail closed
   on unsafe paths/links, exhausted budget or missing controls; retain partial
   ledgers, scheduled/actual offers and submission errors. Test budget boundaries,
   startup failures and ownership-preserving cleanup including neighboring files.
4. Retain clock realtime/monotonic bracket samples and resolution, resource and
   cgroup samples, exact binary/source/config/rule hashes, lifecycle/capture logs,
   oracle/offer/submission ledgers and SQLite/API observations. Clock proof must
   account for host timestamp precision and steps; unknown proof is inconclusive.
   No automatic passing claim: smoke is tooling evidence only, L1/L2 unexecuted.
5. Focused fixture tests then pinned-tool/module/Docker manifest preflight,
   knowledge-check and full make ci-check, static roadmap/JSON/links/diff and
   sensitive review. One completed non-force push; fresh exact refs, exact hosted
   success, full-SHA Vault sync/stable reconciliation/idempotent replay. Persist
   post-push facts locally only. Do not begin R90-196 in this trigger.

## Scope and authority

Own new fixture runner/tests, their Makefile gate, fixture runbook, current tooling
status references, roadmap and this plan/state/evidence. Avoid changes to detector,
legacy SLO schemas/evaluator, release workflows or dependencies. No sudo, host
network mutation, external interfaces, private data, new hardware, release/tag,
L1/L2 execution or threshold relaxation. Rootless unshare and user-systemd delegated
cgroups were proven available; the user bus requires explicit XDG_RUNTIME_DIR.

Stop upload on any required failure or incomplete acceptance mapping. Preserve
partial evidence and diagnose; do not weaken checks. Stop execution on ambiguous
ownership, missing resource/clock evidence or insufficient disk. A measured defect
may require a separately bounded roadmap repair rather than changing thresholds.

## Implementation checkpoint

Sixteen direct fixture regressions plus twelve local-profile methods passed.
The retained independent pre-send oracle and budgeted derive wrapper now cover
copies/scratch and exact raw hash replay. Low-rate smokes have exact
20-frame/two-event accounting and enforced/read-back resources; ~616–781ms delays
were retained even though final-source smoke-07 measured ~74ms.
The earlier ~692/694ms sample and first idle shutdown failure are retained too.
Add only the bounded R90-198 Definition/row as a prerequisite before R90-196; do
not implement it here. Smoke is a completed tooling path, never L1 acceptance.
A separate-run control frame makes the fixture's idle teardown deterministic
without changing capture runtime. Actual buffering/shutdown repair is future work.

Repeated startup/cancellation regressions reach real child processes, preserve
regular/symlink replacements and forward CLI cancellation to the isolated
supervisor. The generic skill now requires namespace-aware interface enumeration
rather than inherited sysfs or inaccessible other-process namespace links.
Complete pinned-tool/module/Docker preflight passed; full ci-check is running.

Final review adds a clock probe before generator startup, continuous drift
rejection and a monotonic outer deadline, plus retained disk/CPU/memory headroom.
The first complete gate passed but is superseded by these changes; repeat the
complete gate. No faster later smoke is selected to hide prior delayed samples.

Wake-error cleanup now waits/kills/reaps the child even when control traffic or
its ledger fails. The direct fault injection passed; final-source smoke-08 also
completed. The superseded second gate was deliberately terminated, with its raw
log retained, rather than allowing obsolete validation to authorize upload.
All 28 focused methods pass. Runtime/legacy sources remain byte-identical.

## Final acceptance comparison

All sixteen direct fixture boundaries reach their named public/API/process or
resource boundary as mapped in the evidence table. Full final-source ci-check
passed, including Docker and 82.3% coverage, after three exact uncached receiver
reruns and a clean complete-suite retry. No unrelated runtime fix was made.
Earlier source-invalidated and real setup-failure logs remain, with raw hashes.
Final source/Makefile/test hashes match the retained smoke-08. No L1/L2 was run.
Only delivery prose is added after the full gate and receives affected static
and knowledge review; one completed batch upload follows. Exact hosted/Vault
facts and resume reconciliation stay local without another closure push.
