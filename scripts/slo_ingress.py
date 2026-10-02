#!/usr/bin/env python3
"""Reference UDP measurement sender with an independent retained offered oracle.

This sends traffic only when explicitly invoked; it does not certify capacity.
"""
from __future__ import annotations

import argparse
import base64
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import re
import socket
import stat
import struct
import sys
import time
from typing import Any, BinaryIO, Callable, Iterator, Sequence

if __package__:
    from . import slo_collect as collector, slo_report as report
else:
    import slo_collect as collector
    import slo_report as report


DEFAULT_MAX_FIXTURE_BYTES = collector.DEFAULT_MAX_INPUT_BYTES


def _checksum(raw: bytes) -> int:
    raw += b"\0" if len(raw) % 2 else b""
    total = sum(struct.unpack(f"!{len(raw) // 2}H", raw))
    while total >> 16:
        total = (total & 0xffff) + (total >> 16)
    return (~total) & 0xffff


def _mac(value: Any) -> bytes:
    if not isinstance(value, str) or not re.fullmatch(r"(?:[0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}", value):
        raise report.EvidenceError("MAC address must contain six colon-separated hex octets")
    return bytes.fromhex(value.replace(":", ""))


def _link(link: dict[str, Any]) -> None:
    report._object(link, {"src_mac", "dst_mac", "src_ip", "dst_ip", "src_port", "dst_port"}, "link")
    for key in ("src_mac", "dst_mac"):
        _mac(link[key])
    for key in ("src_ip", "dst_ip"):
        if not isinstance(link[key], str):
            raise report.EvidenceError("IPv4 addresses must be strings")
        ipaddress.IPv4Address(link[key])
    for key in ("src_port", "dst_port"):
        if report._integer(link[key], key, 1) > 65535:
            raise report.EvidenceError("UDP port exceeds 65535")


def frame(run_id: str, packet_id: str, payload: bytes, link: dict[str, Any], sequence: int) -> bytes:
    """Build untagged Ethernet/IPv4/UDP, with checksums, DF and Ethernet padding."""
    report._identifier(run_id, "run_id")
    report._identifier(packet_id, "packet_id")
    report._integer(sequence, "sequence", 1)
    _link(link)
    marked = f"NSLO1 {run_id} {packet_id}\n".encode("ascii") + payload
    if len(marked) > 1472:
        raise report.EvidenceError("marker plus UDP payload exceeds IPv4 MTU 1500")
    src, dst = (ipaddress.IPv4Address(link[key]).packed for key in ("src_ip", "dst_ip"))
    udp = struct.pack("!HHHH", link["src_port"], link["dst_port"], 8 + len(marked), 0)
    pseudo = src + dst + struct.pack("!BBH", 0, 17, len(udp) + len(marked))
    udp_sum = _checksum(pseudo + udp + marked) or 0xffff
    udp = udp[:6] + struct.pack("!H", udp_sum) + marked
    ip = struct.pack("!BBHHHBBH4s4s", 0x45, 0, 20 + len(udp), sequence & 0xffff,
                     0x4000, 64, 17, 0, src, dst)
    ip = ip[:10] + struct.pack("!H", _checksum(ip)) + ip[12:]
    raw = _mac(link["dst_mac"]) + _mac(link["src_mac"]) + b"\x08\x00" + ip + udp
    return raw + b"\0" * max(0, 60 - len(raw))


class _Ledger:
    def __init__(self, path: Path):
        self.path = path
        self.stream = os.fdopen(os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600), "wb", buffering=0)
        self.digest = hashlib.sha256()
        self.rows = self.size = 0

    def append(self, row: dict[str, Any]) -> None:
        raw = (json.dumps(row, separators=(",", ":"), allow_nan=False) + "\n").encode("utf-8")
        if len(raw) > collector.MAX_LINE_BYTES:
            raise report.EvidenceError("generated ledger row exceeds adapter 256 KiB limit")
        if self.stream.write(raw) != len(raw):
            raise OSError("short ledger write")
        self.digest.update(raw)
        self.rows += 1
        self.size += len(raw)

    def close(self) -> None:
        try:
            os.fsync(self.stream.fileno())
        finally:
            self.stream.close()

    def receipt(self) -> dict[str, Any]:
        return {"file": self.path.name, "rows": self.rows, "bytes": self.size, "sha256": self.digest.hexdigest()}


def _fixture_open(path: Path) -> BinaryIO:
    descriptor = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    try:
        return os.fdopen(descriptor, "rb")
    except BaseException:
        os.close(descriptor)
        raise


def _fixture_metadata(stream: BinaryIO) -> os.stat_result:
    metadata = os.fstat(stream.fileno())
    if not stat.S_ISREG(metadata.st_mode):
        raise report.EvidenceError("fixture must be a regular file")
    if any(not hasattr(metadata, field) for field in collector._INPUT_STAT_FIELDS):
        raise report.EvidenceError("stable fixture metadata unavailable")
    return metadata


def _snapshot_fixture(source: Path, retained: Path, max_bytes: int
                      ) -> tuple[dict[str, Any], os.stat_result]:
    """Close a bounded, synced source snapshot before any submission starts."""
    digest = hashlib.sha256()
    size = count = 0
    with _fixture_open(source) as incoming:
        before = _fixture_metadata(incoming)
        if before.st_size > max_bytes:
            raise report.EvidenceError("fixture exceeds max_fixture_bytes")
        with retained.open("xb", buffering=0) as outgoing:
            os.fchmod(outgoing.fileno(), 0o600)
            while True:
                raw = incoming.readline(min(collector.MAX_LINE_BYTES, max_bytes - size) + 1)
                if not raw:
                    break
                if len(raw) > max_bytes - size:
                    raise report.EvidenceError("fixture exceeds max_fixture_bytes")
                if len(raw) > collector.MAX_LINE_BYTES:
                    raise report.EvidenceError(f"{retained.name} line {count + 1}: exceeds 256 KiB")
                if outgoing.write(raw) != len(raw):
                    raise OSError("short fixture snapshot write")
                digest.update(raw)
                size += len(raw)
                count += 1
            if size != before.st_size:
                raise report.EvidenceError("fixture size differs from admitted source")
            outgoing.flush()
            os.fsync(outgoing.fileno())
            snapshot = _fixture_metadata(outgoing)
            if snapshot.st_size != size:
                raise OSError("fixture snapshot size mismatch")
        collector._input_stat(incoming, before, "fixture")
    return {"file": retained.name, "bytes": size, "rows": count,
            "sha256": digest.hexdigest()}, snapshot


def _fixture_rows(retained: Path, receipt: dict[str, Any], snapshot: os.stat_result
                  ) -> Iterator[dict[str, Any]]:
    """Interpret only the retained snapshot and verify its inventory at EOF."""
    digest = hashlib.sha256()
    size = count = 0
    with _fixture_open(retained) as incoming:
        _fixture_metadata(incoming)
        collector._input_stat(incoming, snapshot, "retained fixture")
        while True:
            raw = incoming.readline(min(collector.MAX_LINE_BYTES, receipt["bytes"] - size) + 1)
            if not raw:
                break
            if len(raw) > receipt["bytes"] - size:
                raise report.EvidenceError("retained fixture exceeds snapshot bytes")
            count += 1
            if len(raw) > collector.MAX_LINE_BYTES:
                raise report.EvidenceError(f"{retained.name} line {count}: exceeds 256 KiB")
            try:
                row = json.loads(raw.decode("utf-8"), object_pairs_hook=report._pairs,
                                 parse_constant=report._constant)
            except (UnicodeError, ValueError, RecursionError) as error:
                raise report.EvidenceError(f"{retained.name} line {count}: invalid JSON row") from error
            if not isinstance(row, dict):
                raise report.EvidenceError(f"{retained.name} line {count}: expected object")
            digest.update(raw)
            size += len(raw)
            yield row
        collector._input_stat(incoming, snapshot, "retained fixture")
        if (size, count, digest.hexdigest()) != (receipt["bytes"], receipt["rows"], receipt["sha256"]):
            raise report.EvidenceError("retained fixture differs from snapshot inventory")


def send_fixture(fixture: Path, output: Path, run_id: str, origin: str,
                 link: dict[str, Any], send: Callable[[bytes], int], *,
                 max_fixture_bytes: int = DEFAULT_MAX_FIXTURE_BYTES) -> dict[str, Any]:
    """Submit a finalized external fixture; send must return the full frame length.

    Retain every initiated offer before send; errors leave partial artifacts and
    no completed submission receipt. Successful send is not proof of NIC delivery.
    """
    report._integer(max_fixture_bytes, "max_fixture_bytes", 1)
    report._identifier(run_id, "run_id")
    started = report._timestamp(origin)
    if not 2000 <= started.year < 2100:
        raise report.EvidenceError("origin must be in [2000,2100)")
    origin_ns = int(started.timestamp()) * report.NS
    _link(link)
    link = dict(link)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    fixture_receipt, snapshot = _snapshot_fixture(
        fixture.expanduser(), output / "fixture.jsonl", max_fixture_bytes)
    source_receipts = [fixture_receipt]
    ledgers: list[_Ledger] = []
    previous = -1
    try:
        offers = _Ledger(output / "offered.jsonl")
        ledgers.append(offers)
        submissions = _Ledger(output / "submissions.jsonl")
        ledgers.append(submissions)
        rows = _fixture_rows(output / "fixture.jsonl", fixture_receipt, snapshot)
        try:
            for sequence, row in enumerate(rows, 1):
                report._object(row, {"offset_ns", "payload_base64", "expected_rule_ids"}, "fixture row")
                offset = report._integer(row["offset_ns"], "offset_ns")
                if offset < previous or offset > 7 * 24 * 3600 * report.NS:
                    raise report.EvidenceError("offsets must be nondecreasing and within seven days")
                previous = offset
                if not isinstance(row["payload_base64"], str):
                    raise report.EvidenceError("payload_base64 must be a string")
                payload = base64.b64decode(row["payload_base64"], validate=True)
                rules = row["expected_rule_ids"]
                if not isinstance(rules, list):
                    raise report.EvidenceError("expected_rule_ids must be an array")
                for rule in rules:
                    report._identifier(rule, "rule_id")
                if len(set(rules)) != len(rules):
                    raise report.EvidenceError("duplicate expected rule")
                packet_id = f"pkt-{sequence}"
                raw = frame(run_id, packet_id, payload, link, sequence)
                expected = [{"rule_id": rule, "event_id": "slo_" + hashlib.sha256(
                    (packet_id + "\0" + rule).encode("utf-8")).hexdigest()} for rule in rules]
                while (remaining := origin_ns + offset - time.time_ns()) > 0:
                    time.sleep(min(remaining / report.NS, 0.25))
                offered_ns = time.time_ns() - origin_ns
                report._integer(offered_ns, "actual offered_ns")
                offers.append({"packet_id": packet_id, "offered_ns": offered_ns,
                               "offered_bytes": len(raw), "expected_alerts": expected})
                send_start = time.time_ns() - origin_ns
                if send_start < offered_ns:
                    raise report.EvidenceError("wall clock moved backward before send")
                try:
                    sent = send(raw)
                    if type(sent) is not int or sent != len(raw):
                        raise OSError("send did not submit the complete frame")
                except Exception:
                    submissions.append({"packet_id": packet_id, "submitted": False,
                                        "send_start_ns": send_start, "send_return_ns": time.time_ns() - origin_ns})
                    raise
                returned = time.time_ns() - origin_ns
                submissions.append({"packet_id": packet_id, "submitted": True,
                                    "send_start_ns": send_start, "send_return_ns": returned,
                                    "scheduled_ns": offset, "lateness_ns": offered_ns - offset,
                                    "frame_sha256": hashlib.sha256(raw).hexdigest()})
                if returned < send_start:
                    raise report.EvidenceError("wall clock moved backward during send")
        finally:
            rows.close()
    finally:
        # Attempt every close even if one fsync fails. A close error prevents receipt.
        close_errors = []
        for ledger in ledgers:
            try:
                ledger.close()
            except OSError as error:
                close_errors.append(error)
        if close_errors:
            raise OSError("failed to sync/close sender ledger") from close_errors[0]
    if not offers.rows:
        raise report.EvidenceError("empty fixture cannot complete a submission run")
    receipt = {"schema_version": 1, "run_id": run_id, "origin": origin,
               "submission_complete": True, "execution_complete": False,
               "slo_compliance_asserted": False, "departmental_review_required": True,
               "link": link, "eligible_byte_boundary": "ethernet_frame_without_fcs_preamble_ifg_with_padding",
               "inputs_and_outputs": source_receipts + [ledger.receipt() for ledger in ledgers],
               "sender_source_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
               "limitations": ["Successful send only establishes kernel submission, not NIC delivery.",
                               "Offer time precedes ledger write and send; inspect scheduling/serialization delay.",
                               "UDP marker remains in payload; fixture and oracle must include it.",
                               "Fixture snapshot preparation can increase lateness; source limits exclude generated ledgers.",
                               "Reference sender throughput and overhead are unmeasured."]}
    report.write_report(output / "submission.json", receipt)
    parent_fd = os.open(output.parent, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(parent_fd)
    finally:
        os.close(parent_fd)
    return receipt


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    for flag in ("interface", "run-id", "origin", "src-mac", "dst-mac", "src-ip", "dst-ip"):
        parser.add_argument("--" + flag, required=True)
    for flag in ("src-port", "dst-port"):
        parser.add_argument("--" + flag, type=int, required=True)
    parser.add_argument("--fixture", type=Path, required=True)
    parser.add_argument("--max-fixture-bytes", type=int, default=DEFAULT_MAX_FIXTURE_BYTES)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args(argv)
    link = {key: getattr(args, key) for key in ("src_mac", "dst_mac", "src_ip", "dst_ip", "src_port", "dst_port")}
    try:
        report._integer(args.max_fixture_bytes, "max_fixture_bytes", 1)
        _link(link)
        with socket.socket(socket.AF_PACKET, socket.SOCK_RAW, socket.htons(0x0800)) as sock:
            sock.bind((args.interface, 0))
            send_fixture(args.fixture, args.output_dir, args.run_id, args.origin, link, sock.send,
                         max_fixture_bytes=args.max_fixture_bytes)
    except (OSError, ValueError, RecursionError, KeyboardInterrupt) as error:
        print(f"[slo-ingress] {error}; retain partial artifacts; no acceptance claim", file=sys.stderr)
        return 2
    print(f"[slo-ingress] submission complete; capture/drain/departmental review still required: {args.output_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
