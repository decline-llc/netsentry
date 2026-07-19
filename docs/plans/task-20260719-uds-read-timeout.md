# Task Plan: R90-13 bound idle UDS receiver connections

## Metadata

- Timestamp: 2026-07-19T01:29:53-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `679ec10a9bb3fddf30db6df831b7dba35ccaeefb`

## Goal

Prevent an accepted local UDS client from occupying a finite receiver handler
slot indefinitely without delivering complete frames.

## Scope

- Add a validated `engine.uds_read_timeout_seconds` setting with a finite
  default and explicit lower/upper bounds.
- Apply the timeout when a connection handler starts and refresh it after each
  complete frame.
- Treat idle expiry as connection lifecycle rather than malformed-frame input.
- Add direct pre-first-frame expiry, deadline-refresh, idle-capacity-reuse, and
  configuration-bound regressions.
- Reconcile public configuration, architecture, testing, changelog, roadmap,
  and task-state documentation.

## Non-Goals

- Do not change the UDS JSON frame schema, maximum frame size, socket mode, or
  connection limit.
- Do not add UDS authentication, peer-credential checks, session sequencing,
  or a mandatory hello state machine.
- Do not change C capture reconnect behavior or HTTP timeouts.
- Do not create a release tag or publish artifacts.

## Risks

- A timeout that is too short can disconnect a legitimately quiet capture
  process and increase reconnect churn.
- Failing to refresh the deadline after valid traffic can terminate active
  sessions.
- Timeout errors counted as decode failures can distort protocol-health
  metrics.

## Validation

- Direct lower/upper configuration-bound regressions.
- Direct partial-frame timeout before the first complete frame.
- Direct proof that the read deadline is refreshed after each complete frame.
- Integration proof that idle expiry releases a saturated connection slot for
  a replacement client.
- Repeated focused receiver/config race tests, full native C/Go race suite,
  E2E smoke, documentation/configuration/knowledge checks, JSON parsing,
  `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- `engine.uds_read_timeout_seconds` defaults to 30 and accepts only 1–3600.
- Every accepted connection has a finite read deadline before its first frame,
  and every complete frame refreshes that deadline.
- A partial or silent frame expires, closes its handler, and releases capacity.
- Idle expiry does not increment the malformed-frame decode-error counter.
- Active frame processing, reconnect, and context-cancel shutdown remain
  compatible.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if safe completion requires a UDS protocol change, authentication or peer
identity policy, C capture changes, operator data, or tag/publication authority.
