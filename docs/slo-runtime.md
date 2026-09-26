# Opt-in engine SLO lifecycle export

R90-116 adds an engine-side measurement export feeding the
[raw ledger adapter](slo-collect.md). This implementation is **untested by user
instruction**; behavioral and acceptance validation belongs to the specialist
department. A compile-only build does not establish runtime correctness or SLO
capacity. The [formal acceptance contract](performance-slo.md) is unchanged.

## Enable for an isolated measurement instance

The ordinary engine does not export lifecycle events. All three flags are
required to enable export:

```bash
./netsentry --config isolated-config.yaml \
  --slo-output-dir engine-run \
  --slo-run-id run-001 \
  --slo-origin 2026-09-26T12:00:00Z
```

This is an invocation example, not an executed run. `engine-run` must not exist;
its parent must exist. The origin must be canonical whole-second UTC in
`[2000,2100)` and equal the eventual manifest's `started_at`. Use a fresh isolated
SUT state and freeze actual SUT/collector commits, resource allocations, rule and
configuration digests, and clock error bounds. Export does not create the offered
packet ledger, fixture oracle or completed-run manifest. There is no SSH runner.

Enabled export requires WAL storage. Each primary/shard SQLite connection,
including replacement connections, receives `synchronous(FULL)` through an
escaped file URI; initialization also verifies effective WAL and synchronous=2.
Non-WAL measurement configuration fails. Ordinary storage settings are unchanged
when export is disabled. Physical device/VM fsync behavior still requires review.

## Ingress metadata and correlation

A valid ordinary UDS packet frame may carry this additional object:

```json
{"slo":{"run_id":"run-001","packet_id":"pkt-000001","arrival_unix_ns":1790424000001000000}}
```

This fragment illustrates the added field; it is not a complete packet frame or
measurement evidence. Existing hello/session, IPv4, payload and timestamp rules
still apply. The receiver validates supplied run/packet IDs against
`[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}` and requires positive int64 arrival nanoseconds.
The exporter additionally requires every processed packet to contain metadata,
the matching run ID, and arrival at/after origin and no later than observation.
The ordinary receiver's JSON decoding policy still applies; this extension does
not introduce canonical JSON or duplicate-member rejection for UDS frames.

`packet_id` must come from the offered oracle and survive ingress/capture/queue
handling. `arrival_unix_ns` must be obtained at **live packet arrival**, on the
same documented Unix wall-clock basis as engine timestamps. Do not substitute
worker-start time, historical offline-PCAP timestamps or component histograms.
The worker emits the supplied arrival value, so subsequent capture buffering,
UDS and queue delays remain inside measured latency. Packets that never reach
the worker have no emitted arrival and remain missing in the offered oracle.

**Native C capture does not yet emit `slo` metadata.** An instrumented ingress
must supply it for this engine feature. R90-117 is queued to connect native live
capture and generator identities. Do not label direct UDS injection as proof of
live packet arrival. The current pre-match processed metric is still unchanged
and is not the new terminal-success observation.

For each expected packet/rule pair, the oracle must use:

```text
event_id = "slo_" + lowercase_hex(SHA256(UTF8(packet_id) + byte(0) + UTF8(rule_id)))
```

Both identities use the bounded grammar above, which excludes NUL. The exporter
uses the full 32-byte digest. This is a measurement correlation ID, **not** the
store's `event_id`, aggregation key or row ID. Storage continues its existing
aggregation/deduplication semantics; retain database evidence and reconcile
fixture expectations against those semantics. Do not derive expected events
from the exporter: missing detections would vanish from the denominator.

## Emission boundaries

| Row | Timestamp and condition |
| --- | --- |
| `arrival` | Supplied live-arrival time minus origin; emitted when a worker begins observing the packet |
| `durable` | Engine wall clock immediately after successful `WriteBatch` return, minus origin; one row per non-null persisted alert |
| `processed` | Engine wall clock after successful packet processing; includes no-match and suppression-completed branches |

JSONL keys exactly match R90-115 (`kind`, `packet_id` or `event_id`, `at_ns`).
Durable time is a conservative **upper bound** on database durable completion:
it includes batch/shard commits, recovery-log cleanup and return overhead. It is
not a timestamp of the hardware flush itself. A partial storage operation that
returns an error emits neither durable rows nor a terminal marker for that
packet; the retained database/recovery log needs separate failure review.

No-match or fully suppressed packets get terminal markers but no durable rows.
If the independent oracle expected an alert, the adapter still counts that
missing alert and excludes the packet from successful processing. A matcher,
suppressor, redactor or writer panic before completion does not manufacture a
terminal marker. Shutdown cancels workers according to the existing lifecycle;
it does not drain every queued packet automatically. Retain the complete offered
ledger and set the manifest's actual observation/drain boundary accordingly.

Concurrent workers share a serialized exporter. Row order across packets is
not timestamp order; the adapter accepts shuffled observations. Repeated rule
IDs within a persisted batch are rejected. There is no unbounded runtime packet
identity set: repeated packet IDs across batches are rejected by the adapter.
Unknown/unexpected durable IDs also cause adapter rejection and must be examined
against the oracle; they are not silently filtered out.

## Files, failures and overhead

The output directory is created with mode 0700 and files with mode 0600:

- `events.jsonl`: raw lifecycle records; no payloads, IP addresses or credentials.
- `close.json`: exclusively published after workers stop and events are synced
  and closed; binds raw bytes by SHA-256, row count, byte count, run ID and origin.

`export_closed=true` only means the exporter completed its file lifecycle.
`execution_complete=false`, `slo_compliance_asserted=false` and
`departmental_review_required=true` are always retained in this receipt. A clean
engine shutdown is not a completed acceptance run. Retain this receipt together
with the adapter bundle and final report; verify the receipt hash equals the
adapter's consumed `events.jsonl` hash. The department supplies the manifest
only after validating actual run completion and the full offered cohort.

File/write/short-write or export metadata errors are sticky, cancel the enabled
engine context and prevent publication of a successful close receipt. The engine
reports an error and exits nonzero after worker shutdown. Existing output is
never overwritten; partial data remain for investigation. A failure after close
receipt publication can leave a visible complete file with uncertain directory
sync; an error does not imply that the receipt is absent. Startup failures can
leave an empty/partial export directory with no receipt. Do not reuse it.

Disk writes occur on the worker path; all workers contend on the export mutex.
File synchronization occurs at close, not after each event. Arrival export delay
and filesystem contention can affect matching and downstream latency; durable
export occurs after the captured durable timestamp and affects subsequent work.
Export overhead, disk requirements and production throughput are unmeasured.
A filesystem stall can delay shutdown. No background drop queue hides export
loss. System-clock discontinuities that violate arrival/order constraints are
rejected locally or by the adapter; other clock distortions require external
calibration and uncertainty accounting.

## Departmental validation still required

- Disabled behavior and existing UDS frames; enabled flags, path preservation,
  malformed/missing/run-mismatched metadata, ID limits and origin boundaries.
- Real upstream timestamp/correlation proof; delayed queues, clock jumps and
  propagated missing packets; multi-rule event-ID agreement with the oracle.
- No-match, suppression, redaction, panic, failed/partial store writes, duplicate
  packets/rules and unexpected alerts; counters and missing-event denominators.
- Concurrent workers and shutdown before readiness/under load; no writes after
  close; write/short-write/sync/close/link/directory failures and retained bytes.
- WAL/full-synchronous primary and daily shards, replacement connections, encoded
  path characters, recovery and aggregation/deduplication correctness; physical
  persistence under the actual VM/storage deployment.
- Recompute raw hashes and adapter/report input; repeated/long acceptance runs,
  raw alert and deadline counts, telemetry coverage, export cost and disk usage.

None of these checks has been executed by the agent. Native ingress wiring and
both profile acceptance runs remain outstanding; development can continue.
