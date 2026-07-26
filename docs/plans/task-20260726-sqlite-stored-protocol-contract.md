# Task Plan: R90-43 validate stored SQLite protocol names

## Metadata

- Timestamp: 2026-07-26T00:05:17-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `9b6ddf92557c1cc7c13bdebc1c4de5166b8067ad`

## Goal

Reject stored alert protocol names that the rule engine cannot canonically emit,
without narrowing the current uint8 IP protocol-number boundary or modifying
historical shard bytes.

## Scope

- Share the rule engine's canonical protocol-name formatter with stored-row
  decoding.
- Accept exactly `TCP`, `UDP`, `ICMP`, and `PROTO_<number>` for otherwise
  unnamed uint8 IP protocol numbers.
- Reject case variants, arbitrary names, malformed or out-of-range `PROTO_`
  values, noncanonical decimal forms, and numeric aliases of named protocols.
- Apply validation through shared primary and historical row decoding.
- Preserve rejected historical shard bytes through a URL-encoded read-only
  path.
- Preserve current protocol query-filter case semantics.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not restrict packet-frame protocol numbers or change the UDS schema.
- Do not add IP protocol parsing or names beyond the current writer contract.
- Do not change protocol query-filter case semantics.
- Do not normalize, repair, delete, or rewrite stored rows.
- Do not change the SQLite schema or recovery-log contract.
- Do not create a release tag or publish artifacts.

## Risks

- Duplicating formatter rules between the rule engine and storage decoder could
  let the contracts drift.
- Treating only TCP and UDP as valid would reject unknown IPv4 protocol numbers
  that the current capture and rule paths can emit.
- Loose numeric parsing could accept leading zeros, whitespace, named-protocol
  numeric aliases, or values outside the uint8 boundary.

## Validation

- Direct primary list/query rejection for a named case variant, arbitrary name,
  missing/negative/leading-zero/padded numeric forms, named-protocol numeric
  aliases, and an out-of-range number.
- Direct compatibility for `TCP`, `UDP`, `ICMP`, and unknown protocol boundary
  names including `PROTO_0` and `PROTO_255`.
- Historical read-only rejection with byte preservation through a directory
  containing spaces.
- Focused alert-store, rule, and shared-model package tests.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Stored protocol text is accepted if and only if it equals the canonical name
  emitted for some uint8 IP protocol number.
- Every distinct noncanonical form promised by the plan has a direct
  regression.
- Primary and historical reads fail with a clear invalid-protocol error.
- Historical rejection leaves the shard byte-for-byte unchanged.
- Named and unknown protocol-number output remains compatible.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires restricting UDS protocol numbers, changing
protocol-filter semantics, data migration, automatic repair or deletion,
operator data, tag creation, or publication authority.

## Delivery Evidence

- Feature commit: `8b030b205f50768c9051354d19ec680b46ba876c`
- Shared canonical protocol-name formatting is implemented across rule emission
  and stored-row decoding.
- Direct regressions cover every planned rejection category, canonical named
  and unknown boundary values, and historical byte preservation through an
  encoded path.
- The initial full native suite exposed a lowercase synthetic shutdown fixture;
  the fixture now matches production `TCP`, twenty uncached focused shutdown
  race reruns pass, and the complete native race rerun passes.
- Twenty uncached focused alert-store race runs, E2E smoke, documentation,
  knowledge, JSON, and diff checks passed.
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range:
  `9b6ddf92557c1cc7c13bdebc1c4de5166b8067ad..8b030b205f50768c9051354d19ec680b46ba876c`
- Vault iteration note, full commit index, `00-MOC` link, and reusable stable
  storage note: verified.
