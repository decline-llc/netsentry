# Task Plan: R90-41 validate stored SQLite MITRE tuples

## Metadata

- Timestamp: 2026-07-25T11:21:21-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `9c62066e6d2bea08295a2c4a7c1ee5069a59386f`

## Goal

Reject stored alerts whose MITRE tactic, technique ID, and technique name are
not either all empty or all nonblank before exposing them through list/query
reads.

## Scope

- Accept a stored MITRE tuple when all three fields are exactly empty.
- Accept a stored MITRE tuple when each field contains a non-whitespace
  character.
- Reject every partial empty/populated tuple shape.
- Reject a tuple containing a whitespace-only tactic, technique ID, or
  technique name.
- Apply validation through shared primary and historical row decoding.
- Preserve historical shard bytes on rejection, including an encoded
  filesystem path.
- Preserve complete tuple text exactly without normalization.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not validate stored tuples against the current MITRE catalog.
- Do not normalize case, trim text, or change alert filter semantics.
- Do not validate or rewrite recovery records in this increment.
- Do not scan, migrate, repair, delete, or rewrite persisted rows.
- Do not change the SQLite schema.
- Do not create a release tag or publish artifacts.

## Risks

- Treating all blank text as empty would expose whitespace-only public metadata.
- Catalog revalidation would make current code decide historical tuple
  validity.
- Validation must remain in the shared decoder so primary and historical reads
  cannot diverge.

## Validation

- Direct rejection for all six non-empty proper subsets of the three MITRE
  fields across list/query decoding.
- Direct whitespace-only rejection for each tuple member.
- Historical read-only rejection with byte preservation through a directory
  requiring URL encoding.
- Direct all-empty and fully populated compatibility without normalization.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Stored MITRE tuples are either exactly all empty or contain non-whitespace
  text in all three fields.
- Every partial or whitespace-only tuple fails before being returned.
- Primary list/query reads share the same rejection behavior.
- Historical rejection preserves shard bytes through the read-only boundary.
- Complete tuple values are returned unchanged; all-empty tuples remain valid.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires current-catalog revalidation, text normalization,
filter changes, recovery validation, automatic row repair or deletion, schema
migration, operator data, tag creation, or publication authority.

## Delivery Evidence

- Shared stored-row MITRE tuple validation implemented with all six partial
  shapes, each whitespace-only field, encoded-path historical preservation,
  all-empty compatibility, and complete unnormalized tuple regressions.
- Twenty uncached focused alert-store race runs passed.
- `make test`, `make e2e-smoke`, `make docs-check`, explicit-Vault
  `make knowledge-check`, task-state JSON parsing, `git diff --check`, and the
  sensitive-information review passed.
- Pending commit, push, fetched-remote verification, post-fetch knowledge
  validation, and exact-range Vault synchronization.
