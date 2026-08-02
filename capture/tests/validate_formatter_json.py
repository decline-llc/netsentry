#!/usr/bin/env python3
"""Decode formatter seed JSONL and enforce the three UDS frame shapes."""

from __future__ import annotations

import base64
import json
import math
import os
import subprocess
import sys


PACKET_KEYS = {
    "timestamp_sec", "timestamp_usec", "src_ip", "dst_ip", "src_port",
    "dst_port", "protocol", "tcp_flags", "payload_len", "payload_preview",
    "is_fragment", "truncated",
}
HEARTBEAT_KEYS = {
    "type", "session_id", "seq", "sent", "dropped", "parse_errors",
    "buf_util_pct", "avg_json_serialize_us", "uds_write_errors",
}
HELLO_KEYS = {
    "type", "version", "session_id", "pid", "hostname", "max_payload_len",
}


def reject_constant(value: str) -> None:
    raise ValueError(f"non-finite JSON number: {value}")


def unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate JSON member: {key}")
        result[key] = value
    return result


def require_exact_type(value: object, expected: type, field: str) -> None:
    if type(value) is not expected:
        raise ValueError(f"{field} must be {expected.__name__}")


def validate_packet(frame: dict[str, object]) -> None:
    if set(frame) != PACKET_KEYS:
        raise ValueError("packet keys do not match the UDS contract")
    for field in (
        "timestamp_sec", "timestamp_usec", "src_port", "dst_port", "protocol",
        "payload_len",
    ):
        require_exact_type(frame[field], int, field)
    for field in ("src_ip", "dst_ip", "tcp_flags", "payload_preview"):
        require_exact_type(frame[field], str, field)
    for field in ("is_fragment", "truncated"):
        require_exact_type(frame[field], bool, field)
    payload = base64.b64decode(frame["payload_preview"], validate=True)
    if len(payload) != frame["payload_len"]:
        raise ValueError("packet payload length/base64 invariant failed")


def validate_heartbeat(frame: dict[str, object]) -> None:
    if set(frame) != HEARTBEAT_KEYS or frame.get("type") != "heartbeat":
        raise ValueError("heartbeat shape or frame kind is invalid")
    require_exact_type(frame["session_id"], str, "session_id")
    for field in (
        "seq", "sent", "dropped", "parse_errors", "buf_util_pct",
        "uds_write_errors",
    ):
        require_exact_type(frame[field], int, field)
    metric = frame["avg_json_serialize_us"]
    if type(metric) not in (int, float) or not math.isfinite(metric):
        raise ValueError("avg_json_serialize_us must be finite JSON number")


def validate_hello(frame: dict[str, object]) -> None:
    if set(frame) != HELLO_KEYS or frame.get("type") != "hello":
        raise ValueError("hello shape or frame kind is invalid")
    for field in ("version", "session_id", "hostname"):
        require_exact_type(frame[field], str, field)
    require_exact_type(frame["pid"], int, "pid")
    require_exact_type(frame["max_payload_len"], int, "max_payload_len")
    if frame["max_payload_len"] != 4096:
        raise ValueError("hello max_payload_len changed")


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: validate_formatter_json.py FORMATTER_FUZZ_BINARY", file=sys.stderr)
        return 2
    environment = os.environ.copy()
    environment.setdefault("ASAN_OPTIONS", "detect_leaks=0")
    completed = subprocess.run(
        [sys.argv[1], "--emit-jsonl"],
        check=True,
        capture_output=True,
        text=True,
        env=environment,
    )
    lines = completed.stdout.splitlines()
    if len(lines) != 3 or any(not line for line in lines):
        raise ValueError("formatter must emit exactly three nonempty JSONL frames")
    frames = [
        json.loads(
            line,
            parse_constant=reject_constant,
            object_pairs_hook=unique_object,
        )
        for line in lines
    ]
    if any(type(frame) is not dict for frame in frames):
        raise ValueError("every JSONL frame must decode as one object")
    validate_packet(frames[0])
    validate_heartbeat(frames[1])
    validate_hello(frames[2])
    print("validate_formatter_json: ok frames=3")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
