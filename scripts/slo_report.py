#!/usr/bin/env python3
"""Summarize supplied SLO observations; never certify deployment compliance."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import re
import sys
import tempfile
from pathlib import Path
from typing import Any, Sequence

NS = 1_000_000_000
MINUTE = 60
DEADLINE_NS = 180_000_000
MAX_INPUT_BYTES = 64 * 1024 * 1024
PROFILES = {
    "staging": {"vcpus": 2, "memory_bytes": 4 * 1024**3, "rules": 2_000,
                "sustained_bps": 100_000_000, "burst_bps": 250_000_000},
    "prod": {"vcpus": 8, "memory_bytes": 16 * 1024**3, "rules": 20_000,
             "sustained_bps": 3_000_000_000, "burst_bps": 7_000_000_000},
}
LOSS_LIMIT_PPM = {"sustained": 500, "burst": 2_000}


class EvidenceError(ValueError):
    """Malformed or internally inconsistent measurement input."""


def _object(value: Any, fields: set[str], label: str) -> dict[str, Any]:
    if not isinstance(value, dict) or set(value) != fields:
        raise EvidenceError(f"{label}: fields do not match schema v1")
    return value


def _integer(value: Any, label: str, minimum: int = 0) -> int:
    if type(value) is not int or not minimum <= value <= 2**63 - 1:
        raise EvidenceError(f"{label}: expected integer in [{minimum}, 2^63-1]")
    return value


def _boolean(value: Any, label: str) -> bool:
    if type(value) is not bool:
        raise EvidenceError(f"{label}: expected boolean")
    return value


def _identifier(value: Any, label: str) -> str:
    if not isinstance(value, str) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}", value):
        raise EvidenceError(f"{label}: expected bounded ASCII identifier")
    return value


def _timestamp(value: Any) -> dt.datetime:
    if not isinstance(value, str) or not re.fullmatch(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z", value):
        raise EvidenceError("started_at: expected UTC YYYY-MM-DDTHH:MM:SSZ")
    try:
        return dt.datetime.strptime(value, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=dt.timezone.utc)
    except ValueError as error:
        raise EvidenceError("started_at: invalid calendar timestamp") from error


def _pairs(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise EvidenceError("duplicate JSON member")
        result[key] = value
    return result


def _constant(_: str) -> Any:
    raise EvidenceError("non-finite JSON number")


def read_observations(path: Path) -> tuple[dict[str, Any], bytes]:
    """Read at most 64 MiB, preserving exact source bytes for the report."""
    with path.open("rb") as stream:
        raw = stream.read(MAX_INPUT_BYTES + 1)
    if len(raw) > MAX_INPUT_BYTES:
        raise EvidenceError("input exceeds 64 MiB; use minute cohorts, not per-packet traces")
    try:
        document = json.loads(raw.decode("utf-8"), object_pairs_hook=_pairs, parse_constant=_constant)
    except (UnicodeError, json.JSONDecodeError, RecursionError) as error:
        raise EvidenceError("input must be a complete UTF-8 JSON document") from error
    return document, raw


def _validate(document: Any) -> None:
    _object(document, {
        "schema_version", "run_id", "profile", "started_at", "duration_seconds",
        "drain_seconds", "observed_through_ns", "execution_complete", "policy",
        "measurement", "resources", "provenance", "packet_windows", "expected_alerts",
    }, "root")
    if type(document["schema_version"]) is not int or document["schema_version"] != 1:
        raise EvidenceError("unsupported schema_version")
    _identifier(document["run_id"], "run_id")
    if not isinstance(document["profile"], str) or document["profile"] not in PROFILES:
        raise EvidenceError("profile must be staging or prod")
    _timestamp(document["started_at"])
    duration = _integer(document["duration_seconds"], "duration_seconds", MINUTE)
    if duration % MINUTE:
        raise EvidenceError("duration_seconds must contain whole one-minute cohorts")
    drain = _integer(document["drain_seconds"], "drain_seconds")
    observed = _integer(document["observed_through_ns"], "observed_through_ns")
    if observed != (duration + drain) * NS:
        raise EvidenceError("observed_through_ns must equal the declared run plus drain boundary")
    if not _boolean(document["execution_complete"], "execution_complete"):
        raise EvidenceError("completed-run report requires execution_complete=true")
    policy = _object(document["policy"], {"minimum_duration_seconds", "minimum_expected_alerts"}, "policy")
    _integer(policy["minimum_duration_seconds"], "minimum_duration_seconds", 301)
    _integer(policy["minimum_expected_alerts"], "minimum_expected_alerts", 1)
    measurement = _object(document["measurement"], {
        "live_arrival", "durable_completion", "clock_method", "clock_uncertainty_ns",
    }, "measurement")
    _boolean(measurement["live_arrival"], "live_arrival")
    _boolean(measurement["durable_completion"], "durable_completion")
    _identifier(measurement["clock_method"], "clock_method")
    _integer(measurement["clock_uncertainty_ns"], "clock_uncertainty_ns")
    resources = _object(document["resources"], {"sut_vcpus", "sut_memory_bytes", "active_rules"}, "resources")
    for key, value in resources.items():
        _integer(value, key, 1)
    provenance = _object(document["provenance"], {
        "sut_commit", "harness_commit", "config_sha256", "rules_sha256", "fixture_sha256",
    }, "provenance")
    for key, value in provenance.items():
        size = 40 if key.endswith("commit") else 64
        if not isinstance(value, str) or not re.fullmatch(rf"[0-9a-f]{{{size}}}", value):
            raise EvidenceError(f"{key}: invalid full digest")
    windows = document["packet_windows"]
    if not isinstance(windows, list) or len(windows) != duration // MINUTE:
        raise EvidenceError("packet_windows must cover every minute exactly once")
    for index, window in enumerate(windows):
        label = f"packet_windows[{index}]"
        _object(window, {"start_second", "phase", "offered_eligible", "fully_processed",
                         "offered_bytes", "observation_complete"}, label)
        if _integer(window["start_second"], label) != index * MINUTE:
            raise EvidenceError(f"{label}: non-contiguous or unordered cohort")
        if window["phase"] not in ("sustained", "burst"):
            raise EvidenceError(f"{label}: phase must be sustained or burst")
        if index and window["phase"] == windows[index - 1]["phase"] == "burst":
            raise EvidenceError("burst cohorts must be isolated 60-second intervals")
        for key in ("offered_eligible", "fully_processed", "offered_bytes"):
            _integer(window[key], f"{label}.{key}")
        if window["fully_processed"] > window["offered_eligible"]:
            raise EvidenceError(f"{label}: fully_processed exceeds offered eligible packets")
        if bool(window["offered_eligible"]) != bool(window["offered_bytes"]):
            raise EvidenceError(f"{label}: offered packet/byte counts disagree")
        if window["offered_bytes"] < window["offered_eligible"]:
            raise EvidenceError(f"{label}: fewer offered bytes than packets")
        _boolean(window["observation_complete"], f"{label}.observation_complete")
    alerts = document["expected_alerts"]
    if not isinstance(alerts, list):
        raise EvidenceError("expected_alerts must be an array")
    event_ids: set[str] = set()
    packet_rules: set[tuple[str, str]] = set()
    packets: dict[str, tuple[int, int | None]] = {}
    packet_counts = [0] * len(windows)
    for index, alert in enumerate(alerts):
        label = f"expected_alerts[{index}]"
        _object(alert, {"event_id", "packet_id", "rule_id", "offered_ns", "arrival_ns", "durable_ns"}, label)
        for key in ("event_id", "packet_id", "rule_id"):
            _identifier(alert[key], f"{label}.{key}")
        identity = (alert["packet_id"], alert["rule_id"])
        if alert["event_id"] in event_ids or identity in packet_rules:
            raise EvidenceError(f"{label}: duplicate expected event or packet/rule pair")
        event_ids.add(alert["event_id"])
        packet_rules.add(identity)
        offered = _integer(alert["offered_ns"], f"{label}.offered_ns")
        if offered >= duration * NS:
            raise EvidenceError(f"{label}: offered time outside the run")
        arrival, durable = alert["arrival_ns"], alert["durable_ns"]
        if arrival is not None:
            _integer(arrival, f"{label}.arrival_ns")
            if not offered <= arrival <= observed:
                raise EvidenceError(f"{label}: arrival time outside offer/observation boundaries")
        if durable is not None:
            _integer(durable, f"{label}.durable_ns")
            if arrival is None or not arrival <= durable <= observed:
                raise EvidenceError(f"{label}: durable time lacks a valid preceding arrival")
        packet_id = alert["packet_id"]
        if packet_id in packets and packets[packet_id] != (offered, arrival):
            raise EvidenceError(f"{label}: one packet has inconsistent timestamps")
        if packet_id not in packets:
            packet_counts[offered // (MINUTE * NS)] += 1
            packets[packet_id] = (offered, arrival)
    for index, count in enumerate(packet_counts):
        if count > windows[index]["offered_eligible"]:
            raise EvidenceError("expected alerts refer to more unique packets than were offered")


def _alert_summary(alerts: list[dict[str, Any]], uncertainty: int) -> dict[str, Any]:
    finite = sorted(alert["durable_ns"] - alert["arrival_ns"] + uncertainty
                    for alert in alerts if alert["durable_ns"] is not None)
    expected = len(alerts)
    missing = expected - len(finite)
    late = sum(value > DEADLINE_NS for value in finite)
    rank = (99 * expected + 99) // 100
    kind = "no_samples" if not expected else "unbounded_missing" if rank > len(finite) else "finite"
    return {
        "expected": expected, "durably_completed": len(finite), "missing": missing,
        "late": late, "deadline_violations": missing + late,
        "p99_upper_bound_ns": finite[rank - 1] if kind == "finite" else None,
        "p99_kind": kind, "p99_rank": rank if expected else None,
        "p99_sample_count": expected,
    }


def _summary(windows: list[dict[str, Any]], alerts: list[dict[str, Any]], uncertainty: int) -> dict[str, Any]:
    phases: dict[str, Any] = {}
    for phase, limit in LOSS_LIMIT_PPM.items():
        rows = [row for row in windows if row["phase"] == phase]
        offered = sum(row["offered_eligible"] for row in rows)
        completed = sum(row["fully_processed"] for row in rows)
        lost = offered - completed
        phases[phase] = {
            "minutes": len(rows), "offered_eligible": offered, "fully_processed": completed,
            "unprocessed": lost, "loss_fraction": lost / offered if offered else None,
            "loss_limit_ppm": limit,
            "loss_limit_exceeded": lost * 1_000_000 > offered * limit if offered else None,
            "offered_bytes": sum(row["offered_bytes"] for row in rows),
        }
    return {"packets_by_phase": phases, "alerts": _alert_summary(alerts, uncertainty)}


def summarize(document: Any) -> dict[str, Any]:
    """Compute cohort summaries without executing traffic or asserting compliance.

    Counters, expected events, clock method and durability observations are supplied
    by the collector. Their physical truth requires departmental evidence review.
    """
    _validate(document)
    windows = document["packet_windows"]
    duration = document["duration_seconds"]
    alerts = document["expected_alerts"]
    uncertainty = document["measurement"]["clock_uncertainty_ns"]
    profile = PROFILES[document["profile"]]
    buckets: list[list[dict[str, Any]]] = [[] for _ in windows]
    for alert in alerts:
        buckets[alert["offered_ns"] // (MINUTE * NS)].append(alert)
    minutes = []
    rolling = []
    coverage: list[str] = []
    failures: list[str] = []
    if not document["measurement"]["live_arrival"]:
        coverage.append("live arrival boundary has not been established")
    if not document["measurement"]["durable_completion"]:
        coverage.append("durable completion boundary has not been established")
    for key, target in (("sut_vcpus", profile["vcpus"]),
                        ("sut_memory_bytes", profile["memory_bytes"]),
                        ("active_rules", profile["rules"])):
        if document["resources"][key] != target:
            coverage.append(f"{key} differs from the proposed profile")
    policy = document["policy"]
    if duration < policy["minimum_duration_seconds"]:
        coverage.append("run is shorter than its declared extended-duration minimum")
    if len(alerts) < policy["minimum_expected_alerts"]:
        coverage.append("expected alert count is below its declared minimum")
    if not alerts:
        coverage.append("no expected alerts were supplied")
    if {row["phase"] for row in windows} != set(LOSS_LIMIT_PPM):
        coverage.append("both sustained and burst observations are required")
    for index, window in enumerate(windows):
        start = index * MINUTE
        result = _summary([window], buckets[index], uncertainty)
        achieved_bps = window["offered_bytes"] * 8 / MINUTE
        result.update(start_second=start, end_second=start + MINUTE,
                      phase=window["phase"], offered_bps=achieved_bps)
        minutes.append(result)
        if not window["observation_complete"]:
            coverage.append(f"minute {index}: cohort accounting is incomplete")
        if not window["offered_eligible"]:
            coverage.append(f"minute {index}: no eligible offered packets")
        target_bps = profile[window["phase"] + "_bps"]
        if window["offered_bytes"] * 8 < target_bps * MINUTE:
            coverage.append(f"minute {index}: offered load is below the proposed target")
        if window["phase"] == "burst" and result["packets_by_phase"]["burst"]["loss_limit_exceeded"]:
            failures.append(f"minute {index}: burst packet loss exceeds 0.2%")
        if index >= 4:
            cohort_alerts = [alert for bucket in buckets[index - 4:index + 1] for alert in bucket]
            summary = _summary(windows[index - 4:index + 1], cohort_alerts, uncertainty)
            summary.update(start_second=(index - 4) * MINUTE, end_second=(index + 1) * MINUTE)
            rolling.append(summary)
    overall = _summary(windows, alerts, uncertainty)
    for label, summary in [("full_run", overall)] + [
        (f"window_ending_{window['end_second']}", window) for window in rolling
    ]:
        alert_metrics = summary["alerts"]
        if alert_metrics["missing"]:
            failures.append(f"{label}: expected security alerts are missing")
        if alert_metrics["p99_kind"] == "unbounded_missing" or (
            alert_metrics["p99_upper_bound_ns"] is not None
            and alert_metrics["p99_upper_bound_ns"] > DEADLINE_NS
        ):
            failures.append(f"{label}: end-to-end p99 exceeds 180 ms")
        for phase, packet_metrics in summary["packets_by_phase"].items():
            if packet_metrics["loss_limit_exceeded"]:
                failures.append(f"{label}: {phase} packet loss exceeds its limit")
    return {
        "schema_version": 1, "artifact_kind": "netsentry_slo_measurement_summary",
        "run_id": document["run_id"], "profile": document["profile"],
        "started_at": document["started_at"], "duration_seconds": duration,
        "drain_seconds": document["drain_seconds"],
        "observed_through_ns": document["observed_through_ns"],
        "window_basis": "offer_time_cohorts_finalized_at_run_drain",
        "status": "inconclusive" if coverage else "measurement_failed" if failures else "review_required",
        "slo_compliance_asserted": False, "departmental_review_required": True,
        "evidence_class": "single_vm_supplied_observations",
        "coverage_gaps": coverage, "measurement_failures": failures,
        "deadline_ns": DEADLINE_NS, "clock_uncertainty_ns": uncertainty,
        "percentile_method": "nearest_rank_over_all_expected_alerts_missing_as_positive_infinity",
        "policy": dict(policy), "provenance": dict(document["provenance"]),
        "resources": dict(document["resources"]), "proposed_profile": dict(profile),
        "full_run": overall, "minute_cohorts": minutes, "rolling_five_minute_cohorts": rolling,
        "limitations": [
            "This tool summarizes supplied observations; it does not collect or independently verify them.",
            "Cohort counters must represent unique eligible packets offered and completed by the drain boundary.",
            "The expected-alert oracle, clock error bound and durable completion require external review.",
            "Hardware, isolation, generator headroom, traffic/rule mix and raw traces require departmental review.",
            "Resource/rate checks are partial; meeting them does not certify the proposed hardware or workload.",
            "Reports retain source observations, not all external collector logs; retain those separately.",
        ],
    }


def write_report(path: Path, report: dict[str, Any]) -> None:
    """Publish a complete JSON file without replacing any existing pathname."""
    path = path.expanduser()
    payload = (json.dumps(report, indent=2, ensure_ascii=False, allow_nan=False) + "\n").encode("utf-8")
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary: Path | None = None
    try:
        with tempfile.NamedTemporaryFile(mode="wb", dir=path.parent, prefix=".slo-report-", delete=False) as stream:
            temporary = Path(stream.name)
            stream.write(payload)
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temporary, path)
        directory = os.open(path.parent, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        if temporary is not None:
            temporary.unlink(missing_ok=True)


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True, help="completed-run observation JSON")
    parser.add_argument("--output", type=Path, help="new output file; existing files are never overwritten")
    args = parser.parse_args(argv)
    try:
        document, raw = read_observations(args.input.expanduser())
        report = summarize(document)
        report["source"] = {"sha256": hashlib.sha256(raw).hexdigest(), "raw_json": raw.decode("utf-8")}
        stamp = _timestamp(document["started_at"]).strftime("%Y%m%d_%H%M%S")
        output = args.output or (
            Path.home() / "benchmarks" / "r90\u201175"
            / f"acceptance_{document['profile']}_{stamp}.json"
        )
        write_report(output, report)
    except (EvidenceError, OSError, ValueError, RecursionError) as error:
        print(f"[slo-report] {error}", file=sys.stderr)
        return 2
    print(f"[slo-report] {report['status']}; departmental review required: {output}")
    return {"review_required": 0, "measurement_failed": 1, "inconclusive": 3}[report["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
