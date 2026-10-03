# R90-144: Isolate rule reload snapshots from caller-owned inputs

## Selection, phase audit and authority

Clean freshly fetched HEAD/origin/main/FETCH_HEAD is
`e96e6524f3e24a1b508120b47b06da4bb1464ed0`. Both R90-143 feature/closure exact Git/Vault ranges,
generated path scope, iteration notes, full index and MOC are verified. The
367-file Vault snapshot JSON SHA-256 reproduces
`d00356a3e58dbfafa3266086a31a4c7357a224e6b58a3f716fc24d7229001d9b`;
12 current stable handoffs and 320 immutable iteration-directory notes are captured.
The 67-commit Sep 4–Oct 2 phase review distinguishes measurement/tooling/replay,
input boundaries, repeated queue audits, historical patched-candidate publication
and main metadata. No missing delivery or qualifying R90-75 measurement appears.

The user now directs core-code development; the agreed first scope is rule
snapshot isolation. The former local queue is empty; register this bounded
source-grounded R90-144 as ready after completed R90-143 and existing rule engine/
API transaction contracts. Reconcile the queue within this implementation
increment before code changes rather than deliver another standalone audit.
R90-75 retains its complete independent departmental contract and does not block
implementation. Horizon stays Oct 2–Dec 30. Define the future R90-145 completed-
packet metric contract without starting it. IPv6 remains a later product/protocol
scope, not silently authorized here.

The user-directed test-department split remains: behavioral/race/CLI/full-suite/
scanner/knowledge/acceptance execution is **not run; delegated by user**. Author
meaningful direct regression source for the department; static Go parse/format and compile-only review,
source/diff/docs/JSON/queue and exact Git/Vault checks are local duties. No runtime
pass, observed race fix, capacity or SLO result may be inferred.

## Scope persisted before runtime and documentation edits

Exactly six paths: `engine/internal/rule/engine.go`, new
`engine/internal/rule/engine_snapshot_test.go`, `docs/architecture.md`, rolling
roadmap, this plan and matching task-state JSON. Local implementation branch:
`feat/r90-144-rule-snapshot-isolation`. On completed static review, create a
focused feature commit, fast-forward main from the verified baseline, push/fetch
verify main, sync exact range and reconcile stable notes. One docs-only closure
on main records verified delivery; verify that range too. No remote branch,
force push, tag, release or workflow dispatch is needed.

## Behavior and ownership

`buildState` currently validates the input pointers, shallow-copies the slice,
then sorts/compiles and retains those pointers. `Match` and `Rules` can therefore
observe subsequent caller changes, violating the intended immutable state.
Reuse existing `cloneRule` for a new owned slice before validation. Its struct
copy preserves scalar/string fields; separate Config bytes and MITRETechs backing
arrays isolate all mutable Rule fields. Validate, sort and compile this same
owned set, and retain only it in the atomically published state. Validation
failure keeps the old state; nil entries retain existing indexed diagnostics;
nil/empty sets retain clearing behavior. Never sort or mutate caller inputs.

Preserve matcher semantics, priorities/early exit, rule/alert/API/file formats,
validation diagnostics, Rules defensive output, API file transactions and the
single Store after full success. Copies allocate during reload, outside Match.
Caller input must remain stable while Reload copies it: this does not support
concurrent mutation during the call or unsafe modifications of immutable strings.
Caller writes after return can proceed independently of readers of the published
snapshot. No new mutex, dependency, quota or runtime option is introduced.

## Acceptance and direct evidence map

| Acceptance | Planned local evidence | Departmental execution, unrun |
| --- | --- | --- |
| Delivery/dependency authority | Fresh clean refs; both prior exact ranges/scope/note/index/MOC; 367-file snapshot; 67-commit phase review; full unfinished contracts | R90-75 independent |
| Own complete Rule graph before validation | Small source diff; cloneRule scalar/raw JSON/MITRE coverage; clone → validate → sort/compile → single Store order | Direct Reload then mutate caller slice/ID/name/type/severity/priority/enabled/early-exit/description/Config bytes/MITRE fields; Rules and Match remain unchanged across payload/IP/port |
| Compatibility and failure preservation | Existing nil handling/diagnostic order/compiler/sort/Match/Rules/API paths unchanged | Nil/empty clearing; nil/duplicate/malformed rejected reload retains exact prior state, prior caller mutations isolated; input order/data unchanged |
| Independent reads after return | No retained caller pointers; defensive Rules output unchanged | Mutate returned Rules slice/objects/Config/MITRE and compare snapshot/alerts; synchronized post-return caller mutation concurrent with Match/Rules under race detector |
| Regression boundary specificity | Direct source review of named test calls, same fixtures/expected comparisons and mutation coverage | Focused package/race tests, then broader rule/API/pipeline/full native as department considers applicable; no fixtures during Reload mutation |
| Documentation and delivery | Go parse/format and compile-only regression review, docs/JSON/full unique roadmap pairs/history/links/fences/six-path scope/diff/sensitive additions; exact feature/closure Git/Vault; stable reconciliation/immutable preservation/identical replay | All execution suites remain delegated |

Regression source must reach public Engine.Reload and both Rules/Match; testing
cloneRule alone is weaker. Post-return concurrency must synchronize after Reload
and must not imply safe concurrent modification during copying. Tests are written
but unexecuted; static review cannot prove their runtime outcome.

## Risk, non-goals and stop conditions

Medium correctness risk: clone timing/validation compatibility and ownership.
Reuse the already established clone helper and preserve every other algorithm.
Exclude matcher optimizations, pipeline changes, IPv6, new API/schema/MITRE
policy, persistence/rollback refactors, dependencies/toolchains, test execution,
external traffic/private data and publication. Stop for ambiguous diagnostics,
unaccounted mutable model fields, required migration, competing user edits,
contradictory Git/Vault or new external/product authority. Finish exactly this
increment; leave R90-145 unstarted. Existing skills need no edit so far.


## Implementation checkpoint and compilation boundary

The runtime diff uses cloneRule for the entire owned slice before unchanged
validateRuleSet; sorting/compilation retain only cloned pointers. All other
functions and matcher/compiler behavior remain unchanged. Six-path scope holds.
Five direct test functions contain three rule-type fixtures, isolated mutation
subcases, defensive output/input ordering, failed-reload state and nil/empty
checks, plus post-return synchronized concurrent reads/writes. Tests are authored
but not executed. Broaden local static validation to compile-only test-source
review (`go test -c`), with the output binary outside the repository and never
invoked. Preflight confirms the owning engine module resolves its exact pinned
Go 1.26.8 toolchain. Compilation is not behavioral/race/full-suite execution.
Existing skill guidance covers the workflow; no skill change is warranted.


## Final static review checkpoint

Exact runtime-source comparison against the fetched baseline permits only the
owned-slice clone/validation substitution and ownership comment; Match, Rules,
cloneRule, all compilers, sort comparator and single publication path are unchanged.
Rule model has only scalar/string fields plus the two separately copied slices;
validation, sorting and compilation all consume cloned pointers. Nil/rejection
semantics and API transaction/source paths are directly reviewed.

Direct regression source covers 39 separate field/slice/nested-data mutation
subcases across payload/IP/port and compares both public Rules and complete Match
alerts. Other cases cover preserved input order/data and priority sorting,
defensive output mutations, exact nil diagnostic, rejected reload retaining the
same prior state, nil/empty clearing and post-return caller mutations. The
concurrent case observes the first caller mutation through a channel before read
assertions, then releases further mutation alongside reads; no fixed sleep or
mutation during Reload. Every promised direct case reaches Engine.Reload and
Rules/Match; none relies on clone-helper-only evidence.

Go 1.26.8 compile-only `go test -c` passes after the final synchronization edit;
the generated binary is outside the repository and was never executed.
Go parse/format, docs-check, 162 task JSON, 149 complete unique roadmap pairs,
full forward contracts, unchanged R90-75 Definition, ordered history, links/fences,
six-path scope, diff/sensitive-addition review pass. Complete 367-file Vault
baseline is unchanged. All behavioral/race/CLI/full-suite/scanner/knowledge/
acceptance execution remains delegated and unrun. Queue restoration within the
core implementation is the recorded planning adjustment; no implementation or
validation failure occurred. No runtime/race/SLO claim. Skills need no edit.
Feature commit/main delivery/Vault and one closure remain; R90-145 unstarted.
