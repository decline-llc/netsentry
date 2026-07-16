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
| R90-05 | Sep 12–Oct 2 | Pending | Prepare v0.1.1 release readiness from validated evidence. | R90-04; passing code quality gates | `make rc-check`, supply-chain, and release gates pass; public docs/evidence identify no unresolved release blocker. |
| R90-06 | Oct 3–14 | Pending | Assemble a release decision package. | R90-05 | Version, commit, evidence, checksums, and intended publication decision are reconciled; do not tag or publish without explicit user authorization. |

## Dependency and Priority Policy

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-05 → R90-06`. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. The approved R90-04 exception permits only anonymized public real-traffic evidence after the required reviews; it does not satisfy R90-05 or R90-06 and does not permit synthetic or generated traffic. All other increments retain their production-derived PCAP requirements.

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

## Current Checkpoint

R90-03a was completed early to remove CI coupling to unversioned hook files. On Jul 15, 2026, R90-04a was added, selected, and completed as the evidence-independent alternate: the non-Docker RC suite and pinned supply-chain baseline passed. On Jul 16, R90-04 used MAWI samplepoint-B trace `200012281400` under the scoped exception: provenance, privacy, sanitization, and sensitive-metadata reviews were approved; corpus pressure processed 544,525 packets with zero parse errors, drops, and UDS write errors. Feature commit `009b2a03776987359661c4ab2776f5d04820db34` is verified on fetched `origin/main`, post-fetch knowledge validation passed, and the exact pushed range is present in the Vault iteration note, full index, and MOC. The public record is `docs/evidence/r90-04-public-traffic-20260716.md`; it is R90-04-only and does not grant release approval. No increment is ready on Jul 16: R90-05 remains pending until its Sep 12 window begins and R90-06 remains pending on R90-05.
