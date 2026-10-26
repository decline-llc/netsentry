# R90-199: retain supervisor probes before rejection

This repair concerns diagnostic integrity, not L1/L2 acceptance. Baseline is
`b33c6211cffcfaf1c6b05fbbd60b92b10878788b`. The old supervisor admitted its
initial probe only in memory and checked active probes before writing them.
The retained Oct21 failed L1 attempt is still inconclusive with a missing
terminal probe; this change cannot reconstruct that historical sample.

## Behavior and direct boundaries

Each initial/active clock and resource row now goes through a checked complete
write and fsync before resource-envelope or clock rejection. The generator is
started only after the initial row is persisted and admitted. An unsuccessful
write, short write or fsync aborts with an incomplete/inconclusive result; its
sample is not appended to the admitted clock sequence. Raw bytes are retained
and sealed even when partial or unsynced; a file seal alone proves no acceptance.
Clock algorithm, 10ms threshold, workload, packet/oracle generation and ownership
boundaries are unchanged. Per-probe synchronization adds evidence I/O overhead;
no timing guarantee is inferred from synthetic failure regressions.

Four new direct test methods exercise ten cases through `execute()` with real
owned Python child sessions and synthetic clocks/resources. They check exact
raw values and fsync before verdict, no pre-admission generator start, child
reaping before test cleanup, inconclusive flags, absent completion manifest and
matching raw-file size/hash seals. Cases cover initial failed kernel probe,
active 20ms clock step, pre-offer/active resource-envelope rejection, and
pre-offer/active write, short-write and fsync failures. Existing active-cancel
fixtures now return serializable resource samples for the newly retained initial
row. Substituted-child provenance binds the Python executable; no built SUT
binary is a test prerequisite. These cases do not transmit measurement traffic.

## Validation and required-gate deviation

Focused fixture/local-SLO: 32 methods passed. Knowledge: 33 methods passed.
Original nine pending paths and all 370 existing raw/validation files rehashed
unchanged. Independent worktree isolates this repair from the stopped attempt.
Roadmap raw row/Definition multisets: 199 equal, unique pairs.

Initial pinned-tool/module/Docker daemon/frontend/base preflight passed. The
first full gate then correctly blocked reachable standard-library
[GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809) on Go1.26.9. No later gate
counts. The retained [official release feed](https://go.dev/dl/?mode=json)
identifies Go1.26.10 as the latest stable patch in the selected line (other
supported line 1.27.3). Linux amd64 archive is 66,936,659 bytes with SHA-256
`ccfe7ae822d2c2f41f4fe7c956199967a4dac22d0ee0d2b2bfdb440aab3b20c5`;
exact size/hash verified before use. Module/lock and current toolchain prose move
together. Scanner and all complete-gate requirements remain enabled.

Complete patched-toolchain make ci-check passed exit 0: all nine locked assets
fetched and verified; zero reachable vulnerabilities; 32 fixture/local-SLO,
42 evidence and 33 knowledge methods; uncached full race and coverage (82.2%);
C parser/sender/15 entrypoint cases, ASan parser/formatter fuzz, E2E, archive and
complete Docker build/content/runtime health checks. Failed original scanner
log remains retained separately. Final source contains only this repair and its
required patched-toolchain prerequisite; affected docs/static/knowledge checks
are rerun after final delivery prose.
Full-gate log SHA-256: a45618950e9d4f85ffa32d03f58a53d57b85a9ee4ddb816cd2e7808756691d19.

This is one completed batch awaiting non-force upload and exact hosted/Vault
verification. Resolve the commit containing this evidence and inspect the local
R90-199 delivery record; only fetched origin/main equality and hosted
completed/success establish remote success. Post-upload facts stay local, with
no closure-only push. The independent worktree is the delivery checkout; the
original worktree keeps its baseline and all nine pending paths byte-for-byte.
No fresh L1/L2, host-clock setting, tag or release occurred.
