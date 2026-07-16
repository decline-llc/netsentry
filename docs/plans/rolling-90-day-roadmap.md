# NetSentry Rolling 90-Day Roadmap

> Window: 2026-07-14 through 2026-10-14. This is the active delivery queue for `$netsentry-next`; refresh unfinished work at each completed increment using Git, task-state, and evidence as authority.

## Status Rules

- **Ready**: every dependency is complete and the earliest window has begun.
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
| R90-06 | Oct 3–14 | Pending | Assemble a release decision package. | R90-05 | Version, commit, evidence, checksums, and intended publication decision are reconciled; do not tag or publish without explicit user authorization. |

## Dependency and Priority Policy

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-04b → R90-05 → R90-06`. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. The approved R90-04 exception permits only anonymized public real-traffic evidence after the required reviews; it does not satisfy R90-05 or R90-06 and does not permit synthetic or generated traffic. R90-04b makes that completed boundary executable before release preparation. The separately approved R90-05 exception applies only to the exact synthetic corpus digest and packet count recorded in `docs/audit/release_exception_r9005.yaml`; it does not change R90-06 or future production-derived PCAP requirements.

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

R90-03a was completed early to remove CI coupling to unversioned hook files. On Jul 15, 2026, R90-04a was added, selected, and completed as the evidence-independent alternate: the non-Docker RC suite and pinned supply-chain baseline passed. On Jul 16, R90-04 used MAWI samplepoint-B trace `200012281400` under the scoped exception: provenance, privacy, sanitization, and sensitive-metadata reviews were approved; corpus pressure processed 544,525 packets with zero parse errors, drops, and UDS write errors. Feature commit `009b2a03776987359661c4ab2776f5d04820db34` is verified on fetched `origin/main`, post-fetch knowledge validation passed, and the exact pushed range is present in the Vault iteration note, full index, and MOC. The public record is `docs/evidence/r90-04-public-traffic-20260716.md`; it is R90-04-only and does not grant release approval. R90-04b completed at `64979f454cfee414cbb216368a8ee2fb34147e4d`: the audit exception is explicitly expired, the release gate rejects its reuse, the historical v0.1.0 fixture remains valid, and fetched `origin/main` plus Vault evidence are verified. R90-05 completed early at `6c3f9ef276c99c13aa9e985b8c849bb5f0791752`: the exact 7,500-packet synthetic corpus was verified by SHA-256, corpus pressure passed without capture errors, the full Docker RC and pinned supply-chain gates passed, and the user explicitly approved final v0.1.1 release-gate acceptance. The pushed commit is verified on fetched `origin/main`; post-fetch knowledge validation and the exact Vault range are verified. The R90-05 exception is now expired and the release gate rejects reuse. Tag creation and publication remain unauthorized. R90-06 remains pending until its Oct 3 window.
