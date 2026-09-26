#!/usr/bin/env python3
"""Adapt department-exported packet/oracle and lifecycle ledgers to SLO observations."""

from __future__ import annotations

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import sqlite3
import sys
import tempfile
from typing import Any, Iterator, Sequence

if __package__:
    from . import slo_report as report
else:
    import slo_report as report

MAX_LINE_BYTES = 256 * 1024
COMMIT_INTERVAL = 10_000


def _encoded(document: Any) -> bytes:
    return (json.dumps(document, indent=2, ensure_ascii=False, allow_nan=False) + "\n").encode("utf-8")


def _retain(path: Path, raw: bytes) -> dict[str, Any]:
    with path.open("xb") as stream:
        stream.write(raw)
        stream.flush()
        os.fsync(stream.fileno())
    return {"file": path.name, "bytes": len(raw), "sha256": hashlib.sha256(raw).hexdigest()}


def _rows(source: Path, retained: Path, receipts: list[dict[str, Any]]) -> Iterator[dict[str, Any]]:
    """Bound each row; hash and retain exactly the bytes being interpreted."""
    digest = hashlib.sha256()
    count = size = 0
    with source.open("rb") as incoming, retained.open("xb") as outgoing:
        while raw := incoming.readline(MAX_LINE_BYTES + 1):
            count += 1
            if len(raw) > MAX_LINE_BYTES:
                raise report.EvidenceError(f"{retained.name} line {count}: exceeds 256 KiB")
            try:
                row = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                                 parse_constant=report._constant)
            except (UnicodeError, ValueError, RecursionError) as error:
                raise report.EvidenceError(f"{retained.name} line {count}: invalid JSON row") from error
            if not isinstance(row, dict):
                raise report.EvidenceError(f"{retained.name} line {count}: expected object")
            outgoing.write(raw)
            digest.update(raw)
            size += len(raw)
            yield row
        outgoing.flush()
        os.fsync(outgoing.fileno())
    receipts.append({"file": retained.name, "bytes": size, "rows": count, "sha256": digest.hexdigest()})


def _manifest(document: Any) -> dict[str, Any]:
    if not isinstance(document, dict) or "expected_alerts" in document:
        raise report.EvidenceError("manifest must be an object without expected_alerts")
    result = copy.deepcopy(document)
    windows = result.get("packet_windows")
    if not isinstance(windows, list):
        raise report.EvidenceError("manifest packet_windows must be an array")
    for window in windows:
        report._object(window, {"start_second", "phase", "observation_complete"}, "manifest window")
        window.update(offered_eligible=0, fully_processed=0, offered_bytes=0)
    result["expected_alerts"] = []
    report._validate(result)
    return result


def _schema(db: sqlite3.Connection) -> None:
    db.execute("PRAGMA foreign_keys=ON")
    db.executescript("""
        CREATE TABLE packets (
            packet_id TEXT PRIMARY KEY NOT NULL,
            offered_ns INTEGER NOT NULL,
            arrival_ns INTEGER,
            processed_ns INTEGER
        );
        CREATE TABLE alerts (
            event_id TEXT PRIMARY KEY NOT NULL,
            packet_id TEXT NOT NULL REFERENCES packets(packet_id),
            rule_id TEXT NOT NULL,
            durable_ns INTEGER,
            UNIQUE(packet_id, rule_id)
        );
    """)


def _offer(db: sqlite3.Connection, row: dict[str, Any], document: dict[str, Any]) -> None:
    report._object(row, {"packet_id", "offered_ns", "offered_bytes", "expected_alerts"}, "offer")
    packet = report._identifier(row["packet_id"], "packet_id")
    offered = report._integer(row["offered_ns"], "offered_ns")
    size = report._integer(row["offered_bytes"], "offered_bytes", 1)
    if offered >= document["duration_seconds"] * report.NS:
        raise report.EvidenceError("offered_ns outside run")
    if not isinstance(row["expected_alerts"], list):
        raise report.EvidenceError("offer expected_alerts must be an array")
    db.execute("INSERT INTO packets(packet_id, offered_ns) VALUES (?, ?)", (packet, offered))
    for alert in row["expected_alerts"]:
        report._object(alert, {"event_id", "rule_id"}, "oracle alert")
        event = report._identifier(alert["event_id"], "event_id")
        rule = report._identifier(alert["rule_id"], "rule_id")
        db.execute("INSERT INTO alerts(event_id, packet_id, rule_id) VALUES (?, ?, ?)", (event, packet, rule))
    window = document["packet_windows"][offered // (report.MINUTE * report.NS)]
    window["offered_eligible"] += 1
    window["offered_bytes"] += size


def _event(db: sqlite3.Connection, row: dict[str, Any], observed: int) -> None:
    kind = row.get("kind")
    if kind not in ("arrival", "processed", "durable"):
        raise report.EvidenceError("event kind must be arrival, processed or durable")
    identity = "event_id" if kind == "durable" else "packet_id"
    report._object(row, {"kind", identity, "at_ns"}, "event")
    key = report._identifier(row[identity], identity)
    at = report._integer(row["at_ns"], "at_ns")
    if at > observed:
        raise report.EvidenceError("event outside observation boundary")
    # Queries are fixed strings: no record value can become a SQL identifier.
    queries = {
        "arrival": "UPDATE packets SET arrival_ns=? WHERE packet_id=? AND arrival_ns IS NULL",
        "processed": "UPDATE packets SET processed_ns=? WHERE packet_id=? AND processed_ns IS NULL",
        "durable": "UPDATE alerts SET durable_ns=? WHERE event_id=? AND durable_ns IS NULL",
    }
    if db.execute(queries[kind], (at, key)).rowcount != 1:
        raise report.EvidenceError("duplicate lifecycle event or identity absent from offered oracle")


def _observations(db: sqlite3.Connection, document: dict[str, Any]) -> None:
    invalid_packet = db.execute("""
        SELECT 1 FROM packets WHERE
            (arrival_ns IS NOT NULL AND arrival_ns < offered_ns) OR
            (processed_ns IS NOT NULL AND (arrival_ns IS NULL OR processed_ns < arrival_ns))
        LIMIT 1
    """).fetchone()
    invalid_alert = db.execute("""
        SELECT 1 FROM alerts a JOIN packets p ON p.packet_id=a.packet_id
        WHERE a.durable_ns IS NOT NULL AND
            (p.arrival_ns IS NULL OR a.durable_ns < p.arrival_ns OR
             (p.processed_ns IS NOT NULL AND a.durable_ns > p.processed_ns))
        LIMIT 1
    """).fetchone()
    if invalid_packet or invalid_alert:
        raise report.EvidenceError("inconsistent offer/arrival/durable/terminal-processing order")
    # The oracle is the left side of accounting. A terminal marker alone does
    # not count as success if any expected alert lacks durable persistence.
    for offered, in db.execute("""
        SELECT p.offered_ns FROM packets p WHERE p.processed_ns IS NOT NULL
        AND NOT EXISTS (SELECT 1 FROM alerts a WHERE a.packet_id=p.packet_id AND a.durable_ns IS NULL)
    """):
        document["packet_windows"][offered // (report.MINUTE * report.NS)]["fully_processed"] += 1
    # Keep output within the reporter limit. Packet identities stay on disk;
    # only the low-volume alert oracle is materialized in memory.
    budget = len(_encoded(document))
    for event, packet, rule, offered, arrival, durable in db.execute("""
        SELECT a.event_id, a.packet_id, a.rule_id, p.offered_ns, p.arrival_ns, a.durable_ns
        FROM alerts a JOIN packets p ON p.packet_id=a.packet_id ORDER BY a.event_id
    """):
        alert = dict(event_id=event, packet_id=packet, rule_id=rule, offered_ns=offered,
                     arrival_ns=arrival, durable_ns=durable)
        # Conservative allowance for enclosing array indentation/separators.
        budget += len(_encoded(alert)) + 64
        if budget > report.MAX_INPUT_BYTES:
            raise report.EvidenceError("alert observation output exceeds reporter 64 MiB budget")
        document["expected_alerts"].append(alert)
    report._validate(document)


def collect(manifest: Path, offered: Path, events: Path, output: Path,
            scratch_dir: Path | None = None) -> dict[str, Any]:
    """Build a new retained bundle; keep partial output on error for inspection.

    Inputs must be finalized department exports. No live capture, test traffic,
    clock calibration, physical durability verification or SLO certification.
    """
    source, raw_manifest = report.read_observations(manifest.expanduser())
    document = _manifest(source)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    receipts = [_retain(output / "manifest.json", raw_manifest)]
    with tempfile.TemporaryDirectory(prefix="netsentry-slo-", dir=scratch_dir) as temporary:
        db = sqlite3.connect(str(Path(temporary) / "identities.sqlite"))
        try:
            _schema(db)
            for number, row in enumerate(_rows(offered.expanduser(), output / "offered.jsonl", receipts), 1):
                _offer(db, row, document)
                if number % COMMIT_INTERVAL == 0:
                    db.commit()
            db.commit()
            for number, row in enumerate(_rows(events.expanduser(), output / "events.jsonl", receipts), 1):
                _event(db, row, document["observed_through_ns"])
                if number % COMMIT_INTERVAL == 0:
                    db.commit()
            db.commit()
            _observations(db, document)
        finally:
            db.close()
    raw_observations = _encoded(document)
    if len(raw_observations) > report.MAX_INPUT_BYTES:
        raise report.EvidenceError("observation output exceeds reporter 64 MiB limit")
    receipts.append(_retain(output / "observations.json", raw_observations))
    receipt = {
        "schema_version": 1, "run_id": document["run_id"], "adapter_complete": True,
        "adapter": "scripts/slo_collect.py",
        "adapter_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "inputs_and_output": receipts,
        "slo_compliance_asserted": False, "departmental_review_required": True,
        "evidence_class": "single_vm_exported_packet_ledgers",
        "limitations": [
            "No live instrumentation is provided; all lifecycle observations are external exports.",
            "Oracle completeness, eligible byte boundary and timestamp/durability truth require review.",
            "Missing expected durable events prevent packet success and remain in latency denominators.",
            "The temporary SQLite identity index is discarded; retained raw inputs permit reconstruction.",
            "Capacity and adapter performance are unmeasured; disk use grows with raw packet count.",
        ],
    }
    # Publish the receipt last; earlier files alone never establish a complete bundle.
    report.write_report(output / "receipt.json", receipt)
    parent = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(parent)
    finally:
        os.close(parent)
    return receipt


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--offered", type=Path, required=True)
    parser.add_argument("--events", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--scratch-dir", type=Path, help="existing directory with space for the SQLite identity index")
    args = parser.parse_args(argv)
    try:
        collect(args.manifest, args.offered, args.events, args.output_dir,
                args.scratch_dir.expanduser() if args.scratch_dir else None)
    except (OSError, ValueError, sqlite3.Error, RecursionError) as error:
        print(f"[slo-collect] {error}; inspect retained partial output before retrying", file=sys.stderr)
        return 2
    print(f"[slo-collect] adapter complete; departmental review required: {args.output_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
