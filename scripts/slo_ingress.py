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
import struct
import sys
import time
from typing import Any, Callable, Sequence

if __package__:
    from . import slo_collect as collector, slo_report as report
else:
    import slo_collect as collector
    import slo_report as report


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


def send_fixture(fixture: Path, output: Path, run_id: str, origin: str,
                 link: dict[str, Any], send: Callable[[bytes], int]) -> dict[str, Any]:
    """Submit a finalized external fixture; send must return the full frame length.

    Retain every initiated offer before send; errors leave partial artifacts and
    no completed submission receipt. Successful send is not proof of NIC delivery.
    """
    report._identifier(run_id, "run_id")
    started = report._timestamp(origin)
    if not 2000 <= started.year < 2100:
        raise report.EvidenceError("origin must be in [2000,2100)")
    origin_ns = int(started.timestamp()) * report.NS
    _link(link)
    link = dict(link)
    output = output.expanduser()
    output.mkdir(mode=0o700, parents=False, exist_ok=False)
    source_receipts: list[dict[str, Any]] = []
    ledgers: list[_Ledger] = []
    previous = -1
    try:
        offers = _Ledger(output / "offered.jsonl")
        ledgers.append(offers)
        submissions = _Ledger(output / "submissions.jsonl")
        ledgers.append(submissions)
        rows = collector._rows(fixture.expanduser(), output / "fixture.jsonl", source_receipts)
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
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args(argv)
    link = {key: getattr(args, key) for key in ("src_mac", "dst_mac", "src_ip", "dst_ip", "src_port", "dst_port")}
    try:
        _link(link)
        with socket.socket(socket.AF_PACKET, socket.SOCK_RAW, socket.htons(0x0800)) as sock:
            sock.bind((args.interface, 0))
            send_fixture(args.fixture, args.output_dir, args.run_id, args.origin, link, sock.send)
    except (OSError, ValueError, RecursionError, KeyboardInterrupt) as error:
        print(f"[slo-ingress] {error}; retain partial artifacts; no acceptance claim", file=sys.stderr)
        return 2
    print(f"[slo-ingress] submission complete; capture/drain/departmental review still required: {args.output_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
