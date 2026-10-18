"""Schema-2 local-v1 supplied evidence; never a live measurement or L1/L2 certificate."""
from __future__ import annotations

import copy
import hashlib
import json
from pathlib import Path
from typing import Any

if __package__:
    from . import slo_report as report
else:
    import slo_report as report

CONTRACT = "local-v1"
DEADLINE_NS = 500_000_000
CLOCK_LIMIT_NS = 10_000_000
POLICY = {"minimum_duration_seconds": 600, "minimum_expected_alerts": 600}
INTEGRITY = {"capture_errors", "submission_errors", "raw_evidence_complete", "identity_accounting_complete"}
REVIEW_CHECKS = {"oracle", "clock", "durability", "load", "resources", "raw_evidence", "identity_accounting"}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise report.EvidenceError(message)


def identity(document: dict[str, Any]) -> None:
    require(type(document.get("schema_version")) is int and document["schema_version"] == 2,
            "local-v1 requires schema_version 2")
    require(document.get("profile") == CONTRACT and document.get("contract_id") == CONTRACT,
            "schema 2 requires profile and contract_id local-v1")


def validate(document: Any) -> None:
    identity(document)
    # Use the unchanged v1 structural validator on an explicit projection, after
    # checking every local member. There is no compatibility-decoder fallback.
    fields = {"schema_version", "run_id", "profile", "started_at", "duration_seconds",
              "drain_seconds", "observed_through_ns", "execution_complete", "policy",
              "measurement", "resources", "provenance", "packet_windows", "expected_alerts",
              "contract_id", "integrity", "anomalies"}
    report._object(document, fields, "local root")
    measurement = report._object(document["measurement"], {
        "live_arrival", "durable_completion", "clock_method", "clock_uncertainty_ns", "clock_checked",
    }, "local measurement")
    report._boolean(measurement["clock_checked"], "clock_checked")
    if measurement["clock_uncertainty_ns"] is not None:
        report._integer(measurement["clock_uncertainty_ns"], "clock_uncertainty_ns")
    resources = report._object(document["resources"], {
        "sut_vcpus", "sut_memory_bytes", "active_rules", "generator_vcpus", "generator_memory_bytes",
    }, "local resources")
    for key in ("generator_vcpus", "generator_memory_bytes"):
        report._integer(resources[key], key, 1)
    integrity = report._object(document["integrity"], INTEGRITY, "integrity")
    for key in ("capture_errors", "submission_errors"):
        report._integer(integrity[key], key)
    for key in ("raw_evidence_complete", "identity_accounting_complete"):
        report._boolean(integrity[key], key)
    anomalies = document["anomalies"]
    require(isinstance(anomalies, list), "anomalies must be an array")
    for item in anomalies:
        report._object(item, {"kind", "identity", "at_ns", "reason"}, "anomaly")
        require(item["kind"] in ("arrival", "processed", "durable"), "invalid anomaly kind")
        require(item["reason"] in ("duplicate", "unexpected", "after_observation"), "invalid anomaly reason")
        report._identifier(item["identity"], "anomaly identity")
        report._integer(item["at_ns"], "anomaly at_ns")
        if item["reason"] == "after_observation":
            require(item["at_ns"] > document["observed_through_ns"], "late anomaly must exceed observation boundary")
    require(isinstance(document["packet_windows"], list), "packet_windows must be an array")
    for row in document["packet_windows"]:
        report._object(row, {"start_second", "phase", "offered_eligible", "fully_processed",
                             "offered_bytes", "observation_complete", "submitted_eligible"}, "local window")
        submitted = report._integer(row["submitted_eligible"], "submitted_eligible")
        require(type(row["offered_eligible"]) is int and submitted <= row["offered_eligible"],
                "submitted exceeds offered")
        require(type(row["fully_processed"]) is int and row["fully_processed"] <= submitted,
                "fully_processed exceeds submitted")
    projected = copy.deepcopy(document)
    for key in ("contract_id", "integrity", "anomalies"):
        del projected[key]
    projected.update(schema_version=1, profile="staging", execution_complete=True)
    report._boolean(document["execution_complete"], "execution_complete")
    del projected["measurement"]["clock_checked"]
    projected["measurement"]["clock_uncertainty_ns"] = measurement["clock_uncertainty_ns"] or 0
    for key in ("generator_vcpus", "generator_memory_bytes"):
        del projected["resources"][key]
    for row in projected["packet_windows"]:
        del row["submitted_eligible"]
    report._validate_v1(projected)
    # These are immutable contract constants, not operator-adjustable minima.
    require(document["policy"] == POLICY, "local-v1 policy must match the frozen contract")


def alerts_summary(alerts: list[dict[str, Any]], uncertainty: int | None) -> dict[str, Any]:
    finite = sorted(row["durable_ns"] - row["arrival_ns"] for row in alerts if row["durable_ns"] is not None)
    count = len(alerts)
    rank = (99 * count + 99) // 100
    kind = "no_samples" if not count else "unbounded_missing" if rank > len(finite) else "finite"
    raw = finite[rank - 1] if kind == "finite" else None
    upper = raw + uncertainty if raw is not None and uncertainty is not None else None
    return dict(expected=count, durably_completed=len(finite), missing=count-len(finite),
                late=sum(value + (uncertainty or 0) > DEADLINE_NS for value in finite),
                deadline_violations=count-len(finite)+sum(value + (uncertainty or 0) > DEADLINE_NS for value in finite),
                p99_kind=kind, p99_rank=rank if count else None, p99_sample_count=count,
                p99_raw_ns=raw, p99_upper_bound_ns=upper)


def packet_summary(rows: list[dict[str, Any]]) -> dict[str, Any]:
    offered = sum(row["offered_eligible"] for row in rows)
    completed = sum(row["fully_processed"] for row in rows)
    lost = offered - completed
    return dict(offered_eligible=offered, submitted_eligible=sum(row["submitted_eligible"] for row in rows),
                fully_processed=completed, unprocessed=lost,
                offered_bytes=sum(row["offered_bytes"] for row in rows),
                mean_frame_bytes=sum(row["offered_bytes"] for row in rows)/offered if offered else None,
                loss_fraction=lost/offered if offered else None, loss_limit_ppm=10_000,
                loss_limit_exceeded=lost*100 > offered if offered else None)


def summarize(document: Any) -> dict[str, Any]:
    validate(document)
    gaps: list[str] = []
    failures: list[str] = []
    windows, alerts = document["packet_windows"], document["expected_alerts"]
    uncertainty = document["measurement"]["clock_uncertainty_ns"]
    if document["duration_seconds"] != 600 or document["drain_seconds"] != 30:
        gaps.append("local-v1 requires 600 seconds workload plus 30 seconds observation")
    if not document["execution_complete"]:
        gaps.append("execution is incomplete")
    for field in ("live_arrival", "durable_completion", "clock_checked"):
        if not document["measurement"][field]:
            gaps.append(field + " has not been established")
    if uncertainty is None or uncertainty > CLOCK_LIMIT_NS:
        gaps.append("clock uncertainty is unknown or exceeds 10 ms")
    for key in ("capture_errors", "submission_errors"):
        if document["integrity"][key]:
            gaps.append(key + " prevents valid measurement")
    for key in ("raw_evidence_complete", "identity_accounting_complete"):
        if not document["integrity"][key]:
            gaps.append(key + " is false")
    resources = document["resources"]
    for key, maximum in (("sut_vcpus", 4), ("sut_memory_bytes", 3*1024**3),
                         ("generator_vcpus", 2), ("generator_memory_bytes", 1024**3)):
        if resources[key] > maximum:
            gaps.append(key + " exceeds local-v1 allocation budget")
    if resources["active_rules"] != 100:
        gaps.append("local-v1 requires 100 active rules")
    seconds = [row["offered_ns"] // report.NS for row in alerts]
    if len(alerts) != 600 or sorted(seconds) != list(range(600)):
        gaps.append("oracle must contain exactly one expected event per second for 600 seconds")
    minutes = []
    for index, row in enumerate(windows):
        phase = "burst" if row["start_second"] in (180, 480) else "sustained"
        expected = 30_000 if phase == "burst" else 6_000
        if row["phase"] != phase:
            gaps.append(f"minute {index}: phase differs from frozen local-v1 schedule")
        if not 95*expected <= row["offered_eligible"]*100 <= 105*expected:
            gaps.append(f"minute {index}: actual offered load outside +/-5%")
        if row["submitted_eligible"] != row["offered_eligible"]:
            gaps.append(f"minute {index}: not all offered packets were submitted")
        if not row["observation_complete"]:
            gaps.append(f"minute {index}: accounting incomplete")
        cohort = [a for a in alerts if row["start_second"]*report.NS <= a["offered_ns"] < (row["start_second"]+60)*report.NS]
        minutes.append(dict(start_second=row["start_second"], phase=row["phase"],
                            target_pps=expected//60, actual_pps=row["offered_eligible"]/60,
                            packets=packet_summary([row]), alerts=alerts_summary(cohort, uncertainty)))
    segments = [("sustained", [row for row in windows if row["phase"] == "sustained"])]
    segments += [(f"burst_{row['start_second']}", [row]) for row in windows if row["phase"] == "burst"]
    packets = {name: packet_summary(rows) for name, rows in segments}
    for name, metrics in packets.items():
        if metrics["loss_limit_exceeded"]:
            failures.append(name + ": packet loss exceeds 1%")
    overall = alerts_summary(alerts, uncertainty)
    if overall["missing"]:
        failures.append("expected security alerts are missing")
    if overall["p99_kind"] == "unbounded_missing" or (
        overall["p99_raw_ns"] is not None and overall["p99_raw_ns"] > DEADLINE_NS
    ) or (overall["p99_upper_bound_ns"] is not None and overall["p99_upper_bound_ns"] > DEADLINE_NS):
        failures.append("host-timestamp-to-WAL p99 upper bound exceeds 500 ms")
    alert_segments = {name: alerts_summary([a for a in alerts if any(
        row["start_second"]*report.NS <= a["offered_ns"] < (row["start_second"]+60)*report.NS
        for row in rows)], uncertainty) for name, rows in segments}
    for name, metrics in alert_segments.items():
        if metrics["missing"]:
            failures.append(name + ": expected security alerts are missing")
        if metrics["p99_kind"] == "unbounded_missing" or (
            metrics["p99_raw_ns"] is not None and metrics["p99_raw_ns"] > DEADLINE_NS
        ) or (metrics["p99_upper_bound_ns"] is not None and metrics["p99_upper_bound_ns"] > DEADLINE_NS):
            failures.append(name + ": p99 upper bound exceeds 500 ms")
    if document["anomalies"]:
        failures.append("unexpected, duplicate or after-observation lifecycle events retained")
    return {
        "schema_version": 2, "artifact_kind": "netsentry_slo_measurement_summary",
        "contract_id": CONTRACT, "profile": CONTRACT, "run_id": document["run_id"],
        "evaluator_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "started_at": document["started_at"], "duration_seconds": document["duration_seconds"],
        "drain_seconds": document["drain_seconds"], "observed_through_ns": document["observed_through_ns"],
        # Numerical failures win while validity gaps remain visible alongside them.
        "status": "measurement_failed" if failures else "inconclusive" if gaps else "review_required",
        "coverage_gaps": gaps, "measurement_failures": failures,
        "deadline_ns": DEADLINE_NS, "clock_uncertainty_ns": uncertainty,
        "percentile_method": "nearest_rank_over_all_expected_alerts_missing_as_positive_infinity",
        "policy": dict(POLICY), "resources": dict(resources), "provenance": dict(document["provenance"]),
        "integrity": dict(document["integrity"]), "anomalies": copy.deepcopy(document["anomalies"]),
        "full_run": {"alerts": overall, "packets_by_phase": packets, "alerts_by_phase": alert_segments}, "minute_cohorts": minutes,
        "local_run_accepted": False, "slo_compliance_asserted": False, "departmental_review_required": False,
        "local_review_required": True, "evidence_class": "single_vm_supplied_local_synthetic_observations",
        "limitations": ["Supplied declarations require retained raw evidence and local review.",
                        "One run cannot establish two-run L1 or L2 stability/recovery acceptance.",
                        "Host capture timestamp to WAL return is not physical NIC or power-loss durability.",
                        "No live traffic, hardware discovery or production certification is performed."],
    }


def reviewed(summary: dict[str, Any], raw: bytes, declaration: Any) -> dict[str, Any]:
    require(isinstance(raw, bytes) and len(raw) <= report.MAX_INPUT_BYTES, "invalid review observation bytes")
    try:
        document = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs, parse_constant=report._constant)
    except (UnicodeError, json.JSONDecodeError, RecursionError) as error:
        raise report.EvidenceError("invalid review observation JSON") from error
    require(json.dumps(summary, sort_keys=True, allow_nan=False) ==
            json.dumps(summarize(document), sort_keys=True, allow_nan=False),
            "review summary differs from supplied observation bytes")
    report._object(declaration, {"schema_version", "contract_id", "run_id", "observations_sha256",
                                 "reviewer", "reviewed_at", "evidence_sha256", "checks"}, "local review")
    require(type(declaration["schema_version"]) is int and declaration["schema_version"] == 2
            and declaration["contract_id"] == CONTRACT, "unsupported review contract")
    require(summary["status"] == "review_required", "review cannot accept a failed or inconclusive measurement")
    require(declaration["run_id"] == summary["run_id"] and
            declaration["observations_sha256"] == hashlib.sha256(raw).hexdigest(), "review observation binding differs")
    report._identifier(declaration["reviewer"], "reviewer")
    report._timestamp(declaration["reviewed_at"])
    report._object(declaration["checks"], REVIEW_CHECKS, "review checks")
    require(all(value is True for value in declaration["checks"].values()), "all local review checks must be true")
    # Digest catalog is an explicit human declaration, not automatic authentication.
    report._object(declaration["evidence_sha256"], {"bundle", "context", "raw_manifest"}, "review evidence")
    import re
    require(all(isinstance(value, str) and re.fullmatch(r"[0-9a-f]{64}", value)
                for value in declaration["evidence_sha256"].values()), "invalid review evidence digest")
    result = copy.deepcopy(summary)
    result.update(status="reviewed_local_acceptance", local_run_accepted=True, local_review_required=False,
                  review=copy.deepcopy(declaration), review_authority="supplied_local_reviewer_declaration")
    return result


def ledger_schema(db: Any) -> None:
    db.execute("ALTER TABLE packets ADD COLUMN submitted INTEGER NOT NULL DEFAULT 1")
    db.execute("CREATE TABLE anomalies (kind TEXT, identity TEXT, at_ns INTEGER, reason TEXT)")


def offer(db: Any, row: dict[str, Any], document: dict[str, Any], legacy_offer: Any) -> None:
    report._object(row, {"packet_id", "offered_ns", "offered_bytes", "expected_alerts", "submitted"}, "local offer")
    submitted = report._boolean(row["submitted"], "submitted")
    projected = {key: value for key, value in row.items() if key != "submitted"}
    legacy_offer(db, projected, document)
    db.execute("UPDATE packets SET submitted=? WHERE packet_id=?", (int(submitted), row["packet_id"]))
    if submitted:
        document["packet_windows"][row["offered_ns"] // (60*report.NS)]["submitted_eligible"] += 1


def event(db: Any, row: dict[str, Any], observed: int) -> None:
    kind = row.get("kind")
    require(kind in ("arrival", "processed", "durable"), "invalid local event kind")
    id_field = "event_id" if kind == "durable" else "packet_id"
    report._object(row, {"kind", id_field, "at_ns"}, "local event")
    key = report._identifier(row[id_field], id_field)
    at = report._integer(row["at_ns"], "at_ns")
    queries = {
        "arrival": "SELECT arrival_ns, submitted FROM packets WHERE packet_id=?",
        "processed": "SELECT processed_ns, submitted FROM packets WHERE packet_id=?",
        "durable": "SELECT a.durable_ns, p.submitted FROM alerts a JOIN packets p ON p.packet_id=a.packet_id WHERE event_id=?",
    }
    current = db.execute(queries[kind], (key,)).fetchone()
    reason = "after_observation" if at > observed else (
        "unexpected" if current is None or not current[1] else "duplicate" if current[0] is not None else None)
    if reason:
        db.execute("INSERT INTO anomalies VALUES (?, ?, ?, ?)", (kind, key, at, reason))
        return
    updates = {
        "arrival": "UPDATE packets SET arrival_ns=? WHERE packet_id=?",
        "processed": "UPDATE packets SET processed_ns=? WHERE packet_id=?",
        "durable": "UPDATE alerts SET durable_ns=? WHERE event_id=?",
    }
    db.execute(updates[kind], (at, key))


def finalize(db: Any, document: dict[str, Any], encoded: Any) -> None:
    failed = db.execute("SELECT count(*) FROM packets WHERE submitted=0").fetchone()[0]
    document["integrity"]["submission_errors"] = max(document["integrity"]["submission_errors"], failed)
    budget = len(encoded(document))
    for kind, key, at, reason in db.execute("SELECT kind, identity, at_ns, reason FROM anomalies ORDER BY rowid"):
        item = dict(kind=kind, identity=key, at_ns=at, reason=reason)
        budget += len(encoded(item)) + 64
        require(budget <= report.MAX_INPUT_BYTES, "local anomaly output exceeds reporter 64 MiB budget")
        document["anomalies"].append(item)
