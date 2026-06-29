#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
ENGINE_PID=""

cleanup() {
    if [[ -n "${ENGINE_PID}" ]] && kill -0 "${ENGINE_PID}" 2>/dev/null; then
        kill "${ENGINE_PID}" 2>/dev/null || true
        wait "${ENGINE_PID}" 2>/dev/null || true
    fi
    rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

PORT="$(python3 - <<'PY'
import socket

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
    s.bind(("127.0.0.1", 0))
    print(s.getsockname()[1])
PY
)"

UDS_PATH="${TMP_DIR}/netsentry.sock"
DB_PATH="${TMP_DIR}/netsentry.db"
CONFIG_PATH="${TMP_DIR}/config.yaml"
PCAP_PATH="/tmp/netsentry_test.pcap"

cat >"${CONFIG_PATH}" <<EOF_CFG
capture:
  mode: "offline"
  offline_file: "${PCAP_PATH}"
  payload_preview_len: 4096
  uds_socket_path: "${UDS_PATH}"
  uds_socket_mode: "0600"
  heartbeat_interval: 5

engine:
  uds_socket_path: "${UDS_PATH}"
  channel_buffer_size: 10000
  worker_count: 1
  db_dir: "${TMP_DIR}"
  db_path: "${DB_PATH}"
  db_shard_daily: false
  db_journal_mode: "WAL"
  db_busy_timeout: 5000
  rules_seed_file: "${ROOT_DIR}/configs/rules.json"
  api_port: ${PORT}
  cors_allowed_origins: ["http://localhost:3000"]
  alert_aggregation_window: 60
  alert_aggregation_max_count: 100
  alert_retention_days: 7
  api_auth_enabled: false
  api_auth_token: ""
  redact_sensitive_fields: true
  health_freshness_limit_seconds: 30
  pprof_enabled: false
  pprof_addr: "127.0.0.1:6060"

logging:
  level: "info"
  format: "json"
  engine_log: "${TMP_DIR}/engine.log"
EOF_CFG

cd "${ROOT_DIR}"

python3 scripts/gen_test_pcap.py >/dev/null

bin/netsentry-engine -config "${CONFIG_PATH}" >"${TMP_DIR}/engine.stdout" 2>"${TMP_DIR}/engine.stderr" &
ENGINE_PID="$!"

for _ in $(seq 1 50); do
    if curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
        break
    fi
    if ! kill -0 "${ENGINE_PID}" 2>/dev/null; then
        echo "[e2e] engine exited before health check passed" >&2
        cat "${TMP_DIR}/engine.stderr" >&2 || true
        exit 1
    fi
    sleep 0.1
done

curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null

bin/netsentry-capture -r "${PCAP_PATH}" -s "${UDS_PATH}" -c 5 >"${TMP_DIR}/capture.stdout" 2>"${TMP_DIR}/capture.stderr"

ALERTS_JSON="${TMP_DIR}/alerts.json"
HEALTH_JSON="${TMP_DIR}/health.json"

for _ in $(seq 1 50); do
    curl -fsS "http://127.0.0.1:${PORT}/api/alerts" >"${ALERTS_JSON}"
    if python3 - "${ALERTS_JSON}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    payload = json.load(f)
sys.exit(0 if payload.get("pagination", {}).get("total") == 5 else 1)
PY
    then
        break
    fi
    sleep 0.1
done

curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${HEALTH_JSON}"

python3 - "${ALERTS_JSON}" "${HEALTH_JSON}" <<'PY'
import json
import sys

alerts_path, health_path = sys.argv[1], sys.argv[2]
with open(alerts_path, "r", encoding="utf-8") as f:
    alerts = json.load(f)
with open(health_path, "r", encoding="utf-8") as f:
    health = json.load(f)

errors = []
pagination = alerts.get("pagination", {})
data = alerts.get("data", [])
if pagination.get("total") != 5:
    errors.append(f"expected 5 total alerts, got {pagination.get('total')!r}")
if len(data) != 5:
    errors.append(f"expected 5 alerts on first page, got {len(data)}")

rule_ids = {alert.get("rule_id") for alert in data}
expected_rules = {"rule-001", "rule-004", "rule-005", "rule-007", "rule-008"}
missing = expected_rules - rule_ids
if missing:
    errors.append(f"missing expected rule ids: {sorted(missing)}")

engine = health.get("engine", {})
throughput = health.get("throughput", {})
capture = health.get("capture", {})
if engine.get("rules_loaded") != 8:
    errors.append(f"expected 8 loaded rules, got {engine.get('rules_loaded')!r}")
if throughput.get("packets_received") != 6:
    errors.append(f"expected 6 packets_received, got {throughput.get('packets_received')!r}")
if throughput.get("packets_processed") != 6:
    errors.append(f"expected 6 packets_processed, got {throughput.get('packets_processed')!r}")
if throughput.get("alerts_generated") != 5:
    errors.append(f"expected 5 alerts_generated, got {throughput.get('alerts_generated')!r}")
if throughput.get("decode_errors") != 0:
    errors.append(f"expected 0 decode_errors, got {throughput.get('decode_errors')!r}")
if capture.get("heartbeat", {}).get("sent") != 6:
    errors.append(f"expected final heartbeat sent=6, got {capture.get('heartbeat', {}).get('sent')!r}")

if errors:
    for err in errors:
        print(f"[e2e] {err}", file=sys.stderr)
    sys.exit(1)

print("[e2e] ok: 6 packets processed, 5 alerts generated, 8 rules loaded")
PY
