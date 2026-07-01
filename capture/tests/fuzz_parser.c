#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>

#include "eth_parser.h"
#include "parser.h"

#define MAX_FUZZ_FRAME 2048

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

static size_t seed_tcp_frame(uint8_t *buf, size_t cap) {
    const uint8_t payload[] = "GET / HTTP/1.1\r\n\r\n";
    size_t payload_len = sizeof(payload) - 1;
    size_t frame_len = 14 + 20 + 20 + payload_len;
    if (cap < frame_len) return 0;

    memset(buf, 0, frame_len);
    put_u16(buf + 12, 0x0800);
    buf[14] = 0x45;
    put_u16(buf + 16, (uint16_t)(20 + 20 + payload_len));
    buf[22] = 64;
    buf[23] = IPPROTO_TCP;
    buf[26] = 10; buf[29] = 1;
    buf[30] = 10; buf[33] = 2;
    put_u16(buf + 34, 12345);
    put_u16(buf + 36, 80);
    buf[46] = 0x50;
    buf[47] = 0x18;
    memcpy(buf + 54, payload, payload_len);
    return frame_len;
}

static void run_mutation_rounds(uint32_t iterations) {
    uint8_t seed[MAX_FUZZ_FRAME];
    uint8_t frame[MAX_FUZZ_FRAME];
    size_t seed_len = seed_tcp_frame(seed, sizeof(seed));
    uint32_t rng = 0xC0FFEEu;

    if (seed_len == 0) abort();
    LLVMFuzzerTestOneInput(seed, seed_len);

    for (uint32_t i = 0; i < iterations; i++) {
        size_t len = xorshift32(&rng) % sizeof(frame);
        if ((i % 3) == 0) {
            memcpy(frame, seed, seed_len);
            len = seed_len;
        } else {
            memset(frame, 0, sizeof(frame));
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
