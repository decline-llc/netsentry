#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-0.1.0-dev}"
IMAGE="${IMAGE:-netsentry:${VERSION}}"
DOCKER_CMD="${DOCKER:-docker}"
SKIP_DOCKER="${SKIP_DOCKER:-0}"

cd "${ROOT_DIR}"

echo "[rc-check] shell syntax"
bash -n scripts/e2e_smoke.sh
bash -n scripts/package_release.sh
bash -n scripts/rc_check.sh

echo "[rc-check] make test"
make test

echo "[rc-check] make e2e-smoke"
make e2e-smoke

echo "[rc-check] make dist VERSION=${VERSION}"
make dist VERSION="${VERSION}"

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
     test -f configs/rules.json'

echo "[rc-check] ok"
