"""Bounded retained-context consumption for the SLO pair comparator."""

from __future__ import annotations

import hashlib
from pathlib import Path
from typing import Any

if __package__:
    from . import slo_bundle as bundle, slo_context as context, slo_report as report
else:
    import slo_bundle as bundle
    import slo_context as context
    import slo_report as report

DEFAULT_MAX_EVIDENCE_BYTES = context.DEFAULT_MAX_EVIDENCE_BYTES

MANIFEST_FIELDS = {
    "schema_version", "artifact_kind", "status", "context_complete", "declaration_schema_valid",
    "observation_binding_complete", "run_identity", "declared_fields", "inventory", "gaps", "mismatches",
    "checks_completed", "max_evidence_bytes", "context_source_sha256", "execution_complete", "facts_verified",
    "comparability_established", "slo_compliance_asserted", "departmental_review_required", "limitations",
}
FIELD_TYPES = {section + "." + key: kind for section, fields in context.FIELDS.items() for key, kind in fields.items()}


def _manifest(document: Any) -> None:
    report._object(document, MANIFEST_FIELDS, "context receipt")
    bundle._require(type(document["schema_version"]) is int and document["schema_version"] == 1,
                    "unsupported context receipt schema")
    bundle._require(document["artifact_kind"] == "netsentry_slo_retained_context", "wrong context receipt kind")
    for key in ("context_complete", "declaration_schema_valid", "observation_binding_complete"):
        report._boolean(document[key], key)
    for key in ("execution_complete", "facts_verified", "comparability_established", "slo_compliance_asserted"):
        bundle._require(document[key] is False, "unsupported authority claim")
    bundle._require(document["departmental_review_required"] is True, "review must remain required")
    for key in ("gaps", "mismatches", "checks_completed", "limitations"):
        bundle._require(isinstance(document[key], list) and all(isinstance(item, str) for item in document[key]),
                        "invalid diagnostics")
    status = "mismatch" if document["mismatches"] else "incomplete" if document["gaps"] else "review_required"
    bundle._require(document["status"] == status and document["context_complete"] == (status == "review_required"),
                    "inconsistent context status")
    report._integer(document["max_evidence_bytes"], "max_evidence_bytes", 1)
    bundle._hash(document["context_source_sha256"])
    identity = document["run_identity"]
    if document["declaration_schema_valid"]:
        report._object(identity, {"run_id", "profile", "started_at", "observations_sha256"}, "run identity")
        report._identifier(identity["run_id"], "run_id")
        bundle._require(isinstance(identity["profile"], str) and identity["profile"] in report.PROFILES,
                        "invalid profile")
        report._timestamp(identity["started_at"])
        bundle._hash(identity["observations_sha256"])
    else:
        bundle._require(identity is None and document["observation_binding_complete"] is False,
                        "identity without valid declaration")
    rows = document["declared_fields"]
    bundle._require(isinstance(rows, list) and len(rows) <= len(FIELD_TYPES), "invalid fields")
    seen = set()
    for row in rows:
        report._object(row, {"field", "value", "evidence_ids"}, "declared field")
        name = row["field"]
        bundle._require(isinstance(name, str) and name in FIELD_TYPES and name not in seen, "unknown/duplicate field")
        seen.add(name)
        refs = row["evidence_ids"]
        bundle._require(isinstance(refs, list) and len(refs) <= context.MAX_REFERENCES, "invalid references")
        checked = [context._evidence_id(item) for item in refs]
        bundle._require(len(set(checked)) == len(checked), "duplicate field reference")
        value, kind = row["value"], FIELD_TYPES[name]
        if value is not None:
            if kind == "text":
                context._text(value)
            elif kind == "positive_integer":
                report._integer(value, name, 1)
            elif kind == "boolean":
                report._boolean(value, name)
            else:
                bundle._require(isinstance(value, str) and value in kind, "invalid enumeration")
        if status == "review_required":
            bundle._require(value is not None and bool(refs), "complete receipt has unknown/unsupported fields")
    bundle._require(seen == (set(FIELD_TYPES) if document["declaration_schema_valid"] else set()),
                    "field coverage contradicts declaration status")
    inventory = document["inventory"]
    bundle._require(isinstance(inventory, list) and len(inventory) <= context.MAX_REFERENCES + 2, "invalid inventory")
    names = set()
    total = 0
    complete_files = set()
    for item in inventory:
        report._object(item, {"file", "bytes", "sha256", "complete"}, "context inventory entry")
        name = item["file"]
        bundle._require(isinstance(name, str) and name not in names, "invalid/duplicate inventory file")
        names.add(name)
        size = report._integer(item["bytes"], "bytes")
        bundle._hash(item["sha256"])
        if report._boolean(item["complete"], "complete"):
            complete_files.add(name)
        if name == "declaration.json":
            bundle._require(size <= context.MAX_DECLARATION_BYTES, "declaration exceeds limit")
        elif name == "observations.json":
            bundle._require(size <= report.MAX_INPUT_BYTES, "observations exceed limit")
        else:
            bundle._require(name.startswith("evidence/") and name.endswith(".bin"), "unknown inventory path")
            context._evidence_id(name[len("evidence/"):-len(".bin")])
            bundle._require(size <= context.MAX_EVIDENCE_FILE_BYTES, "reference exceeds limit")
            total += size
    bundle._require(total <= document["max_evidence_bytes"], "evidence exceeds budget")
    if status == "review_required":
        expected = {"declaration.json", "observations.json"} | {
            "evidence/" + ref + ".bin" for row in rows for ref in row["evidence_ids"]}
        bundle._require(document["observation_binding_complete"] is True and complete_files == names == expected,
                        "complete receipt lacks bound sources")


def _proof(original: dict[str, Any], current: dict[str, Any]) -> set[str]:
    bundle._same(original["run_identity"], current["run_identity"], "run identity differs")
    bundle._same({row["field"]: row for row in original["declared_fields"]},
                 {row["field"]: row for row in current["declared_fields"]}, "declared values/references differ")
    left = {item["file"]: item for item in original["inventory"]}
    right = {item["file"]: item for item in current["inventory"]}
    proven = set()
    for name in left.keys() & right.keys():
        if left[name]["complete"] and right[name]["complete"]:
            bundle._same(left[name], right[name], "original/current source inventory differs")
            proven.add(name)
    return proven


def consume(metadata: bundle._Bundle, side: str, source: Path | None,
            bundle_identity: dict[str, Any] | None, observation_bytes: int | None,
            max_evidence_bytes: int) -> tuple[dict[str, Any], dict[str, Any], list[str], list[str]]:
    """Retain one context and expose only individually supported, bundle-bound fields."""
    root = metadata.output / side / "context"
    root.mkdir(mode=0o700)
    source = source.expanduser() if source is not None else None
    key = side + "/context/context.json"
    metadata.snapshot(None if source is None else source / "context.json", key)
    original = metadata.documents.get(key)
    valid = metadata.check(side + " context receipt schema", lambda: _manifest(original), (key,))
    current = context.retain(
        root / "reconciled", declaration=None if source is None else source / "declaration.json",
        observations=None if source is None else source / "observations.json",
        evidence_dir=None if source is None else source / "evidence", max_evidence_bytes=max_evidence_bytes)
    raw = (root / "reconciled/context.json").read_bytes()
    info = {"original_manifest": metadata.inventory.get(key), "reconciled_status": current["status"],
            "reconciled_manifest": {"file": side + "/context/reconciled/context.json",
                                    "bytes": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
            "bound_to_bundle": False}
    gaps: list[str] = []
    invalid: list[str] = []
    cells: dict[str, Any] = {}
    for label, receipt in (("current", current), ("original", original if valid else None)):
        if receipt is None:
            continue
        if receipt["status"] == "mismatch":
            invalid.append(side + ": " + label + " context has mismatches; inspect retained context.json")
        elif receipt["status"] == "incomplete":
            gaps.append(side + ": " + label + " context is incomplete; inspect retained context.json")
    if not valid or invalid:
        return info, cells, gaps, invalid
    if not all(item["declaration_schema_valid"] and item["observation_binding_complete"]
               for item in (original, current)):
        gaps.append(side + ": context declaration/observation binding is incomplete")
        return info, cells, gaps, invalid
    proven: set[str] = set()
    if not metadata.check(side + " context original/source binding", lambda: proven.update(_proof(original, current))):
        return info, cells, gaps, invalid
    if not {"declaration.json", "observations.json"} <= proven or bundle_identity is None:
        gaps.append(side + ": complete original context sources and a reconciled bundle are required for binding")
        return info, cells, gaps, invalid
    actual_observations = next(item for item in current["inventory"] if item["file"] == "observations.json")

    def bind() -> None:
        bundle._same(current["run_identity"], bundle_identity, "context belongs to a different bundle/run")
        bundle._same(actual_observations["bytes"], observation_bytes, "observation byte count differs")

    if not metadata.check(side + " context/bundle identity and observation bytes", bind):
        return info, cells, gaps, invalid
    info["bound_to_bundle"] = True
    for row in current["declared_fields"]:
        refs = row["evidence_ids"]
        state = "unknown" if row["value"] is None else (
            "known" if refs and all("evidence/" + ref + ".bin" in proven for ref in refs) else "unsupported")
        cells[row["field"]] = {"value": row["value"], "state": state, "evidence_ids": refs}
        if state != "known":
            gaps.append(side + ": context." + row["field"] + " is " + state)
    return info, cells, gaps, invalid


def compare_fields(sides: dict[str, dict[str, Any]]) -> list[dict[str, Any]]:
    rows = []
    for name in sorted(FIELD_TYPES):
        left, right = (sides.get(side, {}).get(name, {"value": None, "state": "unavailable", "evidence_ids": []})
                       for side in ("baseline", "candidate"))
        equal = None
        if left["state"] == right["state"] == "known":
            # Values already have strict, matching schema types; references are provenance, not properties.
            equal = left["value"] == right["value"]
        rows.append({"field": "context." + name, "baseline": left, "candidate": right, "equal": equal})
    return rows
