#!/usr/bin/env python3
"""Reconstruct reference-sender evidence offline from four retained source files."""

from __future__ import annotations

import argparse
import base64
from contextlib import ExitStack
import hashlib
import json
import os
from pathlib import Path
import sys
from typing import Any, BinaryIO, Sequence

if __package__:
    from . import slo_bundle as bundle, slo_collect as collect, slo_ingress as ingress, slo_report as report
else:
    import slo_bundle as bundle
    import slo_collect as collect
    import slo_ingress as ingress
    import slo_report as report

DEFAULT_MAX_BYTES = bundle.DEFAULT_MAX_BYTES
LEDGERS = bundle.SOURCES["sender"][1:]


class _Rows:
    """One bounded decoded row at a time, with a fresh inventory of consumed bytes."""

    def __init__(self, stream: BinaryIO):
        self.stream = stream
        self.digest = hashlib.sha256()
        self.size = self.count = 0

    def read(self) -> dict[str, Any] | None:
        raw = self.stream.readline(collect.MAX_LINE_BYTES + 1)
        if not raw:
            return None
        bundle._require(len(raw) <= collect.MAX_LINE_BYTES, "row exceeds 256 KiB")
        self.digest.update(raw)
        self.size += len(raw)
        self.count += 1
        row = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                         parse_constant=report._constant, parse_float=bundle._finite_float)
        bundle._require(isinstance(row, dict), "row must be an object")
        return row

    def verify(self, retained: dict[str, Any]) -> None:
        bundle._same((self.size, self.count, self.digest.hexdigest()),
                     (retained["bytes"], retained["rows"], retained["sha256"]),
                     "replayed ledger differs from snapshot")


def _correlate(fixture: dict[str, Any], offer: dict[str, Any], submission: dict[str, Any],
               receipt: dict[str, Any], sequence: int, previous: int,
               progress: dict[str, Any]) -> int:
    progress["boundary"] = "fixture schema and schedule"
    report._object(fixture, {"offset_ns", "payload_base64", "expected_rule_ids"}, "fixture")
    offset = report._integer(fixture["offset_ns"], "offset_ns")
    bundle._require(previous <= offset <= 7 * 24 * 3600 * report.NS, "invalid fixture offset")
    progress["boundary"] = "fixture payload and rule oracle"
    bundle._require(isinstance(fixture["payload_base64"], str), "payload_base64 must be a string")
    payload = base64.b64decode(fixture["payload_base64"], validate=True)
    rules = fixture["expected_rule_ids"]
    bundle._require(isinstance(rules, list), "expected_rule_ids must be an array")
    for rule in rules:
        report._identifier(rule, "rule_id")
    bundle._require(len(set(rules)) == len(rules), "duplicate expected rule")
    packet_id = f"pkt-{sequence}"
    expected = [{"rule_id": rule, "event_id": "slo_" + hashlib.sha256(
        (packet_id + "\0" + rule).encode("utf-8")).hexdigest()} for rule in rules]
    progress["boundary"] = "pure frame construction"
    raw = ingress.frame(receipt["run_id"], packet_id, payload, receipt["link"], sequence)
    progress["boundary"] = "offered packet identity, oracle and frame length"
    report._object(offer, {"packet_id", "offered_ns", "offered_bytes", "expected_alerts"}, "offer")
    bundle._same(offer["packet_id"], packet_id, "offer packet sequence differs")
    bundle._same(offer["expected_alerts"], expected, "offered oracle differs")
    bundle._same(offer["offered_bytes"], len(raw), "offered frame length differs")
    offered = report._integer(offer["offered_ns"], "offered_ns")
    progress["boundary"] = "submission identity and frame hash"
    report._object(submission, {"packet_id", "submitted", "send_start_ns", "send_return_ns",
                               "scheduled_ns", "lateness_ns", "frame_sha256"}, "submission")
    bundle._same(submission["packet_id"], packet_id, "submission packet sequence differs")
    bundle._require(submission["submitted"] is True, "submission did not complete")
    bundle._same(submission["frame_sha256"], hashlib.sha256(raw).hexdigest(), "frame hash differs")
    progress["boundary"] = "scheduled, offered and submission timing"
    for field in ("scheduled_ns", "lateness_ns", "send_start_ns", "send_return_ns"):
        report._integer(submission[field], field)
    bundle._same(submission["scheduled_ns"], offset, "scheduled offset differs")
    bundle._same(offered, offset + submission["lateness_ns"], "offered time differs from schedule plus lateness")
    bundle._require(offset <= offered <= submission["send_start_ns"] <= submission["send_return_ns"],
                    "offer/send time order differs")
    return offset


def _replay(retained: bundle._Bundle, receipt: dict[str, Any], progress: dict[str, Any]) -> None:
    with ExitStack() as stack:
        readers = [_Rows(stack.enter_context((retained.output / "sender" / name).open("rb")))
                   for name in LEDGERS]
        previous = -1
        while True:
            progress.update(row=progress["rows_matched"] + 1, boundary="aligned ledger decode")
            rows = []
            for name, reader in zip(LEDGERS, readers):
                progress["boundary"] = "decode " + name
                rows.append(reader.read())
            progress["boundary"] = "ledger row alignment"
            if all(row is None for row in rows):
                break
            bundle._require(all(row is not None for row in rows), "unequal ledger lengths")
            previous = _correlate(*rows, receipt, progress["row"], previous, progress)
            progress["rows_matched"] += 1
        bundle._require(progress["rows_matched"] > 0, "empty sender ledgers")
        progress.update(row=None, boundary="fresh replay inventory binding")
        for name, reader in zip(LEDGERS, readers):
            reader.verify(retained.inventory["sender/" + name])
        progress["boundary"] = "close replay inputs"
    progress.update(row=None, boundary=None)


def reconstruct(output: Path, *, sender: Path | None = None,
                max_bytes: int = DEFAULT_MAX_BYTES) -> dict[str, Any]:
    """Retain and correlate supplied bytes; never send, wait, or certify a run.

    Publication/I/O/interruption can leave partial output without a manifest.
    Preserve it and retry to a fresh path; inspect process exit as well as status.
    """
    report._integer(max_bytes, "max_bytes", 1)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    retained = bundle._Bundle(output, max_bytes)
    source = sender.expanduser() if sender is not None else None
    errors: list[str] = []
    attempted = complete = receipt_valid = False
    receipt: dict[str, Any] = {}
    progress: dict[str, Any] = {"row": None, "rows_matched": 0, "boundary": "source snapshots"}
    try:
        for name in bundle.SOURCES["sender"]:
            progress["boundary"] = "snapshot sender/" + name
            retained.snapshot(None if source is None else source / name, "sender/" + name)
        receipt = retained.documents.get("sender/submission.json", {})

        def validate_receipt() -> None:
            bundle._sender(retained, receipt)
            bundle._require(2000 <= report._timestamp(receipt["origin"]).year < 2100,
                            "sender origin outside supported years")

        progress["boundary"] = "original sender receipt and inventory"
        receipt_valid = retained.check(progress["boundary"], validate_receipt, ("sender/submission.json",))
        if receipt_valid and not (retained.gaps or retained.mismatches or retained.skipped):
            attempted = True
            _replay(retained, receipt, progress)
            complete = True
            retained.checked.append("aligned fixture/oracle/submission derivation and fresh inventory")
        else:
            retained.skipped.append("sender replay requires complete consistent sender sources")
    except (ValueError, TypeError, KeyError, IndexError, OverflowError, RecursionError):
        retained.mismatches.append("sender reconstruction: invalid or inconsistent evidence")
    except OSError:
        errors.append("sender reconstruction I/O failed; retain partial output")
    status = "error" if errors else "mismatch" if retained.mismatches else (
        "incomplete" if retained.gaps or retained.skipped or not complete else "review_required")
    result = {
        "schema_version": 1, "artifact_kind": "netsentry_slo_sender_reconstruction", "status": status,
        "reconstruction_attempted": attempted, "reconstruction_complete": complete,
        "sender_records_match": True if complete else None,
        "progress": progress, "run_id": receipt.get("run_id") if receipt_valid else None,
        "origin": receipt.get("origin") if receipt_valid else None,
        "link": receipt.get("link") if receipt_valid else None,
        "original_sender_source_sha256": receipt.get("sender_source_sha256") if receipt_valid else None,
        "max_retained_input_bytes": max_bytes, "inventory": list(retained.inventory.values()),
        "gaps": retained.gaps, "mismatches": retained.mismatches, "errors": errors,
        "checks_completed": retained.checked, "checks_not_performed": retained.skipped,
        "reconstruction_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "sender_source_sha256": hashlib.sha256(Path(ingress.__file__).read_bytes()).hexdigest(),
        "bundle_source_sha256": hashlib.sha256(Path(bundle.__file__).read_bytes()).hexdigest(),
        "collector_source_sha256": hashlib.sha256(Path(collect.__file__).read_bytes()).hexdigest(),
        "reporter_source_sha256": hashlib.sha256(Path(report.__file__).read_bytes()).hexdigest(),
        "execution_complete": False, "facts_verified": False, "slo_compliance_asserted": False,
        "comparability_established": False, "departmental_review_required": True,
        "limitations": [
            "Agreement establishes derivation of supplied bytes only, not authenticity or complete acquisition.",
            "No send, NIC delivery, clock accuracy, actual scheduling or rule-engine oracle is verified.",
            "Current pure frame construction is used; source digests are not authenticated build identities.",
            "Expected alert order is significant because the reference sender preserves fixture rule order.",
            "Input budget excludes generated metadata; replay retains at most one bounded row per ledger.",
            "A matching prefix is not a complete reconstruction; failure diagnostics stop at the first boundary.",
            "Runtime correctness, performance and departmental acceptance remain untested.",
            "Partial output may lack complete inventory or manifest; retain it and retry to a new directory.",
            "This standalone receipt is not consumed by bundle or pair review.",
        ],
    }
    directory = output / "sender"
    if directory.is_dir():
        descriptor = os.open(directory, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(descriptor)
        finally:
            os.close(descriptor)
    report.write_report(output / "sender-reconstruction.json", result)
    descriptor = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sender", type=Path, help="supplied sender directory; omission records incomplete evidence")
    parser.add_argument("--output-dir", type=Path, required=True, help="new directory in an existing parent")
    parser.add_argument("--max-bytes", type=int, default=DEFAULT_MAX_BYTES, help="total retained input byte budget")
    args = parser.parse_args(argv)
    try:
        result = reconstruct(args.output_dir, sender=args.sender, max_bytes=args.max_bytes)
    except (OSError, ValueError, RecursionError, KeyboardInterrupt):
        print("[slo-sender-reconstruct] operation failed; retain partial output and retry to a new directory",
              file=sys.stderr)
        return 2
    print(f"[slo-sender-reconstruct] {result['status']}; departmental review required")
    return {"review_required": 0, "mismatch": 1, "error": 2, "incomplete": 3}[result["status"]]


if __name__ == "__main__":
    raise SystemExit(main())
