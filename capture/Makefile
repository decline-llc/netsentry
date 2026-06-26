# capture/Makefile
# Builds the netsentry-capture binary and lightweight parser tests.

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

TEST_TARGET := ../bin/test-parser
TEST_SRCS   := tests/test_parser.c \
               src/eth_parser.c \
               src/parser_registry.c \
               src/passthrough_parser.c

.PHONY: all test asan-test clean

all: $(TARGET)

$(TARGET): $(OBJS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

$(TEST_TARGET): $(TEST_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^

test: $(TEST_TARGET)
	$(TEST_TARGET)

asan-test: CFLAGS += -fsanitize=address -fno-omit-frame-pointer -g
asan-test: clean $(TEST_TARGET)
	$(TEST_TARGET)

clean:
	rm -f $(OBJS) $(TARGET) $(TEST_TARGET)
