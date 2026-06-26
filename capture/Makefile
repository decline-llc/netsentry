# capture/Makefile
# Builds the netsentry-capture binary from C sources.

CC      := gcc
CFLAGS  := -std=c11 -Wall -Wextra -O2 \
            -D_DEFAULT_SOURCE -D_POSIX_C_SOURCE=200809L \
            -Iinclude \
            $(shell pkg-config --cflags libpcap 2>/dev/null)
LDFLAGS := $(shell pkg-config --libs libpcap 2>/dev/null || echo "-lpcap")

SRCS := src/main.c \
        src/eth_parser.c \
        src/parser_registry.c \
        src/passthrough_parser.c \
        src/uds_sender.c

OBJS    := $(SRCS:.c=.o)
TARGET  := ../bin/netsentry-capture

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(TARGET)
