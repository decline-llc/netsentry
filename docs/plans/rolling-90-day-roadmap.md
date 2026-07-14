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
| R90-03 | Aug 8–21 | Pending | Add roadmap self-check and deviation-reporting workflow. | R90-02 | Each checkpoint compares actual work to the roadmap and records blockers, re-prioritization, dates, and the next ready increment. |
| R90-04 | Aug 22–Sep 11 | Blocked | Obtain and review sanitized production-derived pcap evidence for v0.1.1, then run corpus-pressure validation. | Approved sanitized traffic source and privacy review | Evidence is path-redacted, explicitly production-derived and sanitized, and passes the documented corpus-pressure/release-evidence requirements. |
| R90-05 | Sep 12–Oct 2 | Pending | Prepare v0.1.1 release readiness from validated evidence. | R90-04; passing code quality gates | `make rc-check`, supply-chain, and release gates pass; public docs/evidence identify no unresolved release blocker. |
| R90-06 | Oct 3–14 | Pending | Assemble a release decision package. | R90-05 | Version, commit, evidence, checksums, and intended publication decision are reconciled; do not tag or publish without explicit user authorization. |

## Dependency and Priority Policy

`R90-01 → R90-02 → R90-03`; `R90-04 → R90-05 → R90-06`. Prioritize an active release blocker or a correctness/security regression over this order only when its task state documents the deviation. R90-04 remains blocked: this repository must not invent, fetch, or treat synthetic traffic as production-derived evidence.

## Current Checkpoint

R90-02 was completed early in response to the knowledge-check request; the reported discrepancies did not reproduce locally (`make knowledge-check`: 33 tests passed). R90-03 remains pending until its scheduled window or a user-directed priority change. No product-release task is authorized merely by its scheduled window.
