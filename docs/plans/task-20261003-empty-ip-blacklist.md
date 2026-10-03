# R90-162: reject IP blacklists with no compiled addresses

## Selection and baseline audit

Fresh fetched clean main HEAD/origin/main/FETCH_HEAD:
`ca989a4cf363c9b0fe7e42145ceb4d306ddadb53`. R90-161 seven-path feature/three-path
closure exact Git/Vault/note/index/MOC verified; abbreviated metadata uniquely
resolves through Git. 103-commit Sep 5–Oct 3 phase moved from SLO tooling to core
correctness; no new qualifying R90-75 evidence. Sole unfinished R90-75 independent
asynchronous departmental acceptance and Oct 3–Dec 31 horizon unchanged. Empty
ready queue reconciled inside source-grounded empty compiled match-set repair.
403 Vault Markdown hashes/14 entire stable notes backed up; no AGENTS. Owning
engine Go 1.26.8 preflighted. Plan/state persisted before runtime/other docs edits.

## Scope, risk and authority

Six paths: `engine/internal/rule/engine.go`, new
`engine/internal/rule/empty_ip_blacklist_test.go`, `docs/architecture.md`, this
plan, `docs/tasks/task-state-20261003-empty-ip-blacklist.json`, rolling roadmap.
Validation currently checks nonzero raw IP slice length, while compileIPRule skips
trimmed blank entries. A blank-only list therefore passes the existing requirement
for at least one IP/CIDR and publishes a rule that can never match. Add only a
three-line compiled-address emptiness guard after the existing loop, returning
the existing at-least-one-address diagnostic through rule-ID wrapping. Preserve
mixed blank/valid entries, trimming, duplicates, exact/CIDR scoping, direction,
protocol and validation ordering; disabled rules retain existing validation.
Risk low: previously accepted blank-only lists now reject. No IPv6 expansion,
normalization, schema, parser, persistence or API policy change. LoadFromFile and
SaveToFile remain parsing/serialization APIs; rejection occurs on actual Reload.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Static/compile evidence is no runtime/SLO pass.

## Acceptance mapped to evidence

1. Exact three-line post-loop guard rejects no compiled exact IPs/CIDRs; every
   other tracked engine source byte-identical. Prior invalid direction/protocol/
   IP/CIDR diagnostics and validation order retained.
2. Public Reload regression: nil/empty IP slices, single empty/ASCII/Unicode blank,
   multiple blanks, enabled and disabled; exact rule-ID diagnostic, retained old
   Rules/RuleCount/Match, unchanged candidate Config/slice/packet. Restore valid
   reload afterward. No private state or swallowed panics.
3. Public successful matching controls for padded exact IP, CIDR, mixed blanks,
   duplicates, disabled valid rule, source/destination/any and TCP/UDP gates,
   owning rule scoping and negative packets; input ownership retained.
4. Public LoadFromFile then actual Engine.Reload over canonical/legacy schema in
   wrapped/array forms: blank-only rejects with old snapshot retained; valid mixed
   entries compile/match; entire file bytes retained after load/reload. Parser and
   serializer gain no standalone validation guarantee.
5. Pinned owning Go 1.26.8 rule/API/pipeline compile-only chain; static direct
   boundaries/source/format/docs/180 JSON/166 unique complete roadmap pairs/all
   prior Definitions/R90-75/history/horizon/links/fences/six paths/diff/sensitive.
   No binaries or benchmarks executed.
6. Feature plus one docs-only closure exact full-SHA push/fetch/Vault scope/note/
   index/MOC; 14 current stable notes reconciled, entire prior prose archived,
   topic tails/immutable notes preserved; identical replay; refresh queue and stop.

## Non-goals and stop conditions

No individual blank-element rejection when valid addresses remain, address-text
canonicalization, expanded IPv6 authority, priority/suppression/payload/port/rule
schema/parser/serializer/API runtime/dependencies/toolchain changes or measured
performance/race/SLO guarantee. Stop for competing edits, ambiguous validation/
Git/Vault or new product/private/external authority. Existing skills cover direct
rejection boundaries and preservation; no redundant skill change. Next selection
only on another trigger.


## Implementation and validation checkpoint

Exact runtime diff is three-line post-loop compiled IP/CIDR emptiness guard;
all other tracked engine/runtime/loader/API source byte-identical. Existing
rule-ID/at-least-one-address diagnostic reused, including disabled validation.
Mixed blank/valid lists still skip blanks; trimming, duplicates, ownership,
per-rule scoping, direction/protocol filters and prior diagnostics retain behavior.
LoadFromFile and SaveToFile remain parser/serializer APIs; Reload validates.

Three external public functions authored/compiled only: twelve enabled/disabled
nil/empty/ASCII/Unicode/multiple blank rejection cases with old Rules/RuleCount/
Match and candidate/packet preservation plus valid retry; four prior direction/
protocol/IP/CIDR diagnostic controls; twelve exact/CIDR/mixed/duplicate/direction/
protocol/disabled/outside matching controls and two owning-rule scope checks;
eight canonical/legacy wrapped/array load-to-actual-Reload cases with valid mixed
matching/blank rejection/published state and entire file-byte preservation.
No private-state manipulation, sleeps, skips or panic swallowing. No runtime pass.

Pinned Go 1.26.8 rule/API/pipeline complete compile-only chain passed;
binaries/benchmarks unexecuted. Static exact source/direct boundaries/Go-format/
docs/180 JSON/166 complete unique roadmap pairs/prior Definitions/R90-75/history/
horizon/links/fences/six paths/diff/sensitive passed. All execution **not run;
delegated by user**. 403 baseline Vault hashes unchanged; no unresolved validation
failure or ambiguity. Existing skills cover rejection boundaries/preservation;
no redundant skill update. Feature and one docs-only closure remain; next increment
not started. No normalization/IPv6 expansion/publication/SLO evidence claim.
