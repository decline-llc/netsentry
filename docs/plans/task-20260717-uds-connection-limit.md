# Task Plan: R90-07 bound concurrent UDS connections

## Metadata

- Timestamp: 2026-07-17T00:00:00-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `f2c9dc20f894dd2f279d0a35b9015e484e1f8f51`

## Goal

Bound the number of concurrent clients handled by the Go UDS receiver so local
clients cannot cause unbounded connection-handler goroutine growth.

## Scope

- Add a validated `engine.uds_max_connections` configuration value.
- Reject accepted connections immediately while the configured capacity is
  exhausted.
- Release capacity when a handler exits so later capture reconnects work.
- Add direct receiver and configuration regression tests.
- Reconcile public configuration, architecture, development, roadmap, and task
  state documentation.

## Non-Goals

- Do not change the UDS frame contract, packet backpressure, or socket
  permissions.
- Do not add authentication to the local UDS protocol.
- Do not create a release tag or publish artifacts.
- Do not begin another v0.2 feature.

## Risks

- A leaked capacity slot could permanently reject valid reconnects.
- A blocking limiter could stall shutdown or obscure overload behavior.
- An invalid zero or excessive limit could silently disable the bound.

## Validation

- Direct config tests for the lower and upper rejection boundaries.
- Direct receiver tests for excess-client rejection and capacity reuse.
- Focused Go tests, configuration/docs checks, `make test`,
  `make knowledge-check`, `git diff --check`, and sensitive-information review.

## Acceptance Criteria

- The repository config explicitly sets a finite concurrent connection limit.
- Configuration rejects values below 1 and above 1024.
- The receiver never starts a handler above the configured capacity, closes an
  excess accepted connection, and accepts a later connection after capacity is
  released.
- Existing reconnect and shutdown behavior continues to pass.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if validation is ambiguous, the limiter requires changing the frame
protocol, or delivery would require tag/publication authority.

## Completion

- Feature commit: `bdca014a5ca3c775125b41d98faf15ffd1b1cf35`
- Fetched `origin/main`: verified at the feature commit
- Focused receiver test: passed 10 consecutive runs
- Full native C/Go race suite: passed
- Config, docs, knowledge, JSON, diff, and sensitive-information checks: passed
- Vault range:
  `f2c9dc20f894dd2f279d0a35b9015e484e1f8f51..bdca014a5ca3c775125b41d98faf15ffd1b1cf35`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
