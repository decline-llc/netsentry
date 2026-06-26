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

/* Connect with exponential backoff.  Blocks until connected or cancelled. */
UDSResult uds_connect(const char *path);

/* Send a serialised JSON line (NUL-terminated). */
UDSResult uds_send_line(const char *json_line);

/* Send a PacketInfo as a JSON line. */
UDSResult uds_send_packet(const PacketInfo *pkt);

/* Send a heartbeat frame. */
UDSResult uds_send_heartbeat(const HeartbeatInfo *hb, const char *session_id);

/* Send the hello/handshake frame. */
UDSResult uds_send_hello(const char *session_id, const char *version,
                          int pid, const char *hostname);

void uds_close(void);

#endif /* NETSENTRY_UDS_SENDER_H */
