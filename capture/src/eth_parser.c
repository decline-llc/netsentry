/*
 * eth_parser.c — Ethernet / VLAN / IP / TCP+UDP frame parser.
 *
 * Fills a PacketInfo from a raw libpcap frame.  All field accesses are
 * bounds-checked; malformed packets are counted and skipped.
 */

#include <arpa/inet.h>
#include <netinet/in.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "packet_types.h"
#include "parser.h"

/* ---- Ethernet / IP constants ---------------------------------------- */
#define ETH_HDR_LEN     14
#define ETHERTYPE_IP    0x0800
#define ETHERTYPE_VLAN  0x8100
#define ETHERTYPE_QINQ  0x88A8
#define IP_MF           0x2000   /* More Fragments flag */
#define IP_OFFSET_MASK  0x1FFF

/* ---- internal helpers ----------------------------------------------- */

static inline uint16_t read_u16_be(const uint8_t *p) {
    return (uint16_t)((p[0] << 8) | p[1]);
}

static void append_tcp_flag(char *dst, size_t dst_len, const char *flag) {
    size_t used;

    if (!dst || !flag || dst_len == 0) return;

    used = strnlen(dst, dst_len);
    if (used >= dst_len - 1) return;

    if (used > 0) {
        size_t remaining = dst_len - used;
        int n = snprintf(dst + used, remaining, "|%s", flag);
        (void)n;
    } else {
        snprintf(dst, dst_len, "%s", flag);
    }
}

/* Parse IPv4 header starting at ip_start; fill PacketInfo.
 * Returns 0 on success, -1 on malformed/unsupported packet. */
static int parse_ipv4(const uint8_t *ip_start, uint32_t avail,
                      PacketInfo *info) {
    if (avail < 20) return -1;   /* minimum IPv4 header */

    uint8_t  ihl      = (ip_start[0] & 0x0F) * 4;
    if (ihl < 20 || ihl > avail) return -1;

    uint16_t total_len = read_u16_be(ip_start + 2);
    if (total_len > avail || total_len < ihl) return -1;

    uint16_t frag_off_word = read_u16_be(ip_start + 6);
    bool is_fragment = (frag_off_word & IP_MF) ||
                       (frag_off_word & IP_OFFSET_MASK);

    uint8_t  protocol = ip_start[9];
    struct in_addr src_addr, dst_addr;
    memcpy(&src_addr.s_addr, ip_start + 12, 4);
    memcpy(&dst_addr.s_addr, ip_start + 16, 4);
    inet_ntop(AF_INET, &src_addr, info->src_ip, NS_MAX_IP_STR);
    inet_ntop(AF_INET, &dst_addr, info->dst_ip, NS_MAX_IP_STR);
    info->protocol    = protocol;
    info->is_fragment = is_fragment;

    if (is_fragment) return 0;   /* skip transport layer for fragments */

    const uint8_t *transport = ip_start + ihl;
    uint32_t       transport_len = total_len - ihl;

    if (protocol == IPPROTO_TCP) {
        if (transport_len < 20) return -1;
        info->src_port = read_u16_be(transport);
        info->dst_port = read_u16_be(transport + 2);

        uint8_t data_offset = ((transport[12] >> 4) & 0x0F) * 4;
        if (data_offset < 20 || data_offset > transport_len) return -1;

        uint8_t flags = transport[13];
        if (flags & 0x02) append_tcp_flag(info->tcp_flags, sizeof(info->tcp_flags), "SYN");
        if (flags & 0x10) append_tcp_flag(info->tcp_flags, sizeof(info->tcp_flags), "ACK");
        if (flags & 0x01) append_tcp_flag(info->tcp_flags, sizeof(info->tcp_flags), "FIN");
        if (flags & 0x04) append_tcp_flag(info->tcp_flags, sizeof(info->tcp_flags), "RST");
        if (flags & 0x08) append_tcp_flag(info->tcp_flags, sizeof(info->tcp_flags), "PSH");

        const uint8_t *payload     = transport + data_offset;
        uint32_t       payload_len = transport_len - data_offset;
        parser_registry_run(payload, payload_len, protocol,
                             info->dst_port, info);

    } else if (protocol == IPPROTO_UDP) {
        if (transport_len < 8) return -1;
        info->src_port = read_u16_be(transport);
        info->dst_port = read_u16_be(transport + 2);
        uint16_t udp_len = read_u16_be(transport + 4);
        if (udp_len < 8 || udp_len > transport_len) return -1;

        const uint8_t *payload     = transport + 8;
        uint32_t       payload_len = udp_len - 8;
        parser_registry_run(payload, payload_len, protocol,
                             info->dst_port, info);
    }
    return 0;
}

/* Top-level frame parser.  Returns 0 on success, -1 on error. */
int parse_frame(const uint8_t *pkt, uint32_t pkt_len,
                int64_t ts_sec, int32_t ts_usec,
                PacketInfo *info) {
    memset(info, 0, sizeof(*info));
    info->timestamp_sec  = ts_sec;
    info->timestamp_usec = ts_usec;

    if (pkt_len < (uint32_t)ETH_HDR_LEN) return -1;

    /* Walk VLAN tags */
    uint16_t     ether_type;
    uint32_t     l3_offset = ETH_HDR_LEN - 2;   /* points at EtherType */
    while (1) {
        if (pkt_len < l3_offset + 2) return -1;
        ether_type = read_u16_be(pkt + l3_offset);
        if (ether_type == ETHERTYPE_VLAN || ether_type == ETHERTYPE_QINQ) {
            l3_offset += 4;   /* skip 2-byte TCI + 2-byte inner EtherType */
        } else {
            l3_offset += 2;   /* skip EtherType, now at L3 */
            break;
        }
    }

    if (ether_type != ETHERTYPE_IP) return -1;   /* IPv6 not supported */

    uint32_t avail = pkt_len - l3_offset;
    return parse_ipv4(pkt + l3_offset, avail, info);
}
