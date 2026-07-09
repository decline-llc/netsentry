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
- Latest local full sudo Docker RC validation passed on 2026-07-08, including Docker build, image content smoke, and runtime `/api/health` smoke.

Blocked before tagging v0.1.0:

- Sustained external C fuzz evidence must be recorded and reviewed.
- Realistic sanitized pcap corpus pressure/query evidence must be recorded and reviewed.
- Version tag `v0.1.0` must be created from the pushed passing release commit, then the checked-in GitHub Release and GHCR workflows must publish the named assets successfully.

## Evidence Commands

Run the standard local release-candidate bundle:

```bash
make rc-check
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
- Fuzz and corpus evidence summaries redact corpus paths by default; set `NETSENTRY_EVIDENCE_INCLUDE_PATHS=1` only for private local debugging evidence that will not be committed.
- Do not commit raw private pcaps, external fuzz corpora, private corpus paths, local operator names, credentials, or environment-specific secrets.
- If evidence must be shared publicly, sanitize the pcap first with `make sanitize-pcap` and review the generated evidence text before committing.
- Record only summarized pass/fail status, run date, command shape, and non-sensitive metrics in public docs.

## Release Checklist

- `make rc-check` passes locally.
- Sudo Docker RC validation passes where Docker is part of the release gate.
- Sustained external fuzz run records zero crashes for the approved campaign window.
- Realistic sanitized pcap corpus run records packet, alert, and query evidence.
- README, changelog, API docs, and development docs match the final behavior.
- No local-only evidence, private corpus paths, credentials, or generated release archives are staged.
- The release commit is pushed only to the approved `decline-llc/netsentry` remote.
- Version tag `v0.1.0` is created from a passing commit.
- The checked-in GitHub Release and named registry image workflows complete successfully from the tag.

## Rollback Notes

If any gate fails after tagging but before publication, stop publication and fix forward on a new commit before creating a replacement tag. If a registry image or GitHub Release has already been published, record the failed artifact, delete or mark it as withdrawn according to project owner direction, and publish a corrected artifact from a new passing tag.
