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
