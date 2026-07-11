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

Blocked before tagging v0.1.0:

- Sustained external C fuzz evidence must be recorded and reviewed.
- Realistic sanitized pcap corpus pressure/query evidence must be recorded and reviewed.
- Version tag `v0.1.0` must be created from the pushed passing release commit, then the checked-in GitHub Release and GHCR workflows must publish the named assets successfully.

The release gate reads `docs/evidence/release-v0.1.0.md` and fails closed
unless the reviewed public record has an approved final decision, at least
1,000,000 fuzz iterations with zero crashes and no ASan findings, passed
realistic sanitized pcap and query evidence, and all sensitive-information
review fields set to `no`. It does not create or approve evidence.

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
production-derived traffic gate without an approved exception.

Use an external or generated corpus for pressure evidence only after the
traffic-pressure gate is approved for the current iteration:

```bash
PCAP_CORPUS=/tmp/netsentry-sanitized-corpus make e2e-corpus-pressure
```

Run the standard local release-candidate bundle:

```bash
make rc-check
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

## Temporary Fuzz-Only Gate Exception

The default release gate requires both sustained external C fuzz evidence and
realistic pcap pressure/query evidence. A project administrator may create a
temporary exception to collect fuzz evidence first. The exception must be
recorded in the active `docs/plans/task-*.md` and `docs/tasks/task-state-*.json`
with approver, UTC approval time, exact scope, expiry/iteration, and evidence
location. The exception may defer traffic pressure, but it must not mark that
gate passed or permit `v0.1.0` tagging. Synthetic generated pcaps and no-corpus
fuzz runs remain auxiliary regression evidence.

## Release Checklist

- `make rc-check` passes locally.
- `make release-gate` passes against the reviewed public evidence record.
- Sudo Docker RC validation passes where Docker is part of the release gate.
- Synthetic extended pressure and no-corpus ASan fuzz checks pass as auxiliary regression signals.
- Sustained external fuzz run records zero crashes for the approved campaign window.
- Realistic sanitized pcap corpus run records packet, alert, and query evidence.
- README, changelog, API docs, and development docs match the final behavior.
- No local-only evidence, private corpus paths, credentials, or generated release archives are staged.
- The release commit is pushed only to the approved `decline-llc/netsentry` remote.
- Version tag `v0.1.0` is created from a passing commit.
- The checked-in GitHub Release and named registry image workflows complete successfully from the tag.

## Rollback Notes

If any gate fails after tagging but before publication, stop publication and fix forward on a new commit before creating a replacement tag. If a registry image or GitHub Release has already been published, record the failed artifact, delete or mark it as withdrawn according to project owner direction, and publish a corrected artifact from a new passing tag.
