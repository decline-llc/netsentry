# R90-200: fixture clock-method metadata compatibility

Baseline: `7de6fdf534834f8629cda06bf42755da96001297`. This is a no-send
synthetic software repair, not live L1/L2 acceptance or real-data evaluation.

## Defect and behavior

The Oct26 L1 retry completed its first 600+30s acquisition, but public derive
rejected the original manifest with `clock_method: expected bounded ASCII
identifier`. The producer copied a descriptive paragraph from clock_verdict
into that identifier field. An earlier manifest test used an independently
authored `synthetic` method and did not reach the actual producer's output.
The new direct producer-to-public-derive regression reproduced the exact error
before the implementation change; its failing log is retained locally.

clock_verdict now returns `host-realtime-monotonic-brackets-v1` as `method`.
Its original explanatory text is retained verbatim as `method_description` in
the raw result; the manifest copies only the identifier. The unchanged algorithm
calculates max(realtime - monotonic_before) minus
min(realtime - monotonic_after), plus 1000ns pcap quantization and maximum clock
resolution. Kernel probe failure or candidate bound above 10000000ns still
rejects the clock. Absolute NTP error remains separately retained and common to
both endpoints; the sampling limitation still applies. No consumer validation,
legacy schema, clock arithmetic, threshold, workload or host setting changes.

## Direct regression boundaries

- Actual frozen declaration/oracle/rules, actual clock_verdict output and actual
  collection_manifest flow through public derive and the versioned collector
  and evaluator. Sparse synthetic per-minute ledgers require no sockets or child
  processes. Assertions verify schema 2 / local-v1 / exact run and method identity,
  exact 2001ns uncertainty, ten derived events, unchanged input bytes and an
  inconclusive, non-acceptance report because offered load is deliberately sparse.
- Historical descriptive metadata still fails through public derive with the
  same bounded-identifier diagnostic; no adapter is created and all source/result
  bytes remain unchanged. Rejected historical manifests are never repaired.
- Clock regression checks the exact method/explanation, 2001ns bound, failed
  kernel probe, drift and clock step, plus inclusive 10000000ns and rejected
  10000001ns boundaries. The legacy golden regression stays unchanged.

## Validation and delivery

Focused fixture/local-SLO: 34 methods passed. Knowledge: 33 methods passed.
Pinned Go1.26.10 official archive size/SHA verified against the repository lock;
actionlint1.7.12, govulncheck1.6.0, engine module, Docker daemon and exact syntax
frontend/base manifests passed preflight. An initial archive-check invocation
used the module cwd with a root-relative lock path; it failed before verification.
The root-relative verification was rerun successfully, then the complete
fail-fast tool/module/Docker preflight passed. No failed-chain test counts.
Full make ci-check completed successfully: nine freshly fetched and verified
locked assets, zero reachable vulnerabilities, 34 fixture/local-SLO, 42 evidence
and 33 knowledge methods, uncached full Go race/coverage (82.2%), C parser/sender/
entrypoint tests and ASan fuzz, E2E, archive and complete Docker image/content/
runtime health checks. Full gate log SHA-256:
`01f28dabdab4f127fdeda9e86399de822a34d9b219a28258bdeb0e2aacf2ccfa`.
A daemon restart lost the command-session handle while the original gate process
continued; its retained log reaches final Docker health, RC success and normal
make completion. No duplicate gate, upload or test result was inferred from an
incomplete log. Source/evaluator seals still match the validated bytes.
Final preservation/static review verifies all 443 original files unchanged,
204 equal unique complete roadmap identifier pairs, JSON, links, Markdown fences,
sensitive diff and unchanged clock-calculation/threshold AST. Only final delivery
prose follows the full gate; affected docs/static/knowledge checks are rerun.

Both earlier pending worktrees and original measurement bytes are isolated from
this increment. The local preservation inventory covers 443 existing files.
User has no real traffic corpus; current tests use synthetic fixtures and the
existing public locked corpus. Private-real-data, production capacity and
physical-NIC/power-loss claims remain unverified. No new L1/L2 is run here.

Resolve the eventual commit containing this evidence and its local common-Git
R90-200 delivery record. Completion requires the full local gate, non-force
batch upload, exact fetched SHA hosted completed/success and verified local
Vault note/index/MOC/stable prose/replay. Post-upload facts stay local without
a closure-only push. Then R90-196 needs two new independent frozen accepted
runs before R90-197 stability/recovery; old attempts remain non-acceptance.
