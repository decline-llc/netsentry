# Native live UDP measurement ingress

R90-117 connects a bounded IPv4/UDP fixture lane to the
[engine lifecycle exporter](slo-runtime.md) and [ledger adapter](slo-collect.md).
Implementation is **untested by explicit user direction**. Only compilation and
static source/syntax/document review were performed. No traffic or acceptance
run was executed. The [SLO acceptance contract](performance-slo.md) still applies.
R90-133 adds the sender fixture boundary with static AST/source/docs review only;
behavioral, CLI, traffic and knowledge suites remain delegated and unrun.

## Wire contract and eligibility

An untagged Ethernet/IPv4/UDP frame carries this exact ASCII prefix in its UDP
payload, followed immediately by the fixture's original payload bytes:

```text
NSLO1 <run_id> <packet_id>\n
```

Here `\n` means one LF byte, and each separator is one ASCII space. IDs follow
`[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}`; spaces/NUL/newlines are forbidden inside IDs.
A colon is a legal ID character, so it is not a separator. The marker remains
in the inspected payload, changes packet length and can affect rule matching.
Prepare the independent rule oracle for the **marked** fixture. Capture never
creates expected alerts or allocates replacement IDs for lost packets.

The reference sender numbers fixture rows from one as `pkt-1`, `pkt-2`, etc.
The oracle's event ID uses the R90-116 algorithm:
`"slo_" + hex(SHA256(packet_id + NUL + rule_id))` over UTF-8 bytes. All submitted
eligible packets, including those never captured or processed, remain in the
offered ledger. The capture-side marker match is not the denominator.

This lane supports untagged Ethernet, IPv4 without fragmentation and UDP on one
explicit destination port. The sender sets DF, computes IPv4/UDP checksums and
supports marker plus original payload up to 1472 bytes (IPv4 MTU 1500). Short
frames are padded to 60 Ethernet bytes. The eligible byte count includes Ethernet,
IP, UDP, marker, original payload and Ethernet padding; it excludes FCS, preamble
and inter-frame gap. No VLAN, TCP, mixed traffic or production-rule coverage is
established by this lane. The proposed 512-byte mean must be recomputed using
this exact frame boundary, not assumed from original application payload size.

## Capture mode

Department invocation example (not executed):

```bash
bin/netsentry-capture -i "$CAPTURE_INTERFACE" -s "$ISOLATED_UDS_PATH" \
  --slo-run-id run-001 --slo-port 19075 --slo-output capture-summary.json
```

All three `--slo-*` options are required together with live `-i`; `-r` is rejected.
The summary pathname must be new and its parent must exist. Ordinary capture
without these options retains its existing packet format/behavior.

Measurement setup requires each of these steps to succeed:

- `pcap_create`, explicit `PCAP_TSTAMP_HOST`, microsecond precision, activation
  without warnings, Ethernet datalink and inbound direction (`PCAP_D_IN`).
- An installed `ip and udp dst port <port>` BPF filter.
- Exclusive creation of a mode-0600 capture summary file.

On the selected lane, capture extracts and validates the marker/run ID, retains
its payload, and attaches the existing engine `slo` metadata. Arrival nanoseconds
are exactly `hdr->ts.tv_sec * 1e9 + hdr->ts.tv_usec * 1000`, with range/overflow
checks. It does not substitute callback time or historical replay timestamps.
Capture, UDS and worker queue delay after the host timestamp remains in E2E
latency. The microsecond field is converted to nanosecond units; this does not
create nanosecond accuracy.

Host packet timestamps have platform-dependent delay/precision and may not mark
physical NIC arrival. This code cannot prove time spent before the host timestamp
or clock agreement with the engine. The department must characterize the actual
capture boundary, timestamp quantization, clock uncertainty and VM/NIC behavior
before claiming the formal live-arrival SLO. Unsupported timestamp/direction
configuration fails rather than falling back to callback timestamps.

Unmarked/foreign-run or non-lane packets increment `ignored`; malformed markers
increment `invalid_markers`. Parse and UDS failures retain their existing
counters. Invalid timestamp arithmetic stops measurement with a nonzero exit.
Missing/corrupted expected packets remain in the independently retained oracle.
A chosen ingress and isolated topology must ensure each offered packet crosses
one inbound capture point; sending on the same interface that sees only outbound
frames will not satisfy this. No interface, namespace or network is configured
by these tools.

On shutdown, `capture-summary.json` records run/port, requested timestamp and
direction policy, sent/drop/parse/marker counters and raw `pcap_stats` counters.
Counter meanings depend on libpcap/platform; they are diagnostic, not unique
eligible loss denominators. `capture_closed` describes capture-loop completion;
`execution_complete` and `slo_compliance_asserted` remain false. Check process exit
status too: late write/fsync/close/directory errors may leave a visible summary.
Errors preserve partial output and never authorize reusing the same pathname.

## Reference sender and independent oracle

Prepare a finalized UTF-8 JSONL fixture externally, one object per packet:

```json
{"offset_ns":0,"payload_base64":"","expected_rule_ids":[]}
{"offset_ns":1000000,"payload_base64":"aGVsbG8=","expected_rule_ids":["rule-7"]}
```

These lines illustrate format, not measured evidence or a verified rule oracle.
Only these three fields are accepted. Offsets are nonnegative, nondecreasing
integers relative to the shared UTC run origin, bounded to seven days. Each line
must fit 256 KiB; payload Base64 is validated, and expected rule IDs must be
bounded, unique within the packet and supplied independently of detections.
Generated ledger rows also obey the adapter's 256 KiB limit.

Department invocation example (not executed):

```bash
python3 scripts/slo_ingress.py \
  --interface "$SEND_INTERFACE" --fixture fixture.jsonl --output-dir sender-run \
  --max-fixture-bytes 68719476736 \
  --run-id run-001 --origin "$UTC_ORIGIN" \
  --src-mac "$SOURCE_MAC" --dst-mac "$DESTINATION_MAC" \
  --src-ip "$SOURCE_IPV4" --dst-ip "$DESTINATION_IPV4" \
  --src-port 19074 --dst-port 19075
```

This command actually sends raw packets when the department invokes it. Linux
AF_PACKET requires appropriate privileges. It uses explicit MAC/IP/port values;
it performs no ARP, routing, interface configuration or topology setup. Sender,
capture and engine must use the same run identity; origin must be canonical
whole-second UTC in `[2000,2100)`, equal to the final manifest `started_at`.
Start capture/engine before the offer schedule and retain actual readiness.

The public API `send_fixture(fixture, output, run_id, origin, link, send, *,
max_fixture_bytes=DEFAULT_MAX_FIXTURE_BYTES)` accepts
Path inputs, six link fields (`src_mac`, `dst_mac`, `src_ip`, `dst_ip`, `src_port`,
`dst_port`) and a caller-owned `send(bytes) -> full_frame_length` function.
`frame(...)` exposes frame construction. A custom sender must also retain its
own transport/interface configuration and prove submission semantics.

R90-133 snapshots the finalized source before the first send. The Python
`max_fixture_bytes` keyword and CLI `--max-fixture-bytes` set a positive integer
limit in `[1, 2^63-1]` (default 64 GiB); Python booleans are rejected. Invalid
budgets fail before output creation, and the CLI validates the budget before
socket setup. The limit bounds source fixture bytes only, excluding generated
ledgers, metadata and total workspace storage. Record the chosen limit with the
departmental invocation; schema-v1 receipt fields remain unchanged.

The supplied final pathname is opened once with non-following/nonblocking flags
and must identify a regular file. Missing, directory, FIFO and symlink inputs
fail acquisition before any packet submission. A known size above the budget
is rejected before copying. Retention admits at most the budget and enforces
the existing 256 KiB raw-row limit before writing. The snapshot writer checks
short writes, flushes/fsyncs and closes; device/inode/size/mtime/ctime metadata
and consumed size must remain consistent before the source closes. Failure
retains available partial evidence and creates no completed submission receipt.

Submission reads only the new mode-0600 `fixture.jsonl`, through one non-following
regular-file handle. Its metadata is compared before reading and at EOF; exact
bytes, rows and SHA-256 must agree with the captured snapshot before completion.
Only one bounded raw row is retained in memory at a time. Metadata observes
changes but does not authenticate acquisition or guarantee continuous mutation
exclusion. Parent traversal and local output manipulation are not authenticated
filesystem boundaries. Source bytes are never repaired or deleted.

The sender streams the retained fixture, waits for scheduled offsets, and records
each actual offer attempt before calling send. Snapshot preparation can increase
lateness; it preserves origin and requested offsets and does not prove adequate
generator throughput. The `offered_ns` timestamp precedes
ledger serialization/write and kernel submission; inspect `send_start_ns`,
`send_return_ns`, schedule/lateness and frame hashes in the separate submission
ledger to bound that delay. Offer cohorts use actual attempt initiation times,
not requested schedule offsets. Late senders can move packets into later minutes
or outside the declared run; the report/department must reject deficient load.

Before any send, its entire oracle row is written to an unbuffered offered file.
A send exception or short submission records failure when possible and aborts
without a completed submission receipt. A crash or interrupt also leaves partial
files. **Do not feed an incomplete sender attempt ledger into a qualifying run**:
failed/ambiguous submissions must be investigated, never silently removed to
improve loss. Even successful socket send proves kernel acceptance, not NIC
transmission; retain independent generator/NIC offered-load evidence for final
qualification. Retained-row validation may discover a malformed later fixture row
after earlier traffic; that run is incomplete and must be retained as such.

## Retained artifacts and downstream handoff

A new mode-0700 sender directory contains:

| File | Meaning |
| --- | --- |
| `fixture.jsonl` | Complete bounded source snapshot before sending; a prefix may remain on failed acquisition |
| `offered.jsonl` | Every initiated offer with bytes and all expected alerts; direct adapter input |
| `submissions.jsonl` | Submission result, timestamps, schedule/lateness and frame hashes |
| `submission.json` | Written last after successful submissions and ledger sync/close; hashes, counts, link configuration and sender source digest |

The receipt always keeps `execution_complete=false`, `slo_compliance_asserted=false`
and departmental review required. Sender completion is not an acceptance pass or
proof of capture/drain completion. Retain the capture summary, engine events and
close receipt, sender bundle, adapter bundle and final report together. Prepare
the final manifest only after verifying actual offered load, run/drain bounds,
resource/clock/durability requirements and external artifact consistency.

```bash
python3 scripts/slo_collect.py --manifest completed-manifest.json \
  --offered sender-run/offered.jsonl --events engine-run/events.jsonl \
  --output-dir retained-run
python3 scripts/slo_report.py --input retained-run/observations.json
```

Existing output is never overwritten. Inspect errors/partial files, use a fresh
output location after resolving failures, and verify all receipt hashes. A late
sync error can leave a visible complete receipt; process success and retained
content must both be reviewed. Scripts do not delete run evidence automatically.

## Unexecuted departmental validation

Required cases include ordinary capture compatibility; live-only flag combinations;
unsupported timestamp/direction/datalink/BPF failures; min/max/colon-containing
IDs, embedded NUL and malformed markers; timestamp overflow and microsecond
conversion; checksum/MTU/DF/padding byte boundaries; independent oracle/event IDs;
background/foreign/duplicate packets and capture/UDS drops; read/send/short-write/
fsync/close/interrupt faults; no-overwrite/partial preservation; clock/schedule
changes and delayed submission; adapter/reporter consistency and complete source
retention. R90-133 adds unexecuted cases for ordinary/space paths; input admission
and source mutation/replacement; invalid/bool/exact/over budgets, the signed
64-bit integer ceiling and row limits;
zero-send acquisition rejection; permission and snapshot/replay inventory checks;
retained-output replacement/mutation; snapshot preparation latency; strict
bundle/sender-replay compatibility; partial preservation and I/O/close faults.
Validate actual generator headroom, NIC counters, long-window alert
sample counts and complete hardware/workload profiles separately.

No test, sender, capture command or acceptance run was executed by the agent.
The Python sender, per-packet ledger and serialized runtime exporter have
unmeasured overhead and do not demonstrate 3/7 Gbps or either proposed profile.

## Cross-artifact handoff

Use the [supplied evidence bundle checker](slo-bundle.md) to retain and compare
sender/capture/engine/adapter/report sources, receipts and completion declarations.
It reports missing inputs and mismatches without running traffic or asserting
compliance. For this UDP lane, the manifest fixture digest identifies the exact
retained sender `fixture.jsonl` bytes. Behavioral validation remains delegated.
