#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>

#include "eth_parser.h"
#include "parser.h"

#define MAX_FUZZ_FRAME 2048
#define MAX_SEEDS 8

typedef struct {
    const char *name;
    size_t len;
    uint8_t data[MAX_FUZZ_FRAME];
} FuzzSeed;

static void init_registry_once(void) {
    static int initialized = 0;
    if (initialized) return;
    parser_registry_register(IPPROTO_TCP, 0, passthrough_parser, "passthrough_tcp");
    parser_registry_register(IPPROTO_UDP, 0, passthrough_parser, "passthrough_udp");
    initialized = 1;
}

static void assert_packet_invariants(const PacketInfo *info) {
    if (info->payload_len > NS_MAX_PAYLOAD_LEN) {
        fprintf(stderr, "payload_len invariant failed: %u\n", info->payload_len);
        abort();
    }
    if (info->src_ip[NS_MAX_IP_STR - 1] != '\0') {
        fprintf(stderr, "src_ip termination invariant failed\n");
        abort();
    }
    if (info->dst_ip[NS_MAX_IP_STR - 1] != '\0') {
        fprintf(stderr, "dst_ip termination invariant failed\n");
        abort();
    }
    if (info->tcp_flags[sizeof(info->tcp_flags) - 1] != '\0') {
        fprintf(stderr, "tcp_flags termination invariant failed\n");
        abort();
    }
}

int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
    init_registry_once();

    if (size > UINT32_MAX) return 0;

    PacketInfo info;
    int rc = parse_frame(data, (uint32_t)size, 1, 0, &info);
    if (rc == 0) {
        assert_packet_invariants(&info);
    }

    return 0;
}

#ifndef NETSENTRY_LIBFUZZER
static uint32_t xorshift32(uint32_t *state) {
    uint32_t x = *state;
    x ^= x << 13;
    x ^= x >> 17;
    x ^= x << 5;
    *state = x;
    return x;
}

static void put_u16(uint8_t *p, uint16_t v) {
    p[0] = (uint8_t)(v >> 8);
    p[1] = (uint8_t)(v & 0xff);
}

static void put_ipv4_addrs(uint8_t *ip) {
    ip[12] = 10; ip[15] = 1;
    ip[16] = 10; ip[19] = 2;
}

static size_t seed_tcp_frame(uint8_t *buf, size_t cap, uint16_t ether_type, size_t l2_extra) {
    const uint8_t payload[] = "GET / HTTP/1.1\r\nHost: fuzz\r\n\r\n";
    size_t payload_len = sizeof(payload) - 1;
    size_t frame_len = 14 + l2_extra + 20 + 20 + payload_len;
    if (cap < frame_len) return 0;

    memset(buf, 0, frame_len);
    put_u16(buf + 12, ether_type);
    if (l2_extra == 4) {
        put_u16(buf + 14, 100);
        put_u16(buf + 16, 0x0800);
    } else if (l2_extra == 8) {
        put_u16(buf + 14, 100);
        put_u16(buf + 16, 0x8100);
        put_u16(buf + 18, 200);
        put_u16(buf + 20, 0x0800);
    }

    uint8_t *ip = buf + 14 + l2_extra;
    ip[0] = 0x45;
    put_u16(ip + 2, (uint16_t)(20 + 20 + payload_len));
    ip[8] = 64;
    ip[9] = IPPROTO_TCP;
    put_ipv4_addrs(ip);

    uint8_t *tcp = ip + 20;
    put_u16(tcp, 12345);
    put_u16(tcp + 2, 80);
    tcp[12] = 0x50;
    tcp[13] = 0x18;
    memcpy(tcp + 20, payload, payload_len);
    return frame_len;
}

static size_t seed_udp_frame(uint8_t *buf, size_t cap) {
    const uint8_t payload[] = "dns-fuzz";
    size_t payload_len = sizeof(payload) - 1;
    size_t frame_len = 14 + 20 + 8 + payload_len;
    if (cap < frame_len) return 0;

    memset(buf, 0, frame_len);
    put_u16(buf + 12, 0x0800);
    uint8_t *ip = buf + 14;
    ip[0] = 0x45;
    put_u16(ip + 2, (uint16_t)(20 + 8 + payload_len));
    ip[8] = 64;
    ip[9] = IPPROTO_UDP;
    put_ipv4_addrs(ip);

    uint8_t *udp = ip + 20;
    put_u16(udp, 5353);
    put_u16(udp + 2, 53);
    put_u16(udp + 4, (uint16_t)(8 + payload_len));
    memcpy(udp + 8, payload, payload_len);
    return frame_len;
}

static size_t seed_fragment_frame(uint8_t *buf, size_t cap) {
    size_t len = seed_udp_frame(buf, cap);
    if (len == 0) return 0;
    uint8_t *ip = buf + 14;
    put_u16(ip + 6, 0x2001);
    return len;
}

static size_t seed_bad_tcp_offset_frame(uint8_t *buf, size_t cap) {
    size_t len = seed_tcp_frame(buf, cap, 0x0800, 0);
    if (len == 0) return 0;
    uint8_t *tcp = buf + 14 + 20;
    tcp[12] = 0xf0;
    return len;
}

static size_t build_seeds(FuzzSeed *seeds, size_t cap) {
    if (cap < MAX_SEEDS) return 0;

    seeds[0].name = "tcp";
    seeds[0].len = seed_tcp_frame(seeds[0].data, sizeof(seeds[0].data), 0x0800, 0);
    seeds[1].name = "udp";
    seeds[1].len = seed_udp_frame(seeds[1].data, sizeof(seeds[1].data));
    seeds[2].name = "vlan-tcp";
    seeds[2].len = seed_tcp_frame(seeds[2].data, sizeof(seeds[2].data), 0x8100, 4);
    seeds[3].name = "qinq-tcp";
    seeds[3].len = seed_tcp_frame(seeds[3].data, sizeof(seeds[3].data), 0x88a8, 8);
    seeds[4].name = "fragment";
    seeds[4].len = seed_fragment_frame(seeds[4].data, sizeof(seeds[4].data));
    seeds[5].name = "bad-tcp-offset";
    seeds[5].len = seed_bad_tcp_offset_frame(seeds[5].data, sizeof(seeds[5].data));
    seeds[6].name = "short-ethernet";
    seeds[6].len = 13;
    memset(seeds[6].data, 0xa5, seeds[6].len);
    seeds[7].name = "empty";
    seeds[7].len = 0;

    return MAX_SEEDS;
}

static void run_mutation_rounds(uint32_t iterations) {
    FuzzSeed seeds[MAX_SEEDS];
    uint8_t frame[MAX_FUZZ_FRAME];
    size_t seed_count = build_seeds(seeds, MAX_SEEDS);
    uint32_t rng = 0xC0FFEEu;

    if (seed_count == 0) abort();
    for (size_t i = 0; i < seed_count; i++) {
        LLVMFuzzerTestOneInput(seeds[i].data, seeds[i].len);
    }

    for (uint32_t i = 0; i < iterations; i++) {
        FuzzSeed *seed = &seeds[xorshift32(&rng) % seed_count];
        size_t len = seed->len;
        memset(frame, 0, sizeof(frame));

        if ((i % 4) == 0 && len > 0) {
            memcpy(frame, seed->data, len);
        } else if ((i % 4) == 1) {
            len = xorshift32(&rng) % sizeof(frame);
        } else {
            size_t copy_len = len;
            if (copy_len > sizeof(frame)) copy_len = sizeof(frame);
            memcpy(frame, seed->data, copy_len);
            len = xorshift32(&rng) % (copy_len + 1);
        }

        uint32_t mutations = 1 + (xorshift32(&rng) % 32);
        for (uint32_t j = 0; j < mutations; j++) {
            size_t idx = xorshift32(&rng) % sizeof(frame);
            frame[idx] = (uint8_t)(xorshift32(&rng) & 0xff);
        }

        LLVMFuzzerTestOneInput(frame, len);
    }
}

static void run_file(const char *path) {
    FILE *fp = fopen(path, "rb");
    if (!fp) {
        perror(path);
        exit(1);
    }

    uint8_t buf[MAX_FUZZ_FRAME];
    size_t len = fread(buf, 1, sizeof(buf), fp);
    if (ferror(fp)) {
        perror(path);
        fclose(fp);
        exit(1);
    }
    fclose(fp);

    LLVMFuzzerTestOneInput(buf, len);
}

int main(int argc, char **argv) {
    uint32_t iterations = 5000;
    const char *env = getenv("FUZZ_ITERATIONS");
    if (env && env[0] != '\0') {
        char *end = NULL;
        unsigned long parsed = strtoul(env, &end, 10);
        if (end && *end == '\0' && parsed > 0 && parsed <= UINT32_MAX) {
            iterations = (uint32_t)parsed;
        }
    }

    if (argc > 1) {
        for (int i = 1; i < argc; i++) {
            run_file(argv[i]);
        }
        printf("fuzz_parser: ok files=%d\n", argc - 1);
        return 0;
    }

    run_mutation_rounds(iterations);
    printf("fuzz_parser: ok iterations=%u\n", iterations);
    return 0;
}
#endif
