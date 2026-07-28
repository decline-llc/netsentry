# Task Plan: R90-47 pin SQLite timestamp comparison semantics

## Metadata

- Timestamp: 2026-07-28T01:26:36-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `e28e3de9fafc099ae082ffc45ae867e20abf19dd`

## Goal

Make SQLite aggregation, ordering, filtering, and retention comparisons
chronological for canonical variable-width UTC RFC3339Nano timestamps without
losing nanosecond precision or rewriting stored rows.

## Scope

- Define one pure-SQL expression that converts canonical UTC RFC3339Nano text
  into a fixed-width nanosecond sort key.
- Use the shared key for UPSERT `first_seen`/`last_seen` selection and latest
  payload/match selection.
- Use the key for primary and per-shard ordering/pagination, inclusive
  `since`/`until` filters, and retention pruning.
- Add an optional expression index for primary `last_seen` ordering and range
  scans; existing writable databases may create the index through normal
  optional-index initialization.
- Keep legacy historical shards read-only and correct even when they do not
  contain the new optional index.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change the SQLite column schema or rewrite, normalize, or migrate
  stored timestamp text.
- Do not change recovery JSON, the public time model, filter inclusivity,
  aggregation windows, retention cutoffs, or shard selection.
- Do not register a custom SQLite function that would make the database schema
  depend on NetSentry-specific extensions.
- Do not require writable initialization of historical shards.
- Do not create a release tag or publish artifacts.

## Risks

- SQLite `julianday` and similar helpers can collapse sub-millisecond values,
  so they cannot satisfy the nanosecond acceptance boundary.
- A fixed-width expression that mishandles absent or trimmed fractions can
  reorder valid writer output.
- Expression-based comparisons can bypass existing raw-text indexes; the new
  optional expression index must cover the primary global order/range path,
  and the legacy-shard fallback must be documented.
- SQL expression drift between UPSERT, query, and pruning paths could leave one
  operation with the old lexical behavior.

## Validation

- Direct aggregation with exact-second, one-nanosecond, and two-nanosecond
  events verifies earliest/latest timestamps and latest payload/match fields.
- Primary list/query pagination verifies chronological order across absent and
  fractional timestamp text.
- Primary and historical `since`/`until` regressions verify inclusive
  nanosecond boundaries.
- Retention pruning verifies a one-nanosecond cutoff without deleting the
  boundary row.
- Query-plan regression verifies the primary order/range path uses the new
  optional expression index without a temporary order sort.
- Existing-database regression verifies the optional index is recreated;
  historical queries remain read-only through an encoded path.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Every planned SQL inequality or ordering operation over aggregate timestamps
  uses the same fixed-width nanosecond key.
- Exact-second and fractional writer output compares by instant through one
  nanosecond without relying on lossy SQLite date conversion.
- Aggregation, primary/historical ordering and filtering, and pruning satisfy
  their direct regressions without changing stored timestamp bytes.
- Primary order/range queries can use the optional expression index; legacy
  historical shards remain correct without index creation or modification.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Validation Evidence

- Focused alert-store race coverage for aggregation, ordering/filtering,
  query-plan selection, optional-index recreation, historical immutability, and
  retention pruning passed uncached with `-count=20`.
- The complete native suite passed with the race detector.
- E2E smoke passed with 6 packets, 5 alerts, and 8 loaded rules.
- Documentation, configuration, knowledge, JSON, and diff checks passed.
- The primary range/order query plan uses
  `idx_alerts_last_seen_time_id` without a temporary order sort.
- A legacy historical shard without the expression index remained
  byte-for-byte unchanged while its nanosecond range results remained
  chronological.

## Delivery Evidence

- Feature commit: `046f89673491b2bab78d6c21eedc067fa9c8584b`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed all 33 tests.
- Exact Vault range:
  `e28e3de9fafc099ae082ffc45ae867e20abf19dd..046f89673491b2bab78d6c21eedc067fa9c8584b`
- Vault feature note, full commit index, MOC link, and reusable stable storage
  note: verified.
- No later engineering increment was selected; the next `$netsentry-next`
  trigger will refresh the rolling roadmap.

## Stop Conditions

Stop if completion requires a schema or stored-data migration, a
NetSentry-specific SQLite extension in durable schema, loss of nanosecond
fidelity, changed public time/filter/retention semantics, writable historical
access, operator data, tag creation, or publication authority.
