# R90-151: consistent negative query limits across store modes

## Scope and selection audit

Ready after completed R90-150 and existing public Store.Query contract.
Fresh clean fetched main HEAD/origin/main/FETCH_HEAD is `ae191f620abf6145c91a4722c0eed36d876fe311`.
81-commit Sep 5–Oct 3 phase: supplied-evidence SLO tools followed by core
snapshot/concurrency/completion/privacy/pagination/storage arithmetic repairs.
R90-150 exact six-path feature and three-path closure Git/Vault scope/note/index/
MOC verified, no missing delivery or new qualifying R90-75 outcome. Sole
unfinished item R90-75 remains independent asynchronous departmental acceptance.
Reconcile the empty local ready queue inside this source-grounded compatibility
repair, without a separate audit-only increment. Horizon Oct 3–Dec 31 unchanged.
Vault baseline 381 Markdown files/14 full stable backups/334 immutable hashes;
snapshot JSON SHA-256
`8000f624e199b1854285626f8b32596e9eb10411c30b74fd3dbf28d35c7c2287`.
Pinned owning engine module resolves Go 1.26.8 linux/amd64.

Six paths: `engine/internal/alert/store.go`, new
`engine/internal/alert/query_limit_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-shard-query-limit.json` and rolling roadmap.
Persist plan/state before behavior/architecture/roadmap edits. Runtime only
queryDailyShards limit normalization: negative means all collected rows, zero
remains default 1000, positive remains explicit limit. Preserve offsets and the
R90-150 remaining-length bounds; do not change primary SQL or HTTP parser.

## Source evidence, risk and authority

queryAlertsDB already treats every negative limit as total filtered rows, while
queryDailyShards maps all nonpositive limits to 1000. Daily reads already collect
all filtered rows from all selected shards, so changing the final slice limit
adds no scan or collection allocation; it returns all rather than truncating.
R90-150 explicitly left this mode inconsistency outside its arithmetic scope;
this separate increment addresses it without rewriting historical authority.
Risk medium: filtering/count/order, over-1000 boundary and normalization parity.
All behavioral/race/CLI/full-suite/scanner/knowledge/acceptance execution remains
**not run; delegated by user**. Static checks/compilation allowed; no runtime,
SQL, memory-performance, file-byte, race or SLO pass from compile evidence.

## Acceptance and evidence map

1. Negative daily limit uses len(all) before existing offset/bound calculation;
   zero defaults to 1000; positive unchanged. Exact source transform plus
   preservation of every other tracked engine file and all surrounding SQL,
   filters/count/sort/lifecycle/recovery/writer/HTTP source.
2. Author public real SQLite Store.Query regression in primary/daily modes with
   1005 distinct synthetic rows and 1003 severity-high rows across two dates in
   encoded directory. Assert two daily files and full positive-limit baseline
   ordering/severity/aggregation values; copy complete Alert values. Direct
   negative -1/-2, filtered >1000 nonzero-offset, default zero, positive >1000,
   normal positive, tail/at/past/MaxInt offsets and negative-offset queries;
   compare hardcoded expected index slices/full values and filtered total.
   Uncapped assertions must exceed 1000 and use actual Store.Query, no fake,
   recover/skip or weak small fixture. Check Count/health and unchanged full
   logical rows after all queries, without a physical byte-preservation claim.
3. Separate primary/daily empty-filter and empty-store direct cases for negative,
   zero and positive limits plus offset boundaries; success/count zero/no panic.
4. Final pinned Go 1.26.8 alert/API compile-only complete chain; Go parse-format,
   docs-check, task JSON, full unique roadmap row/Definition multisets/full
   contracts/prior Definitions/R90-75/history/horizon/links/fences/six-path/diff/
   sensitive review. All execution suites delegated and unrun.
5. Deliver feature plus one three-path docs-only closure with exact full SHA
   push/fresh-fetch refs and exact Vault note/index/MOC, 14 current sections,
   entire previous current prose archived, original topic tails/334 immutable
   hashes retained, all-hash identical-range replay. Refresh queue and stop.

## Non-goals and stop conditions

No HTTP/API/filter/SQL/schema/count-overflow/lifecycle/recovery/writer/retention
change, cursor API, new cap, dependency/toolchain/test execution/IPv6/publication.
No streaming or snapshot-isolation guarantee. Stop for competing changes,
ambiguous validation/Git/Vault, or new product/private-input/external authority.
Only this increment and one closure per trigger; no following implementation.
Existing workflow suffices; no generic skill edit just to narrate delivery.


## Implementation and validation checkpoint

Runtime changes exactly one daily-limit normalization block: negative len(all),
zero 1000, positive unchanged. All other tracked engine files and primary SQL/
HTTP/R90-150 bounds/filter/count/sort/lifecycle/recovery/writer source match
baseline. Two direct functions author 1005-row/1003-high fixtures across two
actual daily files, 13 real Query cases per primary/daily mode plus 64 empty
store/filter combinations. Returned counts and complete copied Alert values,
order/offset/health/Count/final logical rows are asserted; negative filtered
nonzero-offset returns exceed 1000. No fake Query/recover/skip substitutes the
boundary. Final pinned Go 1.26.8 alert/API complete compile-only chain passed;
external binaries never invoked. Source/direct-boundary/Go-format/docs/169 JSON/
155 full unique roadmap/full contract/prior Definitions/R90-75/history/horizon/
links/fences/six-path/diff/sensitive review passes. All execution delegated.
No runtime/SQL/performance outcome is claimed; no validation/compile failure.

Direct truncation evidence review identified a reusable local skill improvement:
use a fixture larger than the truncation boundary and assert count plus content.
The generic netsentry-next Execute instruction 8 is refined; frontmatter,
numbering and Markdown fences pass. This local-only edit is separate from Git
and adds no repository detail. No skill edit merely narrates the outcome.
381-file Vault baseline unchanged. Archive entire prior current prose before
replacement; retain original topic tails and 334 immutable hashes. Exact feature
and one three-path docs-only closure remain; no subsequent increment starts.


## Delivery and queue closeout

Feature `e55c784f5dee3fe45b42b9e802608f5fc2a99760` contains exactly six planned paths. Isolated
`fix/r90-151-shard-query-limit` fast-forwarded freshly verified main;
push/fresh-fetch verified clean HEAD/origin/main/FETCH_HEAD at that SHA.
Exact range `ae191f620abf6145c91a4722c0eed36d876fe311..e55c784f5dee3fe45b42b9e802608f5fc2a99760` six-path scope, iteration note
`04-开发迭代记录/2026-10-03-e55c784f5d-CI知识同步.md`, full index and MOC are verified. Fourteen stable
current sections reconciled; entire prior R90-150 current prose retained under
explicit historical headings and original topic tails intact. All 334 baseline
immutable hashes unchanged; identical replay preserves 382 Markdown files,
snapshot JSON SHA-256 `b79bb3e1447875dc16a67792e394ac8969e848a00fefc526ddc67c858ed7c7d0`. Unique existing local sibling Vault
selected explicitly.

Acceptance matches plan: runtime changes only daily Query limit normalization.
Negative uses len(all), zero remains default 1000, positive unchanged; offset
normalization and safe bounds retain behavior. Existing collection already reads
all filtered rows; this only changes returned slice length. All other tracked
engine files and primary SQL/HTTP/filter/count/sort/lifecycle/recovery/writer
source match baseline. Historical R90-150 negative-limit non-goal remains.
Two authored direct functions use 1005 rows, 1003 high-severity rows across two
dates/two actual daily files/encoded path. Thirteen public real Query cases per
primary/daily mode cover negative -1/-2, filtered nonzero-offset >1000 returns,
zero/positive/offset compatibility; 64 empty store/filter combinations cover
limit/offset boundaries. Positive-limit baseline verifies fixed fixture order/
severity/aggregation, then copies full Alert values before boundary queries.
Expected indices/count/full content/health/Count/final logical rows are asserted,
without fake Query/recover/skip or a too-small fixture masking the truncation.

Final Go 1.26.8 alert/API complete compile-only chain and source/direct-boundary/
Go-format/docs/169 JSON/155 full unique roadmap/full contracts/prior Definitions/
R90-75/history/horizon/links/fences/six-path/diff/sensitive checks pass. No binary
or behavioral/race/CLI/full-suite/scanner/knowledge/acceptance suite executed;
all delegated by user. No runtime/SQL/performance/file-byte/race/SLO pass claim,
implementation/validation/compile/delivery failure or Vault topic loss.
Generic local skill refinement requires truncation fixtures exceeding boundary
with count/content assertions; frontmatter/numbering/fences pass, separate from
repository commit. Only planning deviation: empty ready queue repaired inside
source-grounded compatibility increment. Oct 3–Dec 31 horizon/history unchanged.

One three-path docs-only record closes this increment: resolve full SHA from
Git, push/fresh-fetch and exact feature-tip..closure-tip Vault verification before
reporting; no self-reference follow-up closure. No currently defined dependency-
ready local item after queue refresh. R90-75 retains its full independent
asynchronous departmental contract without blocking development. Next trigger
verifies fetched closure/Vault, audits fresh core-code evidence/queue and persists
separate eligible plan before edits. No subsequent implementation starts here.
IPv6/external publication need separate authority; do not repeat R90-150/R90-151
delivery or R90-59 publication.
