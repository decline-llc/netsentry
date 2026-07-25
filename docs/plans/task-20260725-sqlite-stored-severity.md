# Task Plan: R90-31 validate stored SQLite alert severity

## Metadata

- Timestamp: 2026-07-25T07:01:27-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `17ef01cb24a3bccfa3b3f0aefa9eba64904a9f7f`

## Goal

Reject persisted alert severities outside the public model before returning or
classifying the row.

## Scope

- Accept only `low`, `medium`, `high`, and `critical` stored severities.
- Reject empty, case-variant, and unsupported stored values with a
  field-specific read error.
- Cover primary list/query decoding and historical cross-shard reads.
- Prove a rejected historical shard remains byte-for-byte unchanged.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not normalize case, substitute a default, repair, delete, migrate, or
  rewrite invalid rows.
- Do not change rule/API severity validation, metrics semantics, the SQLite
  schema, or the recovery-log format.
- Do not broaden this increment to other stored textual fields.
- Do not create a release tag or publish artifacts.

## Risks

- Empty severity can be silently classified as low by downstream statistics.
- Arbitrary or case-variant values can escape the documented API enum.
- Historical rejection must stay on the existing read-only query path.

## Validation

- Direct empty, uppercase-known, and unsupported severity rejection
  regressions across list and query decoding.
- Direct historical-shard rejection with byte preservation.
- Healthy primary and cross-shard severity query compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Only the four public severity values decode successfully.
- Empty, case-variant, and unsupported stored values return a field-specific
  error without substitution or normalization.
- Valid rows retain current list, query, count, metrics, and cross-shard
  behavior.
- A rejected historical shard remains byte-for-byte unchanged.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires automatic row repair, deletion, schema migration,
changing the public severity enum, operator data, tag creation, or publication
authority.

## Delivery Evidence

- Feature commit: `856d1788c7f5abea0116b526ad8d7e2ebd5b9e11`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `17ef01cb24a3bccfa3b3f0aefa9eba64904a9f7f..856d1788c7f5abea0116b526ad8d7e2ebd5b9e11`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
