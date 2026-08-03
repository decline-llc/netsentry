# Task Plan: R90-64 sustained parser and formatter fuzz baseline

## Metadata

- Timestamp: 2026-08-03T02:09:05-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `33bc37d9ff71932d6e4ea49cf414f3ed0008415a`

## Goal

Record one reproducible, path-redacted local synthetic ASan baseline that runs
both the C frame-parser and UDS JSON-formatter harnesses at an explicit
sustained iteration budget, with machine-readable evidence validation and no
release or production-traffic claim.

## Scope

- Extend `make fuzz-sustained` to build and execute both current ASan harnesses
  serially and uncached at the same explicit iteration budget.
- Preserve optional local corpus replay for the byte-oriented parser only and
  keep corpus paths redacted from generated evidence by default.
- Record separate parser, formatter, and optional corpus statuses and log tails
  in a versioned evidence format with an honest `local_synthetic` class.
- Add direct evidence-contract regressions for successful dual-harness data,
  rejection of missing/failed harness results, and path redaction.
- Run the accepted one-million-iteration baseline without an external corpus,
  then copy only reviewed non-sensitive summary facts into repository evidence.
- Update public fuzz guidance, changelog, roadmap, and task state.

## Non-Goals

- Do not change either fuzz harness, formatter/parser behavior, the UDS schema,
  capture runtime, or sanitizer flags.
- Do not access a private or external corpus, commit raw generated logs, expose
  an operator path, or characterize the run as public, external, realistic, or
  production-derived traffic evidence.
- Do not modify release-gate acceptance, refresh release artifacts, create a
  tag, publish a release/image, or start R90-59.
- Do not add third-party dependencies or broaden the run into performance or
  throughput benchmarking.

## Risks

- Reusing an existing binary could silently omit the formatter harness; build
  both ASan targets immediately before the sustained run.
- A single aggregate status can hide which harness ran or failed; record each
  harness independently and validate both names, iteration counts, and statuses.
- Generated logs or corpus inventory can reveal sensitive absolute paths; redact
  the resolved corpus root before evidence is written and use no corpus for the
  accepted baseline.
- A successful deterministic mutation run is local synthetic robustness
  evidence only; the public record must prohibit release, production-traffic,
  corpus-provenance, and throughput interpretations.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Both C harnesses run at the recorded sustained budget | Separate parser and formatter status, iteration, elapsed-time, and bounded log-tail fields plus direct output checks |
| Runs are fresh and ASan-capable | Both binaries are rebuilt from versioned ASan targets immediately before serial execution; generated timestamps and durations identify this run |
| Evidence is complete and unambiguous | A versioned Python validator requires the exact harness set, positive equal iteration counts, successful statuses, zero observed sanitizer findings, and coherent corpus fields |
| Optional corpus metadata cannot leak paths | Direct regression passes a path-bearing fixture through the evidence writer and proves the public JSON/Markdown outputs contain only a redacted marker |
| The accepted result remains honestly scoped | Reviewed repository evidence records `local_synthetic`, no external corpus, no production derivation, and no release/publication authority |
| Existing behavior remains compatible | Focused evidence tests, one-million-iteration dual-harness run, C ASan tests, full native tests, shell/Python/docs/evidence/knowledge gates pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `33bc37d9ff71932d6e4ea49cf414f3ed0008415a`.
- Verified the R90-63 feature and closure task state plus exact local Vault
  notes, full index, MOC links, and stable formatter/testing knowledge.
- Reviewed the Jul 14–Aug 2 delivery chain by phase; no missing recent delivery
  record, stale authority, or unresolved validation deviation changes priority.
- Parsed all 78 task-state JSON files and verified all 67 roadmap rows match
  exactly one Definition.
- Audited both unfinished increments: R90-59 remains blocked on exact
  publication authority; R90-64 is dependency-ready and fully defined.

## Validation

- Preflight the repository-required compiler, Make, Bash, and Python tools and
  capture non-sensitive version facts.
- Run focused evidence-contract tests and a small dual-harness integration run.
- Run `FUZZ_SUSTAINED_ITERATIONS=1000000 make fuzz-sustained` with no corpus and
  validate the generated JSON evidence independently.
- Run `make asan-test`, `make test`, `make shell-check`, `make python-check`,
  `make docs-check`, `make evidence-check`, and `make knowledge-check`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and a scoped
  sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Validated Evidence

- Repository-required Make, GCC, Bash, Python, and Go tools were present;
  non-sensitive versions are recorded in the reviewed evidence summary.
- Six direct Python regressions passed for complete dual-harness evidence,
  failed/missing harness rejection, equal budgets, exact log self-reporting,
  sanitizer marker rejection, and corpus path redaction.
- A focused 100-iteration integration forcibly rebuilt both ASan targets,
  replayed one parser corpus fixture from a directory whose name required path
  handling, passed both harnesses, and exposed no absolute path in its JSON or
  Markdown evidence.
- The accepted no-corpus run forcibly rebuilt both targets and passed
  1,000,000 parser plus 1,000,000 formatter mutations. Each harness reported
  its exact budget, exited zero, and recorded zero sanitizer findings; total
  elapsed time was 118.317 seconds.
- Serial C ASan tests passed. After the validation-order correction below, the
  ordinary C tests and every Go package under the race detector with
  `-count=1` passed from a clean C rebuild.
- Shell, Python, documentation, 22 evidence, and 33 knowledge tests passed.

## Execution Deviations

The first full validation chain ran `make asan-test` immediately before
`make test`. Because the C targets share output names and Make does not track
compiler flags, the ordinary C step reused sanitizer-mode binaries. Those C
results were not counted as ordinary-build evidence. The capture outputs were
explicitly cleaned, and the complete ordinary C plus uncached Go race suite
was rerun successfully. No source change was required, and all independent
repository-gate results remained valid.

Pre-push commit inspection also found that replacing the sustained shell file
had dropped its executable bit. No push had occurred. The mode was restored,
shell syntax, six focused evidence tests, a direct executable 10-iteration
dual-harness run, and the knowledge gate passed, then the still-local feature
commit was amended. The verified commit tree retains mode `100755`.

## Delivery Results

- Feature commit:
  `73ab39ef88245b01b3d3418f0d9aeb0f6db1d546` (`test: record
  dual-harness sustained fuzz baseline`).
- The thirteen-path commit was pushed without force, fetched, and verified as
  both `HEAD` and `origin/main`; the post-fetch 33-test knowledge gate passed.
- Exact range
  `33bc37d9ff71932d6e4ea49cf414f3ed0008415a..73ab39ef88245b01b3d3418f0d9aeb0f6db1d546`
  was synchronized twice with identical hashes to the single existing local
  Vault.
- The generated iteration note, full commit index, MOC link, and manually
  reconciled stable ASan-fuzz and testing-gate notes are verified.
- R90-65 is ready but was not started. R90-59 publication, tags, releases,
  images, registries, and workflow dispatch remain outside authority.

## Authority Boundaries

The rolling trigger authorizes the bounded R90-64 evidence tooling, local
synthetic sustained execution, sanitized repository summary, supporting
documentation, commit/push, and local Vault workflow. It does not authorize
private corpus access, production-evidence claims, release acceptance changes,
tags, releases, images, registry publication, or workflow dispatch.

## Stop Conditions

Stop on any crash, sanitizer finding, ambiguous or unequal iteration count,
failed evidence validation, sensitive path exposure, need for private corpus
access, or any attempt to use the result as tag/publication or
production-traffic authority.
