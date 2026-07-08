#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${FUZZ_OUTPUT_DIR:-${ROOT_DIR}/docs/evidence/local}"
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
SUMMARY_JSON="${OUTPUT_DIR}/fuzz-sustained-${RUN_ID}.json"
SUMMARY_MD="${OUTPUT_DIR}/fuzz-sustained-${RUN_ID}.md"
ITERATIONS="${FUZZ_SUSTAINED_ITERATIONS:-1000000}"
CORPUS="${FUZZ_CORPUS:-}"
LOG_FILE="$(mktemp /tmp/netsentry-fuzz-sustained.XXXXXX.log)"

cleanup() {
    rm -f "${LOG_FILE}"
}
trap cleanup EXIT

usage() {
    cat >&2 <<'EOF_USAGE'
Usage:
  make fuzz-sustained

Optional:
  FUZZ_SUSTAINED_ITERATIONS=1000000
  FUZZ_CORPUS=/path/to/external-corpus-file-or-directory
  FUZZ_OUTPUT_DIR=/path/to/evidence

The corpus must stay local unless it has been reviewed for sharing. Evidence is
written as JSON and Markdown; default output under docs/evidence/local/ is ignored.
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
make -C capture ../bin/fuzz-parser

START_EPOCH="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
START_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"

set +e
ASAN_OPTIONS=detect_leaks=0 FUZZ_ITERATIONS="${ITERATIONS}" bin/fuzz-parser >"${LOG_FILE}" 2>&1
MUTATION_STATUS=$?
CORPUS_STATUS=0
if [[ "${MUTATION_STATUS}" -eq 0 && "${#CORPUS_FILES[@]}" -gt 0 ]]; then
    ASAN_OPTIONS=detect_leaks=0 bin/fuzz-parser "${CORPUS_FILES[@]}" >>"${LOG_FILE}" 2>&1
    CORPUS_STATUS=$?
fi
set -e

END_NS="$(python3 - <<'PY'
import time
print(time.monotonic_ns())
PY
)"
END_EPOCH="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

python3 - "${SUMMARY_JSON}" "${SUMMARY_MD}" "${LOG_FILE}" "${RUN_ID}" "${START_EPOCH}" "${END_EPOCH}" "${START_NS}" "${END_NS}" "${ITERATIONS}" "${MUTATION_STATUS}" "${CORPUS_STATUS}" "${CORPUS:-}" "${#CORPUS_FILES[@]}" <<'PY'
import json
import os
import platform
import sys

(
    summary_json,
    summary_md,
    log_path,
    run_id,
    start_epoch,
    end_epoch,
    start_ns,
    end_ns,
    iterations,
    mutation_status,
    corpus_status,
    corpus,
    corpus_count,
) = sys.argv[1:14]

start_ns = int(start_ns)
end_ns = int(end_ns)
iterations = int(iterations)
mutation_status = int(mutation_status)
corpus_status = int(corpus_status)
corpus_count = int(corpus_count)
elapsed = max((end_ns - start_ns) / 1_000_000_000, 0.000001)

with open(log_path, "r", encoding="utf-8", errors="replace") as f:
    log_tail = f.readlines()[-40:]

ok = mutation_status == 0 and corpus_status == 0
summary = {
    "run_id": run_id,
    "status": "pass" if ok else "fail",
    "start": start_epoch,
    "end": end_epoch,
    "elapsed_seconds": elapsed,
    "iterations": iterations,
    "mutation_status": mutation_status,
    "corpus": os.path.abspath(corpus) if corpus else "",
    "corpus_files": corpus_count,
    "corpus_status": corpus_status,
    "environment": {
        "platform": platform.platform(),
        "python": platform.python_version(),
    },
    "log_tail": [line.rstrip("\n") for line in log_tail],
}

with open(summary_json, "w", encoding="utf-8") as f:
    json.dump(summary, f, indent=2, sort_keys=True)
    f.write("\n")

with open(summary_md, "w", encoding="utf-8") as f:
    f.write(f"# Sustained Fuzz Evidence: {run_id}\n\n")
    f.write("## Summary\n\n")
    f.write(f"- Status: {summary['status']}\n")
    f.write(f"- Start: {start_epoch}\n")
    f.write(f"- End: {end_epoch}\n")
    f.write(f"- Elapsed seconds: {elapsed:.3f}\n")
    f.write(f"- Deterministic iterations: {iterations}\n")
    f.write(f"- Mutation status: {mutation_status}\n")
    f.write(f"- Corpus: `{summary['corpus']}`\n" if summary["corpus"] else "- Corpus: not provided\n")
    f.write(f"- Corpus files: {corpus_count}\n")
    f.write(f"- Corpus status: {corpus_status}\n\n")
    f.write("## Log Tail\n\n")
    f.write("```text\n")
    f.write("".join(log_tail))
    if log_tail and not log_tail[-1].endswith("\n"):
        f.write("\n")
    f.write("```\n\n")
    f.write("## Notes\n\n")
    f.write("- Evidence files are local-only by default because corpus paths may be sensitive.\n")
    f.write("- This target exercises the existing ASan C parser fuzz harness.\n")
    f.write("- External corpus quality and duration must be reviewed before treating the result as release evidence.\n")

print(
    "[fuzz-sustained] "
    f"status={summary['status']} iterations={iterations} corpus_files={corpus_count} "
    f"elapsed_sec={elapsed:.3f}"
)
print(f"[fuzz-sustained] wrote {summary_json}")
print(f"[fuzz-sustained] wrote {summary_md}")

sys.exit(0 if ok else 1)
PY
