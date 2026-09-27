#ifndef NETSENTRY_SLO_H
#define NETSENTRY_SLO_H

#include <stddef.h>
#include <string.h>
#include "packet_types.h"

static inline int ns_slo_identifier(const char *id) {
    if (!id) return 0;
    size_t n = strnlen(id, NS_SLO_ID_LEN);
    if (n == 0 || n >= NS_SLO_ID_LEN) return 0;
    for (size_t i = 0; i < n; i++) {
        unsigned char c = (unsigned char)id[i];
        int alnum = (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
                    (c >= '0' && c <= '9');
        if (!alnum && (i == 0 || (c != '_' && c != '.' && c != ':' && c != '-'))) return 0;
    }
    return 1;
}

/* 1: matching marker, 0: unmarked/foreign, -1: malformed, -2: invalid time.
 * Spaces delimit IDs because ':' is legal inside the shared ID grammar.
 * Marker remains part of the inspected application payload. */
static inline int ns_slo_attach(PacketInfo *pkt, const char *run_id) {
    static const char prefix[] = "NSLO1 ";
    if (!pkt || !ns_slo_identifier(run_id) || pkt->payload_len > NS_MAX_PAYLOAD_LEN) return -1;
    if (pkt->payload_len < sizeof(prefix) - 1 ||
        memcmp(pkt->payload, prefix, sizeof(prefix) - 1) != 0) return 0;
    const uint8_t *run = pkt->payload + sizeof(prefix) - 1;
    size_t remaining = pkt->payload_len - (sizeof(prefix) - 1);
    const uint8_t *space = memchr(run, ' ', remaining);
    if (!space) return -1;
    size_t run_len = (size_t)(space - run);
    const uint8_t *packet = space + 1;
    remaining -= run_len + 1;
    const uint8_t *newline = memchr(packet, '\n', remaining);
    if (!newline || run_len == 0 || run_len >= NS_SLO_ID_LEN) return -1;
    size_t packet_len = (size_t)(newline - packet);
    if (packet_len == 0 || packet_len >= NS_SLO_ID_LEN) return -1;
    memcpy(pkt->slo_run_id, run, run_len);
    pkt->slo_run_id[run_len] = '\0';
    memcpy(pkt->slo_packet_id, packet, packet_len);
    pkt->slo_packet_id[packet_len] = '\0';
    if (!ns_slo_identifier(pkt->slo_run_id) || !ns_slo_identifier(pkt->slo_packet_id) ||
        strlen(pkt->slo_run_id) != run_len || strlen(pkt->slo_packet_id) != packet_len) return -1;
    if (strcmp(pkt->slo_run_id, run_id) != 0) return 0;
    if (pkt->timestamp_sec <= 0 || pkt->timestamp_usec < 0 || pkt->timestamp_usec >= 1000000 ||
        pkt->timestamp_sec > (INT64_MAX - (int64_t)pkt->timestamp_usec * 1000) / 1000000000) return -2;
    pkt->slo_arrival_unix_ns = pkt->timestamp_sec * 1000000000 + (int64_t)pkt->timestamp_usec * 1000;
    pkt->slo_enabled = true;
    return 1;
}

#endif
