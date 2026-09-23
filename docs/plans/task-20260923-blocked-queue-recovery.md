# Task Plan: R90-112 blocked-queue recovery reconciliation

## Scope and authority

Refresh the rolling 90-day horizon to Sep 23-Dec 21, 2026 and correct stale
R90-59 recovery instructions using delivered R90-105 evidence. Work is limited
to this plan, its task state, the roadmap, and the active R90-59 task state,
plus the affected local stable Vault prose. Preserve completed historical
records. No runtime, tests, toolchain, candidate, tag, workflow, publication,
benchmark evidence, or product/SLO decision is part of this increment.

## Baseline and audit

- Clean fetched `HEAD == origin/main == FETCH_HEAD`:
  `5a761756de3a981a3047373d8bc8da9a3a441f06`.
- R90-111 feature `bded4d8b4dc12eee7e6f4734a7956374a106dc4f` and
  direct closure are delivered; verify both exact Vault ranges before editing.
- Aug 26-Sep 23 history contains eight documentation commits and no runtime
  change. No commits follow Sep 1. The delivery trend is repeated blocked-queue
  audits; this increment is justified by stale active recovery instructions
  and the horizon, not by a need to repeat completed audits.
- All 127 prior task states parse; 115 roadmap rows and Definitions match as
  complete multisets without duplicates. The initial 33-test knowledge gate
  passes. R90-59 and R90-75 are the only unfinished rows and retain dependencies,
  windows, risks, acceptance, required validation and stop conditions.
- R90-59 still lists R90-105 as unfinished, despite feature
  `c50c184e7797440139b644ac7407ff238075d733` and its completed delivery state.
  Current module selects Go 1.25.14 with language baseline go 1.22.2; this is
  historical delivery evidence, not fresh security validation of a candidate.
- Direct remote tag lookup returns no v0.1.1 tag; GitHub Release lookup returns
  HTTP 404. Preserve the recorded exact-candidate security failure and separate
  authority needed to replace and resign the local tag.

## Acceptance and evidence

| Criterion | Direct evidence |
| --- | --- |
| Current 90-day queue | Sep 23-Dec 21 inclusive date calculation; R90-59/R90-75 complete blocker contracts; one R90-112 row/Definition |
| Accurate R90-59 recovery | R90-105 commit and completed state; completed dependency removed from todo; resume starts with remaining candidate authority and validation |
| Historical authority preserved | Existing candidate/tag identities and failed validation unchanged; Aug 7 grant identified as superseded by Aug 23 exact-object grant |
| Recoverable delivery | Exact R90-111 parent chain and Vault note/index/MOC; fetched baseline; new exact-range sync and stable prose reconciliation |
| Bounded change | Four intended repository paths; no source/test/artifact/publication mutations |

## Validation and stop conditions

Parse all task states; require equal row/Definition multisets without duplicates;
check ordered R90-111 completion and R90-112 selection/completion markers;
verify horizon arithmetic and exact scope. Run docs-check, knowledge-check,
git diff --check and anchored credential/path review before each commit.
Push main without tags or force, fetch-verify, run the knowledge gate and sync
and replay the exact full-SHA range. Verify note, index, MOC and stable prose.
Use one docs-only closure commit for verified delivery facts.

Stop for ambiguous validation or remote/Vault evidence, private/external data,
new candidate/tag authority, a product decision, or work beyond this increment.
R90-59 and R90-75 remain blocked; no later increment is started.

## Local acceptance checkpoint

Both R90-111 exact parent ranges and abbreviated range markers in their
iteration notes match Git; index and MOC links are present. The initial helper
incorrectly expected a full SHA inside the generated note; inspection showed
its documented abbreviated range format. The corrected check validates both
full Git identities and those note markers. Prior closure-range replay passed.
R90-105 state is complete and its feature contains the expected toolchain pin.
The local tag object and candidate match the recorded immutable identities.

All 128 states parse; 116 unique roadmap rows and Definitions match as complete
multisets. The inclusive horizon is exactly 90 days and R90-112 selection
follows R90-111 completion. Documentation, all 33 knowledge tests and diff
formatting pass. Acceptance review confirms only the four planned paths change;
R90-59 keeps its failed historical validation and R90-75 retains both blockers.
No direct runtime regression is promised by this documentation-only plan.
No skill change is needed: existing stale-authority reconciliation rules cover
this repair. Feature delivery and exact-range Vault verification remain.

## Delivery results

Feature `139504de6bc74148244b681956dbc5b50b125cd5` contains exactly the
four planned documentation paths. Main was pushed without force or tags;
fresh fetch proved clean `HEAD == origin/main == FETCH_HEAD` and the post-fetch
33-test knowledge gate passed. Exact range
`5a761756de3a981a3047373d8bc8da9a3a441f06..139504de6bc74148244b681956dbc5b50b125cd5`
was synchronized to the sole local Vault. Iteration note, full index and MOC
links are verified. Stable MOC and release guidance now explain the refreshed
horizon, completed dependency, superseded grant and current recovery boundary.
Identical-range replay preserves Markdown content hash
`2a3afb05293071e13bed34fede7125bcffb1190573b2bd8e228eae6b78efb0c2`.

R90-112 is complete. This docs-only closure records verified feature delivery;
it must receive its own push/fetch, knowledge gate and exact-range Vault sync.
R90-59 and R90-75 remain blocked with complete contracts. There is no next ready
increment; future selection requires material evidence or authority change.
