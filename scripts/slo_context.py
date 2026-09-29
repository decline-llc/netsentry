#!/usr/bin/env python3
"""Retain supplied SLO environment declarations and references; no discovery or certification."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import sys
from typing import Any, Callable, Sequence

if __package__:
    from . import slo_bundle as bundle, slo_report as report
else:
    import slo_bundle as bundle
    import slo_report as report

MAX_DECLARATION_BYTES = 1024**2
MAX_EVIDENCE_FILE_BYTES = 64 * 1024**2
DEFAULT_MAX_EVIDENCE_BYTES = 256 * 1024**2
MAX_REFERENCES = 64
# All keys are required; null is the explicit unknown value for every field.
FIELDS = {
    "hardware": {
        "deployment_type": ("kvm_vm", "physical"), "cpu_arch": ("x86_64", "aarch64", "other"),
        "cpu_model": "text", "cpu_pinning": "text", "sut_vcpus": "positive_integer",
        "sut_memory_bytes": "positive_integer", "storage_model": "text",
        "storage_medium": ("ssd", "nvme", "hdd", "other"), "storage_capacity_bytes": "positive_integer",
        "storage_dedicated_partition": "boolean", "nic_model": "text",
        "nic_mode": ("physical", "sriov_vf", "virtual", "other"),
        "nic_speed_bps": "positive_integer", "rss_enabled": "boolean",
    },
    "toolchain": {
        "os_release": "text", "kernel_release": "text", "go_version": "text",
        "c_compiler": "text", "libpcap_version": "text", "build_flags": "text",
    },
    "isolation": {
        "same_host": "boolean", "separate_workdir": "boolean", "separate_process_groups": "boolean",
        "fresh_runtime": "boolean", "allocation_notes": "text",
    },
}


def _text(value: Any) -> None:
    bundle._require(isinstance(value, str) and 1 <= len(value.encode("utf-8")) <= 512
                    and value == value.strip() and value.isprintable(), "expected bounded printable text")


def _evidence_id(value: Any) -> str:
    bundle._require(isinstance(value, str) and re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]{0,63}", value) is not None,
                    "expected safe evidence ID")
    return value


def validate(document: Any) -> list[str]:
    """Validate declaration v1; return coverage gaps for unknown/unreferenced values."""
    report._object(document, {"schema_version", "artifact_kind", "run_id", "profile", "started_at",
                             "observations_sha256", "hardware", "toolchain", "isolation", "evidence"}, "context")
    bundle._version(document)
    bundle._require(document["artifact_kind"] == "netsentry_slo_run_context", "wrong context kind")
    bundle._require(isinstance(document["profile"], str) and document["profile"] in report.PROFILES,
                    "invalid profile")
    report._timestamp(document["started_at"])
    bundle._hash(document["observations_sha256"])
    catalog = document["evidence"]
    bundle._require(isinstance(catalog, list) and len(catalog) <= MAX_REFERENCES, "invalid evidence catalog")
    ids = set()
    for item in catalog:
        report._object(item, {"id", "description", "bytes", "sha256"}, "evidence reference")
        name = _evidence_id(item["id"])
        bundle._require(name not in ids, "duplicate evidence ID")
        ids.add(name)
        _text(item["description"])
        report._integer(item["bytes"], "evidence bytes", 1)
        bundle._hash(item["sha256"])
    used = set()
    gaps = []
    for section, fields in FIELDS.items():
        report._object(document[section], set(fields), section)
        for key, kind in fields.items():
            field = section + "." + key
            cell = report._object(document[section][key], {"value", "evidence_ids"}, field)
            refs = cell["evidence_ids"]
            bundle._require(isinstance(refs, list) and len(refs) <= MAX_REFERENCES, "invalid evidence IDs")
            checked = [_evidence_id(value) for value in refs]
            bundle._require(len(set(checked)) == len(checked) and set(checked) <= ids,
                            "duplicate or unknown field reference")
            used.update(checked)
            value = cell["value"]
            if value is None:
                gaps.append(field + ": unknown value")
                continue
            if kind == "text":
                _text(value)
            elif kind == "positive_integer":
                report._integer(value, field, 1)
            elif kind == "boolean":
                report._boolean(value, field)
            else:
                bundle._require(isinstance(value, str) and value in kind, "invalid enumeration value")
            if not refs:
                gaps.append(field + ": known value lacks evidence references")
    bundle._require(used == ids, "unreferenced evidence catalog entry")
    return gaps


def _read(path: Path, maximum: int) -> Any:
    with path.open("rb") as stream:
        raw = stream.read(maximum + 1)
    bundle._require(len(raw) <= maximum, "JSON byte limit exceeded")
    return json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                      parse_constant=report._constant, parse_float=bundle._finite_float)


class _Retention:
    def __init__(self, output: Path):
        self.output = output
        self.inventory: dict[str, dict[str, Any]] = {}
        self.gaps: list[str] = []
        self.mismatches: list[str] = []
        self.checked: list[str] = []

    def check(self, label: str, action: Callable[[], Any]) -> bool:
        try:
            action()
        except (ValueError, TypeError, KeyError, IndexError, RecursionError):
            self.mismatches.append(label + ": invalid or inconsistent evidence")
            return False
        self.checked.append(label)
        return True

    def snapshot(self, source: Path | None, key: str, limit: int) -> bool:
        if source is None:
            self.gaps.append(key + ": not supplied")
            return False
        try:
            descriptor = os.open(source.expanduser(), os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        except OSError:
            self.gaps.append(key + ": missing or unreadable regular file")
            return False
        digest = hashlib.sha256()
        size = 0
        complete = True
        with os.fdopen(descriptor, "rb") as incoming:
            before = os.fstat(incoming.fileno())
            if not stat.S_ISREG(before.st_mode):
                self.gaps.append(key + ": not a regular file")
                return False
            destination = self.output / key
            destination.parent.mkdir(mode=0o700, exist_ok=True)
            with destination.open("xb") as outgoing:
                os.fchmod(outgoing.fileno(), 0o600)
                while True:
                    try:
                        raw = incoming.read(min(1024**2, limit - size) + 1)
                    except OSError:
                        self.gaps.append(key + ": read failed; retained prefix only")
                        complete = False
                        break
                    if not raw:
                        break
                    keep = raw[:limit - size]
                    outgoing.write(keep)
                    digest.update(keep)
                    size += len(keep)
                    if len(keep) != len(raw):
                        self.gaps.append(key + ": byte limit exceeded; retained prefix only")
                        complete = False
                        break
                outgoing.flush()
                os.fsync(outgoing.fileno())
            after = os.fstat(incoming.fileno())
            fields = ("st_dev", "st_ino", "st_size", "st_mtime_ns", "st_ctime_ns")
            if any(getattr(before, field) != getattr(after, field) for field in fields) or (
                complete and size != after.st_size
            ):
                self.mismatches.append(key + ": source changed during snapshot")
                complete = False
        self.inventory[key] = dict(file=key, bytes=size, sha256=digest.hexdigest(), complete=complete)
        return complete


def _bind(document: dict[str, Any], observations: Any, receipt: dict[str, Any]) -> None:
    report._validate(observations)
    bundle._same(document["observations_sha256"], receipt["sha256"], "observation digest differs")
    for key in ("run_id", "profile", "started_at"):
        bundle._same(document[key], observations[key], "observation identity differs")
    for key in ("sut_vcpus", "sut_memory_bytes"):
        value = document["hardware"][key]["value"]
        if value is not None:
            bundle._same(value, observations["resources"][key], "allocation differs from observations")


def retain(output: Path, *, declaration: Path | None = None, observations: Path | None = None,
           evidence_dir: Path | None = None,
           max_evidence_bytes: int = DEFAULT_MAX_EVIDENCE_BYTES) -> dict[str, Any]:
    """Snapshot an explicit declaration, observations and opaque reference files."""
    report._integer(max_evidence_bytes, "max_evidence_bytes", 1)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    retained = _Retention(output)
    declaration_complete = retained.snapshot(declaration, "declaration.json", MAX_DECLARATION_BYTES)
    observations_complete = retained.snapshot(observations, "observations.json", report.MAX_INPUT_BYTES)
    documents: dict[str, Any] = {}
    if declaration_complete:
        retained.check("declaration JSON", lambda: documents.update(
            declaration=_read(output / "declaration.json", MAX_DECLARATION_BYTES)))
    if observations_complete:
        retained.check("observations JSON", lambda: documents.update(
            observations=_read(output / "observations.json", report.MAX_INPUT_BYTES)))
        if "observations" in documents:
            if not retained.check("observations schema", lambda: report._validate(documents["observations"])):
                del documents["observations"]
    document = documents.get("declaration")
    schema_valid = False
    bound = False
    fields = []
    if "declaration" in documents:
        schema_valid = retained.check("declaration schema", lambda: retained.gaps.extend(validate(document)))
    if schema_valid:
        for section, names in FIELDS.items():
            fields.extend({"field": section + "." + key, **document[section][key]} for key in names)
        if "observations" in documents:
            bound = retained.check("observation identity/digest/allocation binding", lambda: _bind(
                document, documents["observations"], retained.inventory["observations.json"]))
        remaining = max_evidence_bytes
        for entry in document["evidence"]:
            name = entry["id"] + ".bin"
            key = "evidence/" + name
            source = None if evidence_dir is None else evidence_dir.expanduser() / name
            complete = retained.snapshot(source, key, min(MAX_EVIDENCE_FILE_BYTES, remaining))
            actual = retained.inventory.get(key)
            if actual is not None:
                remaining -= actual["bytes"]
            if complete:
                retained.check("reference " + key, lambda entry=entry, actual=actual: bundle._same(
                    (entry["bytes"], entry["sha256"]), (actual["bytes"], actual["sha256"]),
                    "reference bytes/digest differs"))
    if not bound:
        retained.gaps.append("Observation binding is incomplete; this context is not tied to a valid observation snapshot.")
    status = "mismatch" if retained.mismatches else "incomplete" if retained.gaps else "review_required"
    result = {
        "schema_version": 1, "artifact_kind": "netsentry_slo_retained_context", "status": status,
        "context_complete": status == "review_required", "declaration_schema_valid": schema_valid,
        "observation_binding_complete": bound,
        "run_identity": {key: document[key] for key in ("run_id", "profile", "started_at", "observations_sha256")}
                        if schema_valid else None,
        "declared_fields": fields, "inventory": list(retained.inventory.values()),
        "gaps": retained.gaps, "mismatches": retained.mismatches, "checks_completed": retained.checked,
        "max_evidence_bytes": max_evidence_bytes,
        "context_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "execution_complete": False, "facts_verified": False,
        "comparability_established": False, "slo_compliance_asserted": False,
        "departmental_review_required": True,
        "limitations": [
            "Known values and hash-matched opaque references are declarations, not authenticated machine facts.",
            "Reference contents are retained without interpretation or execution; relevance and truth require review.",
            "Explicit false values remain declarations; profile targets and isolation adequacy are not evaluated here.",
            "Observations are schema/digest-bound, not reconstructed from live traffic or packet ledgers.",
            "No hardware discovery, shell command, service, benchmark or acceptance run is performed.",
            "R90-119 does not automatically consume context; a consumer must freshly verify identity and retained digests.",
            "Clock/durability, workload, physical isolation and process outcomes remain departmental qualification work.",
        ],
    }
    for directory in [path for path in output.iterdir() if path.is_dir()]:
        descriptor = os.open(directory, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(descriptor)
        finally:
            os.close(descriptor)
    report.write_report(output / "context.json", result)
    descriptor = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--declaration", type=Path, help="supplied version-1 declaration JSON")
    parser.add_argument("--observations", type=Path, help="exact observation JSON for this run")
    parser.add_argument("--evidence-dir", type=Path, help="directory containing ID.bin files from the catalog")
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--max-evidence-bytes", type=int, default=DEFAULT_MAX_EVIDENCE_BYTES)
    args = parser.parse_args(argv)
    try:
        result = retain(args.output_dir, declaration=args.declaration, observations=args.observations,
                        evidence_dir=args.evidence_dir, max_evidence_bytes=args.max_evidence_bytes)
    except (OSError, ValueError, RecursionError) as error:
        print(f"[slo-context] {error}; retain partial output and retry to a new directory", file=sys.stderr)
        return 2
    print(f"[slo-context] {result['status']}; departmental review required: {args.output_dir}")
    return {"review_required": 0, "mismatch": 1, "incomplete": 3}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
