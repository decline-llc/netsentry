# R90-152: include historical daily shards in Store.List

## Selection and baseline audit

Clean fetched main HEAD/origin/main/FETCH_HEAD: `4cde70cf49df6ea2f2b7e250f311603dbff86819`.
R90-151 exact six-path feature and three-path closure Git/Vault scope, note,
index and MOC verified. Sep 5–Oct 3 phase has 83 commits covering supplied SLO
evidence tools followed by core correctness repairs; implementation delivery
remains separate from user-delegated behavioral acceptance. No new R90-75
outcome is present. Empty local ready queue reconciled inside this bounded
source-grounded repair. Oct 3–Dec 31 horizon and independent R90-75 unchanged.
Vault Markdown hashes and complete current stable note contents backed up outside
Git before edits. Owning engine module resolves pinned Go 1.26.8 linux/amd64.

## Scope, risk and authority

Six paths: `engine/internal/alert/store.go`, new
`engine/internal/alert/list_shards_test.go`, `docs/architecture.md`, this plan,
`docs/tasks/task-state-20261003-shard-list.json` and rolling roadmap.
Persist this plan/state before source or documentation changes.
List currently queries only s.db even in daily mode, omitting historical rows
that Query/Count expose. Inside the existing shared lifecycle ownership, daily
List delegates to queryDailyShards with explicit limit 1000 and zero offset;
primary List SQL stays unchanged. Do not call public Query recursively under
the lifecycle lock. Reuse existing ordering/read-only/error/health handling.
Risk medium: additional historical reads can expose existing shard errors and
scan/collect the complete set, as Query already does; no performance guarantee.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
is **not run; delegated by user** under the retained Sep 25 instruction.
Static review/docs checking and compilation only; never execute test binaries.

## Acceptance mapped to evidence

1. Exact daily-only dispatch after lifecycle acquisition; retain primary SQL,
   Query/Count/filter/sort/read-only/lifecycle/writer/recovery/API source.
   Evidence: source diff and all other tracked engine files match baseline.
2. Author real public List regression in primary/daily modes with 1005 distinct
   rows across two dates and an encoded directory; assert two actual daily
   files, deterministic timestamp and ID tie ordering, exactly 1000 complete
   copied expected Alert values including historical rows, healthy state,
   unchanged Count/full logical Query results. Fixture exceeds cap; no fake,
   recover or skip. Also test current-empty historical-only rows via public List.
3. Author empty/pre-canceled/closed List compatibility across both modes and a
   corrupt historical shard List rejection with degraded health and unchanged
   corrupt bytes. These authored assertions are unexecuted departmental work.
4. Final Go 1.26.8 alert/API complete compile-only chain, Go format/parse,
   docs-check, task JSON, unique roadmap row/Definition multisets, all prior
   Definitions/R90-75/horizon/history/links/fences/six-path/diff/sensitive review.
5. Feature and one docs-only closure: commit, push, fresh fetch, exact full SHA
   refs and Vault note/index/MOC; reconcile affected current stable prose while
   preserving previous substantive prose and immutable notes; identical-range
   replay. Refresh queue without starting a second increment.

## Non-goals and stop conditions

No primary SQL, Query limits, HTTP schema, filters, Count, lifecycle, retention,
recovery, writer, schema, toolchain/dependency, execution suites, IPv6, release,
external publication or new snapshot/streaming guarantees. No incident claim.
Stop for competing work, ambiguous compilation/static/Git/Vault evidence, or
new product/private-input/external authority. Existing skills suffice; no skill
edit solely to narrate this outcome. Complete exactly one increment.


## Implementation and validation checkpoint

Runtime adds exactly four lines in List after shared lifecycle acquisition:
daily-only private queryDailyShards call with explicit limit 1000, return its
alerts/error. Primary List SQL and all other tracked engine files plus Query/
Count/filter/sort/read-only/lifecycle/writer/recovery/API source match baseline.
Four direct regression functions are authored: real 1005-row primary/daily
fixtures across two dates/two actual files/encoded directory, reversed insertion
and paired equal timestamps exercise ordering/ID tie break; repeated List checks
exactly 1000 full copied Alert values including historical rows, healthy state,
Count and all logical rows. Current-empty historical-only case confirms empty
primary does not imply empty daily List. Empty/pre-canceled/closed cases cover
both modes; corrupt historical List asserts shard error/degraded diagnostic and
unchanged corrupt bytes. No fake List/recover/skip hides the promised boundary.

Pinned Go 1.26.8 alert/API complete compile-only chain passed; external binaries
not invoked. Exact source/direct-boundary/Go-format/docs/170 JSON/156 full unique
roadmap pairs/all prior Definitions/R90-75/history/horizon/links/fences/six-path/
diff/sensitive review passes. Behavioral/race/CLI/full-suite/scanner/knowledge/
traffic/acceptance execution **not run; delegated by user**. No runtime/SQL/
performance/race/physical-preservation/SLO success is inferred from compilation.

First static comparison found one added separator newline at previous Definition
boundary; restored it and reran complete static/docs/diff chain successfully.
No source/compile failure or scope change. Existing skill instructions suffice;
no generic skill update. Feature and one docs-only delivery closure remain;
no subsequent increment is started.
