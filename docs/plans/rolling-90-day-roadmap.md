# NetSentry Rolling 90-Day Roadmap

> Window: 2026-07-22 through 2026-10-20. This is the active delivery queue for `$netsentry-next`; refresh unfinished work at each completed increment using Git, task-state, and evidence as authority. Completed history from the prior horizon is preserved below.

## Status Rules

- **Ready**: every dependency is complete. As of 2026-07-16, roadmap dates are
  forecasting metadata only and never prevent work from starting.
- **Blocked**: requires an explicitly recorded external input, authority, or unresolved validation result.
- **Complete**: acceptance criteria and required evidence are verified, including commit/push/Vault evidence when a repository increment was delivered.
- Complete only one ready increment per `$netsentry-next` trigger. Record deviations before reordering unfinished work.

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
| R90-24 | Jul 23–Oct 20 | In progress | Reject write-affecting SQLite generated columns. | R90-23 | Existing primary and historical databases with virtual or stored generated columns on `alerts` or `alert_events` fail read-only preflight and remain unchanged; ordinary nullable and defaulted column extensions remain compatible and writable. |

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
  evidence. R90-24 is the selected next increment. Publication remains
  unauthorized.

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

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-04b → R90-05 → R90-06 → R90-07 → R90-08 → R90-09 → R90-10 → R90-11 → R90-12 → R90-13 → R90-14 → R90-15 → R90-16 → R90-17 → R90-18 → R90-19 → R90-20 → R90-21 → R90-22 → R90-23 → R90-24`. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. The R90-04 and R90-05 PCAP exceptions remain immutable historical delivery evidence. The later global PCAP waiver supersedes their restrictions for current and future release-gate decisions.

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

R90-24 was selected on Jul 23 after refreshing the empty queue from the clean fetched R90-23 baseline, completed task state, release boundaries, write-critical generated-column metadata, and verified Vault evidence. It exposes columns omitted by `PRAGMA table_info` and rejects virtual or stored generated extensions on `alerts` and `alert_events` before writable initialization while retaining ordinary compatible columns. Implementation and delivery validation are in progress. Publication remains unauthorized.
