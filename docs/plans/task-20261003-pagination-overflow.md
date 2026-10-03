# R90-149: checked alert pagination arithmetic

## Scope and authority

Source-grounded core HTTP correctness repair after completed R90-148. The active
Oct 3–Dec 31 horizon is unchanged. R90-75 remains independent asynchronous
acceptance. All behavioral/race/CLI/full-suite/scanner/knowledge/acceptance
execution remains **not run; delegated by user**. Static checks and compilation
only are permitted. No dependency, toolchain, SQL, route or publication change.

Seven paths: `engine/internal/api/pagination.go`, new
`engine/internal/api/pagination_overflow_test.go`, `docs/api-reference.md`,
`docs/architecture.md`, this plan, `docs/tasks/task-state-20261003-pagination-overflow.json`,
and `docs/plans/rolling-90-day-roadmap.md`. Persist plan/state before runtime,
architecture/API or roadmap edits; use isolated branch.

## Audit and source evidence

Fresh clean HEAD/origin/main/FETCH_HEAD is 6ba400fb15d238846a0441be02b516cd7d775e9a. The last 28-day phase contains
77 commits: supplied-evidence SLO tooling followed by snapshot/concurrency,
completion-counter and redaction core code. R90-148 exact eight-path feature and
three-path closure Git/Vault note/index/MOC are verified; no missing delivery or
new qualifying departmental evidence. Vault baseline: 377 Markdown files,
14 stable full-content backups, 330 immutable iteration notes; snapshot JSON
SHA-256 `b9a3b7888fffcf9b9249ec332a022b07f8056ee02ba9425c8c7befe57115bf3d`.
Only unfinished queue item is R90-75. Register this bounded source-grounded
repair inside implementation rather than shipping an audit-only increment.

parsePagination accepts any positive int page and per_page up to 100, then
pageBounds and toAlertQuery multiply (page-1)*per_page without representability
checks. An accepted extreme page can produce a negative slice index or SQL
offset. Separately, start+per_page can overflow before clamping to a large total.
The public alerts handler parses pagination before either List or Query, so
one parser guard protects both backends without modifying SQL or routing.
Risk: medium; arithmetic boundary, error compatibility and backend consistency.
Pinned owning module engine resolves Go 1.26.8 linux/amd64; no test binaries run.

## Acceptance and evidence map

1. After existing positive-int and per_page-limit diagnostics, reject an offset
   above platform int maximum with `page and per_page exceed maximum pagination offset`.
   Compare page-1 with maxInt/per_page before multiplication. Preserve defaults,
   numeric parsing and no arbitrary lower page cap. Direct parser regressions
   cover normal/defaults, per_page 1/20/100 largest representable pages and first
   overflowing page where representable, plus existing invalid diagnostics.
2. Clamp page bounds using total-start before adding; accept representable large
   offsets and return an empty fallback page when past total. Direct bound cases
   include max-int total/end and empty/ordinary/exact-end/past-end ranges without
   allocating large slices.
3. Through public Handler/ServeHTTP, both fallback
   and query-capable stores reject overflow with HTTP 400 VALIDATION_ERROR,
   request ID and exact detail, before any List/Query/Count call. Cover default
   size and explicit 100. No panic recovery masks the regression.
4. Both handlers accept representable extreme pages, preserve page/per_page/total
   and empty data; ordinary severity-filtered page keeps expected alert. Query
   spy checks exact nonnegative offset/limit and severity; fallback spy checks
   List only. No real DB required because SQL arithmetic source is unchanged.
5. Complete pinned api/alert compile-only chain, Go parse-format, docs-check,
   all task JSON, complete unique roadmap row/Definition multisets, all prior
   Definition/R90-75 preservation, ordered history, 90-day window, links/fences,
   exact seven-path cumulative scope, diff and anchored sensitive review. No
   executed test result or runtime/SQL outcome is claimed.
6. Deliver feature plus one docs-only closure: exact old/new SHA push/fetch
   verification, generated note/index/MOC, reconcile all 14 stable status notes,
   preserve topical tails/330 immutable notes, replay identical range with all
   Markdown hashes unchanged. Closure only roadmap/plan/state, no self-reference
   follow-up commit. Refresh queue without beginning another increment.

## Non-goals and stop conditions

Do not change route/filter/store/query schema, arbitrary pagination cap, nil
store contract, cursor pagination, count overflow, mutation auth, test execution,
R90-75 evidence, IPv6 or external release authority. Stop for unexpected changes,
ambiguous validation, new product/private-data/external authority, or failure
of exact Git/Vault scope verification. No subsequent increment in this trigger.


## Implementation and validation checkpoint

Runtime change is confined to pagination.go: the division guard follows all
existing diagnostics; fallback clamps remaining length before end addition.
All other tracked engine files match fetched baseline. Four authored regressions
contain 17 parser boundary/compatibility cases, nine bound cases including
MaxInt total, four HTTP reject-before-store cases and eight HTTP accepted/filter/
exact-offset cases across both interfaces. No recovery masks a fallback panic.
Final source was compiled with pinned Go 1.26.8 in the complete api/alert
`go test -c` chain; binaries outside the repository were never invoked.
Static source/direct-boundary/Go-format/docs/167 JSON/153 full unique roadmap/
full contract/prior Definitions/R90-75/history/horizon/links/fences/seven-path/
diff/sensitive review passes. Every execution suite remains user-delegated.

First static preservation check detected one separator newline appended to the
prior R90-148 Definition by insertion. A diagnostic assertion initially counted
two newlines and failed without mutation; exact comparison then restored the
single extra newline and the complete static/docs/diff chain reran successfully.
No runtime edit was needed for this formatting deviation; no compile failure.
Existing generic structural skill guidance is sufficient, so no skill change.
377-file Vault baseline is unchanged. In addition to original topic tails,
retain the previous current section's substantive R90-148 material under an
explicit historical heading before replacing current status. Feature delivery
and one three-path docs-only closure remain; no subsequent increment starts.
