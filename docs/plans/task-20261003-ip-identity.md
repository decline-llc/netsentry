# R90-170: match equivalent exact IP blacklist addresses

## Selection and authority

Clean fresh-fetched main baseline: `6bcbdcee40dd0a8b2c9077be7361e7127a64a663`.
R90-169 feature/closure Git, exact range notes/index/MOC and fourteen stable
current references verified. Sep 12–Oct 3 phase audit covers SLO adapters and
inventory admission, patched release/toolchain and bounded core correctness.
All 119 phase commits have iteration note/index coverage. No new qualifying
R90-75 outcome or missing delivery. 173 unique prior roadmap row/Definition
pairs; Oct 3–Dec 31 horizon and independent asynchronous R90-75 full contract
retained. Empty ready queue restored with one source-grounded repair.

Existing compileIPRule accepts net.ParseIP literals but stores their original
text; ipValueMatches compares raw text before parsed CIDR matching. An accepted
IPv4-mapped literal can therefore miss the native canonical IPv4 packet address.
Equivalent IPv6 spellings have the same public Engine mismatch. This is source
inference, not an executed behavioral failure. Go's authoritative
[ParseIP](https://pkg.go.dev/net#ParseIP) and
[IP.String](https://pkg.go.dev/net#IP.String) documentation and repository-pinned
Go 1.26.8 net/ip.go establish parsed identity and canonical string forms.

## Scope, risk, non-goals and stop condition

Six paths: engine/internal/rule/engine.go; new
engine/internal/rule/ip_identity_test.go; docs/architecture.md; this plan;
docs/tasks/task-state-20261003-ip-identity.json; rolling roadmap.
Canonicalize exact compiled keys using the already validated parsed IP; retain
raw canonical lookup and add canonical lookup after the existing packet parse.
Low risk: equivalent spellings now trigger the existing exact-address rule.
Preserve original packet addresses, matched-keyword reason, caller rule config,
CIDR semantics, filters, ordering, early exit and reload publication. No new
address formats, decoder/capture IPv6 support, storage contract, dependencies,
API shape, whitespace packet normalization, performance or SLO claim. Stop on
competing edits, ambiguous compile/static/Git/Vault evidence or new authority.

Behavioral/race/CLI/full-suite/scanner/knowledge/traffic/SLO execution **not run;
delegated by user**. Binaries compile only, unexecuted. No next increment starts.

## Acceptance mapped to evidence

1. Exact two-function source diff: parsed rule keys and parsed packet fallback;
   raw canonical fast path and CIDR traversal retained. Static source/format review.
2. Public Reload/Match cases: mapped dotted/hex/uppercase/expanded spellings
   against native IPv4; reverse mapped packets; compressed/expanded/case IPv6;
   duplicate/mixed blanks and unequal addresses. Assert alert content,
   packet/config/snapshot preservation and one alert per rule. Public Engine
   IPv6 cases imply no native capture/receiver/storage IPv6 support.
3. Per-rule source/dest/any/protocol/disabled filters, exact-before-CIDR reason,
   priority/critical early exit and invalid IP/CIDR/empty reload snapshot
   preservation; file LoadFromFile/SaveToFile/Reload retains original spelling.
   Direct regression assertions authored and compiled; execution delegated.
4. Preflight exact Go 1.26.8 owning engine packages; fail-fast compile-only
   rule/API/cmd/netsentry/pipeline chain. Static docs/JSON/roadmap multisets,
   prior Definitions and R90-75/split/horizon/history/links/fences/scope/diff/
   sensitive review; unchanged pre-delivery Vault hashes.
5. Focused conventional feature commit and one docs-only closure if required.
   Non-force push/immediate fetch exact refs; full-SHA Vault note/index/MOC,
   current stable prose reconciliation with all prior substantive prose and
   immutable iteration notes preserved; identical-range replay hashes stable.
   Refresh future queue without starting next increment. No duplicate closure.

## Checkpoints

Plan/state persisted before runtime/architecture/queue edits. Compile/static
success will not be promoted into runtime/race/SLO evidence. Existing generic
skill guidance covers this repair; no redundant workflow rule planned.

Initial history audit incorrectly required every historical note in the bounded
generated MOC list. Reviewed the versioned twenty-entry policy, corrected the
audit to historical note/index coverage plus latest exact delivery MOC coverage,
and reran it successfully before plan selection. No repository/Vault mutation
or missing delivery resulted.


## R90-170 Compile and Static Checkpoint (2026-10-03)

Runtime diff exactly matches two-function plan: store validated parsed IP.String
keys and look up parsed packet identity after the retained canonical raw lookup,
before unchanged CIDR traversal. Original rule config and packet/alert text
retained; no native IPv6 support expansion. Four direct public regression
functions authored: sixteen address spelling/equality/malformed packet cases,
nine direction/protocol/disabled/CIDR-precedence cases, per-rule priority and
critical early-exit checks, three invalid reload preservation cases and valid
retry, plus file SaveToFile/LoadFromFile/Reload spelling preservation. Complete
alert values, one-per-rule results and caller/snapshot preservation asserted.
No private seams, mocks, sleeps, skips or hidden runtime execution.

Preflighted exact Go 1.26.8 and owning packages; complete fail-fast rule/API/
cmd/netsentry/pipeline compile-only chain passed. Binaries unexecuted. Static
exact transform/format/docs/188 JSON/174 unique pairs/173 prior Definitions/
R90-75/testing split/history/horizon/links/fences/six paths/diff/sensitive review
passed. All 419 baseline Vault Markdown hashes unchanged. Behavioral/race/CLI/
full-suite/scanner/knowledge/traffic/acceptance **not run; delegated by user**.
Authored regression debt remains departmental; no runtime/race/SLO pass or
performance result inferred. No unresolved validation ambiguity. Existing skill
rules suffice; no separate skill update or next increment.


## R90-170 Completion and Forward Queue Refresh (2026-10-03)

Feature `c7e6a56b67f724ebc7a36252df9f0c790780c9a4` contains exactly the six planned paths. Non-force push
and immediate fresh fetch verified clean main HEAD/origin/main/FETCH_HEAD
equality. Exact full-SHA range `6bcbdcee40dd0a8b2c9077be7361e7127a64a663..c7e6a56b67f724ebc7a36252df9f0c790780c9a4` synchronized to
`04-开发迭代记录/2026-10-03-c7e6a56b67-CI知识同步.md`; Git-resolved identifiers, note/index/MOC and six-path scope verified.
Fourteen stable current notes reconciled; all prior substantive current/topic/
history prose archived exactly, excluding only the actual generated CI MOC
region resolved from versioned constants. Legacy MOC content remains exact.
All 372 baseline immutable iteration notes unchanged. Identical feature replay
preserves all 420 Markdown hashes; snapshot JSON SHA-256 `ba9170dd590dc2d658715b45ca686604e7e3351c55c28c605f5843b9843dab6e`.
Existing unique sibling local Vault supplied explicitly; no second/remote Vault.

Acceptance matches plan: two-function parsed exact-key/fallback repair, canonical
raw lookup and CIDR traversal retained; every other runtime path unchanged.
Four public authored functions reach sixteen address spelling/equality/invalid
packet cases, nine filter/disabled/exact-before-CIDR checks, per-rule priority
and critical early exit, three invalid reload preservation cases/valid retry,
and public file SaveToFile/LoadFromFile/Reload spelling preservation. Complete
alert values, packet/caller config/snapshot and file bytes checked directly;
no mock/private seam/sleep/skip. All assertions compiled but unexecuted. Engine
IPv6 cases imply no native capture/receiver/storage IPv6 support. Original alert
address/reason and published rule text preserved. No performance claim.

Preflighted exact Go 1.26.8 four-package fail-fast compile-only chain and complete
static exact source/format/docs/188 JSON/174 unique pairs/173 prior Definitions/
R90-75/testing split/history/horizon/links/fences/scope/diff/sensitive review
passed. Binaries unexecuted. Behavioral/race/CLI/full-suite/scanner/knowledge/
traffic/acceptance **not run; delegated by user**. No runtime/race/SLO pass
inferred; growing authored regression debt remains departmental. Temporary
audit expectation corrected for bounded MOC policy and entire audit rerun;
no missing delivery or unresolved ambiguity. Existing skill rules sufficient.

This single three-path docs-only record closes the same increment. Resolve its
full SHA from Git; verify non-force push/fresh fetched refs and exact
feature..closure Vault note/index/MOC before reporting. Do not repeat verified
feature/closure delivery or add a self-reference closure. Forward queue refreshed:
no other defined local ready increment. R90-75 independent asynchronous full
contract and Oct 3–Dec 31 horizon retained. Next trigger verifies completed
closure/Vault, audits fresh history/code/queue and persists a separate eligible
plan before editing. No next implementation started.
