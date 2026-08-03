#!/usr/bin/env python3
"""Build and validate path-safe sustained fuzz evidence."""

from __future__ import annotations

import argparse
import json
import os
import platform
import re
import sys
from pathlib import Path
from typing import Any

SCHEMA_VERSION = 1
EXPECTED_HARNESSES = {"parser", "formatter"}
RUN_ID_RE = re.compile(r"^[0-9]{8}T[0-9]{6}Z$")
SANITIZER_MARKERS = (
    "AddressSanitizer",
    "UndefinedBehaviorSanitizer",
    "runtime error:",
)


def _elapsed_seconds(start_ns: int, end_ns: int) -> float:
    return max((end_ns - start_ns) / 1_000_000_000, 0.000001)


def _read_log(path: Path, corpus: str, include_paths: bool) -> tuple[list[str], int]:
    text = path.read_text(encoding="utf-8", errors="replace")
    corpus_abs = os.path.abspath(corpus) if corpus else ""
    if corpus_abs and not include_paths:
        text = text.replace(corpus_abs, "<redacted-corpus>")
    findings = sum(text.count(marker) for marker in SANITIZER_MARKERS)
    return text.splitlines()[-40:], findings


def build_summary(args: argparse.Namespace) -> dict[str, Any]:
    include_paths = args.include_paths == "1"
    corpus_abs = os.path.abspath(args.corpus) if args.corpus else ""
    corpus_label = (
        corpus_abs
        if include_paths and corpus_abs
        else ("redacted" if corpus_abs else "")
    )
    parser_log, parser_findings = _read_log(
        args.parser_log, args.corpus, include_paths
    )
    formatter_log, formatter_findings = _read_log(
        args.formatter_log, args.corpus, include_paths
    )
    parser_ok = args.parser_status == 0 and parser_findings == 0
    formatter_ok = args.formatter_status == 0 and formatter_findings == 0
    corpus_ok = args.corpus_status == 0
    ok = parser_ok and formatter_ok and corpus_ok

    return {
        "schema_version": SCHEMA_VERSION,
        "run_id": args.run_id,
        "status": "pass" if ok else "fail",
        "evidence_class": "local_synthetic",
        "production_derived": False,
        "release_or_publication_authority": False,
        "start": args.start,
        "end": args.end,
        "elapsed_seconds": _elapsed_seconds(args.start_ns, args.end_ns),
        "iteration_budget_per_harness": args.iterations,
        "harnesses": {
            "parser": {
                "status": "pass" if parser_ok else "fail",
                "exit_status": args.parser_status,
                "iterations": args.iterations,
                "elapsed_seconds": _elapsed_seconds(
                    args.parser_start_ns, args.parser_end_ns
                ),
                "sanitizer_findings": parser_findings,
                "log_tail": parser_log,
            },
            "formatter": {
                "status": "pass" if formatter_ok else "fail",
                "exit_status": args.formatter_status,
                "iterations": args.iterations,
                "elapsed_seconds": _elapsed_seconds(
                    args.formatter_start_ns, args.formatter_end_ns
                ),
                "sanitizer_findings": formatter_findings,
                "log_tail": formatter_log,
            },
        },
        "corpus": {
            "provided": bool(corpus_abs),
            "path": corpus_label,
            "path_redacted": bool(corpus_abs and not include_paths),
            "files": args.corpus_files,
            "parser_replay_exit_status": args.corpus_status,
        },
        "environment": {
            "platform": platform.platform(),
            "python": platform.python_version(),
            "compiler": args.compiler_version,
            "make": args.make_version,
            "bash": args.bash_version,
        },
    }


def validation_errors(summary: dict[str, Any], require_pass: bool = False) -> list[str]:
    errors: list[str] = []
    if summary.get("schema_version") != SCHEMA_VERSION:
        errors.append(f"schema_version must be {SCHEMA_VERSION}")
    run_id = summary.get("run_id")
    if not isinstance(run_id, str) or not RUN_ID_RE.fullmatch(run_id):
        errors.append("run_id must use compact UTC form YYYYMMDDTHHMMSSZ")
    if summary.get("status") not in {"pass", "fail"}:
        errors.append("status must be pass or fail")
    if summary.get("evidence_class") != "local_synthetic":
        errors.append("evidence_class must be local_synthetic")
    if summary.get("production_derived") is not False:
        errors.append("production_derived must be false")
    if summary.get("release_or_publication_authority") is not False:
        errors.append("release_or_publication_authority must be false")
    budget = summary.get("iteration_budget_per_harness")
    if not isinstance(budget, int) or isinstance(budget, bool) or budget <= 0:
        errors.append("iteration_budget_per_harness must be a positive integer")

    harnesses = summary.get("harnesses")
    if not isinstance(harnesses, dict) or set(harnesses) != EXPECTED_HARNESSES:
        errors.append("harnesses must contain exactly parser and formatter")
    else:
        for name in sorted(EXPECTED_HARNESSES):
            harness = harnesses[name]
            if not isinstance(harness, dict):
                errors.append(f"{name} harness must be an object")
                continue
            if harness.get("status") not in {"pass", "fail"}:
                errors.append(f"{name} status must be pass or fail")
            if not isinstance(harness.get("exit_status"), int):
                errors.append(f"{name} exit_status must be an integer")
            if harness.get("iterations") != budget:
                errors.append(f"{name} iterations must equal the recorded budget")
            elapsed = harness.get("elapsed_seconds")
            if not isinstance(elapsed, (int, float)) or elapsed <= 0:
                errors.append(f"{name} elapsed_seconds must be positive")
            findings = harness.get("sanitizer_findings")
            if not isinstance(findings, int) or isinstance(findings, bool) or findings < 0:
                errors.append(f"{name} sanitizer_findings must be a nonnegative integer")
            if not isinstance(harness.get("log_tail"), list):
                errors.append(f"{name} log_tail must be an array")
            elif harness.get("status") == "pass":
                expected_line = f"fuzz_{'parser' if name == 'parser' else 'uds_formatter'}: ok iterations={budget}"
                if expected_line not in harness["log_tail"]:
                    errors.append(
                        f"{name} log_tail must confirm the recorded iteration budget"
                    )

    corpus = summary.get("corpus")
    if not isinstance(corpus, dict):
        errors.append("corpus must be an object")
    else:
        provided = corpus.get("provided")
        path = corpus.get("path")
        redacted = corpus.get("path_redacted")
        files = corpus.get("files")
        if not isinstance(provided, bool) or not isinstance(redacted, bool):
            errors.append("corpus provided/path_redacted must be booleans")
        if not isinstance(path, str):
            errors.append("corpus path must be a string")
        if not isinstance(files, int) or isinstance(files, bool) or files < 0:
            errors.append("corpus files must be a nonnegative integer")
        if provided and (not path or files <= 0):
            errors.append("provided corpus must record a path label and files")
        if not provided and (path or redacted or files != 0):
            errors.append("absent corpus must have empty path, no redaction, and zero files")
        if redacted and path != "redacted":
            errors.append("redacted corpus path must use the literal redacted marker")

    environment = summary.get("environment")
    if not isinstance(environment, dict):
        errors.append("environment must be an object")
    else:
        for name in ("platform", "python", "compiler", "make", "bash"):
            if not isinstance(environment.get(name), str) or not environment[name]:
                errors.append(f"environment {name} must be a nonempty string")

    if require_pass:
        if summary.get("status") != "pass":
            errors.append("evidence status must be pass")
        if isinstance(harnesses, dict):
            for name in sorted(EXPECTED_HARNESSES & set(harnesses)):
                harness = harnesses[name]
                if isinstance(harness, dict):
                    if harness.get("status") != "pass" or harness.get("exit_status") != 0:
                        errors.append(f"{name} harness must pass with exit status zero")
                    if harness.get("sanitizer_findings") != 0:
                        errors.append(f"{name} harness must have zero sanitizer findings")
        if isinstance(corpus, dict) and corpus.get("parser_replay_exit_status") != 0:
            errors.append("optional parser corpus replay must have exit status zero")
    return errors


def write_markdown(path: Path, summary: dict[str, Any]) -> None:
    corpus = summary["corpus"]
    lines = [
        f"# Sustained Parser and Formatter Fuzz Evidence: {summary['run_id']}",
        "",
        "## Summary",
        "",
        f"- Status: {summary['status']}",
        f"- Evidence class: {summary['evidence_class']}",
        f"- Production-derived: {str(summary['production_derived']).lower()}",
        "- Release or publication authority: false",
        f"- Start: {summary['start']}",
        f"- End: {summary['end']}",
        f"- Elapsed seconds: {summary['elapsed_seconds']:.3f}",
        f"- Iterations per harness: {summary['iteration_budget_per_harness']}",
        f"- Corpus: `{corpus['path']}`" if corpus["path"] else "- Corpus: not provided",
        f"- Corpus path redacted: {str(corpus['path_redacted']).lower()}",
        f"- Corpus files: {corpus['files']}",
        "",
    ]
    for name in ("parser", "formatter"):
        harness = summary["harnesses"][name]
        lines.extend(
            [
                f"## {name.title()} Harness",
                "",
                f"- Status: {harness['status']}",
                f"- Exit status: {harness['exit_status']}",
                f"- Iterations: {harness['iterations']}",
                f"- Elapsed seconds: {harness['elapsed_seconds']:.3f}",
                f"- Sanitizer findings: {harness['sanitizer_findings']}",
                "",
                "```text",
                *harness["log_tail"],
                "```",
                "",
            ]
        )
    lines.extend(
        [
            "## Notes",
            "",
            "- Evidence files are local-only by default because corpus paths may be sensitive.",
            "- Corpus paths are redacted unless NETSENTRY_EVIDENCE_INCLUDE_PATHS=1 is set.",
            "- The deterministic built-in seeds and mutations are local synthetic robustness evidence.",
            "- This result is not production-traffic, throughput, release, tag, or publication authority.",
        ]
    )
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def command_write(args: argparse.Namespace) -> int:
    summary = build_summary(args)
    errors = validation_errors(summary)
    if errors:
        for error in errors:
            print(f"[fuzz-evidence] {error}", file=sys.stderr)
        return 2
    args.json.parent.mkdir(parents=True, exist_ok=True)
    args.markdown.parent.mkdir(parents=True, exist_ok=True)
    args.json.write_text(
        json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    write_markdown(args.markdown, summary)
    print(
        "[fuzz-sustained] "
        f"status={summary['status']} iterations_per_harness={args.iterations} "
        f"corpus_files={args.corpus_files} elapsed_sec={summary['elapsed_seconds']:.3f}"
    )
    print(f"[fuzz-sustained] wrote {args.json}")
    print(f"[fuzz-sustained] wrote {args.markdown}")
    pass_errors = validation_errors(summary, require_pass=True)
    if pass_errors:
        for error in pass_errors:
            print(f"[fuzz-evidence] {error}", file=sys.stderr)
        return 1
    return 0


def command_validate(args: argparse.Namespace) -> int:
    summary = json.loads(args.json.read_text(encoding="utf-8"))
    errors = validation_errors(summary, require_pass=args.require_pass)
    if errors:
        for error in errors:
            print(f"[fuzz-evidence] {error}", file=sys.stderr)
        return 1
    print(f"[fuzz-evidence] ok: {args.json}")
    return 0


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(description=__doc__)
    commands = root.add_subparsers(dest="command", required=True)
    write = commands.add_parser("write")
    for name in ("json", "markdown", "parser-log", "formatter-log"):
        write.add_argument(f"--{name}", type=Path, required=True)
    for name in ("run-id", "start", "end", "corpus", "include-paths"):
        write.add_argument(f"--{name}", required=True)
    for name in (
        "start-ns", "end-ns", "iterations", "parser-status",
        "parser-start-ns", "parser-end-ns", "formatter-status",
        "formatter-start-ns", "formatter-end-ns", "corpus-status",
        "corpus-files",
    ):
        write.add_argument(f"--{name}", type=int, required=True)
    for name in ("compiler-version", "make-version", "bash-version"):
        write.add_argument(f"--{name}", required=True)
    write.set_defaults(func=command_write)

    validate = commands.add_parser("validate")
    validate.add_argument("--json", type=Path, required=True)
    validate.add_argument("--require-pass", action="store_true")
    validate.set_defaults(func=command_validate)
    return root


def main() -> int:
    args = parser().parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
