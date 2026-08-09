# Task Plan: R90-78 rule-file replacement durability

## Metadata

- Timestamp: 2026-08-09T05:28:43-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `4b5b199f37531e69c08cb7fa7b1d814f83047a37`

## Goal

Make canonical rule seed-file replacement explicit and testable across short
write, permission, file-sync, close, rename, and parent-directory durability
boundaries while keeping the active rule snapshot consistent with every
reported commit outcome.

## Scope

- Replace the rule file through a temporary file in the target directory using
  a complete-write check, preserved mode, file sync, close, atomic rename, and
  containing-directory sync and close.
- Introduce a package-local filesystem seam so every lifecycle phase can be
  fault-injected without changing production callers or global process state.
- Classify replacement errors by whether the rename committed the new
  canonical file.
- Preserve the old canonical bytes and active snapshot for every failure
  through rename, and remove the temporary file exactly.
- On a post-rename durability error, load and publish the committed canonical
  rules before returning an explicit API error that says the mutation was
  applied but crash durability could not be confirmed.
- Reconcile current rule-management API, architecture, development, changelog,
  roadmap, task-state, and Vault authority.

## Defined Outcome Contract

| Boundary | Canonical file | Active snapshot | API outcome |
| --- | --- | --- | --- |
| Marshal/create/write-short/write-error/chmod/file-sync/temp-close failure | Prior bytes remain | Prior snapshot remains | `500 INTERNAL_ERROR`; mutation not committed |
| Rename failure | Prior bytes remain | Prior snapshot remains | `500 INTERNAL_ERROR`; mutation not committed |
| Rename succeeds and parent-directory open/sync/close fails | New canonical bytes remain | New canonical rules are loaded and published | `500 RULES_DURABILITY_UNCERTAIN`; mutation committed, crash durability unconfirmed |
| Every phase succeeds | New canonical bytes remain | New canonical rules are loaded and published | Existing success response |

The post-rename result deliberately does not claim rollback: rename has already
made the new file authoritative. Returning a distinct error while synchronizing
active memory makes the committed outcome observable without inventing a
portable crash guarantee.

## Non-Goals

- Do not change rule schema, validation, CRUD success payloads, authentication,
  transaction serialization, hot-reload semantics, or packet matching.
- Do not add cross-process locking, retries, backups, journaling, migration, or
  automatic repair.
- Do not change suppression persistence; R90-79 retains that independent scope.
- Do not claim Windows or non-POSIX portability, production crash evidence, or
  filesystem guarantees beyond the checked Linux lifecycle.
- Do not access private data or exercise tag, release, image, registry, or
  workflow authority.

## Risks

- Treating a post-rename sync failure as an ordinary rejection would leave the
  canonical file changed while active memory and the client assume otherwise.
- Returning success after a parent-directory sync failure would overstate
  crash durability.
- A partial write with a nil error must still fail closed; checking only the
  write error can acknowledge truncated JSON.
- Test seams that mutate package globals could race unrelated tests; the seam
  must be passed directly to the replacement implementation.
- Cleanup after rename must not remove the new canonical path or conceal a
  stale temporary file.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Complete writes are mandatory | Direct short-write and write-error tests reject the mutation and compare exact prior bytes |
| Pre-rename lifecycle is preservation-safe | Direct chmod, file-sync, temp-close, and rename faults preserve prior bytes and leave no temporary file |
| Successful replacement is durability-explicit | Direct operation-order test observes write, chmod, file sync, close, rename, directory open, directory sync, and directory close |
| Post-rename failures are classified | Direct directory-open, directory-sync, and directory-close fault tests report committed replacement and retain exact new bytes |
| API state agrees with committed disk state | Injected post-rename fault returns `RULES_DURABILITY_UNCERTAIN` while canonical reload and active snapshot contain the new rule |
| Rejected mutations preserve active state | Injected pre-rename fault returns `INTERNAL_ERROR`, retains prior canonical bytes and active rules, and permits a later valid request |
| Existing behavior remains compatible | Canonical schema/mode tests, complete rule and API tests, repeated race runs, full native tests, and E2E smoke pass |
| Current claims and delivery records agree | Documentation, roadmap/state JSON, knowledge, diff, staged-scope, and sensitive-information checks pass |

## Trigger Audit

- Read the `netsentry-next` and `netsentry-roadmap` skills plus the active
  rolling roadmap before selection.
- Fetched `origin/main` and verified a clean worktree with
  `HEAD == origin/main == 4b5b199f37531e69c08cb7fa7b1d814f83047a37`.
- Verified the exact R90-77 feature and docs-only closure, both Vault notes,
  full-index rows, MOC links, and current stable rule/config/API authority.
- Reviewed the Jul 20 through Aug 9 phase history: 127 commits comprise 54
  behavior-like changes, 65 delivery-record closures, and 8 other
  documentation changes. The Aug 9 R90-76 audit and R90-77 delivery leave no
  unresolved validation deviation or missing delivery record.
- Parsed all 92 prior task-state JSON files and verified all 84 roadmap rows
  have exactly one Definition.
- Audited the complete unfinished queue. R90-78 is the sole dependency-ready
  local item; R90-79 and R90-80 remain dependency-planned, R90-59 remains
  publication-blocked, and R90-75 remains evidence/product-scope-blocked.
- Direct source/test review confirms `rule.SaveToFile` detects only write
  errors, does not sync the file or parent directory, and cannot distinguish a
  post-rename commit. Existing tests cover canonical format and mode only.

## Validation

- Run focused `engine/internal/rule` and `engine/internal/api` tests uncached,
  then their race-detector variants.
- Run every direct R90-78 regression twenty times uncached under the race
  detector.
- Run `make test`, `make e2e-smoke`, `make docs-check`, and
  `make knowledge-check` fail-fast after preflighting repository-pinned tools.
- Parse all task-state JSON and verify exact roadmap row/Definition coverage
  plus complete unfinished-item fields.
- Run `gofmt`, `git diff --check`, intended staged-diff review, and scoped
  credential, sensitive-path, dependency, schema, config, workflow, release,
  and publication review.
- Push without force or tags, fetch and require
  `HEAD == origin/main == new`, rerun the knowledge gate, then synchronize and
  verify the exact full-SHA Vault range.

## Implementation Checkpoint

- Rule replacement now requires an exact-length write, preserved mode, file
  sync, file close, atomic rename, parent-directory open/sync/close, and
  temporary-path cleanup. The operation seam is passed directly to the helper;
  production and parallel tests do not share mutable fault state.
- Replacement errors retain their underlying cause and classify whether rename
  committed. Stat, create, short-write, write, chmod, file-sync, temp-close,
  and rename faults preserve exact prior bytes. Directory open, sync, and close
  faults retain exact new bytes and report the committed outcome.
- API rule mutation attempts reload and publish whenever replacement reports a
  committed rename. A successful reload followed by a durability error returns
  `500 RULES_DURABILITY_UNCERTAIN`; ordinary pre-rename failures retain the
  existing `500 INTERNAL_ERROR` rejection and release the transaction lock for
  retry.
- Direct tests cover all named lifecycle faults, exact temporary cleanup,
  complete operation ordering, pre-rename API preservation/retry, and
  post-rename canonical/active agreement.
- The first focused run reached the short-write behavior correctly but the new
  test expected the generic injected error instead of `io.ErrShortWrite`. That
  test-only expectation was corrected before accepting behavioral evidence;
  the complete focused ordinary/race sequence and twenty uncached direct race
  repetitions then passed.

## Validated Evidence

- The module selected its pinned Go 1.25.12 toolchain; GCC 13.3.0, GNU Make
  4.3, Bash 5.2.21, Python 3.12.3, curl, timeout, jq, pkg-config, and libpcap
  1.10.4 were available before the complete fail-fast chain.
- Final uncached focused rule/API tests and their race-detector variants pass.
  Every direct R90-78 lifecycle and API outcome regression passes twenty
  uncached race-detector repetitions.
- `make test` passes the C parser/sender tests and every Go package uncached
  under the race detector. `make e2e-smoke` passes with 6 packets processed, 5
  alerts generated, and 8 rules loaded.
- `make docs-check` and the 33-test `make knowledge-check` pass on the complete
  behavior/documentation surface.
- All task-state JSON files parse; every roadmap row has exactly one Definition
  and all unfinished items retain status, dependency, window, risk, acceptance,
  validation, and stop records.
- All 93 task-state JSON files parse and all 84 roadmap rows match exactly one
  Definition with complete unfinished-item records.
- Staged chronological review found the new R90-78 implementation checkpoint
  attached to an older matching roadmap anchor; it was moved to the current
  checkpoint tail before delivery and no behavioral evidence changed.
- `gofmt`, `git diff --check`, exact eleven-path scope, staged-diff,
  credential, sensitive-path, dependency, schema, config, workflow, release,
  and publication reviews pass. No dependency, rule schema, configuration,
  workflow, release artifact, or external mutation was added.

## Authority Boundaries

This trigger authorizes only R90-78 rule-file replacement durability,
phase-aware API reporting, its direct tests, current documentation/task-state
reconciliation, repository validation, commit/push, and local Vault
synchronization. It does not authorize suppression changes, cross-process
coordination, schema or migration policy, private/external data, performance
policy, tag or release mutation, image/registry publication, or workflow
dispatch.

## Stop Conditions

Stop if a deterministic post-rename outcome cannot keep canonical file and
active memory aligned, a required lifecycle phase is not directly injectable,
the target platform cannot support the defined directory-sync boundary without
a product portability decision, validation remains ambiguous after focused
uncached review, or completion needs any authority excluded above.
