# Release Evidence: v0.1.1

> R90-05 evidence review package. The PCAP corpus is synthetic and not
> production-derived. It is accepted only under the exact digest-scoped
> exception in `docs/audit/release_exception_r9005.yaml`.

## Metadata

- Release: v0.1.1
- Commit: pending focused R90-05 delivery commit
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

- Tag: not created
- Tag commit: not applicable
- GitHub Release workflow: not authorized
- Release asset: not published
- Release checksum: local distribution checksum passed
- GHCR workflow: not authorized
- Image: local `netsentry:0.1.1` validation image only
- Reviewer decision: publication remains pending separate authorization
- Notes: R90-05 does not authorize tagging or publication.

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
- Tag publication verified: no
- Approved for release: yes
