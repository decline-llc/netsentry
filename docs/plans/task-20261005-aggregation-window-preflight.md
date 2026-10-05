# R90-180: reject subsecond aggregation windows before startup

## Selection and authority

Clean freshly fetched HEAD/origin/main/FETCH_HEAD baseline:
`ab06fe32219e4abb69a4474c43f4ad33bb593c7d`. R90-179 feature and single
closure exact Git/Vault ranges, notes/index/versioned bounded MOC verified.
Four-week phase audit covers 140 commits with delivery records, 183 unique
roadmap row/Definition pairs, 440 Vault Markdown files, 393 immutable iteration
records and fourteen current stable backups. No missing delivery record or
qualifying R90-75 outcome found. Prior daily fixture/reproduction claims were
corrected by R90-179; recent implementation execution debt remains delegated.

The local ready queue is empty. Source inference: Open accepts any positive
AggregationWindow; normalizeAlert truncates timestamps to that duration, but
alertAggregationID uses WindowStart.Unix, discarding nanoseconds. Distinct
subsecond windows for the same aggregation tuple can share a primary key while
the UPSERT conflict target includes their different window_start values. Such
a batch can fail after recovery append. Select the smallest safe default:
reject positive windows below one second before any path/clock/filesystem work,
preserving durable IDs and all windows at least one second, including fractional
durations. This is source evidence, not an executed failure.

R90-180 depends on verified R90-179 and existing ID/aggregation contracts.
R90-75 retains its complete independent departmental acceptance. Refresh the
unfinished horizon to Oct 5–Jan 2 after the current-date update; prior completed
windows/history remain unchanged. No date-gated dependency or new product
authority is inferred.

## Scope, risks, non-goals and stop condition

Seven intended paths: engine/internal/alert/store.go; new
engine/internal/alert/store_aggregation_window_preflight_test.go;
docs/architecture.md; docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261005-aggregation-window-preflight.json; rolling roadmap.

Low risk: only positive durations below time.Second acquire an early diagnostic.
Insert after current cancellation/durable-WAL/busy-overflow/journal validation,
before defaults/path resolution/recovery read/directory creation/writable open.
Nonpositive durations still default to one minute; one second and all larger
durations retain their current semantics, including fractional windows. Config
already supplies whole-second durations, so no CLI/config behavior change.

No ID/schema/recovery format/migration, retention, active operation, API or
dependency change. No subsecond support claim, private input, department message,
publication or SLO acceptance. Stop for ambiguous compile/static/Git/Vault,
competing edits, new compatibility/product authority or another increment.
Standing user delegation keeps behavioral/race/full/scanner/knowledge/traffic/
acceptance checks **not run; delegated by user**.

## Acceptance and evidence map

| Acceptance | Planned evidence |
|---|---|
| Reject every positive subsecond window before side effects | Public Open at 1 ns, 1 ms, 250 ms, 500 ms and 1 s minus 1 ns, primary/daily and DELETE/durable WAL; absent/healthy/corrupt-sidecars/malformed-recovery/occupied-parent fixtures; exact diagnostic/nil Store, zero clock calls, caller-options and full tree bytes/modes/membership preservation; independently encoded read-only retained-row observer before rejection |
| Preserve established earlier errors | Direct public canceled/expired context, durable-WAL policy, busy overflow and unsupported journal cases with competing subsecond input; exact sentinel or diagnostic, nil Store, zero clock calls and artifacts unchanged |
| Retain accepted/default windows and identity | Public primary/daily DELETE/WAL Open/write/Count/List/Query/close/reopen at nonpositive defaults, exactly 1 s, 1 s plus 1 ns, 1.5 s and 1 minute; two same-tuple alerts one effective window apart, distinct established IDs and full normalized contents/counts/events through independent encoded read-only observation; caller input preserved |
| Reviewable delivery and honest evidence | Pinned Go 1.26.8 alert/API/pipeline/cmd compile-only; exact three-line source delta/format/docs/JSON/full roadmap multisets/prior history/full R90-75/frozen handoff/links/fences/seven-path/sensitive/diff; non-force push/fresh fetch/exact Vault stable preservation/replay and one docs-only closure |

All named declarations are authored and compile-reviewed, not executed; their
assertions must reach the actual primary Path or daily Dir/initial-clock resource.
Compilation establishes no rejection, artifact, durability, HTTP or race pass.

## Delivery and forward queue

Complete exactly R90-180. Commit seven intended paths, non-force push and freshly
verify exact clean refs; sync the full-SHA range to the unique existing sibling
local Vault. Reconcile current stable authority while preserving all substantive
topic/status prose under historical headings and immutable iterations. Replay
the identical range and compare all Markdown hashes. Record verified feature
delivery in one three-path docs-only closure of this increment; push/fetch/sync
that second exact range too. No self-reference closure or repeated delivery.

Refresh the future queue without starting its next increment. No other defined
local ready item; next trigger audits fresh source/history/queue before a separate
eligible plan. R90-75 full independent departmental acceptance remains outstanding.

## Compile and static checkpoint (2026-10-05)

The runtime delta is exactly the planned three-line guard, after earlier
validation and before path/clock/filesystem work. Three public declarations
cover 100 rejected-input fixtures, five earlier-error causes (busy overflow
explicitly skips native 32-bit), and 28 accepted/default-window controls. Daily
fixtures derive their actual path from Dir/initial Now. Healthy inputs are
observed before rejection through an independently encoded read-only handle;
accepted rows/events compare every durable column with independently computed
expected values before and after reopen. No production normalizer or identity
helper substitutes for the expected boundary.

Pinned Go 1.26.8 compiled alert, API, pipeline and cmd/netsentry binaries outside
the repository without executing them. The first fail-fast attempt stopped in
alert on an unused test import; removed it and reran the complete four-package
chain successfully. After strengthening independent expected durable columns,
the complete chain passed again. Later packages in the failed attempt are not
counted as evidence. No unresolved compile/static failure remains.

Complete static review passed: exact source delta and precedence, pinned format,
198 task JSON, 184 unique complete roadmap row/Definition multisets, all 183
prior Definitions/history/full R90-75 terms/frozen handoff retained, unfinished
horizon/links/fences/seven-path scope/sensitive added diff/diff checks. All 440
Vault Markdown hashes still match the selection snapshot. No behavioral, race,
full, scanner, knowledge, traffic or acceptance execution: **not run; delegated
by user**. Compilation establishes no runtime or persistence pass.

Deviation review: empty queue/horizon correction was recorded at selection;
the unused import was a corrected one-off compile error. Existing generic skill
instructions cover public option/resource tracing and fail-fast reruns; no
reusable skill change is warranted. Feature push/fetch/Vault and its one
docs-only closure remain pending.

## Verified feature delivery and acceptance comparison (2026-10-05)

Feature `d7af8e3e20a7f71f16852b48a83c6412dfbaf2dd` contains exactly the seven intended paths. Non-force push
succeeded; a fresh fetch verified clean HEAD/origin/main/FETCH_HEAD at that
exact full SHA. Local Vault sync used full range `ab06fe32219e4abb69a4474c43f4ad33bb593c7d..d7af8e3e20a7f71f16852b48a83c6412dfbaf2dd`.
Iteration `04-开发迭代记录/2026-10-05-d7af8e3e20-CI知识同步.md`, full commit index and versioned bounded MOC links verified,
with abbreviated generated identifiers resolved through Git. Fourteen current
stable notes reconcile actual feature authority/date/references; all prior
non-generated topic/status prose is retained byte-for-byte under explicit
historical headings. All 393 prior immutable iteration records are unchanged.
Identical exact-range replay preserved all 441 Markdown hashes; snapshot JSON
SHA-256: `67cb8d1ded648cd6dd5d27aae1b6687d2e41c6cbbefa5f41ea04f65db035b295`.

Acceptance comparison reaches every planned direct boundary in the authored
declarations: five positive subsecond durations and all 100 resource/state
combinations; nil Store/exact error/zero clock/unchanged caller options/tree
bytes/modes/membership; healthy retained-row/all-column read-only observation
established before rejection; five earlier-error causes with precise native
32-bit overflow skip; and 28 accepted/default controls, including 1 s plus 1 ns
and 1.5 s. Two same-tuple alerts establish distinct expected windows/IDs/events;
full List/Query/Count/totals, independent expected every alerts/events column,
close/reopen, health and caller inputs are asserted. Actual daily resources
follow Dir/initial Now, never an ignored Options.Path. Compile/static review
passed; these are authored assertions, not runtime or persistence results.

Exactly the planned bounded input rejection is implemented. Nonpositive
defaults, accepted durations and durable IDs/schema/recovery format remain
unchanged. Architecture/handoff/current forecast are reconciled; all completed
history and full independent R90-75 acceptance terms are retained. The first
compile failure/import fix and complete reruns are recorded above; no unresolved
failure or scope expansion. Existing generic skill covers the workflow; no edit.
Behavioral/race/full/scanner/knowledge/traffic/acceptance **not run; delegated by
user**. Implementation/evidence authorship is complete; execution debt remains.

## Single closure and resume authority

This three-path documentation delivery record closes R90-180. Resolve its full
SHA from Git after commit, verify non-force push/fresh clean exact refs, and sync
the exact feature..closure range; check its three-path iteration/index/bounded
MOC, fourteen stable reconciliations, prior prose/iterations and identical replay.
The verified feature SHA above is a historical boundary, not an instruction to
repeat delivery. Do not create a self-reference closure or another increment.
Future sessions resolve this sole closure through fresh Git refs and repair
only missing evidence. No other defined local ready increment; next trigger
audits fresh source/history/queue before a separate eligible persisted plan.
R90-75 full independent acceptance remains outstanding; unfinished forecast
Oct 5–Jan 2. No next increment started.
