# R90-199: retain rejected supervisor probes

## Selection and audit

Selected 2026-10-25 from fetched origin/main
`b33c6211cffcfaf1c6b05fbbd60b92b10878788b`; exact hosted run
39355782002 is completed/success. Local delivery record and Vault iteration,
full index and MOC corroborate R90-198. Four-week history moved from correctness
repairs to full CI delivery, local-v1 policy, fixture and capture fixes. The
uncommitted R90-196 attempt stopped inconclusive; its records/raw data remain
in the original worktree and private evidence root. No missing delivery found.

Use a separate worktree; do not stage, rewrite or include those pending files.
Their clock failure and source prove the next ready repair. Queue: R90-199
selected, R90-196 blocked until repair plus stable-clock preflight and two fresh
runs, R90-197 planned after accepted L1, R90-75 deferred by existing authority.
Reconcile each unfinished Definition and the Oct25–Jan22 horizon in this change.

## Scope, acceptance and evidence

Retain initial and active clock/resource probes to samples.jsonl, with a checked
complete write and fsync before any clock/resource verdict can stop execution.
Start the generator only after the initial probe is persisted and admitted.
A failed write/fsync must abort without acceptance; retained files remain sealed
and owned children are reaped. Preserve schema, algorithm and 10ms threshold.

| Acceptance | Direct evidence |
|---|---|
| Initial rejected probe survives with no generator | execute regression reads exact raw row at verdict, rejects kernel probe and checks two owned children reaped |
| Active rejected step survives | execute regression reads exact initial/active probes, reconstructs identical invalid verdict and checks all owned children reaped |
| Resource rejection retains triggering row | execute regression rejects envelope after durable sample, checks exact resource values |
| Persistence errors fail closed | initial/active write, short-write and fsync failure regressions; no verdict on unpersisted sample, no further generator admission |
| Honest partial result and seals | each direct failure verifies inconclusive/incomplete status, no compliance/manifest, raw file size/hash seals |
| Preserve prior evidence and boundaries | compare raw/original-worktree inventories; no workload, host service, threshold, L1/L2, tag or release changes |
| Delivery | focused fixture/SLO checks, tool/module/Docker manifest preflight, full make ci-check, knowledge/diff/sensitive review, one batch non-force push, exact hosted success, Vault verification/replay |

## Risks and stop conditions

A partial write must not count as retained evidence; fsync failure cannot be
reported as a clock verdict. Extra fsync overhead is a diagnostic integrity cost,
not a latency guarantee. Test actual execute lifecycle using owned Python child
processes and synthetic probes; this is not L1 measurement. Stop on missing raw
probe, unsafe teardown, changed evaluator meaning, failed required gate, or need
for host/VM clock changes. Post-upload facts stay in a local delivery record and
Vault; do not upload a closure-only commit. Do not start the next increment.

## Required-gate deviation: patched toolchain

The first full gate stopped at reachable GO-2026-6809 in crypto/x509 on
Go1.26.9. No later check counts. The official release feed was retained before
changing pins: selected latest stable patch in the existing 1.26 line is
Go1.26.10 (other supported line 1.27.3). This required security prerequisite
is part of delivering this increment, without weakening the scanner.
Source: https://go.dev/dl/?mode=json and https://pkg.go.dev/vuln/GO-2026-6809.
Linux amd64 archive: 66,936,659 bytes; SHA-256
`ccfe7ae822d2c2f41f4fe7c956199967a4dac22d0ee0d2b2bfdb440aab3b20c5`.
Update go.mod/lock/current toolchain prose together, verify exact archive and
scanner, then rerun complete make ci-check. Historical 1.26.9 delivery records
remain unchanged.

Current README/acceptance/status prose still asserted no L1 execution. Reconcile
those claims from the retained failed attempt without copying or staging the
original pending files; update only this isolated worktree. Recheck affected
docs/knowledge/static boundaries after final delivery prose.

## Closeout against acceptance

All ten direct supervisor cases and focused suite passed; raw values, fsync
ordering, admitted clock prefix on I/O failure, generator admission, owned-child
reaping and result seals reach execute's actual boundary. Full patched CI passed
with 82.2% coverage and complete Docker runtime checks. Original raw/pending
inventories are unchanged. Deviations are limited to corrected test mocks,
a harmless wrong-cwd setup edit, the mandatory Go security patch, and current
status-prose reconciliation. No unplanned measurement or host change occurred.
Final docs/static/knowledge checks precede one completed batch. Exact fetched
remote/hosted/Vault results are recorded locally afterward; preserve original
pending attempt records. No next increment implementation begins.
