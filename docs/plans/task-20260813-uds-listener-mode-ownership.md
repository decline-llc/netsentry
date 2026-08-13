# Task Plan: R90-94 UDS listener mode and ownership binding

## Metadata

- Timestamp: 2026-08-13T10:39:45-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `50a98397c1145b0915458ab662247b4a68542b27`

## Goal

Apply the configured Unix-socket mode through the listener created by `Start`
and publish pathname ownership only when a later non-following inspection still
identifies that created listener.

## Scope

- Create and mode the listener inside a private same-filesystem directory, then
  publish that verified identity without overwriting the configured pathname.
- Compare the created listener identity with a non-following pathname snapshot
  before publishing `r.ln` and `r.socket`.
- Add deterministic, receiver-local synchronized regressions that replace the
  pathname after listener creation with a regular file, symlink, or Unix
  listener.
- Preserve replacement bytes, modes, identities, and replacement-listener
  service while retaining ordinary configured-mode and shutdown cleanup.
- Reconcile receiver lifecycle documentation, roadmap, task state, delivery,
  and local Vault authority.

## Non-Goals

- Do not change the UDS protocol, configuration schema, public API, stale/active
  peer policy, socket-mode defaults, or general cross-process ownership.
- Do not add an exported/public test seam, dependency, privileged filesystem
  control, fixed-sleep synchronization, or platform-specific unsafe code.
- Do not access private/external data or authorize tags, GitHub Releases, GHCR,
  workflows, performance policy, or any publication action.
- Do not start a later roadmap increment.

## Risks

- Private-path publication must remain on the target filesystem, preserve Unix
  listener connectivity, and leave no staging directory after success/failure.
- Checking only that the pathname is a socket can capture a concurrent
  replacement listener and later remove another service during shutdown.
- A replacement regression before `net.Listen`, during stale probing, or after
  ownership publication would not exercise this increment's promised boundary.
- Cleanup on a failed startup must close the detached created listener without
  removing or mutating the replacement pathname.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Configured mode is applied to the created listener identity | Direct ordinary startup assertion plus private-created-identity implementation review |
| Ownership is published only for the created listener | Private source and non-following published-path identities compared before `r.ln`/`r.socket` assignment, with private identity anchor retained through shutdown |
| Regular-file replacement is preserved | Synchronized post-listen regression compares bytes, mode, identity, startup rejection, and nil installed listener |
| Symlink replacement is preserved without following | Synchronized regression compares link identity/target plus target mode and startup rejection |
| Replacement listener remains separate and live | Synchronized regression compares identity/mode, proves service round trip, and confirms receiver installs no listener |
| Established lifecycle remains compatible | Startup cancellation, stale/active handling, configured mode, owned shutdown cleanup, receiver race, full native, and E2E checks |
| Delivery records are recoverable | Plan/state/roadmap/docs, exact commit and fetched remote SHA, exact-range Vault note/index/MOC, and stable-note reconciliation |

## Validation

- Run the direct post-listen replacement and ordinary mode/cleanup tests.
- Run the acceptance set twenty times uncached under the race detector, then the
  complete receiver package uncached under race.
- Preflight repository-pinned tools, then run the fail-fast native repository
  checks, E2E smoke, documentation, and knowledge gates applicable to the
  changed surface.
- Parse every task-state JSON and compare complete roadmap row/Definition
  multisets with equal counts and no duplicates or asymmetry.
- Run `git diff --check`; review exact intended scope and anchored credential or
  sensitive-path matches before commit.
- Push without force or tags, fetch and require
  `FETCH_HEAD == HEAD == origin/main == new`, rerun `make knowledge-check`, and
  synchronize the exact full-SHA range to the sole local Vault.

## Authority Boundaries

This trigger authorizes only R90-94 implementation, direct regressions,
compatibility documentation, task-state/roadmap evidence, a focused feature
commit, push of `main` without force or tags, and local Vault synchronization.
It does not authorize protocol/configuration/public API changes, private or
external input, performance policy, tags/releases/images/registry publication,
workflow dispatch, or any later increment.

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills, rolling roadmap,
  repository knowledge contract, latest R90-93 plan/state, and current source.
- Fetched `origin/main` and verified a clean
  `HEAD == origin/main == 50a98397c1145b0915458ab662247b4a68542b27`
  baseline.
- Verified the exact R90-93 feature/closure chain, their sole-local-Vault
  notes, full-index rows, MOC links, and stable R90-94-ready authority.
- Reviewed Jul 20 through Aug 13 in four phases: 61, 38, 46, and 14 commits;
  resolved historical validation/transport deviations remain recorded and no
  missing closure, stale stable claim, or unresolved local result changes
  priority.
- Parsed the forward queue: R90-59 and R90-75 retain explicit external
  blockers, while R90-94 is the sole dependency-ready local increment and has
  complete dependency, window, risk, acceptance, validation, and stop fields.
- Confirmed `Start` calls pathname-following `os.Chmod` after `net.Listen` and
  captures a later non-following pathname without proving it is the created
  listener; current tests do not replace the pathname in that interval.

## Implementation Checkpoint

- `Start` now creates the Unix listener under a private mode-0700 directory on
  the configured pathname's filesystem, applies and verifies the configured
  mode there, and publishes the same socket inode with a non-replacing hard
  link. A non-following source/path comparison must pass before `r.ln` and
  `r.socket` are assigned.
- The private hard link remains as an identity anchor while the receiver runs.
  Shutdown completes identity-bound public-path removal and private-anchor
  cleanup before listener close can release waiters; failed startup cleans the
  detached listener and private directory while preserving any competing
  public pathname.
- Three receiver-local synchronized regressions install a regular file,
  symlink, or live Unix listener after private listener creation but before
  publication. Each proves startup rejection, nil published receiver state,
  replacement identity/mode preservation, no private artifact after failure,
  and the specific bytes, symlink target, target mode, or live service promised
  by R90-94. A fourth regression proves ordinary configured mode and owned
  shutdown cleanup.
- The first focused design used `UnixListener.File` plus descriptor `Chmod` and
  `Stat`. Mode application reached the socket, but descriptor metadata did not
  identify the filesystem pathname inode, so ordinary ownership comparison
  failed. Delivery remained blocked; private same-filesystem publication
  replaced that design and the complete corrected focused sequence restarted.
- The corrected direct set and complete receiver package pass normally. The
  acceptance/compatibility set passes twenty times uncached under race, and the
  complete receiver package passes uncached under race. Complete repository
  validation remains the delivery boundary.

## Validated Evidence

- The engine module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU
  Make 4.3, Bash 5.2.21, Python 3.12.3, curl 8.5.0, timeout 9.4, jq 1.7,
  pkg-config 1.8.1, and libpcap 1.10.4 were available before complete
  validation.
- The complete fail-fast repository chain passes: both native C test binaries,
  every Go package uncached under the race detector, E2E smoke with six packets
  processed, five alerts generated, and eight rules loaded, documentation
  checks, and all 33 knowledge tests.
- All 109 task-state JSON files parse; all 98 roadmap rows match exactly one of
  98 Definitions with equal raw counts, no duplicate identifiers, and no
  asymmetric identifiers.
- Every acceptance criterion reaches its direct planned boundary. The three
  replacement tests synchronize after private listener creation and prove
  preservation of bytes/modes, symlink identity/target, or live listener
  identity/service. The ordinary test directly proves configured mode,
  identity-bound shutdown cleanup, and absence of private artifacts.
- `gofmt`, `git diff --check`, exact seven-path scope, credential,
  sensitive-path, dependency, configuration, protocol, public API, release,
  and publication reviews pass. No dependency, configuration schema, protocol,
  public API, workflow, release artifact, private-data access, or external
  mutation was added.
- The rejected descriptor-stat attempt is fully recorded above. Every
  acceptance and repository sequence was rerun after the corrected design; no
  unresolved validation deviation remains. R90-94 awaits only exact staged
  review, feature delivery, fetched remote verification, and exact-range Vault
  synchronization.

## Delivery Results

- Feature commit:
  `2e03300e46f3df1f98e47f72bada5207cc2e8fc3` (`fix: bind UDS listener mode
  and ownership`). It contains exactly the seven validated source, test,
  documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags. The first verification fetch
  returned no usable ref or exit evidence and left synchronization blocked; an
  identical non-mutating retry freshly verified
  `FETCH_HEAD == HEAD == origin/main == 2e03300e46f3df1f98e47f72bada5207cc2e8fc3`
  with fast-forward ancestry from the recorded baseline. The post-fetch
  33-test knowledge gate passed.
- Exact range
  `50a98397c1145b0915458ab662247b4a68542b27..2e03300e46f3df1f98e47f72bada5207cc2e8fc3`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC and UDS prose was reconciled to private created-identity
  publication, retained runtime identity anchor, direct replacement
  preservation, and the completed R90-94 boundary without rewriting immutable
  iteration notes. Replaying the identical range preserved Vault content hash
  `9f66c134e78a28538734d1c0891009c4142e667c7172969f394facef09cee94b`.
- R90-94 is complete. No dependency-ready local increment remains; R90-59 and
  R90-75 retain their external blockers and were not started.

## Stop Conditions

Stop if deterministic proof requires fixed sleeps, privileged filesystem
control, an exported/public seam, following symlinks, a new dependency,
platform-specific unsafe code, protocol/configuration/public API authority,
private data, publication authority, or if validation remains ambiguous.
