#ifndef NETSENTRY_UDS_SENDER_H
#define NETSENTRY_UDS_SENDER_H

#include <stddef.h>
#include "packet_types.h"

#define NS_UDS_PATH_DEFAULT "/tmp/netsentry.sock"

/* Connection state returned by uds_connect() / uds_send_packet() */
typedef enum {
    UDS_OK        =  0,
    UDS_ERR_CONN  = -1,   /* connection failed or lost */
    UDS_ERR_SEND  = -2,   /* send() failed (non-EPIPE) */
    UDS_ERR_PIPE  = -3    /* EPIPE — remote end closed */
} UDSResult;

/* Connect with exponential backoff. max_attempts=0 retries forever. */
UDSResult uds_connect_with_retries(const char *path, unsigned int max_attempts);

/* Connect with exponential backoff until connected. */
UDSResult uds_connect(const char *path);

/* Reconnect to the last configured socket path. */
UDSResult uds_reconnect(void);

/* Reconnect and establish the session before returning the socket to callers. */
UDSResult uds_reconnect_with_hello(const char *session_id, const char *version,
                                   int pid, const char *hostname);

/* Send a serialised JSON line (NUL-terminated). */
UDSResult uds_send_line(const char *json_line);

/* Format JSON frames. Return number of bytes written, or -1 on error/truncation. */
int uds_format_packet_json(const PacketInfo *pkt, char *buf, size_t buf_len);
int uds_format_heartbeat_json(const HeartbeatInfo *hb, const char *session_id,
                              char *buf, size_t buf_len);
int uds_format_hello_json(const char *session_id, const char *version,
                          int pid, const char *hostname,
                          char *buf, size_t buf_len);

/* Send a PacketInfo as a JSON line. */
UDSResult uds_send_packet(const PacketInfo *pkt);

/* Send a heartbeat frame. */
UDSResult uds_send_heartbeat(const HeartbeatInfo *hb, const char *session_id);

/* Send the hello/handshake frame. */
UDSResult uds_send_hello(const char *session_id, const char *version,
                          int pid, const char *hostname);

uint64_t uds_write_errors(void);
double uds_avg_json_serialize_us(void);

void uds_close(void);

#endif /* NETSENTRY_UDS_SENDER_H */
