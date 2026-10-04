# R90-174: Prometheus severity label escaping

## Selection and authority

Fresh fetched clean main `2488c394a394c363a8f19578543cec6983a69014`; R90-173 feature/closure exact Git/Vault
ranges, note/index/MOC verified. Four-week audit covers 127 commits with
note/index coverage, patched candidate/toolchain, SLO input work and native
correctness handoff. No missing delivery or qualifying R90-75 outcome. All 177
unique row/Definition pairs agree. Empty local ready queue restored from source:
Stats accepts dynamic severity labels, but RenderPrometheus uses Go %q, which
can emit unsupported tab/carriage-return/hex/Unicode escapes. This is source
inference, not an executed failure. The authoritative
[Prometheus text format](https://prometheus.io/docs/instrumenting/exposition_formats/)
requires escaping only backslash, double quote and line feed in UTF-8 values.

## Scope, non-goals, risk and stop condition

Eight intended paths: engine/internal/stats/stats.go; new
engine/internal/stats/label_encoding_test.go; new
engine/internal/api/metrics_label_encoding_test.go; docs/architecture.md;
docs/correctness-validation-handoff.md; this plan;
docs/tasks/task-state-20261004-prometheus-labels.json; rolling roadmap.
Low risk: change only severity label encoding with a shared immutable
strings.Replacer; preserve counters, label identity, raw sort order and all
other exposition. No new dependency/parser, label normalization, invalid UTF-8
policy, histogram/name/gauge contract, API shape, release or SLO claim.
Stop for ambiguous static/compile/Git/Vault evidence, competing edits or new
authority. Behavioral/race/full/scanner/knowledge/traffic/acceptance checks
**not run; delegated by user**. Compile-only binaries remain unexecuted.

## Acceptance mapped to evidence

1. Source uses exact Prometheus escaping for severity values; canonical labels
retain byte-identical lines. Static transform review plus renderer assertions.
2. Public ObserveAlerts/Snapshot/RenderPrometheus regressions assert literal
expected bytes for backslash/quote/newline/tab/carriage return/control/DEL/Unicode/
nonbreaking-space and mixed escape cases; independently decode only the three
legal escapes to original values, preserve raw identity/counts/input/snapshot,
assert deterministic sorted lines and no injected samples. Authored and compiled;
execution delegated. Invalid UTF-8 input is explicitly outside scope.
3. Real HTTP Handler /api/metrics regression uses existing store/queue/rule
fixtures and real Stats, asserts exact body label lines, 200/content type,
repeatability of label lines and original health JSON keys/label identity.
4. Architecture and departmental follow-up link the new source/functions/plan/
state without changing the historical 29-repair inventory. Preflight Go 1.26.8,
module-relative stats/API/cmd/pipeline compile-only; docs-check, JSON/roadmap
multisets/preserved prior Definitions/history/links/fences/scope/sensitive/diff
review and unchanged baseline Vault hashes. Test split and full R90-75 contract
unchanged; horizon remains Oct 4–Jan 1.
5. One conventional feature commit plus one optional docs-only closure. Push
without force; fresh fetched refs; exact full-SHA Vault note/index/MOC, reconcile
changed stable authority preserving all prior substantive prose and immutable
notes, replay exact range and compare all Markdown hashes. No next increment.

## Checkpoint

Plan/state persisted before behavior or documentation edits. Existing generic
workflow rules suffice; no skill update is justified by selection alone.


## R90-174 Compile and Static Checkpoint (2026-10-04)

Runtime diff is exactly the shared three-pair immutable strings.Replacer and
severity formatting call; counters, raw keys/order and every other runtime line
remain unchanged. Two renderer declarations cover twelve literal-byte cases,
returned-line decoding through only legal escapes, input/snapshot preservation,
repeatability, total/counts, raw sort order and distinct newline/literal-escape
identities. One HTTP declaration reaches real Stats/Handler with existing
store/queue/rule fixtures, asserts exact mixed control/escape/Unicode bytes,
canonical labels, total, status/content type, repeated label lines and original
health JSON keys/raw identity. No new parser dependency, network/scraper test,
private seam, sleep or invalid-UTF-8 claim. All assertions unexecuted.

Preflighted exact Go 1.26.8/module roots; complete fail-fast stats/API/cmd/pipeline
compile-only chain passed; binaries unexecuted. Static exact source/format/docs/
192 JSON/178 unique roadmap pairs/all 177 prior Definitions and roadmap history/
R90-75/split/horizon/historical handoff/links/fences/eight-path scope/sensitive/
diff passed. All 427 baseline Vault Markdown hashes unchanged. Behavioral/race/
full/scanner/knowledge/traffic/acceptance **not run; delegated by user**. No
runtime/HTTP/parser/race/SLO pass inferred. No scope or validation deviation;
existing skill rules suffice. No next increment started; delivery pending.
