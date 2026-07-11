# Release Evidence: v0.1.0

> This record is reviewed for v0.1.0 under the scoped exception in
> `docs/audit/release_exception_v0.1.0.yaml`. The pcap evidence is synthetic,
> not production-derived; the exception expires before v0.1.1.

## Metadata

- Release: v0.1.0
- Commit: pending final release commit
- Evidence record date: 2026-07-11
- Reviewer: user (explicit approval in conversation)
- Final decision: approved

## Local RC Validation

- Command: `make shell-check python-check docs-check deps-check e2e-smoke`
- Date: 2026-07-11
- Status: pass
- Docker mode: latest sudo Docker RC baseline previously passed
- Coverage summary: 74.2% latest recorded baseline
- Notes: No runtime core module changed in the exception-only release closure.

## Sustained External C Fuzz Evidence

- Command shape: `FUZZ_CORPUS=/approved/local/corpus make fuzz-sustained`
- Date: 2026-07-11
- Status: pass
- Iterations or duration: 1000000 iterations
- Corpus description: reviewed local fuzz inputs; paths intentionally omitted
- Corpus paths included: no
- Corpus files: 6
- Crashes: 0
- ASan findings: no
- Reviewer decision: approved
- Notes: Synthetic/local fuzz evidence; no ASan crash or finding reported.

## Realistic Sanitized Pcap Corpus Evidence

- Command shape: `PCAP_CORPUS=/approved/local/corpus make e2e-corpus-pressure`
- Date: 2026-07-11
- Status: pass
- Corpus description: 100-set deterministic synthetic corpus; exception-backed and not production-derived
- Evidence class: synthetic
- Production-derived corpus: no
- Exception applied: docs/audit/release_exception_v0.1.0.yaml
- Corpus paths included: no
- Pcap files: 600
- Packets processed: 2200
- Alerts generated: 2400
- Parse errors: 0
- Dropped packets: 0
- UDS write errors: 0
- Query evidence: pass
- Reviewer decision: approved
- Notes: 231.957 seconds enhanced-evidence run; 9.4845 packets/sec; 10.3467 alerts/sec; sampled peak RSS 29856 KiB; engine error-log lines 0; alerts query snapshot retained in local evidence.

## Tag Publication Verification

- Tag: pending
- Tag commit: pending
- GitHub Release workflow: pending tag push
- Release asset: pending tag push
- Release checksum: pending tag push
- GHCR workflow: pending tag push
- Image: pending tag push
- Reviewer decision: approved for tag-triggered verification
- Notes: This record must be amended after tag-driven workflows complete.

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
