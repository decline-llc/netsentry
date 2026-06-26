#ifndef NETSENTRY_PARSER_H
#define NETSENTRY_PARSER_H

#include <stdint.h>
#include "packet_types.h"

/* -----------------------------------------------------------------------
 * Protocol parser return codes (Chain-of-Responsibility pattern).
 * v0.1.0: only passthrough_parser is registered.
 * v0.2.0: http_basic, dns_basic will be added via parser_registry_register().
 * ----------------------------------------------------------------------- */
typedef enum {
    PARSE_PASS  =  0,   /* unrecognised — try next parser */
    PARSE_CLAIM =  1,   /* parsed successfully — stop chain */
    PARSE_ERROR = -1    /* malformed data — stop chain, log error */
} ParseResult;

typedef ParseResult (*ParserFunc)(const uint8_t *payload, uint32_t len,
                                  PacketInfo *info);

typedef struct {
    uint16_t    port;       /* 0 = match all ports */
    uint8_t     protocol;   /* IPPROTO_TCP / IPPROTO_UDP */
    ParserFunc  func;
    const char *name;
} ParserEntry;

/* Registry API */
void       parser_registry_register(uint8_t protocol, uint16_t port,
                                     ParserFunc func, const char *name);
ParserFunc parser_registry_lookup(uint8_t protocol, uint16_t port);
void       parser_registry_run(const uint8_t *payload, uint32_t len,
                                uint8_t protocol, uint16_t dst_port,
                                PacketInfo *info);

/* Built-in parsers */
ParseResult passthrough_parser(const uint8_t *payload, uint32_t len,
                               PacketInfo *info);

#endif /* NETSENTRY_PARSER_H */
