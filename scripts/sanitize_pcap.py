#!/usr/bin/env python3
"""
Sanitize an Ethernet pcap while preserving packet framing for NetSentry tests.

Supported packets keep their Ethernet/VLAN/IPv4/TCP/UDP structure and lengths.
MAC addresses, IPv4 addresses, and transport payload bytes are replaced with
deterministic non-sensitive values. Unsupported captured frames are zeroed so
the output file does not retain raw traffic bytes.
"""

import argparse
import hashlib
import ipaddress
import struct
import sys


PCAP_MAGICS = {
    b"\xd4\xc3\xb2\xa1": ("<", b"\xd4\xc3\xb2\xa1"),
    b"\xa1\xb2\xc3\xd4": (">", b"\xa1\xb2\xc3\xd4"),
    b"\x4d\x3c\xb2\xa1": ("<", b"\x4d\x3c\xb2\xa1"),
    b"\xa1\xb2\x3c\x4d": (">", b"\xa1\xb2\x3c\x4d"),
}

LINKTYPE_ETHERNET = 1
ETH_HDR_LEN = 14
ETHERTYPE_IPV4 = 0x0800
ETHERTYPE_VLAN = {0x8100, 0x88A8}


def checksum_ipv4_header(header):
    if len(header) % 2:
        header += b"\x00"
    total = 0
    for idx in range(0, len(header), 2):
        total += (header[idx] << 8) + header[idx + 1]
        total = (total & 0xFFFF) + (total >> 16)
    return (~total) & 0xFFFF


def mapped_ipv4(raw):
    digest = hashlib.blake2s(raw, digest_size=3, person=b"nspcap").digest()
    second = digest[0] & 0x01
    third = digest[1]
    fourth = (digest[2] % 254) + 1
    return ipaddress.IPv4Address(f"198.{18 + second}.{third}.{fourth}").packed


def sanitize_transport(frame, ip_offset, ihl, total_len, payload_byte):
    proto = frame[ip_offset + 9]
    transport_offset = ip_offset + ihl
    transport_len = total_len - ihl
    if transport_len <= 0:
        return

    if proto == 6 and transport_len >= 20:
        data_offset = ((frame[transport_offset + 12] >> 4) & 0x0F) * 4
        if data_offset < 20 or data_offset > transport_len:
            # The parser validates TCP's data offset.  Once the malformed
            # transport bytes are removed, preserve the safe IPv4 envelope but
            # mark it opaque so replay cannot present zeroed bytes as TCP.
            frame[ip_offset + 9] = 0
            frame[transport_offset:ip_offset + total_len] = b"\x00" * transport_len
            return
        payload_offset = transport_offset + data_offset
        payload_len = transport_len - data_offset
        frame[transport_offset + 16:transport_offset + 18] = b"\x00\x00"
        frame[payload_offset:payload_offset + payload_len] = payload_byte * payload_len
        return

    if proto == 17 and transport_len >= 8:
        udp_len = struct.unpack("!H", frame[transport_offset + 4:transport_offset + 6])[0]
        if udp_len < 8 or udp_len > transport_len:
            frame[ip_offset + 9] = 0
            frame[transport_offset:ip_offset + total_len] = b"\x00" * transport_len
            return
        payload_offset = transport_offset + 8
        payload_len = udp_len - 8
        frame[transport_offset + 6:transport_offset + 8] = b"\x00\x00"
        frame[payload_offset:payload_offset + payload_len] = payload_byte * payload_len
        return

    frame[transport_offset:ip_offset + total_len] = b"\x00" * transport_len


def ipv4_offset(frame):
    if len(frame) < ETH_HDR_LEN:
        return None

    offset = 12
    ether_type = struct.unpack("!H", frame[offset:offset + 2])[0]
    offset += 2
    while ether_type in ETHERTYPE_VLAN:
        if len(frame) < offset + 4:
            return None
        ether_type = struct.unpack("!H", frame[offset + 2:offset + 4])[0]
        offset += 4

    if ether_type != ETHERTYPE_IPV4 or len(frame) < offset + 20:
        return None
    return offset


def sanitize_frame(packet, payload_byte):
    frame = bytearray(packet)
    ip_offset = ipv4_offset(frame)
    if ip_offset is None:
        return b"\x00" * len(frame), False

    frame[0:6] = b"\x02\x00\x00\x00\x00\x02"
    frame[6:12] = b"\x02\x00\x00\x00\x00\x01"

    version_ihl = frame[ip_offset]
    if version_ihl >> 4 != 4:
        return b"\x00" * len(frame), False

    ihl = (version_ihl & 0x0F) * 4
    if ihl < 20 or len(frame) < ip_offset + ihl:
        return b"\x00" * len(frame), False

    total_len = struct.unpack("!H", frame[ip_offset + 2:ip_offset + 4])[0]
    if total_len < ihl or len(frame) < ip_offset + total_len:
        return b"\x00" * len(frame), False

    frame[ip_offset + 12:ip_offset + 16] = mapped_ipv4(bytes(frame[ip_offset + 12:ip_offset + 16]))
    frame[ip_offset + 16:ip_offset + 20] = mapped_ipv4(bytes(frame[ip_offset + 16:ip_offset + 20]))

    if ihl > 20:
        frame[ip_offset + 20:ip_offset + ihl] = b"\x00" * (ihl - 20)

    sanitize_transport(frame, ip_offset, ihl, total_len, payload_byte)

    frame[ip_offset + 10:ip_offset + 12] = b"\x00\x00"
    header_sum = checksum_ipv4_header(bytes(frame[ip_offset:ip_offset + ihl]))
    frame[ip_offset + 10:ip_offset + 12] = struct.pack("!H", header_sum)

    if len(frame) > ip_offset + total_len:
        frame[ip_offset + total_len:] = b"\x00" * (len(frame) - ip_offset - total_len)

    return bytes(frame), True


def sanitize_pcap(src_path, dst_path, payload_byte):
    sanitized = 0
    zeroed = 0

    with open(src_path, "rb") as src, open(dst_path, "wb") as dst:
        header = src.read(24)
        if len(header) != 24:
            raise ValueError("input is too short to be a pcap file")

        magic = header[:4]
        if magic not in PCAP_MAGICS:
            raise ValueError("unsupported pcap magic")

        endian, magic_out = PCAP_MAGICS[magic]
        version_major, version_minor, thiszone, sigfigs, snaplen, network = struct.unpack(
            f"{endian}HHIIII", header[4:]
        )
        if network != LINKTYPE_ETHERNET:
            raise ValueError(f"unsupported pcap linktype {network}; only Ethernet is supported")

        dst.write(magic_out)
        dst.write(struct.pack(f"{endian}HHIIII", version_major, version_minor, thiszone, sigfigs, snaplen, network))

        record_struct = struct.Struct(f"{endian}IIII")
        while True:
            record_header = src.read(record_struct.size)
            if not record_header:
                break
            if len(record_header) != record_struct.size:
                raise ValueError("truncated pcap record header")

            ts_sec, ts_usec, incl_len, orig_len = record_struct.unpack(record_header)
            packet = src.read(incl_len)
            if len(packet) != incl_len:
                raise ValueError("truncated pcap packet data")

            clean_packet, ok = sanitize_frame(packet, payload_byte)
            sanitized += 1 if ok else 0
            zeroed += 0 if ok else 1

            dst.write(record_struct.pack(ts_sec, ts_usec, len(clean_packet), orig_len))
            dst.write(clean_packet)

    return sanitized, zeroed


def parse_args():
    parser = argparse.ArgumentParser(description="Sanitize an Ethernet pcap for NetSentry testing.")
    parser.add_argument("-i", "--input", required=True, help="source pcap path")
    parser.add_argument("-o", "--output", required=True, help="sanitized pcap path")
    parser.add_argument(
        "--payload-byte",
        default="X",
        help="single ASCII byte used to overwrite TCP/UDP payloads (default: X)",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    payload = args.payload_byte.encode("ascii")
    if len(payload) != 1:
        print("[sanitize-pcap] --payload-byte must be one ASCII byte", file=sys.stderr)
        return 2

    try:
        sanitized, zeroed = sanitize_pcap(args.input, args.output, payload)
    except (OSError, ValueError) as exc:
        print(f"[sanitize-pcap] {exc}", file=sys.stderr)
        return 1

    print(
        f"[sanitize-pcap] wrote {args.output} "
        f"(sanitized_ipv4={sanitized} zeroed_unsupported={zeroed})"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
