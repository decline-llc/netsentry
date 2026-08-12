# Task Plan: R90-92 UDS shutdown generation preservation

## Metadata

- Timestamp: 2026-08-12T09:10:00-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `29c291a7dffcc37caf0375910e1ad1c6ef0a54a4`

## Goal

Make receiver shutdown preserve a replacement Unix listener when its pathname
reuses the receiver socket's device/inode identity by requiring the captured
non-following change timestamp to match before owned-path removal.

## Scope

- Reuse the established `sameUnixSocketIdentity` generation check in shutdown
  cleanup instead of accepting device/inode equality alone.
- Add a direct real-filesystem regression that releases the owned listener,
  immediately rebinds until device/inode reuse is observed, invokes shutdown
  cleanup, and proves the replacement listener remains serviceable.
- Add direct fail-closed evidence for unavailable generation metadata.
- Preserve ordinary owned socket cleanup and the existing replacement regular
  file and symlink behavior.
- Reconcile current lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority for this increment.

## Non-Goals

- Do not change startup stale/active classification, listener auto-unlink,
  probe timing, context behavior, protocol, configuration, public API, peer
  trust/authentication, or general cross-process ownership.
- Do not add an exported or production-injected test seam, fixed sleeps,
  privileged filesystem control, retry policy, or pathname following.
- Do not access private/external data or perform tag, release, registry,
  workflow, performance-policy, or other publication mutation.

## Risks

- Device/inode equality can identify a new socket after immediate inode reuse;
  removing it would interrupt another listener and violate the preservation
  boundary.
- An unavailable or unsupported `Stat_t` generation signal must fail closed,
  while an over-broad comparison could leak the ordinary owned pathname.
- A regression that merely creates a different inode would not reach the
  promised reuse boundary or prove the generation check matters.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Shutdown rejects device/inode reuse | Direct regression requires `os.SameFile(owned, replacement)` and a different non-following change timestamp before cleanup |
| Replacement service survives cleanup | The same regression invokes `removeOwnedSocket`, confirms the replacement pathname identity, and completes a listener round trip |
| Missing generation metadata fails closed | Direct regression supplies owned metadata without a usable Unix `Stat_t` and proves the current pathname remains |
| Ordinary owned cleanup remains compatible | Existing startup/shutdown regression proves the receiver-created pathname is removed |
| Non-socket replacements remain compatible | Existing regular-file and symlink shutdown regressions remain direct and passing |
| Public lifecycle authority is current | Changelog, architecture, development, roadmap, task plan/state, and stable Vault UDS prose state the generation-bound shutdown rule |
| Delivery evidence is repository-complete | Focused ordinary and repeated uncached receiver race, complete native, E2E, docs, knowledge, JSON/Definition, diff, scope, and sensitive-information checks |

## Trigger Audit

- Read the `netsentry-next` and required `netsentry-roadmap` skills plus the
  active rolling roadmap before selection.
- After transient port-22 and SSH-over-443 failures, fetched through the
  documented IPv4 SSH-over-443 keepalive fallback and verified a clean
  `FETCH_HEAD == HEAD == origin/main == 29c291a7dffcc37caf0375910e1ad1c6ef0a54a4`.
- Verified the exact R90-91 feature/closure parent chain and exact three/three
  paths, completed plan/state, both immutable Vault notes, full-index rows,
  MOC links, and current stable MOC/UDS authority.
- Parsed all 106 prior task-state JSON files and verified all 96 prior roadmap
  rows and Definitions match as complete multisets without duplicates or
  asymmetry.
- Reviewed 155 Jul 20 through Aug 12 commits across four phases: 61 commits
  Jul 20-26, 38 Jul 27-Aug 2, 46 Aug 3-9, and 10 Aug 10-12. The last phase has
  two behavior-like changes, five delivery closures, and three other docs;
  only the R90-91 feature/closure follows the prior audit, with no missing
  record, stale stable authority, or unresolved validation result.
- Confirmed R90-59 and R90-75 retain their external blockers and R90-92 is the
  sole dependency-ready local increment.
- Confirmed shutdown captures non-following socket metadata but
  `removeOwnedSocket` accepts only Unix mode plus `os.SameFile`; current direct
  shutdown replacements cover a regular file and symlink but not a serviceable
  Unix listener that reuses device/inode identity.

## Validation

- Run the direct immediate-inode-reuse replacement-listener regression,
  unavailable-generation regression, ordinary owned cleanup, and regular-file
  and symlink replacement compatibility tests.
- Run that focused set at least twenty times uncached under the race detector
  from the `engine` Go module, then run the complete receiver package uncached
  under race.
- Preflight the module-selected Go toolchain and every repository-pinned local
  tool needed by the complete fail-fast chain.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets, including raw counts, duplicate absence, and asymmetric IDs.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, configuration, protocol, public API,
  release, and publication review.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun the knowledge gate, then
  synchronize and verify the exact full-SHA Vault range and current stable UDS
  authority.

## Authority Boundaries

This trigger authorizes only R90-92's shutdown generation comparison, direct
real-filesystem inode-reuse and unavailable-metadata regressions, compatibility
validation, current documentation and task-state reconciliation, commit/push,
and local Vault synchronization. It does not authorize startup classification,
protocol/configuration/public API changes, peer trust/authentication, private
input, performance policy, tag/release/image/registry publication, workflow
dispatch, or a later roadmap increment.

## Implementation Checkpoint

- `removeOwnedSocket` now requires the same non-following device, inode, and
  change timestamp captured for the receiver-owned socket; missing or changed
  Unix generation metadata returns without removing the current pathname.
- Startup now applies the configured socket mode before capturing ownership
  metadata. The first focused run exposed that the prior post-capture `chmod`
  legitimately advanced change time and caused ordinary owned cleanup to fail
  closed; moving the existing mutation before the snapshot restores cleanup
  without weakening generation identity.
- The direct real-filesystem regression closes and unlinks an owned listener,
  immediately rebinds until `os.SameFile` proves device/inode reuse while the
  change timestamp differs, invokes cleanup, verifies the replacement identity,
  and completes a listener round trip.
- A second direct regression supplies owned metadata without a usable Unix
  `Stat_t`, invokes cleanup, and proves the live pathname and service remain.
  Existing ordinary cleanup plus regular-file and symlink replacements pass.
- The first formatting command used repository-root paths from the `engine`
  module and stopped before tests. The corrected module-relative fail-fast
  sequence was rerun successfully; the existing skill already requires module
  path resolution, so this execution mistake does not warrant another rule.
- The ownership-snapshot ordering lesson is repeatable, so the separate local
  `netsentry-next` skill now requires capture after intended metadata mutations.
  No skill file is included in the repository increment.

## Validated Evidence

- The direct acceptance and compatibility set passed once normally, then 20
  times uncached under the race detector. The complete receiver package also
  passed uncached under race.
- The engine module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU
  Make 4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes: both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation
  checks, and all 33 knowledge tests.
- All 107 task-state JSON files parse; all 96 roadmap rows match exactly one of
  96 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers.
- Every acceptance criterion reaches its planned direct boundary. The real-
  filesystem test explicitly requires device/inode reuse plus changed
  generation before proving replacement identity and service; the unavailable-
  metadata test invokes cleanup directly; ordinary, regular-file, and symlink
  compatibility remains direct.
- `gofmt`, `git diff --check`, exact eight-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.
- The corrected path invocation and pre-snapshot `chmod` findings are recorded
  in the implementation checkpoint. No unresolved validation deviation
  remains; R90-92 awaits only exact staged review, feature delivery, fetched
  remote verification, and exact-range Vault synchronization.

## Delivery Results

- Feature commit:
  `b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e` (`fix: preserve replacement
  UDS socket generations`). It contains exactly the eight validated source,
  test, documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags through the documented IPv4
  SSH-over-443 transport. A fresh fetch verified
  `FETCH_HEAD == HEAD == origin/main == b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e`
  with fast-forward ancestry from the recorded baseline, and the post-fetch
  33-test knowledge gate passed.
- Exact range
  `29c291a7dffcc37caf0375910e1ad1c6ef0a54a4..b3ef17b8850c170b7f517fbb3e5eaa7c7fdf7c1e`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC/UDS prose was reconciled to the delivered generation-bound
  cleanup, ownership-snapshot ordering, and direct inode-reuse plus service
  evidence without rewriting immutable iteration notes. Replaying the
  identical range preserved Vault content hash
  `e368106cf93971de1a76bb46ed0753fb04338a2453f62d3452d277088eea7217`.
- R90-92 is complete. R90-59 and R90-75 retain their external blockers; no
  dependency-ready local increment remains and no later work was started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, privileged filesystem
control, or a public test seam; if non-following generation identity cannot
fail closed while retaining ordinary cleanup; if repeated/full validation is
ambiguous; or if completion needs private data, product policy, protocol,
configuration, public API, or publication authority.
