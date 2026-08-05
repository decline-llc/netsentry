# NetSentry Performance Evidence and Portable-Budget Audit — 2026-08-05

## Audit Baseline

- Repository baseline: `323be1f38fca456a0d17a7801e18bc50c5212075`.
- Branch and remote: clean `main`; fetched `origin/main` matched the baseline.
- Audit period: 2026-07-14 through 2026-08-05.
- Delivery authority: benchmark and pressure source, direct tests, Git commits,
  task plans/states, fetched remote refs, and exact-range local Vault records.
- Performance boundary: local microbenchmarks and synthetic or reviewed-corpus
  pressure results are evidence for their exact fixtures and environments.
  They are not production capacity or cross-host regression thresholds.
- Release boundary: R90-59 remains blocked; this audit authorizes no tag,
  release, registry, image, workflow, benchmark run, or runtime mutation.

## Method

1. Reconcile each R90-70/R90-71 feature and docs-only closure with its parent,
   intended paths, completed plan/state, fetched remote chain, and exact Vault
   range.
2. Inspect each named C and Go benchmark body and the production call boundary
   it measures; do not infer broader pipeline coverage from a nearby harness.
3. Compare the root/module Make commands with the repeat-pcap,
   corpus-pressure, health, and Prometheus execution paths.
4. Inventory checked-in and local-only measurements, their environment and
   commit binding, repetition count, raw output, aggregation, and evidence
   classification.
5. Review recent commits by delivery phase and record material trends,
   validation deviations, stale claims, missing records, and unresolved risks.
6. Parse every task-state JSON file, require one Definition per roadmap row,
   and give every unfinished item complete planning fields.
7. Define only the smallest dependency-ordered evidence queue; keep any
   portable threshold decision blocked until comparable measurements and
   explicit product/SLO scope exist.

## Delivery History Review

| Period | Commits | Main delivery theme |
| --- | ---: | --- |
| Jul 14–20 | 47 | roadmap/remote governance, release evidence, UDS lifecycle, corrupt SQLite and recovery startup preservation |
| Jul 21–27 | 56 | recovery contracts and atomicity, SQLite compatibility, persisted-row validation |
| Jul 28–Aug 5 | 50 | exact recovery encoding, delivery audits, sidecar/restart-free recovery, fuzz/fault evidence, and C/Go benchmark completion |
| **Total** | **153** | one dependency-ordered governance, correctness, recovery, fuzz/fault, and benchmark sequence |

The period contains 66 feature-like commits, 68 `docs: record R90-*` closure
records, and 19 other documentation commits. Counts locate phase trends only;
completion remains tied to direct behavior, fetched refs, and exact Vault
evidence.

Two benchmark-phase validation deviations are material and resolved. R90-70
reproduced an unrelated storage cancellation timeout because the all-package
Go benchmark command also ran ordinary tests; `make bench` now uses
`-run '^$'`, while `make test` retains the complete correctness/race gate.
R90-71 reproduced the active-cancellation regression because its SQLite busy
timeout equaled the outer return deadline; the test-only driver timeout is now
comfortably shorter and passed 20 uncached race runs plus clean package/full
suites. Neither correction changes production behavior or supplies numeric
performance evidence.

The R90-70 feature-range Vault helper first encountered an absent documented
default path, then synchronized the unchanged exact range after the sole local
Vault was supplied explicitly. All later feature and closure ranges are
present. Current delivery must continue explicit single-Vault discovery rather
than silently creating another Vault.

## R90-70 and R90-71 Delivery Reconciliation

| Increment | Git and state | Direct benchmark boundary | Vault and current authority |
| --- | --- | --- | --- |
| R90-70 rule matching | Feature `388487da7205e98dd257ee54a1428673141c7457`; closure `e853f8e22d10c98cc9363356272c6d847421514b`; completed plan/state | `BenchmarkMatcherMatch` measures fixed no-hit/multi-hit Aho-Corasick `Match`; `BenchmarkEngineMatch` measures fixed no-hit/multi-hit full rule `Match` across payload, IP, and port rules. Construction, reload, Base64 setup, correctness checks, and diagnostics are outside timing; bytes and allocations are reported. | Both exact chained notes, full-index rows, and MOC links exist; stable Aho-Corasick/rule/Makefile/testing notes record the delivered boundary. |
| R90-71 alert store | Feature `9f29bf32cc3bbc446d03bd2185900c3dae4a84ef`; closure `323be1f38fca456a0d17a7801e18bc50c5212075`; completed plan/state | `BenchmarkStoreWriteBatch` measures production-durable single/32-alert `WriteBatch` calls with unique identity and outside-timing state cleanup; `BenchmarkStoreQuery` measures indexed exact-rule and timestamp-range queries over one 512-row production-seeded fixture. Cardinality, recovery-log, health, index, and result checks remain outside timing; allocations and alerts/op are reported where applicable. | Both exact chained notes, full-index rows, and MOC links exist; stable SQLite/Makefile/testing notes record the delivered boundary. |

The four commits form the expected uninterrupted feature/closure chain from
the R90-69 closure through the fetched baseline. Their intended paths match
their plans, and no delivery or Vault record is missing.

## Measurement Surface Inventory

| Surface | Timed boundary and fixture | Current output/evidence | Supported claim and remaining gap |
| --- | --- | --- | --- |
| C parser | `clock_gettime(CLOCK_MONOTONIC)` around repeated parsing of fixed Ethernet, VLAN, and Q-in-Q IPv4/TCP frames | iterations, ns/packet, and pps; one historical 2026-06-29 table | Stable local parser microbenchmark. No versioned raw sample set, commit binding, repetition distribution, or current complete-surface environment envelope. |
| C formatter | Repeated packet and heartbeat JSON formatting with fixed bounded structs | iterations, ns/op, ops/sec, serializer statistic | Stable local formatter microbenchmark. It excludes receiver decode, queueing, matching, and storage. |
| C UDS line write | Repeated writes on one preconnected Unix socket to a local drain child | iterations, ns/op, ops/sec, write errors | Local socket-write boundary only. It is not end-to-end IPC or engine throughput. |
| Go Aho-Corasick | `testing.B` no-hit/multi-hit `Matcher.Match` on fixed payloads | ns/op, bytes/sec, B/op, allocs/op during direct/root runs | Stable matcher hot-path evidence. Numeric output is not checked in as a versioned baseline. |
| Go rule engine | `testing.B` no-hit/multi-hit `Engine.Match` on one immutable seven-rule mixed fixture | ns/op, bytes/sec, B/op, allocs/op during direct/root runs | Stable full-match fixture evidence. It does not model production rule-set size or traffic mix. |
| Go durable writes | `testing.B` around only `WriteBatch` for one or 32 unique alerts with recovery append/sync, SQLite commit, and synced clear | ns/op, B/op, allocs/op, alerts/op during direct/root runs | Stable primary-write lifecycle evidence. It excludes fixture creation, validation checks, and cleanup by design and is not a sustained store-capacity result. |
| Go indexed queries | `testing.B` repeated exact-rule and time-range `Query` calls over one fixed 512-row fixture | ns/op, B/op, allocs/op during direct/root runs | Stable indexed-query evidence for two plans/cardinalities. It is not a data-size or concurrent-query curve. |
| Repeat-pcap pressure | Capture invocation through UDS, receiver, four workers, rule matching, SQLite aggregation, drain wait, health and API assertions over a repeated six-packet synthetic pattern | terminal elapsed, pps, alert/s; three dated samples in `docs/performance.md` and one local-only structured record | Functional full-pipeline pressure smoke for one synthetic mix. The 552–1,402 pps historical spread demonstrates environment sensitivity and cannot support a 10% portable gate. |
| Corpus pressure | Per-file and aggregate capture-to-drain timing for operator-supplied sanitized pcap/pcapng, plus errors, sampled RSS, health, metrics, and query snapshot | path-redacted local JSON/Markdown; one approved public-traffic audit result | Evidence for the exact corpus, rules, commit, and host only. Corpus identity/provenance and comparable environments remain external inputs. |
| Runtime metrics | Process-lifetime packet/alert rate gauges, rule/write latency histograms, current/high-water queue depth, errors, storage and capture state | `/api/metrics` and verbose health; correctness tests assert names/values | Operational observation surface. Lifetime averages and fixed buckets are not benchmark samples or regression budgets by themselves. |

## Comparability and Claim Audit

- The historical C table is from 2026-06-29 with Go 1.22.2 and GCC 13.3.0;
  the completed Go benchmark plans validated with the module-selected Go
  1.25.12 on 2026-08-04. They are not one matched complete-surface run.
- R90-70 and R90-71 prove discovery, execution, fixture correctness, and clean
  validation, but intentionally do not version their numeric Go output.
- No checked-in artifact binds all named C/Go cases to one exact clean commit,
  environment fingerprint, command parameters, raw output, and repeated sample
  set. The ignored local evidence directory cannot serve as remote delivery
  authority.
- The three historical repeat-pcap samples use the same nominal fixture but
  vary materially. Without recorded machine load, exact commit, repeated raw
  samples, and a controlled comparison policy, the old `>10%` pps and `>20%`
  RSS roadmap figures in `AUDIT_REPORT.md` are aspirations, not active gates.
- Microbenchmark rates, full-pipeline synthetic pressure, reviewed-corpus
  pressure, and process-lifetime metrics measure different boundaries. They
  must remain separate dimensions in any future evidence schema.
- One local environment can support a reproducible same-host observation and
  variance baseline. It cannot alone establish a portable budget or production
  SLO.

## Forward Queue

| ID | Dependency | Forecast | Status | Risk | Bounded outcome |
| --- | --- | --- | --- | --- | --- |
| R90-73 | R90-72 | Aug 5–Sep 4 | Planned | Medium | Add a versioned, directly tested evidence command that captures every established C/Go benchmark with exact commit/tree state, environment/toolchain fingerprint, command parameters, raw output, parsed metrics, path redaction, and explicit local-synthetic classification, without thresholds. |
| R90-74 | R90-73 | Sep 5–Oct 2 | Planned | Medium | Record at least five uncached complete-surface samples from one clean pinned commit and unchanged local environment, preserving each raw sample plus median/IQR/variation summaries as a same-host observation baseline with no pass/fail budget. |
| R90-75 | R90-74; comparable-environment evidence; explicit budget scope | Oct 3–31 | Blocked | High | Decide whether any regression budget can be portable, same-host-only, or observation-only from matched evidence and product/SLO authority; do not activate a numeric gate from the current single-host data. |

Every roadmap Definition names required validation and a stop condition.
R90-59 remains blocked in parallel on exact version/SHA publication authority.
External corpora remain optional diagnostics, not a performance-budget or
release prerequisite. No R90-73 implementation starts in this audit.

## Audit Conclusion

R90-70 and R90-71 are complete and recoverable from exact Git, task state,
benchmark source, fetched remote, and Vault evidence. The current harnesses
cover useful, deliberately separate C parser/formatter/socket, Go matcher/rule/
store, full-pipeline pressure, and runtime-observation boundaries.

The evidence is not yet comparable enough for a portable budget: current Go
numbers are not versioned, the complete surface has no repeated matched-host
sample set, historical pressure results vary materially, and corpus evidence
is fixture/environment-specific. R90-72 therefore replaces the unsupported
immediate-threshold proposal with a bounded evidence-capture increment, a
single-host repeated baseline, and a separately blocked budget decision.
Publication remains blocked and no later increment was started.
