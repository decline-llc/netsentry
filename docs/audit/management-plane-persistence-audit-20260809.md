# NetSentry Management-Plane Persistence Audit — 2026-08-09

## Audit Baseline

- Repository baseline: `de949bda14a66a407391671f92f0c7b938fb2da5`.
- Branch and remote: clean `main`; fetched `origin/main` matched the baseline.
- Audit period: 2026-07-20 through 2026-08-09.
- Delivery authority: current source, direct tests, Git commits, completed task
  plans/states, fetched remote refs, public documentation, and exact-range
  local Vault records.
- External boundary: R90-59 and R90-75 remain blocked; this audit authorizes no
  compatibility decision, runtime change, tag, publication, performance
  policy, workflow dispatch, or external evidence access.

## Method

1. Reconcile each R90-77 through R90-79 feature and docs-only closure with its
   parent, intended paths, completed plan/state, fetched branch, and two Vault
   records.
2. Trace every promised transaction or persistence boundary from current
   source to a direct regression rather than accepting a nearby package test.
3. Review recent commits at phase level and record only material trends,
   unresolved validation, stale authority, missing records, and open risks.
4. Compare API, architecture, development, changelog, limitation, and stable
   Vault claims with the delivered behavior and evidence classification.
5. Classify remaining topics as complete, externally blocked, or requiring a
   future product/migration decision without selecting that decision.
6. Parse every task-state JSON and require exactly one complete Definition for
   every roadmap row and complete fields for every unfinished item.

## Delivery History Review

| Period | Commits | Main delivery theme |
| --- | ---: | --- |
| Jul 20–26 | 61 | recovery and persisted-row contracts, schema compatibility, and read-only preservation |
| Jul 27–Aug 2 | 38 | recovery structure/encoding, candidate refresh, operator recovery, and committed-prefix retry |
| Aug 3–6 | 22 | sustained fuzz closure, storage fault lifecycle, and benchmark evidence/baseline |
| Aug 7–9 | 10 | local-only tag boundary and management-plane transaction/durability sequence |
| **Total** | **131** | one dependency-ordered correctness, recovery, evidence, release-boundary, and management-plane sequence |

The period contains 56 behavior-like commits, 67 `docs: record R90-*`
delivery closures, and 8 other documentation changes. Material deviations were
resolved at their owning increments: SQLite cancellation evidence retained both
context and driver causes, the benchmark gate stopped running ordinary Go tests,
an unrelated receiver timing result received focused uncached review plus a clean
full rerun, and transient SSH/Vault-path or inferred-SHA mistakes were verified
before synchronization. No unresolved validation result, stale current stable
authority, or missing delivery record changes this audit's conclusion.

## R90-77 Through R90-79 Delivery Reconciliation

| Increment | Exact Git chain | Direct boundary evidence | Task, remote, and Vault result |
| --- | --- | --- | --- |
| R90-77 | Feature `0ae76e167928f0ab1dafe015a997ccd1f61c664f` follows baseline `40798847be8e7bb9270b5c5d7675c27f7addf7b1`; closure `4b5b199f37531e69c08cb7fa7b1d814f83047a37` is its child. The feature changes exactly nine source/test/public-plan/state paths; the closure changes only its three delivery-record paths. | Channel-synchronized create/create, update/delete, and mutation/reload regressions prove the second transaction cannot cross the first transaction's observable state-read boundary. Successful responses are compared with canonical disk and active memory; validation and persistence rejection preserve prior state and permit retry. | Completed plan/state, fetched ancestry, both iteration notes, both full-index rows, both MOC links, and current stable rule/config/API prose pass review. |
| R90-78 | Feature `8d053d1d3c4e390151c224aa8f86852312506eb8` follows R90-77's closure; closure `17a5809f83959714f8801fdfa7e613520e06dd14` is its child. The feature changes exactly eleven source/test/public-plan/state paths; the closure changes only its three delivery-record paths. | One direct table reaches stat, create, short-write, write, chmod, file-sync, file-close, and rename rejection with exact prior bytes and temporary cleanup. Three post-rename cases reach directory open/sync/close and prove committed new bytes; an ordered lifecycle test and API pre/post-rename regressions prove active-state/error agreement. | Completed plan/state, fetched ancestry, both iteration notes, both full-index rows, both MOC links, and current stable rule/config/API prose pass review. |
| R90-79 | Feature `8c621a926ac7ecbd1d730884a1afbc1ebb5e101e` follows R90-78's closure; closure `de949bda14a66a407391671f92f0c7b938fb2da5` is its child and fetched `origin/main`. The feature changes exactly eleven source/test/public-plan/state paths; the closure changes only its three delivery-record paths. | Direct cases reach stat, parent creation, temp creation, short-write, write, chmod, file-sync, file-close, and rename rejection, plus directory open/sync/close committed outcomes. Lifecycle ordering, manager pre-rename retry/post-rename filter publication, and create/update/delete API mapping are each checked directly. | Completed plan/state, fetched ancestry, both iteration notes, both full-index rows, both MOC links, and current stable suppression/config/API prose pass review. |

The regression bodies reach the named boundaries through package-local or
manager-owned seams; none substitutes a generic package pass for a promised
interleaving or lifecycle phase. The recorded focused/race repetitions, full
native tests, E2E smoke, documentation, configuration, and knowledge gates are
historical delivery evidence, while this audit independently verifies their
current test bodies and Git identity.

## Current Claim Reconciliation

| Surface | Current authority | Audit result |
| --- | --- | --- |
| In-process rule transaction order | `engine/internal/api/router.go` holds one server mutex around each authoritative read, validation/persistence or reload, and active snapshot publication; packet matching retains immutable lock-free snapshots. API, architecture, development, changelog, and stable Vault prose say the same. | Complete for one API-server process. No cross-process writer guarantee is claimed. |
| Rule replacement lifecycle | `engine/internal/rule/loader.go` requires exact write, mode preservation, temporary-file sync/close, rename, and parent-directory sync/close. Committed post-rename errors cause active publication and `RULES_DURABILITY_UNCERTAIN`. | Complete for the directly checked local POSIX lifecycle. No retry, backup, migration, portable crash proof, or production crash evidence is claimed. |
| Suppression replacement lifecycle | `engine/internal/alert/suppressor.go` retains the manager mutation lock through candidate compilation, persistence classification, and active publication with the independently tested lifecycle and `SUPPRESSIONS_DURABILITY_UNCERTAIN`. | Complete for the directly checked local POSIX lifecycle. Rule and suppression evidence remain independent. |
| Legacy rule input | The loader intentionally accepts canonical wrapped rules plus prior top-level arrays and legacy scalar/config fields while saves emit the canonical schema. | Compatibility policy, not a discovered defect. Removal, telemetry, deprecation window, and migration tooling need explicit product/migration authority. |
| Cross-process management writers | Current documentation explicitly bounds serialization to one API server; no supported multi-writer topology or lock protocol is specified. | Product/deployment decision. Do not add file locks or distributed coordination without a supported topology and failure contract. |
| Filesystem portability and crash evidence | Fault injection proves ordered calls and deterministic pre/post-rename state on the checked local Linux path; it does not simulate power loss or prove every filesystem/platform. | External/platform evidence plus a portability decision would be required before broader claims or work. |
| Broad protocol and schema evolution | IPv6, TLS decryption, IP/TCP reassembly, broader pcapng/DLT strategy, multi-MITRE migration, and automatic cleanup remain published product boundaries. | Not local-ready. Each requires separate scope and compatibility authority. |
| R90-59 publication | The local signed tag remains absent remotely and publication authority is still withheld. | Externally blocked. This audit does not push tags or publish artifacts. |
| R90-75 performance policy | Same-host observations exist without comparable-environment evidence or product/SLO scope. | Externally blocked and non-blocking. No numeric gate is activated. |

No current public or stable Vault prose overstates the delivered management
plane. The only stale checkpoint wording says R90-80 is ready but unstarted;
the delivery synchronization for this audit must update stable prose after its
feature commit while leaving the six historical iteration notes immutable.

## Forward Queue

| ID | Status | Dependency | Remaining condition |
| --- | --- | --- | --- |
| R90-59 | Blocked on remote-publication authority | R90-59a plus explicit tag-push, GitHub Release, and GHCR authority | Resolve the recorded changelog/artifact boundary and explicitly authorize the external publication actions. |
| R90-75 | Blocked / pending evidence; non-blocking | R90-74 plus comparable-environment evidence and explicit budget scope | Supply matched evidence and choose portable, same-host-only, or observation-only product/SLO scope. |

Both Definitions retain forecast window, risk, acceptance criteria, required
validation, blocker evidence, unblock condition, and stop condition. R90-80
adds no speculative ready increment: legacy schema removal, cross-process
coordination, portable crash claims, and broad protocol work each cross a
product, migration, external-input, or platform-evidence boundary.

## Audit Conclusion

R90-77 through R90-79 form one exact, independently tested, and recoverable
management-plane sequence. The repository now has a defined in-process
transaction order plus preservation-safe, durability-explicit rule and
suppression replacement outcomes, and its public claims stay within the direct
local evidence. R90-80 closes the bounded audit without changing runtime or
choosing future compatibility policy. After delivery, no dependency-ready
local increment remains; R90-59 and R90-75 stay blocked until their recorded
external conditions change.
