# Release Decision Package: v0.1.1

## Decision

- Publication status: **hold**
- Release-gate acceptance: approved
- Tag creation: not authorized and not performed
- GitHub Release publication: not authorized and not performed
- GHCR publication: not authorized and not performed
- Required next authority: explicit final publication authorization for
  `v0.1.1` at the candidate commit below

This package completes R90-06 release-decision reconciliation. It does not
grant or imply publication authority.

## Proposed Release Candidate

- Version: `v0.1.1`
- Proposed tag: `v0.1.1`
- Candidate commit: `ad8a443b5020037c235419f5696c60988d2bba99`
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

The candidate contains the accepted evidence record plus subsequent delivery
reconciliation, the approved roadmap-window and PCAP gate waivers, and
version-to-evidence workflow binding. The complete RC, supply-chain audit, and
release gate were rerun on the candidate worktree.

## Local Artifact

- Platform: `linux/amd64`
- Archive: `netsentry-0.1.1-linux-amd64.tar.gz`
- Byte size: `9,561,869`
- SHA-256:
  `aa88bb4b25e9bb2418e2762f788ebad0159753ebe16868b33ae9d0253b981967`
- Checksum verification: pass
- Archive release-notes version verification: pass
- Repository handling: generated under ignored `dist/`; archive bytes and
  checksum file are not committed

The artifact was rebuilt while `HEAD` and fetched `origin/main` both resolved
to the proposed candidate commit.

## Candidate Validation

| Control | Result |
| --- | --- |
| `VERSION=0.1.1 make rc-check` | Pass, including Docker build, image-content smoke, and runtime health smoke |
| Go statement coverage | 75.4% |
| Parser fuzz smoke | Pass, 5,000 iterations |
| Release archive contents and checksum smoke | Pass |
| `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate` | Pass |
| Pinned supply-chain asset audit | Pass, 9 of 9 fetched assets matched |
| Reachable Go vulnerabilities | 0 |
| Workflow syntax and pin checks | Pass |
| Documentation, Python, evidence, and knowledge checks | Pass |

One supply-chain download attempt ended with a transient TLS EOF after six
matched assets. The unchanged command was rerun and completed with all nine
assets matching their locked byte sizes and SHA-256 digests.

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
explicitly authorize the version and candidate commit before any tag is
created. That authorization may then permit the tag-triggered GitHub Release
and GHCR workflows; it is outside R90-06.
