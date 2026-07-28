# Task Plan: R90-46 validate stored SQLite timestamp encoding

## Metadata

- Timestamp: 2026-07-28T01:00:41-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `8d10cd23965ba3d7c3d7f899777ad3b16baf9022`

## Goal

Accept stored aggregate timestamps only when their text exactly matches the
UTC RFC3339Nano encoding emitted by NetSentry's SQLite writer.

## Scope

- Parse `first_seen`, `last_seen`, and `window_start` as RFC3339Nano and
  compare each original value with the writer's canonical UTC encoding.
- Reject explicit UTC offsets, non-UTC offsets, and redundant fractional
  precision even when they represent the same instant as a canonical value.
- Apply the shared decoder behavior to primary list/query reads and
  historical read-only shard queries.
- Preserve historical shard bytes when a noncanonical timestamp is rejected.
- Preserve healthy writer-generated primary and cross-shard behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the SQLite schema, aggregation window, or public time model.
- Do not change recovery-log JSON encoding or validation.
- Do not validate or expose internal `created_at` or `updated_at` columns.
- Do not change SQLite aggregation, ordering, filtering, or pruning
  expressions; variable-width writer timestamp comparison is a separate
  follow-up.
- Do not scan every stored row during startup or before writes.
- Do not repair, rewrite, normalize, delete, or migrate stored timestamps.
- Do not create a release tag or publish artifacts.

## Risks

- Comparing parsed instants alone would continue accepting text that can sort
  differently under SQLite's lexical timestamp comparisons.
- Requiring a fixed fractional width would reject valid RFC3339Nano output,
  whose fractional component is omitted or trimmed when possible.
- Applying the check after timestamp-order validation could obscure the
  noncanonical field that caused the durable contract failure.

## Validation

- Direct primary list/query rejection for noncanonical `first_seen`,
  `last_seen`, and `window_start` text.
- Cover explicit `+00:00`, a non-UTC offset, and redundant fractional zeros as
  distinct writer-incompatible encodings.
- Historical read-only rejection with byte preservation from a directory path
  requiring SQLite URL encoding.
- Healthy primary and historical writer-generated timestamps remain
  compatible.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Shared row decoding rejects any aggregate timestamp whose original text is
  not exactly `formatTime(parsed)`.
- The error names the offending stored field before ordering or identity
  validation can obscure it.
- Historical rejection leaves the shard byte-for-byte unchanged.
- Canonical UTC RFC3339Nano rows retain existing list, query, ordering,
  filtering, aggregation, and cross-shard behavior.
- The increment does not claim that variable-width canonical RFC3339Nano text
  is chronologically sortable in every SQLite text comparison.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires a schema migration, stored-data rewrite, recovery
format or public time-semantics change, `created_at`/`updated_at` validation, a
full-table startup scan, operator data, tag creation, or publication authority.

## Validation Evidence

- Shared row decoding requires each stored aggregate timestamp to equal
  `formatTime(parsed)` before ordering or aggregation-identity checks.
- Primary list/query regressions reject explicit UTC offsets, non-UTC offsets,
  and redundant fractional precision on all three aggregate timestamp fields.
- An encoded-path historical query rejects noncanonical `last_seen` while
  preserving the shard byte-for-byte.
- Canonical fractional writer output remains compatible.
- Twenty uncached focused alert-store race runs passed.
- The complete native race suite, E2E smoke, documentation, configuration, and
  knowledge checks passed.
- Feature commit: `9c6d574ba0f1f9766e9411b41d54b1ddeafb207b`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range:
  `8d10cd23965ba3d7c3d7f899777ad3b16baf9022..9c6d574ba0f1f9766e9411b41d54b1ddeafb207b`
- Vault iteration note, full commit index, `00-MOC` link, and reusable stable
  storage note: verified.
- Variable-width canonical RFC3339Nano comparison remains explicitly queued as
  R90-47 rather than being included in this increment's delivery claim.
