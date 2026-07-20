# Task Plan: R90-14 enforce the UDS hello/session state machine

## Metadata

- Timestamp: 2026-07-20T04:38:43-07:00
- Branch: main
- Risk Level: Medium
- Remote baseline: `7d33e142222d9541c6ec4535a02876a96df607c2`

## Goal

Require each UDS connection to establish exactly one hello/session boundary
before it can send packet or heartbeat frames.

## Scope

- Verify the checked-in C sender sends hello first on initial and replacement
  connections.
- Enforce hello-before-packet/heartbeat, reject duplicate hello, and require
  heartbeat session IDs to match the connection hello.
- Close only a violating connection and increment decode errors exactly once.
- Preserve valid reconnect, capacity reuse, idle expiry, and cancellation.
- Reconcile protocol documentation, roadmap status, and task state.

## Non-Goals

- Do not add peer authentication or peer-credential policy.
- Do not accept ambiguous legacy packet-first connections.
- Do not change frame schemas, connection limits, or read deadlines.
- Do not create a release tag or publish artifacts.

## Risks

- A receiver-only ordering change can reject the checked-in reconnect path.
- Connection-local state can accidentally become global across concurrent
  capture clients.
- A protocol violation can be double-counted or leave a handler slot occupied.

## Validation

- Direct packet-before-hello, heartbeat-before-hello, duplicate-hello, and
  mismatched-session rejection tests.
- Valid hello, heartbeat, packet, reconnect, capacity, idle, and cancellation
  regressions.
- Direct C reconnect-handshake coverage if a sender change is authorized.
- Focused receiver race tests, full native tests, E2E smoke, documentation,
  knowledge, diff, and sensitive-information checks.

## Acceptance Criteria

- Every connection accepts exactly one valid hello before other frames.
- Heartbeat session IDs match the hello on that connection.
- State violations close only the offending connection, release its capacity,
  and increment decode errors once.
- The checked-in sender remains compatible across reconnect and shutdown.
- The increment is pushed, fetch-verified, and synchronized to the Vault.

## Stop Conditions

Stop if the checked-in C sender does not satisfy the proposed ordering,
compatibility requires ambiguous legacy clients, peer authentication is
required, or work reaches tag/publication authority.

## Reconnect Authorization

The required sender preflight fired the stop condition. The initial connection
sends hello from `capture/src/main.c`, but its `UDS_ERR_PIPE` recovery calls
`uds_reconnect` without sending hello on the replacement socket. The sender
helper in `capture/src/uds_sender.c` only reconnects that socket. A receiver-only
implementation would reject the next packet or heartbeat from a valid checked-in
capture process.

On 2026-07-20, the user explicitly authorized changing the C reconnect path to
send hello before any packet or heartbeat on each replacement connection.
R90-14 resumed with direct C socket-reconnect, Go connection-local state, full
native, and E2E coverage required before delivery.

## Local Validation

- C sender reconnect tests passed 20 consecutive real Unix-socket runs and
  prove hello is the first replacement-connection frame.
- Receiver race tests passed 20 consecutive runs across all four state
  rejection conditions, connection isolation, capacity reuse, concurrent
  connection-local sessions, valid reconnect, idle expiry, and cancellation.
- Full native C and Go race tests, C ASan tests, E2E smoke, documentation,
  knowledge, JSON, and diff checks passed.
- Commit, remote verification, and Vault synchronization remain pending.
