#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${FUZZ_OUTPUT_DIR:-${ROOT_DIR}/docs/evidence/local}"
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
SUMMARY_JSON="${OUTPUT_DIR}/fuzz-sustained-${RUN_ID}.json"
SUMMARY_MD="${OUTPUT_DIR}/fuzz-sustained-${RUN_ID}.md"
ITERATIONS="${FUZZ_SUSTAINED_ITERATIONS:-1000000}"
CORPUS="${FUZZ_CORPUS:-}"
INCLUDE_PATHS="${NETSENTRY_EVIDENCE_INCLUDE_PATHS:-0}"
COMPILER="${CC:-cc}"
PARSER_LOG="$(mktemp /tmp/netsentry-fuzz-parser-sustained.XXXXXX.log)"
FORMATTER_LOG="$(mktemp /tmp/netsentry-fuzz-formatter-sustained.XXXXXX.log)"

cleanup() {
    rm -f "${PARSER_LOG}" "${FORMATTER_LOG}"
}
trap cleanup EXIT

usage() {
    cat >&2 <<'EOF_USAGE'
Usage:
  make fuzz-sustained

Optional:
  FUZZ_SUSTAINED_ITERATIONS=1000000
  FUZZ_CORPUS=/path/to/external-parser-corpus-file-or-directory
  FUZZ_OUTPUT_DIR=/path/to/evidence
  NETSENTRY_EVIDENCE_INCLUDE_PATHS=1

The deterministic parser and formatter harnesses run at the same iteration
budget. An optional external corpus is replayed by the byte-oriented parser
only. The corpus must stay local unless it has been reviewed for sharing.
Evidence is written as JSON and Markdown; default output under
docs/evidence/local/ is ignored. Corpus paths are redacted by default; set
NETSENTRY_EVIDENCE_INCLUDE_PATHS=1 only for private local debugging evidence.
EOF_USAGE
}

case "${ITERATIONS}" in
    ''|*[!0-9]*)
        echo "[fuzz-sustained] FUZZ_SUSTAINED_ITERATIONS must be a positive integer" >&2
        usage
        exit 2
        ;;
esac

if [[ "${ITERATIONS}" -le 0 ]]; then
    echo "[fuzz-sustained] FUZZ_SUSTAINED_ITERATIONS must be greater than zero" >&2
    exit 2
fi

if [[ "${INCLUDE_PATHS}" != "0" && "${INCLUDE_PATHS}" != "1" ]]; then
    echo "[fuzz-sustained] NETSENTRY_EVIDENCE_INCLUDE_PATHS must be 0 or 1" >&2
    exit 2
fi

for tool in make bash python3 "${COMPILER}"; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
        echo "[fuzz-sustained] required tool not found: ${tool}" >&2
        exit 2
    fi
done

CORPUS_FILES=()
if [[ -n "${CORPUS}" ]]; then
    if [[ ! -e "${CORPUS}" ]]; then
        echo "[fuzz-sustained] FUZZ_CORPUS does not exist: ${CORPUS}" >&2
        exit 2
    fi
    if [[ -f "${CORPUS}" ]]; then
        CORPUS_FILES+=("${CORPUS}")
    else
        while IFS= read -r -d '' path; do
            CORPUS_FILES+=("${path}")
        done < <(find "${CORPUS}" -type f -print0 | sort -z)
    fi
    if [[ "${#CORPUS_FILES[@]}" -eq 0 ]]; then
        echo "[fuzz-sustained] no corpus files found under ${CORPUS}" >&2
        exit 2
    fi
fi

mkdir -p "${OUTPUT_DIR}"
cd "${ROOT_DIR}"

# Force both versioned AddressSanitizer targets to rebuild immediately before
# the run; Make does not treat compiler-flag changes as ordinary dependencies.
make -C capture -B ../bin/fuzz-parser ../bin/fuzz-uds-formatter

COMPILER_VERSION="$("${COMPILER}" --version | head -n 1)"
MAKE_VERSION="$(make --version | head -n 1)"
BASH_VERSION_TEXT="$(bash --version | head -n 1)"
START_EPOCH="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
START_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"

PARSER_START_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"
set +e
ASAN_OPTIONS=detect_leaks=0 FUZZ_ITERATIONS="${ITERATIONS}" \
    bin/fuzz-parser >"${PARSER_LOG}" 2>&1
PARSER_STATUS=$?
set -e
PARSER_END_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"

CORPUS_STATUS=0
if [[ "${PARSER_STATUS}" -eq 0 && "${#CORPUS_FILES[@]}" -gt 0 ]]; then
    set +e
    ASAN_OPTIONS=detect_leaks=0 \
        bin/fuzz-parser "${CORPUS_FILES[@]}" >>"${PARSER_LOG}" 2>&1
    CORPUS_STATUS=$?
    set -e
fi

FORMATTER_STATUS=125
FORMATTER_START_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"
if [[ "${PARSER_STATUS}" -eq 0 && "${CORPUS_STATUS}" -eq 0 ]]; then
    set +e
    ASAN_OPTIONS=detect_leaks=0 FUZZ_FORMATTER_ITERATIONS="${ITERATIONS}" \
        bin/fuzz-uds-formatter >"${FORMATTER_LOG}" 2>&1
    FORMATTER_STATUS=$?
    set -e
else
    printf 'fuzz_uds_formatter: not run because an earlier fail-fast stage failed\n' \
        >"${FORMATTER_LOG}"
fi
FORMATTER_END_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"

END_NS="$(python3 -c 'import time; print(time.monotonic_ns())')"
END_EPOCH="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

python3 scripts/fuzz_evidence.py write \
    --json "${SUMMARY_JSON}" \
    --markdown "${SUMMARY_MD}" \
    --parser-log "${PARSER_LOG}" \
    --formatter-log "${FORMATTER_LOG}" \
    --run-id "${RUN_ID}" \
    --start "${START_EPOCH}" \
    --end "${END_EPOCH}" \
    --start-ns "${START_NS}" \
    --end-ns "${END_NS}" \
    --iterations "${ITERATIONS}" \
    --parser-status "${PARSER_STATUS}" \
    --parser-start-ns "${PARSER_START_NS}" \
    --parser-end-ns "${PARSER_END_NS}" \
    --formatter-status "${FORMATTER_STATUS}" \
    --formatter-start-ns "${FORMATTER_START_NS}" \
    --formatter-end-ns "${FORMATTER_END_NS}" \
    --corpus-status "${CORPUS_STATUS}" \
    --corpus "${CORPUS}" \
    --corpus-files "${#CORPUS_FILES[@]}" \
    --include-paths "${INCLUDE_PATHS}" \
    --compiler-version "${COMPILER_VERSION}" \
    --make-version "${MAKE_VERSION}" \
    --bash-version "${BASH_VERSION_TEXT}"
