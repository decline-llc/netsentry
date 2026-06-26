#include <inttypes.h>
#include <netinet/in.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#include "eth_parser.h"
#include "parser.h"

#define ITERATIONS 1000000u
#define CHECK(cond) do { \
    if (!(cond)) { \
        fprintf(stderr, "FAIL %s:%d: %s\n", __FILE__, __LINE__, #cond); \
        exit(1); \
    } \
} while (0)

static volatile uint64_t sink = 0;

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

static size_t write_ipv4(uint8_t *buf, uint8_t proto, uint16_t payload_len) {
    memset(buf, 0, 20);
    buf[0] = 0x45;
    put_u16(buf + 2, (uint16_t)(20 + payload_len));
    buf[8] = 64;
    buf[9] = proto;
    buf[12] = 10; buf[13] = 0; buf[14] = 0; buf[15] = 1;
    buf[16] = 10; buf[17] = 0; buf[18] = 0; buf[19] = 2;
    return 20;
}

static size_t write_tcp(uint8_t *buf, const uint8_t *payload, size_t payload_len) {
    memset(buf, 0, 20);
    put_u16(buf, 12345);
    put_u16(buf + 2, 80);
    buf[12] = 0x50;
    buf[13] = 0x18;
    memcpy(buf + 20, payload, payload_len);
    return 20 + payload_len;
}

static size_t make_tcp_frame(uint8_t *buf, size_t l2_len,
                             const uint8_t *payload, size_t payload_len) {
    size_t off = l2_len;
    size_t tcp_len = 20 + payload_len;
    off += write_ipv4(buf + off, IPPROTO_TCP, (uint16_t)tcp_len);
    off += write_tcp(buf + off, payload, payload_len);
    return off;
}

static uint64_t monotonic_ns(void) {
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) {
        perror("clock_gettime");
        exit(1);
    }
    return (uint64_t)ts.tv_sec * 1000000000ull + (uint64_t)ts.tv_nsec;
}

static void bench_case(const char *name, const uint8_t *frame, size_t len,
                       uint32_t iterations) {
    PacketInfo info;
    CHECK(parse_frame(frame, (uint32_t)len, 1, 0, &info) == 0);

    uint64_t start = monotonic_ns();
    for (uint32_t i = 0; i < iterations; i++) {
        CHECK(parse_frame(frame, (uint32_t)len, 1, (int32_t)i, &info) == 0);
        sink += info.payload_len + info.dst_port + info.protocol;
    }
    uint64_t elapsed = monotonic_ns() - start;
    double ns_per_packet = (double)elapsed / (double)iterations;
    double pps = 1000000000.0 / ns_per_packet;
    printf("bench_parser/%s iterations=%u ns_per_packet=%.2f pps=%.0f\n",
           name, iterations, ns_per_packet, pps);
}

int main(int argc, char **argv) {
    uint32_t iterations = ITERATIONS;
    if (argc > 1) {
        char *end = NULL;
        unsigned long v = strtoul(argv[1], &end, 10);
        if (end && *end == '\0' && v > 0 && v <= UINT32_MAX) {
            iterations = (uint32_t)v;
        }
    }

    parser_registry_register(IPPROTO_TCP, 0, passthrough_parser, "passthrough_tcp");

    const uint8_t payload[] =
        "GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n"
        "Host: example.com\r\nUser-Agent: sqlmap/1.7\r\n\r\n";

    uint8_t plain[512];
    uint8_t vlan[512];
    uint8_t qinq[512];

    size_t plain_len = make_tcp_frame(plain, write_eth(plain, 0x0800),
                                      payload, sizeof(payload) - 1);
    size_t vlan_len = make_tcp_frame(vlan, write_vlan_eth(vlan, 0x0800),
                                     payload, sizeof(payload) - 1);
    size_t qinq_len = make_tcp_frame(qinq, write_qinq_eth(qinq, 0x0800),
                                     payload, sizeof(payload) - 1);

    bench_case("tcp_plain", plain, plain_len, iterations);
    bench_case("tcp_vlan", vlan, vlan_len, iterations);
    bench_case("tcp_qinq", qinq, qinq_len, iterations);
    printf("bench_parser/sink=%" PRIu64 "\n", sink);
    return 0;
}
