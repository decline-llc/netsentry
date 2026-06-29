#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-0.1.0-dev}"
OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_NAME="$(uname -m)"

case "${ARCH_NAME}" in
    x86_64) ARCH_NAME="amd64" ;;
    aarch64 | arm64) ARCH_NAME="arm64" ;;
esac

PACKAGE_NAME="netsentry-${VERSION}-${OS_NAME}-${ARCH_NAME}"
DIST_DIR="${ROOT_DIR}/dist"
STAGE_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "${STAGE_DIR}"
}
trap cleanup EXIT

mkdir -p "${DIST_DIR}"
mkdir -p "${STAGE_DIR}/${PACKAGE_NAME}/bin"

cp "${ROOT_DIR}/bin/netsentry-capture" "${STAGE_DIR}/${PACKAGE_NAME}/bin/"
cp "${ROOT_DIR}/bin/netsentry-engine" "${STAGE_DIR}/${PACKAGE_NAME}/bin/"
cp -R "${ROOT_DIR}/configs" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp -R "${ROOT_DIR}/docs" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/README.md" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/LICENSE" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/CHANGELOG.md" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/SECURITY.md" "${STAGE_DIR}/${PACKAGE_NAME}/"

cat >"${STAGE_DIR}/${PACKAGE_NAME}/RELEASE_NOTES.md" <<EOF_NOTES
# NetSentry ${VERSION}

This archive contains a development snapshot of NetSentry:

- bin/netsentry-capture
- bin/netsentry-engine
- configs/
- docs/

Quick verification after extracting:

\`\`\`bash
./bin/netsentry-engine -config configs/config.yaml
\`\`\`

For the full repository smoke test, use \`make e2e-smoke\` from a source checkout.
EOF_NOTES

tar -C "${STAGE_DIR}" -czf "${DIST_DIR}/${PACKAGE_NAME}.tar.gz" "${PACKAGE_NAME}"
(
    cd "${DIST_DIR}"
    sha256sum "${PACKAGE_NAME}.tar.gz" >"${PACKAGE_NAME}.tar.gz.sha256"
)

echo "[dist] wrote ${DIST_DIR}/${PACKAGE_NAME}.tar.gz"
echo "[dist] wrote ${DIST_DIR}/${PACKAGE_NAME}.tar.gz.sha256"
