# R90-120: Supplied run-context declarations and retained references

## Selection and authority

Clean fetched baseline `79011111b3e7997f8176cf3325952fcd885f178e` equals
HEAD/origin/main/FETCH_HEAD; R90-119 feature/closure Vault index/MOC are verified.
Recent contract/report/ledger/engine/ingress/bundle/comparison deliveries remain
untested under user direction. R90-120 is dependency-ready. The Sep 29–Dec 27
horizon is current; R90-75 acceptance remains departmental and R90-59 publication
retains its independent blocked boundary. No hardware shortage blocks development.

## Design and acceptance map

Implement a versioned standard-library context schema/API/CLI. A declaration
identifies run/profile/start and the exact observation-file SHA-256. Fixed hardware,
toolchain and isolation fields use typed nullable values plus evidence ID lists.
Unknown values remain null; missing fields/invalid types are schema errors.
Known values without references are explicit evidence gaps, not implicit facts.
A bounded evidence catalog has safe IDs, declared bytes and SHA-256; source names
are derived as ID.bin inside an explicitly supplied directory, never free paths.

Retain the declaration, supplied observations and referenced opaque files into a
fresh private directory. Strictly validate schemas, catalog uniqueness/references,
byte/digest identity and run/profile/start binding; declared CPU/memory allocation
must agree with supplied observations when known. Retain unreadable/missing/partial
inputs as gaps, malformed/mismatched evidence as mismatches, and publish the
context receipt last. Matching bytes do not prove a supporting file's meaning,
hardware state, independent isolation, SLO compliance or full workload truth.
No auto-discovery, shell execution, network or production service access occurs.

| Criterion | Planned evidence |
| --- | --- |
| Explicit context schema | Manual/AST review of fixed nullable fields, strict types, unknown gaps, bounded evidence IDs/catalog and reference checks |
| Retained evidence and identity | Source review of regular-file snapshots, safe computed basenames, declared digest/size verification, exact observation binding and allocation consistency |
| Complete departmental handoff | Document all fields, bounds, status precedence and required fault matrix; show how comparison consumers must bind the retained context without claiming certification |
| Delivery | Static AST/JSON/roadmap/history/link/diff/token-pattern review; verified push/fetch and exact local Vault sync/replay |

## Scope and non-goals

Intended paths (nine): scripts/slo_context.py, Makefile syntax registration,
docs/slo-context.md, docs/slo-compare.md, docs/performance-slo.md, roadmap,
this plan, task-state-20260929-slo-context.json and active R90-75 state.
No agent-run tests, CLI smoke, fixtures, benchmarks, acceptance/knowledge suites,
traffic, resource discovery, dependencies, CI changes or release/tag actions.
Schema declarations describe supplied facts; no product target or environment
truth is inferred. R90-119 comparison behavior remains unchanged in this increment;
its integration contract is documented, with consumer wiring queued separately.
Stop only for new private/external authority or an undocumented format. Current
schemas and explicit unknowns suffice; no real evidence is required to implement.

## Delivery and next boundary

Persist before edits. Implement and statically review one increment, verify its
Git/Vault delivery and stable prose replay, then one docs-only closure. Queue
R90-121 to consume retained context packages in the pair comparison, with fresh
identity/digest checks and explicit unknown/differing context diagnostics. Do not
start that consumer integration in this increment.

## Implementation and static review checkpoint

Implemented the nine intended paths. The versioned declaration contains 25 fixed
hardware/toolchain/isolation cells, with null unknowns and bounded evidence IDs.
Known values without references remain gaps; false booleans remain declarations.
Evidence catalogs reject duplicate/unknown/unused IDs and derive safe ID.bin
basenames. Exact raw declaration/observation/reference copies are retained under
finite per-file/combined limits; opaque contents are never parsed or executed.
Observation schema/run/profile/start/hash and known CPU/memory allocation bind
the declaration to its supplied measurement. Partial/missing inputs and malformed/
mismatched sources remain explicit. context.json publishes last without overwrite.

Manual review covered schema/type/enum/UTF-8 and reference bounds, regular-file
copying, source-change detection, remaining-budget arithmetic, allocation binding,
partial retention and file/directory synchronization. Documented status flags
mean implementation-check completeness only; targets, isolation adequacy and
reference relevance still require departmental judgment. R90-119's behavior is
unchanged and consumer wiring is explicitly queued as R90-121, not started.

`make python-check` only AST-parsed source. Static AST extraction matches all 25
fields to the documentation table. All 137 task JSONs parse, 125 unique roadmap
row/Definition pairs agree, and history/link/diff/token-pattern checks pass.
No implementation module, CLI, test, benchmark, acceptance or knowledge suite ran;
no context/measurement fixture or artifact was generated. The department's matrix
covers the unexecuted behaviors/faults. Existing skills already handle user test
delegation and exact delivery; no skill edit is needed.
