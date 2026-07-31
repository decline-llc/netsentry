# Task Plan: R90-58 refresh the v0.1.1 candidate decision package

## Metadata

- Timestamp: 2026-07-31T01:31:14-07:00
- Branch: main
- Risk Level: High
- Remote baseline: `78cd78574e03c8f73ff68248eed2c409d6bca406`

## Goal

Refresh the v0.1.1 hold-state decision package against the current fetched
candidate after the completed hardening sequence, with fresh release-candidate,
supply-chain, release-gate, artifact, checksum, platform, remote, and Vault
evidence.

## Scope

- Treat clean fetched commit
  `78cd78574e03c8f73ff68248eed2c409d6bca406` as the candidate under review.
- Run validation and build the release artifact from an isolated clean
  worktree at that exact commit so planning edits cannot enter the artifact.
- Reconcile the candidate version, commit, archive name, byte size, SHA-256,
  and `linux/amd64` platform.
- Refresh the public decision package and release-readiness status while
  preserving an explicit publication hold.
- Reconcile the roadmap and task state, then deliver a docs-only closure record
  after the feature-range push and Vault verification.

## Non-Goals

- Do not modify runtime code, schemas, protocols, APIs, dependencies, workflows,
  or release-gate policy.
- Do not reuse the historical candidate artifact or checksum as fresh evidence.
- Do not create, move, sign, or push `v0.1.1`.
- Do not publish a GitHub Release, GHCR image, workflow dispatch, or artifact.
- Do not change R90-57's product-decision blocker or R90-59's publication
  authorization boundary.
- Do not commit generated archive bytes, local evidence, private paths,
  credentials, or operator details.

## Risks

- Running against a dirty checkout can bind the artifact to uncommitted files
  instead of the stated candidate.
- Reusing an old archive or checksum after `main` advanced would misidentify
  the proposed tag payload.
- Remote fixture downloads or Docker validation can fail ambiguously because
  of infrastructure rather than candidate behavior.
- A passing release gate can be mistaken for authority to tag or publish.
- A docs-only delivery commit advances `main`; it records the candidate review
  but is not itself silently substituted as the validated candidate.

## Validation

- Verify the isolated worktree commit equals the fetched candidate.
- `VERSION=0.1.1 make rc-check`, including Docker build, image-content smoke,
  runtime health smoke, and generated archive checks.
- `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check`.
- `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate`.
- Independently verify archive filename, platform, byte size, SHA-256, checksum
  file, release-notes version, and archive contents from the candidate
  worktree.
- `make docs-check`, `make evidence-check`, `make knowledge-check`, all
  task-state JSON parsing, Markdown/JSON structure checks, `git diff --check`,
  staged-diff review, and sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, synchronize each exact full-SHA delivery range, and verify
  its Vault note, full index, MOC link, and stable release/testing note.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Exact current candidate | Clean isolated worktree and fetched commit both equal `78cd78574e03c8f73ff68248eed2c409d6bca406` |
| Fresh candidate gates | Full Docker RC, fetched supply-chain audit, and v0.1.1 release gate pass on that worktree |
| Exact artifact identity | Fresh archive name, byte size, SHA-256, checksum verification, release-notes version, and `linux/amd64` platform |
| Explicit hold | Decision package states no tag, GitHub Release, GHCR publication, or publication authority |
| Accurate delivery record | Roadmap/task state distinguish candidate SHA from subsequent documentation commits |
| Repository delivery | Fetched remote equality, post-fetch knowledge gate, and exact-range Vault note/index/MOC/stable-note verification |

## Trigger Audit

- Fetched `origin/main` and verified clean local and remote equality at
  `78cd78574e03c8f73ff68248eed2c409d6bca406`.
- Verified the R90-56 feature and closure commits, both exact Vault notes, the
  full commit index, MOC links, and the 33-test knowledge gate.
- Reconciled 249 July commits. Since July 14 the history contains 55
  implementation/test commits and 63 documentation commits, including the
  expected feature/closure chain through R90-56; no unrecorded delivery
  deviation remains.
- Verified all 62 roadmap rows have matching Definitions and every unfinished
  increment retains status, dependency, window, risk, acceptance criteria,
  required validation, and a stop condition.
- R90-57 remains blocked on its product decision. R90-58 is the only
  dependency-ready increment. R90-59 remains blocked on exact publication
  authorization.

## Authority Boundaries

The user authorized one rolling roadmap increment and its normal commit, push,
and local Vault workflow. This does not authorize a version tag, GitHub Release,
GHCR push, workflow dispatch, product-policy choice, or publication decision.

## Stop Conditions

Stop on candidate/ref drift, ambiguous or unavailable required validation,
artifact/version/checksum/platform mismatch, a need to modify runtime behavior
or release policy, private data, a product decision, tag creation, publication,
or external coordination.

## Execution Deviation

- **Observed:** The first combined validation attempt completed the full Docker
  RC, then the supply-chain target stopped because `actionlint` was absent.
- **Impact:** The release gate did not run, and the partial sequence was not
  accepted as complete evidence.
- **Resolution:** Installed the repository-pinned `actionlint v1.7.12` and
  `govulncheck v1.6.0` into a temporary directory using Go 1.25.12, then reran
  the complete RC, fetched supply-chain audit, and release gate from the clean
  candidate worktree. The rerun passed without changing the candidate.

## Validation Evidence

- Clean detached worktree `HEAD`:
  `78cd78574e03c8f73ff68248eed2c409d6bca406`, matching fetched
  `origin/main`.
- `VERSION=0.1.1 make rc-check`: pass, including native race tests, 78.3% Go
  statement coverage, 5,000-iteration ASan parser fuzz smoke, 6-packet/5-alert
  E2E smoke with 8 rules, archive checks, Docker image-content smoke, and
  Docker runtime health.
- `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check`: pass; 9/9 locked
  external assets matched and `govulncheck` reported zero reachable
  vulnerabilities.
- `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate`: pass.
- Fresh archive: `netsentry-0.1.1-linux-amd64.tar.gz`, 9,760,241 bytes,
  SHA-256
  `c68e09df46d24307c9a0d405a2724573f3382813a8b2611bdb5f3b7d8b068568`;
  checksum, archive contents, versioned release notes, and `linux/amd64`
  platform independently verified.
- `make docs-check`, `make evidence-check`, and `make knowledge-check` passed;
  the knowledge gate ran 33 tests. All task-state JSON parsed, all 62 roadmap
  rows retained Definitions, and `git diff --check` reported no findings.

All candidate, gate, and artifact acceptance criteria have direct evidence.
Repository delivery and exact-range Vault evidence remain pending.
