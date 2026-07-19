# NetSentry Rolling 90-Day Roadmap

> Window: 2026-07-17 through 2026-10-15. This is the active delivery queue for `$netsentry-next`; refresh unfinished work at each completed increment using Git, task-state, and evidence as authority.

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
| R90-13 | Jul 19–Sep 4 | In progress | Bound idle UDS receiver connections. | R90-12 | A validated finite per-connection read timeout applies before the first frame and refreshes after every complete frame; idle expiry releases handler capacity without inflating decode errors, while active traffic and shutdown remain compatible. |

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
  R90-13 is the active engineering increment. No tag or public release is
  authorized.

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

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-04b → R90-05 → R90-06 → R90-07 → R90-08 → R90-09 → R90-10 → R90-11 → R90-12 → R90-13`. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. The R90-04 and R90-05 PCAP exceptions remain immutable historical delivery evidence. The later global PCAP waiver supersedes their restrictions for current and future release-gate decisions.

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

R90-03a was completed early to remove CI coupling to unversioned hook files. On Jul 15, 2026, R90-04a was added, selected, and completed as the evidence-independent alternate: the non-Docker RC suite and pinned supply-chain baseline passed. On Jul 16, R90-04 used MAWI samplepoint-B trace `200012281400` under the scoped exception: provenance, privacy, sanitization, and sensitive-metadata reviews were approved; corpus pressure processed 544,525 packets with zero parse errors, drops, and UDS write errors. Feature commit `009b2a03776987359661c4ab2776f5d04820db34` is verified on fetched `origin/main`, post-fetch knowledge validation passed, and the exact pushed range is present in the Vault iteration note, full index, and MOC. The public record is `docs/evidence/r90-04-public-traffic-20260716.md`; it is R90-04-only and does not grant release approval. R90-04b completed at `64979f454cfee414cbb216368a8ee2fb34147e4d`: the audit exception is explicitly expired, the release gate rejects its reuse, the historical v0.1.0 fixture remains valid, and fetched `origin/main` plus Vault evidence are verified. R90-05 completed early at `6c3f9ef276c99c13aa9e985b8c849bb5f0791752`: the exact 7,500-packet synthetic corpus was verified by SHA-256, corpus pressure passed without capture errors, the full Docker RC and pinned supply-chain gates passed, and the user explicitly approved final v0.1.1 release-gate acceptance. On Jul 16, the user cancelled every roadmap window restriction and every PCAP release-gate restriction. The PCAP waiver delivery at `ec2605e3e8c99749933530d77ad1eb0200b8b47e` is verified on fetched `origin/main`; post-fetch knowledge validation and its exact Vault range are verified. PCAP tooling remains optional, while non-PCAP release controls and the separate tag/publication authority remain enforced. R90-06 completed under the window waiver: candidate `ad8a443b5020037c235419f5696c60988d2bba99` passed the full Docker RC, release gate, and pinned supply-chain audit; its local v0.1.1 linux/amd64 artifact is reconciled by SHA-256 in `docs/evidence/release-decision-v0.1.1.md`; and decision-package commit `c70a48d6e1272b5d0f127b848b761376bb1924a3` is verified on fetched `origin/main` and in the Vault. Publication remains on hold, `v0.1.1` was not created, and no GitHub Release or GHCR image was published. R90-07 completed early at `bdca014a5ca3c775125b41d98faf15ffd1b1cf35`: configuration rejects connection limits outside 1–1024, excess accepted clients are closed without starting handlers, disconnects release capacity for reconnects, the full race suite passed, fetched `origin/main` and post-fetch knowledge validation match, and the exact pushed range is verified in the Vault note, full index, and MOC. R90-08 completed early at `9129d4ecf9df0da9601a027ec118af6f58b96e9a`: HTTP bind failures are synchronous, shutdown waits for receiver handlers, workers, and HTTP before SQLite closes, the active-load drill and immediate-cancel regression passed 25 focused race runs, the full native suite and E2E smoke passed, fetched `origin/main` and post-fetch knowledge validation match, and the exact pushed range is verified in the Vault note, full index, and MOC. R90-09 completed early at `dee7c5f30f11082b76a6ba7f9d3cc6a41be349f4`: an existing non-empty primary database is checked through a separate read-only connection before writable open, corrupt and truncated fixtures return `ErrDatabaseIntegrity` and remain byte-for-byte unchanged, empty and healthy databases still open, the focused race tests, full native suite, and E2E smoke passed, fetched `origin/main` and post-fetch knowledge validation match, and the exact pushed range is verified in the Vault note, full index, and MOC. R90-10 completed early at `d08702d2c3a1e425c27b4bba5238039915603c97`: existing non-empty non-current daily shards receive the same separate read-only preflight before writable initialization, corrupt and truncated write targets return `ErrDatabaseIntegrity` and remain byte-for-byte unchanged, missing/empty/healthy targets remain writable, 20 focused race runs, the full native suite, E2E smoke, and knowledge gate passed, fetched `origin/main` matches, and the exact pushed range is verified in the Vault note, full index, and MOC. R90-11 completed early at `7e5c381880f21498d370789bb3db4c37a8e2254f`: non-current daily-shard query and count use URL-safe SQLite `mode=ro` handles, corrupt/truncated inputs fail all four direct read cases without byte changes, active WAL-backed reads work through an encoded path, 20 focused race runs, the full native suite, E2E smoke, and knowledge gate passed, fetched `origin/main` matches, and the exact pushed range is verified in the Vault note, full index, and MOC. R90-12 completed early at `4bc298b37b5da690b978a28acf6e4ece41956d41`: malformed JSON and otherwise valid records missing the final JSONL newline fail startup with `ErrRecoveryLogIntegrity`, rejected bytes and valid prefixes remain unmodified, a SQLite persistence failure retains valid recovery input, 20 focused race runs, the full native suite, E2E smoke, and knowledge gate passed, fetched `origin/main` matches, and the exact pushed range is verified in the Vault note, full index, and MOC. No later increment is selected; publication remains unauthorized.
