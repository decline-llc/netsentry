#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

DOC_FILES=()
while IFS= read -r file; do
    DOC_FILES+=("${file}")
done < <(
    {
        find . -maxdepth 1 -type f \( -name "README.md" -o -name "CHANGELOG.md" -o -name "SECURITY.md" -o -name "CONTRIBUTING.md" -o -name "CODE_OF_CONDUCT.md" \)
        find docs -type f -name "*.md" ! -path "docs/plans/*"
    } | sort
)

if [[ "${#DOC_FILES[@]}" -eq 0 ]]; then
    echo "[docs-check] no public markdown files found" >&2
    exit 1
fi

STALE_PATTERNS=(
    "Planned tests"
    "More C parser unit tests"
    "Broader UDS sender tests"
    "Full graceful shutdown tests"
    "Broader reconnect integration tests"
    "Race tests for rule reload"
    "End-to-end quickstart regression"
)

for pattern in "${STALE_PATTERNS[@]}"; do
    if grep -nF "${pattern}" "${DOC_FILES[@]}"; then
        echo "[docs-check] stale public documentation wording found: ${pattern}" >&2
        exit 1
    fi
done

echo "[docs-check] ok"
