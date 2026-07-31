# Release Decision Package: v0.1.1

## Decision

- Publication status: **hold**
- Release-gate acceptance: approved
- Tag creation: not authorized and not performed
- GitHub Release publication: not authorized and not performed
- GHCR publication: not authorized and not performed
- Required next authority: explicit final publication authorization for
  `v0.1.1` at the candidate commit below

This package was refreshed by R90-58 after the completed storage and protocol
hardening sequence. It does not grant or imply publication authority.

## Proposed Release Candidate

- Version: `v0.1.1`
- Proposed tag: `v0.1.1`
- Candidate commit: `78cd78574e03c8f73ff68248eed2c409d6bca406`
- Candidate branch: `main`
- Remote verification: fetched `origin/main` matched the candidate commit
- Existing tags at decision time: `v0.1.0` only

If the proposed tag target changes, rebuild the artifact, rerun the applicable
release gates, and replace this decision package before publication.

## Reviewed Evidence Reconciliation

- Release evidence: `docs/evidence/release-v0.1.1.md`
- Evidence acceptance commit:
  `6c3f9ef276c99c13aa9e985b8c849bb5f0791752`
- Reviewer decision: the user explicitly approved v0.1.1 final release-gate
  acceptance
- PCAP policy: the later approved global waiver removes PCAP evidence from the
  current release gate; raw corpora, private paths, and local review materials
  remain excluded from Git and the Vault

The candidate contains the accepted evidence record, approved roadmap-window
and PCAP gate waivers, version-to-evidence workflow binding, and the completed
R90-07 through R90-56 hardening sequence. The complete RC, supply-chain audit,
and release gate were rerun from an isolated clean worktree at the exact
candidate commit.

## Local Artifact

- Platform: `linux/amd64`
- Archive: `netsentry-0.1.1-linux-amd64.tar.gz`
- Byte size: `9,760,241`
- SHA-256:
  `c68e09df46d24307c9a0d405a2724573f3382813a8b2611bdb5f3b7d8b068568`
- Checksum verification: pass
- Archive release-notes version verification: pass
- Repository handling: generated under ignored `dist/`; archive bytes and
  checksum file are not committed

The artifact was freshly rebuilt in the isolated candidate worktree. Its
detached `HEAD` and the source repository's fetched `origin/main` both resolved
to the proposed candidate commit; uncommitted R90-58 planning files were
therefore excluded from the artifact.

## Candidate Validation

| Control | Result |
| --- | --- |
| `VERSION=0.1.1 make rc-check` | Pass, including Docker build, image-content smoke, and runtime health smoke |
| Go statement coverage | 78.3% |
| Parser fuzz smoke | Pass, 5,000 iterations |
| Release archive contents and checksum smoke | Pass |
| `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate` | Pass |
| Pinned supply-chain asset audit | Pass, 9 of 9 fetched assets matched |
| Reachable Go vulnerabilities | 0 |
| Workflow syntax and pin checks | Pass |
| Documentation, Python, evidence, and knowledge checks | Pass |

The first combined validation attempt completed the full Docker RC, then
stopped before the supply-chain audit because the pinned `actionlint` tool was
not installed locally. Exact `actionlint v1.7.12` and `govulncheck v1.6.0`
builds were installed into a temporary directory using Go 1.25.12, and the
entire combined sequence was rerun from the beginning. The accepted rerun
passed all RC checks, fetched and matched all nine locked external assets,
reported zero reachable Go vulnerabilities, and passed the release gate.

## Publication Workflow Readiness

- A `v0.1.1` GitHub Release tag run resolves
  `docs/evidence/release-v0.1.1.md`, reruns the RC and release gate, builds the
  versioned archive, and publishes only after those steps pass.
- A `v0.1.1` Docker tag run resolves the same evidence record and version,
  reruns the RC and release gate, and publishes only after those steps pass.
- Manual Docker runs require a validated repository evidence path and do not
  push unless `push_image` is explicitly enabled.

## Publication Gate

The current decision is **hold**. A later formal publication gate must
explicitly authorize version `v0.1.1` and candidate commit
`78cd78574e03c8f73ff68248eed2c409d6bca406` before any tag is created. That
authorization may then permit the tag-triggered GitHub Release and GHCR
workflows; it is outside R90-58.
