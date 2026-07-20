/*
 * main.c — NetSentry packet capture entry point (v0.1.0).
 *
 * Usage: netsentry-capture [-r <pcap_file>] [-i <iface>]
 *                          [-s <uds_socket>] [-c <connect_retries>]
 *
 * Reads packets from a pcap file (offline mode) or live interface,
 * serialises each as a JSON line, and forwards to the Go engine over
 * a Unix Domain Socket.  Sends a heartbeat every 5 seconds.
 */

#include <errno.h>
#include <pcap.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "eth_parser.h"
#include "packet_types.h"
#include "parser.h"
#include "uds_sender.h"

#define NS_VERSION          "0.1.0"
#define NS_HEARTBEAT_SEC    5
#define NS_DEFAULT_UDS      "/tmp/netsentry.sock"
#define NS_MAX_CONNECT_RETRIES 1000

/* ---- globals ---------------------------------------------------------- */
static volatile sig_atomic_t g_running = 1;

static uint64_t g_sent         = 0;
static uint64_t g_dropped      = 0;
static uint64_t g_parse_errors = 0;
static uint32_t g_hb_seq       = 0;

static char g_session_id[NS_SESSION_ID_LEN];
static char g_hostname[64];

/* ---- signal handler --------------------------------------------------- */
static void sig_handler(int sig) {
    (void)sig;
    g_running = 0;
}

/* ---- session ID ------------------------------------------------------- */
static void gen_session_id(char *out) {
    unsigned int seed = (unsigned int)time(NULL) ^ (unsigned int)getpid();
    snprintf(out, NS_SESSION_ID_LEN, "%08x", seed);
}

static int parse_connect_retries(const char *value, int *result) {
    char *end = NULL;
    unsigned long parsed;

    if (!value || value[0] == '\0' || !result) return -1;
    errno = 0;
    parsed = strtoul(value, &end, 10);
    if (errno == ERANGE || end == value || *end != '\0' ||
        parsed > NS_MAX_CONNECT_RETRIES) {
        return -1;
    }
    *result = (int)parsed;
    return 0;
}

static UDSResult reconnect_session(void) {
    fprintf(stderr, "[capture] UDS connection lost, reconnecting session\n");
    return uds_reconnect_with_hello(g_session_id, NS_VERSION, getpid(),
                                    g_hostname);
}

/* ---- pcap callback ---------------------------------------------------- */
static void packet_handler(uint8_t *user, const struct pcap_pkthdr *hdr,
                            const uint8_t *raw) {
    (void)user;
    PacketInfo info;
    int rc = parse_frame(raw, hdr->caplen,
                          (int64_t)hdr->ts.tv_sec,
                          (int32_t)hdr->ts.tv_usec,
                          &info);
    if (rc < 0) {
        g_parse_errors++;
        return;
    }

    UDSResult r = uds_send_packet(&info);
    if (r == UDS_OK) {
        g_sent++;
    } else {
        g_dropped++;
        if (r == UDS_ERR_PIPE || r == UDS_ERR_CONN) {
            /* Establish hello on the replacement connection before more traffic. */
            if (reconnect_session() != UDS_OK) {
                g_dropped++;
            }
        }
    }
}

/* ---- main ------------------------------------------------------------- */
int main(int argc, char *argv[]) {
    const char *pcap_file  = NULL;
    const char *iface      = NULL;
    const char *uds_path   = NS_DEFAULT_UDS;
    int connect_retries    = -1;

    int opt;
    while ((opt = getopt(argc, argv, "r:i:s:c:")) != -1) {
        switch (opt) {
        case 'r': pcap_file = optarg; break;
        case 'i': iface     = optarg; break;
        case 's': uds_path  = optarg; break;
        case 'c':
            if (parse_connect_retries(optarg, &connect_retries) != 0) {
                fprintf(stderr,
                        "[capture] connect_retries must be an integer from 0 to %d\n",
                        NS_MAX_CONNECT_RETRIES);
                return 2;
            }
            break;
        default:
            fprintf(stderr, "Usage: %s [-r pcap] [-i iface] [-s uds_path] [-c connect_retries]\n",
                    argv[0]);
            return 1;
        }
    }

    if (!pcap_file && !iface) {
        fprintf(stderr, "[capture] must specify -r <pcap> or -i <iface>\n");
        return 1;
    }

    if (connect_retries < 0) {
        connect_retries = pcap_file ? 5 : 0;
    }

    signal(SIGINT,  sig_handler);
    signal(SIGTERM, sig_handler);

    /* Register default passthrough parser */
    parser_registry_register(IPPROTO_TCP, 0, passthrough_parser, "passthrough_tcp");
    parser_registry_register(IPPROTO_UDP, 0, passthrough_parser, "passthrough_udp");

    /* Generate session ID and connect to Go engine */
    gen_session_id(g_session_id);
    fprintf(stderr, "[capture] session_id=%s, connecting to %s\n",
            g_session_id, uds_path);
    if (uds_connect_with_retries(uds_path, (unsigned int)connect_retries) != UDS_OK) {
        fprintf(stderr, "[capture] unable to connect to %s\n", uds_path);
        return 1;
    }

    gethostname(g_hostname, sizeof(g_hostname) - 1);
    if (uds_send_hello(g_session_id, NS_VERSION, getpid(), g_hostname) != UDS_OK) {
        fprintf(stderr, "[capture] failed to send hello frame\n");
        uds_close();
        return 1;
    }

    /* Open pcap source */
    char errbuf[PCAP_ERRBUF_SIZE];
    pcap_t *handle = NULL;
    if (pcap_file) {
        handle = pcap_open_offline(pcap_file, errbuf);
    } else {
        handle = pcap_open_live(iface, 65535, 1, 1000, errbuf);
    }
    if (!handle) {
        fprintf(stderr, "[capture] pcap open failed: %s\n", errbuf);
        uds_close();
        return 1;
    }

    int datalink = pcap_datalink(handle);
    if (datalink != DLT_EN10MB) {
        const char *datalink_name = pcap_datalink_val_to_name(datalink);
        if (!datalink_name) datalink_name = "unknown";
        fprintf(stderr,
                "[capture] unsupported pcap data link type %d (%s); Ethernet is required\n",
                datalink, datalink_name);
        pcap_close(handle);
        uds_close();
        return 2;
    }

    /* Only capture IP packets */
    struct bpf_program fp;
    if (pcap_compile(handle, &fp, "ip", 0, PCAP_NETMASK_UNKNOWN) == 0) {
        if (pcap_setfilter(handle, &fp) != 0) {
            fprintf(stderr, "[capture] warning: pcap filter install failed: %s\n",
                    pcap_geterr(handle));
        }
        pcap_freecode(&fp);
    } else {
        fprintf(stderr, "[capture] warning: pcap filter compile failed: %s\n",
                pcap_geterr(handle));
    }

    fprintf(stderr, "[capture] starting capture (source=%s)\n",
            pcap_file ? pcap_file : iface);

    time_t last_hb = time(NULL);
    while (g_running) {
        int rc = pcap_dispatch(handle, 64, packet_handler, NULL);
        if (rc == PCAP_ERROR) {
            fprintf(stderr, "[capture] pcap_dispatch: %s\n",
                    pcap_geterr(handle));
            break;
        }
        if (rc == PCAP_ERROR_BREAK || (pcap_file && rc == 0)) {
            /* EOF for offline mode */
            break;
        }

        time_t now = time(NULL);
        if (now - last_hb >= NS_HEARTBEAT_SEC) {
            HeartbeatInfo hb = {
                .seq                  = ++g_hb_seq,
                .sent                 = g_sent,
                .dropped              = g_dropped,
                .parse_errors         = g_parse_errors,
                .buf_util_pct         = 0,
                .avg_json_serialize_us = uds_avg_json_serialize_us(),
                .uds_write_errors     = uds_write_errors(),
            };
            UDSResult hb_result = uds_send_heartbeat(&hb, g_session_id);
            if (hb_result == UDS_ERR_PIPE || hb_result == UDS_ERR_CONN) {
                if (reconnect_session() != UDS_OK) {
                    g_dropped++;
                }
            }
            last_hb = now;
        }
    }

    /* Final heartbeat before exit */
    HeartbeatInfo final_hb = {
        .seq                   = ++g_hb_seq,
        .sent                  = g_sent,
        .dropped               = g_dropped,
        .parse_errors          = g_parse_errors,
        .buf_util_pct          = 0,
        .avg_json_serialize_us = uds_avg_json_serialize_us(),
        .uds_write_errors      = uds_write_errors(),
    };
    uds_send_heartbeat(&final_hb, g_session_id);

    fprintf(stderr, "[capture] done. sent=%llu dropped=%llu parse_errors=%llu\n",
            (unsigned long long)g_sent,
            (unsigned long long)g_dropped,
            (unsigned long long)g_parse_errors);

    pcap_close(handle);
    uds_close();
    return 0;
}
