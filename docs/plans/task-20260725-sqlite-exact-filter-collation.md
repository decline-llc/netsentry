# Task Plan: R90-29 pin SQLite exact-filter collation

## Metadata

- Timestamp: 2026-07-25T06:45:27-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `3360e16effae4b0f9e8ecc911907d492e7c6d75a`

## Goal

Make documented exact-match alert filters independent of compatible
operator-declared SQLite column collations.

## Scope

- Apply explicit binary comparison to `rule_id`, `severity`, `src_ip`, and
  `dst_ip` query predicates.
- Prove primary and historical-shard queries do not broaden exact matches when
  compatible columns use `NOCASE`.
- Preserve intentionally case-insensitive protocol and MITRE filters.
- Reconcile query documentation, roadmap status, and task state.

## Non-Goals

- Do not change the SQLite schema, preflight compatibility policy, indexes, or
  aggregation identity.
- Do not change protocol, MITRE, keyword, time, port, or count filter semantics.
- Do not migrate or rewrite operator databases.
- Do not create a release tag or publish artifacts.

## Risks

- Applying binary collation to the wrong predicates could break intentionally
  case-insensitive filters.
- A direct list query could be fixed while filtered counts or cross-shard
  queries retain broadened results.
- Compatible custom collations must not alter the documented exact-match
  contract.

## Validation

- Direct primary-store regressions for rule, severity, source, and destination
  exact filters against compatible `NOCASE` columns.
- Direct historical-shard regression through the cross-shard query path.
- Existing protocol/MITRE case-insensitive query compatibility.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Rule, severity, source, and destination filters use binary equality for both
  row selection and filtered counts.
- Compatible `NOCASE` schemas cannot broaden those exact-match results in
  primary or historical-shard queries.
- Protocol and MITRE filters remain case-insensitive.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires schema migration, rejecting an otherwise compatible
database, changing public filter semantics, operator data, tag creation, or
publication authority.

## Delivery Evidence

- Feature commit: `f12a454c95515dd92549e33e3c56d00449408d89`
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range: `3360e16effae4b0f9e8ecc911907d492e7c6d75a..f12a454c95515dd92549e33e3c56d00449408d89`
- Vault iteration note, full commit index, and `00-MOC` link: verified.
