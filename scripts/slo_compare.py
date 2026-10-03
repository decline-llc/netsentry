#!/usr/bin/env python3
"""Compare declared conditions of retained SLO bundles without certifying capacity."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import sys
from typing import Any, Sequence

if __package__:
    from . import slo_bundle as bundle, slo_report as report, slo_context_compare as context_compare
else:
    import slo_bundle as bundle
    import slo_report as report
    import slo_context_compare as context_compare

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

RECONSTRUCTION_FIELDS = {"reconstruction_policy", "reconstruction", "errors",
                         "max_reconstruction_bytes", "reconstruction_source_sha256"}


SENDER_RECONSTRUCTION_FIELDS = {"sender_reconstruction_policy", "sender_reconstruction", "errors",
                                "max_sender_reconstruction_bytes", "sender_reconstruction_source_sha256",
                                "adapter_reconstruction_required"}


def _replay_summary(document: dict[str, Any], key: str, match_field: str, path: str) -> None:
    info = document[key]
    report._object(info, {"status", "reconstruction_attempted", "reconstruction_complete",
                          "source_binding_complete", match_field, "manifest"}, "reconstruction summary")
    bundle._require(info["status"] in ("review_required", "mismatch", "incomplete", "error"), "invalid replay status")
    for field in ("reconstruction_attempted", "reconstruction_complete", "source_binding_complete"):
        report._boolean(info[field], field)
    if info["reconstruction_complete"]:
        bundle._require(info["reconstruction_attempted"] is True, "completion without replay")
        report._boolean(info[match_field], match_field)
        if match_field == "sender_records_match":
            bundle._require(info[match_field] is True, "sender completion without matching records")
    else:
        bundle._require(info[match_field] is None, "equality without replay completion")
    manifest = info["manifest"]
    if manifest is not None:
        report._object(manifest, {"file", "bytes", "sha256"}, "reconstruction manifest")
        bundle._require(manifest["file"] == path, "unknown replay manifest path")
        bundle._require(report._integer(manifest["bytes"], "bytes", 1) <= bundle.META_LIMIT, "replay manifest too large")
        bundle._hash(manifest["sha256"])
    else:
        bundle._require(not info["reconstruction_complete"] and not info["source_binding_complete"],
                        "completion/binding without manifest")
    if info["status"] == "review_required":
        bundle._require(info["source_binding_complete"] and info["reconstruction_complete"]
                        and info[match_field] is True, "replay success lacks bound agreement")
    if info[match_field] is False:
        bundle._require(info["status"] in ("mismatch", "error"), "observation difference lacks failure")
    if info["status"] == "mismatch":
        bundle._require(bool(document["mismatches"]), "missing replay mismatch diagnostic")
    if info["status"] == "incomplete":
        bundle._require(bool(document["gaps"] or document["checks_not_performed"]), "missing replay gap")
    if document["status"] == "review_required":
        bundle._require(info["status"] == "review_required", "bundle success without replay success")


def _reconstruction_manifest(document: dict[str, Any], *, sender: bool, adapter: bool) -> None:
    errors = document["errors"]
    bundle._require(isinstance(errors, list) and all(isinstance(item, str) for item in errors), "invalid errors")
    selected = []
    if adapter:
        bundle._require(document["reconstruction_policy"] == "retained_adapter_replay_v1", "unsupported replay policy")
        report._integer(document["max_reconstruction_bytes"], "max_reconstruction_bytes", 1)
        bundle._hash(document["reconstruction_source_sha256"])
        _replay_summary(document, "reconstruction", "observations_match", "reconstruction/reconstruction.json")
        selected.append(document["reconstruction"])
    if sender:
        bundle._require(document["sender_reconstruction_policy"] == "retained_sender_replay_v1",
                        "unsupported sender replay policy")
        report._integer(document["max_sender_reconstruction_bytes"], "max_sender_reconstruction_bytes", 1)
        bundle._hash(document["sender_reconstruction_source_sha256"])
        _replay_summary(document, "sender_reconstruction", "sender_records_match",
                       "sender-reconstruction/sender-reconstruction.json")
        selected.append(document["sender_reconstruction"])
    bundle._require(bool(errors) == any(info["status"] == "error" for info in selected),
                    "replay errors contradict selected operation statuses")


def _encoded(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True,
                      allow_nan=False).encode("ascii") + b"\n"


def _digest_rows(rows: Any) -> str:
    digest = hashlib.sha256()
    for row in rows:
        digest.update(_encoded(row))
    return digest.hexdigest()


def _read(path: Path, key: str, entry: Any) -> dict[str, Any]:
    bundle._require(isinstance(entry, dict) and entry.get("complete") is True
                    and entry.get("file") == key, "receipt requires a complete captured inventory")
    expected_size = report._integer(entry.get("bytes"), "receipt bytes")
    bundle._require(expected_size <= bundle.META_LIMIT, "metadata exceeds 1 MiB")
    expected_hash = entry.get("sha256")
    bundle._hash(expected_hash)
    flags = tuple(getattr(os, name, None) for name in ("O_NOFOLLOW", "O_NONBLOCK"))
    bundle._require(all(type(flag) is int and flag > 0 for flag in flags),
                    "non-following/nonblocking receipt admission unavailable")
    fields = ("st_dev", "st_ino", "st_size", "st_mtime_ns", "st_ctime_ns")
    descriptor = os.open(path, os.O_RDONLY | flags[0] | flags[1])
    try:
        stream = os.fdopen(descriptor, "rb")
    except BaseException:
        os.close(descriptor)
        raise
    with stream:
        before = os.fstat(stream.fileno())
        bundle._require(stat.S_ISREG(before.st_mode), "receipt input must be a regular file")
        bundle._require(all(type(getattr(before, field, None)) is int for field in fields),
                        "stable receipt metadata unavailable")
        bundle._require(before.st_size == expected_size, "receipt size differs from inventory")
        before_values = tuple(getattr(before, field) for field in fields)
        raw = stream.read(expected_size + 1)
        bundle._require(len(raw) <= expected_size, "receipt bytes exceed inventory")
        after = os.fstat(stream.fileno())
        bundle._require(all(type(getattr(after, field, None)) is int for field in fields),
                        "stable receipt metadata unavailable")
        bundle._require(before_values == tuple(getattr(after, field) for field in fields),
                        "receipt input changed while reading")
        bundle._require(len(raw) == expected_size and hashlib.sha256(raw).hexdigest() == expected_hash,
                        "receipt bytes differ from captured inventory")
    value = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                       parse_constant=report._constant, parse_float=bundle._finite_float)
    bundle._require(isinstance(value, dict), "metadata must be an object")
    return value


def _manifest(document: dict[str, Any], allow_reconstruction: bool = False,
              allow_sender: bool = False) -> None:
    bundle._require(isinstance(document, dict), "bundle manifest must be an object")
    version = document.get("schema_version")
    bundle._require(type(version) is int and version in (1, 2, 3), "unsupported bundle schema")
    sender = version == 3
    adapter = version == 2
    if sender:
        adapter = report._boolean(document.get("adapter_reconstruction_required"), "adapter_reconstruction_required")
        bundle._require(allow_sender, "v3 requires explicit sender reconstruction mode")
    bundle._require(not adapter or allow_reconstruction, "bundle requires explicit adapter reconstruction mode")
    fields = MANIFEST_FIELDS | (SENDER_RECONSTRUCTION_FIELDS if sender else set())
    if adapter:
        fields |= RECONSTRUCTION_FIELDS
    report._object(document, fields, "bundle manifest")
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
    if sender or adapter:
        _reconstruction_manifest(document, sender=sender, adapter=adapter)
    status = "error" if (sender or adapter) and document["errors"] else "mismatch" if document["mismatches"] else (
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
    inventory = {entry["file"]: entry for entry in original["inventory"]}
    sender = _read(root / "sender/submission.json", "sender/submission.json",
                   inventory.get("sender/submission.json"))
    adapter = _read(root / "adapter/receipt.json", "adapter/receipt.json",
                    inventory.get("adapter/receipt.json"))
    capture = _read(root / "capture/summary.json", "capture/summary.json",
                    inventory.get("capture/summary.json"))
    engine = _read(root / "engine/close.json", "engine/close.json",
                   inventory.get("engine/close.json"))
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
            max_bytes_per_side: int = bundle.DEFAULT_MAX_BYTES,
            baseline_context: Path | None = None, candidate_context: Path | None = None,
            require_context: bool = False,
            max_context_evidence_bytes: int = context_compare.DEFAULT_MAX_EVIDENCE_BYTES,
            reconstruct_ledgers: bool = False, max_reconstruction_bytes: int = bundle.DEFAULT_MAX_BYTES,
            scratch_dir: Path | None = None, reconstruct_sender: bool = False,
            max_sender_reconstruction_bytes: int = bundle.DEFAULT_MAX_BYTES) -> dict[str, Any]:
    """Retain/reconcile two supplied bundles, then compare exact declared conditions."""
    report._integer(max_bytes_per_side, "max_bytes_per_side", 1)
    report._integer(max_context_evidence_bytes, "max_context_evidence_bytes", 1)
    report._boolean(require_context, "require_context")
    report._boolean(reconstruct_ledgers, "reconstruct_ledgers")
    report._integer(max_reconstruction_bytes, "max_reconstruction_bytes", 1)
    report._boolean(reconstruct_sender, "reconstruct_sender")
    report._integer(max_sender_reconstruction_bytes, "max_sender_reconstruction_bytes", 1)
    context_enabled = require_context or baseline_context is not None or candidate_context is not None
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    metadata = bundle._Bundle(output, (4 if context_enabled else 2) * bundle.META_LIMIT)
    sides: dict[str, Any] = {}
    projections: dict[str, dict[str, Any]] = {}
    bundle_identities: dict[str, dict[str, Any]] = {}
    observation_sizes: dict[str, int] = {}
    evidence_gaps: list[str] = []
    invalid: list[str] = []
    errors: list[str] = []
    for side, source in (("baseline", baseline), ("candidate", candidate)):
        (output / side).mkdir(mode=0o700)
        source = source.expanduser() if source is not None else None
        key = side + "/bundle.json"
        metadata.snapshot(None if source is None else source / "bundle.json", key)
        original = metadata.documents.get(key)
        manifest_valid = metadata.check(side + " original bundle schema", lambda: _manifest(
            original, reconstruct_ledgers, reconstruct_sender), (key,))
        root = output / side / "reconciled"
        current = bundle.reconcile(
            root, sender=None if source is None else source / "sender",
            capture=None if source is None else source / "capture/summary.json",
            engine=None if source is None else source / "engine",
            adapter=None if source is None else source / "adapter",
            summary=None if source is None else source / "report/report.json",
            max_bytes=max_bytes_per_side, reconstruct_ledgers=reconstruct_ledgers,
            max_reconstruction_bytes=max_reconstruction_bytes, scratch_dir=scratch_dir,
            reconstruct_sender=reconstruct_sender, max_sender_reconstruction_bytes=max_sender_reconstruction_bytes)
        sides[side] = {"original_manifest": metadata.inventory.get(key),
                       "reconciled_manifest": {"file": side + "/reconciled/bundle.json",
                                               "sha256": hashlib.sha256((root / "bundle.json").read_bytes()).hexdigest()},
                       "reconciled_status": current["status"], "metrics": None}
        for enabled, key_name in ((reconstruct_ledgers, "reconstruction"),
                                  (reconstruct_sender, "sender_reconstruction")):
            if enabled:
                info = dict(current[key_name])
                if info["manifest"] is not None:
                    info["manifest"] = dict(info["manifest"], file=side + "/reconciled/" + info["manifest"]["file"])
                sides[side][key_name] = info
        if current["status"] == "error":
            errors.append(side + ": current reconstruction operation failed; see retained bundle.json")
        elif current["status"] == "mismatch":
            invalid.append(side + ": current reconciliation found mismatches; see retained bundle.json")
        elif current["status"] == "incomplete":
            evidence_gaps.append(side + ": current reconciliation is incomplete; see retained bundle.json")
        if reconstruct_ledgers or reconstruct_sender:
            if current["mismatches"] and current["status"] != "mismatch":
                invalid.append(side + ": current reconciliation also records mismatches")
            if (current["gaps"] or current["checks_not_performed"]) and current["status"] != "incomplete":
                evidence_gaps.append(side + ": current reconciliation also records gaps")
        if not manifest_valid:
            continue
        if original["status"] == "error":
            errors.append(side + ": supplied bundle records a reconstruction operation error")
        elif original["status"] == "mismatch":
            invalid.append(side + ": supplied bundle declares mismatches")
        elif original["status"] == "incomplete":
            evidence_gaps.append(side + ": supplied bundle declares incomplete evidence")
        if reconstruct_ledgers or reconstruct_sender:
            if original["mismatches"] and original["status"] != "mismatch":
                invalid.append(side + ": supplied bundle also records mismatches")
            if (original["gaps"] or original["checks_not_performed"]) and original["status"] != "incomplete":
                evidence_gaps.append(side + ": supplied bundle also records gaps")
        if original["status"] != "review_required" or current["status"] != "review_required":
            continue
        if not metadata.check(side + " original inventory/identity binding", lambda: _bind(original, current)):
            continue
        conditions, metrics = _conditions(root, original)
        projections[side] = conditions
        sides[side]["metrics"] = metrics
        observed = next(item for item in current["inventory"] if item["file"] == "adapter/observations.json")
        bundle_identities[side] = {"run_id": metrics["run_id"], "profile": conditions["profile"],
                                   "started_at": metrics["started_at"], "observations_sha256": observed["sha256"]}
        observation_sizes[side] = observed["bytes"]
    context_rows = []
    if context_enabled:
        context_cells = {}
        for side, source in (("baseline", baseline_context), ("candidate", candidate_context)):
            info, cells, gaps, context_errors = context_compare.consume(
                metadata, side, source, bundle_identities.get(side), observation_sizes.get(side),
                max_context_evidence_bytes)
            sides[side]["context"] = info
            context_cells[side] = cells
            evidence_gaps.extend(gaps)
            invalid.extend(context_errors)
        context_rows = context_compare.compare_fields(context_cells)
        if any(row["equal"] is None for row in context_rows):
            evidence_gaps.append("Context comparison has unknown, unsupported or unavailable fields.")
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
    differences = [row["field"] for row in rows + context_rows if row["equal"] is False]
    status = "error" if errors else "invalid_evidence" if invalid else (
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
    if context_enabled:
        context_differences = [row["field"] for row in context_rows if row["equal"] is False]
        context_unknown = any(row["equal"] is None for row in context_rows)
        result.update(
            schema_version=2, comparison_policy="exact_declared_repeatability_with_context_v1",
            declared_conditions_match=False if differences else None if not rows or context_unknown else True,
            context_conditions_match=False if context_differences else None if context_unknown else True,
            context_comparisons=context_rows, facts_verified=False,
            max_context_evidence_bytes_per_side=max_context_evidence_bytes,
            context_consumer_source_sha256=hashlib.sha256(Path(context_compare.__file__).read_bytes()).hexdigest())
        result["qualification_gaps"][0] = (
            "Context declarations are retained when supplied; CPU/pinning, VM/storage/NIC and actual isolation facts remain unverified.")
        result["limitations"].extend([
            "Context values compare only with original/current source proof and exact bundle binding; null never counts as a match.",
            "Context evidence IDs/digests are provenance, not property equality conditions; reference relevance remains unverified.",
            "Context mode adds per-side evidence budgets, a 1 MiB declaration, 64 MiB observations and 1 MiB original receipt.",
        ])
        # Context retention syncs its own parent; persist the context entry in each side as well.
        for side in SIDES:
            descriptor = os.open(output / side, os.O_RDONLY | os.O_DIRECTORY)
            try:
                os.fsync(descriptor)
            finally:
                os.close(descriptor)
    if reconstruct_ledgers:
        result.update(schema_version=3,
                      comparison_policy=("exact_declared_repeatability_with_context_and_reconstruction_v1"
                                         if context_enabled else "exact_declared_repeatability_with_reconstruction_v1"),
                      reconstruction_required=True, errors=errors,
                      max_reconstruction_bytes_per_side=max_reconstruction_bytes)
        result["qualification_gaps"][3] = (
            "Fresh raw-ledger replay is required; clocks, physical arrival, persistence and process exits remain unverified.")
        result["limitations"].append(
            "Each side must freshly replay and bind retained adapter sources; old reconstruction summaries cannot qualify a side.")
    if reconstruct_sender:
        modes = ("context_and_" if context_enabled else "") + ("reconstruction_and_" if reconstruct_ledgers else "")
        result.update(schema_version=4,
                      comparison_policy="exact_declared_repeatability_with_" + modes + "sender_reconstruction_v1",
                      sender_reconstruction_required=True, adapter_reconstruction_required=reconstruct_ledgers,
                      errors=errors, max_sender_reconstruction_bytes_per_side=max_sender_reconstruction_bytes)
        result["qualification_gaps"][2] = (
            "Fresh sender-source derivation is required; actual offered load, generator headroom and oracle truth remain unverified.")
        result["limitations"].append(
            "Both sides require fresh sender replay bound to four retained files; old receipts cannot qualify a side.")
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
    parser.add_argument("--baseline-context", type=Path, help="retained context package for the baseline run")
    parser.add_argument("--candidate-context", type=Path, help="retained context package for the candidate run")
    parser.add_argument("--require-context", action="store_true", help="require both context packages even when absent")
    parser.add_argument("--max-context-evidence-bytes", type=int, default=context_compare.DEFAULT_MAX_EVIDENCE_BYTES)
    parser.add_argument("--reconstruct-ledgers", action="store_true", help="require fresh raw-ledger replay for both sides")
    parser.add_argument("--max-reconstruction-bytes", type=int, default=bundle.DEFAULT_MAX_BYTES)
    parser.add_argument("--reconstruct-sender", action="store_true", help="require fresh sender replay for both sides")
    parser.add_argument("--max-sender-reconstruction-bytes", type=int, default=bundle.DEFAULT_MAX_BYTES)
    parser.add_argument("--scratch-dir", type=Path, help="temporary reconstruction SQLite index directory")
    args = parser.parse_args(argv)
    try:
        result = compare(args.output_dir, baseline=args.baseline, candidate=args.candidate,
                         max_bytes_per_side=args.max_bytes_per_side,
                         baseline_context=args.baseline_context, candidate_context=args.candidate_context,
                         require_context=args.require_context,
                         max_context_evidence_bytes=args.max_context_evidence_bytes,
                         reconstruct_ledgers=args.reconstruct_ledgers,
                         max_reconstruction_bytes=args.max_reconstruction_bytes, scratch_dir=args.scratch_dir,
                         reconstruct_sender=args.reconstruct_sender,
                         max_sender_reconstruction_bytes=args.max_sender_reconstruction_bytes)
    except (OSError, ValueError, RecursionError) as error:
        print(f"[slo-compare] {error}; retain partial output and retry to a new directory", file=sys.stderr)
        return 2
    print(f"[slo-compare] {result['status']}; departmental review required: {args.output_dir}")
    return {"review_required": 0, "conditions_differ": 1, "error": 2, "incomplete": 3, "invalid_evidence": 4}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
