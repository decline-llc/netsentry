#ifndef NETSENTRY_ETH_PARSER_H
#define NETSENTRY_ETH_PARSER_H

#include <stdint.h>
#include "packet_types.h"

/*
 * Parse a raw libpcap frame into PacketInfo.
 * Returns 0 on success, -1 if the packet is malformed or unsupported.
 * Fragments are parsed up to the IP header; transport layer is skipped
 * and is_fragment is set to true.
 */
int parse_frame(const uint8_t *pkt, uint32_t pkt_len,
                int64_t ts_sec, int32_t ts_usec,
                PacketInfo *info);

#endif /* NETSENTRY_ETH_PARSER_H */
