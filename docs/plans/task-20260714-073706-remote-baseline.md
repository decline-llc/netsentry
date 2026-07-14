# Task Plan: require remote knowledge baseline after push

## Metadata

- Timestamp: 2026-07-14T07:37:06Z
- Branch: main
- Risk Level: Low

## Goal

Make fetched `origin/main`, rather than local Git state alone, the authoritative baseline for post-push knowledge validation and active roadmap/task-state reconciliation.

## Scope

- Fetch `origin/main` immediately after each successful push.
- Require the fetched remote SHA and checked working tree to match before `make knowledge-check` or delivery claims.
- Reconcile active roadmap and task-state records to that remote SHA while retaining historical SHA evidence.
- Record the early completion of R90-03.

## Non-Goals

- Do not rewrite historical commit evidence to the latest remote SHA.
- Do not start R90-04 or acquire production-derived traffic evidence.

## Validation and Stop Conditions

- Fetch and compare `HEAD` with `origin/main`, run `make knowledge-check`, `make docs-check`, and `git diff --check` before commit.
- Stop if the fetched remote ref differs from the proposed delivery commit or knowledge validation fails.
