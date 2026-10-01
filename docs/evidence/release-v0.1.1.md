# Release Evidence: v0.1.1

> R90-05 evidence review package. The PCAP corpus is synthetic and not
> production-derived. It is accepted only under the exact digest-scoped
> exception in `docs/audit/release_exception_r9005.yaml`.

## Metadata

- Release: v0.1.1
- Commit: `6c3f9ef276c99c13aa9e985b8c849bb5f0791752`
- Evidence record date: 2026-07-16
- Reviewer: user explicitly approved v0.1.1 final release-gate acceptance
- Final decision: approved

## Local RC Validation

- Command: `VERSION=0.1.1 make rc-check`
- Date: 2026-07-16
- Status: pass
- Docker mode: full Docker build, image-content smoke, and runtime health smoke passed
- Coverage summary: 75.4% Go statement coverage
- Notes: Evidence regressions, parser fuzz smoke, e2e smoke, distribution archive, checksum, and release-notes smoke passed.

## Sustained External C Fuzz Evidence

- Command shape: `FUZZ_CORPUS=/approved/local/corpus make fuzz-sustained`
- Date: 2026-07-11
- Status: pass
- Iterations or duration: 1000000 iterations
- Corpus description: previously reviewed local fuzz inputs; paths intentionally omitted
- Corpus paths included: no
- Corpus files: 6
- Crashes: 0
- ASan findings: no
- Reviewer decision: approved
- Notes: Existing reviewed 1,000,000-iteration ASan evidence; no crash or finding reported.

## Realistic Sanitized Pcap Corpus Evidence

- Command shape: `PCAP_CORPUS=/approved/local/corpus make e2e-corpus-pressure`
- Date: 2026-07-16
- Status: pass
- Corpus description: reviewed synthetic controlled test traffic; not production-derived
- Evidence class: synthetic
- Production-derived corpus: no
- Exception applied: docs/audit/release_exception_r9005.yaml
- Exception increment: R90-05
- Privacy review: approved
- Provenance validation: approved
- Sanitization review: approved
- Sensitive metadata screening: approved
- Evidence manifest: reviewed path-redacted manifest; local path omitted
- Corpus paths included: no
- Pcap files: 1
- Packets processed: 7500
- Alerts generated: 0
- Parse errors: 0
- Dropped packets: 0
- UDS write errors: 0
- Query evidence: pass
- Reviewer decision: approved
- Manifest integrity verified: yes
- Notes: The reviewed corpus is 711,108 bytes with SHA-256 `509e940bc275d1972c09a4d9fd061e942516e22a0931d44eb9eb24deb7c66e68`. Pressure validation processed all 7,500 packets in 0.203 seconds at approximately 37,014 packets/second with sampled peak RSS 19,084 KiB and zero engine error-log lines.

## Tag Publication Verification

- Tag: signed annotated `v0.1.1`, pushed and fetched
- Tag object: `cbe602ff997c14a375f89acad00ab4d572fa49de`
- Tag commit: `e6f519ade6ad4fa758e8924e66e9a5a1347291a0`
- Signature: verified with release SSH key `SHA256:lanK75hksvHVuDmY55rdL1CJVqj4ZjqgBuE5/kMm4ZU`
- GitHub Release workflow: passed, run `36889806244`
- Release: published 2026-10-01, non-draft and non-prerelease
- Release asset: `netsentry-0.1.1-linux-amd64.tar.gz`, 9,905,076 bytes, SHA-256 `6bbeb5b680f2d94e05ef27b27dee875fd454eb96497ed82e254e67f33b92b8e8`
- Release checksum asset: `netsentry-0.1.1-linux-amd64.tar.gz.sha256`; downloaded checksum verified against the archive
- GHCR workflow: passed, run `36889806404`
- GHCR tags `v0.1.1` and `0.1.1`: both resolve to index `sha256:f4aae2de10c7553c011b05cad8ba86f7e3e3ec9265507b3f36744ea2d26321be`; `linux/amd64` manifest is `sha256:e55caf21991aac2c126e1660cfcb4ab01f061654b750a70b750ce5615c75306e`
- Reviewer decision: publication verified against the exact authorized candidate
- Notes: The independently generated local candidate archive (9,908,296 bytes, SHA-256 `67d02e15a3272e22ca4e86bba9fdd0c6fa02bfa6524bb3e355f8833801526d0b`) differs from the workflow-produced release asset and is not substituted for it.

## R90-59 Candidate Revalidation

- Date: 2026-10-01
- Candidate: `e6f519ade6ad4fa758e8924e66e9a5a1347291a0`, tree `db9356712ed765f820b29038f4ff399c4549877f`; based on the approved historical v0.1.1 payload with only the Go toolchain/lock/documentation patch
- Toolchain: Go `1.26.8`, the latest patch in the newest supported release line at review; official Linux amd64 archive SHA-256 `d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b`
- Behavioral validation: full `VERSION=0.1.1 make rc-check` passed, including C and Go race suites, 78.3% Go statement coverage, 5,000 parser fuzz iterations, E2E (6 packets/5 alerts), distribution archive/checksum, Docker image contents and runtime health
- Supply-chain validation: fetched all 9 locked fixture/license assets with exact hashes; actionlint and pinned `govulncheck v1.6.0` passed with zero reachable vulnerabilities. One vulnerability exists in a required module but is not called by the candidate.
- Release gate: `RELEASE_EVIDENCE=docs/evidence/release-v0.1.1.md make release-gate` passed
- Knowledge validation: candidate RC's knowledge gate passed 33 tests; current main `make knowledge-check` passed 33 tests
- Local candidate archive: `netsentry-0.1.1-linux-amd64.tar.gz`, Linux amd64, 9,908,296 bytes, SHA-256 `67d02e15a3272e22ca4e86bba9fdd0c6fa02bfa6524bb3e355f8833801526d0b`; distinct from the 9,905,076-byte workflow release asset above

## Sensitive Information Review

- Raw pcaps staged: no
- Fuzz corpus files staged: no
- Private corpus paths present: no
- Credentials or tokens present: no
- Local operator notes present: no
- Generated archives staged: no

## Final Release Gate Decision

- Sustained external fuzz evidence reviewed: yes
- Realistic sanitized pcap corpus evidence reviewed: yes
- Local RC validation reviewed: yes
- Tag publication verified: yes
- Approved for release: yes
