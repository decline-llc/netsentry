# R90-118: Supplied SLO evidence bundle reconciliation

## Selection and authority

Clean fetched baseline `c6ae4bce4a8a8c834bc0c15f8aaf1c8b10c34906` matches
HEAD/origin/main/FETCH_HEAD. R90-117 feature and closure are in the local Vault
index/MOC. Recent delivery progressed from contract through reporter, adapter,
engine export and native UDP sender/capture; implementation remains untested by
user direction. R90-118 is the dependency-ready increment. R90-75 acceptance
belongs to the department; R90-59 publication retains its separate authority.
Refresh the forecast horizon to Sep 29–Dec 27; no date or hardware gate applies.

## Design and acceptance evidence

Implement a standard-library API/CLI consuming explicit optional sender directory,
capture summary, engine directory, adapter directory and report paths. Snapshot
only fixed documented filenames into a new private output directory; never follow
paths embedded in receipts. Stream raw ledgers with a finite total byte budget,
bound JSON metadata and individual JSONL rows, hash retained bytes, count rows,
and retain partial artifacts on missing, malformed or excessive input. Publish
a manifest last, distinguishing missing evidence from mismatches and complete
cross-checks requiring department review. No compliance/pass result exists.

Cross-check receipt inventories, exact bytes/digests/row counts, run IDs, available
origins and UDP destination port. Compare sender offers to adapter offers, engine
events to adapter events, manifest metadata to observations, and the report's
embedded observations and recomputed summary. Validate existing observation and
manifest schemas. Component completion must stay separate from the department's
execution/drain assertion; do not infer physical clock, oracle, fsync or process
exit truth. The checker does not replay the adapter's packet identity join.

| Criterion | Planned evidence |
| --- | --- |
| Retained bounded source evidence | Source review of fixed paths, regular-file snapshots, byte/row bounds, exclusive publication and partial files |
| Cross-artifact consistency | Manual review of receipt fields, identifiers, origins, port, source hashes and report recomputation paths; Python AST parsing |
| Honest outcome | Documentation of precedence, incomplete components, unverified claims and department fault matrix; no measurement artifacts generated |
| Delivery | Static JSON/roadmap/link/diff/sensitive-data review; verified push/fetch and exact-range local Vault sync/replay |

## Scope and limits

Intended paths: scripts/slo_bundle.py, Makefile syntax registration,
docs/slo-bundle.md, docs/slo-ingress.md, docs/slo-collect.md,
docs/performance-slo.md, roadmap, this plan, new bundle task state and active
R90-75 state. No network or services, fixtures, acceptance artifacts, new library,
release/tag, CI changes or tests. Unit/integration/benchmark/acceptance/knowledge
and CLI smoke execution are delegated and will not run. Static inspection is
implementation evidence only. Actual evidence, profile qualification, complete
packet replay, authenticity and physical boundaries require department review.
Stop only for new external/private authority or an undocumented input format;
documented schemas suffice for this increment without real run data.

## Delivery and next boundary

Persist this plan/state before editing. Implement and statically review one
increment, push/fetch verify, synchronize exact local Vault range and stable prose,
then one docs-only closure. Queue R90-119 for a bounded supplied-bundle comparison
contract/tool that diagnoses workload/hardware comparability without asserting
capacity; do not start it in this increment.

## Implementation and static review checkpoint

Implemented the ten intended paths. The bundle snapshots thirteen fixed input
files, preserves missing/malformed/truncated sources, validates component receipts,
compares run/origin/UDP port and fixture provenance, hashes/counts raw ledgers,
and recomputes the exact supplied-observation summary. JSONL parsing is streamed;
submission rows additionally require success and ordered timestamps. Receipt paths
cannot select arbitrary files. Finite per-file/total limits preserve prefixes as
gaps; changed sources become mismatches. Publication is exclusive and receipt-last.

Review clarified that the UDP fixture digest identifies the exact retained sender
JSONL, and that matching report bytes do not replay the adapter's packet join.
The bundle records its own and reporter source digests. A consistent failed or
inconclusive report retains that outcome; no compliance/pass result is emitted.
The documented department matrix covers schema, partial/tampered sources, output
faults, ordering, identity replay, real clock/durability and long-run scale.

`make python-check` only AST-parsed files; no module/CLI/business logic ran.
Static source review, 135 task JSONs, 123 unique roadmap row/Definition pairs,
ordered dependency/selection history, local links, diff and token-pattern review
pass. No tests, benchmark, acceptance, knowledge suite or traffic ran; the user
has delegated all such execution. No real/synthetic measurement bundle was
created. R90-119 is defined but not started. Existing skills already encode user
precedence and this delivery workflow; no skill edit is needed.
