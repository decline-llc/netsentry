# R90-128: Post-sender-integration delivery and forward-queue audit

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD baseline is
`f1effc42ddafca8d7d90a7caac54f1fa5f158f25`. R90-127's exact feature/closure
Vault ranges, notes, full index, MOC links and nine stable notes are verified;
closure Vault hash is
`d156113a301b14ac37d533712d7aef6737ef9a633fd7c5e0c96acdcebbaf29b7`.
The Sep 3–Oct 1 phase review covers 32 commits spanning queue recovery/SLO
contract, reporting/acquisition tooling, context/reconstruction and sender
integration. Delivery pairs are accounted for; no new acceptance artifact,
missing closure or stale release/testing authority changes priority.
All 144 prior task states parse; 131 unique roadmap rows/Definitions match.

No local row is ready. Select R90-128 as the smallest documentation-only queue
unblocker. Refresh the active horizon from Sep 30–Dec 28 to Oct 1–Dec 29 because
the date rolled forward; preserve historical windows and completion records.
R90-59 still requires candidate/tag authority and fresh validation. R90-75
acceptance remains departmental, with implementation unblocked. Neither is
silently rescheduled or treated as permission for external action.

## Source-grounded finding and bounded follow-up

`scripts/slo_collect.py` collect accepts manifest/offered/events paths and an
optional scratch directory, with no cumulative retained-input budget. Its shared
_rows opens a source using Path.open("rb") and limits each readline to 256 KiB,
but loops through all rows without a total byte bound. Manifest reading delegates
to slo_report.read_observations, which has a 64 MiB bound but also uses ordinary
Path.open. Neither direct input path establishes non-following regular-file
admission before reading. The inputs are documented as finalized external exports;
that requirement is presently supplied by the operator, not enforced here.

The enclosing bundle/reconstruction path already uses _Bundle.snapshot with a
64 GiB configurable input budget, O_NOFOLLOW/O_NONBLOCK, regular-file checks and
before/after metadata. That wrapper does not cover the departmental standalone
slo_collect invocation. This is a static boundary difference, not an executed
hang, data-loss incident or claim that earlier tools promised a disk quota.
_rows is also used by slo_ingress.send_fixture: avoid silently changing the live
sender while repairing standalone adapter admission. Reconstruction invokes
collect on retained inputs: preserve its explicit larger-budget choices rather
than introducing an accidental second default cap.

Queue R90-129 only: bounded finalized-file admission/retention for the standalone
adapter, using the existing 64 GiB configurable budget convention and preserving
the 64 MiB manifest/256 KiB row limits. Its separate implementation plan must
freeze the exact API/CLI, cumulative accounting, unchanged format/receipt
compatibility and caller budget propagation before editing. Require regular-file
admission through non-following/nonblocking descriptors, consumed-input byte
accounting across all three files, before/after metadata checks, and no completed
receipt on source/budget/I/O failure. Retain partial outputs, never modify sources,
and distinguish retained-input limits from SQLite/output memory/disk costs.
No global reporter-reader change or live-sender behavior change is authorized.

## Scope, acceptance and evidence

Exactly three intended paths: roadmap, this plan and matching R90-128 task state.
No runtime, tests, public tool schema, dependency, fixture, benchmark, CI, traffic,
private-data, publication or R90-129 implementation change.

| Acceptance | Planned static/direct evidence |
| --- | --- |
| Delivery/history audit | Fetched SHA; two exact R90-127 Vault note/index/MOC ranges; phase history and explicit test-delegation limitation |
| Source gap | collect/_rows/read_observations versus _Bundle.snapshot and sender/reconstruction call sites; no executed failure claim |
| Complete queue/horizon | R90-128/R90-129 rows and Definitions with dependency/window/status/risk/acceptance/review/non-goals/stop; Oct 1–Dec 29 active horizon |
| Structural integrity | docs-check, every task JSON, complete unique roadmap multisets, ordered history, local links/fences, exact scope/diff and sensitive-information review |
| Delivery closeout | Commit/push/fetch verification, exact Vault range and stable prose reconciliation/replay, one docs-only closure; R90-129 ready but unstarted |

Tests, knowledge suites and acceptance remain not run, delegated by user. No
behavioral regression evidence is claimed. Stop for a new product/private/external
authority requirement; none is needed for this bounded documentation repair.
No skill change is needed for a finding covered by the existing audit rules.

## Static review checkpoint

The exact three-path documentation repair now records the Oct 1–Dec 29 horizon
and fully defines R90-128/R90-129. Source/call-site review confirms the standalone
input boundary and the shared live-sender/reconstruction compatibility concerns;
no runtime failure was executed or inferred as established evidence. Historical
release/acceptance contracts and completed delivery records remain unchanged.

Static docs-check, 145 task-state JSONs, 133 unique complete roadmap multisets,
unfinished contracts, ordered history, local links/fences, exact scope/diff and
sensitive-information review pass. R90-129 remains planned and unstarted until
this audit is delivered. No behavioral or knowledge tests, CLI smoke, private
input, resource discovery or publication check ran. No scope deviation or skill
update was needed. Commit/push/fetch and exact-range Vault closeout remain.
