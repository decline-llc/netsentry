#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
ENGINE_PID=""
MAX_RSS_KB=0

PCAP_CORPUS="${PCAP_CORPUS:-}"
WAIT_ATTEMPTS="${CORPUS_WAIT_ATTEMPTS:-1200}"
OUTPUT_DIR="${CORPUS_OUTPUT_DIR:-${ROOT_DIR}/docs/evidence/local}"
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
SUMMARY_JSON="${OUTPUT_DIR}/corpus-pressure-${RUN_ID}.json"
SUMMARY_MD="${OUTPUT_DIR}/corpus-pressure-${RUN_ID}.md"
INCLUDE_PATHS="${NETSENTRY_EVIDENCE_INCLUDE_PATHS:-0}"

cleanup() {
    if [[ -n "${ENGINE_PID}" ]] && kill -0 "${ENGINE_PID}" 2>/dev/null; then
        kill "${ENGINE_PID}" 2>/dev/null || true
        wait "${ENGINE_PID}" 2>/dev/null || true
    fi
    rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

sample_engine_memory() {
    if [[ -r "/proc/${ENGINE_PID}/status" ]]; then
        local rss
        rss="$(awk '/^VmRSS:/ {print $2}' "/proc/${ENGINE_PID}/status")"
        if [[ "${rss:-0}" =~ ^[0-9]+$ ]] && (( rss > MAX_RSS_KB )); then
            MAX_RSS_KB="${rss}"
        fi
    fi
}

usage() {
    cat >&2 <<'EOF_USAGE'
Usage:
  PCAP_CORPUS=/path/to/file-or-directory make e2e-corpus-pressure

Optional:
  CORPUS_OUTPUT_DIR=/path/to/evidence
  CORPUS_WAIT_ATTEMPTS=1200
  NETSENTRY_EVIDENCE_INCLUDE_PATHS=1

The corpus must stay local and sanitized before sharing. The script writes
JSON and Markdown evidence summaries; it does not commit pcap files. Corpus
paths are redacted by default; set NETSENTRY_EVIDENCE_INCLUDE_PATHS=1 only for
private local debugging evidence.
EOF_USAGE
}

if [[ -z "${PCAP_CORPUS}" ]]; then
    usage
    exit 2
fi

if [[ ! -e "${PCAP_CORPUS}" ]]; then
    echo "[corpus-pressure] PCAP_CORPUS does not exist: ${PCAP_CORPUS}" >&2
    exit 2
fi

PCAPS=()
if [[ -f "${PCAP_CORPUS}" ]]; then
    PCAPS+=("${PCAP_CORPUS}")
else
    while IFS= read -r -d '' pcap; do
        PCAPS+=("${pcap}")
    done < <(find "${PCAP_CORPUS}" -type f \( -iname '*.pcap' -o -iname '*.pcapng' \) -print0 | sort -z)
fi

if [[ "${#PCAPS[@]}" -eq 0 ]]; then
    echo "[corpus-pressure] no .pcap or .pcapng files found under ${PCAP_CORPUS}" >&2
    exit 2
fi

mkdir -p "${OUTPUT_DIR}"

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
RUN_JSONL="${TMP_DIR}/runs.jsonl"
HEALTH_JSON="${TMP_DIR}/health.json"
ALERTS_JSON="${TMP_DIR}/alerts.json"
METRICS_TXT="${TMP_DIR}/metrics.txt"

cat >"${CONFIG_PATH}" <<EOF_CFG
capture:
  mode: "offline"
  offline_file: ""
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
  suppressions_file: "${ROOT_DIR}/configs/suppressions.json"
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

cd "${ROOT_DIR}"

bin/netsentry-engine -config "${CONFIG_PATH}" >"${TMP_DIR}/engine.stdout" 2>"${TMP_DIR}/engine.stderr" &
ENGINE_PID="$!"

for _ in $(seq 1 50); do
    if curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
        break
    fi
    if ! kill -0 "${ENGINE_PID}" 2>/dev/null; then
        echo "[corpus-pressure] engine exited before health check passed" >&2
        cat "${TMP_DIR}/engine.stderr" >&2 || true
        exit 1
    fi
    sleep 0.1
done

curl -fsS "http://127.0.0.1:${PORT}/api/health" >/dev/null
sample_engine_memory

START_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"

for pcap in "${PCAPS[@]}"; do
    before_json="${TMP_DIR}/before.json"
    after_json="${TMP_DIR}/after.json"
    capture_stdout="${TMP_DIR}/capture.$(basename "${pcap}").stdout"
    capture_stderr="${TMP_DIR}/capture.$(basename "${pcap}").stderr"

    curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${before_json}"
    before_processed="$(python3 - "${before_json}" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as f:
    print(json.load(f).get("throughput", {}).get("packets_processed", 0))
PY
)"

    file_start_ns="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"
    bin/netsentry-capture -r "${pcap}" -s "${UDS_PATH}" -c 5 >"${capture_stdout}" 2>"${capture_stderr}"

    for _ in $(seq 1 "${WAIT_ATTEMPTS}"); do
        curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${after_json}"
        if python3 - "${after_json}" "${before_processed}" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as f:
    health = json.load(f)
before = int(sys.argv[2])
throughput = health.get("throughput", {})
heartbeat = health.get("capture", {}).get("heartbeat", {})
sent = int(heartbeat.get("sent", 0) or 0)
processed = int(throughput.get("packets_processed", 0) or 0)
sys.exit(0 if processed >= before + sent else 1)
PY
        then
            break
        fi
        sleep 0.05
    done

    file_end_ns="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"
    curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${after_json}"
    sample_engine_memory
    python3 - "${pcap}" "${before_json}" "${after_json}" "${file_start_ns}" "${file_end_ns}" "${INCLUDE_PATHS}" >>"${RUN_JSONL}" <<'PY'
import json
import os
import sys

pcap, before_path, after_path = sys.argv[1], sys.argv[2], sys.argv[3]
start_ns, end_ns = int(sys.argv[4]), int(sys.argv[5])
include_paths = sys.argv[6] == "1"

with open(before_path, "r", encoding="utf-8") as f:
    before = json.load(f)
with open(after_path, "r", encoding="utf-8") as f:
    after = json.load(f)

bt = before.get("throughput", {})
at = after.get("throughput", {})
heartbeat = after.get("capture", {}).get("heartbeat", {})
sent = int(heartbeat.get("sent", 0) or 0)
processed_delta = int(at.get("packets_processed", 0) or 0) - int(bt.get("packets_processed", 0) or 0)
alerts_delta = int(at.get("alerts_generated", 0) or 0) - int(bt.get("alerts_generated", 0) or 0)
elapsed = max((end_ns - start_ns) / 1_000_000_000, 0.000001)

print(json.dumps({
    "pcap_name": os.path.basename(pcap),
    "pcap_path": os.path.abspath(pcap) if include_paths else "redacted",
    "pcap_path_redacted": not include_paths,
    "capture_sent": sent,
    "capture_dropped": int(heartbeat.get("dropped", 0) or 0),
    "capture_parse_errors": int(heartbeat.get("parse_errors", 0) or 0),
    "capture_uds_write_errors": int(heartbeat.get("uds_write_errors", 0) or 0),
    "processed_delta": processed_delta,
    "alerts_delta": alerts_delta,
    "elapsed_seconds": elapsed,
    "packets_per_second": processed_delta / elapsed,
    "alerts_per_second": alerts_delta / elapsed,
}, sort_keys=True))
PY
    echo "[corpus-pressure] processed $(basename "${pcap}")"
done

END_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"

curl -fsS "http://127.0.0.1:${PORT}/api/health?verbose=true" >"${HEALTH_JSON}"
curl -fsS "http://127.0.0.1:${PORT}/api/alerts?per_page=100" >"${ALERTS_JSON}"
curl -fsS "http://127.0.0.1:${PORT}/api/metrics" >"${METRICS_TXT}"
sample_engine_memory
ENGINE_ERROR_LINES=0
if [[ -f "${TMP_DIR}/engine.log" ]]; then
    ENGINE_ERROR_LINES="$(grep -Eic 'error|panic|fatal' "${TMP_DIR}/engine.log" || true)"
fi

python3 - "${RUN_JSONL}" "${HEALTH_JSON}" "${ALERTS_JSON}" "${METRICS_TXT}" "${SUMMARY_JSON}" "${SUMMARY_MD}" "${RUN_ID}" "${PCAP_CORPUS}" "${START_NS}" "${END_NS}" "${INCLUDE_PATHS}" "${MAX_RSS_KB}" "${ENGINE_ERROR_LINES}" <<'PY'
import json
import os
import platform
import sys

runs_path, health_path, alerts_path, metrics_path = sys.argv[1:5]
summary_json, summary_md, run_id, corpus = sys.argv[5:9]
start_ns, end_ns = int(sys.argv[9]), int(sys.argv[10])
include_paths = sys.argv[11] == "1"
max_rss_kb = int(sys.argv[12])
engine_error_lines = int(sys.argv[13])

with open(runs_path, "r", encoding="utf-8") as f:
    runs = [json.loads(line) for line in f if line.strip()]
with open(health_path, "r", encoding="utf-8") as f:
    health = json.load(f)
with open(alerts_path, "r", encoding="utf-8") as f:
    alerts = json.load(f)
with open(metrics_path, "r", encoding="utf-8") as f:
    metrics_text = f.read()

elapsed = max((end_ns - start_ns) / 1_000_000_000, 0.000001)
total_processed = sum(item["processed_delta"] for item in runs)
total_alerts = sum(item["alerts_delta"] for item in runs)
total_sent = sum(item["capture_sent"] for item in runs)
total_parse_errors = sum(item["capture_parse_errors"] for item in runs)
total_dropped = sum(item["capture_dropped"] for item in runs)
total_write_errors = sum(item["capture_uds_write_errors"] for item in runs)
corpus_abs = os.path.abspath(corpus)
corpus_label = corpus_abs if include_paths else "redacted"

summary = {
    "run_id": run_id,
    "corpus": corpus_label,
    "corpus_path_redacted": not include_paths,
    "pcap_files": len(runs),
    "elapsed_seconds": elapsed,
    "capture_sent": total_sent,
    "packets_processed": total_processed,
    "alerts_generated": total_alerts,
    "capture_parse_errors": total_parse_errors,
    "capture_dropped": total_dropped,
    "capture_uds_write_errors": total_write_errors,
    "packets_per_second": total_processed / elapsed,
    "alerts_per_second": total_alerts / elapsed,
    "alert_match_rate": total_alerts / total_processed if total_processed else 0,
    "engine_peak_rss_kb_sampled": max_rss_kb,
    "engine_error_log_lines": engine_error_lines,
    "aggregated_alert_rows": alerts.get("pagination", {}).get("total"),
    "query_snapshot": alerts,
    "health": health,
    "runs": runs,
    "environment": {
        "platform": platform.platform(),
        "python": platform.python_version(),
    },
}

with open(summary_json, "w", encoding="utf-8") as f:
    json.dump(summary, f, indent=2, sort_keys=True)
    f.write("\n")

with open(summary_md, "w", encoding="utf-8") as f:
    f.write(f"# Corpus Pressure Evidence: {run_id}\n\n")
    f.write("## Summary\n\n")
    f.write(f"- Corpus: `{corpus_label}`\n")
    f.write(f"- Corpus path redacted: {str(not include_paths).lower()}\n")
    f.write(f"- Pcap files: {len(runs)}\n")
    f.write(f"- Elapsed seconds: {elapsed:.3f}\n")
    f.write(f"- Capture sent: {total_sent}\n")
    f.write(f"- Packets processed: {total_processed}\n")
    f.write(f"- Alerts generated: {total_alerts}\n")
    f.write(f"- Aggregated alert rows: {alerts.get('pagination', {}).get('total')}\n")
    f.write(f"- Packet rate: {total_processed / elapsed:.0f} pps\n")
    f.write(f"- Alert rate: {total_alerts / elapsed:.0f} alerts/sec\n")
    f.write(f"- Alert match rate: {total_alerts / total_processed:.4f}\n" if total_processed else "- Alert match rate: 0\n")
    f.write(f"- Engine peak RSS sampled: {max_rss_kb} KiB\n")
    f.write(f"- Engine error log lines: {engine_error_lines}\n")
    f.write(f"- Capture parse errors: {total_parse_errors}\n")
    f.write(f"- Capture dropped: {total_dropped}\n")
    f.write(f"- Capture UDS write errors: {total_write_errors}\n\n")
    f.write("## Files\n\n")
    f.write("| Pcap | Sent | Processed | Alerts | Parse errors | Dropped | UDS write errors | Seconds |\n")
    f.write("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
    for item in runs:
        f.write(
            f"| `{item['pcap_name']}` | {item['capture_sent']} | {item['processed_delta']} | "
            f"{item['alerts_delta']} | {item['capture_parse_errors']} | {item['capture_dropped']} | "
            f"{item['capture_uds_write_errors']} | {item['elapsed_seconds']:.3f} |\n"
        )
    f.write("\n## Notes\n\n")
    f.write("- Evidence files are local-only by default because corpus paths and operator notes may be sensitive.\n")
    f.write("- Corpus paths are redacted unless NETSENTRY_EVIDENCE_INCLUDE_PATHS=1 is set.\n")
    f.write("- These measurements are local release evidence, not a production throughput guarantee.\n")
    f.write("- Sanitize pcaps before sharing them outside the operator environment.\n")

print(
    "[corpus-pressure] ok: "
    f"files={len(runs)} packets={total_processed} alerts={total_alerts} "
    f"elapsed_sec={elapsed:.3f} pps={total_processed / elapsed:.0f} "
    f"alerts_per_sec={total_alerts / elapsed:.0f}"
)
print(f"[corpus-pressure] wrote {summary_json}")
print(f"[corpus-pressure] wrote {summary_md}")

if total_write_errors != 0:
    print("[corpus-pressure] capture UDS write errors were observed", file=sys.stderr)
    sys.exit(1)
if "netsentry_packets_processed_per_second" not in metrics_text:
    print("[corpus-pressure] expected process rate metrics to be present", file=sys.stderr)
    sys.exit(1)
PY
