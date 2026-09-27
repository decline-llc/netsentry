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
#include <fcntl.h>
#include <getopt.h>
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
#include "slo.h"
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
static const char *g_slo_run_id = NULL;
static unsigned int g_slo_port = 0;
static uint64_t g_slo_ignored = 0;
static uint64_t g_slo_invalid = 0;
static int g_slo_failed = 0;

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

    if (g_slo_run_id) {
        if (info.protocol != IPPROTO_UDP || info.dst_port != g_slo_port || info.is_fragment) {
            g_slo_ignored++;
            return;
        }
        int marked = ns_slo_attach(&info, g_slo_run_id);
        if (marked <= 0) {
            if (marked == 0) g_slo_ignored++;
            else g_slo_invalid++;
            if (marked == -2) {
                g_slo_failed = 1;
                g_running = 0;
            }
            return;
        }
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

static int sync_parent(const char *path) {
    char *copy = strdup(path);
    if (!copy) return -1;
    char *slash = strrchr(copy, '/');
    const char *parent = ".";
    if (slash) {
        if (slash == copy) slash[1] = '\0';
        else *slash = '\0';
        parent = copy;
    }
    int fd = open(parent, O_RDONLY | O_DIRECTORY);
    free(copy);
    if (fd < 0) return -1;
    int rc = fsync(fd);
    if (close(fd) != 0) rc = -1;
    return rc;
}

/* ---- main ------------------------------------------------------------- */
int main(int argc, char *argv[]) {
    const char *pcap_file  = NULL;
    const char *iface      = NULL;
    const char *uds_path   = NS_DEFAULT_UDS;
    int connect_retries    = -1;
    const char *slo_output = NULL;
    static const struct option options[] = {
        {"slo-run-id", required_argument, NULL, 1000},
        {"slo-port", required_argument, NULL, 1001},
        {"slo-output", required_argument, NULL, 1002},
        {NULL, 0, NULL, 0}
    };

    int opt;
    while ((opt = getopt_long(argc, argv, "r:i:s:c:", options, NULL)) != -1) {
        switch (opt) {
        case 'r': pcap_file = optarg; break;
        case 'i': iface     = optarg; break;
        case 's': uds_path  = optarg; break;
        case 1000: g_slo_run_id = optarg; break;
        case 1001: {
            char *end = NULL;
            errno = 0;
            unsigned long port = strtoul(optarg, &end, 10);
            if (errno || !optarg[0] || optarg[0] < '0' || optarg[0] > '9' ||
                *end || port < 1 || port > 65535) {
                fprintf(stderr, "[capture] slo-port must be 1..65535\n");
                return 2;
            }
            g_slo_port = (unsigned int)port;
            break;
        }
        case 1002: slo_output = optarg; break;
        case 'c':
            if (parse_connect_retries(optarg, &connect_retries) != 0) {
                fprintf(stderr,
                        "[capture] connect_retries must be an integer from 0 to %d\n",
                        NS_MAX_CONNECT_RETRIES);
                return 2;
            }
            break;
        default:
            fprintf(stderr, "Usage: %s [-r pcap] [-i iface] [-s uds_path] [-c connect_retries] [--slo-run-id ID --slo-port PORT --slo-output NEW_FILE]\n",
                    argv[0]);
            return 1;
        }
    }

    if ((g_slo_run_id || g_slo_port || slo_output) &&
        (!ns_slo_identifier(g_slo_run_id) || !g_slo_port || !slo_output || !slo_output[0] ||
         !iface || !iface[0] || pcap_file || optind != argc)) {
        fprintf(stderr, "[capture] measurement requires live -i plus all three valid --slo-* options; -r is forbidden\n");
        return 2;
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
    } else if (g_slo_run_id) {
        handle = pcap_create(iface, errbuf);
        if (handle && (pcap_set_snaplen(handle, 65535) != 0 ||
                       pcap_set_promisc(handle, 1) != 0 ||
                       pcap_set_timeout(handle, 1000) != 0 ||
                       pcap_set_tstamp_type(handle, PCAP_TSTAMP_HOST) != 0 ||
                       pcap_set_tstamp_precision(handle, PCAP_TSTAMP_PRECISION_MICRO) != 0 ||
                       pcap_activate(handle) != 0 ||
                       pcap_get_tstamp_precision(handle) != PCAP_TSTAMP_PRECISION_MICRO ||
                       pcap_setdirection(handle, PCAP_D_IN) != 0)) {
            fprintf(stderr, "[capture] measurement timestamp/direction activation failed: %s\n", pcap_geterr(handle));
            pcap_close(handle);
            uds_close();
            return 2;
        }
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
    char filter[80] = "ip";
    if (g_slo_run_id) snprintf(filter, sizeof(filter), "ip and udp dst port %u", g_slo_port);
    int filter_failed = 0;
    if (pcap_compile(handle, &fp, filter, 0, PCAP_NETMASK_UNKNOWN) == 0) {
        if (pcap_setfilter(handle, &fp) != 0) {
            filter_failed = 1;
            fprintf(stderr, "[capture] warning: pcap filter install failed: %s\n",
                    pcap_geterr(handle));
        }
        pcap_freecode(&fp);
    } else {
        filter_failed = 1;
        fprintf(stderr, "[capture] warning: pcap filter compile failed: %s\n",
                pcap_geterr(handle));
    }

    int summary_fd = -1;
    if (g_slo_run_id) {
        if (filter_failed || (summary_fd = open(slo_output, O_WRONLY | O_CREAT | O_EXCL, 0600)) < 0) {
            fprintf(stderr, "[capture] measurement requires a working filter and new summary file\n");
            pcap_close(handle);
            uds_close();
            return 2;
        }
    }

    fprintf(stderr, "[capture] starting capture (source=%s)\n",
            pcap_file ? pcap_file : iface);

    time_t last_hb = time(NULL);
    while (g_running) {
        int rc = pcap_dispatch(handle, 64, packet_handler, NULL);
        if (rc == PCAP_ERROR) {
            if (g_slo_run_id) g_slo_failed = 1;
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

    if (g_slo_run_id) {
        struct pcap_stat ps = {0};
        int stats_ok = pcap_stats(handle, &ps) == 0;
        if (!stats_ok) g_slo_failed = 1;
        char summary[2048];
        int n = snprintf(summary, sizeof(summary),
            "{\"schema_version\":1,\"run_id\":\"%s\",\"capture_closed\":%s,"
            "\"execution_complete\":false,\"slo_compliance_asserted\":false,"
            "\"timestamp_type\":\"host\",\"timestamp_precision\":\"microseconds\","
            "\"direction\":\"inbound\",\"udp_port\":%u,\"sent\":%llu,\"dropped\":%llu,"
            "\"parse_errors\":%llu,\"ignored\":%llu,\"invalid_markers\":%llu,"
            "\"pcap_stats_valid\":%s,\"pcap_received\":%u,\"pcap_dropped\":%u,\"pcap_interface_dropped\":%u}\n",
            g_slo_run_id, g_slo_failed ? "false" : "true", g_slo_port,
            (unsigned long long)g_sent, (unsigned long long)g_dropped,
            (unsigned long long)g_parse_errors, (unsigned long long)g_slo_ignored,
            (unsigned long long)g_slo_invalid, stats_ok ? "true" : "false",
            ps.ps_recv, ps.ps_drop, ps.ps_ifdrop);
        if (n < 0 || (size_t)n >= sizeof(summary) || write(summary_fd, summary, (size_t)n) != n) g_slo_failed = 1;
        if (fsync(summary_fd) != 0) g_slo_failed = 1;
        if (close(summary_fd) != 0) g_slo_failed = 1;
        if (sync_parent(slo_output) != 0) g_slo_failed = 1;
    }

    pcap_close(handle);
    uds_close();
    return g_slo_failed ? 2 : 0;
}
