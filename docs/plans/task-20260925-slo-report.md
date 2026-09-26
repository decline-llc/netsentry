# R90-114: Acceptance measurement report implementation

## Authority and scope

On Sep 25 the user explicitly instructed the agent to skip tests, delegate
those to a specialist department and continue implementation. This overrides
the skills' mandatory local test/knowledge-test gates for this delivery. Do not
run unit, integration, benchmark, acceptance or knowledge test suites. Do not
turn test deferral into a compliance claim. Static syntax, formatting, source
review and Git/Vault evidence reconciliation remain in scope.

Deliver one standard-library Python API/CLI that summarizes department-supplied
completed-run observations into a retained JSON report: unique expected alert
identities, live-arrival to durable-write durations, missing events retained as
unbounded latency observations, late/missing/deadline counts, offered-versus-
fully-processed packet loss, separate sustained/burst and one-minute/five-minute
cohort summaries, extended-run/sample diagnostics, and no automatic SLO claim.
Implement evidence output without overwriting an existing file; embed source
observations and bind the exact input bytes by SHA-256. Reject malformed and
ambiguous records. The tool consumes measurements; it does not collect packets
or validate the physical timestamp/durability boundaries.

Baseline: `99f84420e52d66718e5c6eadce3fd21278016d78`, clean, fetched equal
to HEAD/origin/main/FETCH_HEAD; both R90-113 exact Git/Vault ranges are verified.
Recent delivery is documentation only; the new user instruction removes test
execution as a development prerequisite and permits this local implementation.
R90-75 acceptance and R90-59 publication remain separate unfinished outcomes.

## Intended paths

- `scripts/slo_report.py`: pure summary API, strict JSON input and CLI/output.
- `Makefile`: register the module with the existing static Python syntax target.
- `docs/slo-report.md`: observation schema, invocation and department handoff.
- `docs/performance-slo.md`: implemented surface and delegated-test boundary.
- `docs/plans/rolling-90-day-roadmap.md`: user waiver, R90-114 and R90-75 state.
- This plan and `docs/tasks/task-state-20260925-slo-report.json`.
- `docs/tasks/task-state-20260923-production-slo-acceptance.json`.
- Local stable Vault authority after verified Git delivery.

## Acceptance and planned evidence

| Criterion | Evidence available this increment |
| --- | --- |
| Useful executable artifact summarizer | Reviewed Python API and CLI; static AST parsing; documented exact schema |
| Missing events cannot disappear | Denominator includes every expected alert; nearest-rank p99 includes missing as infinity; missing/deadline counters |
| Loss and window accounting | Exact minute coverage, unique cohort counters, separate phase summaries and trailing five-minute windows |
| Evidence recoverable | Input bytes digest plus embedded observation document; no-clobber JSON output; result marks departmental review required |
| Test ownership honest | No test suite or acceptance run executed; test handoff matrix documents unverified conditions |
| Delivered scope isolated | Eight intended paths; diff/syntax/manual review; verified push/fetch and local Vault replay |

## Non-goals and risks

Do not implement a traffic generator, modify capture/runtime/storage protocols,
activate a CI/release numeric gate, invent test results, certify either profile,
or access a remote runner. The input counts are externally asserted cohort
observations, not independently reconstructed packet identities. The tool must
make that limitation explicit. JSON load is sized for per-minute counters and
low-volume alert events, not per-packet traces at production traffic rates.
Profile hardware/fixture truth and timestamp/durability evidence require the
test department's verification; this tool must never emit SLO compliance true.
The code is untested by explicit user direction; static review cannot replace
behavioral validation. Preserve this risk at feature and closure delivery.

## Delivery and stop conditions

Static AST and JSON parsing, exact path/roadmap coverage, sensitive-data and
`git diff --check` review precede commit. Tests, including `make knowledge-check`,
are explicitly deferred. Push without force/tags, fetch-verify, synchronize the
exact full-SHA range, review stable prose and replay; use one docs-only closure.
Stop on ambiguous source/remote delivery or scope beyond this bounded tool;
missing acceptance hardware/evidence does not block implementation delivery.

## Implementation checkpoint

Implemented `summarize`, strict bounded JSON loading and `write_report` with
non-overwriting publication, plus a CLI and a documented departmental input
schema. The input uses minute cohort counters and low-volume expected events;
it does not load millions of per-packet trace rows. Raw input bytes are retained
with SHA-256 in CLI reports. Missing events remain in nearest-rank p99 as
positive infinity and increment both missing and deadline-violation counters.
Phase/minute/five-minute/full-run summaries retain complete denominators.
Reports distinguish coverage gaps and numerical failures while always requiring
departmental review and never asserting SLO compliance.

Static manual review covered field/type/range/identity/time boundaries,
window membership, missing-as-infinite ordering, integer loss comparisons,
clock-error treatment, input-byte retention and no-clobber publication. Added
the script to the existing Python AST-only syntax target; `make python-check`
and `git diff --check` pass. These checks do not execute the new module or test
cases. The handoff documents all rejection and calculation boundaries for the
department, as well as unverified physical provenance and resource checks.

The scope is eight intended paths. No test suite, new CLI execution, traffic,
benchmark, knowledge test or acceptance run was performed. All behavioral
validation remains delegated. The source is delivered as untested implementation;
no performance evidence or completed-run artifact is manufactured. The existing
skill instruction giving user authority precedence covers this workflow change;
no generic skill change is required.

## Delivery results

Implementation feature `418e5dc5a1443068b25bad2e2ee6307f1338a3e1` contains
exactly the eight intended paths. Manual source and static AST/JSON/roadmap/
links/chronology/sensitive-data/diff reviews passed. Behavioral, benchmark,
acceptance and knowledge tests were not run under the user's explicit testing
delegation. This is not a tested or production-qualified release.

Main was pushed without force/tags and freshly fetched equal to HEAD and
FETCH_HEAD with a clean tree. Exact range
`99f84420e52d66718e5c6eadce3fd21278016d78..418e5dc5a1443068b25bad2e2ee6307f1338a3e1`
was synchronized to the local Vault. Iteration/index/MOC and current stable
implementation/test-ownership authority are verified; historical iteration
records were preserved. Replay retained Markdown content hash
`388277884c0b4c109f2fd31f1dada7c09ad1ed04d210851c870664fd67f811c9`.

R90-114 implementation is complete with tests delegated. This documentation
closure receives static checks, push/fetch and exact-range Vault verification,
without running knowledge or other test suites. R90-75 actual acceptance remains
with the department. Future engineering can implement the collection adapter;
missing production hardware or test results must not block development.
