# Task Plan: R90-34 validate stored SQLite required text

## Metadata

- Timestamp: 2026-07-25T09:26:18-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `cf8f0493b5847c677c8b03922177b18b6a4cbdbb`

## Goal

Reject persisted alert rows whose required public identity, rule, or network
text fields are empty or contain only whitespace.

## Scope

- Require non-blank `event_id`, `rule_id`, `rule_name`, `protocol`, `src_ip`,
  and `dst_ip` during row decoding.
- Cover primary list/query decoding and historical cross-shard reads.
- Prove a rejected historical shard remains byte-for-byte unchanged.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not change aggregation-ID validation or reconcile `event_id` with the
  event ledger.
- Do not add IP syntax, protocol-enum, rule-catalog, or MITRE validation.
- Do not require optional payload, match, or MITRE fields to be non-empty.
- Do not repair, delete, migrate, or rewrite invalid rows.
- Do not create a release tag or publish artifacts.

## Risks

- Checking every non-null text column would incorrectly reject legitimate
  empty optional fields.
- Required-field validation must run before dependent identity checks so the
  reported field is unambiguous.
- Historical rejection must stay on the existing read-only query path.

## Validation

- Direct blank rejection regressions for all six required fields, including
  both empty and whitespace-only values across list and query decoding.
- Direct historical-shard rejection with byte preservation and an encoded
  filesystem path.
- Existing healthy aggregation, pagination, and cross-shard compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Decoded `event_id`, `rule_id`, `rule_name`, `protocol`, `src_ip`, and
  `dst_ip` are non-blank after whitespace trimming.
- Optional text fields retain their current empty-value compatibility.
- Valid rows retain current aggregation, list, query, pagination, and
  cross-shard behavior.
- A rejected historical shard remains byte-for-byte unchanged.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires validating optional fields, changing the public
protocol or address contract, event-ledger reconciliation, automatic row
repair, deletion, schema migration, operator data, tag creation, or publication
authority.

## Delivery Evidence

- Implementation complete.
- Twenty uncached focused alert-store race runs passed.
- Complete native race suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks passed.
- Commit, push, fetched-remote verification, and exact-range Vault
  synchronization remain pending.
