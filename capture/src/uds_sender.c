/*
 * uds_sender.c — Unix Domain Socket sender with exponential-backoff reconnect.
 *
 * Thread-safety: all state is module-global, protected by the assumption
 * that only one goroutine (the UDS writer loop in main.c) calls these
 * functions. Do not call from multiple threads.
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

static uint64_t g_write_errors       = 0;
static uint64_t g_serialize_count    = 0;
static uint64_t g_serialize_total_ns = 0;

/* ---- helpers ---------------------------------------------------------- */
static void sleep_sec(int s) {
    struct timespec ts = {.tv_sec = s, .tv_nsec = 0};
    nanosleep(&ts, NULL);
}

static uint64_t monotonic_ns(void) {
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) {
        return 0;
    }
    return (uint64_t)ts.tv_sec * 1000000000ull + (uint64_t)ts.tv_nsec;
}

static void record_serialize(uint64_t start_ns) {
    uint64_t end_ns = monotonic_ns();
    if (start_ns == 0 || end_ns < start_ns) {
        return;
    }
    g_serialize_count++;
    g_serialize_total_ns += end_ns - start_ns;
}

static int try_connect(const char *path) {
    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) return -1;

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    if (strlen(path) >= sizeof(addr.sun_path)) {
        close(fd);
        errno = ENAMETOOLONG;
        return -1;
    }
    snprintf(addr.sun_path, sizeof(addr.sun_path), "%s", path);

    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        close(fd);
        return -1;
    }
    return fd;
}

static UDSResult map_send_error(void) {
    g_write_errors++;
    if (errno == EPIPE || errno == ECONNRESET) {
        if (g_fd >= 0) {
            close(g_fd);
            g_fd = -1;
        }
        return UDS_ERR_PIPE;
    }
    return UDS_ERR_SEND;
}

static UDSResult send_all(const char *buf, size_t len) {
    size_t off = 0;
    while (off < len) {
        ssize_t sent = send(g_fd, buf + off, len - off, MSG_NOSIGNAL);
        if (sent < 0) {
            if (errno == EINTR) {
                continue;
            }
            return map_send_error();
        }
        if (sent == 0) {
            errno = EPIPE;
            return map_send_error();
        }
        off += (size_t)sent;
    }
    return UDS_OK;
}

static int json_escape(const char *src, char *dst, size_t dst_len) {
    size_t out = 0;
    if (!src || !dst || dst_len == 0) {
        return -1;
    }

    for (const unsigned char *p = (const unsigned char *)src; *p; p++) {
        const char *esc = NULL;
        char tmp[7];

        switch (*p) {
        case '"': esc = "\\\""; break;
        case '\\': esc = "\\\\"; break;
        case '\b': esc = "\\b"; break;
        case '\f': esc = "\\f"; break;
        case '\n': esc = "\\n"; break;
        case '\r': esc = "\\r"; break;
        case '\t': esc = "\\t"; break;
        default:
            if (*p < 0x20) {
                snprintf(tmp, sizeof(tmp), "\\u%04x", *p);
                esc = tmp;
            }
            break;
        }

        if (esc) {
            size_t n = strlen(esc);
            if (out + n >= dst_len) {
                return -1;
            }
            memcpy(dst + out, esc, n);
            out += n;
        } else {
            if (out + 1 >= dst_len) {
                return -1;
            }
            dst[out++] = (char)*p;
        }
    }
    dst[out] = '\0';
    return (int)out;
}

static int base64_payload(const PacketInfo *pkt, char *buf, size_t buf_len) {
    static const char b64[] =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

    if (!pkt || !buf) return -1;
    if (pkt->payload_len > NS_MAX_PAYLOAD_LEN) return -1;

    size_t needed = ((pkt->payload_len + 2u) / 3u) * 4u + 1u;
    if (needed > buf_len) return -1;

    const uint8_t *in = pkt->payload;
    uint32_t in_len = pkt->payload_len;
    char *out = buf;

    for (uint32_t i = 0; i < in_len; i += 3) {
        uint32_t rem = in_len - i;
        uint32_t v = ((uint32_t)in[i] << 16)
                   | (rem > 1 ? (uint32_t)in[i + 1] << 8 : 0)
                   | (rem > 2 ? (uint32_t)in[i + 2]      : 0);
        *out++ = b64[(v >> 18) & 0x3F];
        *out++ = b64[(v >> 12) & 0x3F];
        *out++ = rem > 1 ? b64[(v >> 6) & 0x3F] : '=';
        *out++ = rem > 2 ? b64[v & 0x3F]        : '=';
    }
    *out = '\0';
    return (int)(out - buf);
}

/* ---- public API ------------------------------------------------------- */

UDSResult uds_connect_with_retries(const char *path, unsigned int max_attempts) {
    if (!path || path[0] == '\0') return UDS_ERR_CONN;

    strncpy(g_path, path, sizeof(g_path) - 1);
    g_path[sizeof(g_path) - 1] = '\0';

    int backoff = 1;
    unsigned int attempts = 0;
    while (max_attempts == 0 || attempts < max_attempts) {
        attempts++;
        int fd = try_connect(path);
        if (fd >= 0) {
            g_fd = fd;
            return UDS_OK;
        }
        if (max_attempts != 0 && attempts >= max_attempts) {
            break;
        }
        fprintf(stderr, "[uds] connect to %s failed, retry in %ds\n",
                path, backoff);
        sleep_sec(backoff);
        backoff = (backoff * 2 > 30) ? 30 : backoff * 2;
    }

    fprintf(stderr, "[uds] connect to %s failed after %u attempt(s)\n",
            path, attempts);
    return UDS_ERR_CONN;
}

UDSResult uds_connect(const char *path) {
    return uds_connect_with_retries(path, 0);
}

UDSResult uds_reconnect(void) {
    if (g_path[0] == '\0') return UDS_ERR_CONN;

    char path[sizeof(g_path)];
    snprintf(path, sizeof(path), "%s", g_path);
    return uds_connect(path);
}

UDSResult uds_send_line(const char *json_line) {
    if (g_fd < 0) return UDS_ERR_CONN;
    if (!json_line) return UDS_ERR_SEND;

    UDSResult r = send_all(json_line, strlen(json_line));
    if (r != UDS_OK) return r;
    return send_all("\n", 1);
}

int uds_format_packet_json(const PacketInfo *pkt, char *buf, size_t buf_len) {
    if (!pkt || !buf || buf_len == 0) return -1;

    uint64_t start_ns = monotonic_ns();

    char src_ip[NS_MAX_IP_STR * 6];
    char dst_ip[NS_MAX_IP_STR * 6];
    char tcp_flags[sizeof(pkt->tcp_flags) * 6];
    char b64_buf[((NS_MAX_PAYLOAD_LEN + 2) / 3) * 4 + 1];

    if (json_escape(pkt->src_ip, src_ip, sizeof(src_ip)) < 0) return -1;
    if (json_escape(pkt->dst_ip, dst_ip, sizeof(dst_ip)) < 0) return -1;
    if (json_escape(pkt->tcp_flags, tcp_flags, sizeof(tcp_flags)) < 0) return -1;
    if (base64_payload(pkt, b64_buf, sizeof(b64_buf)) < 0) return -1;

    int n = snprintf(buf, buf_len,
        "{\"timestamp_sec\":%lld,\"timestamp_usec\":%d,"
        "\"src_ip\":\"%s\",\"dst_ip\":\"%s\","
        "\"src_port\":%u,\"dst_port\":%u,\"protocol\":%u,"
        "\"tcp_flags\":\"%s\","
        "\"payload_len\":%u,\"payload_preview\":\"%s\","
        "\"is_fragment\":%s,\"truncated\":%s}",
        (long long)pkt->timestamp_sec, pkt->timestamp_usec,
        src_ip, dst_ip,
        pkt->src_port, pkt->dst_port, pkt->protocol,
        tcp_flags,
        pkt->payload_len, b64_buf,
        pkt->is_fragment ? "true" : "false",
        pkt->truncated   ? "true" : "false");

    if (n < 0 || (size_t)n >= buf_len) return -1;
    record_serialize(start_ns);
    return n;
}

UDSResult uds_send_packet(const PacketInfo *pkt) {
    char buf[NS_MAX_PAYLOAD_LEN * 2 + 512];
    if (uds_format_packet_json(pkt, buf, sizeof(buf)) < 0) return UDS_ERR_SEND;
    return uds_send_line(buf);
}

int uds_format_heartbeat_json(const HeartbeatInfo *hb, const char *session_id,
                              char *buf, size_t buf_len) {
    if (!hb || !session_id || !buf || buf_len == 0) return -1;

    uint64_t start_ns = monotonic_ns();
    char escaped_session[NS_SESSION_ID_LEN * 6];
    if (json_escape(session_id, escaped_session, sizeof(escaped_session)) < 0) {
        return -1;
    }

    int n = snprintf(buf, buf_len,
        "{\"type\":\"heartbeat\",\"session_id\":\"%s\","
        "\"seq\":%u,\"sent\":%llu,\"dropped\":%llu,"
        "\"parse_errors\":%llu,\"buf_util_pct\":%u,"
        "\"avg_json_serialize_us\":%.2f,\"uds_write_errors\":%llu}",
        escaped_session,
        hb->seq, (unsigned long long)hb->sent,
        (unsigned long long)hb->dropped,
        (unsigned long long)hb->parse_errors,
        hb->buf_util_pct, hb->avg_json_serialize_us,
        (unsigned long long)hb->uds_write_errors);
    if (n < 0 || (size_t)n >= buf_len) return -1;
    record_serialize(start_ns);
    return n;
}

UDSResult uds_send_heartbeat(const HeartbeatInfo *hb,
                              const char *session_id) {
    char buf[512];
    if (uds_format_heartbeat_json(hb, session_id, buf, sizeof(buf)) < 0) {
        return UDS_ERR_SEND;
    }
    return uds_send_line(buf);
}

int uds_format_hello_json(const char *session_id, const char *version,
                          int pid, const char *hostname,
                          char *buf, size_t buf_len) {
    if (!session_id || !version || !hostname || !buf || buf_len == 0) return -1;

    uint64_t start_ns = monotonic_ns();
    char escaped_session[NS_SESSION_ID_LEN * 6];
    char escaped_version[64];
    char escaped_hostname[64 * 6];

    if (json_escape(session_id, escaped_session, sizeof(escaped_session)) < 0) return -1;
    if (json_escape(version, escaped_version, sizeof(escaped_version)) < 0) return -1;
    if (json_escape(hostname, escaped_hostname, sizeof(escaped_hostname)) < 0) return -1;

    int n = snprintf(buf, buf_len,
        "{\"type\":\"hello\",\"version\":\"%s\","
        "\"session_id\":\"%s\",\"pid\":%d,\"hostname\":\"%s\","
        "\"max_payload_len\":%d}",
        escaped_version, escaped_session, pid, escaped_hostname, NS_MAX_PAYLOAD_LEN);
    if (n < 0 || (size_t)n >= buf_len) return -1;
    record_serialize(start_ns);
    return n;
}

UDSResult uds_send_hello(const char *session_id, const char *version,
                          int pid, const char *hostname) {
    char buf[512];
    if (uds_format_hello_json(session_id, version, pid, hostname, buf, sizeof(buf)) < 0) {
        return UDS_ERR_SEND;
    }
    return uds_send_line(buf);
}

uint64_t uds_write_errors(void) {
    return g_write_errors;
}

double uds_avg_json_serialize_us(void) {
    if (g_serialize_count == 0) {
        return 0.0;
    }
    return ((double)g_serialize_total_ns / (double)g_serialize_count) / 1000.0;
}

void uds_close(void) {
    if (g_fd >= 0) {
        close(g_fd);
        g_fd = -1;
    }
}
