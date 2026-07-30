# Task Plan: R90-55 derive the recovery contract from model.Alert

## Metadata

- Timestamp: 2026-07-30T09:52:47-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `b52d8df4de2389f6f31670e6dc83884f588f0802`

## Goal

Replace independently maintained recovery field-name, required-field, and
JSON-kind lists with one immutable contract derived from the current
`model.Alert` writer shape.

## Scope

- Build the recovery contract once from the exported fields, Go types, and
  `json` tags on `model.Alert`.
- Preserve declared model field order for deterministic missing-field and
  value-error diagnostics.
- Derive canonical names from JSON tags, required status from supported
  omission options, JSON kinds from supported writer types, and integer
  encoding policy from integral types.
- Keep case-variant alias handling aligned with the existing top-level
  validation behavior.
- Ignore fields the JSON writer ignores, and fail closed during contract
  construction for ambiguous embedded fields, duplicate canonical names, or
  writer types whose output kind cannot be derived safely.
- Drive name, presence, kind, optional-field, and canonical-integer checks from
  the derived contract without per-record reflection.
- Add direct synthetic-model contract tests and retain all existing startup,
  runtime, preservation, precedence, and canonical writer regressions.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change `model.Alert`, its public JSON, field order, or omission
  behavior.
- Do not change recovery semantic validation, numeric spelling rules,
  timestamp rules, duplicate/member diagnostic precedence, or the JSONL
  format.
- Do not use reflection on every recovery record or infer arbitrary custom
  marshaler output.
- Do not generalize the contract to unrelated API or SQLite models.
- Do not add a recovery migration, rewrite rejected input, start R90-56,
  create a release tag, publish artifacts, or change release authority.

## Risks

- Incomplete JSON-tag option handling could mark an optional writer field as
  required or silently ignore a newly emitted field.
- Generic type-to-kind inference can disagree with custom marshalers or the
  `,string` option.
- Reordering a derived contract through a map would change deterministic
  missing-field diagnostics.
- Automatically treating every numeric type as a canonical integer would
  reject legitimate future floating-point writer output.
- Package-initialization failure must be reserved for ambiguous or unsupported
  future model shapes and must not affect the current model.

## Validation

- Direct contract-builder tests cover required and omitted fields, default and
  explicit JSON names, ignored and unexported fields, strings, booleans,
  integers, floating-point numbers, `time.Time`, `omitempty`, `omitzero`, and
  `,string`.
- Direct failure tests cover duplicate case-insensitive names, anonymous
  fields, and unsupported pointer/container/custom model shapes.
- A model-to-contract alignment regression compares canonical writer output
  with every derived `model.Alert` field, required/optional status, JSON kind,
  and integer policy.
- Existing missing-field startup and runtime matrices remain derived from all
  required model fields.
- Existing alias, wrong-kind, optional `raw_payload`, numeric encoding,
  diagnostic precedence, startup/runtime preservation, and canonical replay
  regressions pass unchanged.
- Twenty uncached focused alert-store contract race runs.
- Complete native test suite, E2E smoke, documentation, configuration,
  knowledge, JSON, formatting, diff, and sensitive-information checks.

## Acceptance Criteria

- One contract derived from `model.Alert` supplies every supported recovery
  canonical name, required/optional decision, JSON kind, and integer policy.
- Adding or changing a supported writer field automatically changes recovery
  validation; ambiguous or unsupported shapes fail explicitly instead of
  bypassing validation.
- Contract construction occurs once, and per-record validation uses only the
  immutable result.
- Current public JSON, field order, diagnostics, preservation boundaries, and
  canonical writer replay remain unchanged.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the single local Vault.

## Evidence Map

| Acceptance criterion | Planned evidence |
| --- | --- |
| Single derived contract | Production code contains no independent recovery name/required/kind list |
| Supported changes are automatic | Synthetic model builder test covers names, omission, kinds, integer policy, and writer order |
| Unsupported changes fail closed | Builder rejection tests for ambiguity and unsupported shapes |
| Current behavior unchanged | Existing missing, alias, kind, optional, numeric, precedence, preservation, and replay suites |
| No per-record reflection | Contract is built once and validator looks up immutable entries |
| Delivery | Fetched remote equality, post-fetch knowledge gate, exact-range Vault note/index/MOC, and stable storage note |

## Validation Deviation

- **Observed:** The first combined formatting, diff, and focused-test command
  invoked a Go package path from the repository root even though the Go module
  is rooted under `engine`; the Go command stopped before testing. A later
  pre-commit review resolved the effective tools in both locations and found
  that `engine` uses its pinned Go 1.25.12 toolchain, whose `encoding/json`
  writer also supports `omitzero`.
- **Impact:** The first multi-step result was discarded under the fail-fast
  evidence rule, including earlier formatting and diff output. The initial
  implementation covered `omitempty` but not `omitzero`, so all validation
  completed before that correction was also invalidated.
- **Resolution:** Contract derivation and direct writer alignment now cover
  both omission options, including the difference between
  `time.Time,omitempty` and `time.Time,omitzero`. Twenty focused race runs, the
  complete native race suite, E2E, and the documentation, configuration,
  knowledge, JSON, formatting, and diff gates were rerun from the correct
  roots. The local `netsentry-next` skill now requires direct language checks
  to resolve and use the owning module root.

## Validation Evidence

- One package-initialized contract now derives every recovery field's declared
  order, canonical JSON name, actual supported `omitempty`/`omitzero` behavior,
  JSON kind, and integral encoding policy from `model.Alert`.
- Runtime validation performs only immutable field lookups; there is no
  per-record reflection or independent recovery name, required-field, or kind
  list.
- Direct synthetic-model tests cover default and explicit names, required and
  optional fields, ignored and unexported fields, strings, booleans, signed and
  unsigned integers, floating-point values, `time.Time`, both omission options,
  and `,string`.
- Direct construction failures cover non-struct input, anonymous fields,
  invalid tags, case-insensitive duplicate names, pointers, slices, maps,
  `json.Number`, unsupported quoted shapes, and custom JSON/text marshalers.
- The zero-value and normalized current writers align with the derived
  contract's order, required/optional status, JSON kinds, and canonical integer
  policy; optional empty and populated `raw_payload` both replay.
- Existing duplicate, alias, missing, wrong-kind, numeric-spelling, diagnostic
  precedence, startup/runtime preservation, and replay regressions pass.
- Twenty uncached focused alert-store contract race runs passed.
- The complete C and Go native race suite passed.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge (33 tests), task-state JSON,
  formatting, and diff checks passed.

## Plan Audit

- The trigger audit reconciled the prior R90-54 feature and closure commits,
  fetched `origin/main`, both exact Vault notes, the full index, MOC links, and
  the stable storage note before selecting work.
- The recent delivery history still has no material deviation beyond the dated
  R90-53 audit and the recorded R90-54 validation correction.
- All 62 roadmap rows have matching Definitions. Every unfinished increment
  retains a dependency, delivery window, risk, acceptance criteria, required
  validation, and stop condition.
- R90-56 remains planned behind R90-55 and was not started. R90-57 remains
  blocked on a product decision, R90-58 remains planned behind R90-56, and
  R90-59 remains blocked on explicit publication authorization.
- Every R90-55 acceptance criterion has verified local and delivered evidence;
  the docs-only closure record is the only remaining action in this trigger,
  and R90-56 was not started.

## Delivery Evidence

- Feature commit: `20161c20db271c5dbe9f5acc3f268eb5b8308494`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `b52d8df4de2389f6f31670e6dc83884f588f0802..20161c20db271c5dbe9f5acc3f268eb5b8308494`
- Vault feature note, full commit index, MOC link, idempotent identical-range
  replay, and reusable stable storage note: verified.
- R90-56 is the next ready increment and was not started.

## Stop Conditions

Stop if completion requires changing public JSON or field order, accepting an
ambiguous custom writer shape, per-record reflection, a recovery-format
migration, automatic repair, operator data, external publication authority, or
starting R90-56.
