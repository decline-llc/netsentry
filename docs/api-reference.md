# NetSentry API Reference

> **Status**: development snapshot. This document separates endpoints implemented today from the planned v0.1.0 API contract.

---

## Implemented Today

Base URL: `http://localhost:8080`

### `GET /api/health`

Returns a minimal liveness response by default.

```json
{
  "status": "ok",
  "alerts": 5
}
```

With `verbose=true`, returns capture heartbeat status, engine queue/rule counts, storage status, and throughput counters. Capture status is `unknown` before the first heartbeat, `ok` while the latest heartbeat is within `engine.health_freshness_limit_seconds`, and `stale` after that limit. Storage status is `ok` by default and becomes `degraded` after SQLite write/query errors until a later successful write or full alert list query clears it.

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
    "alerts": 5
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
| `rule_id` | Exact rule ID match. |
| `severity` | One of `low`, `medium`, `high`, `critical`. |
| `src_ip` | Exact source IP match. |
| `dst_ip` | Exact destination IP match. |
| `protocol` | Exact protocol match; compared case-insensitively. |
| `dst_port` | Destination port from `0` to `65535`. |
| `since` | Include alerts whose `last_seen` is greater than or equal to this RFC3339 timestamp. |
| `until` | Include alerts whose `last_seen` is less than or equal to this RFC3339 timestamp. |
| `mitre_tactic` | MITRE tactic match; compared case-insensitively. |
| `mitre_technique_id` | MITRE technique ID match; compared case-insensitively. |
| `matched_keyword` | Case-insensitive substring match against the recorded matched keyword. |
| `min_count` | Minimum `aggregated_count`; must be a positive integer. |

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

Returns Prometheus text format with process counters, packet queue depth, loaded rules, alert counts, storage availability and health gauges, worker counters, rule match latency buckets, alert write latency buckets, and the latest capture heartbeat state when available.

### `GET /api/rules`

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

### `PUT /api/suppressions/{id}`

Replaces an existing suppression, persists the full suppressions file when configured, and atomically swaps the active in-memory filter. If the request body includes `id`, it must match the path ID.

### `DELETE /api/suppressions/{id}`

Deletes an existing suppression, persists the full suppressions file when configured, and returns `204 No Content`.

### `POST /api/suppressions/reload`

Reloads suppressions from `engine.suppressions_file` and atomically swaps the active in-memory filter when validation succeeds.

```json
{
  "reloaded": 1
}
```

### `POST /api/rules`

Creates a rule, writes the canonical wrapped rules file, reloads the saved file, and atomically swaps the active rule snapshot. The request body is a single rule object using the schema below. Duplicate IDs return `RULE_ALREADY_EXISTS`.

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

Current limitations:

- Alert pagination, the stable list envelope, exact-match filters, time range filters, MITRE filters, matched-keyword substring filtering, and minimum aggregate-count filtering exist. Filtering currently runs after the SQLite list query, which is capped to the most recent 1000 aggregated rows.
- Alert storage is SQLite-backed with startup TTL pruning and old daily shard file cleanup; optional daily shard pathing exists, but runtime cross-day rotation and cross-day querying are not implemented yet.
- Validation and internal API errors use the unified error envelope.
- Rules can be listed, created, replaced, deleted, persisted to the configured seed file, and reloaded from disk.
- Optional PSK Bearer authentication protects modifying rule and suppression endpoints when `engine.api_auth_enabled` is true.
- Non-GET API requests emit structured zap audit logs with request ID, method, path, status, authorization outcome, target, remote address, and duration.
- Optional pprof runs on a separate localhost-only server when `engine.pprof_enabled` is true.
- Suppressions load from `engine.suppressions_file` at startup; create, update, delete, and reload operations persist or reload that file before swapping the active in-memory filter.
- Payload previews are redacted before SQLite writes when `engine.redact_sensitive_fields` is true; current redaction covers Authorization, Cookie, Set-Cookie, password, and token patterns.

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

Error responses use:

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
| `GET /api/health` | partial | Minimal and verbose component snapshot responses exist; deeper dependency checks pending. |
| `GET /api/health?verbose=true` | partial | Capture heartbeat freshness, queue depth, rule count, storage status, and throughput counters exist. |
| `GET /api/alerts` | partial | SQLite-backed paginated list with exact-match, time range, MITRE, matched-keyword, and aggregate-count filters exists; cross-day querying is pending. |
| `GET /api/metrics` | partial | Prometheus text output exists for process counters, rule match and alert write latency buckets, queue/rule/alert/storage gauges, worker counters, and capture heartbeat gauges. |
| `GET /api/rules` | partial | Current rule snapshot listing exists. |
| `POST /api/rules` | partial | Creates and persists one rule; optional PSK auth exists. |
| `PUT /api/rules/{id}` | partial | Replaces and persists one rule; optional PSK auth exists. |
| `DELETE /api/rules/{id}` | partial | Deletes and persists one rule; optional PSK auth exists. |
| `POST /api/rules/reload` | partial | Hot reload from `engine.rules_seed_file` exists; optional PSK auth exists. |
| `GET/POST/PUT/DELETE /api/suppressions` | partial | Suppression listing, create, update, delete, and file reload exist; filters newly generated alerts and persists mutations to `engine.suppressions_file`. |
| `POST /api/suppressions/reload` | partial | Hot reload from `engine.suppressions_file` exists; optional PSK auth exists. |
| `GET /debug/pprof/*` | partial | Optional separate localhost-only server when `engine.pprof_enabled` is true; not public API. |

Authentication: modifying rule and suppression endpoints require `Authorization: Bearer <token>` when `engine.api_auth_enabled` is true. The token is configured with `engine.api_auth_token`.

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
