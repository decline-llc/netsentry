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

```json
{
  "alerts": [
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
  "total": 1
}
```

### `GET /api/metrics`

Returns Prometheus text format with basic process counters and gauges.

Current limitations:

- Alert pagination and the stable list envelope exist; filtering is still pending.
- Alert storage is SQLite-backed with startup TTL pruning and old daily shard file cleanup; optional daily shard pathing exists, but runtime cross-day rotation and cross-day querying are not implemented yet.
- Validation and internal API errors use the unified error envelope.
- No authentication yet.
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
| `GET /api/alerts` | partial | SQLite-backed paginated list exists; filtering pending. |
| `GET /api/metrics` | partial | Basic Prometheus text format exists; fuller coverage pending. |
| `GET/POST /api/rules` | planned | Rule listing and hot reload. |
| `GET/PUT/PATCH/DELETE /api/rules/:id` | planned | Rule CRUD. |
| `GET/POST /api/suppressions` | planned | CIDR suppressor component exists; API wiring is pending. |
| `GET /debug/pprof/*` | planned | Separate localhost server, not public API. |

Planned authentication: PSK Bearer token for modifying endpoints, configured under `engine.api_auth_enabled` and `engine.api_auth_token`.

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
