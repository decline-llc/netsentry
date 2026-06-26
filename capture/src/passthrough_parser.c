/*
 * passthrough_parser.c — v0.1.0 default parser.
 *
 * Copies raw payload bytes into PacketInfo without interpreting
 * application-layer content.  AC automaton matching is done in Go.
 */

#include <string.h>
#include "packet_types.h"
#include "parser.h"

ParseResult passthrough_parser(const uint8_t *payload, uint32_t len,
                               PacketInfo *info) {
    if (!payload || !info) return PARSE_ERROR;

    uint32_t copy_len = len;
    bool truncated = false;
    if (copy_len > NS_MAX_PAYLOAD_LEN) {
        copy_len = NS_MAX_PAYLOAD_LEN;
        truncated = true;
    }
    memcpy(info->payload, payload, copy_len);
    info->payload_len = (uint16_t)copy_len;
    info->truncated   = truncated;
    return PARSE_CLAIM;
}
