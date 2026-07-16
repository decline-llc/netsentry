#!/usr/bin/env python3
"""Validate the reviewed non-PCAP evidence required before a NetSentry release."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


MIN_FUZZ_ITERATIONS = 1_000_000
SENSITIVE_LABELS = (
    "Fuzz corpus files staged",
    "Credentials or tokens present",
    "Local operator notes present",
    "Generated archives staged",
)


def sections(text: str) -> dict[str, dict[str, str]]:
    result: dict[str, dict[str, str]] = {"": {}}
    current = ""
    for line in text.splitlines():
        heading = re.match(r"^##\s+(.+?)\s*$", line)
        if heading:
            current = heading.group(1)
            result.setdefault(current, {})
            continue
        field = re.match(r"^-\s+([^:]+):\s*(.*?)\s*$", line)
        if field:
            result.setdefault(current, {})[field.group(1).strip()] = field.group(2).strip()
    return result


def required(section: dict[str, str], label: str, errors: list[str]) -> str:
    value = section.get(label, "")
    if not value:
        errors.append(f"missing field: {label}")
    return value


def positive_integer(value: str, label: str, errors: list[str]) -> int:
    match = re.search(r"\b\d[\d,]*\b", value)
    if not match:
        errors.append(f"{label} must contain a non-negative integer")
        return 0
    return int(match.group(0).replace(",", ""))


def validate(
    path: Path,
    exception_path: Path | None,
    manifest_path: Path | None = None,
    corpus_path: Path | None = None,
) -> list[str]:
    """Validate non-PCAP gates; PCAP compatibility arguments are ignored."""
    del exception_path, manifest_path, corpus_path
    errors: list[str] = []
    if not path.is_file():
        return [f"evidence file not found: {path}"]
    text = path.read_text(encoding="utf-8")
    lowered = text.casefold()
    for forbidden in ("/home/", "/tmp/", "-----begin ", "ghp_", "github_pat_"):
        if forbidden in lowered:
            errors.append(f"forbidden private or credential-like content: {forbidden.strip()}")

    parsed = sections(text)
    metadata = parsed.get("Metadata", {})
    fuzz = parsed.get("Sustained External C Fuzz Evidence", {})
    sensitive = parsed.get("Sensitive Information Review", {})
    final = parsed.get("Final Release Gate Decision", {})

    if required(metadata, "Final decision", errors).casefold() != "approved":
        errors.append("Metadata/Final decision must be approved")
    if required(fuzz, "Status", errors).casefold() not in {"pass", "passed"}:
        errors.append("fuzz Status must be pass")
    if positive_integer(
        required(fuzz, "Iterations or duration", errors),
        "fuzz iterations",
        errors,
    ) < MIN_FUZZ_ITERATIONS:
        errors.append(f"fuzz iterations must be at least {MIN_FUZZ_ITERATIONS}")
    if positive_integer(required(fuzz, "Crashes", errors), "fuzz crashes", errors) != 0:
        errors.append("fuzz Crashes must be 0")
    if required(fuzz, "ASan findings", errors).casefold() not in {"no", "none", "0", "zero"}:
        errors.append("fuzz ASan findings must be no/none/0")
    if required(fuzz, "Reviewer decision", errors).casefold() != "approved":
        errors.append("fuzz Reviewer decision must be approved")

    for label in SENSITIVE_LABELS:
        if required(sensitive, label, errors).casefold() != "no":
            errors.append(f"Sensitive Information Review/{label} must be no")

    if required(final, "Sustained external fuzz evidence reviewed", errors).casefold() != "yes":
        errors.append("final fuzz review must be yes")
    if required(final, "Approved for release", errors).casefold() != "yes":
        errors.append("Approved for release must be yes")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence", default="docs/evidence/release-v0.1.0.md")
    exception_group = parser.add_mutually_exclusive_group()
    exception_group.add_argument(
        "--exception",
        default="docs/audit/release_exception_v0.1.0.yaml",
        help="deprecated compatibility argument; ignored",
    )
    exception_group.add_argument(
        "--no-exception",
        action="store_true",
        help="deprecated compatibility argument; ignored",
    )
    parser.add_argument("--pcap-manifest", type=Path, help="optional diagnostic; ignored")
    parser.add_argument("--pcap-corpus", type=Path, help="optional diagnostic; ignored")
    args = parser.parse_args()
    errors = validate(
        Path(args.evidence),
        None if args.no_exception else Path(args.exception),
        manifest_path=args.pcap_manifest,
        corpus_path=args.pcap_corpus,
    )
    if errors:
        print(f"[release-gate] failed: {args.evidence}", file=sys.stderr)
        for error in errors:
            print(f"[release-gate] - {error}", file=sys.stderr)
        return 1
    print(f"[release-gate] passed: required non-PCAP evidence is complete ({args.evidence})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
