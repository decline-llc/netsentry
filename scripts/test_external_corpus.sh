#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ASSET_DIR="${NETSENTRY_TEST_ASSETS:-${ROOT_DIR}/../NetSentry_TestAssets}"
TMP_DIR="$(mktemp -d)"
LISTENER_PID=""

cleanup() {
    if [[ -n "${LISTENER_PID}" ]] && kill -0 "${LISTENER_PID}" 2>/dev/null; then
        kill "${LISTENER_PID}" 2>/dev/null || true
        wait "${LISTENER_PID}" 2>/dev/null || true
    fi
    rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

if [[ ! -x "${ASSET_DIR}/manage_pcaps.py" ]]; then
    echo "[external-corpus] asset manager missing: ${ASSET_DIR}/manage_pcaps.py" >&2
    exit 2
fi

"${ASSET_DIR}/manage_pcaps.py" verify

SUPPORTED_DIR="${TMP_DIR}/supported"
mkdir -p "${SUPPORTED_DIR}"
for name in \
    pcpp-http.pcap \
    pcpp-ipv4-frag.pcap \
    pcpp-dns.pcap \
    pcpp-qinq.pcapng \
    zeek-quickstart.pcap \
    zeek-ipv4-tcp-good-checksum.pcap
do
    cp "${ASSET_DIR}/pcaps/${name}" "${SUPPORTED_DIR}/${name}"
done

NETSENTRY_EVIDENCE_INCLUDE_PATHS=0 \
PCAP_CORPUS="${SUPPORTED_DIR}" \
CORPUS_OUTPUT_DIR="${TMP_DIR}/evidence" \
    bash "${ROOT_DIR}/scripts/e2e_corpus_pressure.sh"

INVALID_LOG="${TMP_DIR}/invalid-cli.log"
set +e
"${ROOT_DIR}/bin/netsentry-capture" \
    -r "${ASSET_DIR}/pcaps/pcpp-http.pcap" -c invalid \
    >"${TMP_DIR}/invalid-cli.stdout" 2>"${INVALID_LOG}"
invalid_status=$?
set -e
if [[ "${invalid_status}" -ne 2 ]] || ! grep -q "connect_retries must be" "${INVALID_LOG}"; then
    echo "[external-corpus] invalid -c contract failed: status=${invalid_status}" >&2
    cat "${INVALID_LOG}" >&2
    exit 1
fi

UDS_PATH="${TMP_DIR}/dlt.sock"
python3 - "${UDS_PATH}" <<'PY' &
import os
import socket
import sys

path = sys.argv[1]
try:
    os.unlink(path)
except FileNotFoundError:
    pass
with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as listener:
    listener.bind(path)
    listener.listen(1)
    conn, _ = listener.accept()
    with conn:
        data = b""
        while b"\n" not in data:
            chunk = conn.recv(4096)
            if not chunk:
                break
            data += chunk
PY
LISTENER_PID="$!"

for _ in $(seq 1 50); do
    [[ -S "${UDS_PATH}" ]] && break
    sleep 0.02
done

DLT_LOG="${TMP_DIR}/dlt.log"
set +e
"${ROOT_DIR}/bin/netsentry-capture" \
    -r "${ASSET_DIR}/pcaps/pcpp-linux-sll.pcap" -s "${UDS_PATH}" -c 1 \
    >"${TMP_DIR}/dlt.stdout" 2>"${DLT_LOG}"
dlt_status=$?
set -e
wait "${LISTENER_PID}"
LISTENER_PID=""
if [[ "${dlt_status}" -ne 2 ]] || ! grep -q "unsupported pcap data link type" "${DLT_LOG}"; then
    echo "[external-corpus] non-Ethernet DLT contract failed: status=${dlt_status}" >&2
    cat "${DLT_LOG}" >&2
    exit 1
fi

echo "[external-corpus] ok: checksums, 6 supported captures, invalid CLI, and non-Ethernet rejection"
