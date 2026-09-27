# R90-117: Native live UDP measurement ingress

## Selection and authority

Baseline `98541884f60a12adc1a605ce10e918ed5f0b1663` is clean and freshly
fetched equal to HEAD/origin/main/FETCH_HEAD. R90-116 feature/closure exist in
Vault index/MOC. Recent work progresses from SLO contract to report, ledger
adapter and engine export. R90-117 is next ready; R90-75 acceptance and R90-59
publication remain separate. User delegates all tests to the specialist
department; no tests, benchmarks, smoke executions, traffic or knowledge tests.
Static source/syntax/docs checks and compile-only builds remain allowed.

## Bounded design and acceptance

Implement one usable native IPv4/UDP measurement lane. Optional capture flags
require live Ethernet ingress, an explicit run ID, UDP destination port and new
capture summary file. Request PCAP_TSTAMP_HOST with microsecond precision;
use hdr->ts arrival, never callback time. Require inbound capture direction and
measurement BPF filter setup to succeed. Carry existing R90-116 `slo` metadata
through JSONL. Default capture formatting and runtime remain unchanged.

The generator prefixes UDP payload with ASCII `NSLO1 <run_id> <packet_id>\n`.
Both identifiers use the existing bounded grammar. The marker remains in the
matched payload; it changes the fixture and must be included in oracle/rate
qualification. Native capture accepts matching-run markers on the chosen port;
malformed/foreign/parse/UDS/kernel losses remain visible as counters or missing
oracle identities. Offline pcap mode cannot emit live measurement metadata.

Add a standard-library sender CLI/API for externally prepared JSONL fixture rows
with scheduled offsets, Base64 payload and independently expected rule IDs.
Build Ethernet/IPv4/UDP frames with checksums and stable sequential packet IDs.
Before each send attempt, append an R90-115 offered row using actual attempt
time and frame bytes, including minimum Ethernet padding but excluding FCS,
preamble and inter-frame gap. Log send success/failure separately; any failed
or ambiguous send prevents a completed submission receipt. Capture success is
never the offered denominator. Retain fixture copy, offers and send results with
hash/count-bound receipts; exclusive private outputs and no overwrite. This is
a correctness/instrumentation reference sender with unmeasured throughput.

| Criterion | Planned evidence |
| --- | --- |
| Live native correlation | Manual C review of marker bounds, timestamp overflow/precision, live-only flags, inbound/BPF and unchanged default format; compile only |
| Independent offered oracle | Python source/AST review of marker/frame/checksum construction, actual pre-send timestamps, retained missing/failed attempts and receipts |
| Complete handoff | Exact wire/byte/timestamp/eligibility contract, UDP-only scope, rate and clock limits, required department cases |
| Delivery | Static JSON/roadmap/link/diff/sensitive-data review, verified Git push/fetch and exact local Vault ranges |

## Intended paths and non-goals

Capture packet_types.h, new slo.h, main.c, uds_sender.c and capture/Makefile;
new scripts/slo_ingress.py and Makefile syntax registration; new docs/slo-ingress.md;
slo-runtime.md, slo-collect.md, performance-slo.md, architecture.md; roadmap,
this plan, new task-state-20260926-slo-ingress.json and active R90-75 state.
No actual interface, network namespace, service or traffic is started. No fixture
or completed-run artifact is generated. No TCP/mixed-workload certification,
20k-rule throughput claim, release/tag/CI change or new dependency. UDP fixture
support and software timestamps do not certify the production workload or
physical arrival boundary. All behavioral and fault verification is delegated.

## Delivery and next boundary

Persisted before edits. Preflight compiler/libpcap, implement, compile C to a
temporary path and AST-parse Python without running the modules. Deliver one
feature with verified remote and Vault, reconcile stable notes/replay, then one
docs-only closure. Queue the next bounded artifact/runner integration separately;
missing hardware or tests does not block this implementation.

## Implementation checkpoint

Implemented the 16 intended paths. A shared bounded C marker/ID parser carries
libpcap host timestamps into the engine metadata. Live-only setup checks reject
unsupported timestamp/direction/datalink/filter configuration; final capture
summary retains raw counters without claiming complete execution. Ordinary
packet JSON is unchanged when disabled. The reference sender builds checksummed
IPv4/UDP frames, keeps marker bytes in the inspected payload, records each offer
before send, retains independent fixture/oracle/submission logs and publishes
hash/count receipts only after successful submission and close. The marker uses
space separators because colon is legal in the shared ID grammar.

GCC 13.3.0/libpcap 1.10.4 compile-only build of all capture sources to a temporary
binary passed with C11, Wall/Wextra/Werror and O2. No binary was executed.
`make python-check` AST parsing and `git diff --check` pass. Manual review covered
length/NUL/ID/timestamp-overflow boundaries, live flags and PCAP precision,
JSON compatibility, IPv4/UDP checksum/MTU/padding construction, actual offer vs
scheduled times, partial sends, independent loss denominators and exclusive
retained output. Static structure checks parsed 134 states and matched all 122
unique roadmap row/Definition pairs; local links and ordered history agree.

All tests, CLI smoke, traffic, benchmark, acceptance and knowledge tests remain
not run under user delegation. No source-generated fixture or measurement
artifact is presented as evidence. Host timestamps and this UDP-only reference
lane require departmental qualification for the formal workload; sender/export
overhead is unmeasured. No skill change is needed; existing user-precedence
rules already cover this workflow. R90-118 is queued and not started.
