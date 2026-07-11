#!/usr/bin/env python3
"""Validate the reviewed public evidence required before a NetSentry release."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


MIN_FUZZ_ITERATIONS = 1_000_000
SENSITIVE_LABELS = (
    "Raw pcaps staged",
    "Fuzz corpus files staged",
    "Private corpus paths present",
    "Credentials or tokens present",
    "Local operator notes present",
    "Generated archives staged",
)


def sections(text: str) -> dict[str, dict[str, str]]:
    result: dict[str, dict[str, str]] = {}
    current = ""
    result[current] = {}
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


def approved(value: str) -> bool:
    return value.casefold() in {"approved", "pass", "passed", "yes", "zero", "0", "no"}


def positive_integer(value: str, label: str, errors: list[str]) -> int:
    match = re.search(r"\b\d[\d,]*\b", value)
    if not match:
        errors.append(f"{label} must contain a non-negative integer")
        return 0
    return int(match.group(0).replace(",", ""))


def validate(path: Path) -> list[str]:
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
    pcap = parsed.get("Realistic Sanitized Pcap Corpus Evidence", {})
    sensitive = parsed.get("Sensitive Information Review", {})
    final = parsed.get("Final Release Gate Decision", {})

    if required(metadata, "Final decision", errors).casefold() != "approved":
        errors.append("Metadata/Final decision must be approved")
    if required(fuzz, "Status", errors).casefold() not in {"pass", "passed"}:
        errors.append("fuzz Status must be pass")
    if positive_integer(required(fuzz, "Iterations or duration", errors), "fuzz iterations", errors) < MIN_FUZZ_ITERATIONS:
        errors.append(f"fuzz iterations must be at least {MIN_FUZZ_ITERATIONS}")
    if positive_integer(required(fuzz, "Crashes", errors), "fuzz crashes", errors) != 0:
        errors.append("fuzz Crashes must be 0")
    if required(fuzz, "ASan findings", errors).casefold() not in {"no", "none", "0", "zero"}:
        errors.append("fuzz ASan findings must be no/none/0")
    if required(fuzz, "Reviewer decision", errors).casefold() != "approved":
        errors.append("fuzz Reviewer decision must be approved")

    if required(pcap, "Status", errors).casefold() not in {"pass", "passed"}:
        errors.append("pcap Status must be pass")
    if positive_integer(required(pcap, "Pcap files", errors), "pcap files", errors) < 1:
        errors.append("pcap Pcap files must be at least 1")
    if positive_integer(required(pcap, "Packets processed", errors), "packets processed", errors) < 1:
        errors.append("pcap Packets processed must be greater than 0")
    if required(pcap, "Query evidence", errors).casefold() not in {"pass", "passed", "approved"}:
        errors.append("pcap Query evidence must be pass/approved")
    if required(pcap, "Reviewer decision", errors).casefold() != "approved":
        errors.append("pcap Reviewer decision must be approved")

    for label in SENSITIVE_LABELS:
        if required(sensitive, label, errors).casefold() != "no":
            errors.append(f"Sensitive Information Review/{label} must be no")

    if required(final, "Sustained external fuzz evidence reviewed", errors).casefold() != "yes":
        errors.append("final fuzz review must be yes")
    if required(final, "Realistic sanitized pcap corpus evidence reviewed", errors).casefold() != "yes":
        errors.append("final pcap review must be yes")
    if required(final, "Approved for release", errors).casefold() != "yes":
        errors.append("Approved for release must be yes")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence", default="docs/evidence/release-v0.1.0.md")
    args = parser.parse_args()
    errors = validate(Path(args.evidence))
    if errors:
        print(f"[release-gate] failed: {args.evidence}", file=sys.stderr)
        for error in errors:
            print(f"[release-gate] - {error}", file=sys.stderr)
        return 1
    print(f"[release-gate] passed: reviewed evidence is complete ({args.evidence})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
