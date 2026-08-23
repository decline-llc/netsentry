# Task Plan: R90-105 Go 1.25.14 toolchain security refresh

## Metadata

- Timestamp: 2026-08-23T09:39:55-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `8724b816a77c4bdeac899e4848dcb5bcd5232a93`
- Selected toolchain: `go1.25.14`

## Goal

Clear the current-`main` R90-59 supply-chain blocker by updating the selected
Go 1.25 execution toolchain to its latest reviewed security patch while
preserving the `go 1.22.2` language baseline, dependency graph, runtime
behavior, and all publication boundaries.

## Scope

- Record the authoritative Go 1.25.14 release source and official Linux amd64
  archive SHA-256
  `a21ae5633a269bcd7e90cf767e48225633795e99d831742cbf3397064fee7712`.
- Update `engine/go.mod` and `.github/supply-chain-lock.json` together from
  Go 1.25.12 to Go 1.25.14.
- Update only current public toolchain-policy prose; preserve immutable
  historical evidence about prior validations under Go 1.25.12.
- Revalidate the pinned workflow, supply-chain, release-candidate, release,
  documentation, knowledge, Git delivery, and Vault boundaries.
- Carry the R90-59 blocked plan/state as prerequisite evidence while keeping
  R90-59 incomplete.

## Non-Goals

- Do not change `go 1.22.2`, Go dependencies, application behavior, protocol,
  configuration, public API, workflow logic, Action pins, scanner pins, or
  external fixture identities.
- Do not move, recreate, resign, retarget, delete, or push the local `v0.1.1`
  tag; do not dispatch a workflow, create a GitHub Release, or publish a GHCR
  image.
- Do not claim the historical tagged candidate is repaired by a later `main`
  commit or that R90-59 is complete.
- Do not start R90-75, choose a performance budget, or access private data.

## Risks

- Go 1.25.13 is only the scanner-reported minimum fixed version; selecting it
  would omit the later Go 1.25.14 `net/http` security fix.
- Updating the module directive without the lock, or validating with a
  different runtime, can produce a false clean supply-chain result.
- Broadly rewriting historical Go 1.25.12 evidence would corrupt the record of
  what earlier candidates actually used.
- A clean `main` does not change the immutable signed tag target and cannot by
  itself authorize publication.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Latest selected-line patch is reviewed | Official Go release history identifies Go 1.25.14; official download metadata matches the recorded Linux amd64 SHA-256 |
| Language semantics remain stable | `engine/go.mod` retains `go 1.22.2` and changes only the toolchain directive |
| Runtime and lock agree | Exact `go env GOVERSION == go1.25.14`; supply-chain lock/current documentation agree |
| Security gate is clean | Pinned `govulncheck v1.6.0` reports zero reachable vulnerabilities under Go 1.25.14 |
| Supply-chain inputs remain intact | Pinned `actionlint v1.7.12`, structural policy, and all 9 fetched asset size/hash checks pass |
| Repository compatibility remains intact | Complete native and `VERSION=0.1.1 make rc-check` sequence plus release gate pass |
| Publication boundary remains intact | Local tag object/peeled target remain unchanged; no tag push, workflow dispatch, Release, or GHCR mutation occurs |
| Delivery is recoverable | Exact commit and fetched `origin/main`, post-fetch knowledge check, exact-range Vault note/index/MOC, and stable supply-chain knowledge reconciliation |

## Validation

- Verify the official Go 1.25.14 Linux amd64 archive checksum before using its
  compiler, and require `go env GOVERSION` to return exactly `go1.25.14`.
- Preflight the complete repository-pinned tool surface before costly checks.
- Run focused workflow/lock checks, then
  `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check` with the exact pinned
  `actionlint` and `govulncheck` binaries.
- Run the complete fail-fast native and `VERSION=0.1.1 make rc-check` boundary,
  then the selected public release gate.
- Run documentation, task-state JSON, complete roadmap row/Definition multiset,
  knowledge, diff, intended-scope, and anchored sensitive-information checks.
- Verify the local `v0.1.1` tag object and peeled target are unchanged.
- Commit and push only `main` without force or tags; fetch and require
  `FETCH_HEAD == HEAD == origin/main`, rerun `make knowledge-check`, then
  synchronize and verify the exact full-SHA range in the sole local Vault.

## Authority Boundaries

This trigger authorizes only the R90-105 Go 1.25.14 current-`main` toolchain
refresh, current policy/documentation reconciliation, prerequisite R90-59
blocker evidence, validation, a `main` commit/push without tags, and local
Vault synchronization. It does not authorize a language/dependency/behavior
change, tag mutation or push, workflow dispatch, GitHub Release, GHCR/image
publication, R90-59 completion, R90-75, or private input.

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, active roadmap,
  R90-59 blocker plan/state, current toolchain lock/policy, release workflows,
  and latest R90-104 delivery evidence.
- Last verified `HEAD == origin/main` baseline is
  `8724b816a77c4bdeac899e4848dcb5bcd5232a93`; a later SSH fetch failed before
  updating refs during a transient connection reset.
- The exact local signed `v0.1.1` tag remains object
  `f1a38ecb82b9c63e8411f3df040bdea84e985dd8`, peeling to candidate
  `78cd78574e03c8f73ff68248eed2c409d6bca406`; its remote tag and GitHub Release
  were last directly verified absent.
- Exact-candidate structural/actionlint checks passed, but pinned
  `govulncheck v1.6.0` found reachable `GO-2026-6090`, `GO-2026-6089`, and
  `GO-2026-5972` in Go 1.25.12; both tag workflows run that gate before
  publication.
- Official Go release history records security releases Go 1.25.13 on Aug 13
  and Go 1.25.14 on Aug 19, with the later patch fixing `net/http`; the official
  download metadata provides the recorded Linux amd64 checksum.
- Selected R90-105 as the single smallest local unblocker. R90-59 remains
  blocked pending separate new-candidate/tag authority, and R90-75 remains
  unstarted.

## Checkpoints

- The plan, state, evidence map, non-goals, authority boundaries, and stop
  conditions were persisted before the toolchain pin or current policy prose
  changed. The R90-59 blocker record now permits delivery only as R90-105
  prerequisite evidence and remains incomplete.
- The official 59,909,419-byte Go 1.25.14 Linux amd64 archive was downloaded in
  bounded ranges after single-connection TLS/DNS instability and independently
  matched SHA-256
  `a21ae5633a269bcd7e90cf767e48225633795e99d831742cbf3397064fee7712`.
  Go's authenticated toolchain-module path separately installed the runtime,
  which reports `go1.25.14 linux/amd64` and `go env GOVERSION == go1.25.14`.
- Newly version-stamped `actionlint v1.7.12` and `govulncheck v1.6.0` binaries
  were built under Go 1.25.14 after the complete local tool surface and Docker
  daemon were preflighted. Focused workflow validation passed.
- `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check` passed with runtime
  Go 1.25.14, all 7 locked Actions/11 uses, all 9 fetched asset size/hash
  checks, and zero reachable vulnerabilities. The scan found one vulnerability
  in a required module but proved it unreachable. Complete repository and
  release validation remains the delivery boundary.
- The first complete `VERSION=0.1.1 make rc-check` passed every non-Docker
  stage, including both C tests, every Go package uncached under race, 81.3%
  Go statement coverage, both 5,000-iteration sanitizer fuzz targets, E2E
  smoke, archive/checksum/content, and release-note smoke. Docker then spent
  906 seconds fetching its uncached pinned syntax frontend and stopped before
  any image build step when the configured Ubuntu mirror returned EOF while
  resolving `ubuntu:24.04`. No Docker result or downstream release-gate result
  from that stopped sequence is counted. The exact base-image metadata must be
  preflighted before rerunning the Docker build and its content/runtime smoke.
- Three mirror-backed Ubuntu prefetch attempts and one exact-digest mirror
  attempt failed on EOF/reset transport errors. Pulling the same image by the
  explicit Docker Hub hostname then succeeded and verified immutable digest
  `sha256:33ceb71981b602c1a7443a53469e4dba065f7503eab3078a2d7a57a2ab987517`.
  The ordinary Dockerfile frontend tag lookup still failed on the configured
  mirror, so the equivalent build used BuildKit's supported `BUILDKIT_SYNTAX`
  override pinned to already fetched frontend digest
  `sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32`
  without changing the repository or daemon configuration.
- The digest-pinned Docker build selected Go 1.25.14 inside the build stage and
  completed as local image
  `sha256:1fdc62d56aa7fe9c4e4347523676f07094fcc60660cfbf867868900c718a46bb`.
  Image-content and runtime `/api/health` smoke passed. The v0.1.1 non-PCAP
  release gate then passed. The ordinary mutable-tag mirror lookup deviation
  remains recorded; no image was pushed and no publication occurred.
- Final workflow and offline supply-chain checks pass again with exact runtime
  Go 1.25.14 and zero reachable vulnerabilities. Documentation, all 42
  evidence tests, all 33 knowledge tests, all 121 task-state JSON files, and
  the complete 109-row/109-Definition roadmap multiset pass. Formatting is
  clean, and the local tag object/peeled candidate remain exactly
  `f1a38ecb82b9c63e8411f3df040bdea84e985dd8` and
  `78cd78574e03c8f73ff68248eed2c409d6bca406`. R90-105 satisfies local
  acceptance and awaits only scoped commit/push/fetch/Vault delivery.

## Stop Conditions

Stop on any release/checksum/runtime/lock mismatch, reachable vulnerability,
failed asset integrity check, regression, tag identity change, remote-ref
ambiguity, credential uncertainty, or Vault verification gap. Do not recover
through a different Go line, dependency/language change, tag mutation,
publication, or broader increment without new evidence and authority.
