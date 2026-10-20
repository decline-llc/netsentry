# R90-198 measurement capture buffering and idle shutdown

This is local tooling/repair evidence on the existing Ubuntu VM, not L1/L2
acceptance or a production capacity claim. Contract/workload/oracle thresholds
are unchanged. Baseline is 7d01b1abed5f0b02b2acd97b50adfb147ed68473.

## Diagnosis and behavior

R90-195 retained ~692/694ms (other runs ~616–781ms) host-arrival-to-WAL-return
samples and an idle-stop timeout. Its measurement capture used buffered blocking
pcap_dispatch with a 1000ms packet-buffer timeout; the fixture later sent a
separate control frame for teardown. These observations are consistent with
waiting on buffer delivery, not established detector/storage capacity failure.
No exact per-stage attribution is claimed from the old aggregate samples.

Libpcap 1.10.4 upstream documents [packet-buffer timeout limitations](https://github.com/the-tcpdump-group/libpcap/blob/libpcap-1.10.4/pcap.3pcap.in),
[immediate delivery](https://github.com/the-tcpdump-group/libpcap/blob/libpcap-1.10.4/pcap_set_immediate_mode.3pcap.in)
and [nonblocking dispatch](https://github.com/the-tcpdump-group/libpcap/blob/libpcap-1.10.4/pcap_setnonblock.3pcap).
A packet-buffer timeout does not guarantee an idle read returns. Measurement-only
setup now requires immediate mode before activation, nonblocking mode afterward
and a selectable descriptor. Empty nonblocking dispatch polls for at most 50ms;
EINTR retries through the signal flag, other poll errors fail measurement.
Queued packets dispatch without an intervening idle wait. Inbound Ethernet,
host microsecond timestamps, mandatory filter and exclusive/fsynced summary
remain. Summary adds immediate/nonblocking/poll metadata. Ordinary live/offline
capture retain their read/setup paths. Under downstream UDS blocking or CPU
starvation, 50ms is not an end-to-end latency or shutdown promise.

Fixture shutdown now sends only SIGTERM and records monotonic stop durations;
there is no wake/control traffic. Deadline escalation still kills and reaps only
the owned child and retains an incomplete outcome.

## Direct acceptance coverage

| Planned boundary | Direct evidence |
| --- | --- |
| Immediate before activation; host timestamp/direction/filter unchanged | Native real-entrypoint stubs assert ordering, values and mandatory UDP filter |
| New setup rejection and cleanup | SC_IMMEDIATE_FAIL, SC_NONBLOCK_FAIL, SC_FD_FAIL: exit 2, both resources closed, no dispatch/summary |
| Poll failure rejection | SC_POLL_FAIL (EIO), SC_POLL_ERR/HUP/NVAL: exit 2 and capture_closed=false |
| Interrupt and empty-wait retries | SC_INTERRUPT, SC_TIMEOUT each require a second dispatch/poll |
| Cancellation before readiness | SC_EARLY_CANCEL raises SIGTERM during activation: zero dispatch/poll, closed summary/resources |
| Idle and active cancellation | SC_IDLE zero packet; SC_ACTIVE_CANCEL dispatches real marked Ethernet/UDP through parser, asserts timestamp/identity and one successful UDS packet |
| Ordinary/offline preservation | SC_ORDINARY/SC_OFFLINE use original opens, no immediate/nonblock/poll calls |
| Real idle before first eligible packet | Real binary in new owned rootless veth namespace; readiness then zero transmission, SIGINT and SIGTERM separately; zero capture_received/sent/ignored and successful summaries |
| Owned deadline cleanup | Python test_stop_deadline_still_reaps_owned_child retains SIGTERM-ignore kill/reap assertion; existing early-cancel/replacement fixtures retained |
| Low-rate real pipeline | Two independent 10pps/2s + 3s observation smokes, frozen 100 rules, independent SQLite/API and lifecycle accounting |

Fifteen native scenarios passed ordinarily and under ASan. Normal C outputs were
cleaned/rebuilt after sanitizer mode; 28 Python fixture/local-profile methods
passed. The first native compile exposed enum names colliding with signal.h;
scenario names were prefixed and the full focused chain rerun. No checks were
removed. Generic local skill now requires idle-before-input cancellation with
no injected wake traffic; its edit is separate from this repository increment.

## Retained local observations

| Run | Offered/submitted | Arrival/processed | Oracle/durable/SQLite | Host arrival to WAL return | Capture stop |
| --- | --- | --- | --- | --- | --- |
| r90-198-smoke-01 | 20/20 | 20/20 | 2/2/2 | 5.982ms, 4.747ms | 32.64ms |
| r90-198-smoke-02 | 20/20 | 20/20 | 2/2/2 | 6.768ms, 4.754ms | 33.73ms |

Both runs preserve exact identities with zero missing/extra/duplicate events,
zero capture drops/parse errors/invalid markers, valid shared-clock interval
probes and read-back separate cgroup budgets/affinities. These four alert samples
are not a percentile or sustained-load result. Idle probe SIGINT/SIGTERM each
closed within one second with no captured packet and no forced kill. Ethernet
readiness probing occurred before capture startup. Absolute UTC/NIC accuracy and
between-probe paired clock steps are not established by the sampled clock bound.

Raw artifacts remain local outside Git/Vault under the explicit budgeted root;
source seals, all raw bytes/hashes, idle-probe source/log and validation logs are
retained in the local delivery record. Historical R90-195 raw inventory was
rehashed without modifying it. Raw-inventory manifest SHA256: d544b1f1b7e44f57ee686723e6e07b262e310c5b74cc9d301964f60d0bca5553.

## Complete gate and delivery

Pinned Go 1.26.9, actionlint 1.7.12, govulncheck 1.6.0, engine module and Docker
daemon/frontend docker/dockerfile:1/base ubuntu:24.04 preflight succeeded.
Full make ci-check passed exit 0 with 82.3% all-package coverage: nine fetched
locked assets, no reachable vulnerabilities, 28 SLO/fixture + 42 evidence + 33
knowledge methods, uncached race/coverage, ordinary C entrypoint/parser/sender,
both ASan fuzz gates, E2E, archive and unmodified full Docker build/content/health.
Full log SHA256: `291d2c5e19ab1e0dc8ea354c33051fb4314bf9f768b6e25d82b34ccdb6c995f1`.
Source/test/build hashes match the frozen source and both actual smoke binaries.
Only final delivery prose is added after that gate; affected static/knowledge
and sensitive review are rerun before commit. All 202 row/Definition multiset
pairs are unique and complete; historical definitions and frozen workload,
legacy evaluators, engine/runtime outside capture are preserved.
One completed batch push follows successful checks; exact hosted and Vault facts
remain local, without an extra closure push. R90-196 is next only after verified
completion; no L1/L2 started in this increment.
