# NetSentry v0.1.0 Release Readiness

> Status: release-candidate preparation. This file tracks public release gates and evidence commands. It must not contain private pcap paths, private fuzz corpora, credentials, or local operator notes.

## Current Gate State

Ready:

- Native Makefile release-candidate bundle is wired through `make rc-check`.
- GitHub Actions release-candidate checks are present.
- GitHub Release publication workflow is present for version tags and uploads the packaged tarball plus checksum.
- GHCR publishing workflow is present for version tags or explicit manual publishing.
- Local release archive packaging is available through `make dist`.
- Local `make release-artifacts VERSION=0.1.0` validates release-version format before building publishable archive assets.
- Local Docker image build is available through `make docker-build`.
- The local `origin` fetch and push URLs use the project SSH standard `git@github.com:decline-llc/netsentry.git`.
- Latest local full sudo Docker RC validation passed on 2026-07-08, including Docker build, image content smoke, and runtime `/api/health` smoke.
- Latest local non-Docker RC validation passed on 2026-07-10 with `SKIP_DOCKER=1 make rc-check`, covering syntax, docs, Python, config, dependencies, C/Go tests, race tests, coverage 74.2%, ASan fuzz smoke, e2e smoke, dist archive smoke, and release notes smoke.
- Synthetic extended validation passed on 2026-07-10: `PRESSURE_REPEATS=10000 PRESSURE_WAIT_ATTEMPTS=1200 make e2e-pressure` processed 60,000 packets, generated 50,000 alerts, and returned 5 aggregated rows in 214.956 seconds. This is repeat-pcap synthetic evidence only.
- Synthetic deterministic ASan parser fuzz passed on 2026-07-10: `FUZZ_SUSTAINED_ITERATIONS=1000000 make fuzz-sustained` completed 1,000,000 iterations with zero corpus files and no reported crash. This is no-corpus synthetic evidence only.
- Standardized sanitized synthetic corpus generation is available through `make gen-sanitized-corpus`; it writes three deterministic Ethernet pcaps, three matching pcapng files, and a manifest outside the repository by default.
- Approved local fuzz evidence passed on 2026-07-11: `FUZZ_CORPUS=/tmp/netsentry-fuzz-pcap-inputs make fuzz-sustained` completed 1,000,000 ASan mutation iterations plus 6 pcap/pcapng inputs with zero reported crashes. The corpus path is redacted in local evidence. This is approved local synthetic evidence, not external-corpus release evidence.
- The tag publication workflows now run `make release-gate` after `make rc-check` and before building release assets or logging in to GHCR.
- The approved v0.1.0 exception in `docs/audit/release_exception_v0.1.0.yaml` scopes out only real production-derived pcap evidence and expires before v0.1.1.
- The 2026-07-12 supply-chain gate pins the Go CI toolchain to `go1.25.12`, pins every third-party Action to a reviewed full commit SHA, validates all workflows with `actionlint v1.7.12`, scans reachable Go code with `govulncheck v1.6.0`, and fetches/verifies all 9 locked external fixture/license files only in an ephemeral directory.
- The v0.1.1 pre-evidence quality baseline passed on 2026-07-15: the pinned supply-chain check verified 9/9 locked external fixture/license hashes and found zero reachable Go vulnerabilities; `SKIP_DOCKER=1 make rc-check` passed with 75.4% Go coverage, parser fuzz, e2e smoke, and distribution checks. This is local, non-Docker validation only and does not provide production-derived traffic evidence or v0.1.1 release approval.
- R90-04 public-real-traffic validation passed on 2026-07-16 using one locally re-sanitized MAWI samplepoint-B trace: 544,525 packets processed with zero capture parse errors, drops, UDS write errors, or engine error-log lines. The reviewed, path-redacted record is [`r90-04-public-traffic-20260716.md`](evidence/r90-04-public-traffic-20260716.md). It is approved for R90-04 only; it neither approves a release nor satisfies later production-derived requirements.

v0.1.0 publication result:

- The signed `v0.1.0` tag, GitHub Release assets, tag-triggered Docker workflow, and public `ghcr.io/decline-llc/netsentry:v0.1.0` manifest were verified on 2026-07-11. Do not recreate or move the immutable tag.
- The v0.1.0 exception does not apply to v0.1.1. Separately, the R90-04-only exception in `docs/audit/release_exception_r9004.yaml` permits anonymized public real-traffic PCAP evidence in place of internal production-derived PCAP evidence, but only after dedicated privacy, provenance, sanitization, and sensitive-metadata reviews. Synthetic/generated traffic remains prohibited; this does not waive R90-05, R90-06, or future production-derived-PCAP requirements.

The release gate reads `docs/evidence/release-v0.1.0.md` and fails closed
unless the reviewed public record has an approved final decision, at least
1,000,000 fuzz iterations with zero crashes and no ASan findings, passed
synthetic pcap and query evidence covered by a valid scoped exception, and all
sensitive-information review fields set to `no`. It does not create or approve
evidence.

## Evidence Commands

Generate the built-in sanitized simulation corpus:

```bash
make gen-sanitized-corpus CORPUS_DIR=/tmp/netsentry-sanitized-corpus
```

For larger synthetic pressure input, generate differentiated sets outside Git:

```bash
make gen-sanitized-corpus CORPUS_DIR=/tmp/netsentry-synthetic-100 CORPUS_SETS=100
```

The generated corpus is explicitly synthetic and cannot close the realistic
production-derived traffic gate. It is also prohibited under the R90-04 public-real-traffic exception.

Use a reviewed external corpus for R90-04 pressure evidence only after its
privacy, provenance, sanitization, and sensitive-metadata controls are approved.
Generated corpora remain synthetic auxiliary input and are prohibited under the
R90-04 exception:

```bash
PCAP_CORPUS=/tmp/netsentry-sanitized-corpus make e2e-corpus-pressure
```

Run the standard local release-candidate bundle:

```bash
make rc-check
```

Run the pinned CI supply-chain gate, including remote external-asset integrity:

```bash
(cd engine && go install golang.org/x/vuln/cmd/govulncheck@v1.6.0)
(cd engine && go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12)
SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check
```

Validate the reviewed evidence before a release workflow:

```bash
make release-gate
```

When Docker requires elevated privileges, pass Docker through sudo:

```bash
DOCKER="sudo docker" make rc-check
```

Run sustained C parser fuzz evidence:

```bash
make fuzz-sustained
FUZZ_CORPUS=/path/to/local-corpus make fuzz-sustained
```

Run realistic pcap corpus pressure evidence:

```bash
PCAP_CORPUS=/path/to/sanitized-pcaps make e2e-corpus-pressure
```

Synthetic extended validation (auxiliary only; not a release-gate substitute):

```bash
PRESSURE_REPEATS=10000 PRESSURE_WAIT_ATTEMPTS=1200 make e2e-pressure
FUZZ_SUSTAINED_ITERATIONS=1000000 make fuzz-sustained
```

The 2026-07-10 pressure run used the repository's generated six-packet repeat
pcap. The fuzz run used no external corpus. Both results are useful regression
signals, but neither closes the external fuzz or realistic sanitized pcap gates.

Create local release artifacts:

```bash
VERSION=0.1.0 make dist
make release-artifacts VERSION=0.1.0
IMAGE=netsentry:0.1.0 DOCKER="sudo docker" make docker-build
```

Publish the release from a passing tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Evidence Handling

- `docs/evidence/local/` is ignored by Git and is the default location for local JSON and Markdown evidence.
- Sanitized public evidence records should use `docs/evidence/release-evidence-template.md` and be committed only after review.
- Fuzz and corpus evidence summaries redact corpus paths by default; set `NETSENTRY_EVIDENCE_INCLUDE_PATHS=1` only for private local debugging evidence that will not be committed.
- Do not commit raw private pcaps, external fuzz corpora, private corpus paths, local operator names, credentials, or environment-specific secrets.
- If evidence must be shared publicly, sanitize the pcap first with `make sanitize-pcap` and review the generated evidence text before committing.
- Record only summarized pass/fail status, run date, command shape, and non-sensitive metrics in public docs.

## v0.1.0 Scoped Gate Exception

The approved exception is recorded in
`docs/audit/release_exception_v0.1.0.yaml`. It applies only to the missing
real production-derived pcap requirement for v0.1.0, expires before v0.1.1,
and does not waive fuzz, ASan, synthetic pressure, query, sensitive-information,
or general quality checks. Real business-traffic corpus evidence remains a
mandatory v0.1.1 follow-up.

## R90-04 Scoped Public-Traffic Exception

`docs/audit/release_exception_r9004.yaml` applies only to R90-04. It permits an anonymized public real-traffic dataset instead of internal production-derived PCAP evidence only when the public evidence record identifies `public-anonymized-real`, references the exception and R90-04, and records approved privacy review, provenance validation, sanitization review, and sensitive-metadata screening. It expires when R90-04 completes and does not alter any subsequent increment. Synthetic or generated traffic is prohibited.

The exception is recorded as expired at R90-04 completion commit
`009b2a03776987359661c4ab2776f5d04820db34`. The release gate rejects it even
when every R90-04 review field is otherwise valid, preventing reuse for R90-05,
tag publication, or image publication.

## Release Checklist

- `make rc-check` passes locally.
- `SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check` passes with immutable Action refs, the locked Go toolchain, zero reachable vulnerabilities, and 9/9 external hashes.
- `make release-gate` passes against a reviewed release record using only an
  exception valid for that release or production-derived PCAP evidence; the
  expired R90-04 record cannot satisfy this gate.
- Sudo Docker RC validation passes where Docker is part of the release gate.
- Synthetic extended pressure and no-corpus ASan fuzz checks pass as auxiliary regression signals.
- Approved v0.1.0 exception record is present and scope-limited.
- Synthetic corpus pressure run records packet, alert, resource, and query evidence.
- Real production-derived pcap corpus is scheduled before v0.1.1.
- README, changelog, API docs, and development docs match the final behavior.
- No local-only evidence, private corpus paths, credentials, or generated release archives are staged.
- The release commit is pushed only to the approved `decline-llc/netsentry` remote.
- Version tag `v0.1.0` is created from a passing commit.
- The checked-in GitHub Release and named registry image workflows complete successfully from the tag.

## Rollback Notes

If any gate fails after tagging but before publication, stop publication and fix forward on a new commit before creating a replacement tag. If a registry image or GitHub Release has already been published, record the failed artifact, delete or mark it as withdrawn according to project owner direction, and publish a corrected artifact from a new passing tag.
