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
cp "${ROOT_DIR}/README.en.md" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/AUDIT_REPORT.md" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/LICENSE" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/CHANGELOG.md" "${STAGE_DIR}/${PACKAGE_NAME}/"
cp "${ROOT_DIR}/SECURITY.md" "${STAGE_DIR}/${PACKAGE_NAME}/"

cat >"${STAGE_DIR}/${PACKAGE_NAME}/RELEASE_NOTES.md" <<EOF_NOTES
# NetSentry ${VERSION}

This archive contains a development snapshot of NetSentry v0.1.0.

## Package Contents

- bin/netsentry-capture
- bin/netsentry-engine
- configs/
- docs/
- README.md
- README.en.md
- AUDIT_REPORT.md
- CHANGELOG.md
- SECURITY.md

## Quick Verification

After extracting the archive, start the engine with the packaged sample config:

\`\`\`bash
./bin/netsentry-engine -config configs/config.yaml
\`\`\`

In a source checkout, the deterministic end-to-end smoke test remains:

\`\`\`bash
make e2e-smoke
\`\`\`

## v0.1.0 Boundaries

- Offline pcap analysis is the primary path.
- Supported packet parsing covers Ethernet, VLAN/Q-in-Q, IPv4, TCP, and UDP passthrough.
- Detection rules cover payload, IP blacklist, and port blacklist matching.
- TCP stream reassembly, IP fragment reassembly, TLS decryption, IPv6, and full application-layer parsing are outside this snapshot.

## Release-Candidate Evidence

The source repository release-candidate bundle is \`make rc-check\`. It covers documentation consistency, dependency verification, C/Go tests, coverage snapshot, deterministic parser fuzz smoke, e2e smoke, archive checksum/content checks, Docker image content smoke, and Docker runtime health smoke.

## References

- README.md / README.en.md for bilingual quickstart and current behavior.
- AUDIT_REPORT.md for the prioritized audit and three-month roadmap.
- docs/development.md for local build and validation commands.
- docs/api-reference.md for API behavior.
- CHANGELOG.md for current gaps and release history.
EOF_NOTES

tar -C "${STAGE_DIR}" -czf "${DIST_DIR}/${PACKAGE_NAME}.tar.gz" "${PACKAGE_NAME}"
(
    cd "${DIST_DIR}"
    sha256sum "${PACKAGE_NAME}.tar.gz" >"${PACKAGE_NAME}.tar.gz.sha256"
)

echo "[dist] wrote ${DIST_DIR}/${PACKAGE_NAME}.tar.gz"
echo "[dist] wrote ${DIST_DIR}/${PACKAGE_NAME}.tar.gz.sha256"
