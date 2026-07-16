#!/usr/bin/env python3
"""Regression tests for privacy-preserving PCAP sanitization."""

import struct
import unittest

from scripts.sanitize_pcap import sanitize_frame


class SanitizeFrameTests(unittest.TestCase):
    @staticmethod
    def ipv4_frame(protocol, transport):
        ethernet = b"\x00" * 12 + b"\x08\x00"
        ipv4 = bytearray(20)
        ipv4[0] = 0x45
        ipv4[2:4] = struct.pack("!H", 20 + len(transport))
        ipv4[8] = 64
        ipv4[9] = protocol
        ipv4[12:16] = b"\x0a\x00\x00\x01"
        ipv4[16:20] = b"\x0a\x00\x00\x02"
        return ethernet + bytes(ipv4) + bytes(transport)

    def test_invalid_tcp_data_offset_becomes_opaque_ipv4(self):
        """A zeroed malformed TCP payload must not remain marked as TCP."""
        tcp = bytearray(20)
        tcp[12] = 0xF0

        clean, supported = sanitize_frame(self.ipv4_frame(6, tcp), b"X")

        self.assertTrue(supported)
        self.assertEqual(clean[23], 0)
        self.assertEqual(clean[34:], b"\x00" * 20)

    def test_invalid_udp_length_becomes_opaque_ipv4(self):
        """A zeroed malformed UDP payload must not remain marked as UDP."""
        udp = bytearray(8)
        udp[4:6] = struct.pack("!H", 9)

        clean, supported = sanitize_frame(self.ipv4_frame(17, udp), b"X")

        self.assertTrue(supported)
        self.assertEqual(clean[23], 0)
        self.assertEqual(clean[34:], b"\x00" * 8)


if __name__ == "__main__":
    unittest.main()
