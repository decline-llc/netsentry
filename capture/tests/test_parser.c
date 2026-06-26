#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>

#include "eth_parser.h"
#include "parser.h"

#define CHECK(cond) do { \
    if (!(cond)) { \
        fprintf(stderr, "FAIL %s:%d: %s\n", __FILE__, __LINE__, #cond); \
        exit(1); \
    } \
} while (0)

static void put_u16(uint8_t *p, uint16_t v) {
    p[0] = (uint8_t)(v >> 8);
    p[1] = (uint8_t)(v & 0xff);
}

static size_t write_eth(uint8_t *buf, uint16_t ether_type) {
    memset(buf, 0, 12);
    put_u16(buf + 12, ether_type);
    return 14;
}

static size_t write_vlan_eth(uint8_t *buf, uint16_t inner_type) {
    memset(buf, 0, 12);
    put_u16(buf + 12, 0x8100);
    put_u16(buf + 14, 100);
    put_u16(buf + 16, inner_type);
    return 18;
}

static size_t write_qinq_eth(uint8_t *buf, uint16_t inner_type) {
    memset(buf, 0, 12);
    put_u16(buf + 12, 0x88A8);
    put_u16(buf + 14, 7);
    put_u16(buf + 16, 0x8100);
    put_u16(buf + 18, 100);
    put_u16(buf + 20, inner_type);
    return 22;
}

static size_t write_ipv4(uint8_t *buf, uint8_t proto, uint16_t payload_len,
                         uint16_t frag_word) {
    memset(buf, 0, 20);
    buf[0] = 0x45;
    put_u16(buf + 2, (uint16_t)(20 + payload_len));
    put_u16(buf + 6, frag_word);
    buf[8] = 64;
    buf[9] = proto;
    buf[12] = 10; buf[13] = 0; buf[14] = 0; buf[15] = 1;
    buf[16] = 10; buf[17] = 0; buf[18] = 0; buf[19] = 2;
    return 20;
}

static size_t write_tcp(uint8_t *buf, uint16_t sport, uint16_t dport,
                        const uint8_t *payload, size_t payload_len) {
    memset(buf, 0, 20);
    put_u16(buf, sport);
    put_u16(buf + 2, dport);
    buf[12] = 0x50;
    buf[13] = 0x18;
    memcpy(buf + 20, payload, payload_len);
    return 20 + payload_len;
}

static size_t write_udp(uint8_t *buf, uint16_t sport, uint16_t dport,
                        const uint8_t *payload, size_t payload_len) {
    memset(buf, 0, 8);
    put_u16(buf, sport);
    put_u16(buf + 2, dport);
    put_u16(buf + 4, (uint16_t)(8 + payload_len));
    memcpy(buf + 8, payload, payload_len);
    return 8 + payload_len;
}

static size_t make_tcp_frame(uint8_t *buf, size_t l2_len,
                             const uint8_t *payload, size_t payload_len) {
    size_t off = l2_len;
    size_t tcp_len = 20 + payload_len;
    off += write_ipv4(buf + off, IPPROTO_TCP, (uint16_t)tcp_len, 0);
    off += write_tcp(buf + off, 12345, 80, payload, payload_len);
    return off;
}

static void test_short_frame_rejected(void) {
    uint8_t frame[8] = {0};
    PacketInfo info;
    CHECK(parse_frame(frame, sizeof(frame), 1, 2, &info) == -1);
}

static void test_tcp_frame_parsed(void) {
    uint8_t frame[256];
    const uint8_t payload[] = "GET / HTTP/1.1\r\n\r\n";
    size_t off = write_eth(frame, 0x0800);
    size_t len = make_tcp_frame(frame, off, payload, sizeof(payload) - 1);

    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)len, 123, 456, &info) == 0);
    CHECK(info.timestamp_sec == 123);
    CHECK(info.timestamp_usec == 456);
    CHECK(strcmp(info.src_ip, "10.0.0.1") == 0);
    CHECK(strcmp(info.dst_ip, "10.0.0.2") == 0);
    CHECK(info.src_port == 12345);
    CHECK(info.dst_port == 80);
    CHECK(info.protocol == IPPROTO_TCP);
    CHECK(strcmp(info.tcp_flags, "ACK|PSH") == 0);
    CHECK(info.payload_len == sizeof(payload) - 1);
    CHECK(memcmp(info.payload, payload, sizeof(payload) - 1) == 0);
    CHECK(!info.is_fragment);
    CHECK(!info.truncated);
}

static void test_udp_frame_parsed(void) {
    uint8_t frame[256];
    const uint8_t payload[] = {0xde, 0xad, 0xbe, 0xef};
    size_t off = write_eth(frame, 0x0800);
    off += write_ipv4(frame + off, IPPROTO_UDP, (uint16_t)(8 + sizeof(payload)), 0);
    off += write_udp(frame + off, 5353, 53, payload, sizeof(payload));

    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)off, 1, 0, &info) == 0);
    CHECK(info.protocol == IPPROTO_UDP);
    CHECK(info.src_port == 5353);
    CHECK(info.dst_port == 53);
    CHECK(info.payload_len == sizeof(payload));
    CHECK(memcmp(info.payload, payload, sizeof(payload)) == 0);
}

static void test_vlan_and_qinq_offsets(void) {
    uint8_t frame[256];
    const uint8_t payload[] = "vlan";
    size_t off = write_vlan_eth(frame, 0x0800);
    size_t len = make_tcp_frame(frame, off, payload, sizeof(payload) - 1);
    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)len, 1, 0, &info) == 0);
    CHECK(info.dst_port == 80);
    CHECK(info.payload_len == sizeof(payload) - 1);

    memset(frame, 0, sizeof(frame));
    off = write_qinq_eth(frame, 0x0800);
    len = make_tcp_frame(frame, off, payload, sizeof(payload) - 1);
    CHECK(parse_frame(frame, (uint32_t)len, 1, 0, &info) == 0);
    CHECK(info.dst_port == 80);
    CHECK(info.payload_len == sizeof(payload) - 1);
}

static void test_fragment_skips_transport(void) {
    uint8_t frame[128];
    size_t off = write_eth(frame, 0x0800);
    off += write_ipv4(frame + off, IPPROTO_TCP, 0, 0x2000);

    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)off, 1, 0, &info) == 0);
    CHECK(info.is_fragment);
    CHECK(info.protocol == IPPROTO_TCP);
    CHECK(info.src_port == 0);
    CHECK(info.dst_port == 0);
    CHECK(info.payload_len == 0);
}

static void test_malformed_tcp_data_offset_rejected(void) {
    uint8_t frame[128];
    size_t off = write_eth(frame, 0x0800);
    off += write_ipv4(frame + off, IPPROTO_TCP, 20, 0);
    uint8_t *tcp = frame + off;
    memset(tcp, 0, 20);
    put_u16(tcp, 12345);
    put_u16(tcp + 2, 80);
    tcp[12] = 0xf0;
    off += 20;

    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)off, 1, 0, &info) == -1);
}

int main(void) {
    parser_registry_register(IPPROTO_TCP, 0, passthrough_parser, "passthrough_tcp");
    parser_registry_register(IPPROTO_UDP, 0, passthrough_parser, "passthrough_udp");

    test_short_frame_rejected();
    test_tcp_frame_parsed();
    test_udp_frame_parsed();
    test_vlan_and_qinq_offsets();
    test_fragment_skips_transport();
    test_malformed_tcp_data_offset_rejected();

    puts("test_parser: ok");
    return 0;
}
