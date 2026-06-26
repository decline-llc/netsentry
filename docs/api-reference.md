# NetSentry — API Reference

> **Version**: v0.1.0 (planned)
> **Base URL**: `http://localhost:8080`
> **Authentication**: `Authorization: Bearer <token>` (PSK)

---

## Authentication

All endpoints except `/api/health` and `/api/metrics` require a Pre-Shared Key (PSK) token.

```
Authorization: Bearer <your-token>
```

Configure the token in `config.yaml`:

```yaml
engine:
  api_auth_token: "${NETSENTRY_API_TOKEN}"
  api_auth_enabled: true
```

**Unauthenticated requests** return:
```json
HTTP 401
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid Authorization header",
    "request_id": "req_a1b2c3d4"
  }
}
```

---

## Unified Response Formats

### Success (list)

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

### Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {"field": "severity", "reason": "must be one of: low, medium, high, critical"}
    ],
    "request_id": "req_a1b2c3d4"
  }
}
```

### HTTP Status Codes

| Code | Scenario |
|------|----------|
| 200 | Query success |
| 201 | Resource created (`POST /api/rules`) |
| 204 | Delete success — no response body |
| 400 | Request validation failed |
| 401 | Unauthenticated |
| 404 | Resource not found |
| 409 | Conflict (e.g. duplicate `rule_id`) |
| 429 | Rate limit exceeded |
| 500 | Internal error |

---

## Pagination

All list endpoints accept:

| Parameter | Type | Default | Max |
|-----------|------|---------|-----|
| `page` | int | 1 | — |
| `per_page` | int | 20 | 100 |

---

## Endpoints

### Health

#### `GET /api/health`

Lightweight liveness probe (no auth required).

```json
{"status": "ok"}
```

#### `GET /api/health?verbose=true`

Full health snapshot (no auth required).

```json
{
  "status": "ok",
  "uptime_seconds": 3600,
  "version": "0.1.0",
  "data_freshness_seconds": 2.3,
  "components": {
    "capture": {
      "status": "connected",
      "last_heartbeat_ago_ms": 1200,
      "session_id": "a1b2c3d4",
      "restarts_total": 0,
      "parse_errors_total": 3,
      "packets_sent_total": 98765
    },
    "engine": {
      "status": "ok",
      "goroutines": 8,
      "heap_alloc_mb": 45,
      "channel_depth": 342,
      "channel_capacity": 10000,
      "channel_dropped_total": 0
    },
    "sqlite": {
      "status": "ok",
      "wal_size_mb": 2,
      "write_latency_p99_ms": 5,
      "current_db": "alerts_2026-06-25.db"
    }
  },
  "throughput": {
    "packets_total": 98765,
    "packets_dropped": 0,
    "alerts_total": 67,
    "alerts_by_severity": {"critical": 2, "high": 15, "medium": 30, "low": 20}
  }
}
```

When `status` is `"degraded"`, check `components` for the failing subsystem.

---

### Metrics

#### `GET /api/metrics`

Prometheus text format (no auth required). Scrape by Prometheus or `curl`.

```
# HELP netsentry_packets_received_total Total packets received from C capture module
# TYPE netsentry_packets_received_total counter
netsentry_packets_received_total 98765
...
```

See [architecture.md](architecture.md) §9 for the full metric list.

---

### Alerts

#### `GET /api/alerts`

List alerts with filtering and pagination.

**Query parameters**:

| Parameter | Type | Description |
|-----------|------|-------------|
| `date` | string | Query a specific day's DB, e.g. `2026-06-25`. Overrides `start_time`/`end_time`. |
| `start_time` | RFC3339 | Start of time range (must be same UTC day as `end_time` in v0.1.0) |
| `end_time` | RFC3339 | End of time range |
| `severity` | string | Comma-separated: `low,medium,high,critical` |
| `src_ip` | string | CIDR or single IP, e.g. `10.0.0.0/24` |
| `dst_ip` | string | CIDR or single IP |
| `dst_port` | string | Comma-separated ports, e.g. `80,443` |
| `rule_id` | string | Exact rule ID |
| `mitre_tactic` | string | e.g. `Discovery` |
| `mitre_technique_id` | string | e.g. `T1046` |
| `aggregated_count_min` | int | Minimum aggregated count |
| `sort_by` | string | `timestamp` (default), `severity`, `aggregated_count` |
| `sort_order` | string | `asc` or `desc` (default) |
| `page` | int | Page number (default 1) |
| `per_page` | int | Rows per page (default 20, max 100) |

Multiple values for a parameter are OR-combined; different parameters are AND-combined.

> **v0.1.0 cross-day limitation**: If `start_time` and `end_time` span multiple UTC days, the API returns HTTP 400 with code `CROSS_DAY_QUERY_UNSUPPORTED`. Query each day separately or use the `date` parameter.

**Response**:

```json
{
  "data": [
    {
      "id": "alert-0001",
      "event_id": "evt_a1b2c3d4e5f6",
      "rule_id": "rule-001",
      "rule_name": "SQL Injection Detection",
      "timestamp": "2026-06-25T10:30:00Z",
      "src_ip": "10.0.0.99",
      "dst_ip": "192.168.1.10",
      "dst_port": 80,
      "protocol": "TCP",
      "severity": "high",
      "aggregated_count": 12,
      "first_seen": "2026-06-25T10:29:50Z",
      "last_seen": "2026-06-25T10:30:00Z",
      "mitre_tactic": "Initial Access",
      "mitre_technique_id": "T1190",
      "mitre_technique_name": "Exploit Public-Facing Application",
      "matched_keyword": "UNION SELECT",
      "payload_preview": "GET /search?q=UNION+SELECT..."
    }
  ],
  "pagination": {"page": 1, "per_page": 20, "total": 67}
}
```

#### `GET /api/alerts/:id`

Single alert detail. Includes `raw_payload` field (full JSON context, also redacted).

#### `POST /api/alerts/batch-delete`

Delete multiple alerts by ID array.

**Request body**:

```json
{"ids": ["alert-0001", "alert-0002"]}
```

**Response**: `204 No Content`

---

### Traffic Stats

#### `GET /api/stats`

Traffic statistics snapshot.

```json
{
  "packets_received_total": 98765,
  "packets_dropped_total": 0,
  "alerts_total": 67,
  "alerts_by_severity": {"critical": 2, "high": 15, "medium": 30, "low": 20},
  "rules_loaded": 12,
  "rules_enabled": 10,
  "uptime_seconds": 3600
}
```

---

### Rules

#### `GET /api/rules`

List all rules (paginated).

#### `POST /api/rules`

Create a rule. Returns `201` with the created rule.

**Request body** (full schema):

```json
{
  "id": "rule-001",
  "name": "SQL Injection Detection",
  "type": "payload_match",
  "severity": "high",
  "priority": 150,
  "enabled": true,
  "early_exit": false,
  "config": {
    "keywords": ["UNION SELECT", "DROP TABLE", "OR 1=1"],
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
  "description": "Detect SQL injection patterns in HTTP traffic"
}
```

**Rule types**:

| Type | `config` fields |
|------|----------------|
| `payload_match` | `keywords`, `case_insensitive`, `protocols`, `ports`, `direction`, `depth`, `offset` |
| `ip_blacklist` | `ips` (IP or CIDR list), `direction` (`src`/`dst`/`any`) |
| `port_blacklist` | `ports`, `protocols` |
| `frequency_threshold` | `threshold`, `window_secs`, `group_by` |

**Field constraints**:

| Field | Constraint |
|-------|-----------|
| `id` | Must match `^rule-[0-9]{3,}$` |
| `severity` | One of `low`, `medium`, `high`, `critical` |
| `priority` | Integer 0–1000 (default 100) |
| `early_exit` | Only meaningful when `severity == critical` |
| `direction` | `src`, `dest`, or `any` |
| `depth` | 1–65535 (default 4096) |

Rate limit: 5 write operations/second on rule endpoints.

#### `GET /api/rules/:id`

Get a single rule by ID.

#### `PUT /api/rules/:id`

Full update of a rule (replaces all fields).

#### `PATCH /api/rules/:id`

Partial update (only supplied fields are changed).

#### `DELETE /api/rules/:id`

Delete a rule. Returns `204 No Content`.

#### `POST /api/rules/batch`

Bulk import rules.

**Request body**:

```json
{"rules": [ ... ]}
```

#### `PATCH /api/rules/batch`

Bulk enable/disable rules.

**Request body**:

```json
{"ids": ["rule-001", "rule-002"], "enabled": false}
```

---

### Suppressions

Suppression rules silence alerts matching specific conditions (IP range, time window).

#### `GET /api/suppressions`

List all suppression rules (paginated).

#### `POST /api/suppressions`

Create a suppression rule.

**Request body**:

```json
{
  "id": "suppress-001",
  "rule_id": "rule-001",
  "src_ip": "10.0.0.0/24",
  "dst_ip": null,
  "reason": "Internal vulnerability scanner — Nessus weekly scan",
  "expires_at": "2026-07-01T00:00:00Z",
  "enabled": true
}
```

`rule_id: null` suppresses all rules matching the given IP. `expires_at: null` means never expires.

#### `GET /api/suppressions/:id`

Get a single suppression rule.

#### `PUT /api/suppressions/:id` / `PATCH /api/suppressions/:id`

Full / partial update.

#### `DELETE /api/suppressions/:id`

Delete a suppression rule. Returns `204 No Content`.

---

### Debug (pprof)

> **Security**: pprof endpoints are served on a separate port (`127.0.0.1:6060`) in production and require PSK authentication. Never expose to the public internet.

| Endpoint | Description |
|----------|-------------|
| `GET /debug/pprof/` | Index of available profiles |
| `GET /debug/pprof/profile?seconds=30` | 30 s CPU profile |
| `GET /debug/pprof/heap` | Heap memory profile |
| `GET /debug/pprof/goroutine?debug=2` | Goroutine stack dump |

```bash
# Example: capture 30 s CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## Rate Limits

| Endpoint group | Limit |
|----------------|-------|
| Rule write operations (`POST`/`PUT`/`PATCH`/`DELETE /api/rules*`) | 5 req/s |
| Query endpoints (`GET /api/alerts`, `GET /api/rules`) | 30 req/s |

Exceeded requests return `HTTP 429`.

---

## CORS

Allowed origins are configured in `config.yaml`:

```yaml
engine:
  cors_allowed_origins: ["http://localhost:3000"]
```

Allowed methods: `GET, POST, PUT, PATCH, DELETE`. Wildcard `*` is not set.

---

## Sensitive Data Redaction

`payload_preview` fields in API responses are automatically redacted for common sensitive patterns (cookies, authorization headers, passwords). Configure in `config.yaml`:

```yaml
engine:
  redact_sensitive_fields: true
  redact_patterns:
    - 'Cookie:\s*[^\r\n]+'
    - 'Authorization:\s*[^\r\n]+'
    - 'password=[^&\s]+'
```

Matched substrings are replaced with `[REDACTED]`.

---

## MITRE ATT&CK Filter Examples

```bash
# All Discovery alerts
GET /api/alerts?mitre_tactic=Discovery

# Specific technique
GET /api/alerts?mitre_technique_id=T1046

# Critical Initial Access alerts
GET /api/alerts?mitre_tactic=Initial+Access&severity=critical
```

**Supported tactics**: `Initial Access`, `Execution`, `Persistence`, `Privilege Escalation`, `Defense Evasion`, `Credential Access`, `Discovery`, `Lateral Movement`, `Collection`, `Command and Control`, `Exfiltration`, `Impact`.
