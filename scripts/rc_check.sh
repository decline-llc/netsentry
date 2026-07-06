#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-0.1.0-dev}"
IMAGE="${IMAGE:-netsentry:${VERSION}}"
DOCKER_CMD="${DOCKER:-docker}"
SKIP_DOCKER="${SKIP_DOCKER:-0}"
DOCKER_CONTAINER_ID=""
OS_NAME="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_NAME="$(uname -m)"

case "${ARCH_NAME}" in
    x86_64) ARCH_NAME="amd64" ;;
    aarch64 | arm64) ARCH_NAME="arm64" ;;
esac

PACKAGE_NAME="netsentry-${VERSION}-${OS_NAME}-${ARCH_NAME}"
ARCHIVE_PATH="dist/${PACKAGE_NAME}.tar.gz"

cleanup() {
    if [[ -n "${DOCKER_CONTAINER_ID}" ]]; then
        read -r -a docker_parts <<<"${DOCKER_CMD}"
        "${docker_parts[@]}" rm -f "${DOCKER_CONTAINER_ID}" >/dev/null 2>&1 || true
    fi
}
trap cleanup EXIT

cd "${ROOT_DIR}"

echo "[rc-check] shell syntax"
bash -n scripts/e2e_smoke.sh
bash -n scripts/e2e_pressure.sh
bash -n scripts/package_release.sh
bash -n scripts/rc_check.sh

echo "[rc-check] make deps-check"
make deps-check

echo "[rc-check] make test"
make test

echo "[rc-check] make test-coverage"
make test-coverage

echo "[rc-check] make fuzz-parser"
make fuzz-parser

echo "[rc-check] make e2e-smoke"
make e2e-smoke

echo "[rc-check] make dist VERSION=${VERSION}"
make dist VERSION="${VERSION}"

echo "[rc-check] dist archive smoke"
(
    cd dist
    sha256sum -c "${PACKAGE_NAME}.tar.gz.sha256"
)
tar -tzf "${ARCHIVE_PATH}" \
    "${PACKAGE_NAME}/bin/netsentry-capture" \
    "${PACKAGE_NAME}/bin/netsentry-engine" \
    "${PACKAGE_NAME}/configs/config.yaml" \
    "${PACKAGE_NAME}/configs/rules.json" \
    "${PACKAGE_NAME}/configs/suppressions.json" \
    "${PACKAGE_NAME}/docs/development.md" \
    "${PACKAGE_NAME}/README.md" \
    "${PACKAGE_NAME}/CHANGELOG.md" \
    "${PACKAGE_NAME}/RELEASE_NOTES.md" >/dev/null

if [[ "${SKIP_DOCKER}" == "1" ]]; then
    echo "[rc-check] docker checks skipped"
    exit 0
fi

echo "[rc-check] make docker-build IMAGE=${IMAGE}"
make docker-build IMAGE="${IMAGE}" DOCKER="${DOCKER_CMD}"

echo "[rc-check] docker image smoke"
read -r -a docker_parts <<<"${DOCKER_CMD}"
"${docker_parts[@]}" run --rm --entrypoint sh "${IMAGE}" -c \
    'test -x /usr/local/bin/netsentry-engine &&
     test -x /usr/local/bin/netsentry-capture &&
     test -f configs/config.yaml &&
     test -f configs/rules.json &&
     test -f configs/suppressions.json'

echo "[rc-check] docker runtime health smoke"
DOCKER_CONTAINER_ID="$("${docker_parts[@]}" run --rm -d "${IMAGE}")"
DOCKER_HEALTH_OK=0
for _ in $(seq 1 50); do
    if "${docker_parts[@]}" exec "${DOCKER_CONTAINER_ID}" \
        curl -fsS "http://127.0.0.1:8080/api/health" >/dev/null 2>&1; then
        DOCKER_HEALTH_OK=1
        break
    fi
    sleep 0.2
done

if [[ "${DOCKER_HEALTH_OK}" == "1" ]]; then
    "${docker_parts[@]}" rm -f "${DOCKER_CONTAINER_ID}" >/dev/null
    DOCKER_CONTAINER_ID=""
    echo "[rc-check] docker runtime health ok"
    echo "[rc-check] ok"
    exit 0
fi

echo "[rc-check] docker runtime health failed" >&2
"${docker_parts[@]}" logs "${DOCKER_CONTAINER_ID}" >&2 || true
exit 1
