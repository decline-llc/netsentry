/*
 * main.c — NetSentry packet capture entry point (v0.1.0).
 *
 * Usage: netsentry-capture [-r <pcap_file>] [-i <iface>]
 *                          [-s <uds_socket>] [-p <payload_len>]
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

/* ---- globals ---------------------------------------------------------- */
static volatile int g_running = 1;

static uint64_t g_sent         = 0;
static uint64_t g_dropped      = 0;
static uint64_t g_parse_errors = 0;
static uint32_t g_hb_seq       = 0;

static char g_session_id[NS_SESSION_ID_LEN];

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
        if (r == UDS_ERR_PIPE) {
            /* Go engine disconnected — attempt reconnect */
            fprintf(stderr, "[capture] UDS pipe broken, reconnecting…\n");
            uds_connect(NS_DEFAULT_UDS);
        }
    }
}

/* ---- main ------------------------------------------------------------- */
int main(int argc, char *argv[]) {
    const char *pcap_file  = NULL;
    const char *iface      = NULL;
    const char *uds_path   = NS_DEFAULT_UDS;

    int opt;
    while ((opt = getopt(argc, argv, "r:i:s:")) != -1) {
        switch (opt) {
        case 'r': pcap_file = optarg; break;
        case 'i': iface     = optarg; break;
        case 's': uds_path  = optarg; break;
        default:
            fprintf(stderr, "Usage: %s [-r pcap] [-i iface] [-s uds_path]\n",
                    argv[0]);
            return 1;
        }
    }

    if (!pcap_file && !iface) {
        fprintf(stderr, "[capture] must specify -r <pcap> or -i <iface>\n");
        return 1;
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
    uds_connect(uds_path);

    char hostname[64] = {0};
    gethostname(hostname, sizeof(hostname) - 1);
    uds_send_hello(g_session_id, NS_VERSION, getpid(), hostname);

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
        return 1;
    }

    /* Only capture IP packets */
    struct bpf_program fp;
    if (pcap_compile(handle, &fp, "ip", 0, PCAP_NETMASK_UNKNOWN) == 0) {
        pcap_setfilter(handle, &fp);
        pcap_freecode(&fp);
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
                .avg_json_serialize_us = 0.0,
                .uds_write_errors     = 0,
            };
            uds_send_heartbeat(&hb, g_session_id);
            last_hb = now;
        }
    }

    /* Final heartbeat before exit */
    HeartbeatInfo final_hb = {
        .seq          = ++g_hb_seq,
        .sent         = g_sent,
        .dropped      = g_dropped,
        .parse_errors = g_parse_errors,
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
