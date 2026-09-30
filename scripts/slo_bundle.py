#!/usr/bin/env python3
"""Retain and reconcile supplied SLO artifacts; never certify SLO compliance."""

from __future__ import annotations

import argparse
import copy
import hashlib
import ipaddress
import json
import math
import os
from pathlib import Path
import re
import stat
import sys
from typing import Any, Callable, Sequence

if __package__:
    from . import slo_collect as collect, slo_report as report
else:
    import slo_collect as collect
    import slo_report as report

DEFAULT_MAX_BYTES = 64 * 1024**3
META_LIMIT = 1024**2
SOURCES = {
    "sender": ("submission.json", "fixture.jsonl", "offered.jsonl", "submissions.jsonl"),
    "capture": ("summary.json",),
    "engine": ("close.json", "events.jsonl"),
    "adapter": ("receipt.json", "manifest.json", "offered.jsonl", "events.jsonl", "observations.json"),
    "report": ("report.json",),
}


def _require(condition: bool, message: str) -> None:
    if not condition:
        raise report.EvidenceError(message)


def _finite_float(raw: str) -> float:
    value = float(raw)
    _require(math.isfinite(value), "non-finite JSON number")
    return value


def _hash(value: Any) -> None:
    _require(isinstance(value, str) and re.fullmatch(r"[0-9a-f]{64}", value) is not None,
             "expected lowercase SHA-256")


def _same(left: Any, right: Any, label: str) -> None:
    # JSON encoding also distinguishes booleans from numbers (True == 1 in Python).
    _require(json.dumps(left, sort_keys=True, allow_nan=False) ==
             json.dumps(right, sort_keys=True, allow_nan=False), label)


def _version(document: dict[str, Any]) -> None:
    _require(type(document.get("schema_version")) is int and document["schema_version"] == 1,
             "unsupported schema_version")
    report._identifier(document.get("run_id"), "run_id")


def _flags(document: dict[str, Any], completion: str, *, component: bool = True,
           department: bool = True) -> None:
    _version(document)
    _require(document.get(completion) is True, f"{completion} must be true")
    _require(document.get("slo_compliance_asserted") is False, "compliance claim is forbidden")
    if component:
        _require(document.get("execution_complete") is False,
                 "component receipt must not claim completed execution")
    if "limitations" in document:
        _require(isinstance(document["limitations"], list) and
                 all(isinstance(item, str) for item in document["limitations"]), "invalid limitations")
    if department:
        _require(document.get("departmental_review_required") is True,
                 "departmental review must remain required")


class _Bundle:
    def __init__(self, output: Path, maximum: int):
        self.output = output
        self.remaining = maximum
        self.inventory: dict[str, dict[str, Any]] = {}
        self.documents: dict[str, dict[str, Any]] = {}
        self.gaps: list[str] = []
        self.mismatches: list[str] = []
        self.checked: list[str] = []
        self.skipped: list[str] = []

    def snapshot(self, source: Path | None, key: str) -> None:
        if source is None:
            self.gaps.append(f"{key}: not supplied")
            return
        if key == "report/report.json":
            limit = 512 * META_LIMIT
        elif key in ("adapter/manifest.json", "adapter/observations.json"):
            limit = report.MAX_INPUT_BYTES
        else:
            limit = self.remaining if key.endswith(".jsonl") else META_LIMIT
        limit = min(limit, self.remaining)
        try:
            descriptor = os.open(source, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        except OSError:
            self.gaps.append(f"{key}: missing or unreadable regular file")
            return
        digest = hashlib.sha256()
        size = 0
        complete = True
        with os.fdopen(descriptor, "rb") as incoming:
            before = os.fstat(incoming.fileno())
            if not stat.S_ISREG(before.st_mode):
                self.gaps.append(f"{key}: source is not a regular file")
                return
            destination = self.output / key
            destination.parent.mkdir(mode=0o700, exist_ok=True)
            with destination.open("xb") as outgoing:
                os.fchmod(outgoing.fileno(), 0o600)
                while True:
                    try:
                        raw = incoming.read(min(1024**2, limit - size) + 1)
                    except OSError:
                        self.gaps.append(f"{key}: input read failed; retained prefix only")
                        complete = False
                        break
                    if not raw:
                        break
                    keep = raw[:limit - size]
                    outgoing.write(keep)
                    digest.update(keep)
                    size += len(keep)
                    if len(keep) != len(raw):
                        self.gaps.append(f"{key}: byte limit exceeded; retained prefix only")
                        complete = False
                        break
                outgoing.flush()
                os.fsync(outgoing.fileno())
            after = os.fstat(incoming.fileno())
            fields = ("st_dev", "st_ino", "st_size", "st_mtime_ns", "st_ctime_ns")
            if any(getattr(before, field) != getattr(after, field) for field in fields) or (
                complete and size != after.st_size
            ):
                self.mismatches.append(f"{key}: source changed during snapshot")
                complete = False
        self.remaining -= size
        self.inventory[key] = dict(file=key, bytes=size, sha256=digest.hexdigest(), complete=complete)
        if complete:
            self.check(f"decode {key}", lambda: self.decode(key))

    def decode(self, key: str) -> None:
        path = self.output / key
        if key.endswith(".jsonl"):
            count = 0
            with path.open("rb") as stream:
                while raw := stream.readline(collect.MAX_LINE_BYTES + 1):
                    _require(len(raw) <= collect.MAX_LINE_BYTES, "JSONL row exceeds 256 KiB")
                    row = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                                     parse_constant=report._constant, parse_float=_finite_float)
                    _require(isinstance(row, dict), "JSONL row must be an object")
                    if key == "sender/submissions.jsonl":
                        report._object(row, {"packet_id", "submitted", "send_start_ns", "send_return_ns",
                                             "scheduled_ns", "lateness_ns", "frame_sha256"}, "submission")
                        report._identifier(row["packet_id"], "packet_id")
                        _require(row["submitted"] is True, "incomplete submission")
                        for field in ("send_start_ns", "send_return_ns", "scheduled_ns", "lateness_ns"):
                            report._integer(row[field], field)
                        _require(row["scheduled_ns"] + row["lateness_ns"] <= row["send_start_ns"] <= row["send_return_ns"],
                                 "submission time order mismatch")
                        _hash(row["frame_sha256"])
                    count += 1
            self.inventory[key]["rows"] = count
        else:
            document = json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=report._pairs,
                                  parse_constant=report._constant, parse_float=_finite_float)
            _require(isinstance(document, dict), "metadata must be an object")
            self.documents[key] = document

    def check(self, label: str, action: Callable[[], None], needs: Sequence[str] = ()) -> bool:
        if any(key not in self.documents for key in needs):
            self.skipped.append(label)
            return False
        try:
            action()
        except (ValueError, TypeError, KeyError, IndexError, RecursionError):
            # Diagnostics identify the boundary, without echoing untrusted raw values/paths.
            self.mismatches.append(label + ": invalid or inconsistent evidence")
            return False
        self.checked.append(label)
        return True

    def receipt(self, role: str, entries: Any, names: Sequence[str]) -> None:
        _require(isinstance(entries, list), "receipt inventory must be an array")
        _require(len(entries) == len(names), "receipt inventory length mismatch")
        seen = set()
        for entry in entries:
            _require(isinstance(entry, dict), "receipt entry must be an object")
            name = entry.get("file")
            _require(isinstance(name, str) and name in names and name not in seen,
                     "unexpected or duplicate receipt file")
            seen.add(name)
            fields = {"file", "bytes", "sha256"} | ({"rows"} if name.endswith(".jsonl") else set())
            report._object(entry, fields, "receipt entry")
            report._integer(entry["bytes"], "bytes")
            _hash(entry["sha256"])
            if "rows" in entry:
                report._integer(entry["rows"], "rows")
            key = role + "/" + name
            retained = self.inventory.get(key)
            if retained is None or not retained["complete"]:
                self.skipped.append("receipt digest " + key)
                continue
            for field in fields - {"file"}:
                _same(entry[field], retained.get(field), f"{key}: {field} mismatch")

    def identical(self, left: str, right: str) -> None:
        a, b = self.inventory.get(left), self.inventory.get(right)
        label = left + " == " + right
        if a is None or b is None or not a["complete"] or not b["complete"]:
            self.skipped.append(label)
            return
        self.check(label, lambda: _same((a["bytes"], a["sha256"]),
                                       (b["bytes"], b["sha256"]), "raw source mismatch"))


def _sender(bundle: _Bundle, document: dict[str, Any]) -> None:
    report._object(document, {"schema_version", "run_id", "origin", "submission_complete",
                             "execution_complete", "slo_compliance_asserted", "departmental_review_required",
                             "link", "eligible_byte_boundary", "inputs_and_outputs", "sender_source_sha256",
                             "limitations"}, "sender receipt")
    _flags(document, "submission_complete")
    report._timestamp(document["origin"])
    _hash(document["sender_source_sha256"])
    _require(document["eligible_byte_boundary"] ==
             "ethernet_frame_without_fcs_preamble_ifg_with_padding", "byte boundary mismatch")
    link = document["link"]
    report._object(link, {"src_mac", "dst_mac", "src_ip", "dst_ip", "src_port", "dst_port"}, "link")
    for key in ("src_mac", "dst_mac"):
        _require(isinstance(link[key], str) and re.fullmatch(r"(?:[0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}", link[key]) is not None,
                 "invalid MAC address")
    for key in ("src_ip", "dst_ip"):
        _require(isinstance(link[key], str), "IPv4 address must be a string")
        ipaddress.IPv4Address(link[key])
    for key in ("src_port", "dst_port"):
        _require(1 <= report._integer(link[key], key) <= 65535, "invalid UDP port")
    bundle.receipt("sender", document["inputs_and_outputs"], SOURCES["sender"][1:])
    rows = [bundle.inventory.get("sender/" + name, {}).get("rows") for name in SOURCES["sender"][1:]]
    if all(value is not None for value in rows):
        _require(rows[0] > 0 and rows[0] == rows[1] == rows[2], "sender row counts differ or empty")


def _capture(document: dict[str, Any]) -> None:
    report._object(document, {"schema_version", "run_id", "capture_closed", "execution_complete",
                             "slo_compliance_asserted", "timestamp_type", "timestamp_precision", "direction",
                             "udp_port", "sent", "dropped", "parse_errors", "ignored", "invalid_markers",
                             "pcap_stats_valid", "pcap_received", "pcap_dropped", "pcap_interface_dropped"},
                   "capture summary")
    _flags(document, "capture_closed", department=False)
    for key, expected in (("timestamp_type", "host"), ("timestamp_precision", "microseconds"),
                          ("direction", "inbound")):
        _require(document[key] == expected, key + " mismatch")
    _require(1 <= report._integer(document["udp_port"], "udp_port") <= 65535, "invalid UDP port")
    report._boolean(document["pcap_stats_valid"], "pcap_stats_valid")
    for key in ("sent", "dropped", "parse_errors", "ignored", "invalid_markers",
                "pcap_received", "pcap_dropped", "pcap_interface_dropped"):
        report._integer(document[key], key)


def _engine(bundle: _Bundle, document: dict[str, Any]) -> None:
    report._object(document, {"schema_version", "run_id", "origin", "export_closed", "execution_complete",
                             "slo_compliance_asserted", "departmental_review_required", "file", "rows", "bytes",
                             "sha256", "clock", "durable_boundary", "event_identity"}, "engine receipt")
    _flags(document, "export_closed")
    report._timestamp(document["origin"])
    _require(document["clock"] == "supplied_live_arrival_and_engine_unix_wall_clock", "clock mismatch")
    _require(document["durable_boundary"] == "successful_full_synchronous_WAL_WriteBatch_return_upper_bound",
             "durable boundary mismatch")
    _require(document["event_identity"] == "slo_ plus SHA256(UTF8(packet_id) + NUL + UTF8(rule_id))",
             "event identity mismatch")
    bundle.receipt("engine", [{key: document[key] for key in ("file", "rows", "bytes", "sha256")}],
                   ("events.jsonl",))


def _adapter(bundle: _Bundle, document: dict[str, Any]) -> None:
    report._object(document, {"schema_version", "run_id", "adapter_complete", "adapter", "adapter_source_sha256",
                             "inputs_and_output", "slo_compliance_asserted", "departmental_review_required",
                             "evidence_class", "limitations"}, "adapter receipt")
    _flags(document, "adapter_complete", component=False)
    _require(document["adapter"] == "scripts/slo_collect.py", "adapter mismatch")
    _hash(document["adapter_source_sha256"])
    _require(document["evidence_class"] == "single_vm_exported_packet_ledgers", "evidence class mismatch")
    bundle.receipt("adapter", document["inputs_and_output"], SOURCES["adapter"][1:])


def _manifest_observations(manifest: dict[str, Any], observations: dict[str, Any]) -> None:
    _version(manifest)
    _version(observations)
    collect._manifest(manifest)
    report._validate(observations)
    projection = copy.deepcopy(observations)
    del projection["expected_alerts"]
    for window in projection["packet_windows"]:
        for key in ("offered_eligible", "fully_processed", "offered_bytes"):
            del window[key]
    _same(manifest, projection, "manifest metadata differs from observations")


def _summary(bundle: _Bundle, summary: dict[str, Any], observations: dict[str, Any]) -> None:
    source = summary["source"]
    report._object(source, {"sha256", "raw_json"}, "report source")
    _hash(source["sha256"])
    _require(isinstance(source["raw_json"], str), "report source must retain raw JSON")
    raw = (bundle.output / "adapter/observations.json").read_bytes()
    _require(source["raw_json"].encode("utf-8") == raw, "embedded observations differ")
    _require(source["sha256"] == hashlib.sha256(raw).hexdigest(), "report source digest differs")
    expected = report.summarize(observations)
    expected["source"] = source
    _same(summary, expected, "report differs from recomputed supplied-observation summary")


def reconcile(output: Path, *, sender: Path | None = None, capture: Path | None = None,
              engine: Path | None = None, adapter: Path | None = None,
              summary: Path | None = None, max_bytes: int = DEFAULT_MAX_BYTES,
              reconstruct_ledgers: bool = False, max_reconstruction_bytes: int = DEFAULT_MAX_BYTES,
              scratch_dir: Path | None = None) -> dict[str, Any]:
    """Snapshot fixed filenames and publish bundle.json last; no live work occurs."""
    report._integer(max_bytes, "max_bytes", 1)
    report._boolean(reconstruct_ledgers, "reconstruct_ledgers")
    report._integer(max_reconstruction_bytes, "max_reconstruction_bytes", 1)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    bundle = _Bundle(output, max_bytes)
    roots = dict(sender=sender, capture=capture, engine=engine, adapter=adapter, report=summary)
    for role, names in SOURCES.items():
        root = roots[role]
        for name in names:
            source = None
            if root is not None:
                source = root.expanduser()
                if role not in ("capture", "report"):
                    source = source / name
            bundle.snapshot(source, role + "/" + name)
    documents = bundle.documents
    validators = {
        "sender/submission.json": lambda doc: _sender(bundle, doc),
        "capture/summary.json": _capture,
        "engine/close.json": lambda doc: _engine(bundle, doc),
        "adapter/receipt.json": lambda doc: _adapter(bundle, doc),
        "adapter/manifest.json": collect._manifest,
        "adapter/observations.json": report._validate,
    }
    valid = {}
    for key, validate in validators.items():
        if bundle.check("schema/receipt " + key, lambda key=key, validate=validate: validate(documents[key]), (key,)):
            valid[key] = documents[key]
    identities = [(key, doc["run_id"]) for key, doc in valid.items()]
    if identities:
        bundle.check("run IDs", lambda: _require(len({value for _, value in identities}) == 1, "run IDs differ"))
    origins = [doc["origin"] if "origin" in doc else doc["started_at"]
               for doc in valid.values() if "origin" in doc or "started_at" in doc]
    if origins:
        bundle.check("origins", lambda: _require(len(set(origins)) == 1, "origins differ"))
    sender_doc, capture_doc = valid.get("sender/submission.json"), valid.get("capture/summary.json")
    if sender_doc is not None and capture_doc is not None:
        bundle.check("UDP destination port", lambda: _same(sender_doc["link"]["dst_port"],
                                                           capture_doc["udp_port"], "ports differ"))
    else:
        bundle.skipped.append("UDP destination port")
    fixture = bundle.inventory.get("sender/fixture.jsonl")
    manifest = valid.get("adapter/manifest.json")
    if fixture is not None and fixture["complete"] and manifest is not None:
        bundle.check("fixture provenance digest", lambda: _same(fixture["sha256"],
                     manifest["provenance"]["fixture_sha256"], "fixture digest differs"))
    else:
        bundle.skipped.append("fixture provenance digest")
    bundle.identical("sender/offered.jsonl", "adapter/offered.jsonl")
    bundle.identical("engine/events.jsonl", "adapter/events.jsonl")
    bundle.check("manifest/observation metadata", lambda: _manifest_observations(
        documents["adapter/manifest.json"], documents["adapter/observations.json"]),
        ("adapter/manifest.json", "adapter/observations.json"))
    bundle.check("report source and recomputed summary", lambda: _summary(
        bundle, documents["report/report.json"], documents["adapter/observations.json"]),
        ("report/report.json", "adapter/observations.json"))
    reconstruction = None
    errors: list[str] = []
    if reconstruct_ledgers:
        # Lazy import: reconstruction reuses this module's snapshot/schema helpers.
        if __package__:
            from . import slo_reconstruct as replay
        else:
            import slo_reconstruct as replay
        reconstruction, gaps, mismatches, errors = replay.integrate(
            bundle, max_bytes=max_reconstruction_bytes, scratch_dir=scratch_dir)
        bundle.gaps.extend(gaps)
        bundle.mismatches.extend(mismatches)
    status = "error" if errors else "mismatch" if bundle.mismatches else (
        "incomplete" if bundle.gaps or bundle.skipped else "review_required")
    result = {
        "schema_version": 1, "artifact_kind": "netsentry_slo_evidence_bundle",
        "status": status, "bundle_complete": status == "review_required",
        "snapshots_complete": not bundle.gaps and all(item["complete"] for item in bundle.inventory.values()),
        "cross_checks_complete": status == "review_required",
        "execution_complete": False, "slo_compliance_asserted": False,
        "departmental_review_required": True,
        "evidence_class": "single_vm_supplied_artifact_snapshots",
        "max_retained_input_bytes": max_bytes,
        "bundle_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "reporter_source_sha256": hashlib.sha256(Path(report.__file__).read_bytes()).hexdigest(),
        "inventory": list(bundle.inventory.values()),
        "supplied_run_ids": dict(identities), "supplied_origins": origins,
        "gaps": bundle.gaps, "mismatches": bundle.mismatches,
        "checks_completed": bundle.checked, "checks_not_performed": bundle.skipped,
        "reported_measurement_status": documents.get("report/report.json", {}).get("status"),
        "limitations": [
            "Consistency is not authenticity, SLO compliance or measured capacity.",
            "Component closure does not prove process exit success or completed cohort/drain accounting.",
            "JSONL syntax/counts, submission success/times and receipts are checked; packet identity/oracle replay is not performed.",
            "Submitted frames, physical arrival, clock accuracy and storage durability require departmental review.",
            "Capture counters are retained diagnostics; missing pcap stats and drops require review, not denominator filtering.",
            "Source-code digests are declarations; source versions/config/rules are not authenticated here.",
            "Report metrics are recomputed from supplied observations, not reconstructed from raw packet ledgers.",
            "Default 64 GiB retained-input budget may need increasing for extended acceptance runs.",
        ],
    }
    if reconstruct_ledgers:
        result.update(schema_version=2, reconstruction_policy="retained_adapter_replay_v1",
                      reconstruction=reconstruction, errors=errors,
                      max_reconstruction_bytes=max_reconstruction_bytes,
                      reconstruction_source_sha256=hashlib.sha256(Path(replay.__file__).read_bytes()).hexdigest())
        result["limitations"][2] = "Packet identity/oracle replay is required in this mode; inspect reconstruction and source binding."
        result["limitations"][6] = "Fresh raw-ledger derivation is checked when inputs qualify; acquisition truth remains unverified."
        result["limitations"].append("Reconstruction adds its own input copies, rebuilt outputs and temporary SQLite disk use.")
    # Persist child directory entries before the completion manifest is published.
    for directory in [path for path in output.iterdir() if path.is_dir()]:
        descriptor = os.open(directory, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(descriptor)
        finally:
            os.close(descriptor)
    report.write_report(output / "bundle.json", result)
    descriptor = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("sender", "capture", "engine", "adapter", "summary"):
        parser.add_argument("--" + name, type=Path, help="supplied input; omission is retained as an evidence gap")
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--max-bytes", type=int, default=DEFAULT_MAX_BYTES, help="total retained input byte budget")
    parser.add_argument("--reconstruct-ledgers", action="store_true", help="require fresh retained adapter replay")
    parser.add_argument("--max-reconstruction-bytes", type=int, default=DEFAULT_MAX_BYTES)
    parser.add_argument("--scratch-dir", type=Path, help="temporary reconstruction SQLite index directory")
    args = parser.parse_args(argv)
    try:
        result = reconcile(args.output_dir, sender=args.sender, capture=args.capture, engine=args.engine,
                           adapter=args.adapter, summary=args.summary, max_bytes=args.max_bytes,
                           reconstruct_ledgers=args.reconstruct_ledgers,
                           max_reconstruction_bytes=args.max_reconstruction_bytes, scratch_dir=args.scratch_dir)
    except (OSError, ValueError, RecursionError) as error:
        print(f"[slo-bundle] {error}; retain partial output and use a new directory", file=sys.stderr)
        return 2
    print(f"[slo-bundle] {result['status']}; departmental review required: {args.output_dir}")
    return {"review_required": 0, "mismatch": 1, "error": 2, "incomplete": 3}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
