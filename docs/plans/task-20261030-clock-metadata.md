# R90-200: fixture clock metadata handoff

Selected 2026-10-30 from fetched origin/main
`7de6fdf534834f8629cda06bf42755da96001297` in an isolated clean worktree.
R90-199 exact hosted success and local Vault delivery are verified. The original
main and R90-196 retry worktrees contain preserved pending records, excluded here.
The latest retry completed acquisition but the collector rejected clock_method;
its numerical diagnostics do not establish accepted L1. Four-week history shows
bounded correctness, fixture and evidence repairs; no new feature expansion.

## Scope and authority

Repair only the fixture's emitted clock-method representation. Keep a stable
bounded identifier and retain the full algorithm description separately in the
raw clock result. Preserve the clock calculation, 10ms threshold, workload,
legacy schemas, rejected manifests and stricter packet-loss diagnostics.
No real traffic corpus is available or needed for this no-send repair. Real-data
and production conclusions remain unverified; required software gates remain.
Do not run L1/L2, modify host clocks, publish a release or mix pending work.

## Acceptance and evidence

| Criterion | Required evidence |
| --- | --- |
| Actual producer metadata accepted downstream | No-send fixture frozen declaration and actual clock_verdict output through collection_manifest and public derive; verify schema/contract/run identity and exact uncertainty |
| Clock semantics unchanged | Existing drift/step tests plus exact bound/threshold assertions and preserved description; legacy golden regression |
| Rejected input remains rejected and unchanged | Regression using the old descriptive method through derive, byte comparison before/after, no adapter/report acceptance |
| Historical and pending artifacts preserved | Local hash inventory before/after, including both earlier attempts and pending worktrees |
| Reviewable complete delivery | Focused tests, pinned tool/module/Docker preflight, full make ci-check, knowledge/diff/sensitive review, one batch upload, exact fetched SHA hosted success and Vault verification |

## Risk, validation and stop

Do not turn diagnostic metadata into an acceptance certificate or relax a
consumer validator. Run direct regression before the complete gate; preflight
exact Go1.26.10, actionlint1.7.12, govulncheck1.6.0, engine module and Docker
daemon/frontend/base metadata. Stop delivery on required gate failure, ambiguous
clock meaning, evidence mutation or additional unsupported contract errors.
Record deviations and refresh unfinished queue without starting its next item.
Post-push facts stay in local delivery records and Vault; no fact-only upload.

## Local outcome and delivery resume

The actual producer-to-public-derive regression reproduced the identifier error
before repair and passes afterward with exact method and 2001ns uncertainty.
Rejected old metadata remains rejected without source mutation, legacy golden
and exact 10ms boundary regressions pass. All 443 historical/pending files remain
unchanged, with the clock calculation/threshold AST unchanged. Full ci-check
passed through complete Docker runtime health; see the sanitized evidence for
its log hash and daemon-recovery/tool-path deviations. Every acceptance boundary
maps to a direct regression or retained verification, not a live-acceptance claim.
Final prose is followed by affected docs/static/knowledge checks before commit.

Resolve this increment's commit and common-Git local delivery record; only exact
fetched remote equality, hosted success and verified Vault complete delivery.
Once verified do not repeat any upload or send closure-only facts. Refresh the
queue for two fresh R90-196 runs without starting them in this increment.
