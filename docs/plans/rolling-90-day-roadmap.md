# NetSentry Rolling 90-Day Roadmap

> Window: 2026-08-10 through 2026-11-08. This is the active delivery queue for `$netsentry-next`; refresh unfinished work at each completed increment using Git, task-state, and evidence as authority. Completed history from the prior horizon is preserved below.

## Status Rules

- **Ready**: every dependency is complete. As of 2026-07-16, roadmap dates are
  forecasting metadata only and never prevent work from starting.
- **Planned**: at least one internal dependency is unfinished; no external
  authority or input currently blocks the item.
- **Blocked**: requires an explicitly recorded external input, authority, or unresolved validation result.
- **Complete**: acceptance criteria and required evidence are verified, including commit/push/Vault evidence when a repository increment was delivered.
- Complete only one ready increment per `$netsentry-next` trigger. Record deviations before reordering unfinished work.

## Global Eligibility Policy

- **Schedule authority:** The Jul 16, 2026 global schedule-window waiver remains
  active. Dates are forecasts only and never gate selection or completion.
- **Prerequisite authority:** On Aug 1, 2026, the user cancelled additional
  prerequisite review as an eligibility gate. When internal dependencies are
  complete, `$netsentry-next` may select the safest bounded default and record
  it in the increment plan instead of waiting for a separate product review.
- **Unchanged boundaries:** Required validation and acceptance evidence remain
  mandatory. This policy does not authorize private-data access, destructive
  recovery, automatic evidence deletion, version tags, GitHub Releases, image
  publication, workflow dispatch, or any other external mutation that needs
  explicit action authority.

## Per-Trigger Plan Audit

1. **Baseline audit:** work from the repository root; fetch the active remote;
   isolate pre-existing user changes; require local and fetched refs to agree;
   verify the latest completed plan/state and exact Vault note, index, and MOC.
2. **History audit:** review the previous two to four weeks at phase level and
   record material deviations, missing delivery records, stale authority, and
   unresolved risks rather than treating commit volume as completion evidence.
3. **Forward-plan audit:** require every unfinished item to have status,
   dependency, window, risk, acceptance criteria, required validation, and stop
   condition. Reconcile an empty, stale, contradictory, or incomplete queue
   before implementation.
4. **Selection audit:** choose exactly one highest-priority dependency-ready
   increment and persist its plan/state, non-goals, evidence map, and authority
   boundaries before editing.
5. **Execution audit:** compare progress to the plan at meaningful checkpoints;
   record every validation deviation and do not start unrelated work while a
   result is ambiguous.
6. **Pre-commit audit:** require acceptance evidence, applicable focused and
   repository checks, `make knowledge-check`, JSON/diff validation, intended
   staged paths only, and a sensitive-information review.
7. **Delivery audit:** record full old/new SHAs; push without force; fetch and
   require `HEAD == origin/main == new`; rerun the knowledge gate; synchronize
   and verify the exact Vault range.
8. **Closeout audit:** persist verified delivery facts in one docs-only closure
   commit when needed, deliver and synchronize that second exact range as part
   of the same increment, refresh but do not start the next item, and leave
   accurate resume instructions.

## Phased Delivery Queue

| ID | Window | Status | Increment | Dependencies | Acceptance criteria |
|---|---|---|---|---|---|
| R90-01 | Jul 14–24 | Complete | Rebuild rolling roadmap capability and initial plan. | None | `netsentry-roadmap` is discoverable; `$netsentry-next` loads this roadmap and ends after one eligible increment; roadmap records windows, dependencies, validation, and acceptance criteria. |
| R90-02 | Jul 14 | Complete early | Add Git lifecycle decision policy and task-state reconciliation. | R90-01 | Every repository change must pass local `make knowledge-check` before commit; a failure blocks delivery until its roadmap/state/evidence cause is reconciled and the check is rerun successfully. |
| R90-03 | Jul 14 | Complete early | Add remote-baseline roadmap self-check and deviation-reporting workflow. | R90-02 | After every push, fetch `origin/main`, require active state to match its SHA, then run `make knowledge-check`; any ref or validation drift blocks delivery until reconciled. |
| R90-03a | Jul 14 | Complete early | Decouple post-push sync tests from local Git hooks. | R90-03 | Versioned Python sync APIs are tested directly; `.git/hooks/post-push` is only a thin local wrapper; `make knowledge-check` passes without hook files. |
| R90-04a | Jul 15 | Complete | Revalidate the v0.1.1 code-quality baseline independently of production-traffic evidence. | R90-03a | Passed non-Docker RC and pinned supply-chain baselines are recorded; no release-ready or production-evidence claim is made. |
| R90-04 | Jul 15–Sep 11 | Complete | Review approved anonymized public real-traffic PCAP evidence, then run corpus-pressure validation. | R90-04 scoped exception | Path-redacted MAWI real-traffic evidence passed dedicated privacy, provenance, sanitization, sensitive-metadata, and corpus-pressure review; the exception expires for this increment. |
| R90-04b | Jul 16 | Complete | Enforce the completed R90-04 exception boundary before R90-05. | R90-04 | The audit record is expired and the release gate directly rejects R90-04-backed release approval while preserving the historical v0.1.0 gate. |
| R90-05 | Jul 16 | Complete early | Prepare v0.1.1 release readiness from validated evidence. | R90-04; passing code quality gates | `make rc-check`, supply-chain, and release gates pass; public docs/evidence identify no unresolved release blocker. |
| R90-06 | Window waived; forecast was Oct 3–14 | Complete under waiver | Assemble a release decision package. | R90-05 | Version, commit, evidence, checksums, and intended publication decision are reconciled; do not tag or publish without explicit user authorization. |
| R90-07 | Jul 17–24 | Complete early | Bound concurrent Go UDS receiver connections. | R90-06 | A validated finite connection limit rejects excess clients, releases capacity after disconnect, and preserves reconnect/shutdown behavior. |
| R90-08 | Jul 17–31 | Complete early | Add an active-load full-engine shutdown drill. | R90-07 | One integration test exercises receiver, worker, HTTP, and SQLite teardown with in-flight work and proves bounded clean shutdown without writes after store close. |
| R90-09 | Jul 18–Aug 7 | Complete early | Fail closed on corrupt SQLite startup state. | R90-08 | A deterministic regression proves corrupt or truncated SQLite input causes a clear startup error without overwriting the database, and recovery guidance preserves operator data. |
| R90-10 | Jul 18–Aug 14 | Complete early | Preserve corrupt historical daily shards on write. | R90-09 | Opening an existing non-current daily shard for a write uses the same read-only integrity preflight; corrupt/truncated shards reject the write and remain byte-for-byte unchanged. |
| R90-11 | Jul 18–Aug 21 | Complete early | Make historical daily-shard reads strictly read-only. | R90-10 | Query and count open non-current shards with a read-only SQLite handle; corrupt/truncated inputs fail without changing shard bytes, while healthy cross-shard results remain unchanged. |
| R90-12 | Jul 18–Aug 28 | Complete early | Preserve malformed recovery logs during startup replay. | R90-11 | Corrupt and truncated JSONL recovery logs fail startup with a clear error and remain byte-for-byte unchanged; valid logs still replay and truncate only after successful persistence. |
| R90-13 | Jul 19–Sep 4 | Complete early | Bound idle UDS receiver connections. | R90-12 | A validated finite per-connection read timeout applies before the first frame and refreshes after every complete frame; idle expiry releases handler capacity without inflating decode errors, while active traffic and shutdown remain compatible. |
| R90-14 | Jul 19–Sep 11 | Complete early | Enforce the per-connection UDS hello/session state machine. | R90-13 | Each connection requires exactly one valid hello before heartbeat or packet frames; heartbeat session IDs must match that hello; state violations close only the offending connection, increment decode errors once, and preserve valid reconnect/shutdown behavior. |
| R90-15 | Jul 20–Sep 18 | Complete early | Reject incompatible existing SQLite schemas before writable initialization. | R90-14 | A structurally valid but non-NetSentry or incompatible existing database fails startup clearly and remains byte-for-byte unchanged; compatible existing, empty, and missing databases retain current behavior. |
| R90-16 | Jul 20–Sep 25 | Complete early | Reject semantically invalid recovery-log records before replay. | R90-15 | Newline-terminated, syntactically valid JSON records that cannot satisfy the durable normalized-alert contract fail startup clearly; the complete recovery log remains unchanged, no valid prefix is persisted, and valid replay behavior is preserved. |
| R90-17 | Jul 20–Oct 2 | Complete early | Preflight recovery logs before writable SQLite initialization. | R90-16 | Invalid recovery input fails before a missing database can be created or a compatible existing database can be modified; valid replay and initialization behavior remain unchanged. |
| R90-18 | Jul 21–Oct 9 | Complete early | Reject inconsistent normalized recovery records before replay. | R90-17 | Recovery records whose durable ID, first/last timestamps, window start, or aggregate count cannot be emitted by the normalized writer fail before SQLite initialization; the complete log and target database remain unchanged, while valid replay behavior is preserved. |
| R90-19 | Jul 22–Oct 15 | Complete early | Preflight recovery logs before runtime append. | R90-18 | A runtime write rejects an already malformed or semantically invalid recovery log before appending or touching SQLite; the complete log and database remain unchanged, while valid pending-log persistence remains compatible. |
| R90-20 | Jul 22–Oct 20 | Complete early | Bound recovery-record encoding and replay. | R90-19 | Valid writer-generated records above the scanner's former 64 KiB ceiling persist and replay; records above the explicit 4 MiB durable limit fail before append, leaving the recovery log and database unchanged. |
| R90-21 | Jul 22–Oct 20 | Complete early | Reject write-blocking SQLite schema extensions. | R90-20 | Existing primary and historical databases with unknown `NOT NULL` columns lacking a usable non-NULL default fail read-only preflight and remain unchanged; nullable and non-NULL-defaulted extra columns remain compatible. |
| R90-22 | Jul 23–Oct 20 | Complete early | Reject write-blocking SQLite uniqueness extensions. | R90-21 | Existing primary and historical databases with extra unique indexes that do not contain a binary-collated canonical write identity fail read-only preflight and remain unchanged; non-unique indexes and uniqueness extensions containing an existing safe identity remain compatible and writable. |
| R90-23 | Jul 23–Oct 20 | Complete early | Reject write-affecting SQLite triggers. | R90-22 | Existing primary and historical databases with triggers attached to `alerts` or `alert_events` fail read-only preflight and remain unchanged; triggers confined to unrelated operator tables remain compatible and NetSentry writes succeed. |
| R90-24 | Jul 23–Oct 20 | Complete early | Reject write-affecting SQLite generated columns. | R90-23 | Existing primary and historical databases with virtual or stored generated columns on `alerts` or `alert_events` fail read-only preflight and remain unchanged; ordinary nullable and defaulted column extensions remain compatible and writable. |
| R90-25 | Jul 24–Oct 20 | Complete early | Reject write-affecting SQLite check constraints. | R90-24 | Existing primary and historical databases with `CHECK` constraints on `alerts` or `alert_events` fail read-only preflight and remain unchanged; constraints confined to unrelated operator tables remain compatible and NetSentry writes succeed. |
| R90-26 | Jul 24–Oct 20 | Complete early | Reject write-affecting SQLite foreign keys. | R90-25 | Existing primary and historical databases with foreign-key relationships whose source or target is `alerts` or `alert_events` fail read-only preflight and remain unchanged; relationships confined to unrelated operator tables remain compatible and NetSentry writes succeed. |
| R90-27 | Jul 25–Oct 20 | Complete early | Require binary collation on SQLite aggregation uniqueness. | R90-26 | Existing primary and historical databases whose canonical alert aggregation uniqueness uses a non-binary collation fail read-only preflight and remain unchanged; a binary-collated canonical key remains compatible and preserves distinct NetSentry identities. |
| R90-28 | Jul 25–Oct 20 | Complete early | Honor SQLite identifier case semantics during schema preflight. | R90-27 | Existing primary and historical databases with case-variant required table, column, aggregation-key, and safe unique-key identifiers pass the same validation and remain writable; all write-safety checks remain enforced. |
| R90-29 | Jul 25–Oct 20 | Complete early | Pin SQLite exact-filter collation. | R90-28 | Rule, severity, source, and destination filters retain binary exact-match semantics for compatible primary and historical schemas regardless of declared column collation; intentionally case-insensitive filters remain unchanged. |
| R90-30 | Jul 25–Oct 20 | Complete early | Validate stored SQLite alert numerics. | R90-29 | Primary and historical reads reject destination ports outside `0..65535` and aggregate counts below one without silently narrowing values or modifying historical shard bytes; valid rows remain compatible. |
| R90-31 | Jul 25–Oct 20 | Complete early | Validate stored SQLite alert severity. | R90-30 | Primary and historical reads accept only the four public severity values; empty, case-variant, and unsupported values fail without substitution or historical shard modification. |
| R90-32 | Jul 25–Oct 20 | Complete early | Validate stored SQLite timestamp ordering. | R90-31 | Primary and historical reads reject rows with `first_seen > last_seen` or `window_start > first_seen` without modifying historical shard bytes; valid historical aggregation windows remain compatible. |
| R90-33 | Jul 25–Oct 20 | Complete early | Validate stored SQLite aggregation identity. | R90-32 | Primary and historical reads reject empty or altered alert IDs that disagree with the canonical aggregation tuple without modifying historical shard bytes; valid aggregation identities remain compatible. |
| R90-34 | Jul 25–Oct 20 | Complete early | Validate stored SQLite required text. | R90-33 | Primary and historical reads reject blank required identity, rule, and network text without modifying historical shard bytes; optional empty text and valid rows remain compatible. |
| R90-35 | Jul 25–Oct 20 | Complete early | Enforce the UDS IPv4 address contract. | R90-34 | UDS packet frames accept only strict IPv4 source and destination addresses; ordinary and IPv4-mapped IPv6 text fails once without queueing a packet, while valid capture traffic remains compatible. |
| R90-36 | Jul 25–Oct 20 | Complete early | Enforce the recovery IPv4 address contract. | R90-35 | Startup and runtime recovery preflight reject malformed, ordinary IPv6, or IPv4-mapped IPv6 source/destination addresses before modifying the complete log or missing/existing SQLite state; valid IPv4 replay remains compatible. |
| R90-37 | Jul 25–Oct 20 | Complete early | Validate stored SQLite IPv4 addresses. | R90-36 | Primary and historical reads reject malformed, ordinary IPv6, and IPv4-mapped IPv6 source/destination text before identity derivation; valid IPv4 rows remain compatible and historical rejection preserves shard bytes. |
| R90-38 | Jul 25–Oct 20 | Complete early | Validate recovery event identity. | R90-37 | Startup and runtime recovery preflight reject nonblank `event_id` values that differ from the deterministic event identity before modifying the complete log or missing/existing SQLite state; valid idempotent replay remains compatible. |
| R90-39 | Jul 25–Oct 20 | Complete early | Validate recovery severity. | R90-38 | Startup and runtime recovery preflight accept only `low`, `medium`, `high`, or `critical`; empty, case-variant, and unsupported severities fail before modifying the complete log or missing/existing SQLite state, while all four public values remain compatible. |
| R90-40 | Jul 25–Oct 20 | Complete early | Validate recovery rule names. | R90-39 | Startup and runtime recovery preflight reject missing, empty, or whitespace-only `rule_name` before modifying the complete log or missing/existing SQLite state, while nonblank names replay without normalization. |
| R90-41 | Jul 25–Oct 20 | Complete early | Validate stored SQLite MITRE tuples. | R90-40 | Primary and historical reads accept MITRE tactic/ID/name only when all three are empty or all three are nonblank; partial and whitespace-only tuple members fail without normalization or historical shard modification. |
| R90-42 | Jul 25–Oct 20 | Complete early | Validate recovery MITRE tuples. | R90-41 | Startup and runtime recovery preflight accept MITRE tactic/ID/name only when all three are empty or all three are nonblank; every partial and whitespace-only tuple fails before modifying the complete log or missing/existing SQLite state, while valid tuple text remains unchanged. |
| R90-43 | Jul 26–Oct 24 | Complete early | Validate stored SQLite protocol names. | R90-42 | Primary and historical reads accept exactly the canonical writer-emittable `TCP`, `UDP`, `ICMP`, and `PROTO_<0..255>` names; case variants, arbitrary names, malformed/out-of-range numeric forms, and numeric aliases of named protocols fail without historical shard modification. |
| R90-44 | Jul 27–Oct 25 | Complete early | Validate recovery protocol names. | R90-43 | Startup and runtime recovery preflight accept exactly the canonical writer-emittable `TCP`, `UDP`, `ICMP`, and `PROTO_<0..255>` names; every noncanonical form fails before modifying the complete log or missing/existing SQLite state. |
| R90-45 | Jul 27–Oct 25 | Complete early | Preflight the current recovery batch. | R90-44 | Every newly normalized alert passes the complete durable recovery contract before any current-batch append or SQLite write; a later invalid record cannot partially append a valid prefix, alter an existing pending log/database, or degrade healthy storage. |
| R90-46 | Jul 28–Oct 26 | Complete early | Validate stored SQLite timestamp encoding. | R90-45 | Primary and historical row reads accept aggregate timestamps only in the exact UTC RFC3339Nano text emitted by the writer; parseable offsets and nonminimal fractional forms fail without historical shard modification, while canonical rows remain compatible. |
| R90-47 | Jul 28–Oct 26 | Complete early | Pin SQLite timestamp comparison semantics. | R90-46 | Aggregation updates, alert ordering/pagination, time filters, and retention pruning compare canonical variable-width RFC3339Nano values by instant with nanosecond fidelity; mixed fractional widths remain chronological without rewriting stored rows. |
| R90-48 | Jul 29–Oct 27 | Complete early | Validate recovery timestamp encoding. | R90-47 | Startup and runtime recovery preflight accept `timestamp`, `first_seen`, `last_seen`, and `window_start` only as exact canonical UTC RFC3339Nano strings emitted by the writer; parseable offsets and nonminimal fractional forms fail before modifying the complete log or missing/existing SQLite state. |
| R90-49 | Jul 29–Oct 27 | Complete early | Reject duplicate recovery JSON fields. | R90-48 | Startup and runtime recovery preflight reject exact duplicate top-level JSON names and case-variant aliases targeting the same durable field before last-value decoding can obscure input; the complete log and missing/existing SQLite state remain unchanged while canonical writer records remain compatible. |
| R90-50 | Jul 29–Oct 27 | Complete early | Enforce canonical recovery JSON field names. | R90-49 | Startup and runtime recovery preflight reject a single unknown top-level name or noncanonical case alias before model decoding; every current writer field, including optional `raw_payload`, remains compatible and rejected input preserves the complete log plus missing/existing SQLite state. |
| R90-51 | Jul 30–Oct 28 | Complete early | Require complete recovery JSON records. | R90-50 | Startup and runtime recovery preflight require every non-`omitempty` field emitted by the current writer before model decoding; optional `raw_payload` remains compatible, diagnostic precedence is preserved, and rejected input leaves the complete log plus missing/existing SQLite state unchanged. |
| R90-52 | Jul 30–Oct 28 | Complete early | Enforce recovery JSON value types. | R90-51 | Startup and runtime recovery preflight require every present top-level value to use the non-null JSON kind emitted by the current writer; optional `raw_payload` remains optional but string-typed, diagnostic precedence is preserved, and rejected input leaves the complete log plus missing/existing SQLite state unchanged. |
| R90-53 | Jul 30–Aug 1 | Complete early | Audit recent delivery and future planning. | R90-52 | A dated audit reconciles recent commits, plans/states, remote/Vault evidence, and release boundaries; every roadmap item has a complete definition; future work spans the active horizon; each trigger has an explicit plan-audit and two-commit closeout sequence. |
| R90-54 | Jul 31–Aug 7 | Complete early | Enforce canonical recovery JSON numeric encoding. | R90-53 | Recovery `dst_port` and `aggregated_count` accept only the exact base-10 integer spelling emitted by the writer; alternate exponent, fractional, sign, and leading-zero forms fail before durable mutation while canonical replay remains compatible. |
| R90-55 | Aug 8–21 | Complete early | Eliminate recovery field-contract drift. | R90-54 | One authoritative model contract drives canonical field names, required/optional status, and JSON kinds; adding or changing a writer field cannot silently bypass reader validation. |
| R90-56 | Aug 22–Sep 18 | Complete early | Preserve corrupt SQLite sidecars during preflight. | R90-55 | Deterministic primary and historical fixtures with corrupt or inconsistent WAL/SHM state fail read-only preflight clearly and preserve the database plus sidecar bytes; healthy active-WAL reads remain compatible. |
| R90-57 | Forecast Sep 19–Oct 2; waived | Complete early | Define restart-free emergency recovery semantics. | R90-56 | An operator-triggered, fail-closed state machine defines probe, recovery, retry, concurrency, and evidence-preservation boundaries without duplicate writes or automatic cleanup; implementation remains a separate increment. |
| R90-58 | Oct 3–21 | Complete early | Refresh the v0.1.1 candidate decision package. | R90-56 | Version, current candidate commit, gates, artifacts, checksums, platform, and hold decision are reconciled from fresh evidence without tagging or publishing. |
| R90-59a | Aug 7 | Complete | Create the authorized local v0.1.1 tag without remote publication. | R90-58; exact tag-only authorization | A signed annotated local `v0.1.1` tag resolves exactly to the authorized candidate after candidate changelog/evidence review and smoke validation; the remote tag remains absent and no workflow, GitHub Release, or GHCR action occurs. |
| R90-59 | Oct 22–28 | Blocked on remote-publication authority | Execute the remote v0.1.1 publication gate. | R90-59a; explicit tag-push, GitHub Release, and GHCR authorization | Only an explicitly authorized tag may be pushed; GitHub Release and GHCR results must be verified directly, while absence of the remaining authority preserves the local-only tag without external mutation. |
| R90-60 | Forecast Aug 1–Oct 30; waived | Complete early | Implement operator-triggered restart-free storage recovery. | R90-57 | One authenticated request serializes recovery against store lifecycle operations, preflights durable input before the writable boundary, replays or probes idempotently, exposes bounded health/audit outcomes, and leaves failures in sticky emergency without automatic cleanup or retry. |
| R90-61 | Aug 2 | Complete | Audit post-recovery delivery and restore the forward queue. | R90-60 | A dated audit reconciles recent commits, plans/states, fetched remote and Vault evidence, records the committed-prefix test gap, and restores a complete evidence-grounded queue without runtime or publication changes. |
| R90-62 | Aug 3–Sep 4 | Complete early | Prove committed-prefix multi-shard recovery retry. | R90-61 | Deterministic direct regressions cancel or fail recovery after an earlier shard commit, retain the complete log and emergency state, and prove explicit retry completes every event once without aggregate inflation. |
| R90-63 | Sep 5–Oct 9 | Complete early | Add a dedicated C UDS JSON formatter fuzz boundary. | R90-62 | An ASan-capable deterministic harness covers packet, heartbeat, and hello formatting across escaping, payload, integer, and output-boundary inputs; valid output remains canonical JSONL and failures never overrun or expose partial buffers as success. |
| R90-64 | Oct 10–31 | Complete early | Record a sustained parser and formatter fuzz baseline. | R90-63 | Reproducible sustained ASan runs exercise both C harnesses with path-redacted corpus metadata, no crashes or sanitizer findings, and an honest local/synthetic evidence classification without a release or production-traffic claim. |
| R90-65 | Aug 3–14 | Complete early | Audit the completed fuzz delivery and scope the next local hardening queue. | R90-64 | A dated audit reconciles the dual-harness evidence, public remaining-gap claims, code/tests, fetched remote, and exact Vault records; external-input gaps are separated from bounded local work, and every added increment has a complete dependency, window, risk, acceptance, validation, and stop definition. |
| R90-66 | Aug 15–Sep 4 | Complete early | Prove primary write interruption recovery. | R90-65 | Real SQLite contention and active cancellation after durable recovery append but before primary commit leave no partial database mutation, retain the complete log, and permit one explicit retry to persist each event once without aggregate inflation. |
| R90-67 | Sep 5–Oct 2 | Complete early | Inject recovery-log append lifecycle faults. | R90-66 | Direct open, short-write, sync, and close failures occur before SQLite mutation, retain the exact pre-existing valid log prefix, expose the failing phase, and leave complete or incomplete appended evidence fail-closed without automatic deletion. |
| R90-68 | Oct 3–31 | Complete early | Harden post-commit recovery-log clearing. | R90-67 | Direct open/truncate, sync, and close failures after a primary or daily-shard commit cannot lose an alert or inflate an aggregate; every retained-log or already-cleared outcome remains explicit and one operator retry returns healthy. |
| R90-69 | Aug 4–14 | Complete early | Audit the completed local storage-fault sequence and restore the forward queue. | R90-68 | A dated audit reconciles R90-66 through R90-68 code, direct tests, task states, fetched remote, exact Vault evidence, and current public gap claims; it restores a complete evidence-grounded queue without runtime or publication changes. |
| R90-70 | Aug 4–Sep 4 | Complete early | Add Go rule-matching microbenchmarks. | R90-69 | `make bench` executes deterministic Aho-Corasick and full rule-engine cases for no-hit and multi-hit payloads; setup and correctness checks remain outside timed regions, allocations are reported, and no host-independent or production threshold is claimed. |
| R90-71 | Sep 5–Oct 2 | Complete early | Add Go alert-store microbenchmarks. | R90-70 | `make bench` executes bounded primary SQLite write and filtered-query cases with unique event identity, production recovery durability intact, deterministic cardinality checks outside timed regions, and no operator data or production throughput claim. |
| R90-72 | Oct 3–31 | Complete early | Audit local performance evidence and scope a portable budget. | R90-71 | A dated audit reconciles the complete C/Go benchmark surface, local pressure tooling, public performance claims, and exact delivery/Vault evidence, then defines only a supportable baseline or budget queue without inventing cross-host or production thresholds. |
| R90-73 | Aug 5–Sep 4 | Complete early | Add versioned local benchmark evidence capture. | R90-72 | One directly tested command captures every established C/Go benchmark with exact Git/tree state, environment/toolchain fingerprint, parameters, raw output, parsed metrics, path redaction, and local-synthetic classification without applying a threshold. |
| R90-74 | Sep 5–Oct 2 | Complete early | Record a repeated single-host benchmark baseline. | R90-73 | At least five uncached complete-surface samples from one clean pinned commit and unchanged environment retain every raw result plus median/IQR/variation summaries as observation-only local evidence. |
| R90-75 | Oct 3–31 | Blocked / pending evidence; non-blocking | Decide portable performance-budget scope. | R90-74; comparable-environment evidence; explicit budget scope | Matched evidence and product/SLO authority decide whether a budget can be portable, same-host-only, or observation-only; current single-host data cannot activate a numeric gate or prevent unrelated dependency-ready roadmap work. |
| R90-76 | Aug 9 | Complete | Audit post-tag delivery and restore the forward queue. | R90-59a; R90-74 | A dated audit reconciles the local-tag feature/closure, recent delivery phases, fetched remote, exact Vault evidence, current code/tests, and blocked authorities, then restores only evidence-grounded local work without runtime or publication changes. |
| R90-77 | Aug 10–Sep 4 | Complete early | Serialize rule-management transactions. | R90-76 | Concurrent rule create/update/delete/reload operations cannot lose a successful mutation or leave canonical disk and active memory disagreeing; direct synchronized race regressions reach each promised interleaving. |
| R90-78 | Sep 5–25 | Complete early | Harden rule-file replacement durability. | R90-77 | Rule seed replacement explicitly handles short write, file sync, close, rename, and parent-directory sync with preservation-safe pre-rename failures and a defined post-rename memory/disk outcome. |
| R90-79 | Sep 26–Oct 16 | Complete early | Harden suppression-file replacement durability. | R90-78 | Suppression replacement directly proves the same lifecycle boundaries while retaining serialized mutation, exact prior-file preservation before rename, and active-filter agreement with every reported outcome. |
| R90-80 | Oct 17–31 | Complete early | Audit management-plane persistence and future compatibility scope. | R90-79 | A dated audit reconciles the rule/suppression transaction and durability sequence, current public claims, Git/task-state/remote/Vault evidence, and classifies remaining migration or product work without silently selecting a compatibility policy. |
| R90-81 | Aug 9 | Complete | Audit post-management-plane delivery and restore the local reliability queue. | R90-80 | A dated audit reconciles the R90-80 feature/closure, recent delivery phases, fetched remote, exact Vault evidence, current tests, and recurring validation deviations, then restores only evidence-grounded local reliability work without runtime or publication changes. |
| R90-82 | Aug 10–Sep 4 | Complete early | Stabilize receiver idle-capacity release evidence. | R90-81 | Direct receiver tests synchronize on an observable handler-capacity boundary rather than a shared heartbeat/session poll, prove timeout-driven slot release and replacement acceptance, and pass repeated uncached race execution without changing production timeout semantics. |
| R90-83 | Aug 9 | Complete | Audit post-receiver delivery and restore the local filesystem-lifecycle queue. | R90-82 | A dated audit reconciles the R90-82 feature/closure, recent phases, fetched remote, exact Vault evidence, and the current UDS pathname lifecycle, then restores only a directly evidenced preservation increment without runtime or publication changes. |
| R90-84 | Aug 10–Sep 5 | Complete early | Preserve non-socket UDS pathname occupants. | R90-83 | Receiver startup rejects a pre-existing non-socket or symlink pathname without modifying it, and shutdown removes only the socket identity created by that receiver while preserving a replacement path; stale/active socket policy remains unchanged. |
| R90-85 | Aug 9 | Complete | Audit post-pathname delivery and repair roadmap chronology. | R90-84 | A dated audit reconciles the R90-84 feature/closure, recent phases, fetched remote, exact Vault evidence, corrects mutable delivery-history ordering, and restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-86 | Aug 10–Sep 5 | Complete early | Reject receiver startup with an already-canceled context. | R90-85 | `Start` returns an error matching `context.Canceled` before pathname mutation or listener creation; direct absent-path and pre-existing Unix-socket preservation regressions pass while live startup and post-readiness cancellation remain compatible. |
| R90-87 | Aug 10 | Complete | Audit post-cancellation delivery and restore the active-socket lifecycle queue. | R90-86 | A dated audit reconciles the R90-86 feature/closure, recent phases, fetched remote, exact Vault evidence, and current pre-existing-socket behavior, then restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-88 | Aug 11–Sep 12 | Complete early | Preserve an active UDS listener during receiver startup. | R90-87 | Startup rejects a currently connectable existing Unix listener without replacing its pathname identity or breaking its service, still reclaims a stale socket, and preserves a replacement identity if the pathname changes during classification. |
| R90-89 | Aug 11 | Complete | Audit post-listener delivery and restore the cancellation-aware startup queue. | R90-88 | A dated audit reconciles the R90-88 feature/closure, recent phases, fetched remote, exact Vault evidence, and current pre-readiness probe cancellation behavior, then restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-90 | Aug 12–Sep 12 | Complete early | Make the existing-socket probe context-aware. | R90-89 | Cancellation during a blocked pre-readiness Unix-socket probe returns an error matching the context sentinel promptly, preserves the captured pathname identity, and installs no receiver listener while active/stale classification remains compatible. |
| R90-91 | Aug 12 | Complete | Audit post-probe delivery and restore the shutdown pathname-generation queue. | R90-90 | A dated audit reconciles the R90-90 feature/closure, recent phases, fetched remote, exact Vault evidence, and current owned-socket shutdown cleanup, then restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-92 | Aug 13–Sep 12 | Complete early | Preserve an immediate replacement Unix socket during receiver shutdown. | R90-91 | Shutdown removes its owned pathname only when non-following device, inode, and change-time identity still match; a direct immediate-inode-reuse regression preserves a replacement listener while ordinary owned cleanup and regular-file/symlink replacement behavior remain compatible. |
| R90-93 | Aug 12 | Complete | Audit post-generation delivery and restore the listener-creation ownership queue. | R90-92 | A dated audit reconciles the R90-92 feature/closure, recent phases, fetched remote, exact Vault evidence, and current post-listen mode/ownership boundary, then restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-94 | Aug 13–Sep 13 | Complete early | Bind UDS mode application and ownership capture to the created listener. | R90-93 | Startup applies the configured mode through the created listener identity and publishes ownership only if the non-following pathname still matches; direct regular-file, symlink-target, and replacement-listener races preserve replacement state and service while ordinary mode and shutdown cleanup remain compatible. |
| R90-95 | Aug 14 | Complete | Audit post-listener-ownership delivery and restore the pre-readiness cancellation queue. | R90-94 | A dated audit reconciles the R90-94 feature/closure, recent phases, fetched remote, exact Vault evidence, and current post-private-listener cancellation boundary, then restores at most one directly evidenced local follow-on without runtime or publication changes. |
| R90-96 | Aug 15–Sep 14 | Planned | Reject cancellation after private UDS listener creation. | R90-95 | Cancellation synchronized after private listener creation but before pathname publication returns the context sentinel, publishes no listener ownership, and leaves neither public nor private listener artifacts while existing startup/probe/shutdown behavior remains compatible. |

## R90-01 Definition

- **Goal:** establish one versioned 90-day delivery authority and a
  one-increment `$netsentry-next` workflow.
- **Risk:** a parallel queue or multi-increment trigger could make delivery and
  recovery evidence ambiguous.
- **Required validation:** skill discovery, roadmap structure, dependency and
  acceptance review, documentation checks, and knowledge synchronization.
- **Stop condition:** stop if selecting work requires a second authority,
  private input, or more than one increment.

## R90-02 Definition

- **Goal:** make the local knowledge gate and task-state reconciliation
  mandatory before repository delivery.
- **Risk:** committing against stale task/Vault evidence can make a later
  session repeat or misreport delivery.
- **Required validation:** direct passing/failing knowledge-gate behavior,
  roadmap/state reconciliation review, documentation checks, and exact diff
  inspection.
- **Stop condition:** stop while the knowledge gate fails or its evidence
  conflict is unresolved.

## R90-03 Definition

- **Goal:** make fetched `origin/main` the post-push delivery and planning
  baseline.
- **Risk:** local-only success can conceal a failed push or remote drift.
- **Required validation:** push/fetch ref comparison, post-fetch knowledge
  validation, task-state reconciliation, and Vault range verification.
- **Stop condition:** stop if fetched refs differ, fetch/push status is
  ambiguous, or remote authority changes.

## R90-03a Definition

- **Goal:** keep knowledge-sync business logic versioned and tests independent
  of local hook files.
- **Risk:** CI or a clean checkout cannot reproduce behavior implemented only
  under `.git/hooks`.
- **Required validation:** direct versioned Python API tests, hook-free
  repository knowledge checks, idempotency fixtures, and documentation checks.
- **Stop condition:** stop if tests require provisioning or executing local
  hooks in CI.

## R90-04 Definition

- **Goal:** validate the authorized public real-traffic corpus under the
  R90-04-only evidence exception.
- **Risk:** public traffic can still expose sensitive metadata or be
  misrepresented as production-derived release approval.
- **Required validation:** privacy, provenance, sanitization,
  sensitive-metadata, integrity, and corpus-pressure review under the exact
  scoped exception.
- **Stop condition:** stop on unapproved traffic, private paths, failed review,
  digest drift, or an attempt to reuse the exception outside R90-04.

## R90-04a Definition

- **Goal:** record a current v0.1.1 code-quality baseline independently of any
  production-derived or public real-traffic evidence.
- **Risk:** a passing non-Docker quality baseline can be misrepresented as
  traffic evidence, release readiness, or publication approval.
- **Required validation:** repository-pinned supply-chain checks; non-Docker RC
  quality, race, coverage, fuzz, E2E, and archive checks; documentation,
  knowledge, and diff checks; explicit evidence and publication boundary review.
- **Stop condition:** stop if completion requires traffic acquisition or review,
  corpus-pressure evidence, private input, release approval, tagging, or
  publication; do not treat R90-04a as satisfying R90-04, R90-05, or R90-06.
- **Selected plan:**
  [`task-20260715-090000-r90-04a.md`](task-20260715-090000-r90-04a.md),
  from recorded remote baseline
  `b3d143ba6ee714f5518f32684fd96b9ea0925a0a`.

## R90-04b Definition

- **Goal:** expire the R90-04 exception and prevent its historical evidence
  from authorizing later release decisions.
- **Risk:** a technically valid old record could be reused beyond its approved
  scope.
- **Required validation:** direct release-gate rejection of R90-04 reuse,
  preservation of historical v0.1.0 behavior, audit/documentation checks, and
  knowledge validation.
- **Stop condition:** stop if expiry would rewrite historical evidence or
  weaken another release gate.

## R90-05 Definition

- **Goal:** prepare v0.1.1 release readiness from the exact approved evidence
  and quality baseline.
- **Risk:** synthetic or scoped-exception evidence can be generalized into an
  unsupported release claim.
- **Required validation:** RC, supply-chain, evidence-integrity, release-gate,
  documentation, and exact exception-boundary checks.
- **Stop condition:** stop on evidence/digest drift, unavailable required
  validation, private-data need, tagging, or publication authority.

## R90-06 Definition

- **Goal:** assemble a hold-state v0.1.1 decision package without creating a
  tag or artifact publication.
- **Risk:** a reconciled candidate package can be mistaken for final
  publication authorization or remain pinned after `main` advances.
- **Required validation:** exact version/commit/artifact/checksum/platform
  reconciliation, RC and release gates, fetched remote verification, and an
  explicit hold decision.
- **Stop condition:** stop before tag creation, GitHub Release, GHCR push, or
  any candidate change that lacks fresh evidence.

## R90-07 Definition

- **Goal:** prevent unbounded UDS connection-handler goroutine growth while
  preserving the capture reconnect path.
- **Risk:** a leaked limiter slot can reject valid capture reconnects; a
  blocking overload path can interfere with shutdown.
- **Required validation:** direct lower/upper config-bound regressions, direct
  excess-client rejection and capacity-reuse regressions, focused receiver and
  config tests, full native tests, documentation/configuration checks, and the
  knowledge gate.
- **Stop condition:** stop if the limit requires a frame-protocol change, an
  overload result is ambiguous, or work reaches tag/publication authority.

## R90-08 Definition

- **Goal:** close the documented active-load shutdown validation gap across the
  full Go engine lifecycle.
- **Risk:** timing-sensitive orchestration can create a flaky test or hide a
  real write-after-close race.
- **Required validation:** a direct integration regression with bounded waits,
  repeated focused race runs, the full native test suite, and the knowledge
  gate.
- **Stop condition:** stop if deterministic orchestration requires production
  traffic, privileged external services, or a runtime architecture change
  broader than shutdown validation.

## R90-09 Definition

- **Goal:** close the first bounded SQLite corruption/fault-injection gap by
  making corrupt startup behavior explicit and recoverable.
- **Risk:** an attempted repair path could overwrite operator data or turn a
  clear startup failure into silent data loss.
- **Required validation:** direct corrupt and truncated database regressions,
  focused alert-store tests, full native tests, documentation checks, and the
  knowledge gate.
- **Stop condition:** stop if safe completion requires automatic database
  repair, deletion, access to operator data, or a broader storage redesign.

## R90-10 Definition

- **Goal:** apply the R90-09 preservation boundary when a running daily-shard
  store targets an existing non-current shard.
- **Risk:** shard initialization can mutate a corrupt historical database
  before returning an error.
- **Required validation:** direct corrupt/truncated historical-shard write
  regressions with byte preservation, focused alert-store race tests, full
  native tests, documentation checks, and the knowledge gate.
- **Stop condition:** stop if completion requires automatic shard repair,
  deletion, operator data, or a redesign of cross-shard storage.

## R90-11 Definition

- **Goal:** remove writable SQLite handles from non-current daily-shard query
  and count paths after R90-10 protected their write path.
- **Risk:** read-only DSN handling can break healthy WAL-backed shard reads or
  obscure useful SQLite errors.
- **Required validation:** direct corrupt/truncated query and count
  preservation regressions, healthy cross-shard compatibility tests, focused
  alert-store race tests, full native tests, documentation checks, and the
  knowledge gate.
- **Stop condition:** stop if safe read-only access requires automatic shard
  repair, snapshots, operator data, or a broader query/storage redesign.

## R90-12 Definition

- **Goal:** extend the storage preservation boundary to durable JSONL recovery
  input before startup replay can truncate it.
- **Risk:** an ambiguous partial-line policy could discard the last recoverable
  alert or turn a clear startup failure into silent data loss.
- **Required validation:** direct corrupt-record and truncated-record startup
  regressions with byte preservation, valid replay compatibility, focused
  alert-store race tests, full native tests, documentation checks, and the
  knowledge gate.
- **Stop condition:** stop if completion requires automatic recovery-log
  repair, partial-record acceptance, operator data, or a replay-format redesign.

## R90-13 Definition

- **Goal:** close the remaining handler-slot exhaustion path after R90-07 by
  expiring connections that deliver no complete frame within a bounded period.
- **Risk:** an overly short or unrefreshed deadline can disconnect healthy
  capture sessions; timeout errors can be misclassified as malformed input.
- **Required validation:** direct config-bound, pre-first-frame timeout,
  per-frame refresh, idle-capacity-reuse, reconnect, and cancellation tests;
  focused receiver/config race tests, full native tests, documentation/config
  checks, E2E smoke, and the knowledge gate.
- **Stop condition:** stop if completion requires a frame-protocol change, UDS
  authentication/peer policy, C capture changes, operator data, or
  tag/publication authority.

## R90-14 Definition

- **Goal:** make the documented hello handshake a per-connection ordering and
  session boundary instead of accepting packet/heartbeat traffic without it.
- **Risk:** connection-local state can accidentally become global, reject a
  valid reconnect, or allow a violating client to keep its handler slot.
- **Required validation:** direct packet-before-hello, heartbeat-before-hello,
  duplicate-hello, and mismatched-session rejection tests; valid hello,
  heartbeat, packet, reconnect, capacity, cancellation, focused receiver race,
  full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if the checked-in C sender does not satisfy the
  proposed ordering, compatibility requires accepting ambiguous legacy
  clients, peer authentication is required, or work reaches tag/publication
  authority.

### R90-14 Reconnect Authorization

- **Detected:** 2026-07-20 during the required sender-ordering preflight.
- **Evidence:** `capture/src/main.c` sends hello only after the initial
  connection. When `uds_send_packet` reports `UDS_ERR_PIPE`, `packet_handler`
  calls `uds_reconnect` but does not send hello on the replacement connection.
  `capture/src/uds_sender.c` confirms that `uds_reconnect` only reconnects the
  socket.
- **Impact:** enforcing a hello as the first frame on every connection would
  close a valid checked-in capture reconnect when its next packet or heartbeat
  arrives, violating the increment's reconnect compatibility criterion.
- **Unblock condition:** obtain product authority to change the C reconnect
  lifecycle so every successful replacement connection sends hello before any
  packet or heartbeat, and include direct C plus end-to-end reconnect coverage;
  or approve a different explicit compatibility contract.
- **Authorization:** On 2026-07-20, the user explicitly authorized changing the
  C reconnect path to resend hello before any packet or heartbeat.
- **Current effect:** The blocker is resolved. R90-14 may update the checked-in
  sender and receiver together, with direct C socket-reconnect, Go
  connection-local state, full native, and E2E validation.

## R90-15 Definition

- **Goal:** extend the R90-09 preservation boundary from corrupt SQLite bytes
  to structurally valid existing databases that do not satisfy NetSentry's
  required alert-store schema.
- **Risk:** an over-strict schema check can reject a compatible database, while
  writable initialization of an unrelated or incompatible database can modify
  operator data before returning an error.
- **Required validation:** direct unrelated-schema and incompatible-alert-table
  startup regressions with byte preservation; compatible existing, empty, and
  missing database compatibility; focused alert-store race tests, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires automatic schema
  migration, repair, deletion, operator data, or a broader storage redesign.

## R90-16 Definition

- **Goal:** extend the R90-12 recovery-input boundary from JSON syntax and line
  termination to the semantic invariants of records written by NetSentry's own
  normalized recovery logger.
- **Risk:** an incomplete validator can persist empty/corrupt alert identities,
  while an over-strict validator can reject legitimate historical recovery
  input.
- **Required validation:** direct null/empty and missing durable-identity/network
  field regressions with full-log byte preservation and no prefix persistence;
  valid replay/idempotency compatibility, focused alert-store race tests, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if the durable semantic contract is ambiguous,
  compatibility requires accepting records that current NetSentry cannot
  generate, automatic log repair is required, or operator data is needed.

## R90-17 Definition

- **Goal:** move the complete R90-12/R90-16 recovery-log integrity boundary
  ahead of every writable SQLite open and initialization step.
- **Risk:** reading the log twice can introduce a validation/replay race, while
  replaying a stale snapshot can ignore an unexpected concurrent append.
- **Required validation:** direct malformed and semantic-invalid startup cases
  proving a missing database remains absent and a compatible existing database
  plus optional-index state remain byte-for-byte unchanged; valid missing and
  existing database replay compatibility, focused alert-store race tests, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires locking recovery input
  across processes, changing the recovery format, automatic repair, operator
  data, or tag/publication authority.

## R90-18 Definition

- **Goal:** complete the normalized recovery-record semantic boundary by
  rejecting internally inconsistent identity, time-window, and count fields
  that the durable writer cannot emit.
- **Risk:** deriving the expected identity or window with different rules from
  the writer can reject valid recovery input, while replay normalization can
  otherwise conceal tampered or partially corrupted fields.
- **Required validation:** direct durable-ID, first-seen, last-seen,
  window-start, and aggregate-count rejection regressions with full-log and
  missing/existing database preservation; valid replay/idempotency
  compatibility, focused alert-store race tests, full native, documentation,
  E2E, and knowledge checks.
- **Stop condition:** stop if the normalized writer contract is ambiguous,
  compatibility requires accepting records the current writer cannot emit,
  automatic log repair is required, operator data is needed, or work reaches
  tag/publication authority.

## R90-19 Definition

- **Goal:** extend the recovery-input preservation boundary from startup to
  normal runtime writes before they append new durable records.
- **Risk:** an extra preflight can accidentally drop valid pending records or
  create a check/append race, while appending first mutates invalid operator
  evidence before the existing integrity failure is reported.
- **Required validation:** direct malformed, truncated, semantic-invalid, and
  normalized-invariant runtime rejection regressions with full-log and database
  byte preservation; valid pending-log persistence compatibility; repeated
  focused alert-store race tests, full native, documentation, E2E, and
  knowledge checks.
- **Stop condition:** stop if safe completion requires cross-process recovery
  locking, changing the recovery format, automatic repair, operator data, or
  tag/publication authority.

## R90-20 Definition

- **Goal:** align recovery-log writing and reading on one explicit bounded
  record size so the store never rejects its own successfully appended output.
- **Risk:** raising the scanner ceiling without a writer bound permits
  excessive allocation, while checking records during streaming can partially
  append a batch before a later oversized record fails.
- **Required validation:** direct above-64-KiB runtime write and startup replay
  compatibility; exact 4-MiB boundary acceptance; above-limit runtime and
  direct-append rejection with full log and database preservation; existing
  malformed/truncated behavior; repeated focused alert-store race tests, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if the durable size contract requires an on-disk
  format migration, accepting unbounded records, automatic repair, operator
  data, or tag/publication authority.

## R90-21 Definition

- **Goal:** close the required-schema gap where an extra mandatory column can
  pass preflight even though NetSentry's fixed inserts cannot populate it.
- **Risk:** rejecting every unknown column would break compatible operator
  extensions, while inspecting defaults incorrectly can accept a write-blocking
  schema or reject a valid nullable/defaulted extension.
- **Required validation:** direct `alerts` and `alert_events` unknown
  `NOT NULL`-without-default startup rejections plus a literal-NULL-default
  rejection with byte preservation; a historical-shard rejection; nullable
  and non-NULL-defaulted compatibility
  with successful writes; focused alert-store race tests, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires schema migration,
  evaluating arbitrary default expressions, rewriting operator tables,
  operator data, or tag/publication authority.

## R90-22 Definition

- **Goal:** close the remaining schema-preflight gap where an extra uniqueness
  constraint can reject valid fixed-column alert or event writes only after
  writable initialization.
- **Risk:** rejecting every operator index would break compatible query
  extensions, while treating a subset or expression uniqueness constraint as
  harmless can preserve a write blocker.
- **Required validation:** direct primary `alerts` subset,
  `alert_events` timestamp-only, expression-only, partial-subset, and
  non-binary-collated identity unique-index rejections with byte preservation;
  a historical-shard rejection; non-unique and binary-identity-containing
  unique-index compatibility with successful writes;
  repeated focused alert-store race tests, full native, documentation, E2E,
  and knowledge checks.
- **Stop condition:** stop if safe completion requires evaluating arbitrary
  index expressions, schema migration, rewriting operator indexes, operator
  data, or tag/publication authority.

## R90-23 Definition

- **Goal:** close the schema-preflight gap where a trigger attached to a
  write-critical table can abort, redirect, or add side effects to valid
  NetSentry writes only after writable initialization.
- **Risk:** inspecting trigger bodies would require interpreting arbitrary SQL,
  while rejecting triggers on unrelated operator tables would unnecessarily
  narrow compatible extensions.
- **Required validation:** direct `alerts` `BEFORE INSERT`, `alerts`
  `AFTER UPDATE`, `alert_events`, and case-variant table-name trigger
  rejections with byte preservation; a historical-shard rejection;
  unrelated-table trigger compatibility with successful writes; repeated
  focused alert-store race tests, full native, documentation, E2E, and
  knowledge checks.
- **Stop condition:** stop if safe completion requires interpreting or
  rewriting trigger SQL, schema migration, operator data, or tag/publication
  authority.

## R90-24 Definition

- **Goal:** close the required-column preflight gap where `PRAGMA table_info`
  hides generated columns whose arbitrary expressions can abort or alter valid
  fixed-column NetSentry writes.
- **Risk:** parsing generated expressions would reproduce SQLite semantics
  incompletely, while rejecting ordinary nullable/defaulted columns would
  break the compatibility retained by R90-21.
- **Required validation:** direct virtual and stored generated-column
  rejections across `alerts` and `alert_events` with byte preservation; a
  historical-shard rejection; ordinary nullable/defaulted column compatibility
  with successful writes; repeated focused alert-store race tests, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires parsing or evaluating
  generated expressions, schema migration, rewriting operator columns,
  operator data, or tag/publication authority.

## R90-25 Definition

- **Goal:** close the schema-preflight gap where a `CHECK` constraint attached
  to a write-critical table can reject valid fixed-column NetSentry writes
  only after writable initialization.
- **Risk:** matching raw schema text without SQLite lexical boundaries can
  mistake strings, comments, or quoted identifiers for constraints, while
  evaluating arbitrary constraint expressions would reproduce SQLite
  semantics incompletely.
- **Required validation:** direct table-level and column-level `CHECK`
  rejections across `alerts` and `alert_events`, including case-variant
  keywords and false-positive lexical boundaries, with byte preservation; a
  historical-shard rejection; unrelated-table constraint compatibility with
  successful writes; repeated focused alert-store race tests, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires evaluating constraint
  expressions, schema migration, rewriting operator constraints, operator
  data, or tag/publication authority.

## R90-26 Definition

- **Goal:** close the schema-preflight gap where foreign-key relationships can
  reject or cascade NetSentry inserts, updates, and retention deletes when
  SQLite foreign-key enforcement is active.
- **Risk:** inspecting only outgoing relationships misses unrelated tables
  that reference write-critical tables, while rejecting relationships confined
  to operator tables would unnecessarily narrow compatible extensions.
- **Required validation:** direct outgoing `alerts` and `alert_events`
  rejections plus an incoming and case-variant relationship rejection with
  byte preservation; a historical-shard rejection; unrelated-table
  relationship compatibility with successful writes; repeated focused
  alert-store race tests, full native, documentation, E2E, and knowledge
  checks.
- **Stop condition:** stop if safe completion requires enabling or evaluating
  foreign-key actions, schema migration, rewriting operator relationships,
  operator data, or tag/publication authority.

## R90-27 Definition

- **Goal:** close the required aggregation-schema gap where a canonical
  uniqueness key with non-binary collation can merge distinct NetSentry alert
  identities even though its column order passes preflight.
- **Risk:** inspecting only column names misses SQLite collation semantics,
  while rejecting additional compatible indexes would narrow the extension
  policy established by R90-22.
- **Required validation:** direct inline and explicit non-binary aggregation
  uniqueness rejections with byte preservation; a historical-shard rejection;
  binary-collated canonical-key compatibility with successful writes of
  identities that differ only by case; repeated focused alert-store race tests,
  full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing the canonical
  aggregation identity, schema migration, rewriting operator indexes, operator
  data, or tag/publication authority.

## R90-28 Definition

- **Goal:** align required-column and unique-index metadata comparisons with
  SQLite's case-insensitive identifier semantics.
- **Risk:** partial normalization can still reject compatible index metadata,
  while broad normalization could weaken unknown-column or binary-collation
  checks.
- **Required validation:** direct primary and historical case-variant
  compatibility writes; existing incompatible-schema and non-binary-collation
  regressions; twenty focused alert-store race runs, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires schema migration,
  identifier rewriting, weakening a write-safety constraint, operator data, or
  tag/publication authority.

## R90-29 Definition

- **Goal:** make documented exact-match alert predicates independent of
  compatible operator-declared SQLite column collations.
- **Risk:** applying binary collation too broadly could break intentionally
  case-insensitive protocol or MITRE filters, while fixing only list selection
  could leave filtered counts or cross-shard results inconsistent.
- **Required validation:** direct rule, severity, source, and destination
  primary-query regressions against compatible `NOCASE` columns; a historical
  cross-shard regression; existing protocol/MITRE compatibility; twenty focused
  alert-store race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires schema migration,
  rejecting a compatible database, changing public filter semantics, operator
  data, or tag/publication authority.

## R90-30 Definition

- **Goal:** reject persisted numeric alert fields that cannot satisfy the
  public model before conversion or return.
- **Risk:** integer narrowing can silently wrap invalid ports, while accepting
  non-positive aggregate counts exposes states the writer cannot generate.
- **Required validation:** direct negative and above-65535 port rejection;
  direct zero and negative aggregate-count rejection; historical read-only
  rejection with byte preservation; healthy primary/cross-shard compatibility;
  twenty focused alert-store race runs, full native, documentation, E2E, and
  knowledge checks.
- **Stop condition:** stop if safe completion requires automatic row repair,
  deletion, schema migration, a full-table startup scan, operator data, or
  tag/publication authority.

## R90-31 Definition

- **Goal:** reject persisted severity values outside the public alert enum
  before returning or classifying a row.
- **Risk:** empty severity can be silently classified as low downstream, while
  arbitrary or case-variant values can escape the documented API contract.
- **Required validation:** direct empty, uppercase-known, and unsupported
  severity rejection across list/query decoding; historical read-only rejection
  with byte preservation; healthy severity compatibility; twenty focused
  alert-store race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires automatic row repair,
  deletion, schema migration, changing the public severity enum, operator data,
  or tag/publication authority.

## R90-32 Definition

- **Goal:** reject syntactically valid persisted timestamp ordering that the
  aggregation writer cannot produce.
- **Risk:** reversed first/last timestamps break aggregate ordering, while a
  window start after the first event cannot describe the stored aggregate.
- **Required validation:** direct first-after-last and window-after-first
  rejection across list/query decoding; historical read-only rejection with
  byte preservation; healthy timestamp compatibility; twenty focused
  alert-store race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires automatic row repair,
  deletion, schema migration, assuming a historical aggregation-window
  duration, operator data, or tag/publication authority.

## R90-33 Definition

- **Goal:** reject persisted alert IDs that the aggregation writer cannot
  derive from the row's canonical aggregation tuple.
- **Risk:** duplicated identity derivation can drift between writer and reader,
  while accepting an altered ID exposes an identity unrelated to the stored
  aggregation key.
- **Required validation:** direct empty and altered ID rejection across
  list/query decoding; historical read-only rejection with byte preservation
  through an encoded filesystem path; healthy aggregation and cross-shard
  compatibility; twenty focused alert-store race runs, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing the aggregation
  identity, automatic row repair, deletion, schema migration, event-ledger
  reconciliation, operator data, or tag/publication authority.

## R90-34 Definition

- **Goal:** reject persisted required public text fields that the rule,
  receiver, and durable recovery contracts do not permit to be blank.
- **Risk:** validating every non-null text column would reject legitimate empty
  optional fields, while dependent identity validation can obscure which
  required field is corrupt.
- **Required validation:** direct blank `event_id`, `rule_id`, `rule_name`,
  `protocol`, `src_ip`, and `dst_ip` rejection across list/query decoding;
  historical read-only rejection with byte preservation through an encoded
  filesystem path; optional-text and healthy aggregation compatibility; twenty
  focused alert-store race runs, full native, documentation, E2E, and knowledge
  checks.
- **Stop condition:** stop if safe completion requires validating optional
  fields, changing the public protocol or address contract, event-ledger
  reconciliation, automatic row repair, deletion, schema migration, operator
  data, or tag/publication authority.

## R90-35 Definition

- **Goal:** align UDS packet address validation with the C capture parser and
  documented IPv4-only v0.1 contract.
- **Risk:** generic IP parsing accepts IPv6, while `To4`-style checks can also
  accept IPv4-mapped IPv6 text; rejected frames must not enter the packet queue
  or double-count decode errors.
- **Required validation:** direct ordinary and IPv4-mapped IPv6 source and
  destination rejection; malformed-address and valid IPv4 compatibility;
  decode-error and no-enqueue assertions; twenty focused receiver race runs,
  full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires IPv6 product support, a
  packet-schema change, C capture changes, address normalization, stored-data
  migration, operator data, or tag/publication authority.

## R90-36 Definition

- **Goal:** extend the strict IPv4 ingress contract to durable recovery records
  before startup replay or runtime append can modify state.
- **Risk:** dependent identity validation can obscure an invalid address, while
  late validation can create or initialize SQLite or append a new recovery
  record before failing.
- **Required validation:** direct malformed, ordinary IPv6, and IPv4-mapped
  IPv6 source/destination rejection; missing and existing database startup
  preservation with a valid prefix; runtime log/database preservation; valid
  replay/write compatibility; twenty focused alert-store race runs, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires IPv6 support, a
  recovery-format change, address normalization, stored-row migration,
  automatic log repair, operator data, or tag/publication authority.

### R90-36 Validation Deviation

- **Observed:** The first full native race suite failed the four
  `TestStoreExactFiltersOverrideCompatibleNoCaseColumns` cases because their
  shared fixture wrote IPv6 source/destination text through the recovery path.
- **Impact:** Delivery was held pending a clean full-suite rerun. The strict
  recovery behavior itself passed all focused tests.
- **Resolution:** Keep rule/severity compatibility on valid IPv4 writer input.
  Seed the source/destination collation-only rows below the recovery boundary
  because that test exercises persisted SQL comparison semantics. The affected
  test then passed twenty uncached race runs, and the combined focused and
  complete native race suites passed.

## R90-37 Definition

- **Goal:** extend the strict IPv4 contract to persisted SQLite alerts before
  rows can be exposed through list or query reads.
- **Risk:** dependent aggregation-identity validation can obscure an invalid
  address, while applying validation outside the shared row decoder can leave
  primary and historical behavior inconsistent.
- **Required validation:** direct malformed, ordinary IPv6, and IPv4-mapped
  IPv6 source/destination rejection across list/query decoding; historical
  read-only rejection with byte preservation through an encoded filesystem
  path; healthy primary and cross-shard compatibility; twenty focused
  alert-store race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires IPv6 product support,
  address normalization, automatic row repair, deletion, schema migration, a
  full-table startup scan, operator data, or tag/publication authority.

### R90-37 Validation Deviation

- **Observed:** The first full native race suite failed the source and
  destination cases in
  `TestStoreExactFiltersOverrideCompatibleNoCaseColumns` because R90-36 had
  moved their case-variant IPv6 rows below recovery, while R90-37 now correctly
  rejects those rows at shared stored-row decoding.
- **Impact:** Delivery was held pending fixture reconciliation and a clean
  full-suite rerun. R90-37's direct behavior and twenty focused race runs
  passed.
- **Resolution:** Preserve the compatible `NOCASE` schema coverage with
  distinct valid IPv4 source/destination rows. Case-variant address text is no
  longer a valid stored-row fixture under the strict IPv4 contract. The
  affected test then passed twenty uncached race runs, and the combined focused
  and complete native race suites passed.

## R90-38 Definition

- **Goal:** require each durable recovery record's `event_id` to match the
  deterministic event identity used by the writer and idempotency ledger.
- **Risk:** accepting an altered identity can bypass or collide with replay
  deduplication, while late validation can create or initialize SQLite or append
  a new recovery record before failing.
- **Required validation:** direct nonblank event-identity mismatch with a valid
  prefix; missing and existing database startup preservation; runtime
  log/database preservation; valid replay/idempotency compatibility; twenty
  focused alert-store race runs, full native, documentation, E2E, and knowledge
  checks.
- **Stop condition:** stop if safe completion requires changing event-ID
  derivation, rewriting the recovery format, stored-row or event-ledger
  reconciliation, automatic repair, operator data, or tag/publication
  authority.

## R90-39 Definition

- **Goal:** apply the existing public severity enum to durable recovery records
  before startup replay or runtime append can modify state.
- **Risk:** empty or arbitrary severity can be persisted only to fail later
  during row decoding, while normalization would conceal invalid durable input.
- **Required validation:** direct empty, case-variant, and unsupported severity
  rejection with a valid prefix; missing and existing database startup
  preservation; runtime log/database preservation; direct compatibility for all
  four public severity values; twenty focused alert-store race runs, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing or normalizing
  the public severity enum, rewriting the recovery format, stored-row migration,
  automatic repair, operator data, or tag/publication authority.

### R90-39 Validation Deviation

- **Observed:** The first complete native race suite failed because the
  collation-independent severity-filter regression used `WriteBatch` to create
  an intentionally case-variant stored severity.
- **Cause and impact:** The new recovery contract correctly rejects that
  invalid writer input, so the fixture no longer reached its separate query
  concern. The regression now writes valid alerts and directly updates the
  intentionally invalid stored row before exercising the binary query filter.
- **Resolution:** The affected focused race run and the complete native race
  suite passed after the fixture correction. Scope, dates, and runtime behavior
  did not change.

## R90-40 Definition

- **Goal:** align durable recovery records with the existing rule-loader and
  stored-row requirement that `rule_name` is nonblank.
- **Risk:** blank recovery rule names can currently persist and fail only on a
  later row read, while normalization would alter durable public text.
- **Required validation:** direct missing, empty, and whitespace-only rule-name
  rejection with a valid prefix; missing and existing database startup
  preservation; runtime log/database preservation; nonblank padded-name replay
  without normalization; stored-row compatibility; twenty focused alert-store
  race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires normalizing rule names,
  changing rule schema or alert identity, rewriting the recovery format,
  stored-row migration, automatic repair, operator data, or tag/publication
  authority.

## R90-41 Definition

- **Goal:** align stored alert MITRE tuple decoding with the rule engine's
  all-empty or fully populated emission contract.
- **Risk:** partial tuples can expose an ID without its tactic/name or
  whitespace-only public metadata, while catalog revalidation could reject
  legitimate historical complete tuples.
- **Required validation:** all six partial empty/populated tuple shapes and
  whitespace-only tactic, technique ID, and technique name rejection across
  list/query decoding; historical read-only rejection with byte preservation
  through an encoded filesystem path; complete and all-empty compatibility;
  twenty focused alert-store race runs, full native, documentation, E2E, and
  knowledge checks.
- **Stop condition:** stop if safe completion requires revalidating stored
  tuples against the current MITRE catalog, normalizing text, changing filter
  semantics, recovery-format validation, automatic row repair, deletion,
  schema migration, operator data, or tag/publication authority.

## R90-42 Definition

- **Goal:** align durable recovery records with the rule engine and stored-row
  all-empty-or-fully-populated MITRE tuple contract.
- **Risk:** partial tuples can currently persist and fail only on a later row
  read, while current-catalog validation or normalization could reject or alter
  legitimate historical complete tuples.
- **Required validation:** all six partial empty/populated tuple shapes and
  whitespace-only tactic, technique ID, and technique name rejection with a
  valid prefix; missing and existing database startup preservation; runtime
  log/database preservation; all-empty and complete unnormalized replay
  compatibility; stored-row compatibility; twenty focused alert-store race
  runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires current-catalog
  revalidation, text normalization, a recovery-format change, stored-row
  migration, automatic repair, operator data, or tag/publication authority.

## R90-43 Definition

- **Goal:** align stored alert protocol decoding with the rule engine's
  canonical IP protocol-name emission contract.
- **Risk:** arbitrary stored protocol text can currently escape through the API,
  while over-restricting unknown IP protocol numbers could reject values the
  current writer legitimately emits.
- **Required validation:** case-variant named protocol, unsupported name,
  malformed, noncanonical, named-protocol numeric alias, and out-of-range
  `PROTO_` rejection across list/query decoding; historical read-only rejection
  with byte preservation through an encoded filesystem path; named and unknown
  boundary-value compatibility; focused alert-store, rule, and shared-model
  tests, twenty focused alert-store race runs, full native, documentation, E2E,
  and knowledge checks.
- **Stop condition:** stop if safe completion requires restricting the UDS IP
  protocol number, changing query-filter case semantics, rewriting stored data,
  schema migration, operator data, or tag/publication authority.

## R90-44 Definition

- **Goal:** align durable recovery records with the shared canonical IP
  protocol-name contract already enforced at rule emission and stored-row
  decoding.
- **Risk:** nonblank but noncanonical recovery protocol text can currently pass
  preflight and modify state before a later stored-row read rejects it, while
  over-restricting unknown IP protocol numbers could reject values the current
  writer legitimately emits.
- **Required validation:** direct startup and runtime rejection for a
  case-variant named protocol, unsupported name, malformed, noncanonical,
  named-protocol numeric alias, and out-of-range `PROTO_` form; complete-log
  plus missing/existing database preservation; named and unknown
  boundary-value replay/write compatibility; twenty focused alert-store race
  runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires restricting the UDS IP
  protocol number, changing query-filter semantics, a recovery-format change,
  stored-row migration, automatic repair, operator data, or tag/publication
  authority.

## R90-45 Definition

- **Goal:** close the runtime boundary where `WriteBatch` validates only the
  existing recovery log before appending newly normalized records.
- **Risk:** validating during or after append can partially write a valid
  prefix, while treating invalid caller input as a storage fault can degrade a
  healthy store.
- **Required validation:** direct valid-prefix current-batch rejection for
  every reachable required-text, event-identity, severity, MITRE, protocol, and
  IPv4 validation category; pre-existing valid-log plus SQLite byte
  preservation; healthy-status preservation; valid pending/current
  compatibility; twenty focused alert-store race runs, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a recovery-format or
  SQLite-schema change, cross-process locking, changing public alert semantics,
  automatic repair, operator data, or tag/publication authority.

## R90-46 Definition

- **Goal:** align stored aggregate timestamp decoding with the exact UTC
  RFC3339Nano text emitted by NetSentry before SQLite text comparisons see
  alternate encodings.
- **Risk:** accepting parseable but noncanonical offsets or redundant
  fractional precision can make SQLite lexical comparisons disagree with the
  decoded instants, while over-validation could reject legitimate writer
  output.
- **Required validation:** direct `first_seen`, `last_seen`, and
  `window_start` rejection for explicit UTC offsets, non-UTC offsets, and
  nonminimal fractional forms across list/query decoding; historical read-only
  rejection with byte preservation through an encoded filesystem path;
  healthy primary and cross-shard compatibility; twenty focused alert-store
  race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a schema migration,
  rewriting stored timestamps, changing recovery JSON or public time
  semantics, validating `created_at`/`updated_at`, a full-table startup scan,
  operator data, or tag/publication authority.

### R90-46 Scope Observation

- **Observed:** Go's canonical RFC3339Nano output omits or trims fractional
  seconds, so exact writer-format validation does not by itself prove that
  every variable-width timestamp string has chronological lexical order.
- **Impact:** R90-46 rejects alternate offset and redundant-precision input but
  makes no claim that SQLite time comparison, aggregation, ordering, or pruning
  semantics are fully corrected.
- **Follow-up:** Refresh the queue after R90-46 delivery with a separate
  increment that pins SQL time comparisons without silently migrating or
  rewriting stored rows.

## R90-47 Definition

- **Goal:** make every SQL comparison over stored aggregate timestamps preserve
  chronological order for canonical RFC3339Nano values with absent, trimmed,
  or full fractional seconds.
- **Risk:** SQLite date helpers can lose sub-millisecond precision or bypass
  existing indexes, while a fixed-width storage migration would rewrite
  operator data outside the accepted preservation boundary.
- **Required validation:** direct mixed-width and sub-millisecond aggregation
  earliest/latest selection; primary and historical ordering/pagination plus
  `since`/`until` filtering; retention-boundary pruning; query-plan or focused
  performance review; twenty focused alert-store race runs, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a schema or stored-data
  migration, loses nanosecond ordering fidelity, changes public time/filter
  semantics, needs operator data, or reaches tag/publication authority.

## R90-48 Definition

- **Goal:** align every durable recovery timestamp with the exact UTC
  RFC3339Nano string emitted by NetSentry rather than accepting alternate JSON
  timestamp spellings that decode to the same instant.
- **Risk:** validating only decoded `time.Time` values loses the original
  offset and fractional spelling, while raw JSON inspection could accidentally
  diverge from the model decoder or reject canonical writer output.
- **Required validation:** direct `timestamp`, `first_seen`, `last_seen`, and
  `window_start` rejection for explicit UTC offsets, equivalent non-UTC
  offsets, and nonminimal fractional forms; startup preservation for missing
  and compatible existing databases; runtime log/database preservation;
  canonical replay compatibility; twenty focused alert-store race runs, full
  native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing the recovery
  JSON format, accepting a timestamp the current writer cannot emit, automatic
  log repair, operator data, or tag/publication authority.

## R90-49 Definition

- **Goal:** reject ambiguous duplicate top-level names in durable recovery JSON
  before Go's decoder silently keeps the last value.
- **Risk:** checking only exact text misses case-variant names that target the
  same exported model field, while recursively policing unknown nested JSON
  would broaden the durable schema beyond this increment.
- **Required validation:** direct identical and conflicting duplicate durable
  field rejection; case-variant alias rejection under Go's field-matching
  semantics; duplicate unknown top-level name rejection; valid-prefix and
  missing/existing database preservation at startup; runtime log/database
  preservation; canonical writer compatibility; twenty focused alert-store
  race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing the recovery
  JSON format, recursively constraining unknown nested values, automatic log
  repair, operator data, or tag/publication authority.

## R90-50 Definition

- **Goal:** make the durable recovery member vocabulary equal the current
  writer's exact JSON tags instead of silently ignoring unknown names or
  accepting case-insensitive aliases.
- **Risk:** omitting an optional writer field from the allowlist would make the
  store reject its own output, while returning a field-name error too early
  could obscure duplicate or malformed JSON diagnostics.
- **Required validation:** direct scalar and nested unknown top-level rejection;
  direct case-variant supported-name rejection; duplicate and malformed error
  precedence; complete-log plus missing/existing database preservation at
  startup; runtime log/database preservation; canonical writer compatibility
  with empty and populated optional `raw_payload`; twenty focused alert-store
  race runs, full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a versioned recovery
  migration, accepting a field the current writer cannot emit, recursively
  constraining value objects, automatic log repair, operator data, or
  tag/publication authority.

## R90-51 Definition

- **Goal:** prevent Go zero-value decoding from accepting incomplete durable
  recovery objects that the current writer cannot emit.
- **Risk:** a required-field list can drift from the writer, while checking
  presence before completing the structural parse can obscure duplicate,
  unsupported-name, or malformed-record diagnostics.
- **Required validation:** direct removal of every non-`omitempty` writer field
  at startup and runtime; complete-log plus missing/existing database
  preservation; duplicate, unsupported-name, and malformed diagnostic
  precedence; canonical writer compatibility with omitted and populated
  `raw_payload`; twenty focused alert-store race runs, full native,
  documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a versioned recovery
  migration, making `raw_payload` mandatory, changing JSON value/null
  semantics, changing the recovery format, automatic repair, operator data,
  or tag/publication authority.

## R90-52 Definition

- **Goal:** prevent JSON `null` or a mismatched top-level JSON kind from
  becoming an accepted Go zero value that the current recovery writer cannot
  emit.
- **Risk:** a field-kind map can drift from the writer, while returning a value
  error before completing the structural parse can obscure duplicate,
  unsupported-name, or malformed-record diagnostics.
- **Required validation:** direct null rejection for every writer field;
  representative wrong-kind text, timestamp, and numeric rejections; complete
  log plus missing/existing database preservation at startup; runtime
  log/database preservation; duplicate, unsupported-name, and malformed
  diagnostic precedence; writer-kind alignment and canonical replay with
  omitted and populated `raw_payload`; twenty focused alert-store race runs,
  full native, documentation, E2E, and knowledge checks.
- **Stop condition:** stop if safe completion requires a versioned recovery
  migration, making `raw_payload` mandatory, recursively constraining JSON
  values, changing canonical numeric spelling, automatic repair, operator
  data, or tag/publication authority.

## R90-53 Definition

- **Goal:** reconcile recent delivery evidence, restore a complete forward
  queue, and make plan auditing repeatable on every trigger.
- **Audit record:**
  [`delivery-plan-audit-20260730.md`](../audit/delivery-plan-audit-20260730.md).
- **Risk:** volume-based conclusions, rewritten history, or speculative future
  work can create false confidence and false commitments.
- **Required validation:** commit phase/count review, all task-state JSON
  parsing, complete roadmap entry/definition coverage, unfinished-item field
  audit, documentation and knowledge checks, diff review, and sensitive-data
  review.
- **Stop condition:** stop if completion requires historical rewrite, runtime
  implementation, private evidence, product/release authority, or starting
  R90-54.

## R90-54 Definition

- **Goal:** align recovery numeric representation with the exact integer JSON
  spelling emitted by the writer.
- **Risk:** semantic decoding can discard exponent, fractional, sign, or
  leading-zero representation differences before validation.
- **Required validation:** direct startup/runtime rejection for every planned
  alternate numeric spelling across both numeric fields; missing/existing
  database and log byte preservation; canonical writer compatibility;
  focused race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if safe completion requires a recovery migration,
  accepting a writer-impossible number, operator data, or publication.

## R90-55 Definition

- **Goal:** remove independently maintained recovery name, presence, and kind
  lists that can drift from `model.Alert` writer behavior.
- **Risk:** runtime reflection or incomplete tag parsing can weaken deterministic
  diagnostics or mishandle `omitempty`.
- **Required validation:** direct model-to-contract alignment tests, missing,
  alias, value-kind, optional-field, and canonical writer regressions; focused
  race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if the shared contract changes public JSON, field
  order, diagnostic precedence, or requires a format migration.

## R90-56 Definition

- **Goal:** close a bounded SQLite fault-injection gap around corrupt or
  inconsistent WAL/SHM sidecars.
- **Risk:** opening a sidecar fixture incorrectly can checkpoint, delete,
  recreate, or otherwise modify operator evidence.
- **Required validation:** deterministic primary and encoded-path historical
  sidecar rejection with separate read-only handles and byte comparisons;
  healthy direct and database-symlink active-WAL compatibility; focused race,
  full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if deterministic validation would mutate fixtures,
  require operator data, depend on privileged storage faults, or perform
  automatic repair.

## R90-57 Definition

- **Goal:** define safe restart-free recovery semantics for sticky storage
  emergency mode before any implementation.
- **Risk:** automatic retry can duplicate writes, race current writers, hide
  persistent faults, or delete recovery evidence.
- **Required validation:** reviewed state-machine invariants, concurrency and
  operator-control threat review, failure/retry test plan, documentation, and
  knowledge checks.
- **Authorization:** the Aug 1 global eligibility instruction removes the
  additional product-review gate; R90-57 uses the safest bounded default of an
  operator-triggered probe with no background retry or automatic cleanup.
- **Stop condition:** stop if defining the state machine requires runtime/API
  implementation, automatic cleanup, deleting recovery evidence, private
  operator data, or more than this single documentation increment.

## R90-58 Definition

- **Goal:** refresh the v0.1.1 hold-state candidate package after the completed
  hardening sequence.
- **Risk:** reusing the historical candidate, artifact, or checksum after
  `main` advanced would bind publication evidence to the wrong code.
- **Required validation:** fresh RC, supply-chain, release gate, artifact
  checksum/platform reconciliation, fetched remote verification, and explicit
  hold-state documentation.
- **Stop condition:** stop on ambiguous validation, unavailable required
  infrastructure, version/SHA drift, tag creation, or publication.

## R90-59 Definition

- **Goal:** execute separately authorized remote v0.1.1 publication and verify
  immutable external outcomes after the local tag boundary is complete.
- **Risk:** tagging the wrong commit or inferring workflow/registry success can
  create an unrecoverable public release mismatch.
- **Required validation:** exact remaining authorization/version/SHA match,
  local tag and signature revalidation before push, GitHub Release
  assets/checksums, GHCR digest/platform, workflow result, documentation,
  remote, and Vault evidence.
- **Blocker evidence:** On Aug 7 the user authorized only creation of local tag
  `v0.1.1` at candidate
  `78cd78574e03c8f73ff68248eed2c409d6bca406`; GitHub Release and GHCR remain
  unauthorized. The current tag-push workflows would trigger both outputs.
- **Unblock condition:** after changelog and smoke review, the user explicitly
  authorizes pushing the exact tag and the tag-triggered GitHub Release and
  GHCR actions.
- **Stop condition:** remain blocked without explicit publication authority;
  stop on any SHA, tag, digest, platform, workflow, or artifact ambiguity.

## R90-59a Definition

- **Goal:** create an authenticated local release reference for the exact
  authorized v0.1.1 candidate without crossing the remote publication boundary.
- **Risk:** pushing the tag would immediately trigger both currently
  unauthorized publication workflows, while an unsigned or mistargeted local
  tag would not provide an acceptable immutable release reference.
- **Required validation:** exact user authorization/version/SHA reconciliation;
  candidate changelog and release-evidence review; isolated clean candidate RC,
  E2E/archive smoke, and release gate; annotated tag signature and peeled-target
  verification; direct remote-tag absence; documentation, knowledge, remote
  branch, and Vault checks.
- **Stop condition:** stop on candidate, changelog, smoke, tag, or signature
  ambiguity; inability to keep the tag local; any request to push a tag,
  dispatch a workflow, create a GitHub Release, publish GHCR, change the
  candidate/workflows, access private data, or start R90-75.
- **Selected plan:**
  [`task-20260807-v0.1.1-local-tag.md`](task-20260807-v0.1.1-local-tag.md),
  from clean fetched baseline
  `c19067172f1c626a59ba11b3201b276092721192`. The tag remains local because
  both checked-in tag-push workflows perform external publication.

### R90-59a Publication Boundary Observation

- **Changelog:** Candidate `CHANGELOG.md` has no versioned `0.1.1` heading;
  the release content remains under `[Unreleased]`. This does not alter the
  explicitly authorized local tag target, but remote publication remains
  blocked pending the user's changelog approval.
- **Fresh artifact:** The accepted fresh smoke build produced a 9,760,151-byte
  archive with SHA-256 `fd91e8f3...`, distinct from the historical R90-58
  9,760,241-byte archive with SHA-256 `c68e09df...`. The artifacts are not
  treated as equivalent; later publication must reconcile its exact output.
- **Tag:** Local signed annotated tag object `f1a38ecb82b9c63e8411f3df040bdea84e985dd8`
  peels exactly to the authorized candidate and verifies with the expected SSH
  signer. The remote tag remains absent, so neither publication workflow ran.

## R90-60 Definition

- **Goal:** implement the R90-57 operator-triggered recovery state machine for
  primary and daily-sharded alert stores plus its authenticated API control.
- **Risk:** incorrect lifecycle ownership can use a closing handle, duplicate
  replay, erase recovery evidence, or allow concurrent recovery attempts.
- **Required validation:** direct ownership, cancellation-before-readiness,
  preflight byte-preservation, empty/pending-log success, writable failure,
  idempotent retry, daily-shard, encoded-path, authentication, health, audit,
  focused repeated race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if completion requires background retry, automatic
  cleanup, database/recovery-format migration, private operator data, release
  publication, or behavior beyond this one recovery-control increment.

## R90-61 Definition

- **Goal:** reconcile the completed restart-free recovery delivery and repair
  the now-empty dependency-ready engineering queue from current repository
  evidence.
- **Audit record:**
  [`delivery-plan-audit-20260802.md`](../audit/delivery-plan-audit-20260802.md).
- **Risk:** speculative planning or historical rewrite can create unsupported
  commitments or conceal a validation gap.
- **Required validation:** recent phase/count review, all task-state JSON
  parsing, exact row/Definition coverage, unfinished-item field audit,
  documentation, knowledge, diff, and sensitive-information checks.
- **Stop condition:** stop if completion requires runtime implementation,
  private evidence, historical rewrite, a product/release decision, or starting
  R90-62.

## R90-62 Definition

- **Goal:** directly prove the committed-prefix retry invariant promised by
  the R90-57/R90-60 storage-recovery design for daily shards.
- **Risk:** nondeterministic shard order or a test-only timing hook can create a
  flaky proof, while a real later-shard failure can leave earlier commits that
  must not inflate on retry.
- **Required validation:** deterministic shard-order review; direct later-shard
  failure and active-replay cancellation after an earlier commit; full-log and
  sticky-emergency preservation; idempotent explicit retry with one event and
  aggregate count per input; twenty uncached focused race runs, full native,
  E2E, documentation, and knowledge checks.
- **Stop condition:** stop if deterministic proof requires production data,
  cross-process recovery ownership, rollback-by-copy, automatic cleanup, a
  storage-format migration, or publication authority.

## R90-63 Definition

- **Goal:** extend the existing C ASan fuzz boundary from frame parsing to the
  handwritten UDS packet, heartbeat, and hello JSON formatters.
- **Risk:** unconstrained structured-input generation can manufacture invalid
  C strings or make success assertions meaningless, while a harness that only
  checks for crashes can miss truncated output accepted as valid.
- **Required validation:** deterministic structured seeds and mutations;
  sanitizer coverage for escaping, payload boundaries, integer extremes, and
  exact-fit/undersized output buffers; successful JSONL decode and frame-kind
  invariants; direct truncation rejection; C tests, shell/docs checks, full
  native, E2E, and knowledge checks.
- **Stop condition:** stop if completion requires changing the UDS wire schema,
  accepting noncanonical JSON, adding a C runtime dependency, private corpora,
  or publication authority.

## R90-64 Definition

- **Goal:** record one current reproducible sustained ASan baseline across the
  parser and formatter harnesses after R90-63 closes the harness gap.
- **Risk:** cached, path-bearing, or underspecified results can be mistaken for
  repeated execution, public corpus provenance, or production throughput
  evidence.
- **Required validation:** repository-pinned tool preflight; uncached sustained
  parser and formatter runs at the recorded iteration budget; optional corpus
  inventory with paths redacted; zero crashes and sanitizer findings; evidence
  schema/content checks, full native, documentation, and knowledge checks.
- **Stop condition:** stop on a crash, sanitizer finding, ambiguous iteration
  count, sensitive path exposure, need for private corpus access, or an attempt
  to use the result as tag/publication or production-traffic authority.
- **Selected plan:**
  [`task-20260803-sustained-fuzz-baseline.md`](task-20260803-sustained-fuzz-baseline.md),
  from clean fetched baseline
  `33bc37d9ff71932d6e4ea49cf414f3ed0008415a`. The accepted run uses both
  built-in deterministic harnesses at 1,000,000 iterations each without an
  external corpus and records only path-redacted local synthetic evidence.

## R90-65 Definition

- **Goal:** reconcile the completed dual-harness fuzz delivery against current
  public gap claims and restore a bounded dependency-ready local hardening
  queue without inventing external evidence.
- **Risk:** broad corruption/fault-injection language can produce speculative
  or duplicate work, while external fuzz/traffic gaps can be incorrectly
  treated as locally satisfiable.
- **Required validation:** exact R90-64 feature/closure Git, task-state, remote,
  and Vault evidence; code/test comparison for every public remaining-gap
  claim; task-state JSON parsing; exact roadmap row/Definition coverage;
  complete fields for each new unfinished increment; documentation, knowledge,
  diff, and sensitive-information checks.
- **Stop condition:** stop without implementation if the next bounded queue
  requires private/external corpora, a product or release decision, historical
  evidence rewrite, publication authority, or starting a later increment.
- **Selected plan:**
  [`task-20260803-fuzz-delivery-audit.md`](task-20260803-fuzz-delivery-audit.md),
  from clean fetched baseline
  `23983e1ac696b923a4595e7b97f0e7e1d935dc97`. The audit treats historical
  plans as immutable evidence, separates external-input gaps from ready local
  work, and does not implement R90-66.

## R90-66 Definition

- **Goal:** directly prove ordinary primary-store writes are replay-safe when
  SQLite contention or context cancellation interrupts work after the durable
  recovery append but before transaction commit.
- **Risk:** a test that cancels before `WriteBatch` starts or merely closes the
  database does not reach the active transaction boundary; a timing-only test
  can also pass without proving the log was durable first.
- **Required validation:** real independent SQLite lock contention; active
  cancellation synchronized on observation of the complete recovery record;
  exact log preservation; independent read-only proof of no event or aggregate
  mutation before retry; one explicit retry with one event and aggregate count
  per input; twenty uncached focused race runs, full native, E2E,
  documentation, and knowledge checks.
- **Stop condition:** stop if deterministic proof needs a production failpoint,
  fixed sleeps, cross-process ownership support, automatic evidence cleanup,
  a storage-format migration, or publication authority.
- **Selected plan:**
  [`task-20260803-primary-write-interruption-recovery.md`](task-20260803-primary-write-interruption-recovery.md),
  from clean fetched baseline
  `667cedc72dec9ce58fc7c12aff3be2d37e9ab835`. Both direct cases use real
  SQLite contention, one pre-opened read-only observer, and no production
  failpoint or fixed sleep.

## R90-67 Definition

- **Goal:** make every recovery-log append lifecycle failure directly
  injectable and prove it cannot mutate SQLite or erase an earlier valid log
  prefix.
- **Risk:** broad filesystem simulation can alter production semantics, while
  checking only open failure misses partial write, sync, and close outcomes
  after bytes have reached the file.
- **Required validation:** direct open, short-write, sync, and close failure
  regressions; exact pre-existing-prefix preservation; independent read-only
  SQLite non-mutation proof; precise phase diagnostics and health state;
  complete appended records replay once while an incomplete suffix remains
  fail-closed and preserved; successful-path compatibility, focused race, full
  native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if coverage requires privileged mounts, destructive
  host faults, automatic truncation of failed evidence, a recovery-format
  change, cross-process write ownership, or publication authority.
- **Selected plan:**
  [`task-20260804-recovery-log-append-lifecycle.md`](task-20260804-recovery-log-append-lifecycle.md),
  from clean fetched baseline
  `2f62acf9025969a50dd0295f3881ce7cd2784ec6`. Injection is store-local and
  preserves the production `os.OpenFile` path by default; the direct evidence
  uses a pre-opened read-only SQLite observer and real file bytes.

## R90-68 Definition

- **Goal:** make recovery-log clearing durable and directly prove every failure
  after committed primary or daily-shard persistence remains lossless and
  idempotently recoverable.
- **Risk:** an injected error can occur before truncation, after the file is
  already empty, or during durability/close handling; treating those states as
  identical can overstate retained evidence or conceal a committed alert.
- **Required validation:** direct open/truncate, sync, and close failure
  regressions after observed database commit; exact classification of retained
  versus already-cleared log state; independent read-only proof that every
  event exists once with no aggregate inflation; explicit retry from each
  outcome to healthy state; primary and encoded daily-shard paths, focused
  race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if completion requires deleting an uncommitted log,
  rolling back a committed SQLite transaction, weakening sticky emergency
  semantics, filesystem-specific privileged infrastructure, a format
  migration, or publication authority.
- **Selected plan:**
  [`task-20260804-recovery-log-clearing-lifecycle.md`](task-20260804-recovery-log-clearing-lifecycle.md),
  from clean fetched baseline
  `cac3178512a84356364f82261f2b7dffdfdf8e58`. Every phase is exercised after
  an independently observed commit for both an ordinary primary database and a
  pre-existing non-current daily shard under an encoded filesystem path.

## R90-69 Definition

- **Goal:** reconcile the completed R90-66 through R90-68 storage-fault
  sequence and restore a bounded dependency-ready queue from current code,
  direct tests, public gap claims, fetched remote, and exact Vault evidence.
- **Risk:** inventing speculative fault work or treating broad historical gap
  prose as current authority can reopen completed boundaries or create false
  delivery commitments.
- **Required validation:** exact R90-66/R90-67/R90-68 feature and closure Git,
  task-state, remote, note/index/MOC, and stable-knowledge evidence; code/test
  comparison for current public remaining-gap claims; task-state JSON parsing;
  exact roadmap row/Definition coverage; complete fields for every new
  unfinished increment; documentation, knowledge, diff, and
  sensitive-information checks.
- **Stop condition:** stop without runtime implementation if the next bounded
  queue requires private/external input, a product or release decision,
  historical evidence rewrite, publication authority, or starting a later
  increment.
- **Selected plan:**
  [`task-20260804-storage-fault-delivery-audit.md`](task-20260804-storage-fault-delivery-audit.md),
  from clean fetched baseline
  `159fcf92122b387b3b80ecc5853150a6de1450d0`. The audit treats all six
  R90-66 through R90-68 commits and Vault notes as one delivered chain and
  does not implement R90-70.

## R90-70 Definition

- **Goal:** make the existing Go half of `make bench` exercise stable
  Aho-Corasick and full rule-engine matching hot paths instead of discovering
  no Go benchmarks.
- **Risk:** benchmark setup, Base64 preparation, mutable shared output, or
  unverified dead-code elimination can make reported time and allocations
  meaningless or flaky.
- **Required validation:** deterministic no-hit and multi-hit benchmark
  fixtures; construction/setup and correctness assertions outside timed
  regions; allocation reporting; explicit benchmark discovery/execution from
  the owning Go module and through `make bench`; focused rule tests, full
  native, documentation, and knowledge checks.
- **Stop condition:** stop if completion requires changing matcher semantics,
  optimizing production code, adding a benchmark dependency, external corpus
  access, host-independent thresholds, or a production throughput claim.
- **Selected plan:**
  [`task-20260804-go-rule-matching-benchmarks.md`](task-20260804-go-rule-matching-benchmarks.md),
  from clean fetched baseline
  `fffea8c7d030b84f836137fb22e94ae552a8e677`. The bounded fixture set covers
  Aho-Corasick plus immutable full-engine payload/IP/port matching without
  production-code or dependency changes.

## R90-71 Definition

- **Goal:** add bounded Go microbenchmarks for primary SQLite alert writes and
  filtered queries using the same durability and query paths as production.
- **Risk:** timing database creation, reusing duplicate event IDs, unbounded
  table growth, or weakening recovery durability can produce fast but invalid
  results.
- **Required validation:** deterministic single and batched write plus indexed
  filtered-query cases; unique event identity and bounded fixture cardinality;
  database setup/cleanup and correctness assertions outside timed regions;
  allocation reporting; explicit execution through the module and
  `make bench`; focused alert tests, full native, documentation, and knowledge
  checks.
- **Stop condition:** stop if completion requires disabling recovery-log
  durability, changing SQLite schema or behavior, persistent operator data,
  unbounded benchmark growth, host-independent thresholds, or a production
  throughput claim.
- **Selected plan:**
  [`task-20260804-go-alert-store-benchmarks.md`](task-20260804-go-alert-store-benchmarks.md),
  from clean fetched baseline
  `e853f8e22d10c98cc9363356272c6d847421514b`. The bounded cases keep real
  primary recovery durability enabled, clear write rows only outside timing,
  and seed one fixed indexed-query corpus through production `WriteBatch`.

## R90-72 Definition

- **Goal:** reconcile the complete local C/Go benchmark and repeat-pressure
  surface, then define the smallest defensible performance evidence or budget
  increment from comparable measurements.
- **Risk:** a single host result, stale June baseline, or synthetic repeat-pcap
  rate can be mislabeled as a portable regression threshold or production
  capacity guarantee.
- **Required validation:** exact R90-70/R90-71 feature and closure Git,
  task-state, remote, note/index/MOC, and stable-knowledge evidence; direct
  execution-path comparison for C, Go, metrics, and pressure tooling; current
  public performance-claim review; task-state JSON parsing; exact roadmap
  row/Definition coverage; documentation, knowledge, diff, and
  sensitive-information checks.
- **Stop condition:** stop without runtime or threshold changes if a portable
  budget requires external traffic, multiple comparable environments, a
  product/SLO decision, private data, historical rewrite, publication
  authority, or starting a later increment.
- **Audit record:**
  [`performance-evidence-audit-20260805.md`](../audit/performance-evidence-audit-20260805.md).
- **Selected plan:**
  [`task-20260805-performance-evidence-audit.md`](task-20260805-performance-evidence-audit.md),
  from clean fetched baseline
  `323be1f38fca456a0d17a7801e18bc50c5212075`. The documentation-only audit
  separates every measurement boundary and does not run a new benchmark,
  activate a threshold, or start R90-73.

## R90-73 Definition

- **Goal:** give the established C and Go microbenchmarks one versioned,
  machine-readable local evidence envelope without changing their measured
  behavior.
- **Risk:** permissive parsing can omit a benchmark or silently accept a
  partial run, while environment collection or raw output can leak sensitive
  host paths.
- **Required validation:** fixture-driven parser tests for every named C/Go
  case and malformed/partial output; exact clean/dirty Git state, OS/kernel/
  architecture/toolchain and command-parameter capture; default path
  redaction; one bounded direct complete-surface run; shell, Python, docs,
  knowledge, and full native checks.
- **Stop condition:** stop if completion requires changing benchmark/runtime
  semantics, collecting private host data, accepting a partial surface,
  applying a numeric threshold, external corpus input, or publication
  authority.

## R90-74 Definition

- **Goal:** establish a repeated observation-only baseline for the complete
  benchmark surface on one unchanged local environment and exact clean commit.
- **Risk:** cached, thermally unstable, background-loaded, or environment-drift
  samples can create a misleading variance summary, while an aggregate without
  raw samples prevents later review.
- **Required validation:** at least five uncached complete evidence captures;
  identical commit/tree, environment, toolchain, fixture, and command
  parameters; every raw sample retained; median, interquartile range, and
  variation summaries recomputed by a tested versioned API; full native,
  documentation, evidence, and knowledge checks.
- **Stop condition:** stop on environment drift, incomplete or ambiguous
  samples, excessive unexplained variance, sensitive metadata, pressure or
  corpus substitution, threshold activation, or publication authority.

## R90-75 Definition

- **Goal:** decide from matched evidence and explicit product/SLO scope whether
  NetSentry can support a portable, same-host-only, or observation-only
  regression policy.
- **Risk:** turning a single-host variance band into a universal gate can
  create false failures and false production-capacity claims.
- **Required validation:** completed R90-74 evidence; at least one independently
  provisioned comparable-environment evidence set using the same schema and
  exact benchmark commit; documented product/SLO scope; statistical and
  fixture comparability review; direct threshold-policy tests if any gate is
  proposed; documentation and knowledge checks.
- **Blocker evidence:** this trigger contains neither a comparable-environment
  evidence set nor authority to choose a product/SLO regression scope.
- **Unblock condition:** supply the matched evidence and explicitly choose the
  budget scope after R90-74 completes.
- **Stop condition:** remain blocked without both inputs; stop on corpus,
  commit, environment, metric, statistical, production-claim, or publication
  ambiguity.

## R90-76 Definition

- **Goal:** reconcile the completed local-tag boundary and recent delivery,
  then restore the empty dependency-ready queue from current repository
  evidence without runtime or external mutation.
- **Risk:** stale feature-only resume authority can obscure the fetched closure,
  while speculative queue filling can reopen completed work or cross product,
  publication, performance-budget, or compatibility boundaries.
- **Required validation:** exact R90-59a feature/closure/tag/remote/task/Vault
  evidence; dated phase-level history and current code/test gap review; all
  task-state JSON parsing; exact roadmap row/Definition coverage;
  documentation, knowledge, diff, staged-scope, and sensitive-information
  checks.
- **Stop condition:** stop if completion requires source/test behavior changes,
  private or external input, a product/compatibility decision, changelog or
  artifact approval, a performance threshold, tag/publication mutation,
  immutable-evidence rewrite, or starting a later increment.
- **Audit record:**
  [`delivery-plan-audit-20260809.md`](../audit/delivery-plan-audit-20260809.md).
- **Selected plan:**
  [`task-20260809-delivery-queue-audit.md`](task-20260809-delivery-queue-audit.md),
  from clean fetched baseline
  `5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc`.

## R90-77 Definition

- **Goal:** serialize the complete file-backed rule create, update, delete, and
  explicit reload transaction without blocking concurrent packet matching.
- **Risk:** an incomplete lock boundary can still lose an accepted mutation,
  deadlock a handler, or let disk and the immutable active snapshot diverge.
- **Required validation:** synchronized create/create, update/delete, and
  mutation/reload interleavings; successful-response, canonical-file, and
  active-snapshot agreement; validation/persistence failure preservation;
  focused repeated race, complete API/rule, full native, E2E, documentation,
  and knowledge checks.
- **Stop condition:** stop if completion requires changing public rule
  semantics/schema, cross-process file locking, migration policy, disabling hot
  reload, private data, or publication authority.
- **Selected plan:**
  [`task-20260809-rule-transaction-serialization.md`](task-20260809-rule-transaction-serialization.md),
  from clean fetched baseline
  `40798847be8e7bb9270b5c5d7675c27f7addf7b1`.

## R90-78 Definition

- **Goal:** make successful rule seed-file replacement durability-explicit and
  make every failure phase preservation-safe and observable to the API layer.
- **Risk:** a short write or missing file/directory sync can acknowledge an
  incomplete or crash-volatile mutation; an error after rename can create a
  disk/memory split if commit state is not classified.
- **Required validation:** direct short-write, chmod, file-sync, close, rename,
  and parent-directory-sync fault injection; byte-for-byte pre-rename
  preservation; exact temporary-file cleanup; explicit post-rename committed
  outcome; canonical reload and active-state agreement; focused repeated race,
  complete API/rule, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop on ambiguous post-rename state, platform semantics
  that require a product portability decision, rule-schema change, migration,
  external data, or publication authority.
- **Selected plan:**
  [`task-20260809-rule-file-durability.md`](task-20260809-rule-file-durability.md),
  from clean fetched baseline
  `4b5b199f37531e69c08cb7fa7b1d814f83047a37`.

## R90-79 Definition

- **Goal:** apply an independently tested durability and preservation contract
  to suppression-file replacement while retaining its serialized in-memory
  filter swap.
- **Risk:** reusing rule-file assumptions without direct suppression coverage
  can acknowledge crash-volatile state, expose a disk/filter split, or weaken
  the manager's existing mutation lock.
- **Required validation:** direct short-write, chmod, file-sync, close, rename,
  and parent-directory-sync faults; prior-file and temporary-file evidence;
  explicit post-rename outcome; active filter/file agreement; focused repeated
  race, complete alert/API, full native, E2E, documentation, and knowledge
  checks.
- **Stop condition:** stop if completion broadens suppression semantics, changes
  config schema, requires cross-process locking or migration policy, accesses
  private data, or needs publication authority.

## R90-80 Definition

- **Goal:** reconcile the completed management-plane concurrency/durability
  sequence and identify only evidence-supported follow-on work through the end
  of the active horizon.
- **Risk:** treating legacy schema support or broad protocol limitations as
  defects can silently choose compatibility or product policy; treating fault
  tests as production evidence can overstate reliability.
- **Required validation:** exact R90-77 through R90-79 code, direct tests,
  feature/closure, task-state, fetched remote, and Vault evidence; current API,
  architecture, development, and limitation review; task-state JSON and
  roadmap coverage; documentation, knowledge, diff, and sensitive-information
  checks.
- **Stop condition:** stop if completion requires choosing legacy-schema
  removal, migration or product scope, changing runtime/tests, external input,
  performance policy, or publication authority.
- **Audit record:**
  [`management-plane-persistence-audit-20260809.md`](../audit/management-plane-persistence-audit-20260809.md).
- **Selected plan:**
  [`task-20260809-management-plane-persistence-audit.md`](task-20260809-management-plane-persistence-audit.md),
  from clean fetched baseline
  `de949bda14a66a407391671f92f0c7b938fb2da5`.

## R90-81 Definition

- **Goal:** reconcile the completed R90-80 delivery closure and restore the
  empty local queue from recurring, directly verifiable validation evidence.
- **Risk:** treating isolated clean-rerun deviations as either a proven runtime
  defect or harmless noise can respectively broaden scope or preserve a weak
  release gate; speculative queue filling can cross product boundaries.
- **Required validation:** exact R90-80 feature/closure/remote/task/Vault
  evidence; dated phase-level history and recurring-deviation review; direct
  source-to-test boundary review; all task-state JSON parsing; exact roadmap
  row/Definition coverage; documentation, knowledge, diff, staged-scope, and
  sensitive-information checks.
- **Stop condition:** stop if completion requires source/test behavior changes,
  a runtime diagnosis unsupported by direct evidence, private/external input,
  product or compatibility policy, publication mutation, immutable-evidence
  rewrite, or starting R90-82.
- **Audit record:**
  [`post-management-plane-delivery-audit-20260809.md`](../audit/post-management-plane-delivery-audit-20260809.md).
- **Selected plan:**
  [`task-20260809-post-management-plane-delivery-audit.md`](task-20260809-post-management-plane-delivery-audit.md),
  from clean fetched baseline
  `49ae9eb95c6ff500e3c525bff30d7a13a43b6938`.

## R90-82 Definition

- **Goal:** make the receiver idle-timeout capacity-release regression observe
  the actual handler-slot boundary deterministically instead of inferring it
  through the process-wide latest-session snapshot.
- **Risk:** a test-only synchronization seam can accidentally change receiver
  behavior or mask a real timeout/capacity liveness defect; a fixed sleep or
  broad retry loop can preserve the same ambiguity under a different bound.
- **Required validation:** direct timeout-driven first-handler exit and slot
  release observation; replacement acceptance without shared-session polling;
  existing protocol-violation and ordinary disconnect capacity reuse; repeated
  uncached receiver race runs; full native, E2E, documentation, and knowledge
  checks.
- **Stop condition:** stop if deterministic proof requires a public runtime API,
  protocol/configuration change, relaxed timeout semantics, production traffic,
  external services, or publication authority.
- **Selected plan:**
  [`task-20260809-receiver-idle-capacity-evidence.md`](task-20260809-receiver-idle-capacity-evidence.md),
  from clean fetched baseline
  `9541d44db18b9c13e521b83be8aae79a9e5068be`.

## R90-83 Definition

- **Goal:** reconcile R90-82 delivery and restore only a directly evidenced
  local receiver-filesystem reliability queue after all prior local work
  completed.
- **Risk:** an audit can overstate unconditional pathname removal as a broader
  active-socket defect, or silently choose stale-socket and peer policy while
  attempting to restore local work.
- **Required validation:** exact R90-82 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 9 phase
  review; direct receiver startup/shutdown source and test mapping; complete
  unfinished-item fields; exact row/Definition multiset comparison; task-state
  JSON, documentation, knowledge, formatting, scope, and sensitive-information
  checks.
- **Stop condition:** stop if exact R90-82 evidence is missing or
  contradictory, the pathname gap cannot be bounded without active/stale
  socket or peer policy, validation remains ambiguous, or completion requires
  runtime/test changes, private/external input, product/performance policy,
  publication authority, or starting R90-84.

## R90-84 Definition

- **Goal:** keep receiver startup and shutdown within the filesystem identity
  the receiver is authorized to create and remove, without changing the
  existing stale-socket reclamation policy.
- **Risk:** a broad cleanup check can break ordinary restart, follow symlinks,
  delete operator data, or remove a pathname another process replaced after
  listener creation.
- **Required validation:** direct regular-file and symlink startup rejections
  with exact content/link preservation and no listener; ordinary absent-path
  startup compatibility; ordinary owned-socket shutdown cleanup; replacement
  regular-file and symlink preservation after the owned socket pathname is
  displaced; focused receiver race repetition, full native, E2E,
  documentation, and knowledge checks.
- **Stop condition:** stop if safe completion requires changing active/stale
  socket reclamation, dialing or authenticating an existing peer,
  cross-process locking, platform-specific ownership promises, operator data,
  or tag/publication authority.
- **Selected plan:**
  [`task-20260809-uds-pathname-preservation.md`](task-20260809-uds-pathname-preservation.md),
  from clean fetched baseline
  `5c4253d18283c80ec27b7c2c1f383616eac2a89e`.

## R90-85 Definition

- **Goal:** reconcile R90-84 delivery, repair its mutable roadmap chronology,
  and restore only directly evidenced local work after the ready queue emptied.
- **Risk:** correct commit facts in the wrong delivery order can mislead resume
  logic; speculative queue filling can turn a missing regression into a
  claimed defect or cross preserved socket/product boundaries.
- **Required validation:** exact R90-84 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 9 phase
  review; direct receiver startup/cancellation source and test mapping;
  chronological history review; complete unfinished-item fields; exact
  row/Definition multiset comparison; task-state JSON, documentation,
  knowledge, formatting, scope, and sensitive-information checks.
- **Stop condition:** stop if exact R90-84 evidence is missing or contradictory,
  chronology repair would alter immutable evidence, a follow-on needs product,
  compatibility, private/external, performance, or publication authority,
  validation is ambiguous, or completion would start runtime/test work.
- **Selected plan:**
  [`task-20260809-post-pathname-delivery-audit.md`](task-20260809-post-pathname-delivery-audit.md),
  from clean fetched baseline
  `79f6250de30c3128ecaec31e81ae19eecc9109d8`.

## R90-86 Definition

- **Goal:** make cancellation before receiver readiness fail closed before any
  configured-path mutation or listener creation.
- **Risk:** checking cancellation after stale-socket removal can destroy the
  prior identity before reporting cancellation, while changing later
  cancellation ordering can regress ordinary shutdown and path cleanup.
- **Required validation:** direct already-canceled absent-path and pre-existing
  Unix-socket identity-preservation regressions with `errors.Is` context
  sentinel checks; live absent/stale-socket startup and post-readiness active
  cancellation compatibility; repeated uncached receiver race runs; full
  native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if safe completion requires active/stale peer
  classification, cross-process path locking, changing post-readiness cleanup,
  protocol/configuration/public API changes, private data, or publication
  authority.
- **Selected plan:**
  [`task-20260809-receiver-pre-canceled-start.md`](task-20260809-receiver-pre-canceled-start.md),
  from clean fetched baseline
  `ab63ee3ef53fdb7a764ca0863dac36580d0318fa`.

## R90-87 Definition

- **Goal:** reconcile R90-86 delivery and restore only directly evidenced
  local work after the dependency-ready queue emptied.
- **Risk:** speculative queue filling can turn unconditional pathname removal
  into an unsupported liveness, trust, or production-defect claim.
- **Required validation:** exact R90-86 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 10 phase
  review; direct receiver existing-socket source, caller, test, and public-doc
  mapping; complete unfinished-item fields; exact row/Definition multiset
  comparison; task-state JSON, documentation, knowledge, formatting, scope,
  and sensitive-information checks.
- **Stop condition:** stop if exact R90-86 evidence is missing or contradictory,
  a follow-on needs peer trust/authentication, protocol, private/external,
  performance, or publication authority, validation is ambiguous, or
  completion would start runtime/test work.
- **Selected plan:**
  [`task-20260810-post-cancellation-delivery-audit.md`](task-20260810-post-cancellation-delivery-audit.md),
  from clean fetched baseline
  `6ea917e976d71432a4beb72967f73f2abf5c908b`.

## R90-88 Definition

- **Goal:** retain established stale-socket reclamation without unlinking a
  pathname currently owned by a connectable Unix listener.
- **Risk:** a liveness probe can perturb the existing peer, a pathname can be
  replaced between classification and removal, and treating reachability as
  authentication can overstate the trust boundary.
- **Required validation:** direct active-listener pathname-identity and
  continued-service preservation; stale-socket reclamation; identity-bound
  preservation when the pathname changes during classification; regular-file,
  symlink, already-canceled, ordinary startup, reconnect, cancellation, and
  owned-cleanup compatibility; repeated uncached receiver race, full native,
  E2E, documentation, and knowledge checks.
- **Stop condition:** stop if safe completion requires trusting or authenticating
  the peer, changing the hello/frame protocol or capture sender, cross-process
  locking beyond identity-bound pathname handling, private data, or
  publication authority.

## R90-89 Definition

- **Goal:** reconcile R90-88 delivery and restore only the directly evidenced
  cancellation-aware local startup work after the dependency-ready queue
  emptied.
- **Risk:** treating a fixed one-second probe bound as context cancellation can
  overstate shutdown responsiveness, while speculative queue filling can turn
  a missing direct regression into an unsupported production-defect claim.
- **Required validation:** exact R90-88 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 11 phase
  review; direct receiver startup, probe, cancellation, test, and public-doc
  mapping; complete unfinished-item fields; exact row/Definition multiset
  comparison; task-state JSON, documentation, knowledge, formatting, scope,
  and sensitive-information checks.
- **Stop condition:** stop if exact R90-88 evidence is missing or contradictory,
  a follow-on needs protocol/configuration/public API, private/external,
  performance, or publication authority, validation is ambiguous, or
  completion would start runtime/test work.
- **Selected plan:**
  [`task-20260811-post-listener-delivery-audit.md`](task-20260811-post-listener-delivery-audit.md),
  from clean fetched baseline
  `56d7d0b8005601299292b47d49bee7fc1e651753`.

## R90-90 Definition

- **Goal:** make cancellation during the bounded existing-socket liveness probe
  terminate receiver startup promptly before listener readiness.
- **Risk:** retaining `net.DialTimeout` can delay cancellation for the complete
  fixed probe bound, while changing probe error classification can accidentally
  reclaim an ambiguous or active pathname.
- **Required validation:** a direct receiver-local synchronized probe regression
  cancels after probe entry and checks prompt `errors.Is` context-sentinel
  return, original pathname identity, and absent receiver listener; direct
  active-listener, ambiguous-probe, replacement-identity, stale-reclamation,
  pre-canceled, ordinary startup, and post-readiness cancellation compatibility;
  repeated uncached receiver race, full native, E2E, documentation, and
  knowledge checks.
- **Stop condition:** stop if deterministic cancellation requires sleeps or a
  public test seam, if probe cancellation cannot preserve refusal-only stale
  classification and pathname identity, or if completion needs protocol,
  configuration, public API, private data, or publication authority.
- **Selected plan:**
  [`task-20260811-uds-probe-cancellation.md`](task-20260811-uds-probe-cancellation.md),
  from clean fetched baseline
  `22ba8ce639d79547875885f4ce107321273dd3b7`.

## R90-91 Definition

- **Goal:** reconcile R90-90 delivery and restore only the directly evidenced
  pathname-generation cleanup work after the dependency-ready queue emptied.
- **Risk:** device/inode reuse is a bounded filesystem race, not evidence of an
  observed production incident; speculative queue filling or a weak
  replacement test could overstate the gap.
- **Required validation:** exact R90-90 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 12 phase
  review; direct receiver ownership, cleanup, replacement-test, and public-doc
  mapping; complete unfinished-item fields; exact row/Definition multiset
  comparison; task-state JSON, documentation, knowledge, formatting, scope,
  and sensitive-information checks.
- **Stop condition:** stop if exact R90-90 evidence is missing or contradictory,
  a follow-on needs protocol/configuration/public API, private/external,
  performance, or publication authority, validation is ambiguous, or
  completion would start runtime/test work.
- **Selected plan:**
  [`task-20260812-post-probe-delivery-audit.md`](task-20260812-post-probe-delivery-audit.md),
  from clean fetched baseline
  `c0b1eb2dae8dd90eda745eacc87b0a6ece01a450`.

## R90-92 Definition

- **Goal:** make receiver shutdown preserve a pathname occupant that replaces
  its owned Unix socket even when the filesystem immediately reuses the
  original device/inode identity.
- **Risk:** device/inode equality alone can misclassify a new listener as the
  receiver's owned socket; an over-broad cleanup change could instead leak the
  ordinary owned pathname or follow a symlink.
- **Required validation:** a direct synchronized immediate-replacement
  regression proves inode reuse and replacement-listener service preservation;
  missing or changed non-following generation metadata fails closed; direct
  ordinary owned cleanup, regular-file and symlink replacement compatibility;
  repeated uncached receiver race, full native, E2E, documentation, and
  knowledge checks.
- **Stop condition:** stop if deterministic proof requires fixed sleeps,
  privileged filesystem control, a public test seam, following symlinks, or if
  completion needs protocol/configuration/public API, private data, or
  publication authority.
- **Selected plan:**
  [`task-20260812-uds-shutdown-generation-preservation.md`](task-20260812-uds-shutdown-generation-preservation.md),
  from clean fetched baseline
  `29c291a7dffcc37caf0375910e1ad1c6ef0a54a4`.

## R90-93 Definition

- **Goal:** reconcile R90-92 delivery and restore only the directly evidenced
  post-listen mode/ownership work after the dependency-ready queue emptied.
- **Risk:** a pathname-based metadata operation is a bounded local race, not
  evidence of an observed production incident; speculative queue filling or
  tests outside the listener-creation boundary could overstate the gap.
- **Required validation:** exact R90-92 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 12 phase
  review; direct listener creation, mode, ownership, replacement-test, Go
  contract, and public-doc mapping; complete unfinished-item fields; exact
  row/Definition multiset comparison; task-state JSON, documentation,
  knowledge, formatting, scope, and sensitive-information checks.
- **Stop condition:** stop if exact R90-92 evidence is missing or contradictory,
  a follow-on needs protocol/configuration/public API, private/external,
  performance, or publication authority, validation is ambiguous, or
  completion would start runtime/test work.
- **Selected plan:**
  [`task-20260812-post-generation-delivery-audit.md`](task-20260812-post-generation-delivery-audit.md),
  from clean fetched baseline
  `c59c3aca6a67b1975f178734d6b0f81a6bcab6b8`.

## R90-94 Definition

- **Goal:** make the configured UDS mode and captured ownership refer to the
  listener actually created by `Start`, even if its pathname is replaced
  before readiness.
- **Risk:** pathname-based `chmod` follows symlinks and can mutate a replacement
  target, while capturing a replacement socket as owned can publish a detached
  listener and later remove another service during shutdown.
- **Required validation:** direct synchronized post-listen replacement
  regressions preserve regular-file bytes/mode, symlink identity plus target
  mode, and replacement-listener mode/identity/service; created-listener-bound
  mode application and non-following pathname identity rejection; ordinary
  configured-mode, startup cancellation, shutdown cleanup, repeated uncached
  receiver race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if deterministic proof requires fixed sleeps,
  privileged filesystem control, an exported/public test seam, following
  symlinks, a new dependency or platform-specific unsafe implementation, or if
  completion needs protocol/configuration/public API, private data, or
  publication authority.

## R90-95 Definition

- **Goal:** reconcile R90-94 delivery and restore only the directly evidenced
  post-private-listener cancellation work after the dependency-ready queue
  emptied.
- **Risk:** cancellation can race many filesystem operations, but this audit
  must not promise interruptibility beyond the deterministic private-created,
  not-yet-published boundary or treat a transient transport deviation as an
  unresolved delivery failure.
- **Required validation:** exact R90-94 feature/closure Git, task-state,
  fetched-remote, and dual-Vault reconciliation; Jul 20 through Aug 14 phase
  review; direct startup cancellation, private-listener creation, publication,
  ownership, test, and public-doc mapping; complete unfinished-item fields;
  exact row/Definition multiset comparison; task-state JSON, documentation,
  knowledge, formatting, scope, and sensitive-information checks.
- **Stop condition:** stop if exact R90-94 evidence is missing or contradictory,
  a follow-on needs protocol/configuration/public API, private/external,
  performance, or publication authority, validation is ambiguous, or
  completion would start runtime/test work.
- **Selected plan:**
  [`task-20260814-post-listener-ownership-delivery-audit.md`](task-20260814-post-listener-ownership-delivery-audit.md),
  from clean fetched baseline
  `0dbf05acf1dcd233a9be6f76d54b947d77ff0290`.

## R90-96 Definition

- **Goal:** fail pre-readiness startup when the context is canceled after the
  private listener exists but before its pathname and ownership are published.
- **Risk:** returning success after cancellation can briefly publish a listener
  that shutdown immediately removes, while a weak regression could cancel at
  an already-covered boundary and leave the actual private-creation interval
  untested.
- **Required validation:** a direct synchronized regression cancels through the
  existing private-listener-created seam and requires an error matching the
  context sentinel, nil published receiver listener/ownership, absent public
  pathname, and no private staging artifacts; already-canceled startup,
  cancellation during the existing-socket probe, ordinary live startup,
  post-readiness shutdown, configured mode/ownership, repeated uncached
  receiver race, full native, E2E, documentation, and knowledge checks.
- **Stop condition:** stop if deterministic proof needs fixed sleeps, an
  exported/public seam, interruptible-filesystem guarantees, a dependency, or
  protocol/configuration/public API, private-data, performance-policy, or
  publication authority.

### R90-71 Validation Deviation

- **Observed:** The first uncached complete alert-package run hit the existing
  `TestStorePrimaryWriteActiveCancellationRetainsRecoveryLogForIdempotentRetry`
  five-second return boundary after active cancellation.
- **Impact:** Delivery is held pending focused uncached reproducibility review
  and clean alert-package plus complete native reruns. The new benchmark
  functions were excluded from the failing command, and no production storage
  source changed.
- **Reproduction and correction:** A 20-count uncached race command reproduced
  the timeout. The fixture used an equal 5-second SQLite busy timeout and outer
  return deadline; it now uses a 1-second driver timeout with the original
  5-second assertion while preserving active-boundary, context-cause, exact
  durable-state, and retry coverage. Production behavior is unchanged.
- **Resolution evidence:** The corrected exact regression passed 20 uncached
  race executions, followed by clean uncached complete alert-package runs both
  normally and under the race detector. Full repository validation remains the
  delivery boundary.

### R90-49 Validation Deviation

- **Observed:** The first complete native race suite hit the existing
  `TestStartIdleTimeoutReleasesConnectionCapacity` timing boundary when the
  replacement hello write received a broken pipe after the prior idle
  connection expired.
- **Impact:** Delivery remains blocked pending reproducibility assessment and a
  clean complete native rerun; the changed alert package passed and no receiver
  code was modified.
- **Resolution evidence:** Twenty uncached focused receiver race executions
  passed, followed by a clean complete native race-suite rerun. The timing
  event did not reproduce, so R90-49 validation may continue.

### R90-43 Validation Deviation

- **Observed:** The first full native race suite reached the new stored-row
  contract through `TestActiveLoadFullEngineShutdown`; its synthetic matcher
  emitted lowercase `tcp`, unlike the production rule engine's canonical
  `TCP`, so the alert API correctly returned a storage decode error.
- **Impact:** Delivery was held while the fixture mismatch was investigated; no
  shutdown orchestration or production protocol behavior changed.
- **Resolution:** The synthetic matcher now emits the production contract's
  canonical value. Twenty uncached focused shutdown race executions and the
  complete uncached native rerun pass.

### R90-24 Validation Deviation

- **Observed:** The first full native race suite hit the existing
  `TestStartIdleTimeoutReleasesConnectionCapacity` timing boundary: the
  replacement session was not observed before its bounded wait expired.
- **Impact:** Delivery was held while the unrelated result was ambiguous; no
  receiver behavior was changed.
- **Resolution:** Twenty uncached focused receiver race executions and the
  complete uncached native rerun pass. The timing event did not reproduce, so
  R90-24 validation may continue.

## Global Schedule-Window Waiver

- **Authorization:** On Jul 16, 2026, the user cancelled every roadmap planning
  window restriction.
- **Effect:** Earliest and latest dates remain visible only as historical
  forecasts. Dependency-ready increments may start immediately, and passing a
  forecast end date does not by itself block or defer work.
- **Unchanged controls:** Dependencies, evidence requirements, acceptance
  criteria, stop conditions, private-data boundaries, release decisions,
  tagging, and publication authorization remain fully enforced.
- **Current result:** The empty queue was refreshed on Jul 19 from verified
  Git, task-state, audit, code/test, release-boundary, and Vault evidence.
  R90-13 and R90-14 are complete. R90-14's sender-compatibility blocker was
  explicitly resolved on Jul 20 by authorizing hello on every C replacement
  connection; its fetched remote, post-fetch knowledge gate, and exact Vault
  range are verified. R90-15 is the next ready increment. No tag or public
  release is authorized. R90-15 completed early from the clean fetched
  `origin/main` baseline and verified R90-14 Vault evidence; its fetched remote,
  post-fetch knowledge gate, and exact Vault range are verified. R90-16
  completed early at `40b58c2c5160262efc42e3d8d7e5e588cd71fcc6`:
  syntactically valid recovery records now require the normalized writer's
  durable identity, timestamp/window/count, and network fields before replay;
  every direct semantic rejection preserves the complete log and persists no
  valid prefix. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact Vault note, full index, and MOC are
  verified. R90-17 completed early at
  `9a13283c3124ca270f39ca9ec63573e94283438c`: complete structural and
  semantic recovery validation now precedes directory creation and writable
  SQLite initialization, rejected input preserves missing and compatible
  databases, and valid replay uses the exact validated snapshot. Twenty
  focused race runs, the full native suite, E2E smoke, documentation, and
  knowledge checks passed; fetched `origin/main`, the post-fetch knowledge
  gate, and the exact Vault note, full index, and MOC are verified. The queue
  was refreshed on Jul 21 from the clean fetched baseline, completed task
  state, release boundaries, storage fault-injection gaps, and the existing
  Vault. R90-18 completed early at
  `cb2fd7d1889b33a01829226becb44260f1668651`: recovery records must now
  match the normalized writer's durable ID, first/last timestamps, aggregation
  window, and single-event count before SQLite initialization. All direct
  rejection cases preserve the full log and missing/existing database state.
  Twenty focused race runs, the full native suite, E2E smoke, documentation,
  and knowledge checks passed; fetched `origin/main`, the post-fetch knowledge
  gate, and the exact Vault note, full index, and MOC are verified. The queue
  was refreshed on Jul 22 from the clean fetched baseline, completed task
  state, release boundaries, the runtime recovery write path, and verified
  Vault evidence. R90-19 completed early at
  `9c93c8f82dfad07e17fcf57e4ba0818136b02710`: runtime writes now reject
  invalid existing recovery input before append or SQLite access, while valid
  pending records remain compatible. Twenty focused race runs, the full native
  suite, E2E smoke, documentation, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, and the exact Vault note, full
  index, and MOC are verified. The horizon was refreshed on Jul 22 through
  Oct 20 from the clean fetched baseline, completed task state, release
  boundaries, recovery reader/writer limits, and verified Vault evidence.
  R90-20 completed early at
  `1009187f1dae2cc1de8abde1738b159f3c4bd8e9`: writer batches are fully
  encoded and checked before append, reader capacity accepts records through
  4 MiB, and oversized output preserves the log and database. Twenty focused
  race runs, the full native suite, E2E smoke, documentation, and knowledge
  checks passed; fetched `origin/main`, the post-fetch knowledge gate, and the
  exact full-SHA Vault note, index, and MOC are verified. The queue was
  refreshed on Jul 22 from the clean fetched baseline, completed task state,
  release boundaries, SQLite write-critical schema constraints, and verified
  Vault evidence. R90-21 completed early at
  `352cf8fc96ab70a73a0b3f7e3da0cf4f32245160`: both write-critical tables
  now reject unknown mandatory columns without usable defaults before writable
  initialization, while compatible extensions remain writable. Twenty focused
  race runs, the full native suite, E2E smoke, documentation, and knowledge
  checks passed; fetched `origin/main`, the post-fetch knowledge gate, and the
  exact full-SHA Vault note, index, and MOC are verified. The queue was
  refreshed on Jul 23 from the clean fetched baseline, completed task state,
  release boundaries, SQLite uniqueness constraints, and verified Vault
  evidence. R90-22 completed early at
  `b62cbff41ec3f72adfa07030dcba17058a3e239e`: both write-critical tables
  now reject extra unique indexes lacking a binary-collated canonical write
  identity before writable initialization, while compatible index extensions
  remain writable. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 23 from the clean fetched
  baseline, completed task state, release boundaries, write-critical SQLite
  trigger metadata, and verified Vault evidence. R90-23 completed early at
  `c74982c13356cfa2733ed51bc890840b238d7cfe`: triggers attached to either
  write-critical table now fail before writable initialization, while
  unrelated operator-table triggers remain active and compatible. Twenty
  focused race runs, the full native suite, E2E smoke, documentation, and
  knowledge checks passed; fetched `origin/main`, the post-fetch knowledge
  gate, and the exact full-SHA Vault note, index, and MOC are verified. No
  later engineering increment is selected; refresh the rolling roadmap on the
  next `$netsentry-next` trigger. The queue was refreshed on Jul 23 from the
  clean fetched baseline, completed task state, release boundaries,
  write-critical generated-column metadata, and verified Vault
  evidence. R90-24 completed early at
  `4b342ae65b10279448b438e43b1947f1cfb282fc`: complete column metadata
  now exposes and rejects virtual or stored generated columns before writable
  initialization, while ordinary compatible extensions remain writable.
  Twenty focused generated-column race runs, twenty focused receiver reruns
  after one non-reproduced timing event, the clean full native rerun, E2E
  smoke, documentation, and knowledge checks passed; fetched `origin/main`,
  the post-fetch knowledge gate, and the exact full-SHA Vault note, index, and
  MOC are verified. The queue was refreshed on Jul 24 from the clean fetched
  baseline, completed task state, release boundaries, write-critical SQLite
  constraint metadata, and verified Vault evidence. R90-25 completed early at
  `1a4f565b1ef07b91a0c5ce80efc7cc78c382bb5b`: lexical schema inspection now
  rejects `CHECK` constraints on both write-critical tables before writable
  initialization without false positives from strings, comments, quoted
  identifiers, or identifier substrings. Twenty focused race runs, the full
  native suite, E2E smoke, documentation, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, and the exact full-SHA Vault
  note, index, and MOC are verified. The queue was refreshed on Jul 24 from the
  clean fetched baseline, completed task state, release boundaries, SQLite
  foreign-key metadata, and verified Vault evidence. R90-26 completed early at
  `0ddba61bde65fe1bb5ca9757bc87d06123409251`: read-only metadata inspection
  now rejects outgoing and incoming foreign-key relationships involving both
  write-critical tables, including case-variant and implicit-primary-key
  references. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, aggregation-index
  collation metadata, and verified Vault evidence. R90-27 completed early at
  `6a40a0aaf9b21d5d8a9ce08b7939d5b7b4ec8241`: the exact canonical
  aggregation uniqueness key now requires binary collation in column order,
  while compatible binary indexes preserve case-distinct identities. Twenty
  focused race runs, the full native suite, E2E smoke, documentation, and
  knowledge checks passed; fetched `origin/main`, the post-fetch knowledge
  gate, and the exact full-SHA Vault note, index, and MOC are verified. The
  queue was refreshed on Jul 25 from the clean fetched baseline, completed task
  state, release boundaries, SQLite identifier metadata semantics, and verified
  Vault evidence. R90-28 completed early at
  `41d4c94517d503175dc288fe763f1e860c55ed02`: required-column and
  unique-index key comparisons now follow SQLite's case-insensitive identifier
  semantics, while type, nullability, key order, collation, and every other
  write-safety check remain enforced. Twenty focused race runs, the full native
  suite, E2E smoke, documentation, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, and the exact full-SHA Vault
  note, index, and MOC are verified. The queue was refreshed on Jul 25 from the
  clean fetched baseline, completed task state, release boundaries,
  exact-filter query semantics, and verified Vault evidence. R90-29 completed
  early at `f12a454c95515dd92549e33e3c56d00449408d89`: rule, severity,
  source, and destination predicates now explicitly use binary comparison,
  preventing compatible custom column collations from broadening primary or
  historical results while protocol and MITRE matching remain
  case-insensitive. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, persisted-row numeric
  decoding, and verified Vault evidence. R90-30 completed early at
  `23679d6fbf6619315b6260e614dad62b2f3c2863`: primary and historical
  reads now reject ports outside `0..65535` before conversion and aggregate
  counts below one, while the read-only historical rejection preserves shard
  bytes. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, persisted severity
  decoding, and verified Vault evidence. R90-31 completed early at
  `856d1788c7f5abea0116b526ad8d7e2ebd5b9e11`: primary and historical
  reads now accept only the four public severity values; empty, case-variant,
  and unsupported text fails without substitution, while historical rejection
  preserves shard bytes. Twenty focused race runs, the full native suite, E2E
  smoke, documentation, and knowledge checks passed; fetched `origin/main`,
  the post-fetch knowledge gate, and the exact full-SHA Vault note, index, and
  MOC are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, persisted timestamp
  ordering, and verified Vault evidence. R90-32 completed early at
  `5d8eb60015c977e5f371846faff85ab015002615`: primary and historical
  reads now enforce `window_start <= first_seen <= last_seen` without assuming
  a historical window duration, while historical rejection preserves shard
  bytes. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, persisted aggregation
  identity behavior, and verified Vault evidence. R90-33 completed early at
  `824de0ee51fa5841d17021e797ba1f293c7aa128`: writer normalization and
  row decoding now share canonical aggregation-ID derivation, while primary
  and historical reads reject mismatches and historical rejection preserves
  shard bytes. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, persisted required-text
  behavior, and verified Vault evidence. R90-34 completed early at
  `9aaad8c837e89434f5eebd51f3397899df31027e`: primary and historical
  reads now reject blank required event, rule, protocol, and network identity
  while legitimate empty optional text remains compatible and historical
  rejection preserves shard bytes. Twenty focused race runs, the full native
  suite, E2E smoke, documentation, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, and the exact full-SHA Vault
  note, index, and MOC are verified. The queue was refreshed on Jul 25 from the
  clean fetched baseline, completed task state, release boundaries, the
  IPv4-only packet contract, and verified Vault evidence. R90-35 completed
  early at `ab90caa9b2148f9cfd706445bbd35c6646ac44a5`: UDS packet
  validation now rejects ordinary and IPv4-mapped IPv6 in either address
  position with one decode error and no queued packet, while valid IPv4 traffic
  remains compatible. Twenty focused race runs, the full native suite, E2E
  smoke, documentation, and knowledge checks passed; fetched `origin/main`,
  the post-fetch knowledge gate, and the exact full-SHA Vault note, index, and
  MOC are verified. The queue was refreshed on Jul 25 from the clean fetched
  baseline, completed task state, release boundaries, the strict IPv4 recovery
  boundary, and verified Vault evidence. R90-36 completed early at
  `a7eca65c9a8d327480821c22b1c42ae165f238b3`: startup replay and runtime
  preflight now reject malformed, ordinary IPv6, and IPv4-mapped IPv6 source or
  destination addresses before modifying the recovery log or missing/existing
  SQLite state, while valid IPv4 replay remains compatible. Twenty focused
  race runs, twenty affected-fixture race runs, the full native suite, E2E
  smoke, documentation, and knowledge checks passed; fetched `origin/main`,
  the post-fetch knowledge gate, and the exact full-SHA Vault note, index, and
  MOC are verified. The queue was refreshed on Jul 25 from the clean fetched
  reconciliation baseline, completed task state, release boundaries, the
  remaining stored-address contract, and verified Vault evidence. R90-37
  completed early at `fecf62d317d92e64a7816dacb337c6f444610086`:
  shared primary and historical row decoding now rejects malformed, ordinary
  IPv6, and IPv4-mapped IPv6 source or destination addresses before dependent
  aggregation-identity validation, while valid IPv4 rows remain compatible and
  historical rejection preserves shard bytes. Twenty focused race runs,
  twenty affected-fixture race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  reconciliation baseline, completed task state, release boundaries, recovery
  idempotency invariants, and verified Vault evidence. R90-38 completed early
  at `99b081bf941af5ab4a257900c3d08cfd339c5dc2`: startup and runtime
  recovery preflight now rejects an altered nonblank `event_id` before
  modifying the complete log or missing/existing SQLite state, while valid
  replay and duplicate-event idempotency remain compatible. Twenty focused
  race runs, the full native suite, E2E smoke, documentation, and knowledge
  checks passed; fetched `origin/main`, the post-fetch knowledge gate, and the
  exact full-SHA Vault note, index, and MOC are verified. The queue was
  refreshed on Jul 25 from the clean fetched reconciliation baseline, completed
  task state, release boundaries, the public severity enum, and verified Vault
  evidence. R90-39 completed early at
  `601a6dd5a47083a925727ef2b35b841566a72a24`: startup and runtime recovery
  preflight now accept exactly the four public severity values and reject
  empty, case-variant, or unsupported severity before modifying the complete
  log or missing/existing SQLite state, while stored-row validation remains
  compatible. Twenty focused race runs, the corrected collation fixture, the
  full native suite, E2E smoke, documentation, and knowledge checks passed;
  fetched `origin/main`, the post-fetch knowledge gate, and the exact full-SHA
  Vault note, index, and MOC are verified. The queue was refreshed on Jul 25
  from the clean fetched reconciliation baseline, completed task state, release
  boundaries, rule-loader and stored-row required-text behavior, and verified
  Vault evidence. R90-40 completed early at
  `75692ad0a9eb17a3672f60fbccf990204c50945f`: startup and runtime recovery
  preflight now reject missing, empty, or whitespace-only `rule_name` before
  modifying the complete log or missing/existing SQLite state, while padded
  nonblank names replay unchanged and stored-row required-text behavior remains
  compatible. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  reconciliation baseline, completed task state, release boundaries, rule
  engine MITRE emission, stored-row decoding, and verified Vault evidence.
  R90-41 completed early at
  `5e1425a6c81aac720a8a7743aee782bf2a5f61ed`: shared primary and historical
  row decoding now rejects every partial MITRE tactic/ID/name tuple and each
  whitespace-only member, while all-empty and fully populated historical
  values remain compatible without normalization or current-catalog
  revalidation. Twenty focused race runs, the full native suite, E2E smoke,
  documentation, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, and the exact full-SHA Vault note, index, and MOC
  are verified. The queue was refreshed on Jul 25 from the clean fetched
  reconciliation baseline, completed task state, release boundaries, the
  shared stored-row MITRE contract, recovery preflight behavior, and verified
  Vault evidence. R90-42 completed early at
  `4780f02688fabf75e89b954bc4f3f0982c0d1f6a`: shared startup and runtime
  recovery preflight now rejects every partial MITRE tactic/ID/name tuple and
  each whitespace-only member before modifying the complete log or
  missing/existing SQLite state, while all-empty and complete padded values
  replay unchanged without current-catalog revalidation. Twenty focused race
  runs, the full native suite, E2E smoke, documentation, and knowledge checks
  passed; fetched `origin/main`, the post-fetch knowledge gate, the exact
  full-SHA Vault note, index, MOC, and stable storage note are verified. No
  later engineering increment was selected. The horizon was refreshed on
  Jul 26 through Oct 24 from the clean fetched reconciliation baseline,
  completed task state, release boundaries, canonical protocol emission,
  stored-row decoding, and verified Vault evidence. R90-43 is selected as the
  highest-priority dependency-ready correctness increment. R90-43 completed
  early at `8b030b205f50768c9051354d19ec680b46ba876c`: rule emission and
  stored-row decoding now share canonical protocol names; noncanonical primary
  and historical values fail clearly, while historical rejection preserves
  shard bytes. Twenty focused alert-store race runs, twenty affected
  shutdown-fixture race runs, the full native suite, E2E smoke, documentation,
  and knowledge checks passed; fetched `origin/main`, the post-fetch knowledge
  gate, the exact full-SHA Vault note, index, MOC, and stable storage note are
  verified. No later engineering increment was selected. The horizon was
  refreshed on Jul 27 through Oct 25 from the clean fetched reconciliation
  baseline, completed task state, release boundaries, canonical protocol
  emission, stored-row decoding, recovery preflight behavior, and verified
  Vault evidence. R90-44 is selected as the highest-priority dependency-ready
  correctness increment. R90-44 completed early at
  `a87b2161bf65b726d827a805f21aa209bd71ed3b`: shared startup and runtime
  recovery preflight now rejects every planned noncanonical protocol form
  before modifying the complete log or missing/existing SQLite state, while
  canonical named and unknown protocol records remain compatible. Twenty
  focused alert-store race runs, the full native suite, E2E smoke,
  documentation, config, and knowledge checks passed; fetched `origin/main`,
  the post-fetch knowledge gate, and the exact full-SHA Vault note, index, MOC,
  and stable storage note are verified. No later engineering increment is
  selected. The horizon was refreshed from the clean fetched reconciliation
  baseline, completed task state, release boundaries, durable recovery
  validation, current `WriteBatch` ordering, and verified Vault evidence.
  R90-45 is selected as the highest-priority dependency-ready correctness
  increment. R90-45 completed early at
  `3990a1b228deddb3f43ef957af0eb102fbc170e4`: `WriteBatch` now
  validates the complete normalized current batch after existing-log preflight
  and before append, so a later invalid record cannot partially append a valid
  prefix, alter the pending log or SQLite, persist an alert, or degrade healthy
  storage. Twenty focused alert-store race runs, the full native suite, E2E
  smoke, documentation, config, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, and the exact full-SHA Vault
  note, index, MOC, and stable storage note are verified. No later engineering
  increment is selected; refresh the rolling roadmap on the next
  `$netsentry-next` trigger. The horizon was refreshed on Jul 28 through
  Oct 26 from the clean fetched reconciliation baseline, completed task state,
  release boundaries, SQLite text-ordering behavior, stored-row decoding, and
  verified Vault evidence. R90-46 is selected as the highest-priority
  dependency-ready correctness increment. R90-46 completed early at
  `9c6d574ba0f1f9766e9411b41d54b1ddeafb207b`: shared row decoding now
  rejects parseable timestamp encodings that differ from writer output before
  ordering or identity checks, while canonical rows remain compatible and
  historical rejection preserves shard bytes. Twenty focused alert-store race
  runs, the full native suite, E2E smoke, documentation, config, and knowledge
  checks passed; fetched `origin/main`, the post-fetch knowledge gate, the
  exact full-SHA Vault note, index, MOC, and stable storage note are verified.
  R90-47 completed early at
  `046f89673491b2bab78d6c21eedc067fa9c8584b`: UPSERT timestamp
  selection, primary and historical ordering/filtering, and retention pruning
  now share a fixed-width nanosecond key; the writable primary uses an optional
  expression index while unindexed historical shards remain read-only and
  correct. Twenty uncached focused alert-store race runs, the complete native
  race suite, E2E smoke, documentation, config, and knowledge checks passed;
  fetched `origin/main`, the post-fetch knowledge gate, the exact full-SHA
  Vault note, index, MOC, and stable storage note are verified. The horizon was
  refreshed on Jul 29 through Oct 27 from the clean fetched reconciliation
  baseline, completed task state, release boundaries, raw recovery timestamp
  decoding, and verified Vault evidence. R90-48 completed early at
  `6df3d8f45b2c581cf49c3b40e00198ba59dbc20e`: startup and runtime recovery
  preflight now reject alternate offset and fractional spellings for all four
  durable timestamps before representation-dependent semantic checks, while
  canonical writer output remains compatible. Twenty uncached focused
  alert-store race runs, the complete native suite, E2E smoke, documentation,
  configuration, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, exact full-SHA Vault note, index, MOC, and stable
  storage note are verified. The queue was refreshed from the clean fetched
  reconciliation baseline, completed task state, release boundaries, Go JSON
  member decoding, and verified Vault evidence. R90-49 completed early at
  `e015e9726bb5359bbd447b10d43953abda5b5149`: startup and runtime recovery
  preflight now reject exact duplicate top-level names and case-variant aliases
  before last-value decoding, while malformed diagnostics, single extensions,
  nested unknown values, and canonical writer output remain compatible.
  Twenty uncached focused alert-store race runs, the recorded receiver timing
  deviation and its twenty focused reruns, the clean complete native rerun,
  E2E smoke, documentation, configuration, and knowledge checks passed;
  fetched `origin/main`, the post-fetch knowledge gate, exact full-SHA Vault
  note, index, MOC, and stable storage note are verified. R90-50 completed
  early at `e49f2feea7fe3a3915998895f3c6e755b2ec3d17`: startup and runtime
  recovery preflight now reject unknown scalar and nested top-level members
  plus case-variant supported names, while duplicate and malformed diagnostics
  retain precedence and canonical writer output including optional
  `raw_payload` remains compatible. Twenty uncached focused alert-store race
  runs, the complete native race suite, E2E smoke, documentation,
  configuration, and knowledge checks passed; fetched `origin/main`, the
  post-fetch knowledge gate, exact full-SHA Vault note, index, MOC, and stable
  storage note are verified. No later engineering increment is selected;
  refresh the rolling roadmap on the next `$netsentry-next` trigger.
  Publication remains unauthorized. The horizon was refreshed on Jul 30
  through Oct 28 from the clean fetched reconciliation baseline, completed task
  state, release boundaries, recovery writer field presence, and verified Vault
  evidence. R90-51 is selected as the highest-priority dependency-ready
  correctness increment. R90-51 completed early at
  `4a27cece77f0f94b18982677c7562fac1e754b93`: startup and runtime recovery
  preflight now require all 19 non-`omitempty` writer fields before model
  decoding, while `raw_payload` remains optional and duplicate,
  unsupported-name, and malformed diagnostics retain precedence. Twenty
  uncached focused alert-store race runs, the complete native race suite, E2E
  smoke, documentation, configuration, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, exact full-SHA Vault note,
  index, MOC, and stable storage note are verified. No later engineering
  increment is selected; refresh the rolling roadmap on the next
  `$netsentry-next` trigger. Publication remains unauthorized. The horizon was
  refreshed on Jul 30 from the clean fetched reconciliation baseline,
  completed task state, release boundaries, top-level recovery value decoding,
  and verified Vault evidence. R90-52 is selected as the highest-priority
  dependency-ready correctness increment. R90-52 completed early at
  `f4985bb7fc3b6f50a5f90aa13d4d482cd712695c`: startup and runtime recovery
  preflight now reject `null` in every writer field and mismatched top-level
  JSON kinds before model decoding, while optional `raw_payload` remains
  compatible and structural diagnostics retain precedence. Twenty uncached
  focused alert-store race runs, the complete native race suite, E2E smoke,
  documentation, configuration, and knowledge checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, exact full-SHA Vault note,
  index, MOC, and stable storage note are verified. No later engineering
  increment is selected; refresh the rolling roadmap on the next
  `$netsentry-next` trigger. Publication remains unauthorized. R90-53 completed
  early at `4eb67e5cec8efdb969d4de4a2dbdea00b1da6ce0`: the dated audit reconciles
  241 July commits, 68 pre-audit task states, completed remote/Vault evidence,
  and the v0.1.1 hold boundary; the roadmap now has 62 entries and 62
  Definitions, an eight-step per-trigger audit, and planned or blocked work
  through Oct 28. Documentation, knowledge, JSON, definition coverage, skill
  structure, diff, and sensitive-information checks passed; fetched
  `origin/main`, the post-fetch knowledge gate, exact full-SHA Vault note,
  index, MOC, and stable testing/release note are verified. R90-54 is ready but
  was not started. Publication remains unauthorized. R90-54 completed early at
  `1e138805cdc133b87acd722f319fcc0cc624196f`: startup and runtime recovery
  preflight now reject exponent, fractional, and negative-sign spellings for
  both durable numeric fields before model decoding, while JSON-forbidden
  leading-zero forms retain malformed diagnostics and canonical writer output
  remains compatible. Direct preservation regressions, twenty uncached
  focused race runs, the complete native race suite, E2E smoke,
  documentation, configuration, knowledge, JSON, formatting, diff, and
  sensitive-information checks passed; fetched `origin/main`, the post-fetch
  knowledge gate, exact full-SHA Vault note, index, MOC, and stable storage
  note are verified. R90-55 is ready but was not started. Publication remains
  unauthorized. R90-55 completed early at
  `20161c20db271c5dbe9f5acc3f268eb5b8308494`: recovery name, presence, JSON
  kind, and integral-encoding validation now use one contract derived once
  from `model.Alert`, including the module writer's `omitempty` and `omitzero`
  behavior; ambiguous or unsupported future shapes fail contract construction.
  Twenty uncached focused race runs, the complete native race suite, E2E
  smoke, documentation, configuration, knowledge, JSON, formatting, diff, and
  sensitive-information checks passed after the final omission correction.
  Fetched `origin/main`, the post-fetch knowledge gate, exact full-SHA Vault
  note, index, MOC, idempotent replay, and stable storage note are verified.
  R90-56 is ready but was not started. Publication remains unauthorized.
  R90-56 was selected on Jul 30 from clean fetched baseline
  `cac88a4320dc820d9def98c8f5af775a0af5dfa2` after the per-trigger Git,
  recent-history, task-state, roadmap, and Vault audits passed. Direct
  fault-injection experiments confirmed that SQLite `mode=ro` alone can
  modify SHM evidence, while conditional `readonly_shm=1` preserves sidecars
  and active-WAL visibility. The bounded plan is persisted at
  `docs/plans/task-20260730-sqlite-sidecar-preflight.md`; no later increment
  or publication action is authorized.

## Global PCAP Release-Gate Waiver

- **Authorization:** On Jul 16, 2026, the user cancelled every PCAP package
  restriction.
- **Effect:** PCAP presence, source, evidence class, production derivation,
  sanitization/provenance/privacy approvals, sensitive-metadata review, packet
  count, byte size, digest, manifest, pressure/query evidence, and PCAP reviewer
  decisions cannot block release-gate acceptance.
- **Optional capability:** PCAP sanitizer, manifest, integrity, and pressure
  tooling remains available for diagnostics and engineering evidence.
- **Unchanged boundaries:** Raw PCAP bytes, private paths, credentials, and
  sensitive operator data remain prohibited from Git and the Vault. Fuzz, RC,
  supply-chain, final release decision, tagging, and publication controls remain
  enforced.

## Dependency and Priority Policy

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`;
`R90-04 → R90-04b → R90-05 → R90-06 → R90-07 → R90-08 → R90-09 → R90-10 → R90-11 → R90-12 → R90-13 → R90-14 → R90-15 → R90-16 → R90-17 → R90-18 → R90-19 → R90-20 → R90-21 → R90-22 → R90-23 → R90-24 → R90-25 → R90-26 → R90-27 → R90-28 → R90-29 → R90-30 → R90-31 → R90-32 → R90-33 → R90-34 → R90-35 → R90-36 → R90-37 → R90-38 → R90-39 → R90-40 → R90-41 → R90-42 → R90-43 → R90-44 → R90-45 → R90-46 → R90-47 → R90-48 → R90-49 → R90-50 → R90-51 → R90-52 → R90-53 → R90-54 → R90-55 → R90-56 → R90-57 → R90-60 → R90-61 → R90-62 → R90-63 → R90-64 → R90-65 → R90-66 → R90-67 → R90-68 → R90-69 → R90-70 → R90-71 → R90-72 → R90-73 → R90-74 → R90-75`;
`(R90-59a + R90-74) → R90-76 → R90-77 → R90-78 → R90-79 → R90-80 → R90-81 → R90-82`;
`R90-56 → R90-58 → R90-59a → R90-59`. R90-75 is blocked on
comparable-environment evidence plus explicit product/SLO budget scope but is
not a dependency for unrelated future work. R90-59 is blocked on explicit
remote tag-push, GitHub Release, and GHCR authorization after R90-59a. R90-04a
is an evidence-independent quality
increment and does not satisfy any R90-04 dependency. The R90-04 and R90-05
PCAP exceptions remain immutable historical delivery evidence. The later global
PCAP waiver supersedes their restrictions for current and future release-gate
decisions.

## R90-04 Scoped Evidence Exception

- **Authority and scope:** `docs/audit/release_exception_r9004.yaml` authorizes an R90-04-only alternative to internal production-derived PCAP evidence.
- **Allowed evidence:** anonymized, publicly released, real network traffic only. Synthetic or generated traffic is permanently prohibited.
- **Required controls:** approve dedicated privacy review, provenance validation, sanitization review, and sensitive-metadata screening before corpus-pressure validation or official-evidence use.
- **Boundary:** this exception expires when R90-04 completes and does not amend R90-05, R90-06, or future increment requirements.

## R90-05 Authorized Schedule Deviation

- **Authorization:** On Jul 16, 2026, the user explicitly waived only the Sep 12
  scheduled start constraint and authorized R90-05 to begin immediately.
- **Later policy change:** On Jul 16, the user separately approved the exact
  synthetic corpus recorded in `docs/audit/release_exception_r9005.yaml` as an
  R90-05-only substitute for production-derived PCAP evidence.
- **Impact:** Work begins 58 days early. R90-06, tagging, release approval, and
  publication remain outside this authorization.
- **Stop condition:** Stop if completion requires private corpus access,
  interactive privileged validation, a new evidence exception, release
  approval, tagging, or publication.

## R90-05 Corpus Handoff Timeline — Superseded

- **External prerequisite:** Release/privacy owners must provide an approved
  sanitized production-derived PCAP corpus together with complete provenance,
  sanitization, privacy-review, packet-count, and SHA-256 manifest inputs.
- **Alignment checkpoint:** Obtain the responsible owner and committed delivery
  date by Jul 20, 2026. Target corpus approval and handoff no later than Sep 25,
  leaving the final week of the R90-05 window for validation and acceptance.
- **Validation turnaround:** Within one business day of handoff, generate and
  verify the path-redacted manifest, run corpus pressure and the full Docker RC,
  and prepare the sanitized v0.1.1 evidence record. Complete release-gate review
  and final acceptance by Oct 2.
- **Schedule risk:** If the owner or delivery date is not confirmed by Jul 20,
  or the approved corpus is not available by Sep 25, record R90-05 and R90-06
  schedule impact immediately; do not substitute synthetic, public, or
  unreviewed traffic.
- **Supersession:** The Jul 16 R90-05-only synthetic exception satisfied this
  external handoff dependency for the approved digest only. Preserve these
  dates as historical planning evidence; do not apply the exception to R90-06.

## Current Checkpoint

R90-58 completed early at
`6ed01a710ff17d11e196a5fb8685401407376395` for clean fetched candidate
`78cd78574e03c8f73ff68248eed2c409d6bca406`. The trigger audit verified the
complete R90-56 feature and closure chain, exact Vault evidence, all 249 July
commits, 62/62 row-to-Definition coverage, and complete future-item planning.
An isolated detached worktree at the exact candidate passed the full Docker RC
with 78.3% Go coverage, 5,000-iteration ASan parser fuzz smoke, and E2E smoke;
the pinned supply-chain audit fetched and matched all nine assets and reported
zero reachable Go vulnerabilities; the v0.1.1 release gate passed. The fresh
`linux/amd64` archive is 9,760,241 bytes with SHA-256
`c68e09df46d24307c9a0d405a2724573f3382813a8b2611bdb5f3b7d8b068568`.
The first combined sequence stopped after RC because pinned local tools were
absent; after installing the exact temporary tool versions, the entire sequence
was rerun successfully. The feature commit is pushed and fetched
`origin/main` equals it; the post-fetch knowledge gate, exact full-SHA Vault
note, full index, MOC link, idempotent replay, and stable release/testing note
are verified. At that closeout, R90-57 remained blocked on its product decision
and R90-59 remained blocked on explicit authorization for exact version
`v0.1.1` and candidate
`78cd78574e03c8f73ff68248eed2c409d6bca406`; tagging and publication remained
unauthorized. On Aug 1, the user reaffirmed the global schedule waiver
and cancelled additional prerequisite review as an eligibility gate. R90-57 is
therefore selected from clean fetched baseline
`46bbf8a0535c30e707b7dfbaefee9cab27a81d84` using the fail-closed default of
operator-triggered recovery with no background retry or automatic evidence
cleanup. R90-57 completed at
`6b53430e333118b5fcebeb77f6c59302a58d4382`: the state machine covers healthy,
degraded, emergency, recovering, and closed; one owner and an exclusive
lifecycle barrier protect handle replacement; read-only preflight precedes any
writable boundary; cancellation, partial replay, daily shards, empty-log proof,
and evidence preservation have direct implementation test requirements. The
feature commit is pushed and fetched `origin/main` equals it; documentation,
evidence, knowledge, JSON, definition, diff, sensitive-information, exact Vault
note/index/MOC, and stable SQLite-storage knowledge checks passed. Runtime/API
implementation was not started. R90-59 remains blocked on explicit publication
action authority. The next trigger found that the forward queue omitted the
runtime increment explicitly deferred by R90-57. R90-60 is therefore added and
selected from clean fetched baseline
`59904b79424f80d760d3a9aac9c9617ef1e975cb`; its bounded implementation plan
and task state were persisted before runtime changes. The lifecycle gate,
single recovery owner, preservation-safe preflight, idempotent replay or empty
log write probe, mandatory-auth API, bounded health/audit surface, shutdown
cancellation, and direct regressions are now implemented. Focused race tests,
twenty uncached repetitions, full native tests, E2E smoke, documentation,
evidence, knowledge, JSON, definition, and diff checks pass; feature delivery
and exact remote/Vault evidence passed. The feature completed at
`a4a4adf662e1accf11528dc2440000426fe5fa28`; it was pushed without force,
fetched equal to `origin/main`, and passed the post-fetch knowledge gate. Exact
range
`59904b79424f80d760d3a9aac9c9617ef1e975cb..a4a4adf662e1accf11528dc2440000426fe5fa28`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and updated stable SQLite-storage knowledge are verified.
R90-59 remains blocked on explicit publication action authority; no later
increment is selected. The Aug 2 trigger verified the clean fetched R90-60
closure and exact Vault evidence, then found that R90-59 was the only
unfinished row and remained externally blocked. R90-61 is selected as the
smallest safe documentation-only queue unblocker. Its code/test audit records
that committed-prefix multi-shard recovery retry is promised by architecture
and development guidance but lacks a direct later-shard failure or
active-replay cancellation regression. R90-62 is planned as the next
correctness increment, followed by the documented C formatter and sustained
fuzz gaps. R90-59 remains blocked; no runtime, tag, release, registry, or
workflow action is authorized by this audit. R90-61 completed at
`99963311d80a279e532cf8b7d43a9945ada70b46`: its four-path documentation
commit was pushed without force, fetched equal to `origin/main`, and passed the
post-fetch knowledge gate. Exact range
`3f3acbbb0b12046f1db7a7892c818a6d8f732649..99963311d80a279e532cf8b7d43a9945ada70b46`
was synchronized idempotently to the single local Vault; the iteration note,
full index, MOC link, and updated stable testing/release knowledge are
verified. R90-62 is ready but was not started. R90-59 remains blocked.
The next trigger fetched and verified the R90-61 closure plus exact Vault
evidence, then selected R90-62 from clean baseline
`89806508802fd8d8165f9606995d19bba0ef6da0`. Daily-shard recovery now sorts
its serial replay paths, and direct real-SQLite regressions exercise a locked
later shard plus active context cancellation after an independently observed
earlier commit. Both retain the complete log and sticky emergency state before
an explicit retry proves one event and aggregate count one per input. Twenty
uncached focused race executions and the complete alert-package race run pass;
the complete native race suite, E2E smoke, documentation, evidence, and
knowledge gates also pass. R90-63 and R90-59 were not started.
R90-62 completed early at
`981cb1e3a0041301f42629522cff844e04764c6f`: the eight-path feature commit was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate. Exact range
`89806508802fd8d8165f9606995d19bba0ef6da0..981cb1e3a0041301f42629522cff844e04764c6f`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and updated stable SQLite/testing knowledge are verified.
R90-63 is ready but was not started. R90-59 remains blocked.
The next trigger fetched and verified the R90-62 closure plus exact Vault
evidence, then selected R90-63 from clean baseline
`f5dc37e48513de31633aaa7a812e619a3d171e90`. A dedicated C ASan boundary now
derives bounded packet, heartbeat, and hello formatter inputs; structured seeds
and deterministic mutations cover escaping, payload, integer, exact-fit, and
undersized-buffer behavior with canary protection. Representative output also
passes independent strict JSONL decoding and frame-shape checks. Default and
100,000-mutation focused runs, ordinary and ASan C tests, full native race,
E2E, shell, Python, documentation, evidence, and knowledge gates pass. R90-64
and R90-59 were not started. R90-63 completed early at
`357455a22f62b4d85c16c431fde70320d27c28a9`: the fourteen-path feature commit
was pushed without force, fetched equal to `origin/main`, and passed the
post-fetch knowledge gate. Exact range
`f5dc37e48513de31633aaa7a812e619a3d171e90..357455a22f62b4d85c16c431fde70320d27c28a9`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and updated stable testing/ASan-fuzz knowledge are
verified. R90-64 is ready but was not started. R90-59 remains blocked.
The Aug 3 trigger fetched and verified the R90-63 closure plus exact Vault
evidence, audited all 78 prior task-state files and 67 roadmap definitions, and
selected R90-64 from clean baseline
`33bc37d9ff71932d6e4ea49cf414f3ed0008415a`. `make fuzz-sustained` now forces
fresh ASan parser/formatter builds, runs both at one explicit budget, and uses
a versioned validator for exact harness/iteration/status, sanitizer, corpus
redaction, and evidence-class fields. The accepted no-corpus local synthetic
run passed 1,000,000 mutations per harness with zero sanitizer findings in
118.317 seconds; a separate path-bearing fixture proved JSON/Markdown
redaction. Serial ASan C tests, an explicitly clean ordinary C rebuild, every
Go package under uncached race, shell, Python, documentation, evidence, and
knowledge gates passed. The feature completed early at
`73ab39ef88245b01b3d3418f0d9aeb0f6db1d546`; it was pushed without force,
fetched equal to `origin/main`, and passed the post-fetch knowledge gate. Exact
range
`33bc37d9ff71932d6e4ea49cf414f3ed0008415a..73ab39ef88245b01b3d3418f0d9aeb0f6db1d546`
was synchronized idempotently to the single local Vault; its note, full index,
MOC link, and stable fuzz/testing knowledge are verified. R90-65 is ready but
was not started. R90-59 remains blocked on exact publication authority.
The next Aug 3 trigger fetched and verified the R90-64 closure at
`23983e1ac696b923a4595e7b97f0e7e1d935dc97` plus both exact Vault notes,
index entries, MOC links, and stable fuzz/testing updates. The dated R90-65
audit reviewed 139 commits across three phases and reconciled the public gap
claims with code, direct tests, R90-04 traffic evidence, and the R90-64 local
synthetic baseline. Larger reviewed fuzz corpora and more diverse alert-bearing
traffic remain external-input diagnostics, not ready local work or R90-59
prerequisites. The broad local storage-fault claim is narrowed to R90-66
through R90-68: primary write interruption after durable log append,
recovery-log append lifecycle faults, and post-commit log-clearing faults.
R90-65 completed early at
`84e83a17fa0560a8a0cc76e34701a730696c5f44`: its six-path documentation
feature was pushed without force, fetched equal to `origin/main`, and passed
the post-fetch knowledge gate. Exact range
`23983e1ac696b923a4595e7b97f0e7e1d935dc97..84e83a17fa0560a8a0cc76e34701a730696c5f44`
was synchronized idempotently to the single local Vault; the iteration note,
full index, MOC link, and corrected stable fuzz/testing authority are verified.
Historical iteration notes remain unchanged. R90-66 is ready but was not
started, and R90-59 remains blocked on exact publication authority.
The next Aug 3 trigger fetched and verified the R90-65 closure at
`667cedc72dec9ce58fc7c12aff3be2d37e9ab835` plus exact Vault notes, index,
MOC, and current stable authority. All 80 task states and 71 roadmap
Definitions reconcile. R90-66 is selected as the sole dependency-ready item:
it adds direct ordinary-primary contention and active-cancellation evidence
after the durable recovery append and before commit, using a pre-opened
read-only observer and observable SQLite connection readiness. R90-67, R90-68,
and R90-59 were not started.
The direct contention and active-cancellation cases now pass: each uses a real
independent write reservation, retains the exact recovery log, proves zero
event/aggregate rows before retry through one pre-opened observer, and proves
one event plus aggregate count one after one retry. The active case exposed and
corrected a lost `context.Canceled` classification while preserving the SQLite
interruption diagnostic. Twenty final uncached focused race runs, the complete
alert package, full native tests, E2E, documentation, and knowledge checks
pass. R90-67, R90-68, and R90-59 remain unstarted.
R90-66 completed early at
`260d53d6b5804ca37dc83b083486d429a5e9c983`: the exact eight-path feature was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate. Exact range
`667cedc72dec9ce58fc7c12aff3be2d37e9ab835..260d53d6b5804ca37dc83b083486d429a5e9c983`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current SQLite/testing authority are verified.
R90-67 is ready but was not started, R90-68 remains planned, and R90-59 remains
blocked on exact publication authority.
The Aug 4 trigger fetched and verified the R90-66 docs-only closure at
`2f62acf9025969a50dd0295f3881ce7cd2784ec6`, both exact Vault iteration notes,
the full index, MOC links, and current stable SQLite/testing authority. All 81
prior task states parse and all 71 roadmap rows match one Definition. R90-67 is
selected as the sole dependency-ready increment with a store-local append-file
seam and direct open, short-write, sync, and close preservation evidence.
R90-68 and R90-59 were not started.
The store-local seam, explicit short-write rejection, and direct four-phase
regression are implemented. Twenty uncached focused race executions, the
complete alert package race suite, full native tests, E2E smoke, documentation,
knowledge, JSON, definition, formatting, diff, and sensitive-information checks
pass. R90-67 remains in progress until feature push, fetch verification, and
exact-range Vault synchronization complete.
R90-67 completed early at
`1a9732514d4cf061a52821f9b487fa10aebbf35e`: its exact eight-path feature was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate. Exact range
`2f62acf9025969a50dd0295f3881ce7cd2784ec6..1a9732514d4cf061a52821f9b487fa10aebbf35e`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current stable SQLite/testing/MOC authority are
verified. R90-68 is ready but was not started, and R90-59 remains blocked on
exact publication authority.
The next Aug 4 trigger fetched and verified the R90-67 docs-only closure at
`cac3178512a84356364f82261f2b7dffdfdf8e58`, both exact Vault notes, the full
index, MOC links, and current stable SQLite/testing/MOC authority. All 82 prior
task states parse and all 71 roadmap rows match one Definition. R90-68 is
selected as the sole dependency-ready increment with direct post-commit
open/truncate, sync, and close evidence across primary and encoded daily-shard
paths. R90-59 was not started.
The clear path now syncs the truncated file before close, and its per-Store
fault seam reaches all three phases. Six direct race cases prove independently
observed post-commit cardinality, exact retained versus already-cleared log
state, sticky phase-specific emergency, and one healthy explicit recovery for
ordinary primary plus encoded historical-shard paths. Full validation remained
the delivery boundary at that checkpoint; R90-59 was not started.
Twenty uncached focused race executions, the complete alert package race suite,
full native tests, E2E smoke, documentation, knowledge, JSON, definition,
formatting, diff, and sensitive-information checks pass. R90-68 remains in
progress until feature push, fetch verification, and exact-range Vault
synchronization complete.
R90-68 completed early at
`574dfd9e43959656e33373db82cb88dc2b3184f2`: its exact eight-path feature was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate. Exact range
`cac3178512a84356364f82261f2b7dffdfdf8e58..574dfd9e43959656e33373db82cb88dc2b3184f2`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current stable SQLite/testing/MOC authority are
verified. R90-69 is ready but was not started, and R90-59 remains blocked on
exact publication authority.
The next Aug 4 trigger fetched and verified the R90-68 docs-only closure at
`159fcf92122b387b3b80ecc5853150a6de1450d0`, all six R90-66 through R90-68
Vault notes, the full index, MOC links, and current stable SQLite/testing/MOC
authority. All 83 prior task states parse and all 72 roadmap rows match one
Definition. The direct test bodies match every promised storage-fault boundary,
so the completed sequence is not reopened. R90-69 is selected as the sole
dependency-ready documentation audit; R90-59 remains blocked.
The dated audit reviews 147 commits across three phases and reconciles current
public gaps with code, tests, and Make targets. External fuzz/traffic remains
input-dependent, product-scale protocol and migration work remains outside
this trigger, and the concrete local gap is that `make bench` invokes Go
benchmark discovery while the module contains no `Benchmark*` function.
R90-70 through R90-72 split matcher benchmarks, SQLite benchmarks, and a later
performance evidence/budget audit through Oct 31. None was started.
All 84 task-state JSON files parse, all 75 roadmap rows match one Definition,
and documentation, knowledge, formatting, diff, staged-scope, and
sensitive-information checks pass. R90-69 remains in progress until feature
push, fetch verification, and exact-range Vault synchronization complete.
R90-69 completed early at
`1a612273dd49a216710441dc2eae9e0e2b4d16f7`: its exact six-path documentation
feature was pushed without force, fetched equal to `origin/main`, and passed
the post-fetch knowledge gate. Exact range
`159fcf92122b387b3b80ecc5853150a6de1450d0..1a612273dd49a216710441dc2eae9e0e2b4d16f7`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current stable Makefile/testing/MOC authority are
verified. R90-70 is ready but was not started, and R90-59 remains blocked on
exact publication authority.
The next Aug 4 trigger fetched and verified the R90-69 docs-only closure at
`fffea8c7d030b84f836137fb22e94ae552a8e677`, both exact Vault notes, the full
index, MOC links, and current stable Makefile/testing/MOC authority. All 84
prior task states parse and all 75 roadmap rows match one Definition. R90-70 is
selected as the sole dependency-ready increment with deterministic
Aho-Corasick and immutable full-engine no-hit/multi-hit matching benchmarks;
R90-71/R90-72 were not started, and R90-59 remains blocked.
The two benchmark families now execute with fixture construction, Base64
preparation, correctness assertions, and diagnostics outside timed regions.
Each case reports allocations and bytes, retains its local result against
dead-code elimination without a shared mutable sink, and the multi-hit engine
fixture traverses payload, IP, and port rules. Focused rule tests and bounded
direct benchmark execution pass; full validation remains the delivery
boundary.
Two exact root benchmark reruns then exposed the same existing storage
cancellation test timeout because the all-package Go benchmark command also
ran ordinary tests concurrently with long benchmark packages. The exact test
passed alone and across 20 uncached race repetitions. The dedicated benchmark
command now uses `-run '^$'`; one final root run passes all C cases and all four
ten-second Go cases, while `make test` separately passes the complete native
race suite. E2E, documentation, knowledge, JSON, definition, formatting,
scope, and sensitive-information checks pass. R90-70 remains in progress until
feature push, fetch verification, and exact-range Vault synchronization
complete.
R90-70 completed early at
`388487da7205e98dd257ee54a1428673141c7457`: its exact ten-path benchmark
feature was pushed without force, fetched equal to `origin/main`, and passed
the post-fetch knowledge gate. Exact range
`fffea8c7d030b84f836137fb22e94ae552a8e677..388487da7205e98dd257ee54a1428673141c7457`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current stable Makefile/Aho-Corasick/rule-engine/
testing/MOC authority are verified. The helper's first attempt used the stale
documented default Vault path and failed before writing; the same exact range
succeeded with the sole discovered Vault supplied explicitly.
R90-71 is ready but was not started, R90-72 remains planned, and R90-59 remains
blocked on exact publication authority.
The next Aug 4 trigger fetched and verified the R90-70 docs-only closure at
`e853f8e22d10c98cc9363356272c6d847421514b`, both exact Vault notes, the full
index, MOC links, and current stable benchmark/testing authority. All 85 prior
task states parse and all 75 roadmap rows match one Definition. R90-71 is
selected as the sole dependency-ready increment with durable single/batched
primary writes and fixed-cardinality indexed filtered queries; R90-72 and
R90-59 were not started.
The four benchmark cases now execute with unique event identity, real recovery
durability, bounded row cleanup, a fixed 512-row production-seeded query
fixture, and direct rule/time index assertions. The equal-deadline cancellation
deviation was corrected test-only and passed 20 uncached race executions plus
clean normal/race alert-package runs. The final root benchmark exposes 1/32
alerts per write operation; full native race, E2E, documentation, knowledge,
JSON, definition, formatting, scope, and sensitive-information checks pass.
R90-71 remains in progress until feature push, fetched verification, and exact
Vault synchronization complete. R90-72 and R90-59 remain unstarted.
R90-71 completed early at
`9f29bf32cc3bbc446d03bd2185900c3dae4a84ef`: its exact nine-path feature was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate. Exact range
`e853f8e22d10c98cc9363356272c6d847421514b..9f29bf32cc3bbc446d03bd2185900c3dae4a84ef`
was synchronized idempotently to the single local Vault; its iteration note,
full index, MOC link, and current SQLite, Makefile, testing, and MOC authority
are verified. R90-72 is ready but was not started, and R90-59 remains blocked
on exact publication authority.
The Aug 5 trigger fetched and verified the R90-71 docs-only closure at
`323be1f38fca456a0d17a7801e18bc50c5212075`, both exact R90-70/R90-71
feature and closure pairs, all four Vault notes/index rows/MOC links, and
current stable benchmark authority. All 86 prior task states parse and all 75
roadmap rows match one Definition. R90-72 is selected as the sole
dependency-ready documentation audit; R90-59 remains blocked.
The audit reviews 153 commits across three phases and reconciles every C/Go
microbenchmark, repeat-pcap and corpus-pressure path, runtime metric, public
claim, and checked-in/local-only evidence boundary. Current numeric Go output
is not versioned, the complete surface has no repeated matched-host sample set,
and historical synthetic pressure varies from 552 to 1,402 pps, so no portable
or 10% regression threshold is supportable. R90-73 through R90-75 now separate
versioned evidence capture, a repeated single-host observation baseline, and a
budget decision blocked on comparable-environment evidence plus explicit
product/SLO scope. None was started; R90-72 remains in progress until its
documentation feature is pushed, fetched, and synchronized.
R90-72 completed early at
`13b259f3779840a8a410803dfd209f19bbb71649`: its exact eight-path
documentation feature was pushed without force, fetched equal to
`origin/main`, and passed the complete rerun of the post-fetch 33-test
knowledge gate. The first post-fetch command used an incorrectly inferred full
SHA and stopped before the gate; the complete sequence was rerun with
`git rev-parse HEAD` as authority. Exact range
`323be1f38fca456a0d17a7801e18bc50c5212075..13b259f3779840a8a410803dfd209f19bbb71649`
was synchronized idempotently to the single local Vault. Its iteration note,
full index, MOC link, and reconciled stable MOC/Makefile/testing authority are
verified. R90-73 is ready but was not started; R90-75 and R90-59 remain blocked
on their recorded external authority conditions.
The Aug 6 trigger fetched and verified the R90-72 docs-only closure at
`b20845a8b7b4584e9cfa49aadc5ee663c17a2fe2`, both exact R90-72 Vault notes,
the full index, MOC links, and current stable performance authority. All 87
prior task states parse and all 78 roadmap rows match one Definition. R90-73
is selected as the sole highest-priority dependency-ready increment; R90-74
remains planned, while R90-75 and R90-59 remain blocked.
The versioned capture command now retains exact Git/tree and environment/
toolchain context, redacted raw output, and strictly parsed metrics for all six
C and eight Go cases without changing their timed boundaries. Fourteen focused
tests cover complete/partial/malformed output, raw/parsed equality, path
redaction, command parameters, and clean/dirty Git state. A bounded direct Make
capture passed the complete surface and independent validation with no
unredacted sensitive absolute path. Full shell, Python, docs, evidence,
knowledge, native race, JSON/Definition, and diff checks pass. R90-73 remains
in progress until feature push, fetched verification, and exact-range Vault
synchronization complete; no numeric baseline, threshold, or later increment
was started.
R90-73 completed early at
`e9fc0dc39fb08f4a5d667732bf594bd3edeb7120`: its exact eight-path feature was
pushed without force, fetched equal to `origin/main`, and passed the post-fetch
33-test knowledge gate. The immediate port-22 verification fetch disconnected
after the successful push; SSH-over-443 fetched the same exact remote SHA
before synchronization. Exact range
`b20845a8b7b4584e9cfa49aadc5ee663c17a2fe2..e9fc0dc39fb08f4a5d667732bf594bd3edeb7120`
was synchronized idempotently to the sole local Vault. Its iteration note,
full index, MOC link, and reconciled stable MOC/Makefile/testing authority are
verified. R90-74 is ready but was not started; R90-75 and R90-59 remain blocked
on their recorded external conditions.
The next Aug 6 trigger fetched and verified the R90-73 docs-only closure at
`b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`, both exact R90-73 Vault notes,
the full index, MOC links, and current stable benchmark authority. All 88 prior
task states parse and all 78 roadmap rows match one Definition. R90-74 is
selected as the sole dependency-ready increment; R90-75 and R90-59 remain
blocked.
Five sequential uncached default-parameter captures from one isolated clean
detached worktree share exact commit/tree, environment/toolchain fingerprint,
commands, and the complete six-C/eight-Go surface. Every raw JSON is retained
and SHA-256-bound. A tested versioned API recomputes 43 metric-series median,
inclusive IQR, sample deviation, coefficient-of-variation, and range summaries
and rejects sample/context/metric/digest/aggregate drift. The largest observed
CV is 11.252396% for matcher no-hit latency; no threshold or portable/
production claim is applied. Focused aggregation, direct recomputation,
Python, docs, and diff checks pass. The first full native gate exposed one
unchanged receiver idle-timeout failure; its exact test passed 20 uncached race
runs, no unrelated source changed, and the restarted complete fail-fast chain
passed all native, evidence, knowledge, JSON/Definition, and diff checks.
R90-74 remains in progress until push/fetch verification and exact-range Vault
synchronization.
R90-74 completed early at
`77e1ec005e077e1e66049a5a4eb809afd87fa23c`: its exact 15-path feature was
pushed without force through SSH-over-443, fetched equal to `origin/main`, and
passed the post-fetch 33-test knowledge gate plus direct baseline recomputation.
Exact range
`b3d4f8f82e8913093be518ffe426f1d6dc8eee7f..77e1ec005e077e1e66049a5a4eb809afd87fa23c`
was synchronized idempotently to the sole local Vault. Its iteration note,
full index, MOC link, and reconciled stable MOC/Makefile/testing authority are
verified. R90-75 remains blocked on comparable-environment evidence plus an
explicit product/SLO budget decision; R90-59 remains blocked on exact
publication authority. No next increment is dependency-ready.
The Aug 7 trigger keeps R90-75 blocked as pending evidence and explicitly
non-blocking; no comparable-environment data or product/SLO budget is inferred.
The user authorized only a local `v0.1.1` tag at exact candidate
`78cd78574e03c8f73ff68248eed2c409d6bca406` and withheld GitHub Release and
GHCR authority. Direct workflow review proves that pushing a `v*` tag would
trigger both external publications, so R90-59a is selected as a bounded local
signed-tag increment. R90-59 remains blocked on later tag-push and publication
authority; R90-75 is not started.
The exact candidate then passed the full v0.1.1 RC and release gate, including
native race, 78.3% coverage, ASan fuzz, E2E, archive, Docker image, and runtime
health smoke. Signed annotated local tag `v0.1.1` was created and verified at
the exact candidate; the remote tag remains absent. Candidate changelog review
found no `0.1.1` heading, and the fresh archive digest differs from R90-58, so
both facts remain explicit R90-59 remote-publication blockers. R90-59a awaits
repository delivery only; no workflow, GitHub Release, GHCR, or R90-75 work
started.
R90-59a completed at branch evidence commit
`afb435ce8c4e708c8b7b52c5b609d1f07e232891`: `main` was pushed with
`--no-follow-tags`, fetched equal to `origin/main`, and passed the post-fetch
knowledge gate while direct remote lookup continued to show no `v0.1.1` tag.
Exact branch range
`c19067172f1c626a59ba11b3201b276092721192..afb435ce8c4e708c8b7b52c5b609d1f07e232891`
was synchronized idempotently to the sole local Vault; its note, full index,
MOC link, and stable release authority are verified. R90-59 remains blocked on
the recorded changelog, artifact, and explicit remote-publication conditions.
R90-75 remains pending-evidence and non-blocking. No next increment is ready.
The Aug 9 trigger fetched and verified the R90-59a docs-only closure at
`5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc`, both exact R90-59a Vault notes,
full-index rows, MOC links, current stable release authority, the unchanged
signed local tag, and continued remote tag/GitHub Release/GHCR absence. All 90
prior task states parse and all 79 prior roadmap rows match one Definition.
R90-59 and R90-75 remain blocked on their recorded external conditions, so
R90-76 is selected as the documentation-only smallest safe queue unblocker.
The 161-commit phase audit found no missing recent delivery record or unresolved
validation deviation. Direct source and test review identified one bounded
local correctness gap: rule create/update/delete/reload transactions are not
serialized even though each replaces the full file and active snapshot. The
rule and suppression temporary-file paths also lack direct short-write,
file-sync, rename, and parent-directory-sync lifecycle evidence. R90-77 through
R90-80 now sequence transaction serialization, separately reviewable rule and
suppression durability, and a final management-plane audit through Oct 31.
None was started; R90-76 remains in progress until its documentation feature is
pushed, fetched, and synchronized.
All 91 task-state JSON files parse, all 84 roadmap rows match one Definition,
and every unfinished item has a complete status, dependency, window, risk,
acceptance, validation, and stop record. Documentation, the 33-test knowledge
gate, formatting, exact six-path scope, credential-prefix, sensitive-path, and
local/remote tag-state checks pass. R90-76 is validated and awaits only its
documentation feature delivery; R90-77 remains unstarted.
R90-76 completed at
`f3ddeda97375b5b92fbf0b0cdd08b21095e38fc0`: its exact six-path
documentation feature was pushed without force, fetched equal to
`origin/main`, and passed the post-fetch 33-test knowledge gate. Exact range
`5f6bf2ab4ae211e64f005b930de2ad3e84ee15fc..f3ddeda97375b5b92fbf0b0cdd08b21095e38fc0`
was synchronized idempotently to the sole local Vault. Its iteration note,
full-index row, MOC link, and reconciled stable MOC/rule/config/API authority
are verified. The remote `v0.1.1` tag remains absent. R90-77 is ready but was
not started; R90-59 and R90-75 remain blocked on their recorded external
conditions.
The next trigger fetched and verified the R90-76 docs-only closure at
`40798847be8e7bb9270b5c5d7675c27f7addf7b1` plus both exact Vault notes,
full-index rows, MOC links, and stable rule/config/API authority. The fresh
history and forward-queue audit found no new material deviation: R90-59 and
R90-75 retain their external blockers, R90-78 through R90-80 retain complete
dependency-ordered definitions, and R90-77 is the sole ready local increment.
R90-77 is selected with a persisted plan/state before behavior changes; no
later increment or publication action is started.
R90-77 now holds one API-server management mutex across the authoritative rule
state/file read, validation, canonical replacement when applicable, and active
snapshot publication for create, update, delete, and explicit reload. Direct
channel-synchronized create/create, update/delete, and mutation/reload tests
prove the second transaction cannot cross a blocked first transaction and that
both successful outcomes agree on disk and in memory. Validation and
persistence failures preserve prior bytes/state and release the lock for a
later valid request. Twenty uncached focused race repetitions, complete focused
ordinary/race tests, full native tests, E2E smoke, documentation, and the
33-test knowledge gate pass. The exact nine-path increment is validated and
awaits delivery; R90-78 remains unstarted.
R90-77 completed early at
`0ae76e167928f0ab1dafe015a997ccd1f61c664f`: its exact nine-path feature was
pushed without force or tags, fetched equal to `origin/main`, and passed the
post-fetch 33-test knowledge gate. Exact range
`40798847be8e7bb9270b5c5d7675c27f7addf7b1..0ae76e167928f0ab1dafe015a997ccd1f61c664f`
was synchronized idempotently to the sole local Vault. Its iteration note,
full-index row, MOC link, and current MOC/rule/config/API stable authority are
verified; stale pre-delivery stable prose was reconciled without rewriting
immutable iteration notes. R90-78 is ready but was not started; R90-59 and
R90-75 retain their external blockers.
The next trigger fetched and verified the R90-77 docs-only closure at
`4b5b199f37531e69c08cb7fa7b1d814f83047a37`, both exact R90-77 Vault notes,
full-index rows, MOC links, and current stable rule/config/API authority. The
Jul 20 through Aug 9 phase audit found no unresolved validation deviation or
missing delivery record; all 92 prior task states parse and all 84 roadmap
rows match one Definition. R90-78 is selected as the sole dependency-ready
local increment with a persisted lifecycle outcome contract and direct fault
evidence map. R90-79 and R90-80 remain dependency-planned; R90-59 and R90-75
retain their external blockers. No later increment or publication action is
started.
R90-78 now requires exact-length temporary writes, preserved mode, file sync,
file close, atomic rename, and containing-directory sync and close before a
successful rule mutation response. Direct faults cover stat, create,
short-write, write, chmod, file-sync, temp-close, rename, directory-open,
directory-sync, and directory-close boundaries with exact prior/new bytes and
temporary cleanup. Post-rename durability errors publish the committed
canonical rules to active memory and return
`RULES_DURABILITY_UNCERTAIN`; pre-rename errors retain prior file/state and
permit retry. Twenty uncached direct race repetitions, complete focused
ordinary/race tests, full native tests, E2E smoke, documentation, and the
33-test knowledge gate pass. The exact increment is validated and awaits
delivery; R90-79 remains unstarted.
R90-78 completed early at
`8d053d1d3c4e390151c224aa8f86852312506eb8`: its exact eleven-path feature is
the fetched `origin/main` tip with fast-forward ancestry from the recorded
baseline, and the post-fetch 33-test knowledge gate passes. Exact range
`4b5b199f37531e69c08cb7fa7b1d814f83047a37..8d053d1d3c4e390151c224aa8f86852312506eb8`
is synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current MOC/rule/config/API stable authority are
verified. The signed local `v0.1.1` tag remains absent remotely. R90-79 is ready
but was not started; R90-59 and R90-75 retain their external blockers.
The next trigger fetched and verified the R90-78 docs-only closure at
`17a5809f83959714f8801fdfa7e613520e06dd14`, both exact R90-78 Vault notes,
full-index rows, MOC links, and current stable rule/config/API authority. The
Jul 20 through Aug 9 phase audit found no new unresolved validation deviation,
stale stable authority, or missing delivery record; all 93 prior task states
parse and all 84 roadmap rows match one Definition. R90-79 is selected as the
sole dependency-ready local increment with a persisted suppression lifecycle
outcome contract and direct fault-evidence map. R90-80 remains
dependency-planned; R90-59 and R90-75 retain their external blockers. No later
increment or publication action is started.
R90-94 now creates and modes its listener at a private same-filesystem path,
publishes that verified socket identity with a non-replacing hard link, and
requires a non-following private/public identity match before receiver
ownership is assigned. The private link anchors identity through shutdown and
is removed with the public owned path before listener close. Direct
post-creation regular-file, symlink, and live-listener replacement regressions
preserve replacement bytes, modes, identities, target, and service without
publishing receiver ownership or leaving a private artifact; ordinary mode and
owned cleanup remain compatible. The first descriptor-stat design applied mode
but exposed kernel socket metadata rather than the filesystem pathname
identity, so it was rejected before acceptance evidence. The corrected focused
set and complete receiver package pass normally; twenty uncached acceptance
race runs and the complete receiver race package pass. Complete repository
validation remains pending, and no later increment is started.
The complete fail-fast repository chain passes both native C tests, every Go
package uncached under race, E2E smoke, documentation, and all 33 knowledge
tests. All 109 task-state JSON files parse and all 98 roadmap rows match exactly
one Definition with equal raw counts, no duplicate identifiers, and no
asymmetry. Each R90-94 acceptance criterion reaches its direct promised
boundary; formatting, exact seven-path scope, dependency/configuration/
protocol/public-API/release boundaries, and sensitive-information review pass.
The rejected descriptor-stat design is the only validation deviation and is
fully superseded by the corrected clean sequences. R90-94 satisfies its local
acceptance evidence and awaits only feature delivery, fetched remote
verification, and exact-range Vault synchronization. No later increment is
started.
R90-94 completed early at
`2e03300e46f3df1f98e47f72bada5207cc2e8fc3`: its exact seven-path feature was
pushed without force or tags. The first verification fetch returned no usable
ref or exit evidence, so Vault work remained blocked; an identical non-mutating
retry then verified `FETCH_HEAD == HEAD == origin/main` at the feature commit
with fast-forward ancestry from the recorded baseline. The post-fetch 33-test
knowledge gate passed. Exact range
`50a98397c1145b0915458ab662247b4a68542b27..2e03300e46f3df1f98e47f72bada5207cc2e8fc3`
was synchronized to the sole local Vault; its iteration note, full-index row,
and MOC link are verified. Stable MOC/UDS prose now records private created-
identity publication, the retained runtime identity anchor, direct replacement
preservation, and the completed R90-94 boundary. Identical-range replay
preserved Vault content hash
`9f66c134e78a28538734d1c0891009c4142e667c7172969f394facef09cee94b`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The Aug 14 trigger found the configured GitHub SSH port 22 route closed, then
fetched successfully through the documented SSH-over-443 transport and
verified the clean R90-94 docs-only closure at
`0dbf05acf1dcd233a9be6f76d54b947d77ff0290`. Both exact R90-94 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority are verified.
All 109 prior task states parse and all 98 prior roadmap row and Definition
multisets match without duplicates or asymmetry. The 161-commit Jul 20 through
Aug 14 phase review found no missing closure, stale stable authority, or
unresolved local validation result that changes priority. R90-59 and R90-75
retain their external blockers and no local row is ready, so R90-95 is selected
as the documentation-only smallest safe queue unblocker with a persisted
plan/state. Current `Start` checks cancellation before pathname preparation and
during the existing-socket probe, but `createUnixListener` receives no context;
the synchronized private-listener-created seam can observe cancellation while
startup still proceeds toward mode application, pathname publication,
ownership assignment, and a nil return. R90-96 records only cancellation after
private listener creation and before publication, and remains unstarted.
R90-95 completed at
`e109f91e012906546495eaa1fc18ee9aad71e064`: its exact three-path
documentation audit was pushed without force or tags through the documented
SSH-over-443 transport, freshly fetched with
`FETCH_HEAD == HEAD == origin/main`, and passed the post-fetch 33-test
knowledge gate. Exact range
`0dbf05acf1dcd233a9be6f76d54b947d77ff0290..e109f91e012906546495eaa1fc18ee9aad71e064`
was synchronized to the sole local Vault; its iteration note, full-index row,
and MOC link are verified. Stable MOC prose now records the audited
post-private-listener cancellation boundary and ready/unstarted R90-96
follow-on. Identical-range replay preserved Vault content hash
`595dcb5e24835a4ffef0c5e91188fd0313c1efe5e5d661d2b1f50ca46e4c1b00`.
R90-96 is the next ready local increment and remains unstarted; R90-59 and
R90-75 retain their external blockers.
R90-79 now requires exact-length temporary writes, preserved mode, file sync,
file close, atomic rename, and containing-directory sync and close before a
successful suppression mutation response. Direct faults cover stat, parent
creation, temp creation, short write, write, chmod, file-sync, temp-close,
rename, directory-open, directory-sync, and directory-close boundaries with
exact prior/new bytes and temporary cleanup. Post-rename durability errors
publish the committed candidate rules/filter and return
`SUPPRESSIONS_DURABILITY_UNCERTAIN`; pre-rename errors retain prior file/filter
state and permit retry. Complete focused ordinary/race tests and twenty
uncached direct race repetitions pass; full repository validation remains the
delivery boundary. R90-80 remains unstarted.
The complete repository validation boundary passes: C tests and every Go
package pass uncached under the race detector; E2E smoke processes six packets,
generates five alerts, and loads eight rules; repository config/suppression,
documentation, and all 33 knowledge tests pass. All 94 task-state JSON files
parse, all 84 roadmap rows match one Definition, every unfinished record has
complete selection fields, and formatting, exact eleven-path scope,
credential/sensitive-path, dependency, schema, config, workflow, release, and
publication reviews pass. R90-79 is validated and awaits only feature delivery;
R90-80 remains unstarted.
R90-79 completed early at
`8c621a926ac7ecbd1d730884a1afbc1ebb5e101e`: its exact eleven-path feature was
pushed without force or tags, fetched equal to `origin/main`, and passed the
post-fetch 33-test knowledge gate. Exact range
`17a5809f83959714f8801fdfa7e613520e06dd14..8c621a926ac7ecbd1d730884a1afbc1ebb5e101e`
was synchronized idempotently to the sole local Vault. Its iteration note,
full-index row, MOC link, and reconciled stable MOC/config/suppression/API
authority are verified. R90-80 is ready but was not started; R90-59 and R90-75
retain their external blockers.
The next trigger fetched and verified the R90-79 docs-only closure at
`de949bda14a66a407391671f92f0c7b938fb2da5`, both exact R90-79 Vault notes,
full-index rows, MOC links, and current stable rule/config/suppression/API
authority. The Jul 20 through Aug 9 phase review found no unresolved validation
result, stale current stable authority, or missing delivery record; all 94
prior task states parse and all 84 roadmap rows match one Definition. R90-80
is selected as the sole dependency-ready documentation audit. R90-59 and
R90-75 retain their external blockers; no runtime, policy, or publication work
is started.
The audit verifies the exact R90-77 through R90-79 feature/closure parent
chain, intended paths, direct synchronized/lifecycle/API regressions, completed
states, fetched remote, and all six Vault notes/index rows/MOC links. Current
API, architecture, development, changelog, and stable Vault claims correctly
limit the delivered behavior to one API-server process and the checked local
POSIX lifecycle. Legacy schema removal, cross-process writers, portable crash
evidence, and broader protocol work each require product, migration, platform,
or external-input authority and are not converted into speculative ready work.
R90-80's documentation audit is complete; repository validation and delivery
remain pending.
All 95 task-state JSON files parse, all 84 roadmap rows match one Definition,
and R90-59, R90-75, and R90-80 retain complete unfinished-item fields.
Documentation, the 33-test knowledge gate, formatting, exact six-path scope,
credential/sensitive-path, source/test/config/workflow/generated-evidence,
release, and publication reviews pass. R90-80 is validated and awaits only its
documentation delivery; no later increment is started.
R90-80 completed early at
`7d0d7884a9ba18e51113a74081b9bb1ae6206fa3`: its exact six-path
documentation audit was pushed without force or tags, fetched equal to
`origin/main`, and passed the post-fetch 33-test knowledge gate. Exact range
`de949bda14a66a407391671f92f0c7b938fb2da5..7d0d7884a9ba18e51113a74081b9bb1ae6206fa3`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC/config/API prose was
reconciled to the completed audit without changing immutable iteration notes;
identical-range replay preserved Vault content hash
`baed830f8a62bf5e0d732f1f93d9dd276e137c6ec2faa151eb36d065ba3bf51e`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers.
The next Aug 9 trigger fetched and verified the R90-80 docs-only closure at
`49ae9eb95c6ff500e3c525bff30d7a13a43b6938`, both exact R90-80 Vault notes,
full-index rows, MOC links, and current stable management-plane authority. The
Jul 20 through Aug 9 phase review found no missing delivery record or unresolved
behavioral validation result, but direct history and test review confirmed
receiver timing deviations in three separate full validation runs before
focused uncached reruns passed. Only the Aug 6 record names
`TestStartIdleTimeoutReleasesConnectionCapacity` exactly; Jul 29 records the
idle-timeout family with a broken-pipe symptom and Jul 23 records only a receiver
timing boundary. Because no local increment was ready, R90-81 is selected as the
documentation-only smallest safe queue unblocker. R90-82 records the bounded
direct-test evidence gap but is not started; R90-59 and R90-75 retain their
external blockers.
The R90-81 audit preserves the different specificity of the three historical
receiver timing records, verifies R90-80's exact feature/closure/remote/Vault
chain, and maps the exact Aug 6 idle-capacity failure to the current indirect
shared-session polling boundary. It restores R90-82 as a bounded test-evidence
increment without claiming a production defect or starting runtime/test work.
Repository validation and R90-81 delivery remain pending.
The first structural prevalidation found that completed historical row R90-04a
lacked its own Definition, contradicting prior row/Definition coverage claims.
R90-81 restored the missing definition directly from the completed plan/state
and preserved its non-traffic, non-release boundary; no historical delivery
record or immutable evidence was rewritten. Complete validation remains pending.
All 96 task-state JSON files parse and all 86 roadmap rows now match exactly one
Definition. The first structural check exposed the missing completed R90-04a
Definition; after evidence-grounded repair, documentation, the 33-test knowledge
gate, and formatting checks pass. Exact documentation scope and sensitive-data
review remain the pre-commit delivery boundary; R90-82 remains unstarted.
R90-81 completed at
`669e49965b5c76e659290469fea026af6c003c09`: its exact four-path
documentation feature was pushed without force or tags, fetched equal to
`origin/main`, and passed the post-fetch 33-test knowledge gate. Exact range
`49ae9eb95c6ff500e3c525bff30d7a13a43b6938..669e49965b5c76e659290469fea026af6c003c09`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC and test-gate prose now
records the R90-04a repair plus receiver evidence-strength boundary without
rewriting immutable iteration notes; identical-range replay preserved Vault
content hash `c030352a5d46d54eb18c4a03d369149a1194136ea3cdfbc592c204f226bafbe8`.
One initial idempotency replay ran from the Vault directory and failed before
mutation because the thin hook resolves its versioned script from the repository
working directory; replay from the repository root passed and remained
content-stable. R90-82 is ready but was not started; R90-59 and R90-75 retain
their external blockers.
The Aug 9 R90-82 trigger fetched and verified the R90-81 docs-only closure at
`9541d44db18b9c13e521b83be8aae79a9e5068be`, its exact Vault iteration note,
full-index row, MOC link, and current stable receiver test authority. The Jul 20
through Aug 9 phase review found no newer missing delivery record, stale stable
authority, or unresolved validation result that changes priority. R90-59 and
R90-75 retain their recorded external blockers, and every unfinished item keeps
a complete dependency, window, risk, acceptance, validation, and stop record.
R90-82 is selected as the sole dependency-ready increment with a persisted
plan/state and direct evidence map before receiver or test changes; no later
increment or publication action is started.
R90-82 now exposes the existing internal limiter to package tests as available
tokens without adding a public seam: each handler claims one token and returns
one only after exit. The protocol-violation, ordinary-disconnect, and idle-timeout
regressions directly claim that released token, then prove replacement packet
delivery without shared latest-session polling. The first focused command used
repository-relative source paths from the Go module and stopped before formatting
or tests; the corrected full focused sequence passed, followed by twenty
uncached race executions of all three direct regressions. Complete repository
validation remains the delivery boundary; no later increment is started.
R90-82's exact structural gate then found 86 roadmap rows but 87 Definition
headings: R90-81 had added a complete R90-04a Definition near the queue while an
older R90-04a Definition remained under the global policy history. The prior
R90-81 closure's exact-one claim was therefore incorrect. Delivery stayed
blocked while the duplicate was traced to Git history; the newer complete
definition remains active and the redundant older heading/prose was removed
without changing the completed R90-04a row or immutable R90-81 evidence. The
full structural and repository validation sequences must be rerun before
delivery.
The corrected structural gate proves 86 queue rows map one-to-one to 86
Definitions, all three unfinished records are complete, all 97 task-state JSON
files parse, and the diff is clean. The complete fail-fast repository rerun then
passes both C tests, every Go package uncached under race, E2E smoke, docs, and
all 33 knowledge tests. R90-82 satisfies its local acceptance evidence and
remains in progress only for exact staged-scope review, feature delivery,
fetched remote verification, and exact-range Vault synchronization; no later
increment is started.
R90-82 completed early at
`6118a0fb628a2a0ae0527c0783f436f96314a353`: its exact five-path feature was
pushed without force or tags, fetched equal to `origin/main`, and passed the
post-fetch 33-test knowledge gate. Exact range
`9541d44db18b9c13e521b83be8aae79a9e5068be..6118a0fb628a2a0ae0527c0783f436f96314a353`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC, test-gate, and UDS prose
now records the delivered available-token boundary plus the corrected R90-04a
duplicate-definition authority without rewriting immutable iteration notes;
identical-range replay preserved Vault content hash
`bdacddb75b810a02a1d87373646989491f0c47e4327b351ca9908d4c3e442a00`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The next Aug 9 trigger fetched and verified the R90-82 docs-only closure at
`49c31cf5682c232d1bc66d830b366d36603b7048`, both exact R90-82 Vault notes,
full-index rows, MOC links, and current stable receiver authority. The Jul 20
through Aug 9 phase review found no newer missing delivery record, stale stable
authority, or unresolved validation result. All 97 prior task states parse and
all 86 prior roadmap rows match one Definition. With R90-59 and R90-75 still
externally blocked and no ready row remaining, R90-83 is selected as the
documentation-only smallest safe queue unblocker. Current source directly
shows unconditional removal of the configured UDS pathname at startup and
shutdown, while tests do not cover pre-existing non-socket/symlink occupants or
a path replaced after listener creation. R90-84 records only that bounded
filesystem-identity preservation gap and remains unstarted; active/stale socket
policy, runtime/test work, and publication actions are not started.
All 98 task-state JSON files parse and all 88 roadmap rows map one-to-one to 88
Definitions with no duplicate identifiers. Documentation, the 33-test
knowledge gate, formatting, exact four-path scope, credential/sensitive-path,
source/test/config/workflow/generated-evidence, release, and publication
reviews pass. R90-83 satisfies its local acceptance evidence and awaits only
documentation delivery; R90-84 remains unstarted.
R90-83 completed at
`658bda36a75f8b0b5a5ed9ec7fec65087f1c9afc`: its exact four-path
documentation audit was pushed without force or tags, fetched equal to
`origin/main`, and passed the post-fetch 33-test knowledge gate. The first push
attempt returned without useful transport output and direct ref verification
showed the remote still at the recorded baseline; one retry was performed only
after that evidence and succeeded. Exact range
`49c31cf5682c232d1bc66d830b366d36603b7048..658bda36a75f8b0b5a5ed9ec7fec65087f1c9afc`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC and UDS prose now records
the bounded pathname-preservation gap and planned/unstarted R90-84 without
rewriting immutable iteration notes; identical-range replay preserved Vault
content hash `693eb7097cb835c549ed6d3ac4dca503d2b87d2d059e1d0921d53809ecd51f43`.
R90-84 is ready but was not started; R90-59 and R90-75 retain their external
blockers.
The Aug 9 R90-84 trigger fetched and verified the R90-83 docs-only closure at
`5c4253d18283c80ec27b7c2c1f383616eac2a89e`, its exact Vault iteration note,
full-index row, MOC link, and current stable UDS pathname authority. The Jul 20
through Aug 9 phase review found no newer missing delivery record, stale stable
authority, or unresolved validation result that changes priority. All 98 prior
task states parse and all 88 roadmap row and Definition multisets match exactly
without duplicate identifiers. R90-59 and R90-75 retain their external
blockers; R90-84 is selected as the sole dependency-ready increment with a
persisted ownership contract and direct evidence map before receiver or test
changes. No active/stale peer policy, cross-process locking, protocol change,
or publication action is started.
R90-84 now rejects pre-existing regular files and symlinks through non-following
pathname classification, retains pre-existing Unix-socket reclamation, disables
listener auto-unlink, and removes the pathname only when it matches the captured
created-socket identity. Five direct regressions cover every promised startup,
stale-socket, owned-cleanup, and replacement-path boundary. The first repeated
race run exposed that accept-loop completion could precede explicit cleanup
when unlink followed listener close; cleanup now occurs before close and the
complete focused sequence must be rerun. No later increment is started.
The corrected five direct pathname regressions pass once normally and twenty
times uncached under the race detector, followed by the complete receiver race
package. The complete fail-fast repository chain passes both C tests, every Go
package uncached under race, E2E smoke, documentation, and all 33 knowledge
tests. All 99 task states parse and all 88 roadmap rows match the complete
Definition multiset without duplicates. R90-84 satisfies its local acceptance
evidence and exact eight-path staged review; it awaits only feature delivery,
fetched remote verification, and exact-range Vault synchronization. No later
increment is started.
R90-84 completed early at
`8fc16241921dcb2817e2c138e59e36a6ab774b02`: its exact eight-path feature was
pushed without force or tags, fetched equal to `origin/main`, and passed the
post-fetch 33-test knowledge gate. Exact range
`5c4253d18283c80ec27b7c2c1f383616eac2a89e..8fc16241921dcb2817e2c138e59e36a6ab774b02`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC and UDS prose now records
the delivered pathname ownership boundary without rewriting immutable
iteration notes; identical-range replay preserved Vault content hash
`2eea69c66524a2c9664f896036e9e24fdd8d0269878bffe8a8f7162f2b7fe4a1`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The next Aug 9 trigger fetched and verified the R90-84 docs-only closure at
`79f6250de30c3128ecaec31e81ae19eecc9109d8`, both exact R90-84 Vault notes,
full-index rows, MOC links, and current stable UDS/MOC authority. The Jul 20
through Aug 9 phase review found no missing delivery record or unresolved
behavioral validation result, but the mutable roadmap placed the R90-84
completion paragraph before R90-82 completion while its tail still ended at
R90-84's pre-delivery checkpoint. With R90-59 and R90-75 externally blocked
and no ready row, R90-85 is selected as the documentation-only smallest safe
unblocker to repair chronology and audit the directly untested pre-canceled
receiver-start boundary. No runtime/test or publication work is started.
The R90-85 audit verifies the exact R90-84 feature/closure parent chain,
intended paths, completed state, fetched remote, both Vault notes/index/MOC
links, and current stable MOC/UDS authority. It moves the unchanged R90-84
completion facts after the increment's successful validation checkpoint and
records 141 Jul 20 through Aug 9 commits: 58 behavior-like changes, 72 delivery
closures, and 11 other documentation changes, with no unresolved behavioral
validation result or missing record. Current source launches the context watcher
only after listener/path state is published, while every direct cancellation
test cancels after `Start`; R90-86 captures only that deterministic
pre-canceled preservation boundary and remains unstarted. Repository validation
and R90-85 delivery remain pending.
All 100 task-state JSON files parse and all 90 roadmap rows match the complete
Definition multiset with equal raw counts, no duplicate identifiers, and no
asymmetry. Explicit marker-order validation proves the mutable R90-82 through
R90-85 delivery history is chronological. Documentation, all 33 knowledge
tests, and formatting pass. R90-85 satisfies its local acceptance evidence and
exact four-path staged-scope and sensitive-information review; it awaits only
delivery. R90-86 remains unstarted.
R90-85 completed at
`73a5f03ab685408d802084ce5864d33cfa3bf03b`: its exact four-path
documentation audit was pushed without force or tags, fetched equal to
`origin/main`, and passed the post-fetch 33-test knowledge gate. Exact range
`79f6250de30c3128ecaec31e81ae19eecc9109d8..73a5f03ab685408d802084ce5864d33cfa3bf03b`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, and MOC link are verified. Stable MOC and UDS prose now records
the corrected chronology and planned/unstarted R90-86 boundary without
rewriting immutable iteration notes; identical-range replay preserved Vault
content hash `8b9cb2e84a8335f7f9f025ca63234577a36580639b18c3459fb77c3fae09dda8`.
R90-86 is ready but was not started; R90-59 and R90-75 retain their external
blockers.
The next Aug 9 trigger fetched and verified the R90-85 docs-only closure at
`ab63ee3ef53fdb7a764ca0863dac36580d0318fa`, both exact R90-85 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority. All 100 prior
task states parse and all 90 roadmap row and Definition multisets match exactly
without duplicate or asymmetric identifiers. Recent delivery review found no
missing closure, stale stable authority, or unresolved validation result that
changes priority. R90-59 and R90-75 retain their external blockers; R90-86 is
selected as the sole dependency-ready increment with a persisted ownership and
evidence contract before receiver or test changes. No later increment or
publication action is started.
R90-86 now rejects an already-canceled context before pathname inspection,
stale-socket removal, or listener creation and wraps the original cancellation
sentinel. Two direct regressions prove an absent path remains absent and a
pre-existing Unix socket keeps the same filesystem identity with no receiver
listener installed. The corrected focused sequence passes once normally,
twenty times uncached under race, and as part of the complete receiver race
package; established live startup and post-readiness cancellation remain
compatible.
The complete fail-fast repository chain passes both C tests, every Go package
uncached under race, E2E smoke, documentation, and all 33 knowledge tests. All
101 task states parse and all 90 roadmap row and Definition multisets match
without duplicate or asymmetric identifiers. R90-86 satisfies its local
acceptance evidence and exact eight-path scope review; it awaits only feature
delivery, fetched remote verification, and exact-range Vault synchronization.
No later increment is started.
R90-86 completed early at
`97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3`: its exact eight-path feature was
pushed without force or tags, fetched equal to `origin/main`, and passed the
post-fetch 33-test knowledge gate. Exact range
`ab63ee3ef53fdb7a764ca0863dac36580d0318fa..97ef7c12b2ce254d2a6a57b8d5cf084f6e8ee4a3`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current stable MOC/UDS authority are verified.
Identical-range replay preserved Vault content hash
`41b42418edb5033763e8aa923f9f000765f6bac6cce27270f3c039c1884bc639`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The Aug 10 trigger used the documented SSH-over-443 fallback after port 22
closed and a first 443 fetch ended in a transient broken pipe; a keepalive
retry fetched and verified the clean R90-86 docs-only closure baseline at
`6ea917e976d71432a4beb72967f73f2abf5c908b`. Both exact R90-86 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority are verified.
All 101 prior task states parse and all 90 prior roadmap row and Definition
multisets match without duplicate or asymmetric identifiers. Recent phase
review found no missing closure, stale stable authority, or unresolved
validation result that changes priority. R90-59 and R90-75 retain their
external blockers and no local row is ready, so R90-87 is selected as the
documentation-only smallest safe queue unblocker with a persisted plan/state.
Current `removeExistingSocket` removes every pre-existing non-symlink Unix
socket after `Lstat`; the sole reclamation regression closes its listener
first, and no direct test preserves a live listener or its pathname identity.
R90-88 records only that bounded preservation outcome and remains unstarted.
The R90-87 audit reconciles the exact R90-86 feature/closure chain, 145 commits
across four recent phases, both immutable Vault notes, and current stable
authority. It distinguishes reachability from trust, records the transient
fetch and resolved R90-86 setup deviations, refreshes the horizon through
Nov 8, and restores only R90-88 behind this audit. All 102 task states parse;
all 92 roadmap rows and Definitions match as complete multisets without
duplicates or asymmetry; R90-59, R90-75, R90-87, and R90-88 retain complete
unfinished-item fields. Documentation, all 33 knowledge tests, formatting,
scope, and sensitive-information review pass. R90-87 satisfies its local
acceptance evidence and exact four-path scope; it awaits only documentation
delivery, fetched remote verification, and exact-range Vault synchronization.
R90-88 is not started.
R90-87 completed at
`d0cb83bec99d881c012738bb4a11a4ca2629e3cb`: its exact four-path documentation
audit was pushed without force or tags, fetched equal to `origin/main`, and
passed the post-fetch 33-test knowledge gate. Exact range
`6ea917e976d71432a4beb72967f73f2abf5c908b..d0cb83bec99d881c012738bb4a11a4ca2629e3cb`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current stable MOC/UDS authority are verified.
Identical-range replay preserved Vault content hash
`48ff083888de53a14ff4305d1de359edc9b2fe09068139077c8592043bcd5286`.
R90-88 is ready but was not started; R90-59 and R90-75 retain their external
blockers.
The next Aug 10 trigger fetched and verified the R90-87 docs-only closure at
`1dcda25ce728336a984892ae849dffeb1d01b4d6`, both exact R90-87 Vault notes,
full-index rows, MOC links, current stable MOC/UDS authority, and reconciled
Vault hash `1f14b6313b9692b419f9bc4a3c0ee4eb03b9cd0178e0f0032da11ee3e25335ef`.
All 102 prior task states parse and all 92 roadmap row and Definition
multisets match without duplicates or asymmetry. The 147-commit recent phase
review found no missing closure, stale stable authority, or unresolved
validation result that changes priority. R90-59 and R90-75 retain their
external blockers; R90-88 is selected as the sole dependency-ready increment
with a persisted ownership and evidence contract before receiver or
documentation changes. No later increment or publication action is started.
R90-88 now probes a pre-existing Unix socket with a bounded local connection,
rejects a connectable listener, and treats only connection refusal as a stale
candidate before re-inspecting its non-following identity. The first focused
run exposed immediate inode reuse: device/inode equality alone removed a
replacement listener and incorrectly allowed startup. Delivery remained
blocked while the check added the captured change timestamp. The corrected
active-listener, ambiguous-probe, immediate-replacement, and stale-reclamation
regressions pass normally and twenty times uncached under race; the complete
receiver race package also passes. Full repository validation remains pending.
The complete fail-fast repository chain passes both C tests, every Go package
uncached under race, E2E smoke, documentation, and all 33 knowledge tests. All
103 task states parse and all 92 roadmap row and Definition multisets match
without duplicates or asymmetry. R90-88 satisfies its local acceptance
evidence, including direct continued-service and immediate-replacement
boundaries, and exact eight-path scope review; it awaits only feature delivery,
fetched remote verification, and exact-range Vault synchronization. No later
increment is started.
R90-88 completed early at
`b551b71ebb7cf4d6cdee0d249a68490412e925eb`: its exact eight-path feature was
pushed without force or tags. The first verification fetch returned no usable
exit/ref evidence, so synchronization stayed blocked until an identical retry
fetched `FETCH_HEAD == HEAD == origin/main` at the feature commit. The
post-fetch 33-test knowledge gate passed. Exact range
`1dcda25ce728336a984892ae849dffeb1d01b4d6..b551b71ebb7cf4d6cdee0d249a68490412e925eb`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current stable MOC/UDS authority are verified.
Identical-range replay preserved Vault content hash
`aab59bbf7fa2486f302e4eaa0bfbe35cc68bce955f3f85ef34cff52b989e565e`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The Aug 11 trigger fetched and verified the clean R90-88 docs-only closure at
`56d7d0b8005601299292b47d49bee7fc1e651753`, both exact R90-88 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority. All 103 prior
task states parse and all 92 prior roadmap row and Definition multisets match
without duplicates or asymmetry. The 149-commit Jul 20 through Aug 11 phase
review found no missing closure, stale stable authority, or unresolved
validation result that changes priority. R90-59 and R90-75 retain their
external blockers and no local row is ready, so R90-89 is selected as the
documentation-only smallest safe queue unblocker with a persisted plan/state.
Current `Start` checks only an already-canceled context before a fixed
one-second `net.DialTimeout` probe that cannot observe later cancellation;
direct tests cover pre-canceled startup and post-readiness shutdown but not
cancellation synchronized during that pre-readiness probe. R90-90 records only
that bounded prompt-cancellation and pathname-preservation outcome and remains
unstarted.
The R90-89 audit reconciles the exact R90-88 feature/closure parent chain,
intended paths, completed state, fetched remote, both immutable Vault notes,
and current stable authority. It records 149 commits across four recent phases:
60 behavior-like changes, 76 delivery closures, and 13 other documentation
changes, with no missing record, stale stable authority, or unresolved
validation result that changes priority. Public lifecycle prose does not claim
prompt during-probe cancellation. Source and direct-test mapping confirms the
probe seam has no context parameter and uses fixed `net.DialTimeout`, while
pre-canceled and post-readiness cancellation are covered on either side of the
untested boundary. Only planned R90-90 is restored to pass `Start`'s context
through the bounded probe and directly prove prompt cancellation plus pathname
identity preservation; it remains unstarted pending R90-89 validation and
delivery.
All 104 task-state JSON files parse and all 94 roadmap rows match the complete
Definition multiset with equal raw counts, no duplicate identifiers, and no
asymmetry. R90-59, R90-75, R90-89, and R90-90 retain complete unfinished-item
fields. Documentation, all 33 knowledge tests, formatting, exact four-path
scope, and sensitive-information review pass. R90-89 satisfies its local
acceptance evidence and awaits only documentation delivery, fetched remote
verification, and exact-range Vault synchronization. R90-90 remains unstarted.
R90-89 completed at
`d30729c9b6e2331fca834123a3f876f1c3b91df1`: its exact four-path
documentation audit was pushed without force or tags. The first trigger's push
produced no output or completion evidence and was interrupted; ordinary SSH,
SSH-over-443, and HTTPS/API verification then timed out, so the remote result
remained ambiguous and Vault work stayed blocked. On the resumed trigger, a
fresh fetch proved `origin/main` was still the recorded baseline. The single
authorized retry succeeded, and a second fresh fetch verified
`FETCH_HEAD == HEAD == origin/main` at the feature commit. The post-fetch
33-test knowledge gate passed. Exact range
`56d7d0b8005601299292b47d49bee7fc1e651753..d30729c9b6e2331fca834123a3f876f1c3b91df1`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current stable MOC/UDS authority are verified.
Identical-range replay preserved Vault content hash
`340e0654c221cf3e5bba249cdfcb9abc87505eb954e237677de42c3730c015d3`.
R90-90 is ready but was not started; R90-59 and R90-75 retain their external
blockers.
The next Aug 11 trigger fetched and verified the clean R90-89 docs-only closure
at `22ba8ce639d79547875885f4ce107321273dd3b7`, both exact R90-89 Vault
notes, full-index rows, MOC links, current stable MOC/UDS authority, and
reconciled closure-range Vault hash
`fbe378e8ad8a865ad65aca0c78a118d280481b876f4170a75aa9dba05160c022`.
All 104 prior task states parse and all 94 prior roadmap row and Definition
multisets match without duplicates or asymmetry. The 151-commit Jul 20 through
Aug 11 phase review found no missing closure, stale stable authority, or
unresolved validation result that changes priority. R90-59 and R90-75 retain
their external blockers; R90-90 is selected as the sole dependency-ready
increment with a persisted context, pathname-preservation, and evidence
contract before receiver or test changes. No later increment or publication
action is started.
R90-90 now passes the startup context through pathname preparation and a
receiver-local bounded `DialContext` probe without changing refusal-only stale
classification. Its direct regression synchronizes on probe entry before
canceling, then proves prompt `context.Canceled` matching, no installed
receiver listener, and complete captured device/inode/change-time identity
preservation. The acceptance and compatibility set passes normally, twenty
times uncached under race, and as part of the complete receiver race package.
No implementation or focused-validation deviation occurred; complete
repository validation remains the delivery boundary.
The complete fail-fast repository chain passes both C tests, every Go package
uncached under race, E2E smoke, documentation, and all 33 knowledge tests. All
105 task states parse and all 94 roadmap row and Definition multisets match
without duplicate or asymmetric identifiers. R90-90 satisfies every local
acceptance criterion through the direct synchronized regression and exact
eight-path scope review; it awaits only feature delivery, fetched remote
verification, and exact-range Vault synchronization. No later increment is
started.
The first closure edit anchored the R90-90 completion paragraph on a repeated
generic `started.` sentence and placed it inside older R90-59a history. The
chronology check rejected that placement before validation; the unchanged
completion facts now follow R90-90 selection, implementation, and successful
validation, and the ordered-marker gate must pass before closure delivery.
R90-90 completed early at
`c17870eb7f829b7451ab866b00fead4ef6b72e92`: its exact eight-path feature
was pushed without force or tags, fetched equal to `origin/main`, and passed
the post-fetch 33-test knowledge gate. Exact range
`22ba8ce639d79547875885f4ce107321273dd3b7..c17870eb7f829b7451ab866b00fead4ef6b72e92`
was synchronized idempotently to the sole local Vault; its iteration note,
full-index row, MOC link, and current stable MOC/UDS authority are verified.
Identical-range replay preserved Vault content hash
`07ee79b395aab016fc0e7617999a4f3bca5a8e5b59c772e5a4efec703b3ba997`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The Aug 12 trigger fetched and verified the clean R90-90 docs-only closure at
`c0b1eb2dae8dd90eda745eacc87b0a6ece01a450`, both exact R90-90 Vault
notes, full-index rows, MOC links, and current stable MOC/UDS authority. All
105 prior task states parse and all 94 prior roadmap row and Definition
multisets match without duplicates or asymmetry. The 153-commit Jul 20 through
Aug 12 phase review found the R90-90 closure-placement deviation resolved
before delivery and no missing record, stale stable authority, or unresolved
validation result that changes priority. R90-59 and R90-75 retain their
external blockers and no local row is ready, so R90-91 is selected as the
documentation-only smallest safe queue unblocker with a persisted plan/state.
Current receiver shutdown captures its created socket with non-following
metadata but `removeOwnedSocket` accepts Unix mode plus device/inode identity
alone. The existing regular-file and symlink replacement tests do not exercise
a replacement Unix listener or immediate inode reuse; the already-delivered
startup classification regression proves the local filesystem can reuse that
identity and that change time is required as a generation signal. R90-92
records only fail-closed generation-aware cleanup plus direct immediate
replacement-listener preservation and remains unstarted.
All 106 task-state JSON files parse and all 96 roadmap rows match the complete
Definition multiset with equal raw counts, no duplicate identifiers, and no
asymmetry. Ordered history proves R90-90 completion precedes R90-91 selection
and R90-92 planning. Documentation, all 33 knowledge tests, formatting, exact
three-path scope, and sensitive-information review pass. The first history
edit matched an older generic no-ready marker; it was moved after the exact
R90-90 completion tail before validation, with no evidence or scope change.
R90-91 satisfies its local acceptance evidence and awaits only documentation
delivery, fetched remote verification, and exact-range Vault synchronization.
R90-92 remains planned and unstarted.
R90-91 completed at
`972a6714caf91e089a220eeec88c16944a47757d`: its exact three-path
documentation audit was pushed without force or tags. The push produced no
output or completion evidence and was interrupted after bounded polling; no
retry occurred. A fresh fetch then proved the remote had advanced and verified
`FETCH_HEAD == HEAD == origin/main` at the feature commit with fast-forward
ancestry from the recorded baseline. The post-fetch 33-test knowledge gate
passed. Exact range
`c0b1eb2dae8dd90eda745eacc87b0a6ece01a450..972a6714caf91e089a220eeec88c16944a47757d`
was synchronized to the sole local Vault; its iteration note, full-index row,
MOC link, and current stable MOC/UDS authority are verified. Identical-range
replay preserved reconciled Vault content hash
`49af611ecaa37348a699ae392715c157025c9abc51355c8eecea114ae447d2a2`.
R90-92 is the next ready local increment and remains unstarted; R90-59 and
R90-75 retain their external blockers.
The next Aug 12 trigger initially found GitHub SSH unavailable on ports 22 and
443, then fetched successfully through the documented IPv4 SSH-over-443
keepalive retry and verified the clean R90-91 docs-only closure at
`29c291a7dffcc37caf0375910e1ad1c6ef0a54a4`. Both exact R90-91 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority are verified.
All 106 prior task states parse and all 96 prior roadmap row and Definition
multisets match without duplicates or asymmetry. The 155-commit Jul 20 through
Aug 12 phase review found only the R90-91 feature and closure since the prior
audit, with no missing record, stale stable authority, or unresolved validation
result that changes priority. R90-59 and R90-75 retain their external blockers;
R90-92 is selected as the sole dependency-ready local increment with a
persisted generation-identity and direct-evidence contract before receiver or
test changes. No later increment or publication action is started.
R90-92 shutdown cleanup now requires the receiver's captured non-following
device/inode/change-time identity. The first focused run exposed that startup
captured ownership before its intended `chmod`, so ordinary cleanup correctly
failed closed after the change timestamp advanced. The existing mode mutation
now precedes the ownership snapshot. A direct real-filesystem regression proves
device/inode reuse with a changed generation, preserves the replacement
listener identity, and completes a service round trip; a separate missing-
generation regression also fails closed. The corrected focused acceptance and
compatibility set passes; repeated and full validation remain the delivery
boundary.
The direct acceptance and compatibility set then passed 20 uncached race
executions, followed by a clean complete receiver-package race run. The
module-selected Go 1.25.12 toolchain and repository tool surface were
preflighted before the complete chain. Both C tests, every Go package uncached
under race, E2E smoke, documentation, and all 33 knowledge tests pass. All 107
task states parse and all 96 roadmap rows match exactly one Definition without
duplicate or asymmetric identifiers. R90-92 satisfies its local acceptance
evidence and exact eight-path scope review; it awaits only feature delivery,
fetched remote verification, and exact-range Vault synchronization. No later
increment is started.
R90-92 completed early at
`b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e`: its exact eight-path feature
was pushed without force or tags through the documented IPv4 SSH-over-443
transport, then freshly fetched with
`FETCH_HEAD == HEAD == origin/main` at the feature commit and fast-forward
ancestry from the recorded baseline. The post-fetch 33-test knowledge gate
passed. Exact range
`29c291a7dffcc37caf0375910e1ad1c6ef0a54a4..b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e`
was synchronized to the sole local Vault; its iteration note, full-index row,
and MOC link are verified. Stable MOC/UDS prose now records the delivered
generation-bound cleanup and direct replacement-listener evidence; identical-
range replay preserved Vault content hash
`e368106cf93971de1a76bb46ed0753fb04338a2453f62d3452d277088eea7217`.
No dependency-ready local increment remains. R90-59 and R90-75 retain their
recorded external blockers, and neither was started.
The next Aug 12 trigger fetched and verified the clean R90-92 docs-only closure
at `c59c3aca6a67b1975f178734d6b0f81a6bcab6b8`, both exact R90-92 Vault
notes, full-index rows, MOC links, and current stable MOC/UDS authority. All
107 prior task states parse and all 96 prior roadmap row and Definition
multisets match without duplicates or asymmetry. The 157-commit Jul 20 through
Aug 12 phase review found only the R90-92 feature and closure since the prior
audit, with no missing record, stale stable authority, or unresolved validation
result that changes priority. R90-59 and R90-75 retain their external blockers
and no local row is ready, so R90-93 is selected as the documentation-only
smallest safe queue unblocker with a persisted plan/state. Current startup uses
pathname-based `os.Chmod` after `net.Listen` but before it verifies and captures
the non-following pathname; Go documents that pathname-based `Chmod` follows a
symlink. Current direct replacement evidence covers pre-start paths, the stale
probe, and shutdown, not replacement after listener creation. R90-94 records
only created-listener-bound mode application plus fail-closed pathname
ownership and remains unstarted.
All 108 task-state JSON files parse and all 98 roadmap rows match the complete
Definition multiset with equal raw counts, no duplicate identifiers, and no
asymmetry. Ordered history proves R90-92 completion precedes R90-93 selection
and R90-94 planning. Documentation, all 33 knowledge tests, formatting, exact
three-path scope, and sensitive-information review pass. The first two history
edits matched older no-ready markers; the unchanged audit paragraph was finally
anchored after the exact R90-92 completion tail before validation. R90-93
satisfies its local acceptance evidence and awaits only documentation delivery,
fetched remote verification, and exact-range Vault synchronization. R90-94
remains planned and unstarted.
R90-93 completed at
`0628005cd7c606dd14e6cf0fa2c4fb12042ecf65`: its exact three-path
documentation audit was pushed without force or tags, freshly fetched with
`FETCH_HEAD == HEAD == origin/main`, and passed the post-fetch 33-test
knowledge gate. Exact range
`c59c3aca6a67b1975f178734d6b0f81a6bcab6b8..0628005cd7c606dd14e6cf0fa2c4fb12042ecf65`
was synchronized to the sole local Vault; its iteration note, full-index row,
and MOC link are verified. Stable MOC/UDS prose now records the audited
post-listen pathname boundary and ready/unstarted R90-94 follow-on; identical-
range replay preserved Vault content hash
`e085f102fbb71ac087ae5f3d629a91bad6fb5db4aa0d09f0134253690fb69624`.
R90-94 is ready but was not started; R90-59 and R90-75 retain their external
blockers.
The Aug 13 trigger fetched and verified the clean R90-93 docs-only closure at
`50a98397c1145b0915458ab662247b4a68542b27`, both exact R90-93 Vault notes,
full-index rows, MOC links, and current stable MOC/UDS authority. The 159-commit
Jul 20 through Aug 13 phase review found no missing closure, stale stable claim,
or unresolved local validation result that changes priority. R90-59 and R90-75
retain their explicit external blockers; R90-94 is selected as the sole
dependency-ready local increment with a persisted listener-identity,
replacement-preservation, evidence, non-goal, and authority contract before
receiver or compatibility-documentation changes. Current `Start` applies the
configured mode through the pathname after `net.Listen`, then captures a later
non-following pathname without proving it identifies the created listener;
existing direct tests do not replace the pathname in that interval. No later
increment or publication action is started.
