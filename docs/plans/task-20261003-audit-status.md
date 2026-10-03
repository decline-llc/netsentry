# R90-153: record committed HTTP audit response status

## Selection and baseline audit

Clean fetched main HEAD/origin/main/FETCH_HEAD: `a18e66babf7b17d012e874574f77fa00861d9375`.
R90-152 exact six-path feature/three-path closure Git/Vault note/scope/index/MOC
verified. Sep 5–Oct 3 phase covers 85 commits: supplied SLO evidence tooling
followed by core correctness repairs. No new R90-75 acceptance artifact; its
full independent asynchronous departmental contract and Oct 3–Dec 31 horizon
remain unchanged. Restore empty local ready queue inside this source-grounded
correctness repair, without a separate audit-only delivery. Vault 385-file hash
snapshot and complete 14 current stable note contents backed up outside Git.
Pinned owning engine module resolves Go 1.26.8 linux/amd64.

## Scope, risk and authority

Six paths: `engine/internal/api/audit.go`, new
`engine/internal/api/audit_status_test.go`, `docs/api-reference.md`, this plan,
`docs/tasks/task-state-20261003-audit-status.json` and rolling roadmap.
Persist plan/state before source/docs changes. Runtime only auditResponseWriter
WriteHeader: ignore headers after committed final status; forward accepted
headers before recording status; record >=200 or 101 as final, leaving other
1xx open. Preserve implicit Write status 200 and no-write audit default 200.
Pinned local net/http server source directly confirms first-final-only behavior,
forwarding non-101 1xx without final commitment, and 101 terminal handling.
Current wrapper overwrites status on repeated final headers and treats 1xx as
final: logs can disagree with wire response and status-derived authorization.
Risk medium: response-status bookkeeping affects audit evidence, not actual
endpoint auth policy. No observed production incident claim.
All behavioral/race/CLI/full-suite/scanner/knowledge/traffic/acceptance execution
**not run; delegated by user**. Static review/docs checks and compile only;
never execute produced test binaries. No live HTTP or audit correctness pass
from compilation.

## Acceptance mapped to evidence

1. Exact WriteHeader-only source transform: first committed final status retained,
   non-101 1xx forwarded without commitment, terminal 101 recorded, underlying
   invalid-code validation occurs before status recording. Primary audit fields,
   request IDs, GET skip, storage phase/authorization, Write and all other
   tracked engine source unchanged. Source plus pinned net/http mapping evidence.
2. Author real net/http httptest.Server mutation cases through audit middleware:
   repeated explicit finals including unauthorized-first/success-first, implicit
   body success followed by explicit final, no-write default, repeated 1xx then
   explicit final/body or implicit-body/handler-return success. Compare actual
   client status/body and observed audit status/authorized/target/request ID.
   Observe 1xx with httptrace, synchronize on handler completion before reading
   audit entries, bounded request context/client, no fixed sleeps or fake writer
   replacing wire boundary. Include GET-skip and absent-logger compatibility.
3. Direct wrapper/header forwarding for terminal 101, repeated finals and
   underlying invalid-code panic propagation without poisoning status; use real
   ResponseRecorder for final statuses. A spy only verifies header forwarding,
   not substitutes actual net/http 1xx evidence. Direct cases authored/unexecuted.
4. Pinned Go 1.26.8 API/alert compile-only complete chain; Go-format/parse/docs,
   task JSON, unique complete roadmap row/Definition multisets, all prior
   Definitions/R90-75/history/horizon/links/fences/six-path/diff/sensitive review.
5. Commit feature and one docs-only closure; push/fresh-fetch full exact SHA refs,
   exact Vault note/index/MOC, affected stable current prose reconciled with
   prior substantive prose and immutable hashes retained; identical replay;
   refresh queue and stop without next implementation.

## Non-goals and stop conditions

No router/auth/schema/log-field/response-body/request-ID/GET/storage-phase changes,
optional ResponseWriter interfaces, async audit policy, panic recovery in runtime,
other core modules, dependencies/toolchain, tests execution, IPv6 or publication.
Stop for ambiguous static/compile/Git/Vault, competing work, unsupported HTTP
semantics, new product/private/external authority, or a second increment.
Existing skill guidance suffices; no skill edit just to narrate this outcome.


## Implementation and validation checkpoint

Runtime changes only auditResponseWriter.WriteHeader: after an existing final
commit, return before forwarding; otherwise forward before recording status,
then record >=200 or terminal 101. Other 1xx leave status zero. Implicit Write
200/no-write default 200 and all other tracked engine/audit fields/request IDs/
GET skip/storage phase/status-derived authorization/router/auth source unchanged.
Pinned Go net/http source directly confirms the same state transitions and
client trace behavior. No new endpoint authorization semantics or panic recovery.

Three direct regression functions authored: 12 actual net/http server/client
cases through wrapped audit middleware, comparing wire status/body/request ID,
trace-observed 1xx and completed audit status/authorization/target/path/method;
explicit/repeated finals, unauthorized-first/success-first/failure/no-content,
implicit body, no-write, informational explicit/body/handler-return, GET skip,
absent logger. Request/client deadlines and handler-completion channel avoid
fixed sleeps and log timing races. Five direct forwarding cases include terminal
101 and ignored later headers; two real ResponseRecorder invalid-code cases
assert underlying panic propagation with zero cached status, then final 401
retained through attempted 200. Recovery exists only in this deliberate panic
assertion. Spy does not replace real informational wire evidence.

Pinned Go 1.26.8 final API/alert complete compile-only chain passed; external
binaries unexecuted. Exact source/local standard mapping/direct-boundary/Go-format/
docs/171 JSON/157 unique complete roadmap/prior Definitions/R90-75/history/horizon/
links/fences/six-path/diff/sensitive review passed. Behavioral/race/CLI/full-suite/
scanner/knowledge/traffic/acceptance execution **not run; delegated by user**.
No live HTTP/audit/runtime/SLO success inferred. No implementation, compilation
or validation failure; baseline Vault hashes unchanged. Existing skills suffice;
no generic skill edit. Feature plus one docs-only closure remain, without next
implementation or release action.


## Delivery and queue closeout

Feature `6f443423ea3d5eb83ec4319acb84774eaba65241` contains exactly six planned paths. Isolated
fix/r90-153-audit-status fast-forwarded freshly verified main; push/fresh-fetch
verified clean HEAD/origin/main/FETCH_HEAD at that SHA. Exact range
`a18e66babf7b17d012e874574f77fa00861d9375..6f443423ea3d5eb83ec4319acb84774eaba65241` six-path scope, iteration note
`04-开发迭代记录/2026-10-03-6f443423ea-CI知识同步.md`, full index and MOC verified. Fourteen stable current notes
reconciled; entire previous substantive current prose archived under explicit
R90-152 historical headings. Original topic tails and all 338 baseline immutable
iteration hashes retained; only the documented generated MOC region refreshes.
Identical exact-range replay preserves 386 Markdown files; snapshot JSON SHA-256
`c3d9747620fd54fe0d4dcff79c39630d3cba4499c9f31ed90430360af698b086`. Unique existing sibling local Vault
selected explicitly; no remote Vault or new empty Vault created.

Acceptance matches persisted plan: exact WriteHeader-only guard/forward/commit
transform retains first final status, non-101 informational headers do not
commit status, terminal 101 does, underlying invalid-code validation precedes
cache mutation. Implicit Write/default audit 200 and all other tracked engine/
audit fields/request IDs/GET skip/storage phase/status-derived indicator/router/
endpoint authorization policy remain unchanged. Pinned net/http server/client
source establishes the same final/informational and trace semantics.

Three direct regression functions reach promised boundaries: 12 actual net/http
server/client wrapped-audit cases compare wire status/body/request ID, trace-
observed 1xx and completed audit fields, with client/context deadlines and an
observable completion channel; five forwarding cases include terminal 101;
two real ResponseRecorder invalid-code cases explicitly assert deliberate
underlying panic/zero cached status then final 401 retained through 200. The
spy only observes forwarding, not substitutes real informational-wire evidence.
All are authored/compiled, unexecuted. Final Go 1.26.8 API/alert complete
compile-only chain and source/local-standard/direct-boundary/Go-format/docs/
171 JSON/157 unique complete roadmap/prior Definitions/R90-75/history/horizon/
links/fences/six-path/diff/sensitive review pass. Behavioral/race/CLI/full-suite/
scanner/knowledge/traffic/acceptance execution **not run; delegated by user**.
No live HTTP/audit/runtime/race/SLO pass, observed incident, source/compilation/
validation/delivery failure or Vault topic loss claimed. Existing skills
sufficed; no generic outcome-only edit. The sole planning deviation is empty
ready queue restored inside this source-grounded core repair.

This single three-path docs-only record closes the same increment: resolve its
full SHA from Git, push/fresh-fetch and verify exact feature-tip..closure-tip
Vault note/index/MOC/stable prose before reporting. No self-reference follow-up
closure. No currently defined dependency-ready local item after queue refresh.
R90-75 full independent asynchronous departmental contract and Oct 3–Dec 31
horizon remain unchanged. Next trigger verifies fetched closure/Vault, audits
fresh code/queue and persists a separate eligible plan before edits. No next
increment started; do not repeat R90-152/R90-153 delivery or R90-59 publication.
IPv6/external publication needs separate authority.
