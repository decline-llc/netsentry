# CI repair and validated delivery — 2026-10-17

The user requests repair of recurring GitHub CI failures and complete local
validation before uploading. This explicitly authorizes the CI, scanner, runtime,
and knowledge checks needed for this repair, superseding the prior delegated-test
restriction for this task. Preserve independent R90-75 acceptance boundaries.

Baseline: main at 01a5e447aa5a26b436f868b448c1e31e4d4181fa, clean checkout.
The latest 25 GitHub runs failed; latest run 38876302871 fails its supply-chain
step before RC execution. Check annotations identify reachable net/http calls.
Last successful main CI: 37176164511 (2026-10-04). Feature and closure pushes
have repeatedly triggered the same failure; local compile-only evidence cannot
establish the scanner or RC result. Historical delegated records remain historical.

Plan and acceptance evidence:
1. Recover failure details and reproduce the scanner failure locally.
2. Verify the latest supported patch in the selected Go line from go.dev;
   update the exact toolchain and lock together if the scanner proves necessary.
3. Share one complete local/CI validation entry point; retain all security,
   correctness, Docker and publication gates. Document validation-before-push.
4. Preflight exact tools, Docker daemon/frontend/base, then run full supply-chain
   asset verification and RC checks, including uncached race/coverage, Python
   evidence/knowledge checks, C sanitizer fuzz, E2E, archive and Docker health.
   Diagnose and repair reproducible failures before any push.
5. Review complete diff and sensitive content. Commit all completed repair and
   evidence together, push once, fetch-verify and wait for that SHA's GitHub CI.
   Sync the exact pushed range to the existing local Vault. Record delivery
   locally without an additional remote docs-only push/CI run.

Risk: low for orchestration/pins; behavioral failures may require a bounded
regression repair supported by reproduction. Stop upload on failed/unavailable
required validation or unexpected remote changes. Do not disable checks or
notifications, create tags/releases/images, contact others, or start roadmap work.

Diagnosis: the 1.26.8 local scanner reports GO-2026-6617, GO-2026-6613,
GO-2026-6612, GO-2026-6611, GO-2026-6608, GO-2026-6607 and GO-2026-6603,
all fixed in 1.26.9. GitHub step/annotations independently identify the scanner
and reachable HTTP paths; direct blob log retrieval hit a TLS timeout.
The official release feed lists 1.27.2 and 1.26.9; choose 1.26.9 to retain the
selected release line. Official Linux amd64 archive: 66,935,201 bytes,
SHA-256 `42d158b4d8f7b61ac0a830567c940a86098fb7aac52e467a5ebec03ef5cc2f8d`;
downloaded and verified before use. The replacement scanner has passed with
zero reachable vulnerabilities and all nine fixture/license hashes verified.

## Full-gate deviation: reproducible fixture failures

The first complete gate passed scanning but failed nine alert test declarations.
A separate uncached race run repeated the four representative failure boundaries
twice: subsecond-window healthy WAL observation, expired active-shard WAL
observation, mixed-prefix unscoped suppression, and path-encoding writes.
Six failing declarations share pathEncodingAlert's arbitrary noncanonical EventID;
two observer helpers unconditionally demand readonly_shm even after the last
writer removes clean WAL sidecars; one older suppression fixture supplies an
explicit all-empty RuleIDs list now correctly rejected by the rule-scope contract.

Bounded repair scope adds only four existing test files. Omit the optional event
ID so public normalization generates its canonical identity; condition independently
encoded read-only observer protection on existing sidecars; use the accepted empty
rule-ID list for an unscoped suppression. Preserve all path, row, artifact,
identity, caller-input and scope assertions. Re-run the complete affected test
families twice uncached under race, then restart the complete ci-check chain.
Do not weaken production validators or delete failing tests.

## Local acceptance checkpoint

Affected families passed twice uncached under race. The restarted complete
`make ci-check` passed with exit 0: fetched supply-chain gate, 42 evidence and
33 knowledge tests, all-package uncached race and coverage (82.3%), C tests,
both 5,000-iteration ASan fuzz targets, E2E, archive and unmodified Docker build,
image-content and runtime health smoke. No Docker skip or gate suppression.
All required local boundaries now pass; only final evidence/status prose changed
afterward and receives document/JSON/diff/sensitive-content review before commit.
Exact-SHA hosted verification and exact-range Vault synchronization are the
post-push acceptance steps; record those locally without an extra remote closure.
