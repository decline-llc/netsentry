# R90-131: Refresh main's supported execution toolchain

## Selection and authority

Clean fetched `HEAD == origin/main == FETCH_HEAD` is
`1ec7c9555e7e4788618f88efb9050fb536c72ad1`. The latest R90-59 closure's exact
range `1808fcdda909a1432721b54bea2762a8f75c409a..1ec7c9555e7e4788618f88efb9050fb536c72ad1`
has a verified Vault iteration note, full-index row and MOC link. The sole
existing sibling Vault contains 341 Markdown notes. Review covers 41 commits in the Sep 17–Oct 1
delivery phases: SLO tooling implementation with delegated execution evidence;
collector input hardening and queue audits; separately validated R90-59
candidate publication and closure. No main-only execution result follows from
the historical release candidate. The forward queue selects ready R90-131;
R90-75 remains departmental acceptance with its independent evidence contract.

The Sep 25 instruction delegates behavioral and knowledge suites. The Oct 1
R90-59 testing grant was task-scoped. This trigger performs static review and
delivery checks only; no tests, scanner, RC, build or acceptance run is authorized.

## Decision and scope

Official [download metadata](https://go.dev/dl/?mode=json) reviewed Oct 1 lists
stable Go 1.27.1 and 1.26.8. The [release policy/history](https://go.dev/doc/devel/release)
supports the two latest major lines, 1.27 and 1.26; 1.25 is no longer supported.
Select the latest patch in the supported 1.26 line, Go 1.26.8, to minimize the
major-line jump and align with the independently validated candidate selection.
This decision does not establish main compatibility or absence of vulnerabilities.

The official Linux amd64 archive `go1.26.8.linux-amd64.tar.gz` is 66,897,291
bytes with metadata SHA-256
`d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b`.
Record this upstream identity; do not claim a fresh archive download/hash check.

Intended paths: `engine/go.mod`, `.github/supply-chain-lock.json`, both READMEs,
`docs/supply-chain.md`, `docs/release-readiness.md`, the rolling roadmap, this
plan and `docs/tasks/task-state-20261001-main-go-toolchain.json`. Only the
toolchain directive changes module behavior; keep the language baseline,
dependencies, runtime sources, tools, Actions, workflows and publication objects.
Update affected current stable Vault prose; preserve immutable iteration notes.

## Acceptance and evidence map

| Acceptance | Planned evidence |
| --- | --- |
| Supported exact patch selection | Live official stable-release metadata and policy; selected line, archive bytes/checksum/source recorded in lock and plan |
| Pin and consumer consistency | Static module/lock comparison, all three setup-go consumers of `engine/go.mod`, Docker's module-driven download and build; unchanged language/dependency/workflow diffs |
| Honest validation boundary | Docs/state distinguish current main from published candidate; native/RC, scanner, workflow execution and knowledge suites recorded as not run, delegated by user |
| Queue and history integrity | All task JSON parses; complete unique row/Definition multisets; ordered prior completion, selection and completion; complete unfinished-item contracts |
| Safe delivery | Docs check, diff/scope/sensitive-information review; focused commit, verified push/fetch and exact Vault range; stable prose reconciliation and identical-range replay |

## Risks, non-goals and stop conditions

Changing compilers may expose main-only compatibility or vulnerability findings;
departmental execution evidence remains outstanding. Review dates describe a
snapshot, not perpetual support. Stop on ambiguous version/checksum identity,
conflicting consumers, necessary language/dependency migration, remote or Vault
identity ambiguity, or new authority. Runtime changes, test execution, CI dispatch,
tag/Release/image mutation, private data and R90-75 acceptance are outside scope.
Complete exactly this increment; refresh the future queue without starting it.

## Static review checkpoint

Live metadata matches the selected lock archive filename, size and checksum and
latest-supported-patch snapshot. The module diff changes only the toolchain; all
lock entries outside the Go review retain their prior values. All three setup-go
steps read `engine/go.mod`; Docker copies the module before `go mod download`
and subsequently builds from it. Native execution and Docker resolution remain
delegated. Docs check, 149 task-state JSON parses, 135 unique matching roadmap
row/Definition pairs, local link/fence, scope and diff review pass. No behavioral
or vulnerability result is claimed. Delivery is the remaining work.

## Delivery and deviations

Feature `7505be8457e99a89965a276694e9e22e9eae0913` contains exactly the nine
planned paths. Push and fresh fetch verified clean HEAD/origin/main/FETCH_HEAD
at that SHA; the publication tag object and peeled candidate were unchanged.
Exact range `1ec7c9555e7e4788618f88efb9050fb536c72ad1..7505be8457e99a89965a276694e9e22e9eae0913`
was synchronized and its iteration note, index row and MOC link verified.

Eleven stale current handoffs and one additional Actions/Docker stable note
were reconciled. That extra local note surfaced during consumer review and
expanded only stable knowledge coverage; repository scope stayed at nine paths.
All 294 existing iteration-directory notes remained unchanged. Replay
preserved the 342-file Markdown snapshot JSON hash
`6a8b18f17ce5195153ed2412eeb00ef7c553114fa1381c7f15c67d9b50fdcfef`.
The initial safety review stopped before staging because it scanned old pathname
examples; rerunning the complete review on intended additions passed. No source
or behavioral validation deviation occurred, and no tests were run. All
execution/knowledge checks remain explicitly delegated.

Every acceptance criterion has the promised static or delivery evidence; no
direct behavioral regression was promised or claimed. This docs-only closure
records the delivered feature as complete; its own final SHA and exact range
must be resolved from Git, push/fetch-verified and synchronized before reporting
final delivery. There is no additional local ready increment. R90-75 retains
its independent departmental contract; re-audit fresh evidence on the next
trigger without repeating this increment or release publication.
