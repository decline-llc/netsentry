# Task Plan: R90-35 enforce the UDS IPv4 address contract

## Metadata

- Timestamp: 2026-07-25T09:34:42-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `42721e10ef84dd7cf24306f03e23c8ad8d401870`

## Goal

Reject UDS packet frames whose source or destination address is not a strict
IPv4 address, matching the capture parser and documented v0.1 protocol boundary.

## Scope

- Parse UDS packet addresses with an address type that distinguishes IPv4 from
  IPv6 and IPv4-mapped IPv6.
- Reject ordinary IPv6 and IPv4-mapped IPv6 source or destination text.
- Preserve valid IPv4 packet delivery and existing malformed-address behavior.
- Reconcile receiver documentation, roadmap status, and task state.

## Non-Goals

- Do not add IPv6 parsing, capture, matching, storage, or API support.
- Do not change the C sender or packet JSON schema.
- Do not normalize or rewrite address text.
- Do not add persisted SQLite address validation in this increment.
- Do not create a release tag or publish artifacts.

## Risks

- `IP.To4`-style checks can accidentally accept IPv4-mapped IPv6 text.
- Rejection must increment decode errors exactly once and must not enqueue the
  invalid packet.
- A broader address-normalization change could alter aggregation identities.

## Validation

- Direct ordinary IPv6 source and destination rejection.
- Direct IPv4-mapped IPv6 source and destination rejection.
- Direct malformed-address compatibility and valid IPv4 packet delivery.
- Decode-error accounting and no-enqueue assertions for rejected frames.
- Twenty uncached focused receiver race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Only strict IPv4 source and destination addresses pass UDS packet validation.
- Ordinary and IPv4-mapped IPv6 text fails with one decode error and no queued
  packet.
- Valid IPv4 capture frames retain current behavior.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires IPv6 product support, a packet-schema change, C
capture changes, address normalization, stored-data migration, operator data,
tag creation, or publication authority.

## Delivery Evidence

- Implementation complete.
- Twenty uncached focused receiver race runs passed.
- Complete native race suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks passed.
- Commit, push, fetched-remote verification, and exact-range Vault
  synchronization remain pending.
