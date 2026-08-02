#include <limits.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "packet_types.h"
#include "uds_sender.h"

#define FORMAT_BUFFER_CAP (NS_MAX_PAYLOAD_LEN * 2u + 2048u)
#define GUARD_BYTES 32u
#define MAX_FUZZ_INPUT 16384u

typedef int (*FormatCallback)(const void *, char *, size_t);

typedef struct {
    const uint8_t *data;
    size_t size;
    size_t offset;
} InputReader;

typedef struct {
    PacketInfo packet;
} PacketContext;

typedef struct {
    HeartbeatInfo heartbeat;
    char session_id[NS_SESSION_ID_LEN];
} HeartbeatContext;

typedef struct {
    char session_id[NS_SESSION_ID_LEN];
    char version[11];
    int pid;
    char hostname[64];
} HelloContext;

static void fail(const char *message) {
    fprintf(stderr, "fuzz_uds_formatter: %s\n", message);
    abort();
}

static int format_packet(const void *context, char *buf, size_t buf_len) {
    const PacketContext *packet = context;
    return uds_format_packet_json(&packet->packet, buf, buf_len);
}

static int format_heartbeat(const void *context, char *buf, size_t buf_len) {
    const HeartbeatContext *heartbeat = context;
    return uds_format_heartbeat_json(&heartbeat->heartbeat,
                                     heartbeat->session_id, buf, buf_len);
}

static int format_hello(const void *context, char *buf, size_t buf_len) {
    const HelloContext *hello = context;
    return uds_format_hello_json(hello->session_id, hello->version, hello->pid,
                                 hello->hostname, buf, buf_len);
}

static int invoke_guarded(FormatCallback formatter, const void *context,
                          size_t buf_len, char *snapshot,
                          size_t snapshot_len) {
    uint8_t storage[FORMAT_BUFFER_CAP + GUARD_BYTES * 2u];
    memset(storage, 0xa5, sizeof(storage));
    char *output = (char *)(storage + GUARD_BYTES);

    if (buf_len > FORMAT_BUFFER_CAP) fail("invalid guarded buffer length");
    int result = formatter(context, output, buf_len);

    for (size_t i = 0; i < GUARD_BYTES; i++) {
        if (storage[i] != 0xa5) fail("write before output buffer");
    }
    for (size_t i = GUARD_BYTES + buf_len; i < sizeof(storage); i++) {
        if (storage[i] != 0xa5) fail("write after output buffer");
    }

    if (result >= 0) {
        if (buf_len == 0 || (size_t)result >= buf_len) {
            fail("formatter exposed an out-of-range success");
        }
        if (output[result] != '\0' || strlen(output) != (size_t)result) {
            fail("successful output length or termination mismatch");
        }
        if (snapshot) {
            if ((size_t)result + 1u > snapshot_len) {
                fail("snapshot buffer too small");
            }
            memcpy(snapshot, output, (size_t)result + 1u);
        }
    }
    return result;
}

static void assert_frame_kind(const char *output, const char *kind) {
    if (output[0] != '{' || output[strlen(output) - 1u] != '}') {
        fail("successful output is not one JSON object");
    }
    if (strchr(output, '\n') || strchr(output, '\r')) {
        fail("successful output contains a raw line delimiter");
    }
    if (strcmp(kind, "packet") == 0) {
        if (strncmp(output, "{\"timestamp_sec\":", 17) != 0 ||
            strstr(output, "\"type\":") != NULL) {
            fail("packet frame-kind invariant failed");
        }
    } else {
        char prefix[48];
        int n = snprintf(prefix, sizeof(prefix), "{\"type\":\"%s\"", kind);
        if (n < 0 || (size_t)n >= sizeof(prefix) ||
            strncmp(output, prefix, (size_t)n) != 0) {
            fail("typed frame-kind invariant failed");
        }
    }
}

static void exercise_formatter(FormatCallback formatter, const void *context,
                               const char *kind) {
    char expected[FORMAT_BUFFER_CAP];
    char exact[FORMAT_BUFFER_CAP];
    int result = invoke_guarded(formatter, context, FORMAT_BUFFER_CAP,
                                expected, sizeof(expected));
    if (result <= 0) fail("bounded structured input did not format");
    assert_frame_kind(expected, kind);

    size_t exact_len = (size_t)result + 1u;
    int exact_result = invoke_guarded(formatter, context, exact_len,
                                      exact, sizeof(exact));
    if (exact_result != result || strcmp(exact, expected) != 0) {
        fail("exact-fit output changed or failed");
    }
    if (invoke_guarded(formatter, context, (size_t)result, NULL, 0) >= 0) {
        fail("one-byte-short output succeeded");
    }
    if (invoke_guarded(formatter, context, 1u, NULL, 0) >= 0) {
        fail("single-byte output succeeded");
    }
    if (invoke_guarded(formatter, context, 0u, NULL, 0) >= 0) {
        fail("zero-byte output succeeded");
    }
}

static uint8_t read_byte(InputReader *reader) {
    uint8_t value;
    if (reader->size == 0) {
        value = (uint8_t)(0x5au + reader->offset * 37u);
    } else {
        value = reader->data[reader->offset % reader->size];
    }
    reader->offset++;
    return value;
}

static uint64_t read_u64(InputReader *reader) {
    uint64_t value = 0;
    for (unsigned int i = 0; i < 8; i++) {
        value |= (uint64_t)read_byte(reader) << (i * 8u);
    }
    return value;
}

static void read_text(InputReader *reader, char *output, size_t capacity) {
    static const unsigned char alphabet[] = {
        'a', 'Z', '0', '.', '-', '_', '/', ':', ' ', '"', '\\',
        '\b', '\f', '\n', '\r', '\t', 0x01, 0x1f
    };
    if (capacity == 0) fail("zero-capacity text destination");
    size_t length = (size_t)read_byte(reader) % capacity;
    for (size_t i = 0; i < length; i++) {
        output[i] = (char)alphabet[read_byte(reader) % sizeof(alphabet)];
    }
    output[length] = '\0';
}

static void derive_packet(InputReader *reader, PacketContext *context) {
    PacketInfo *packet = &context->packet;
    memset(packet, 0, sizeof(*packet));
    packet->timestamp_sec = (int64_t)read_u64(reader);
    packet->timestamp_usec = (int32_t)read_u64(reader);
    read_text(reader, packet->src_ip, sizeof(packet->src_ip));
    read_text(reader, packet->dst_ip, sizeof(packet->dst_ip));
    packet->src_port = (uint16_t)read_u64(reader);
    packet->dst_port = (uint16_t)read_u64(reader);
    packet->protocol = read_byte(reader);
    read_text(reader, packet->tcp_flags, sizeof(packet->tcp_flags));

    uint8_t payload_choice = read_byte(reader);
    if (payload_choice == 0) {
        packet->payload_len = 0;
    } else if (payload_choice == 1) {
        packet->payload_len = 1;
    } else if (payload_choice == 2) {
        packet->payload_len = NS_MAX_PAYLOAD_LEN;
    } else {
        packet->payload_len =
            (uint16_t)(read_u64(reader) % (NS_MAX_PAYLOAD_LEN + 1u));
    }
    for (uint16_t i = 0; i < packet->payload_len; i++) {
        packet->payload[i] = read_byte(reader);
    }
    packet->is_fragment = (read_byte(reader) & 1u) != 0;
    packet->truncated = (read_byte(reader) & 1u) != 0;
}

static void derive_heartbeat(InputReader *reader, HeartbeatContext *context) {
    memset(context, 0, sizeof(*context));
    read_text(reader, context->session_id, sizeof(context->session_id));
    context->heartbeat.seq = (uint32_t)read_u64(reader);
    context->heartbeat.sent = read_u64(reader);
    context->heartbeat.dropped = read_u64(reader);
    context->heartbeat.parse_errors = read_u64(reader);
    context->heartbeat.buf_util_pct = (uint32_t)read_u64(reader);
    int32_t metric = (int32_t)read_u64(reader);
    context->heartbeat.avg_json_serialize_us = (double)metric / 100.0;
    context->heartbeat.uds_write_errors = read_u64(reader);
}

static void derive_hello(InputReader *reader, HelloContext *context) {
    memset(context, 0, sizeof(*context));
    read_text(reader, context->session_id, sizeof(context->session_id));
    read_text(reader, context->version, sizeof(context->version));
    context->pid = (int)(uint32_t)read_u64(reader);
    read_text(reader, context->hostname, sizeof(context->hostname));
}

int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
    InputReader reader = {.data = data, .size = size, .offset = 0};
    switch (read_byte(&reader) % 3u) {
    case 0: {
        PacketContext context;
        derive_packet(&reader, &context);
        exercise_formatter(format_packet, &context, "packet");
        break;
    }
    case 1: {
        HeartbeatContext context;
        derive_heartbeat(&reader, &context);
        exercise_formatter(format_heartbeat, &context, "heartbeat");
        break;
    }
    default: {
        HelloContext context;
        derive_hello(&reader, &context);
        exercise_formatter(format_hello, &context, "hello");
        break;
    }
    }
    return 0;
}

#ifndef NETSENTRY_LIBFUZZER
static void fill_pattern(char *output, size_t capacity,
                         const unsigned char *pattern, size_t pattern_len) {
    if (capacity == 0 || pattern_len == 0) fail("invalid text pattern");
    for (size_t i = 0; i + 1u < capacity; i++) {
        output[i] = (char)pattern[i % pattern_len];
    }
    output[capacity - 1u] = '\0';
}

static void run_structured_seeds(void) {
    static const unsigned char escape_pattern[] = {
        '"', '\\', '\b', '\f', '\n', '\r', '\t', 0x01, 0x1f
    };
    PacketContext packet = {0};
    packet.packet.timestamp_sec = INT64_MIN;
    packet.packet.timestamp_usec = INT32_MAX;
    fill_pattern(packet.packet.src_ip, sizeof(packet.packet.src_ip),
                 escape_pattern, sizeof(escape_pattern));
    fill_pattern(packet.packet.dst_ip, sizeof(packet.packet.dst_ip),
                 escape_pattern, sizeof(escape_pattern));
    fill_pattern(packet.packet.tcp_flags, sizeof(packet.packet.tcp_flags),
                 escape_pattern, sizeof(escape_pattern));
    packet.packet.src_port = UINT16_MAX;
    packet.packet.dst_port = UINT16_MAX;
    packet.packet.protocol = UINT8_MAX;
    packet.packet.payload_len = NS_MAX_PAYLOAD_LEN;
    for (uint16_t i = 0; i < packet.packet.payload_len; i++) {
        packet.packet.payload[i] = (uint8_t)i;
    }
    packet.packet.is_fragment = true;
    packet.packet.truncated = true;
    exercise_formatter(format_packet, &packet, "packet");
    memset(packet.packet.payload, 0, sizeof(packet.packet.payload));
    packet.packet.payload_len = 0;
    packet.packet.timestamp_sec = INT64_MAX;
    packet.packet.timestamp_usec = INT32_MIN;
    exercise_formatter(format_packet, &packet, "packet");

    HeartbeatContext heartbeat = {0};
    fill_pattern(heartbeat.session_id, sizeof(heartbeat.session_id),
                 escape_pattern, sizeof(escape_pattern));
    heartbeat.heartbeat.seq = UINT32_MAX;
    heartbeat.heartbeat.sent = UINT64_MAX;
    heartbeat.heartbeat.dropped = UINT64_MAX;
    heartbeat.heartbeat.parse_errors = UINT64_MAX;
    heartbeat.heartbeat.buf_util_pct = UINT32_MAX;
    heartbeat.heartbeat.avg_json_serialize_us = -21474836.48;
    heartbeat.heartbeat.uds_write_errors = UINT64_MAX;
    exercise_formatter(format_heartbeat, &heartbeat, "heartbeat");

    HelloContext hello = {0};
    fill_pattern(hello.session_id, sizeof(hello.session_id),
                 escape_pattern, sizeof(escape_pattern));
    fill_pattern(hello.version, sizeof(hello.version),
                 escape_pattern, sizeof(escape_pattern));
    fill_pattern(hello.hostname, sizeof(hello.hostname),
                 escape_pattern, sizeof(escape_pattern));
    hello.pid = INT_MIN;
    exercise_formatter(format_hello, &hello, "hello");
    hello.pid = INT_MAX;
    exercise_formatter(format_hello, &hello, "hello");
}

static uint32_t xorshift32(uint32_t *state) {
    uint32_t value = *state;
    value ^= value << 13;
    value ^= value >> 17;
    value ^= value << 5;
    *state = value;
    return value;
}

static void run_mutation_rounds(uint32_t iterations) {
    static const uint8_t seeds[][32] = {
        {0},
        {0, 2, 0xff, 0x00, '"', '\\', '\n'},
        {1, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
        {2, 0x00, 0x01, 0x7f, 0x80, 0xfe, 0xff},
    };
    uint8_t input[sizeof(seeds[0])];
    uint32_t rng = 0x46524d54u;

    run_structured_seeds();
    for (size_t i = 0; i < sizeof(seeds) / sizeof(seeds[0]); i++) {
        LLVMFuzzerTestOneInput(seeds[i], sizeof(seeds[i]));
    }
    for (uint32_t i = 0; i < iterations; i++) {
        size_t seed_index = xorshift32(&rng) %
                            (sizeof(seeds) / sizeof(seeds[0]));
        memcpy(input, seeds[seed_index], sizeof(input));
        uint32_t mutations = 1u + xorshift32(&rng) % 16u;
        for (uint32_t j = 0; j < mutations; j++) {
            input[xorshift32(&rng) % sizeof(input)] =
                (uint8_t)xorshift32(&rng);
        }
        size_t length = xorshift32(&rng) % (sizeof(input) + 1u);
        LLVMFuzzerTestOneInput(input, length);
    }
}

static void emit_jsonl(void) {
    char output[FORMAT_BUFFER_CAP];
    PacketContext packet = {0};
    packet.packet.timestamp_sec = 1719300000;
    packet.packet.timestamp_usec = 123456;
    strcpy(packet.packet.src_ip, "10.0.0.1");
    strcpy(packet.packet.dst_ip, "10.0.0.2");
    packet.packet.src_port = 12345;
    packet.packet.dst_port = 443;
    packet.packet.protocol = 6;
    strcpy(packet.packet.tcp_flags, "ACK\"\\");
    packet.packet.payload_len = 3;
    memcpy(packet.packet.payload, "abc", 3);
    int result = format_packet(&packet, output, sizeof(output));
    if (result <= 0) fail("could not emit packet seed");
    printf("%s\n", output);

    HeartbeatContext heartbeat = {0};
    strcpy(heartbeat.session_id, "abcd1234");
    heartbeat.heartbeat.seq = 9;
    heartbeat.heartbeat.sent = 10;
    heartbeat.heartbeat.dropped = 2;
    heartbeat.heartbeat.parse_errors = 3;
    heartbeat.heartbeat.buf_util_pct = 4;
    heartbeat.heartbeat.avg_json_serialize_us = 12.5;
    heartbeat.heartbeat.uds_write_errors = 7;
    result = format_heartbeat(&heartbeat, output, sizeof(output));
    if (result <= 0) fail("could not emit heartbeat seed");
    printf("%s\n", output);

    HelloContext hello = {0};
    strcpy(hello.session_id, "abcd1234");
    strcpy(hello.version, "0.1\"x");
    hello.pid = 123;
    strcpy(hello.hostname, "host\n\\name");
    result = format_hello(&hello, output, sizeof(output));
    if (result <= 0) fail("could not emit hello seed");
    printf("%s\n", output);
}

static void run_file(const char *path) {
    FILE *file = fopen(path, "rb");
    if (!file) {
        perror(path);
        exit(1);
    }
    uint8_t input[MAX_FUZZ_INPUT];
    size_t length = fread(input, 1, sizeof(input), file);
    if (ferror(file)) {
        perror(path);
        fclose(file);
        exit(1);
    }
    fclose(file);
    LLVMFuzzerTestOneInput(input, length);
}

int main(int argc, char **argv) {
    if (argc == 2 && strcmp(argv[1], "--emit-jsonl") == 0) {
        emit_jsonl();
        return 0;
    }
    if (argc > 1) {
        for (int i = 1; i < argc; i++) run_file(argv[i]);
        printf("fuzz_uds_formatter: ok files=%d\n", argc - 1);
        return 0;
    }

    uint32_t iterations = 5000;
    const char *environment = getenv("FUZZ_FORMATTER_ITERATIONS");
    if (environment && environment[0] != '\0') {
        char *end = NULL;
        unsigned long parsed = strtoul(environment, &end, 10);
        if (end && *end == '\0' && parsed > 0 && parsed <= UINT32_MAX) {
            iterations = (uint32_t)parsed;
        }
    }
    run_mutation_rounds(iterations);
    printf("fuzz_uds_formatter: ok iterations=%u\n", iterations);
    return 0;
}
#endif
