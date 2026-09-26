# SLO measurement report tool and test-department handoff

`python3 scripts/slo_report.py --input observations.json` summarizes a completed
run supplied by the test department. It executes no traffic and collects no
runtime data. Implementation is delivered with **tests deferred by the user's
2026-09-25 instruction**. No behavioral test or acceptance execution is claimed.

## Invocation and output

```bash
python3 scripts/slo_report.py --input observations.json
# Optional explicit NEW output path:
python3 scripts/slo_report.py --input observations.json --output acceptance.json
```

The default destination is:

```text
~/benchmarks/r90‑75/acceptance_{staging|prod}_{YYYYMMDD_HHMMSS}.json
```

The directory retains the user-specified U+2011 hyphen. The filename timestamp
comes from the run's `started_at` in UTC, not the report-generation time. An
existing pathname is never overwritten. JSON is written to a private temporary
file, flushed/synchronized and published using an exclusive hard link on the
same filesystem. Directory synchronization follows publication. If publication
succeeds but directory synchronization fails, the CLI reports an error and
leaves the published file for inspection; do not assume an error means no file
exists. No report is created for a structurally invalid or unfinished run.

The JSON embeds the exact UTF-8 input text under `source.raw_json` and its
SHA-256 under `source.sha256`. Retain the collector's complete raw logs,
configuration, clock/durability proof and packet identity ledgers separately
as required by the [acceptance contract](performance-slo.md). Embedded minute
counts are not a substitute for those underlying ledgers. Reports may contain
local measurement metadata; they remain local evidence by default.

| Exit | Report status | Meaning |
| --- | --- | --- |
| 0 | `review_required` | No implemented numerical check flagged a problem; departmental review still required |
| 1 | `measurement_failed` | Supplied complete measurements violate an implemented latency, missing-alert or loss check |
| 2 | No new report guaranteed | Input/schema/filesystem error; inspect destination if publication had already occurred |
| 3 | `inconclusive` | Coverage, declared run/sample policy, offered load or checked resource fields are insufficient; numerical failures are still retained |

Every report sets `slo_compliance_asserted=false` and
`departmental_review_required=true`. A zero exit is **not SLO acceptance**.
The tool does not activate any CI or release gate.

## Input schema v1

Input is a single UTF-8 JSON object, at most 64 MiB. All listed fields are
required; unknown fields, duplicate object members and non-finite JSON numbers
are rejected. Integer fields reject booleans and must fit a nonnegative signed
64-bit range unless a stricter minimum is specified. Identifiers are 1-128
ASCII characters: start with a letter/digit, then letters/digits or `_.:-`.
This schema is for minute counters plus relatively low-volume expected alerts,
not one JSON record per offered packet at million-packet-per-second rates.

| Root field | Type / contract |
| --- | --- |
| `schema_version` | Integer `1` |
| `run_id` | Identifier |
| `profile` | `staging` or `prod` |
| `started_at` | Valid UTC timestamp `YYYY-MM-DDTHH:MM:SSZ` |
| `duration_seconds` | Positive multiple of 60 |
| `drain_seconds` | Nonnegative declared post-offer drain duration |
| `observed_through_ns` | Exactly `(duration_seconds + drain_seconds) * 1000000000` |
| `execution_complete` | Boolean `true`; this is supplied by the collector, not independently attested |
| `policy` | Fields below, fixed before execution |
| `measurement` | Fields below |
| `resources` | Fields below |
| `provenance` | Fields below |
| `packet_windows` | Exactly one ordered record for every minute, including empty/incomplete cohorts |
| `expected_alerts` | Every event predicted by the frozen workload/rule oracle, including missing events |

`policy` contains only:

- `minimum_duration_seconds`: integer greater than 300. It enforces the chosen
  extended-run floor, not a scientifically established tail sample requirement.
- `minimum_expected_alerts`: positive integer. The department chooses and
  records this floor before a run; sparse p99 still requires statistical review.

`measurement` contains only:

- `live_arrival`: boolean asserting the required live-arrival timestamp boundary.
- `durable_completion`: boolean asserting the durable-persistence boundary.
- `clock_method`: bounded identifier naming the method documented with raw logs.
- `clock_uncertainty_ns`: nonnegative upper bound on elapsed-time error after
  clock normalization. This is added to each finite measured latency. It is the
  bound for the interval, not an unspecified per-clock error.

Either false boundary assertion produces `inconclusive`. True assertions cannot
prove the instrumentation is correct; the department must inspect the evidence.
All event times below are integer nanosecond offsets on the same normalized
run timeline. Offline PCAP timestamps and component timers cannot substitute.

`resources` contains only `sut_vcpus`, `sut_memory_bytes`, and `active_rules`,
all positive integers. The current checks require exact agreement with:

| Profile | SUT vCPU allocation | SUT memory allocation | Active rules |
| --- | ---: | ---: | ---: |
| staging | 2 | 4294967296 | 2000 |
| prod | 8 | 17179869184 | 20000 |

A mismatch makes the analysis inconclusive, but does not prevent writing a
report. These are reported allocations, not local-host autodetection. SSD/NIC
properties, pinning, isolation, generator headroom, traffic distribution, alert
rate, rule complexity and configuration truth are **not verified by this tool**.
Runs at a different profile remain useful observations without qualifying the
proposed profile. Development does not wait for matching local hardware.

`provenance` contains only `sut_commit`, `harness_commit` (40-character lowercase
hex full Git SHAs), and `config_sha256`, `rules_sha256`, `fixture_sha256`
(64-character lowercase hex). Shape is validated; artifact existence/content
must be reconciled by the department against retained raw evidence.

### Packet minute cohorts

Each `packet_windows` object has exactly these fields:

| Field | Meaning |
| --- | --- |
| `start_second` | `0`, `60`, `120`, ... in exact order; no gaps or duplicates |
| `phase` | `sustained` or `burst`; each burst is one isolated 60-second cohort, with no adjacent burst cohorts |
| `offered_eligible` | Unique eligible packets offered in this minute, selected before the run |
| `fully_processed` | Unique packets from that same cohort fully processed by the declared final drain boundary |
| `offered_bytes` | Bytes in those offered eligible packets under the frozen byte-counting convention |
| `observation_complete` | Whether the collector fully reconciled this cohort |

Completed counts may not exceed offered counts. Byte counts must be consistent
with empty/nonempty cohorts and at least one byte per offered packet. The tool
cannot independently prove uniqueness from aggregate counts; retain the
collector's identity ledger. Incomplete observation, empty offered cohorts or
missing either phase produce an inconclusive result.

The tool checks the offered byte rate against 100/250 Mbps for staging and
3/7 Gbps for prod, by phase; below-target load is inconclusive. It does not
certify rate pacing, packet-size distribution, generator saturation or an
upper load bound. Failures observed at excess load need contextual review and
are not proof of failure at the nominal target.

### Expected alert events

Each `expected_alerts` object has exactly:

| Field | Meaning |
| --- | --- |
| `event_id` | Unique expected-event identifier |
| `packet_id` | Offered-packet identifier |
| `rule_id` | Expected matching rule identifier |
| `offered_ns` | Integer offset in `[0, duration_seconds * 1000000000)` |
| `arrival_ns` | Live-arrival offset, or `null` if no arrival was observed |
| `durable_ns` | Successful durable-completion offset, or `null` if missing by final drain |

Event IDs and `(packet_id, rule_id)` pairs must be unique. Multiple rules for a
packet must share the same offer/arrival timestamps. The number of distinct
alert-bearing packets cannot exceed the offered packet count of their minute.
Durable timestamps require an arrival; timestamps must satisfy
`offered <= arrival <= durable <= observed_through` when present. Pending or
failed alerts remain in the list with `durable_ns=null`, including packets
whose arrival is also missing. Do not replace absent completion with zero or
omit its expected event. The tool cannot reconstruct an omitted oracle event;
compare the oracle/hash and full trace during departmental review.

## Calculation boundaries

Each minute and each trailing five-minute window is an **offer-time cohort**
finalized at the run's declared drain deadline. These retrospective reports
are not live operational window counters. Completed alerts that finish after
the offer window still belong to their original cohort, retaining their full
latency. Window boundaries cannot hide queued work or losses.

Latency is `durable_ns - arrival_ns + clock_uncertainty_ns`. Missing completion
is positive infinity in the conceptual latency distribution. Nearest-rank p99
uses rank `ceil(0.99 * expected_alert_count)`, including every missing event.
JSON never emits numeric infinity: `p99_kind=unbounded_missing` with a null
`p99_upper_bound_ns` represents an unbounded rank. `no_samples` also has null
p99 but an explicit different kind. A finite p99 may coexist with missing
samples; missing events still create failures and deadline violations.

The output includes expected, durably completed, missing, late and deadline-
violation counts beside p99 and its sample count. Exactly 180 ms is on time;
finite durations above it are late. Missing plus late equals deadline
violations. Missing expected security alerts are always recorded as failures.

Loss is `(offered_eligible - fully_processed) / offered_eligible`, with separate
sustained/burst denominators. Comparisons use integer arithmetic against 500
and 2000 parts per million, respectively. The tool checks phase loss and p99
for the full run and trailing five-minute windows; it also checks every
individual burst's loss. Reports retain all one-minute cohorts and raw input.
Coverage gaps take status precedence over numerical failure flags without
removing those flags. Even an otherwise clean result requires evidence review.

## Departmental validation handoff — not executed

The user delegated testing. This implementation has only static syntax, diff
and manual source review; the following behavioral validation remains with the
test department:

- Valid staging/prod reports, exact targets, threshold equality, high/low load,
  resource mismatch, sparse/empty samples, short runs and absent phases.
- Missing arrivals/completions, finite p99 with rare missing samples, unbounded
  p99, exactly/just beyond 180 ms and nonzero clock uncertainty.
- Late completion across minute/window boundaries, final-drain cutoff,
  sustained/burst denominators, isolated burst failure and incomplete cohorts.
- Every schema rejection: malformed/duplicate JSON, booleans as integers,
  overflow, unknown/missing fields, bad SHA/timestamp/identifier, window gaps,
  ordering/phase/adjacent bursts, bad counters, duplicate events and conflicting
  packet timestamps, impossible time ordering and too many alert-bearing packets.
- Source-byte digest reconstruction, Unicode destination, existing-file and
  symlink collision preservation, serialization/write/link/sync failures,
  incomplete-run rejection and truthful non-success/exit behavior.
- Instrumentation and real-load acceptance against the full
  [contract](performance-slo.md), including retained identity ledgers,
  hardware/clock/durability proof and extended tail evidence. This summary tool
  does not supply those measurements or constitute a qualifying run itself.
