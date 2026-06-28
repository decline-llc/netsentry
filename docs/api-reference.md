# NetSentry API Reference

> **Status**: development snapshot. This document separates endpoints implemented today from the planned v0.1.0 API contract.

---

## Implemented Today

Base URL: `http://localhost:8080`

### `GET /api/health`

Returns a minimal liveness response. Query parameters are currently ignored.

```json
{
  "status": "ok",
  "alerts": 5
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

Returns Prometheus text format with basic process counters and gauges.

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

- Alert pagination, the stable list envelope, and basic exact-match filters exist; advanced query features are still pending.
- Alert storage is SQLite-backed with startup TTL pruning and old daily shard file cleanup; optional daily shard pathing exists, but runtime cross-day rotation and cross-day querying are not implemented yet.
- Validation and internal API errors use the unified error envelope.
- Rules can be listed, created, replaced, deleted, persisted to the configured seed file, and reloaded from disk.
- Optional PSK Bearer authentication protects modifying rule endpoints when `engine.api_auth_enabled` is true.
- No payload redaction yet.

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
| `GET /api/health` | partial | Minimal response exists; verbose component snapshot pending. |
| `GET /api/health?verbose=true` | planned | Capture heartbeat, channel depth, storage and throughput details. |
| `GET /api/alerts` | partial | SQLite-backed paginated list with basic exact-match filters exists; advanced query features pending. |
| `GET /api/metrics` | partial | Basic Prometheus text format exists; fuller coverage pending. |
| `GET /api/rules` | partial | Current rule snapshot listing exists. |
| `POST /api/rules` | partial | Creates and persists one rule; optional PSK auth exists. |
| `PUT /api/rules/{id}` | partial | Replaces and persists one rule; optional PSK auth exists. |
| `DELETE /api/rules/{id}` | partial | Deletes and persists one rule; optional PSK auth exists. |
| `POST /api/rules/reload` | partial | Hot reload from `engine.rules_seed_file` exists; optional PSK auth exists. |
| `GET/POST /api/suppressions` | planned | CIDR suppressor component exists; API wiring is pending. |
| `GET /debug/pprof/*` | planned | Separate localhost server, not public API. |

Authentication: modifying rule endpoints require `Authorization: Bearer <token>` when `engine.api_auth_enabled` is true. The token is configured with `engine.api_auth_token`.

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
