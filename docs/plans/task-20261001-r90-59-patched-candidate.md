# Task Plan: R90-59 patched v0.1.1 candidate and publication

## Metadata

- Timestamp: 2026-10-01
- Branch: main
- Risk Level: High
- Starting remote baseline: `5dced1bc9576f769d770a227d3989fbe0c0f4ea4`
- Historical tag object: `f1a38ecb82b9c63e8411f3df040bdea84e985dd8`
- Historical candidate: `78cd78574e03c8f73ff68248eed2c409d6bca406`
- Target tag: `v0.1.1`

## Goal

Prepare and validate a patched v0.1.1 candidate, replace and resign the local
tag under the user's explicit authorization, publish through the checked-in
tag workflows, verify all remote artifacts, and update the current audit.

## Scope

- Select the smallest release candidate that includes the reviewed Go 1.26.8
  security patch and the exact approved v0.1.1 payload.
- Refresh the release changelog/evidence as needed without overstating
  synthetic traffic as production evidence.
- Complete full behavioral/release-candidate checks, fetched supply-chain
  checks with zero reachable findings, release gate, and knowledge checks.
- Replace and resign only local `v0.1.1` after successful validation, then push
  the candidate branch and exact tag to the configured origin.
- Verify tag workflows, GitHub Release assets/checksum, and GHCR digest/platform.
- Update R90-59 roadmap, task state, latest audit, and local knowledge vault.

## Authority

The user explicitly authorizes a patched candidate and replacement/resigning
of `v0.1.1`, followed by revalidation. Publication remains subject to the
existing R90-59 explicit publication grant and successful exact-candidate
gates. Only this version/tag and its checked-in tag-triggered workflows are in
scope. R90-75 remains separate.

## Acceptance and evidence

| Criterion | Evidence |
| --- | --- |
| Patch is current and correctly pinned | Official Go support policy/release source, exact toolchain pin and checksum |
| Candidate content is reviewed | Exact commit/tree, changelog and release evidence review |
| Behavioral gates pass | Full `VERSION=0.1.1 make rc-check` including native, fuzz, E2E, Docker and runtime smoke |
| Supply-chain gate passes | Fetched asset/license verification and zero reachable `govulncheck` findings |
| Release gate passes | Exact candidate `make release-gate` |
| Knowledge checks pass | `make knowledge-check` before and after verified branch delivery |
| Replacement tag is authentic | New signed annotated tag object, signature, exact peeled candidate |
| Publication is verified | Remote tag, both workflows, release assets/checksum, GHCR digest and platform |
| Audit is recoverable | Roadmap/state/evidence, fetched remote, exact Vault synchronization |

## Non-goals and risks

- Do not modify unrelated runtime behavior, workflows, dependencies, versions,
  or publish additional tags/images/platforms.
- Do not use historical artifacts as substitutes for newly generated assets.
- Do not claim production-derived evidence from synthetic fixtures.
- Stop before tag replacement/push if the signing identity, any candidate gate,
  release evidence, remote identity, or artifact result is ambiguous or fails.

## Validation sequence

1. Confirm clean fetched baseline, tag identity/signature, absent remote tag and
   release, signing identity, current Go patch, Docker/frontend/base-image
   availability, and all exact pinned tool versions.
2. Finalize and review the isolated candidate and release evidence.
3. Run full candidate RC, supply-chain fetched-asset/vulnerability gate, and
   release gate fail-fast; independently verify archive/checksum/platform.
4. Run behavioral tests and `make knowledge-check`; inspect docs/JSON/roadmap
   structure, diff, staged scope, and sensitive-data scan.
5. Create/verify replacement local signed tag, push only the exact tag, fetch
   and verify its object and candidate, wait for both workflows, and inspect all
   artifacts.
6. Record exact evidence, run knowledge checks after fetch, push/fetch-verify
   the audit closure, and synchronize each exact full-SHA range to the local
   Vault.

## Initial observations

- Fetched baseline is clean at `5dced1bc9576f769d770a227d3989fbe0c0f4ea4`.
- The existing local signed tag still targets the historical candidate; direct
  remote tag lookup returned no `v0.1.1` ref.
- Current main pins Go 1.25.14, but official policy and release history show
  Go 1.26.8 is the latest patch in the selected supported 1.26 line; the historical
  tagged candidate remains on Go 1.25.12.
- The prior SSH tag signature and current release signing identity match at
  fingerprint `SHA256:lanK75hksvHVuDmY55rdL1CJVqj4ZjqgBuE5/kMm4ZU`; direct
  verification passes using a temporary allowed-signers file.
- Pinned `govulncheck v1.6.0` and `actionlint v1.7.12` are installed under a
  temporary task-specific GOBIN. Go 1.26.8 resolves in the candidate worktree.
- Docker 29.6.2 and Buildx 0.35.0 are available; the daemon is responsive and
  the pinned Dockerfile frontend plus Ubuntu 24.04 base image resolve.

## Candidate and validation results

- Candidate commit: `e6f519ade6ad4fa758e8924e66e9a5a1347291a0`.
- Candidate tree: `db9356712ed765f820b29038f4ff399c4549877f`.
- The candidate is the approved historical v0.1.1 payload plus only the Go
  toolchain/supply-chain-lock/documentation patch to Go 1.26.8.
- The official Go release policy supports the two newest major release lines;
  Go 1.26.8 is the latest patch in the selected supported 1.26 line at this review.
  The official Linux amd64 archive SHA-256 is
  `d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b`.
- `VERSION=0.1.1 make rc-check` passed: C and Go race tests, 78.3% Go
  statement coverage, 5,000 parser fuzz iterations, E2E (6 packets and 5
  alerts), distribution archive/checksum, Docker image-content smoke and
  runtime-health smoke.
- `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check` passed with all 9
  locked fixture/license hashes, actionlint v1.7.12, and zero reachable
  vulnerabilities under govulncheck v1.6.0. The scanner reported one
  vulnerability in a required module that candidate code does not call.
- `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate` passed.
  The RC knowledge gate passed 33 tests; current-main `make knowledge-check`
  passed 33 tests. Documentation, JSON, and diff checks passed.
- The local linux/amd64 archive is 9,908,296 bytes with SHA-256
  `67d02e15a3272e22ca4e86bba9fdd0c6fa02bfa6524bb3e355f8833801526d0b`.
  It is distinct from the later workflow-produced release archive.

## Tag and publication results

- Replaced and resigned the local tag with object
  `cbe602ff997c14a375f89acad00ab4d572fa49de`; the SSH signature verifies with
  the historical release key and it peels exactly to the candidate above.
- Pushed only `refs/tags/v0.1.1`. GitHub SSH port 22 closed the first
  connection; authenticated SSH-over-443 succeeded. A transient verification
  fetch failure was resolved by retrying the same fetch. Fetched tag object,
  signature, and peeled candidate match the local ref exactly.
- Release run `36889806244` and Docker Publish run `36889806404` both succeeded
  for the exact candidate SHA.
- GitHub Release is published, non-draft and non-prerelease. Its
  `netsentry-0.1.1-linux-amd64.tar.gz` asset is 9,905,076 bytes with SHA-256
  `6bbeb5b680f2d94e05ef27b27dee875fd454eb96497ed82e254e67f33b92b8e8`; the
  paired 101-byte checksum file verifies the downloaded archive. The release
  archive is distinct from the local build artifact.
- GHCR tags `v0.1.1` and `0.1.1` resolve to the same index digest
  `sha256:f4aae2de10c7553c011b05cad8ba86f7e3e3ec9265507b3f36744ea2d26321be`;
  the `linux/amd64` manifest is
  `sha256:e55caf21991aac2c126e1660cfcb4ab01f061654b750a70b750ce5615c75306e`.

## Verified audit delivery

Audit commit `1808fcdda909a1432721b54bea2762a8f75c409a` is pushed and freshly
fetched; HEAD, origin/main and FETCH_HEAD match. The post-fetch knowledge gate
passed 33 tests in the publication turn. Exact range
`5dced1bc9576f769d770a227d3989fbe0c0f4ea4..1808fcdda909a1432721b54bea2762a8f75c409a`
is synchronized and its iteration note, full-index row and MOC link verified.
The next trigger reproduced the recorded 340-file Markdown hash
`bce2ac3d5b24a92df4f06277261f00d9b3afc2a7f9342cfb40c4b32f969c953a`.
These completed actions must not be repeated as pending work.

The original scope mentioned a candidate-branch push; actual publication pushed
only the exact tag, which retains the candidate and triggers both workflows.
No candidate branch publication was needed. The older phrase "newest supported
line" was imprecise: 1.26.8 belongs to the selected supported 1.26 line; the
[official release history](https://go.dev/doc/devel/release) identifies 1.27 as
the newer major line. This correction changes no release object or test result.


## Selected delivery-record closure (next Oct 1 trigger)

The fetched baseline is `1808fcdda909a1432721b54bea2762a8f75c409a`.
The prior audit commit and its exact Vault range are already delivered; the
remaining work is one docs-only delivery-record closure of R90-59.

- **Scope:** reconcile this plan, active task state, latest roadmap audit and
  current stable Vault prose with verified delivery; correct the release-line
  description in release evidence, and define the next main-toolchain increment
  from the verified support-policy gap. Preserve historical
  iteration notes and all published release objects.
- **Acceptance/evidence:** verify the prior full-SHA range, generated note,
  index/MOC and 340-file Vault hash; remove obsolete commit/push/sync resume
  instructions; confirm every unfinished row has a complete contract; validate
  documentation, task JSON, roadmap multisets and knowledge checks; push/fetch
  this one closure and synchronize its exact range with idempotent replay.
- **Risk/stop:** stop on remote/ref divergence, missing historical evidence,
  unexpected user changes, ambiguous validation, or a required new authority.
- **Non-goals:** no runtime/toolchain change, release/tag mutation, acceptance
  execution, or next increment. R90-75 stays departmental. The prior scoped
  R90-59 knowledge-check authorization covers this remaining closure.
- **Deviation:** the previous turn verified delivery and Vault but left its
  plan/state at a pre-delivery checkpoint. Current stable notes only partially
  superseded the old publication blocker. This closure records those facts and
  repairs current handoff prose without repeating the release.


## Delivery-record closure evidence

- Prior audit push/fetch, exact Vault note/index/MOC and remote tag object/target
  verified directly; SSH-over-443 resolved the port-22 tag-lookup failure.
- Reconciled 11 current stable notes. All 293 pre-existing immutable iteration
  notes retain their original SHA-256 values. Replaying the identical prior
  range preserves the reconciled 340-file Markdown hash
  `17f197976d7ea921c5c3491b6c1de2be1d1e14fbdd30ea379421f3c2a8b84953`.
- `make docs-check`, 33-test `make knowledge-check`, release-evidence gate and
  `git diff --check` pass. All 148 task-state JSON files parse; all 135 roadmap
  rows and Definitions match as unique complete multisets. No behavioral suite
  rerun is needed for this documentation closure; the candidate's historical
  full-RC results are preserved without claiming main SLO validation.
- R90-131 is ready and unstarted. R90-75 still requires departmental evidence.
  The next trigger should verify the fetched closure commit and its exact Vault
  range, then audit/select the next item; no completed publication or audit
  operation is a pending task.
