/*
 * parser_registry.c — simple linear registry for protocol parsers.
 *
 * v0.1.0: at most 16 registered parsers.  Registration is done once
 * at startup before any packet is processed, so no locking is needed.
 */

#include <stdio.h>
#include <string.h>
#include "packet_types.h"
#include "parser.h"

#define MAX_PARSERS 16

static ParserEntry registry[MAX_PARSERS];
static int         registry_count = 0;

void parser_registry_register(uint8_t protocol, uint16_t port,
                               ParserFunc func, const char *name) {
    if (registry_count >= MAX_PARSERS) {
        fprintf(stderr, "[parser_registry] max parsers (%d) reached, "
                        "cannot register \"%s\"\n", MAX_PARSERS, name);
        return;
    }
    registry[registry_count++] = (ParserEntry){
        .port     = port,
        .protocol = protocol,
        .func     = func,
        .name     = name,
    };
}

ParserFunc parser_registry_lookup(uint8_t protocol, uint16_t port) {
    for (int i = 0; i < registry_count; i++) {
        const ParserEntry *e = &registry[i];
        if (e->protocol != protocol) continue;
        if (e->port != 0 && e->port != port) continue;
        return e->func;
    }
    return NULL;
}

void parser_registry_run(const uint8_t *payload, uint32_t len,
                          uint8_t protocol, uint16_t dst_port,
                          PacketInfo *info) {
    for (int i = 0; i < registry_count; i++) {
        const ParserEntry *e = &registry[i];
        if (e->protocol != protocol) continue;
        if (e->port != 0 && e->port != dst_port) continue;

        ParseResult r = e->func(payload, len, info);
        if (r == PARSE_CLAIM || r == PARSE_ERROR) return;
        /* PARSE_PASS → continue to next parser */
    }
}
