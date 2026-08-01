# NetSentry Rolling 90-Day Roadmap

> Window: 2026-08-01 through 2026-10-30. This is the active delivery queue for `$netsentry-next`; refresh unfinished work at each completed increment using Git, task-state, and evidence as authority. Completed history from the prior horizon is preserved below.

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
| R90-59 | Oct 22–28 | Blocked | Execute the v0.1.1 publication gate. | R90-58; explicit publication authorization | Only an explicitly authorized version and commit may be tagged; GitHub Release and GHCR results must be verified directly, while absence of authority preserves the hold without mutation. |
| R90-60 | Forecast Aug 1–Oct 30; waived | In progress | Implement operator-triggered restart-free storage recovery. | R90-57 | One authenticated request serializes recovery against store lifecycle operations, preflights durable input before the writable boundary, replays or probes idempotently, exposes bounded health/audit outcomes, and leaves failures in sticky emergency without automatic cleanup or retry. |

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

- **Goal:** execute a separately authorized v0.1.1 publication decision and
  verify immutable external outcomes.
- **Risk:** tagging the wrong commit or inferring workflow/registry success can
  create an unrecoverable public release mismatch.
- **Required validation:** exact authorization/version/SHA match, annotated tag
  and signature checks, GitHub Release assets/checksums, GHCR digest/platform,
  workflow result, documentation, remote, and Vault evidence.
- **Blocker evidence:** the v0.1.1 decision package records publication status
  as hold and states that tag, GitHub Release, and GHCR publication are not
  authorized.
- **Unblock condition:** the user explicitly authorizes the exact version and
  candidate commit after R90-58 evidence is complete.
- **Stop condition:** remain blocked without explicit publication authority;
  stop on any SHA, tag, digest, platform, workflow, or artifact ambiguity.

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

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-04b → R90-05 → R90-06 → R90-07 → R90-08 → R90-09 → R90-10 → R90-11 → R90-12 → R90-13 → R90-14 → R90-15 → R90-16 → R90-17 → R90-18 → R90-19 → R90-20 → R90-21 → R90-22 → R90-23 → R90-24 → R90-25 → R90-26 → R90-27 → R90-28 → R90-29 → R90-30 → R90-31 → R90-32 → R90-33 → R90-34 → R90-35 → R90-36 → R90-37 → R90-38 → R90-39 → R90-40 → R90-41 → R90-42 → R90-43 → R90-44 → R90-45 → R90-46 → R90-47 → R90-48 → R90-49 → R90-50 → R90-51 → R90-52 → R90-53 → R90-54 → R90-55 → R90-56 → R90-57 → R90-60`; `R90-56 → R90-58 → R90-59`, with R90-59 blocked on explicit publication authorization. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. The R90-04 and R90-05 PCAP exceptions remain immutable historical delivery evidence. The later global PCAP waiver supersedes their restrictions for current and future release-gate decisions.

## R90-04 Scoped Evidence Exception

- **Authority and scope:** `docs/audit/release_exception_r9004.yaml` authorizes an R90-04-only alternative to internal production-derived PCAP evidence.
- **Allowed evidence:** anonymized, publicly released, real network traffic only. Synthetic or generated traffic is permanently prohibited.
- **Required controls:** approve dedicated privacy review, provenance validation, sanitization review, and sensitive-metadata screening before corpus-pressure validation or official-evidence use.
- **Boundary:** this exception expires when R90-04 completes and does not amend R90-05, R90-06, or future increment requirements.

## R90-04a Definition

- **Goal:** establish a current, reproducible v0.1.1 code-quality baseline while the privacy-controlled traffic-evidence process is unavailable.
- **Window:** Jul 15–Aug 21, 2026; selected as the next ready increment by explicit user direction on Jul 15.
- **Risk:** a passing quality suite could be misread as release approval or as replacement traffic evidence.
- **Required validation:** run the applicable non-evidence quality, dependency, workflow, and release-candidate checks; record any unavailable check precisely.
- **Stop condition:** stop without starting R90-04 if a required check is ambiguous or if continuation would require private traffic, privacy-review authority, a release decision, tagging, or publication.

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
and exact remote/Vault evidence remain pending.
