# Task Plan: R90-63 C UDS JSON formatter fuzz boundary

## Metadata

- Timestamp: 2026-08-02T01:40:42-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `f5dc37e48513de31633aaa7a812e619a3d171e90`

## Goal

Add a deterministic, ASan-capable fuzz boundary around the handwritten C
packet, heartbeat, and hello JSON formatters, proving valid JSONL output and
safe success/failure behavior at escaping, payload, integer, and output-buffer
boundaries without changing the UDS wire contract.

## Scope

- Add one C fuzz harness with `LLVMFuzzerTestOneInput` plus deterministic seed
  and mutation execution for all three formatter entrypoints.
- Derive only bounded, NUL-terminated C strings from arbitrary input while
  exercising JSON control escapes, quotes, backslashes, and output expansion.
- Cover empty, ordinary, and maximum payloads; signed and unsigned integer
  extremes; and finite floating-point metrics.
- For every successful format, require exact-fit reproduction, reject a buffer
  one byte too small, and guard both sides of the destination against writes.
- Emit representative packet, heartbeat, and hello JSONL for a dependency-free
  Python decoder that verifies frame-kind, key, type, and base64 invariants.
- Expose the deterministic ASan smoke through the capture and root Makefiles,
  then update public development guidance, changelog, roadmap, and task state.

## Non-Goals

- Do not change formatter output, the UDS schema, receiver behavior, capture
  parsing, session rules, payload limits, or public APIs.
- Do not add a C JSON parser, third-party C runtime dependency, random seed,
  private/external corpus, sustained evidence run, or release claim.
- Do not modify the existing sustained parser workflow; R90-64 owns combined
  sustained parser/formatter evidence after this harness is delivered.
- Do not start R90-64 or the authority-blocked R90-59 publication work.

## Risks

- Arbitrary bytes can create unterminated or invalid-text C inputs and make a
  crash-only result meaningless; all fuzzer-derived strings must use bounded
  valid JSON text alphabets and explicit terminators.
- `%f` can render non-finite values as invalid JSON; deterministic structured
  input must retain integer-derived finite metrics while reaching numeric
  boundaries on the integer fields.
- A formatter may return failure after `snprintf` has written a truncated
  prefix; the boundary must distinguish this safe failure from successful
  exposure and prove surrounding canaries remain intact.
- A string-only assertion can accept malformed JSON; emitted frames require an
  independent standard-library JSON decode and schema-shape check.

## Acceptance Criteria and Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| All formatters have one dedicated fuzz boundary | `LLVMFuzzerTestOneInput` deterministically selects and exercises packet, heartbeat, or hello structured inputs |
| Escaping, payload, and numeric edges are reached | Named direct seeds cover JSON escapes, empty/maximum payloads, signed extrema, and unsigned maxima before deterministic mutation rounds |
| Success and truncation are unambiguous | Every successful seed/mutation is repeated with exact-fit and undersized buffers; return values, complete bytes, NUL termination, and canaries are asserted |
| Successful frames remain canonical JSONL | Harness-emitted packet, heartbeat, and hello lines pass a Python standard-library decoder with exact frame-key/type and base64-length invariants |
| Existing behavior remains compatible | Existing and ASan C tests, formatter fuzz smoke, full native tests, E2E smoke, shell/Python/docs, and knowledge gates pass |

## Trigger Audit

- Fetched `origin/main` and verified clean local/remote equality at
  `f5dc37e48513de31633aaa7a812e619a3d171e90`.
- Verified the R90-62 feature/closure records, exact closure Vault note, full
  index, MOC links, and stable SQLite/testing knowledge at the recorded feature
  version.
- Reconciled 135 commits since Jul 14 across the active delivery phases; no
  missing delivery record, stale authority, or unresolved validation deviation
  changes selection priority.
- Parsed all 77 task-state JSON files and verified all 67 roadmap rows match
  exactly one Definition.
- Confirmed R90-63 is the highest-priority ready increment; R90-64 depends on
  it and R90-59 still lacks explicit publication authority.

## Validation

- Build and run the new deterministic formatter ASan target, including its
  independent JSONL validation, at focused and elevated mutation counts.
- Run existing C tests and serial ASan C tests.
- Run complete `make test` and `make e2e-smoke`.
- Run `make shell-check`, `make python-check`, `make docs-check`, `make
  evidence-check`, and `make knowledge-check`.
- Parse every task-state JSON and verify exact roadmap row/Definition coverage.
- Run `git diff --check`, intended staged-diff review, and a scoped
  sensitive-information review.
- Push without force, fetch and require `HEAD == origin/main`, rerun the
  knowledge gate, then synchronize and verify the exact full-SHA Vault range.

## Validated Evidence

- The formatter target compiled with AddressSanitizer and passed its default
  5,000-mutation smoke plus an elevated 100,000-mutation run.
- Named structured cases reached packet payload lengths zero and 4096, JSON
  control/quote/backslash escaping, signed integer minima/maxima, unsigned
  maxima, and finite metric formatting across packet, heartbeat, and hello.
- Every harness invocation proved complete full-buffer output, identical
  exact-fit output, rejection at one byte short plus one/zero-byte capacities,
  NUL/length agreement, no raw line delimiter, and intact pre/post canaries.
- Three representative emitted frames passed strict Python standard-library
  JSONL decoding with duplicate/non-finite rejection, exact frame keys and
  scalar types, packet base64/length agreement, and hello payload-limit checks.
- Conventional C sender tests directly reject packet, heartbeat, and hello
  truncation; both ordinary and serial ASan C test runs passed.
- `make test` passed every C test and every Go package under the race detector.
- `make e2e-smoke` passed with 6 packets processed, 5 alerts generated, and 8
  rules loaded.
- `make shell-check`, `make python-check`, `make docs-check`, `make
  evidence-check` (16 tests), and `make knowledge-check` (33 tests) passed.

## Execution Deviations

None. The bounded structured-input, independent JSONL-decoder, and no-wire-
change design executed as planned.

## Delivery Results

- Feature commit:
  `357455a22f62b4d85c16c431fde70320d27c28a9` (`test: add UDS
  formatter fuzz boundary`).
- The fourteen-path commit was pushed without force, fetched, and verified
  equal to `origin/main`; the post-fetch 33-test knowledge gate passed.
- Exact range
  `f5dc37e48513de31633aaa7a812e619a3d171e90..357455a22f62b4d85c16c431fde70320d27c28a9`
  was synchronized twice with identical iteration-note, full-index, and MOC
  hashes to the single existing local Vault.
- The generated note, full commit index, MOC link, and manually reconciled
  stable testing-matrix and ASan-fuzz notes identify the deterministic
  structured boundary, independent decoder, and honest non-sustained scope.
- R90-64 is ready but was not started; R90-59 publication, tags, releases,
  images, and workflow dispatch remain outside authority.

## Authority Boundaries

The rolling trigger authorizes the bounded R90-63 harness, validator, build
targets, supporting documentation, commit/push, and local Vault workflow. It
does not authorize wire-format changes, external/private corpus access,
sustained evidence claims, tags, releases, images, registry publication, or
workflow dispatch.

## Stop Conditions

Stop if valid coverage requires a UDS schema change, noncanonical JSON,
unbounded/invalid C strings, a new C runtime dependency, private corpus access,
publication authority, or if any sanitizer/decoder result is ambiguous.
