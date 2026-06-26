/*
 * uds_sender.c — Unix Domain Socket sender with exponential-backoff reconnect.
 *
 * Thread-safety: all state is module-global, protected by the assumption
 * that only one goroutine (the UDS writer loop in main.c) calls these
 * functions.  Do not call from multiple threads.
 */

#include <errno.h>
#include <netinet/in.h>
#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <time.h>
#include <unistd.h>

#include "packet_types.h"
#include "uds_sender.h"

/* ---- internal state -------------------------------------------------- */
static int  g_fd        = -1;
static char g_path[108] = {0};   /* UNIX_PATH_MAX */

/* ---- helpers ---------------------------------------------------------- */
static void sleep_sec(int s) {
    struct timespec ts = {.tv_sec = s, .tv_nsec = 0};
    nanosleep(&ts, NULL);
}

static int try_connect(const char *path) {
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) return -1;

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, path, sizeof(addr.sun_path) - 1);

    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        close(fd);
        return -1;
    }
    return fd;
}

/* ---- public API ------------------------------------------------------- */

UDSResult uds_connect(const char *path) {
    strncpy(g_path, path, sizeof(g_path) - 1);

    int backoff = 1;
    while (1) {
        int fd = try_connect(path);
        if (fd >= 0) {
            g_fd = fd;
            return UDS_OK;
        }
        fprintf(stderr, "[uds] connect to %s failed, retry in %ds\n",
                path, backoff);
        sleep_sec(backoff);
        backoff = (backoff * 2 > 30) ? 30 : backoff * 2;
    }
}

UDSResult uds_send_line(const char *json_line) {
    if (g_fd < 0) return UDS_ERR_CONN;

    size_t len = strlen(json_line);
    ssize_t sent = send(g_fd, json_line, len, MSG_NOSIGNAL);
    if (sent < 0) {
        if (errno == EPIPE || errno == ECONNRESET) {
            close(g_fd);
            g_fd = -1;
            return UDS_ERR_PIPE;
        }
        return UDS_ERR_SEND;
    }
    /* Send newline separator */
    send(g_fd, "\n", 1, MSG_NOSIGNAL);
    return UDS_OK;
}

UDSResult uds_send_packet(const PacketInfo *pkt) {
    if (!pkt) return UDS_ERR_SEND;

    /* Base64-encode payload */
    static const char b64[] =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

    char b64_buf[((NS_MAX_PAYLOAD_LEN + 2) / 3) * 4 + 1];
    const uint8_t *in = pkt->payload;
    int in_len = pkt->payload_len;
    char *out = b64_buf;

    for (int i = 0; i < in_len; i += 3) {
        int rem = in_len - i;
        uint32_t v = ((uint32_t)in[i] << 16)
                   | (rem > 1 ? (uint32_t)in[i+1] << 8 : 0)
                   | (rem > 2 ? (uint32_t)in[i+2]       : 0);
        *out++ = b64[(v >> 18) & 0x3F];
        *out++ = b64[(v >> 12) & 0x3F];
        *out++ = rem > 1 ? b64[(v >> 6) & 0x3F] : '=';
        *out++ = rem > 2 ? b64[v & 0x3F]        : '=';
    }
    *out = '\0';

    char buf[NS_MAX_PAYLOAD_LEN * 2 + 512];
    int n = snprintf(buf, sizeof(buf),
        "{\"timestamp_sec\":%lld,\"timestamp_usec\":%d,"
        "\"src_ip\":\"%s\",\"dst_ip\":\"%s\","
        "\"src_port\":%u,\"dst_port\":%u,\"protocol\":%u,"
        "\"tcp_flags\":\"%s\","
        "\"payload_len\":%u,\"payload_preview\":\"%s\","
        "\"is_fragment\":%s,\"truncated\":%s}",
        (long long)pkt->timestamp_sec, pkt->timestamp_usec,
        pkt->src_ip, pkt->dst_ip,
        pkt->src_port, pkt->dst_port, pkt->protocol,
        pkt->tcp_flags,
        pkt->payload_len, b64_buf,
        pkt->is_fragment ? "true" : "false",
        pkt->truncated   ? "true" : "false");

    if (n < 0 || (size_t)n >= sizeof(buf)) return UDS_ERR_SEND;
    return uds_send_line(buf);
}

UDSResult uds_send_heartbeat(const HeartbeatInfo *hb,
                              const char *session_id) {
    char buf[512];
    int n = snprintf(buf, sizeof(buf),
        "{\"type\":\"heartbeat\",\"session_id\":\"%s\","
        "\"seq\":%u,\"sent\":%llu,\"dropped\":%llu,"
        "\"parse_errors\":%llu,\"buf_util_pct\":%u,"
        "\"avg_json_serialize_us\":%.2f,\"uds_write_errors\":%llu}",
        session_id,
        hb->seq, (unsigned long long)hb->sent,
        (unsigned long long)hb->dropped,
        (unsigned long long)hb->parse_errors,
        hb->buf_util_pct, hb->avg_json_serialize_us,
        (unsigned long long)hb->uds_write_errors);
    if (n < 0 || (size_t)n >= sizeof(buf)) return UDS_ERR_SEND;
    return uds_send_line(buf);
}

UDSResult uds_send_hello(const char *session_id, const char *version,
                          int pid, const char *hostname) {
    char buf[256];
    int n = snprintf(buf, sizeof(buf),
        "{\"type\":\"hello\",\"version\":\"%s\","
        "\"session_id\":\"%s\",\"pid\":%d,\"hostname\":\"%s\","
        "\"max_payload_len\":%d}",
        version, session_id, pid, hostname, NS_MAX_PAYLOAD_LEN);
    if (n < 0 || (size_t)n >= sizeof(buf)) return UDS_ERR_SEND;
    return uds_send_line(buf);
}

void uds_close(void) {
    if (g_fd >= 0) {
        close(g_fd);
        g_fd = -1;
    }
}
