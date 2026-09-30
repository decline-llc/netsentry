#!/usr/bin/env python3
"""Reconstruct supplied SLO adapter observations from retained raw ledgers."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import sqlite3
import sys
from typing import Any, Sequence

if __package__:
    from . import slo_bundle as bundle, slo_collect as collect, slo_report as report
else:
    import slo_bundle as bundle
    import slo_collect as collect
    import slo_report as report

DEFAULT_MAX_BYTES = bundle.DEFAULT_MAX_BYTES
INPUTS = ("manifest.json", "offered.jsonl", "events.jsonl")


def _comparisons(original: dict[str, Any], rebuilt: dict[str, Any]) -> list[dict[str, Any]]:
    rows = []
    for field in sorted(original):
        left, right = original[field], rebuilt[field]
        if field == "expected_alerts":
            # Both documents passed the reporter's unique event/packet-rule validation.
            left = sorted(left, key=lambda alert: alert["event_id"])
            right = sorted(right, key=lambda alert: alert["event_id"])
        rows.append({"field": field, "equal": json.dumps(left, sort_keys=True, allow_nan=False) ==
                     json.dumps(right, sort_keys=True, allow_nan=False)})
    return rows


def _rebuilt_proof(retained: bundle._Bundle, receipt: dict[str, Any], raw: bytes,
                   document: dict[str, Any]) -> None:
    entries = {entry["file"]: entry for entry in receipt["inputs_and_output"]}
    bundle._require(set(entries) == set(INPUTS) | {"observations.json"}, "unexpected rebuilt inventory")
    for name in INPUTS:
        current = entries[name]
        previous = retained.inventory["adapter/" + name]
        for field in ("bytes", "sha256") + (("rows",) if name.endswith(".jsonl") else ()):
            bundle._same(current[field], previous[field], "rebuilt inputs differ from retained sources")
    observed = entries["observations.json"]
    bundle._same((observed["bytes"], observed["sha256"]),
                 (len(raw), hashlib.sha256(raw).hexdigest()), "rebuilt observations differ from receipt")
    report._validate(document)
    bundle._same(receipt["run_id"], retained.documents["adapter/receipt.json"]["run_id"],
                 "rebuilt receipt run differs")
    bundle._manifest_observations(retained.documents["adapter/manifest.json"], document)


def reconstruct(output: Path, *, adapter: Path | None = None,
                max_bytes: int = DEFAULT_MAX_BYTES, scratch_dir: Path | None = None) -> dict[str, Any]:
    """Snapshot fixed inputs, replay the adapter and publish diagnostic provenance.

    No live execution, source repair, measurement certification or evidence deletion.
    Resource errors may leave partial files; inspect status and process exit together.
    """
    report._integer(max_bytes, "max_bytes", 1)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    retained = bundle._Bundle(output, max_bytes)
    source = adapter.expanduser() if adapter is not None else None
    for name in bundle.SOURCES["adapter"]:
        retained.snapshot(None if source is None else source / name, "adapter/" + name)
    documents = retained.documents
    receipt_valid = retained.check("original adapter receipt and inventory", lambda: bundle._adapter(
        retained, documents["adapter/receipt.json"]), ("adapter/receipt.json",))
    metadata_valid = retained.check("manifest and supplied observation schema/metadata", lambda:
        bundle._manifest_observations(documents["adapter/manifest.json"], documents["adapter/observations.json"]),
        ("adapter/manifest.json", "adapter/observations.json"))
    identity_valid = retained.check("original receipt/manifest run identity", lambda: bundle._same(
        documents["adapter/receipt.json"]["run_id"], documents["adapter/manifest.json"]["run_id"],
        "run identity differs"), ("adapter/receipt.json", "adapter/manifest.json"))
    ready = receipt_valid and metadata_valid and identity_valid and not (
        retained.gaps or retained.mismatches or retained.skipped)
    attempted = complete = False
    rows: list[dict[str, Any]] = []
    byte_match = None
    rebuilt_info = None
    errors: list[str] = []
    if ready:
        attempted = True
        root = output / "adapter"
        try:
            receipt = collect.collect(root / "manifest.json", root / "offered.jsonl", root / "events.jsonl",
                                      output / "rebuilt", scratch_dir.expanduser() if scratch_dir else None)
            document, raw = report.read_observations(output / "rebuilt/observations.json")
            complete = retained.check("rebuilt input/output receipt binding", lambda:
                                      _rebuilt_proof(retained, receipt, raw, document))
            raw_receipt = (output / "rebuilt/receipt.json").read_bytes()
            rebuilt_info = {"directory": "rebuilt", "inventory": receipt["inputs_and_output"],
                            "receipt": {"file": "rebuilt/receipt.json", "bytes": len(raw_receipt),
                                        "sha256": hashlib.sha256(raw_receipt).hexdigest()},
                            "adapter_source_sha256": receipt["adapter_source_sha256"]}
            if complete:
                rows = _comparisons(documents["adapter/observations.json"], document)
                original = retained.inventory["adapter/observations.json"]
                byte_match = (original["bytes"], original["sha256"]) == (len(raw), hashlib.sha256(raw).hexdigest())
        except (ValueError, TypeError, KeyError, IndexError, OverflowError, RecursionError,
                sqlite3.IntegrityError, sqlite3.DataError):
            retained.mismatches.append("raw ledger reconstruction: invalid or inconsistent evidence")
            complete = False
        except (OSError, sqlite3.Error):
            errors.append("reconstruction I/O or SQLite operation failed; inspect retained partial rebuilt directory")
            complete = False
    else:
        retained.skipped.append("raw ledger reconstruction requires complete consistent adapter sources")
    differences = [row["field"] for row in rows if not row["equal"]]
    if differences:
        retained.mismatches.append("rebuilt observations differ from supplied observations")
    status = "error" if errors else "mismatch" if retained.mismatches else (
        "incomplete" if retained.gaps or retained.skipped or not complete else "review_required")
    result = {
        "schema_version": 1, "artifact_kind": "netsentry_slo_ledger_reconstruction", "status": status,
        "reconstruction_attempted": attempted, "reconstruction_complete": complete,
        "observations_match": not differences if complete else None,
        "observation_bytes_match": byte_match if complete else None,
        "field_comparisons": rows, "differing_fields": differences,
        "run_id": documents["adapter/manifest.json"]["run_id"] if metadata_valid else None,
        "original_adapter_source_sha256": documents["adapter/receipt.json"]["adapter_source_sha256"]
                                           if receipt_valid else None,
        "max_retained_input_bytes": max_bytes,
        "inventory": list(retained.inventory.values()), "rebuilt": rebuilt_info,
        "gaps": retained.gaps, "mismatches": retained.mismatches, "errors": errors,
        "checks_completed": retained.checked, "checks_not_performed": retained.skipped,
        "reconstruction_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "adapter_source_sha256": hashlib.sha256(Path(collect.__file__).read_bytes()).hexdigest(),
        "reporter_source_sha256": hashlib.sha256(Path(report.__file__).read_bytes()).hexdigest(),
        "bundle_source_sha256": hashlib.sha256(Path(bundle.__file__).read_bytes()).hexdigest(),
        "execution_complete": False, "facts_verified": False, "slo_compliance_asserted": False,
        "comparability_established": False, "departmental_review_required": True,
        "limitations": [
            "Agreement means derivation from these retained ledgers only, not authentic or complete acquisitions.",
            "Live arrival, clock accuracy, physical durability, offered oracle and process exits require review.",
            "Current adapter interpretation is used; original tool digests are retained, not authenticated builds.",
            "Expected alerts compare in event-ID order; minute order and all observation values remain significant.",
            "Missing lifecycle values remain null; missing expected alerts are never filtered from comparison.",
            "The input budget excludes rebuilt copies, generated metadata and the temporary SQLite identity index.",
            "Disk/time/memory cost and runtime correctness remain unmeasured; no tests or acceptance run is performed.",
            "Partial rebuilt files may lack a complete receipt/inventory; preserve the directory and use a fresh retry path.",
            "Bundle and pair reconstruction mode performs fresh replay; this receipt alone cannot qualify evidence.",
        ],
    }
    for directory in (output / "adapter", output / "rebuilt"):
        if directory.is_dir():
            descriptor = os.open(directory, os.O_RDONLY | os.O_DIRECTORY)
            try:
                os.fsync(descriptor)
            finally:
                os.close(descriptor)
    report.write_report(output / "reconstruction.json", result)
    descriptor = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)
    return result


def integrate(retained: bundle._Bundle, *, max_bytes: int,
              scratch_dir: Path | None = None) -> tuple[dict[str, Any], list[str], list[str], list[str]]:
    """Freshly replay a reconciled bundle's own snapshots, never an old receipt."""
    info = {"status": "incomplete", "reconstruction_attempted": False,
            "reconstruction_complete": False, "source_binding_complete": False,
            "observations_match": None, "manifest": None}
    if retained.gaps or retained.mismatches or retained.skipped:
        return info, ["Reconstruction requires all existing bundle checks to complete."], [], []
    gaps: list[str] = []
    mismatches: list[str] = []
    errors: list[str] = []
    try:
        current = reconstruct(retained.output / "reconstruction", adapter=retained.output / "adapter",
                              max_bytes=max_bytes, scratch_dir=scratch_dir)
        raw = (retained.output / "reconstruction/reconstruction.json").read_bytes()
        info.update(status=current["status"], reconstruction_attempted=current["reconstruction_attempted"],
                    reconstruction_complete=current["reconstruction_complete"],
                    observations_match=current["observations_match"],
                    manifest={"file": "reconstruction/reconstruction.json", "bytes": len(raw),
                              "sha256": hashlib.sha256(raw).hexdigest()})
        gaps.extend("Reconstruction: " + item for item in current["gaps"] + current["checks_not_performed"])
        mismatches.extend("Reconstruction: " + item for item in current["mismatches"])
        errors.extend("Reconstruction: " + item for item in current["errors"])
        fresh = {entry["file"]: entry for entry in current["inventory"]}
        names = {"adapter/" + name for name in bundle.SOURCES["adapter"]}
        if set(fresh) != names or any(not fresh[key]["complete"] for key in names):
            gaps.append("Reconstruction source inventory is incomplete; enclosing bundle binding unavailable.")
        else:
            try:
                bundle._same({key: retained.inventory[key] for key in names}, fresh,
                             "reconstruction sources differ from enclosing bundle")
            except (ValueError, KeyError):
                mismatches.append("Reconstruction source inventory differs from enclosing bundle.")
            else:
                info["source_binding_complete"] = True
    except (OSError, sqlite3.Error):
        errors.append("Reconstruction I/O or SQLite failure; retain partial reconstruction directory.")
    info["status"] = "error" if errors else "mismatch" if mismatches else (
        "incomplete" if gaps or not info["reconstruction_complete"] else "review_required")
    return info, gaps, mismatches, errors


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--adapter", type=Path, help="supplied adapter package; omission records incomplete evidence")
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--max-bytes", type=int, default=DEFAULT_MAX_BYTES, help="total retained input byte budget")
    parser.add_argument("--scratch-dir", type=Path, help="existing directory for the temporary SQLite identity index")
    args = parser.parse_args(argv)
    try:
        result = reconstruct(args.output_dir, adapter=args.adapter, max_bytes=args.max_bytes,
                             scratch_dir=args.scratch_dir)
    except (OSError, ValueError, sqlite3.Error, RecursionError) as error:
        print(f"[slo-reconstruct] {error}; inspect partial output and retry to a new directory", file=sys.stderr)
        return 2
    print(f"[slo-reconstruct] {result['status']}; departmental review required: {args.output_dir}")
    return {"review_required": 0, "mismatch": 1, "error": 2, "incomplete": 3}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
