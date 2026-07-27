# Task Plan: R90-44 validate recovery protocol names

## Metadata

- Timestamp: 2026-07-27T04:53:16-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `d4334654a8eafee84f9cc1de69ea6d733a4a2160`

## Goal

Reject recovery-log protocol names that the normalized alert writer cannot
canonically emit before startup replay or runtime append can modify the
complete log or missing/existing SQLite state.

## Scope

- Apply the shared canonical protocol-name validator during recovery preflight.
- Accept exactly `TCP`, `UDP`, `ICMP`, and `PROTO_<number>` for otherwise
  unnamed uint8 IP protocol numbers.
- Reject named case variants, arbitrary names, malformed or out-of-range
  `PROTO_` values, noncanonical decimal forms, and numeric aliases of named
  protocols.
- Cover startup rejection after a valid prefix for missing and compatible
  existing databases.
- Cover runtime rejection while preserving the complete existing recovery log
  and SQLite database.
- Preserve valid canonical replay and runtime-write behavior.
- Reconcile storage documentation, roadmap status, and task state.

## Non-Goals

- Do not restrict packet-frame protocol numbers or change the UDS schema.
- Do not add IP protocol parsing or names beyond the current writer contract.
- Do not change protocol query-filter semantics.
- Do not normalize, repair, delete, or rewrite recovery records or stored rows.
- Do not change the SQLite schema.
- Do not create a release tag or publish artifacts.

## Risks

- Running protocol validation after dependent identity checks could obscure the
  durable-field error or allow state modification before rejection.
- Treating only named protocols as valid would reject unknown IPv4 protocol
  numbers that the current writer legitimately emits.
- Tests that mutate a normalized protocol must recompute deterministic record
  identities so they isolate protocol validation.

## Validation

- Direct startup rejection after a valid prefix for a named case variant,
  arbitrary name, missing/negative/leading-zero/padded numeric forms,
  named-protocol numeric aliases, and an out-of-range number.
- Missing and compatible existing database preservation for every startup
  rejection condition.
- Direct runtime rejection for every distinct noncanonical category with
  complete recovery-log and database byte preservation.
- Direct startup replay and runtime-write compatibility for `TCP`, `UDP`,
  `ICMP`, `PROTO_0`, `PROTO_2`, and `PROTO_255`.
- Twenty uncached focused alert-store race runs.
- Complete native test suite, E2E smoke, documentation, knowledge, JSON, diff,
  and sensitive-information checks.

## Acceptance Criteria

- Recovery protocol text is accepted if and only if it equals the canonical
  name emitted for some uint8 IP protocol number.
- Every distinct rejection condition promised by the plan has a direct
  startup and runtime regression.
- Rejection reports the recovery record number and a clear invalid-protocol
  error.
- Startup rejection preserves the complete log and leaves a missing database
  absent or a compatible existing database byte-for-byte unchanged.
- Runtime rejection preserves the complete log and database byte-for-byte.
- Named and unknown canonical protocol records replay and write successfully.
- Required validation passes and delivery is verified against fetched
  `origin/main` and the local Vault.

## Stop Conditions

Stop if completion requires restricting UDS protocol numbers, changing
protocol-filter semantics, a recovery-format change, data migration, automatic
repair or deletion, operator data, tag creation, or publication authority.

## Delivery Evidence

- Feature commit: `a87b2161bf65b726d827a805f21aa209bd71ed3b`
- Recovery preflight now applies the same canonical protocol-name validator as
  stored-row decoding before dependent identity validation.
- Direct startup and runtime regressions cover every planned noncanonical form,
  complete-log preservation, missing/existing database preservation, and clear
  record-numbered errors.
- Direct replay and runtime-write regressions cover `TCP`, `UDP`, `ICMP`,
  `PROTO_0`, `PROTO_2`, and `PROTO_255`.
- Twenty uncached focused alert-store race runs, the complete native race
  suite, E2E smoke, documentation, config, knowledge, JSON, and diff checks
  passed.
- Fetched `origin/main`: verified equal to the feature commit after push.
- Post-fetch knowledge check: passed.
- Vault range:
  `d4334654a8eafee84f9cc1de69ea6d733a4a2160..a87b2161bf65b726d827a805f21aa209bd71ed3b`
- Vault iteration note, full commit index, `00-MOC` link, and reusable stable
  storage note: verified.
