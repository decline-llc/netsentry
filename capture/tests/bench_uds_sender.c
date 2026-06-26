#include <errno.h>
#include <inttypes.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

#include "packet_types.h"
#include "uds_sender.h"

#define ITERATIONS 100000u
#define CHECK(cond) do { \
    if (!(cond)) { \
        fprintf(stderr, "FAIL %s:%d: %s\n", __FILE__, __LINE__, #cond); \
        exit(1); \
    } \
} while (0)

static volatile uint64_t sink = 0;

static uint64_t monotonic_ns(void) {
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) {
        perror("clock_gettime");
        exit(1);
    }
    return (uint64_t)ts.tv_sec * 1000000000ull + (uint64_t)ts.tv_nsec;
}

static uint32_t parse_iterations(int argc, char **argv) {
    uint32_t iterations = ITERATIONS;
    if (argc > 1) {
        char *end = NULL;
        unsigned long v = strtoul(argv[1], &end, 10);
        if (end && *end == '\0' && v > 0 && v <= UINT32_MAX) {
            iterations = (uint32_t)v;
        }
    }
    return iterations;
}

static PacketInfo sample_packet(void) {
    PacketInfo pkt;
    memset(&pkt, 0, sizeof(pkt));
    pkt.timestamp_sec = 1719300000;
    pkt.timestamp_usec = 123456;
    strcpy(pkt.src_ip, "10.0.0.3");
    strcpy(pkt.dst_ip, "10.0.0.2");
    pkt.src_port = 54322;
    pkt.dst_port = 80;
    pkt.protocol = 6;
    strcpy(pkt.tcp_flags, "ACK|PSH");
    const char payload[] =
        "GET /search?q=1'+union+select+1,2,3-- HTTP/1.1\r\n"
        "Host: example.com\r\nUser-Agent: sqlmap/1.7\r\n\r\n";
    pkt.payload_len = (uint16_t)(sizeof(payload) - 1);
    memcpy(pkt.payload, payload, pkt.payload_len);
    pkt.is_fragment = false;
    pkt.truncated = false;
    return pkt;
}

static void print_result(const char *name, uint32_t iterations, uint64_t elapsed) {
    double ns_per_op = (double)elapsed / (double)iterations;
    double ops_per_sec = 1000000000.0 / ns_per_op;
    printf("bench_uds_sender/%s iterations=%u ns_per_op=%.2f ops_per_sec=%.0f\n",
           name, iterations, ns_per_op, ops_per_sec);
}

static void bench_packet_json(uint32_t iterations) {
    PacketInfo pkt = sample_packet();
    char buf[NS_MAX_PAYLOAD_LEN * 2 + 512];
    CHECK(uds_format_packet_json(&pkt, buf, sizeof(buf)) > 0);

    uint64_t start = monotonic_ns();
    for (uint32_t i = 0; i < iterations; i++) {
        int n = uds_format_packet_json(&pkt, buf, sizeof(buf));
        CHECK(n > 0);
        sink += (uint64_t)n;
    }
    print_result("format_packet_json", iterations, monotonic_ns() - start);
}

static void bench_heartbeat_json(uint32_t iterations) {
    HeartbeatInfo hb = {
        .seq = 42,
        .sent = 1000,
        .dropped = 2,
        .parse_errors = 3,
        .buf_util_pct = 0,
        .avg_json_serialize_us = 1.25,
        .uds_write_errors = 4,
    };
    char buf[512];
    CHECK(uds_format_heartbeat_json(&hb, "abcd1234", buf, sizeof(buf)) > 0);

    uint64_t start = monotonic_ns();
    for (uint32_t i = 0; i < iterations; i++) {
        hb.seq = i;
        int n = uds_format_heartbeat_json(&hb, "abcd1234", buf, sizeof(buf));
        CHECK(n > 0);
        sink += (uint64_t)n;
    }
    print_result("format_heartbeat_json", iterations, monotonic_ns() - start);
}

static int start_drain_listener(const char *path, pid_t *child_pid) {
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) {
        perror("socket");
        return -1;
    }

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    if (strlen(path) >= sizeof(addr.sun_path)) {
        close(fd);
        errno = ENAMETOOLONG;
        return -1;
    }
    snprintf(addr.sun_path, sizeof(addr.sun_path), "%s", path);
    unlink(path);

    if (bind(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("bind");
        close(fd);
        return -1;
    }
    if (listen(fd, 1) < 0) {
        perror("listen");
        close(fd);
        return -1;
    }

    pid_t pid = fork();
    if (pid < 0) {
        perror("fork");
        close(fd);
        return -1;
    }
    if (pid == 0) {
        int conn = accept(fd, NULL, NULL);
        close(fd);
        if (conn < 0) {
            _exit(2);
        }
        char buf[8192];
        while (read(conn, buf, sizeof(buf)) > 0) {
        }
        close(conn);
        _exit(0);
    }

    close(fd);
    *child_pid = pid;
    return 0;
}

static void bench_uds_send_line(uint32_t iterations) {
    char path[108];
    snprintf(path, sizeof(path), "/tmp/netsentry_bench_uds_%ld.sock", (long)getpid());

    pid_t child = -1;
    CHECK(start_drain_listener(path, &child) == 0);
    CHECK(uds_connect_with_retries(path, 20) == UDS_OK);

    PacketInfo pkt = sample_packet();
    char line[NS_MAX_PAYLOAD_LEN * 2 + 512];
    CHECK(uds_format_packet_json(&pkt, line, sizeof(line)) > 0);

    uint64_t start = monotonic_ns();
    for (uint32_t i = 0; i < iterations; i++) {
        CHECK(uds_send_line(line) == UDS_OK);
        sink += line[0];
    }
    uint64_t elapsed = monotonic_ns() - start;

    uds_close();
    int status = 0;
    CHECK(waitpid(child, &status, 0) == child);
    CHECK(WIFEXITED(status) && WEXITSTATUS(status) == 0);
    unlink(path);

    print_result("uds_send_line", iterations, elapsed);
}

int main(int argc, char **argv) {
    signal(SIGPIPE, SIG_IGN);
    uint32_t iterations = parse_iterations(argc, argv);
    bench_packet_json(iterations);
    bench_heartbeat_json(iterations);
    bench_uds_send_line(iterations);
    printf("bench_uds_sender/sink=%" PRIu64 " avg_json_serialize_us=%.2f write_errors=%" PRIu64 "\n",
           sink, uds_avg_json_serialize_us(), uds_write_errors());
    return 0;
}
