# NetSentry API Reference

> **Status**: development snapshot. This document separates endpoints implemented today from the planned v0.1.0 API contract.

---

## Implemented Today

Default base URL: `http://127.0.0.1:8080`. `engine.api_listen_host` defaults to loopback; a non-loopback value is rejected unless Bearer authentication is enabled with a non-empty token.

### `GET /api/health`

Returns a minimal liveness response by default.

```json
{
  "status": "ok",
  "alerts": 5
}
```

With `verbose=true`, returns capture heartbeat status, engine queue/rule counts, storage status and available filesystem bytes when the store path is known, and throughput counters. Capture status is `unknown` before the first heartbeat, `ok` while the latest heartbeat is within `engine.health_freshness_limit_seconds`, and `stale` after that limit. Storage status is `ok` by default, becomes `degraded` after ordinary SQLite write/query errors until a later successful write or full alert list query clears it, and becomes `emergency` for disk-full, quota, read-only filesystem, or disk I/O failures. Emergency mode is intentionally sticky until restart after operator cleanup.

```json
{
  "status": "ok",
  "alerts": 5,
  "capture": {
    "status": "ok",
    "session_id": "capture-123",
    "last_heartbeat_at": "2026-06-27T12:00:00Z",
    "heartbeat_age_seconds": 1.2,
    "freshness_limit_seconds": 30
  },
  "engine": {
    "queue_depth": 0,
    "rules_loaded": 8
  },
  "storage": {
    "status": "ok",
    "alerts": 5,
    "available_bytes": 123456789
  },
  "throughput": {
    "frames_total": 12,
    "packets_received": 5,
    "packets_processed": 5,
    "decode_errors": 0
  }
}
```

### `GET /api/alerts`

Returns SQLite-backed aggregated alerts ordered by most recent activity.

Query parameters:

| Name | Description |
| --- | --- |
| `page` | Positive page number. Defaults to `1`. |
| `per_page` | Page size from `1` to `100`. Defaults to `20`. |
| `rule_id` | Exact, case-sensitive rule ID match. |
| `severity` | Exact, case-sensitive match for one of `low`, `medium`, `high`, `critical`. |
| `src_ip` | Exact, case-sensitive source IP text match. |
| `dst_ip` | Exact, case-sensitive destination IP text match. |
| `protocol` | Exact protocol match; compared case-insensitively. |
| `dst_port` | Destination port from `0` to `65535`. |
| `since` | Include alerts whose `last_seen` is greater than or equal to this RFC3339 timestamp. |
| `until` | Include alerts whose `last_seen` is less than or equal to this RFC3339 timestamp. |
| `mitre_tactic` | MITRE tactic match; compared case-insensitively. |
| `mitre_technique_id` | MITRE technique ID match; compared case-insensitively. |
| `matched_keyword` | Case-insensitive substring match against the recorded matched keyword. |
| `min_count` | Minimum `aggregated_count`; must be a positive integer. |

Pagination requires both the positive `page` value and `(page - 1) * per_page`
to fit the engine platform's signed `int`. An unrepresentable offset returns
HTTP 400 `VALIDATION_ERROR` with detail
`page and per_page exceed maximum pagination offset`, before a storage call.
A representable page beyond the filtered total returns an empty `data` array
with the requested pagination and filtered `total`; no additional page cap is
imposed. Defaults and the `per_page <= 100` limit are unchanged.

```json
{
  "data": [
    {
      "id": "rule-001-1",
      "event_id": "rule-001-1",
      "rule_id": "rule-001",
      "rule_name": "SQL Injection - Union Select",
      "timestamp": "2026-06-26T14:27:25.000001Z",
      "src_ip": "10.0.0.3",
      "dst_ip": "10.0.0.2",
      "dst_port": 80,
      "protocol": "TCP",
      "severity": "high",
      "aggregated_count": 1,
      "mitre_tactic": "Initial Access",
      "mitre_technique_id": "T1190",
      "mitre_technique_name": "Exploit Public-Facing Application",
      "payload_preview": "GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n\r\n",
      "matched_keyword": "--"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 1
  }
}
```

### `GET /api/metrics`

Returns Prometheus text format with process counters, process-lifetime packet and alert rate gauges, current and high-water packet queue depth, loaded rules, alert counts, storage availability and health gauges, worker counters, rule match latency buckets, alert write latency buckets, and the latest capture heartbeat state when available. The `netsentry_packets_processed_per_second` and `netsentry_alerts_generated_per_second` gauges are process-lifetime averages derived from local counters, not sliding-window throughput guarantees.

R90-145 adds `netsentry_packets_completed_total` (counter): successful terminal
Worker processing, including any enabled lifecycle export. No-alert and fully
suppressed packets count once without a write; alert-producing packets count
only after writer success and any enabled Durable/Processed export succeeds.
Write/export errors and recovered panics do not count completion. The counter
appears at zero before work, including when stats are absent.

`netsentry_packets_processed_total` retains its pre-match increment, and the
existing received counter, processed-rate gauge and verbose-health JSON keep
their original fields and values. Use the new counter to observe terminal work;
write success is not a universal durability proof, and exporter acceptance does
not mean final receipt/fsync publication. Counts reset on process restart and
active Stats snapshots sample atomics independently. Aggregate count differences
do not establish offered-load loss or SLO acceptance.

Direct Worker/Stats/HTTP regression source is authored and compile-reviewed;
behavioral/race/CLI/full-suite/knowledge and acceptance execution remains
**not run; delegated by user**. See
[the implementation plan](plans/task-20261002-packet-completion-counter.md).

### `GET /api/rules`

For `payload_match`, `ip_blacklist` and `port_blacklist` configuration, an
explicit `any` member of `protocols` allows every protocol, including when
combined with `TCP`, `UDP` or `ICMP`. Protocol names ignore case and surrounding
spaces. Unsupported members still reject the rule even beside `any`. Blank
members are ignored; omitted/empty/blank-only lists are unrestricted, while a
blank mixed with named members does not broaden that named union. Other rule
gates remain active. Create/update/reload preserve the original protocol list
in canonical persisted configuration. Direct engine and file-backed HTTP
assertions are authored and compiled; behavioral execution remains delegated.

Returns the currently loaded rule snapshot in priority order.

```json
{
  "data": [
    {
      "id": "rule-001",
      "name": "SQL Injection - Union Select",
      "type": "payload_match",
      "severity": "high",
      "priority": 150,
      "enabled": true,
      "early_exit": false,
      "config": {
        "keywords": ["UNION SELECT", "DROP TABLE"],
        "case_insensitive": true,
        "protocols": ["TCP"],
        "ports": [80, 8080, 443],
        "direction": "dest",
        "depth": 4096,
        "offset": 0
      }
    }
  ]
}
```


### `GET /api/suppressions`

Returns the active suppression rules in insertion order. At startup, suppressions are loaded from `engine.suppressions_file` when configured.

```json
{
  "data": [
    {
      "id": "internal-subnet",
      "enabled": true,
      "rule_ids": ["rule-001"],
      "src_cidrs": ["10.0.0.0/24"],
      "dst_cidrs": [],
      "any_cidrs": []
    }
  ]
}
```

### `POST /api/suppressions`

Adds a suppression rule and immediately applies it to newly generated alerts. Enabled suppressions require at least one `src_cidrs`, `dst_cidrs`, or `any_cidrs` entry. When `engine.suppressions_file` is configured, successful creates are persisted to that JSON file before the in-memory snapshot is updated.

For enabled suppressions, omitted/null/empty `rule_ids` means all rules. A
nonempty list must contain at least one nonempty ID; `[""]` and `["", ""]`
are rejected rather than widening the scope. Empty entries mixed with valid IDs
are skipped; IDs otherwise match exactly without trimming or existence checks.
Disabled suppressions retain their existing compilation skip. Create/update
rejections use `400 VALIDATION_ERROR`; invalid file reload retains the existing
`500 INTERNAL_ERROR`. Rejection preserves the active rules/filter and does not
rewrite the suppression file. Structural standalone file load/save helpers do
not enforce this compiler requirement.

### `PUT /api/suppressions/{id}`

Replaces an existing suppression, persists the full suppressions file when configured, and atomically swaps the active in-memory filter. If the request body includes `id`, it must match the path ID.

### `DELETE /api/suppressions/{id}`

Deletes an existing suppression, persists the full suppressions file when configured, and returns `204 No Content`.

### `POST /api/suppressions/reload`

Reloads suppressions from `engine.suppressions_file` and atomically swaps the active in-memory filter when validation succeeds.

Within one suppression manager, the mutation lock serializes the authoritative
reload read through publication with create, update and delete. Reload errors
leave the prior active rules/filter intact; a missing file clears the active set
without creating a file. Reload I/O may delay other manager operations, including
List/Filter readers. External writers and standalone file helpers are outside
this coordination boundary; endpoint schemas and errors are unchanged.

```json
{
  "reloaded": 1
}
```

When configured, suppression create, update, and delete replace the canonical
file using an exact temporary write, mode preservation, file sync/close,
atomic rename, and parent-directory sync/close before reporting success. A
failure through rename returns `INTERNAL_ERROR` and preserves the prior file
and active filter. A parent-directory durability failure after rename returns
`SUPPRESSIONS_DURABILITY_UNCERTAIN`; the mutation is already applied to the
canonical file and active filter, but crash durability was not confirmed.

### `POST /api/rules`

Creates a rule, writes the canonical wrapped rules file, reloads the saved file, and atomically swaps the active rule snapshot. The request body is a single rule object using the schema below. Duplicate IDs return `RULE_ALREADY_EXISTS`.

Rule and suppression mutation bodies are limited to 1 MiB, reject unknown fields, and must contain exactly one JSON document.

### `PUT /api/rules/{id}`

Replaces an existing rule, persists the full rules file, reloads it, and atomically swaps the active snapshot. If the body includes `id`, it must match the path ID.

### `DELETE /api/rules/{id}`

Deletes an existing rule, persists the full rules file, reloads it, and returns `204 No Content`.

### `POST /api/rules/reload`

Reloads rules from `engine.rules_seed_file` and atomically swaps the active rule snapshot when validation succeeds.

```json
{
  "reloaded": 8
}
```

File-backed rule create, update, delete, and explicit reload requests are
serialized as complete in-process management transactions. A successful
response therefore agrees with both the canonical seed file and the active
snapshot after all earlier successful rule transactions. Packet matching stays
lock-free. Rule mutations require a complete temporary-file write, preserve the
existing file mode, sync and close the temporary file, atomically rename it,
and sync and close the containing directory before returning success. A failure
through rename leaves both the prior canonical bytes and active snapshot
unchanged. If rename commits but containing-directory durability cannot be
confirmed, the server loads the new canonical file into the active snapshot and
returns `500 RULES_DURABILITY_UNCERTAIN`; the mutation was applied and must not
be treated as a safe rejection. This does not coordinate a second NetSentry
process writing the same file or claim portability beyond the checked local
filesystem lifecycle.

Current limitations:

- Alert pagination, the stable list envelope, exact-match filters, time range filters, MITRE filters, matched-keyword substring filtering, and minimum aggregate-count filtering exist. The SQLite-backed store pins rule, severity, source, and destination exact matches to binary comparison regardless of a compatible database's declared column collation, while protocol and MITRE filters remain case-insensitive. It applies those filters and pagination in SQL, with indexes for common exact/range filters; matched-keyword substring filtering remains a regular SQL substring predicate.
- Alert storage is SQLite-backed with JSONL recovery-log replay, startup TTL pruning, old daily shard file cleanup, and sticky emergency mode for disk-full/read-only/I/O failures. Emergency mode remains sticky until restart or a successful authenticated operator recovery request. When `engine.db_shard_daily` is enabled, alert writes use each alert timestamp to select `netsentry-YYYY-MM-DD.db`, alert queries scan matching shards and apply the same filters, ordering, and pagination across shards, and health and metrics alert counts also sum matching shard files.
- Validation, unsupported method, and internal API errors use the unified error envelope.
- Rules can be listed, created, replaced, deleted, persisted to the configured seed file, and reloaded from disk. File-backed mutations and explicit reload are serialized within one API server; cross-process writers remain outside that boundary.
- A rule mutation whose rename commits but parent-directory durability cannot
  be confirmed returns `RULES_DURABILITY_UNCERTAIN`; canonical disk and active
  memory contain the committed mutation even though the response is an error.
- Optional PSK Bearer authentication protects modifying rule and suppression endpoints when `engine.api_auth_enabled` is true.
- The HTTP listener has explicit read/header/write/idle timeouts, a 16 KiB header limit, and loopback-only defaults.
- Non-GET API requests emit structured zap audit logs with request ID, method, path, status, authorization outcome, target, remote address, and duration.
  Audit status retains the first committed final response, including terminal
  101 Switching Protocols; before final commitment, other informational 1xx
  headers are forwarded without committing the logged status. A body write commits implicit 200, and a handler
  that never commits a final header or body logs the normal default 200. Later
  headers cannot replace the final audit status or its existing status-derived
  authorization indicator. Endpoint authorization policy and audit fields are
  unchanged. Direct real-HTTP status/log regressions are authored and compiled
  only; execution remains delegated to the user's test department.
- Optional pprof runs on a separate localhost-only server when `engine.pprof_enabled` is true.
- Suppressions load from `engine.suppressions_file` at startup; create, update, delete, and reload operations persist or reload that file before swapping the active in-memory filter.
- Payload previews are redacted before SQLite writes when `engine.redact_sensitive_fields` is true; current redaction covers Authorization, Cookie, Set-Cookie, password, and token patterns.
- Complete quoted JSON password/token values recognize escape pairs, including
  escaped quotes and backslashes, while preserving surrounding formatting.
  Keys retain literal case-insensitive matching. Redaction remains best effort
  on bounded previews; escaped keys, extra sensitive fields and malformed or
  truncated values have no new sanitization guarantee. Existing configuration,
  endpoint fields, header/pair behavior and replacement marker are unchanged.

---

## `POST /api/storage/recovery`

Starts one synchronous, operator-triggered recovery attempt for a store in
sticky emergency mode. The endpoint does not perform cleanup or repair and no
timer, health read, or ordinary write can trigger recovery.

- API authentication must be enabled. If it is disabled, the endpoint returns
  `409 STORAGE_RECOVERY_AUTH_REQUIRED`. Once enabled, a valid Bearer token is
  mandatory even on loopback; missing or invalid credentials return `401`
  before storage inspection.
- Only `storage.status=emergency` is accepted. A healthy or degraded store
  returns `409 STORAGE_RECOVERY_NOT_NEEDED`; a second request during an active
  attempt returns `409 STORAGE_RECOVERY_IN_PROGRESS`. Neither conflict touches
  SQLite or the recovery log.
- The request owns one bounded synchronous attempt. Client cancellation or
  server shutdown cancels the attempt; no detached retry continues afterward.
- `200` means read-only preflight, writable proof/replay, complete recovery-log
  truncation, and the transition to healthy all succeeded. The response is
  `{"status":"ok","phase":"complete"}`.
- A preservation-safe preflight rejection or writable replay failure returns
  `503 STORAGE_RECOVERY_FAILED`, leaves the store in emergency, retains the
  recovery log, and reports only stable `phase` and `writable_attempted`
  details. The `X-NetSentry-Recovery-Phase` response header carries the phase
  for success, conflict, and recovery failure responses.
- Audit logging records the request metadata, authorization result, duration,
  storage target, and response phase. It never records the underlying recovery
  error, recovery-log content, credentials, or private filesystem paths.

Verbose health exposes `recovering`, attempt start time, current phase, and the
last terminal recovery result/time when available. Polling health is
observational and cannot trigger or advance recovery. During the exclusive
recovery window it does not enter the ordinary database count path.

---

## Planned v0.1.0 API Contract

List responses use:

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 234
  }
}
```

Error responses use the envelope below. Unsupported methods return `METHOD_NOT_ALLOWED`
and include an `Allow` header listing the supported methods for that endpoint.

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [],
    "request_id": "req_xxx"
  }
}
```

Planned endpoints:

| Endpoint | Status | Notes |
| --- | --- | --- |
| `GET /api/health` | partial | Minimal and verbose component snapshot responses exist. |
| `GET /api/health?verbose=true` | partial | Capture heartbeat freshness, queue depth, rule count, storage status including emergency mode, storage available bytes, and throughput counters exist. |
| `GET /api/alerts` | partial | SQLite-backed paginated list with exact-match, time range, MITRE, matched-keyword, and aggregate-count filters exists; daily-shard mode queries across matching shard files. |
| `GET /api/metrics` | partial | Prometheus text output exists for process counters, process-lifetime packet and alert rate gauges, rule match and alert write latency buckets, current/high-water queue depth, rule/alert/storage gauges, worker counters, and capture heartbeat gauges. |
| `POST /api/storage/recovery` | implemented | Authenticated explicit recovery from sticky emergency mode, with preservation-safe preflight, serialized replay/write probe, stable conflicts, and phase reporting. |
| `GET /api/rules` | partial | Current rule snapshot listing exists. |
| `POST /api/rules` | partial | Creates and persists one rule; optional PSK auth exists. |
| `PUT /api/rules/{id}` | partial | Replaces and persists one rule; optional PSK auth exists. |
| `DELETE /api/rules/{id}` | partial | Deletes and persists one rule; optional PSK auth exists. |
| `POST /api/rules/reload` | partial | Hot reload from `engine.rules_seed_file` exists; optional PSK auth exists. |
| `GET/POST/PUT/DELETE /api/suppressions` | partial | Suppression listing, create, update, delete, and file reload exist; filters newly generated alerts and persists mutations to `engine.suppressions_file`. |
| `POST /api/suppressions/reload` | partial | Hot reload from `engine.suppressions_file` exists; optional PSK auth exists. |
| `GET /debug/pprof/*` | partial | Optional separate localhost-only server when `engine.pprof_enabled` is true; not public API. |

Authentication: modifying rule and suppression endpoints require `Authorization: Bearer <token>` when `engine.api_auth_enabled` is true. The token is configured with `engine.api_auth_token`.

MITRE mapping: the v0.1 alert schema stores at most one technique per rule. Rule reload validates supported technique ID/tactic/name tuples against the versioned in-code catalog and rejects multiple mappings instead of silently dropping them.

---

## Rule JSON Schema

The canonical seed rule format is:

```json
{
  "rules": [
    {
      "id": "rule-001",
      "name": "SQL Injection Detection",
      "type": "payload_match",
      "severity": "high",
      "priority": 150,
      "enabled": true,
      "early_exit": false,
      "config": {
        "keywords": ["UNION SELECT", "DROP TABLE"],
        "case_insensitive": true,
        "protocols": ["TCP"],
        "ports": [80, 8080, 443],
        "direction": "dest",
        "depth": 4096,
        "offset": 0
      },
      "mitre_techniques": [
        {
          "tactic": "Initial Access",
          "technique_id": "T1190",
          "technique_name": "Exploit Public-Facing Application"
        }
      ],
      "description": "Detect SQL injection patterns in cleartext payloads"
    }
  ]
}
```

The loader still accepts the previous top-level array and legacy `payload_match` / `ip_blacklist` fields during migration.
