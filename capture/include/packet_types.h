#ifndef NETSENTRY_PACKET_TYPES_H
#define NETSENTRY_PACKET_TYPES_H

#include <stdint.h>
#include <stdbool.h>

#define NS_MAX_PAYLOAD_LEN   4096
#define NS_MAX_IP_STR        40
#define NS_SESSION_ID_LEN    9   /* 8 hex chars + NUL */
#define NS_TCP_FLAGS_LEN     32

typedef struct {
    int64_t  timestamp_sec;
    int32_t  timestamp_usec;
    char     src_ip[NS_MAX_IP_STR];
    char     dst_ip[NS_MAX_IP_STR];
    uint16_t src_port;
    uint16_t dst_port;
    uint8_t  protocol;          /* IPPROTO_TCP / IPPROTO_UDP */
    char     tcp_flags[NS_TCP_FLAGS_LEN]; /* e.g. "SYN", "ACK", "FIN|ACK" */
    uint16_t payload_len;
    uint8_t  payload[NS_MAX_PAYLOAD_LEN];
    bool     is_fragment;
    bool     truncated;
} PacketInfo;

typedef struct {
    char     session_id[NS_SESSION_ID_LEN];
    uint32_t seq;
    uint64_t sent;
    uint64_t dropped;
    uint64_t parse_errors;
    uint32_t buf_util_pct;
    double   avg_json_serialize_us;
    uint64_t uds_write_errors;
} HeartbeatInfo;

#endif /* NETSENTRY_PACKET_TYPES_H */
