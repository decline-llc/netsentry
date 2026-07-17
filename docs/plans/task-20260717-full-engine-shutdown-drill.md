# Task Plan: R90-08 active-load full-engine shutdown drill

## Metadata

- Timestamp: 2026-07-17T00:57:08-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `80b9bf84915de20d55d13c38ed019c09c1f896d7`

## Goal

Close the documented full-engine shutdown validation gap by exercising the real
UDS receiver, pipeline worker, HTTP API, and SQLite store while packet matching
is in flight, then proving teardown finishes in a bounded time without a store
write after close.

## Scope

- Make HTTP API startup return bind failures synchronously and expose a done
  signal that represents completed shutdown.
- Make the engine wait for receiver, worker, and HTTP API shutdown before the
  deferred SQLite close.
- Add one integration regression using real receiver, worker, HTTP, and SQLite
  components with deterministic channels around the in-flight matcher.
- Reconcile architecture, development, changelog, roadmap, and task state.

## Non-Goals

- Do not change signal handling, packet/frame contracts, API contracts, or
  SQLite schemas.
- Do not add timing-only sleeps as shutdown synchronization.
- Do not exercise production traffic, privileged services, release tagging, or
  publication.
- Do not begin another roadmap increment.

## Risks

- An HTTP shutdown signal may report completion before active handlers finish.
- A timing-sensitive test may be flaky and conceal rather than expose ordering.
- Closing SQLite before workers or API handlers stop can create post-close
  access failures.

## Validation

- Repeat the focused integration test under the race detector.
- Run the full native C/Go race suite.
- Run documentation and knowledge checks, `git diff --check`, JSON parsing, and
  sensitive-information review.

## Acceptance Criteria

- HTTP bind errors are returned to startup instead of only logged in a
  background goroutine.
- Engine shutdown waits for receiver handlers, pipeline workers, and HTTP
  shutdown before the SQLite store is closed.
- The integration drill processes at least one alert, cancels with another
  match in flight, proves shutdown waits for that work, finishes within a
  bounded timeout, closes the API listener, and records zero writes after store
  close.
- The focused race test is stable across repeated runs and the full native test
  suite passes.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if deterministic validation requires production traffic, privileged
external services, a broad runtime architecture change, or any tag/publication
authority.

## Completion

- Feature commit: `9129d4ecf9df0da9601a027ec118af6f58b96e9a`
- Fetched `origin/main`: verified at the feature commit
- Focused race validation: 25 consecutive passes
- Full native C/Go race suite and `make e2e-smoke`: passed
- Documentation, knowledge, JSON, diff, and sensitive-information checks:
  passed
- Vault range:
  `80b9bf84915de20d55d13c38ed019c09c1f896d7..9129d4ecf9df0da9601a027ec118af6f58b96e9a`
- Vault iteration note, full index, and MOC link: verified
- Tag/publication actions: none
