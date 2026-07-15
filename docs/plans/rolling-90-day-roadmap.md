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
| R90-04a | Jul 15–Aug 21 | Ready | Revalidate the v0.1.1 code-quality baseline independently of production-traffic evidence. | R90-03a | A fresh non-evidence quality baseline is recorded from passing applicable code, dependency, workflow, and release-candidate checks; it makes no release-ready or production-evidence claim. |
| R90-04 | Aug 22–Sep 11 | Blocked | Obtain and review sanitized production-derived pcap evidence for v0.1.1, then run corpus-pressure validation. | Approved sanitized traffic source and privacy review | Evidence is path-redacted, explicitly production-derived and sanitized, and passes the documented corpus-pressure/release-evidence requirements. |
| R90-05 | Sep 12–Oct 2 | Pending | Prepare v0.1.1 release readiness from validated evidence. | R90-04; passing code quality gates | `make rc-check`, supply-chain, and release gates pass; public docs/evidence identify no unresolved release blocker. |
| R90-06 | Oct 3–14 | Pending | Assemble a release decision package. | R90-05 | Version, commit, evidence, checksums, and intended publication decision are reconciled; do not tag or publish without explicit user authorization. |

## Dependency and Priority Policy

`R90-01 → R90-02 → R90-03`; `R90-03a → R90-04a`; `R90-04 → R90-05 → R90-06`. R90-04a is an evidence-independent quality increment and does not satisfy any R90-04 dependency. Prioritize an active release blocker or a correctness/security regression over this order only when its task state documents the deviation. R90-04 remains blocked: this repository must not invent, fetch, or treat synthetic traffic as production-derived evidence.

## R90-04a Definition

- **Goal:** establish a current, reproducible v0.1.1 code-quality baseline while the privacy-controlled traffic-evidence process is unavailable.
- **Window:** Jul 15–Aug 21, 2026; selected as the next ready increment by explicit user direction on Jul 15.
- **Risk:** a passing quality suite could be misread as release approval or as replacement traffic evidence.
- **Required validation:** run the applicable non-evidence quality, dependency, workflow, and release-candidate checks; record any unavailable check precisely.
- **Stop condition:** stop without starting R90-04 if a required check is ambiguous or if continuation would require private traffic, privacy-review authority, a release decision, tagging, or publication.

## Current Checkpoint

R90-03a was completed early to remove CI coupling to unversioned hook files. On Jul 15, 2026, R90-04a was added and selected as the evidence-independent alternate while R90-04 remains blocked pending approved sanitized traffic evidence. This is a scheduling deviation: privacy review and traffic artifacts are unavailable, so the impact is that v0.1.1 evidence and release-preparation work remain blocked while code-quality validation can continue. No product-release task is authorized merely by its scheduled window.
