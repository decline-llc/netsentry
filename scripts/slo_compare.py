#!/usr/bin/env python3
"""Compare declared conditions of retained SLO bundles without certifying capacity."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import sys
from typing import Any, Sequence

if __package__:
    from . import slo_bundle as bundle, slo_report as report
else:
    import slo_bundle as bundle
    import slo_report as report

SIDES = ("baseline", "candidate")
FILES = {role + "/" + name for role, names in bundle.SOURCES.items() for name in names}
MANIFEST_FIELDS = {
    "schema_version", "artifact_kind", "status", "bundle_complete", "snapshots_complete",
    "cross_checks_complete", "execution_complete", "slo_compliance_asserted",
    "departmental_review_required", "evidence_class", "max_retained_input_bytes",
    "bundle_source_sha256", "reporter_source_sha256", "inventory", "supplied_run_ids",
    "supplied_origins", "gaps", "mismatches", "checks_completed", "checks_not_performed",
    "reported_measurement_status", "limitations",
}


def _encoded(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True,
                      allow_nan=False).encode("ascii") + b"\n"


def _digest_rows(rows: Any) -> str:
    digest = hashlib.sha256()
    for row in rows:
        digest.update(_encoded(row))
    return digest.hexdigest()


def _read(path: Path) -> dict[str, Any]:
    # Called only for our successfully reconciled, bounded metadata snapshots.
    with path.open("rb") as stream:
        raw = stream.read(bundle.META_LIMIT + 1)
    bundle._require(len(raw) <= bundle.META_LIMIT, "metadata exceeds 1 MiB")
    value = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                       parse_constant=report._constant, parse_float=bundle._finite_float)
    bundle._require(isinstance(value, dict), "metadata must be an object")
    return value


def _manifest(document: dict[str, Any]) -> None:
    report._object(document, MANIFEST_FIELDS, "bundle manifest")
    bundle._require(type(document["schema_version"]) is int and document["schema_version"] == 1,
                    "unsupported bundle schema")
    bundle._require(document["artifact_kind"] == "netsentry_slo_evidence_bundle", "wrong artifact kind")
    bundle._require(document["evidence_class"] == "single_vm_supplied_artifact_snapshots", "wrong evidence class")
    for key in ("bundle_complete", "snapshots_complete", "cross_checks_complete", "execution_complete",
                "slo_compliance_asserted", "departmental_review_required"):
        report._boolean(document[key], key)
    bundle._require(document["execution_complete"] is False and document["slo_compliance_asserted"] is False
                    and document["departmental_review_required"] is True, "invalid authority flags")
    for key in ("gaps", "mismatches", "checks_completed", "checks_not_performed", "limitations"):
        bundle._require(isinstance(document[key], list) and all(isinstance(item, str) for item in document[key]),
                        "invalid diagnostic list")
    status = "mismatch" if document["mismatches"] else (
        "incomplete" if document["gaps"] or document["checks_not_performed"] else "review_required")
    bundle._require(document["status"] == status, "status contradicts diagnostics")
    bundle._require(document["bundle_complete"] == document["cross_checks_complete"] == (status == "review_required"),
                    "completion contradicts status")
    report._integer(document["max_retained_input_bytes"], "max_retained_input_bytes", 1)
    for key in ("bundle_source_sha256", "reporter_source_sha256"):
        bundle._hash(document[key])
    bundle._require(isinstance(document["inventory"], list), "inventory must be an array")
    seen = set()
    total = 0
    complete = True
    for entry in document["inventory"]:
        bundle._require(isinstance(entry, dict), "invalid inventory entry")
        name = entry.get("file")
        bundle._require(isinstance(name, str) and name in FILES and name not in seen, "unknown/duplicate file")
        seen.add(name)
        fields = {"file", "bytes", "sha256", "complete"}
        if "rows" in entry:
            bundle._require(name.endswith(".jsonl"), "rows on non-ledger")
            fields.add("rows")
            report._integer(entry["rows"], "rows")
        report._object(entry, fields, "inventory entry")
        total += report._integer(entry["bytes"], "bytes")
        bundle._hash(entry["sha256"])
        complete = report._boolean(entry["complete"], "complete") and complete
        if status == "review_required" and name.endswith(".jsonl"):
            bundle._require("rows" in entry, "completed ledger lacks row count")
    bundle._require(total <= document["max_retained_input_bytes"], "inventory exceeds declared budget")
    snapshots_complete = not document["gaps"] and complete and seen == FILES
    bundle._require(document["snapshots_complete"] == snapshots_complete, "snapshot completeness mismatch")
    if status == "review_required":
        bundle._require(snapshots_complete, "completed bundle lacks snapshots")
    ids = document["supplied_run_ids"]
    bundle._require(isinstance(ids, dict), "invalid run IDs")
    allowed_ids = {"sender/submission.json", "capture/summary.json", "engine/close.json",
                   "adapter/receipt.json", "adapter/manifest.json", "adapter/observations.json"}
    bundle._require(set(ids) <= allowed_ids, "unknown run-ID source")
    for value in ids.values():
        report._identifier(value, "run_id")
    bundle._require(isinstance(document["supplied_origins"], list), "invalid origins")
    for value in document["supplied_origins"]:
        report._timestamp(value)
    bundle._require(document["reported_measurement_status"] in
                    (None, "inconclusive", "measurement_failed", "review_required"), "invalid report status")


def _bind(original: dict[str, Any], current: dict[str, Any]) -> None:
    # Require both manifests to be complete before this binding can qualify a side.
    for key in ("status", "bundle_complete", "snapshots_complete", "cross_checks_complete",
                "supplied_run_ids", "supplied_origins", "reported_measurement_status"):
        bundle._same(original[key], current[key], "bundle identity/summary differs")
    left = {entry["file"]: entry for entry in original["inventory"]}
    right = {entry["file"]: entry for entry in current["inventory"]}
    bundle._same(left, right, "retained source inventory differs")


def _conditions(root: Path, original: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any]]:
    observations, _ = report.read_observations(root / "adapter/observations.json")
    report._validate(observations)
    sender = _read(root / "sender/submission.json")
    adapter = _read(root / "adapter/receipt.json")
    capture = _read(root / "capture/summary.json")
    engine = _read(root / "engine/close.json")
    conditions = {key: observations[key] for key in ("profile", "duration_seconds", "drain_seconds")}
    for section in ("policy", "resources", "measurement", "provenance"):
        for key, value in observations[section].items():
            conditions[section + "." + key] = value
    conditions["workload.offered_cohorts_sha256"] = _digest_rows(
        {key: row[key] for key in ("start_second", "phase", "offered_eligible", "offered_bytes", "observation_complete")}
        for row in observations["packet_windows"])
    alerts = observations["expected_alerts"]
    # Exclude arrival/durable outcomes and offer-time jitter from oracle identity.
    conditions["workload.expected_oracle_sha256"] = _digest_rows(
        sorted((row["event_id"], row["packet_id"], row["rule_id"]) for row in alerts))
    counts = [0] * len(observations["packet_windows"])
    for row in alerts:
        counts[row["offered_ns"] // (report.MINUTE * report.NS)] += 1
    conditions["workload.expected_alerts_per_minute_sha256"] = _digest_rows(counts)
    conditions["workload.expected_alert_count"] = len(alerts)
    conditions["sender.link"] = sender["link"]
    conditions["sender.eligible_byte_boundary"] = sender["eligible_byte_boundary"]
    for key in ("timestamp_type", "timestamp_precision", "direction", "udp_port"):
        conditions["capture." + key] = capture[key]
    for key in ("clock", "durable_boundary", "event_identity"):
        conditions["engine." + key] = engine[key]
    for key, value in (("sender", sender["sender_source_sha256"]),
                       ("adapter", adapter["adapter_source_sha256"]),
                       ("bundle", original["bundle_source_sha256"]),
                       ("reporter", original["reporter_source_sha256"])):
        conditions["tools." + key + "_source_sha256"] = value
    summary = report.summarize(observations)
    evidence = {
        "run_id": observations["run_id"], "started_at": observations["started_at"],
        "duration_seconds": observations["duration_seconds"], "drain_seconds": observations["drain_seconds"],
        "observed_through_ns": observations["observed_through_ns"],
        "measurement_status": summary["status"], "full_run": summary["full_run"],
        "coverage_gap_count": len(summary["coverage_gaps"]),
        "measurement_failure_count": len(summary["measurement_failures"]),
        "cohort_details": "adapter/observations.json and report/report.json in this side's reconciled directory",
    }
    return conditions, evidence


def compare(output: Path, *, baseline: Path | None = None, candidate: Path | None = None,
            max_bytes_per_side: int = bundle.DEFAULT_MAX_BYTES) -> dict[str, Any]:
    """Retain/reconcile two supplied bundles, then compare exact declared conditions."""
    report._integer(max_bytes_per_side, "max_bytes_per_side", 1)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    metadata = bundle._Bundle(output, 2 * bundle.META_LIMIT)
    sides: dict[str, Any] = {}
    projections: dict[str, dict[str, Any]] = {}
    evidence_gaps: list[str] = []
    invalid: list[str] = []
    for side, source in (("baseline", baseline), ("candidate", candidate)):
        (output / side).mkdir(mode=0o700)
        source = source.expanduser() if source is not None else None
        key = side + "/bundle.json"
        metadata.snapshot(None if source is None else source / "bundle.json", key)
        original = metadata.documents.get(key)
        manifest_valid = metadata.check(side + " original bundle schema", lambda: _manifest(original), (key,))
        root = output / side / "reconciled"
        current = bundle.reconcile(
            root, sender=None if source is None else source / "sender",
            capture=None if source is None else source / "capture/summary.json",
            engine=None if source is None else source / "engine",
            adapter=None if source is None else source / "adapter",
            summary=None if source is None else source / "report/report.json",
            max_bytes=max_bytes_per_side)
        sides[side] = {"original_manifest": metadata.inventory.get(key),
                       "reconciled_manifest": {"file": side + "/reconciled/bundle.json",
                                               "sha256": hashlib.sha256((root / "bundle.json").read_bytes()).hexdigest()},
                       "reconciled_status": current["status"], "metrics": None}
        if current["status"] == "mismatch":
            invalid.append(side + ": current reconciliation found mismatches; see retained bundle.json")
        elif current["status"] == "incomplete":
            evidence_gaps.append(side + ": current reconciliation is incomplete; see retained bundle.json")
        if not manifest_valid:
            continue
        if original["status"] == "mismatch":
            invalid.append(side + ": supplied bundle declares mismatches")
        elif original["status"] == "incomplete":
            evidence_gaps.append(side + ": supplied bundle declares incomplete evidence")
        if original["status"] != "review_required" or current["status"] != "review_required":
            continue
        if not metadata.check(side + " original inventory/identity binding", lambda: _bind(original, current)):
            continue
        conditions, metrics = _conditions(root, original)
        projections[side] = conditions
        sides[side]["metrics"] = metrics
    rows = []
    pair_gaps = []
    if set(projections) == set(SIDES):
        left, right = (sides[side]["metrics"] for side in SIDES)
        if left["run_id"] == right["run_id"]:
            pair_gaps.append("Run IDs are identical; this pair does not establish two distinct runs.")
        # Canonical whole-second UTC origins and integral nanosecond observation bounds.
        start = [int(report._timestamp(item["started_at"]).timestamp()) * report.NS for item in (left, right)]
        end = [start[index] + item["observed_through_ns"] for index, item in enumerate((left, right))]
        if max(start) < min(end):
            pair_gaps.append("Observation intervals overlap; shared-VM resource independence is not established.")
        for key in sorted(projections["baseline"]):
            a, b = projections["baseline"][key], projections["candidate"][key]
            rows.append({"field": key, "baseline": a, "candidate": b, "equal": _encoded(a) == _encoded(b)})
    else:
        evidence_gaps.append("Two complete, bound and currently reconciled bundles are required for field comparison.")
    invalid.extend(metadata.mismatches)
    evidence_gaps.extend(metadata.gaps)
    evidence_gaps.extend(metadata.skipped)
    differences = [row["field"] for row in rows if not row["equal"]]
    status = "invalid_evidence" if invalid else (
        "incomplete" if evidence_gaps or pair_gaps else "conditions_differ" if differences else "review_required")
    result = {
        "schema_version": 1, "artifact_kind": "netsentry_slo_declared_comparison",
        "comparison_policy": "exact_declared_repeatability_v1", "status": status,
        "declared_conditions_match": not differences if rows else None,
        "comparability_established": False, "slo_compliance_asserted": False,
        "regression_asserted": False, "departmental_review_required": True,
        "comparison_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "max_input_bytes_per_side": max_bytes_per_side, "sides": sides,
        "condition_comparisons": rows, "differing_fields": differences,
        "input_checks_completed": metadata.checked,
        "invalid_evidence": invalid, "evidence_gaps": evidence_gaps, "pair_qualification_gaps": pair_gaps,
        "qualification_gaps": [
            "CPU model/pinning, VM allocation, storage/NIC capabilities and actual isolation are absent from bundle schemas.",
            "Compiler/runtime/OS/kernel versions and source/config/rule provenance need independent supporting evidence.",
            "Oracle completeness, offered-load timing/distribution, generator headroom and full workload qualification require review.",
            "Raw packet join replay, clock calibration, physical arrival, persistence and process exit status remain unverified.",
            "Matching supplied declarations and digests do not authenticate independently measured runs or establish comparability.",
        ],
        "limitations": [
            "This is a conservative same-commit repeatability comparison; commit or tooling changes are differences, not regressions.",
            "Exact minute offered counts/bytes and alert-oracle identity are compared; sub-minute offer jitter is not compared.",
            "Run/origin and measured arrival/durable outcomes are not equality conditions; failed/inconclusive metrics remain visible.",
            "No ratio, statistical significance, capacity, SLO gate or hardware independence is inferred.",
            "Budgets cover retained source copies per side plus at most 1 MiB for each original manifest; outputs need additional disk.",
            "Current reconciliation is retained alongside original receipts; changing input tool versions requires departmental review.",
        ],
    }
    # reconcile() synced side directory entries; publish the pair manifest last.
    report.write_report(output / "comparison.json", result)
    descriptor = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", type=Path, help="supplied R90-118 bundle directory")
    parser.add_argument("--candidate", type=Path, help="supplied R90-118 bundle directory")
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--max-bytes-per-side", type=int, default=bundle.DEFAULT_MAX_BYTES)
    args = parser.parse_args(argv)
    try:
        result = compare(args.output_dir, baseline=args.baseline, candidate=args.candidate,
                         max_bytes_per_side=args.max_bytes_per_side)
    except (OSError, ValueError, RecursionError) as error:
        print(f"[slo-compare] {error}; retain partial output and retry to a new directory", file=sys.stderr)
        return 2
    print(f"[slo-compare] {result['status']}; departmental review required: {args.output_dir}")
    return {"review_required": 0, "conditions_differ": 1, "incomplete": 3, "invalid_evidence": 4}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
