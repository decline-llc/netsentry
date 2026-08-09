# Task Plan: R90-79 suppression-file replacement durability

## Metadata

- Timestamp: 2026-08-09T06:05:06-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `17a5809f83959714f8801fdfa7e613520e06dd14`

## Goal

Make canonical suppression-file replacement explicit and directly testable
across short write, permission, file-sync, close, rename, and parent-directory
durability boundaries while retaining the manager's serialized mutation and
keeping its active rules/filter consistent with every reported commit outcome.

## Scope

- Replace the suppression file through a temporary file in the target
  directory using an exact-length write, preserved mode, file sync, close,
  atomic rename, and containing-directory sync and close.
- Retain creation of missing parent directories and inject every replacement
  lifecycle operation through a manager-owned package-local seam rather than
  package-global mutable state.
- Classify persistence errors by whether rename committed the canonical file.
- Preserve exact prior canonical bytes, active rules, and active filter for
  every failure through rename, and remove the temporary file exactly.
- On a post-rename durability error, publish the already-compiled candidate
  rules/filter before returning an explicit API error that says the mutation
  was applied but crash durability could not be confirmed.
- Reconcile current suppression API, architecture, development, changelog,
  roadmap, task-state, and Vault authority.

## Defined Outcome Contract

| Boundary | Canonical file | Active rules/filter | API outcome |
| --- | --- | --- | --- |
| Validate/marshal/stat/directory-create/temp-create/write-short/write-error/chmod/file-sync/temp-close failure | Prior bytes remain | Prior state remains | `500 INTERNAL_ERROR` for persistence faults; mutation not committed |
| Rename failure | Prior bytes remain | Prior state remains | `500 INTERNAL_ERROR`; mutation not committed |
| Rename succeeds and parent-directory open/sync/close fails | New canonical bytes remain | New candidate rules/filter are published | `500 SUPPRESSIONS_DURABILITY_UNCERTAIN`; mutation committed, crash durability unconfirmed |
| Every phase succeeds | New canonical bytes remain | New candidate rules/filter are published | Existing success response |

The post-rename result deliberately does not claim rollback. The manager
retains its mutation lock through candidate compilation, persistence outcome
classification, and active-state publication, so a committed error cannot
leave the canonical file and filter describing different suppression sets.

## Non-Goals

- Do not change suppression schema, validation, matching semantics, CRUD
  success payloads, authentication, or reload behavior.
- Do not add cross-process locking, retries, backups, journaling, migration,
  automatic repair, or rule-file changes.
- Do not broaden the manager transaction model beyond the existing serialized
  mutation and active-filter swap boundary.
- Do not claim Windows or non-POSIX portability, production crash evidence, or
  filesystem guarantees beyond the checked Linux lifecycle.
- Do not access private data or exercise tag, release, image, registry, or
  workflow authority.

## Risks

- Treating a post-rename directory-sync failure as an ordinary rejection would
  leave canonical disk changed while the active filter and client assume the
  prior state.
- Returning success after the parent-directory durability boundary fails would
  overstate crash durability.
- A partial write with a nil error must fail closed; checking only the write
  error can acknowledge truncated JSON.
- Injected operations stored in package globals could race unrelated alert and
  API tests; the seam must belong to the manager or direct save invocation.
- Cleanup after rename must not remove the new canonical path or conceal a
  stale temporary file.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Complete writes are mandatory | Direct short-write and write-error regressions reject the mutation and compare exact prior bytes |
| Pre-rename lifecycle is preservation-safe | Direct stat, directory-create, temp-create, chmod, file-sync, temp-close, and rename faults preserve prior bytes and leave no temporary file |
| Successful replacement is durability-explicit | Direct operation-order regression observes stat, directory creation, temp creation, write, chmod, file sync, close, rename, directory open, directory sync, and directory close |
| Post-rename failures are classified | Direct directory-open, directory-sync, and directory-close regressions report committed replacement and retain exact new bytes |
| Manager state agrees with committed disk | Injected post-rename faults return a committed error while active rules/filter and canonical file contain the candidate suppression |
| Rejected mutations preserve active state | Injected pre-rename fault retains prior file/rules/filter and permits a later valid mutation under the released lock |
| API reports the committed ambiguity | Direct create/update/delete error mapping returns `SUPPRESSIONS_DURABILITY_UNCERTAIN` for a classified committed persistence result and retains existing error mapping otherwise |
| Existing behavior remains compatible | Canonical schema/mode tests, complete alert and API tests, repeated race runs, full native tests, and E2E smoke pass |
| Current claims and delivery records agree | Documentation, roadmap/state JSON, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 17a5809f83959714f8801fdfa7e613520e06dd14`.
- Verified the exact R90-78 feature and docs-only closure, both Vault notes,
  full-index rows, MOC links, and current stable rule/config/API authority.
- Reviewed the Jul 20 through Aug 9 phase history. The existing Aug 9 audit and
  completed R90-77/R90-78 sequence leave no new unresolved validation
  deviation, stale stable authority, or missing delivery record.
- Parsed all 93 prior task-state JSON files and verified all 84 roadmap rows
  have exactly one Definition.
- Audited every unfinished item. R90-79 is the sole dependency-ready local
  item; R90-80 remains dependency-planned, R90-59 remains
  publication-blocked, and R90-75 remains evidence/product-scope-blocked.
- Direct source/test review confirms `SaveSuppressionsToFile` detects only
  write errors and applies chmod/close/rename without file or directory sync;
  the manager publishes memory only on a nil save result and the API cannot
  distinguish a post-rename commit. Existing tests cover canonical format,
  ordinary persistence, mutation, and reload only.

## Validation

- Run focused `engine/internal/alert` and `engine/internal/api` tests uncached,
  then their race-detector variants.
- Run every direct R90-79 lifecycle, manager-outcome, and API mapping regression
  twenty times uncached under the race detector.
- Preflight repository-pinned tools, then run `make test`, `make e2e-smoke`,
  `make docs-check`, and `make knowledge-check` fail-fast.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage
  plus complete unfinished-item fields.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, schema, config, workflow, release,
  and publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Suppression replacement now requires an exact-length write, preserved mode,
  file sync, file close, atomic rename, parent-directory open/sync/close, and
  exact temporary-path cleanup while retaining missing-parent creation.
- The operation seam belongs to each manager or direct save call; production
  and parallel tests do not share mutable fault state. Replacement errors keep
  their underlying cause and classify whether rename committed.
- Direct faults for stat, parent creation, temp creation, short-write, write,
  chmod, file-sync, temp-close, rename, directory-open, directory-sync, and
  directory-close prove exact pre-rename preservation, exact post-rename
  candidate bytes, temporary cleanup, and complete operation ordering.
- The manager retains its mutation lock through candidate compilation,
  persistence classification, and active rules/filter publication. A
  post-rename durability error publishes the committed candidate before
  returning; a pre-rename error retains prior file/filter state and releases
  the lock for a later successful retry.
- The API maps classified create, update, and delete results to
  `SUPPRESSIONS_DURABILITY_UNCERTAIN` while uncommitted persistence failures
  retain `INTERNAL_ERROR`.
- Complete uncached focused alert/API tests and their race-detector variants
  pass. Every direct R90-79 lifecycle, manager-outcome, and API-mapping
  regression passes twenty uncached race-detector repetitions.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl, timeout, jq, pkg-config, and libpcap
  1.10.4 were available before the complete fail-fast chain.
- Final uncached focused alert/API tests and their race-detector variants pass.
  Every direct R90-79 lifecycle, missing-parent, manager-outcome, and API
  mapping regression passes twenty uncached race-detector repetitions.
- `make test` passes the C parser/sender tests and every Go package uncached
  under the race detector. `make e2e-smoke` passes with 6 packets processed, 5
  alerts generated, and 8 rules loaded.
- `make config-check` passes current repository configuration, rules, and
  suppressions. `make docs-check` and the 33-test `make knowledge-check` pass
  on the complete behavior/documentation surface.
- All 94 task-state JSON files parse; every one of the 84 roadmap rows has
  exactly one Definition and every unfinished item retains status, dependency,
  window, risk, acceptance criteria, required validation, and stop condition.
- `gofmt`, `git diff --check`, exact eleven-path scope, credential,
  sensitive-path, dependency, schema, config, workflow, release, and
  publication reviews pass. No dependency, suppression schema, configuration,
  workflow, release artifact, or external mutation was added.

## Delivery Results

- Feature commit:
  `8c621a926ac7ecbd1d730884a1afbc1ebb5e101e` (`fix: harden suppression file
  replacement durability`). It contains exactly the eleven validated source,
  test, documentation, roadmap, plan, and task-state paths.
- `main` was pushed without force or tags, then a fresh fetch verified
  `HEAD == origin/main == 8c621a926ac7ecbd1d730884a1afbc1ebb5e101e`
  with fast-forward ancestry from the recorded baseline. The post-fetch
  33-test knowledge gate passed.
- Exact range
  `17a5809f83959714f8801fdfa7e613520e06dd14..8c621a926ac7ecbd1d730884a1afbc1ebb5e101e`
  was synchronized to the sole local Vault. The generated iteration note,
  full-index row, and MOC link are verified.
- Stable MOC/config/suppression/API prose left stale by deterministic
  synchronization was reconciled to the delivered pre-rename preservation and
  post-rename committed-but-durability-uncertain boundary without rewriting
  immutable iteration notes. Replaying the identical feature range preserved
  the exact reconciled Vault tree hash
  `42df9837208932ea74aca96153e394380b617fca0e4ad6ad2081326868c73268`.
- R90-79 is complete. R90-80 is ready but unstarted; R90-59 and R90-75 retain
  their recorded external blockers.

## Authority Boundaries

This trigger authorizes only R90-79 suppression-file replacement durability,
phase-aware manager/API reporting, direct tests, current documentation and
task-state reconciliation, repository validation, commit/push, and local Vault
synchronization. It does not authorize rule behavior, cross-process
coordination, schema or migration policy, private/external data, performance
policy, tag or release mutation, image/registry publication, or workflow
dispatch.

## Stop Conditions

Stop if a deterministic post-rename outcome cannot keep canonical file and
active filter aligned, a required lifecycle phase is not directly injectable,
the target platform cannot support the defined directory-sync boundary without
a product portability decision, validation remains ambiguous after focused
uncached review, or completion needs any authority excluded above.
