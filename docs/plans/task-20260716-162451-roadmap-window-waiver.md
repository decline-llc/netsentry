# Task Plan: global roadmap window waiver

## Metadata

- Timestamp: 2026-07-16T16:24:51Z
- Branch: main
- Risk Level: Low
- Remote baseline: `f8d766e9dbadbb0433ad0dd31047ae04f0055950`

## Goal

Remove every planning-window start and end restriction from unfinished
NetSentry roadmap work while preserving dependency, evidence, release, tagging,
and publication controls.

## Scope

- Make roadmap dates informational forecasts rather than eligibility gates.
- Reclassify unfinished increments using dependency readiness.
- Reconcile completed R90-05 resume guidance with the global waiver.

## Non-Goals

- Do not begin or complete R90-06 in this documentation increment.
- Do not approve a release, create or move a tag, or publish artifacts/images.
- Do not weaken evidence, security, privacy, or acceptance requirements.

## Validation

- `make docs-check`
- `make knowledge-check`
- JSON parsing for the updated task state
- `git diff --check`
- Sensitive-information review

## Acceptance Criteria

- No unfinished increment is blocked solely by a roadmap date.
- R90-06 is recorded as ready because R90-05 is complete.
- Dates remain available as historical forecasts.
- Tagging and publication still require separate explicit authorization.

## Stop Conditions

Stop if the waiver is ambiguous about dependency or publication authority, or
if validation exposes unrelated repository drift.
