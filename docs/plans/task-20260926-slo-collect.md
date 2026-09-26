# R90-115: Raw packet ledger adapter

## Selection and authority

Baseline `1a1893bc7c18077aa4bedb57f15e6956eccc0daa` is clean and freshly
fetched equal to HEAD/origin/main/FETCH_HEAD. R90-113 formalized the SLO;
R90-114 delivered the report consumer with tests delegated. Recent September
history moved from blocked-queue reconciliation through contract to implementation.
Its feature/closure are present in the local Vault index and MOC. R90-75 real
acceptance and R90-59 publication remain separate pending outcomes.

Select one adapter increment depending on R90-114. User explicitly delegates
all unit/integration/benchmark/acceptance/knowledge tests to the specialist
department; do not execute them or the new CLI. Static review remains allowed.
No missing hardware or acceptance results block this implementation delivery.

## Scope, evidence and intended paths

Implement `scripts/slo_collect.py`: consume a frozen run manifest, one JSONL
oracle row per offered eligible packet, and JSONL arrival/processing/durable
events. Use a temporary SQLite identity index to correlate out-of-order events
without loading every packet into RAM. Derive minute loss counters from unique
packets, keep all oracle alerts including missing observations, and produce the
existing reporter's exact observation schema. Retain exact raw source copies,
SHA-256 receipts and a no-overwrite output bundle; a final receipt marks adapter
completion, never SLO compliance. Inputs are department-exported records, not
implemented live runtime instrumentation.

| Acceptance | Planned evidence |
| --- | --- |
| Unique offered identities and correlated lifecycle events | Manual source review of SQLite constraints, strict JSON schemas and temporal checks |
| Missing alerts and packets retained | Review oracle-driven joins, complete packet criteria, minute cohort accounting |
| Recomputable report input | Review retained source bytes/digests, bounded lines/output, final receipt and reporter schema validation |
| Honest collection boundary | Document external live timestamps, terminal processing and durable instrumentation requirements |
| Isolated delivery | AST-only syntax, static JSON/roadmap/link/diff review, verified Git push/fetch and exact Vault ranges |

Paths: new script; Makefile syntax registration; `docs/slo-collect.md`;
`docs/slo-report.md`; `docs/performance-slo.md`; rolling roadmap; this plan;
`docs/tasks/task-state-20260926-slo-collect.json`; active R90-75 state.

## Risks and non-goals

Behavioral validation is not run. Department must test duplicates, invalid and
out-of-order records, clocks, time boundaries, missing alerts, multi-rule packets,
window counts, limits, preservation and interrupted/disk-full writes. SQLite and
raw retention consume disk proportional to input; throughput is unmeasured.
No traffic, SSH, protocol/runtime/storage changes, new third-party dependencies,
CI weakening, compliance assertion, release/tag/image publication or invented
artifacts. Oracle completeness and physical timestamp truth remain external.
Current processed counter is not terminal completion; histograms are not E2E.

## Delivery

Persisted before implementation. Review source and static structure; commit and
push one feature without force/tags, fetch verify, synchronize exact full-SHA
Vault range and reconcile stable prose. Record verified facts in one docs-only
closure and deliver that exact range too. Stop on ambiguous delivery or scope
beyond this adapter. Future runtime collection integration is a separate increment.

## Implementation checkpoint

Implemented the nine-path scope. SQLite uniqueness and fixed parameterized
queries correlate all offers before unordered lifecycle rows. Final temporal
checks reject inconsistent observations; terminal markers with missing expected
durable alerts do not count as successful packets. Oracle alerts remain in
reporter input with null missing timestamps. Raw manifest/JSONL copies and the
observation output are bound by byte counts and SHA-256, with adapter source
hash and a final receipt. Partial errors preserve output for inspection.

Manual review covered field/type/time/identity invariants, missing observations,
minute membership, SQLite joins, resource bounds, exact source retention and
receipt-last/no-overwrite behavior. AST-only `make python-check` passes; 132
state JSONs parse, 120 unique row/Definition pairs agree, changed-document links
resolve and diff formatting is clean. No module/business logic, CLI, behavioral
suite, benchmark, acceptance traffic or knowledge test was executed. Behavioral
risks remain delegated; docs explicitly describe external acquisition and
unmeasured adapter scale. R90-116 runtime exports are queued, not started.

## Delivery results

Feature `f2be37c0b4d9b867a15dc1f3a4d2f460d78e7be7` contains exactly the nine planned paths.
Static AST/source/JSON/roadmap/link/diff/sensitive-data review completed; all
behavioral, benchmark, acceptance and knowledge tests remain not run under the
user's explicit delegation. No traffic or completed-run evidence was generated.

Push without force/tags succeeded. Fresh fetch verified clean
HEAD/origin/main/FETCH_HEAD equality. Exact range `1a1893bc7c18077aa4bedb57f15e6956eccc0daa..f2be37c0b4d9b867a15dc1f3a4d2f460d78e7be7`
was synchronized to the local Vault; iteration/index/MOC were verified. Four
stable notes now explain the adapter, test ownership, raw retention and pending
live instrumentation; historical iteration records are unchanged. Replay
preserved Markdown SHA-256 `02a01880173ce1a2f9c2125166ae1b4db312129a547293c70da48f885d1d62d9`.

R90-115 implementation is complete with tests delegated. R90-116 is ready for
its own plan and bounded runtime export increment; it is not started here.
R90-75 remains departmental acceptance, not a development blocker. This single
docs-only closure receives static review, push/fetch and exact-range Vault
verification too; no tests or knowledge suite are run.
