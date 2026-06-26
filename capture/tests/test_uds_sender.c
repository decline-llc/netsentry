#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/wait.h>
#include <unistd.h>

#include "packet_types.h"
#include "uds_sender.h"

#define CHECK(cond) do { \
    if (!(cond)) { \
        fprintf(stderr, "FAIL %s:%d: %s\n", __FILE__, __LINE__, #cond); \
        exit(1); \
    } \
} while (0)


static pid_t start_one_line_listener(const char *path) {
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    CHECK(fd >= 0);

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    CHECK(strlen(path) < sizeof(addr.sun_path));
    snprintf(addr.sun_path, sizeof(addr.sun_path), "%s", path);
    unlink(path);
    CHECK(bind(fd, (struct sockaddr *)&addr, sizeof(addr)) == 0);
    CHECK(listen(fd, 1) == 0);

    pid_t pid = fork();
    CHECK(pid >= 0);
    if (pid == 0) {
        int conn = accept(fd, NULL, NULL);
        close(fd);
        if (conn < 0) {
            _exit(2);
        }
        char ch;
        while (read(conn, &ch, 1) == 1) {
            if (ch == '\n') {
                break;
            }
        }
        close(conn);
        _exit(0);
    }

    close(fd);
    return pid;
}

static void wait_listener(pid_t pid) {
    int status = 0;
    CHECK(waitpid(pid, &status, 0) == pid);
    CHECK(WIFEXITED(status) && WEXITSTATUS(status) == 0);
}

static void test_hello_escapes_json_strings(void) {
    char buf[512];
    int n = uds_format_hello_json("abcd1234", "0.1\"x", 123,
                                  "host\n\\name", buf, sizeof(buf));
    CHECK(n > 0);
    CHECK(strstr(buf, "\"type\":\"hello\"") != NULL);
    CHECK(strstr(buf, "\"version\":\"0.1\\\"x\"") != NULL);
    CHECK(strstr(buf, "\"hostname\":\"host\\n\\\\name\"") != NULL);
    CHECK(strchr(buf, '\n') == NULL);
}

static void test_heartbeat_contains_metrics(void) {
    char buf[512];
    HeartbeatInfo hb = {
        .seq = 9,
        .sent = 10,
        .dropped = 2,
        .parse_errors = 3,
        .buf_util_pct = 4,
        .avg_json_serialize_us = 12.5,
        .uds_write_errors = 7,
    };

    int n = uds_format_heartbeat_json(&hb, "abcd1234", buf, sizeof(buf));
    CHECK(n > 0);
    CHECK(strstr(buf, "\"type\":\"heartbeat\"") != NULL);
    CHECK(strstr(buf, "\"seq\":9") != NULL);
    CHECK(strstr(buf, "\"sent\":10") != NULL);
    CHECK(strstr(buf, "\"dropped\":2") != NULL);
    CHECK(strstr(buf, "\"parse_errors\":3") != NULL);
    CHECK(strstr(buf, "\"buf_util_pct\":4") != NULL);
    CHECK(strstr(buf, "\"avg_json_serialize_us\":12.50") != NULL);
    CHECK(strstr(buf, "\"uds_write_errors\":7") != NULL);
}

static void test_packet_base64_and_escapes_flags(void) {
    char buf[1024];
    PacketInfo pkt;
    memset(&pkt, 0, sizeof(pkt));
    pkt.timestamp_sec = 1719300000;
    pkt.timestamp_usec = 123456;
    strcpy(pkt.src_ip, "10.0.0.1");
    strcpy(pkt.dst_ip, "10.0.0.2");
    pkt.src_port = 12345;
    pkt.dst_port = 80;
    pkt.protocol = 6;
    strcpy(pkt.tcp_flags, "A\"\\");
    pkt.payload_len = 3;
    pkt.payload[0] = 'a';
    pkt.payload[1] = 'b';
    pkt.payload[2] = 'c';
    pkt.is_fragment = true;
    pkt.truncated = false;

    int n = uds_format_packet_json(&pkt, buf, sizeof(buf));
    CHECK(n > 0);
    CHECK(strstr(buf, "\"src_ip\":\"10.0.0.1\"") != NULL);
    CHECK(strstr(buf, "\"dst_port\":80") != NULL);
    CHECK(strstr(buf, "\"tcp_flags\":\"A\\\"\\\\\"") != NULL);
    CHECK(strstr(buf, "\"payload_preview\":\"YWJj\"") != NULL);
    CHECK(strstr(buf, "\"is_fragment\":true") != NULL);
    CHECK(strstr(buf, "\"truncated\":false") != NULL);
}

static void test_limited_connect_fails_fast(void) {
    const char *path = "/tmp/netsentry_missing_socket_for_test";
    unlink(path);
    CHECK(uds_connect_with_retries(path, 1) == UDS_ERR_CONN);
}


static void test_reconnect_reuses_last_path(void) {
    char path[108];
    snprintf(path, sizeof(path), "/tmp/netsentry_reconnect_test_%ld.sock", (long)getpid());

    pid_t first = start_one_line_listener(path);
    CHECK(uds_connect_with_retries(path, 5) == UDS_OK);
    CHECK(uds_send_line("{\"first\":true}") == UDS_OK);
    wait_listener(first);

    UDSResult failed = UDS_OK;
    for (int i = 0; i < 16; i++) {
        failed = uds_send_line("{\"after_close\":true}");
        if (failed != UDS_OK) {
            break;
        }
    }
    CHECK(failed != UDS_OK);
    CHECK(uds_write_errors() > 0);

    pid_t second = start_one_line_listener(path);
    CHECK(uds_reconnect() == UDS_OK);
    CHECK(uds_send_line("{\"second\":true}") == UDS_OK);
    uds_close();
    wait_listener(second);
    unlink(path);
}

static void test_format_rejects_truncation(void) {
    char tiny[16];
    HeartbeatInfo hb = {.seq = 1};
    CHECK(uds_format_heartbeat_json(&hb, "abcd1234", tiny, sizeof(tiny)) < 0);
}

int main(void) {
    signal(SIGPIPE, SIG_IGN);
    test_hello_escapes_json_strings();
    test_heartbeat_contains_metrics();
    test_packet_base64_and_escapes_flags();
    test_format_rejects_truncation();
    test_limited_connect_fails_fast();
    test_reconnect_reuses_last_path();
    printf("test_uds_sender: ok\n");
    return 0;
}
