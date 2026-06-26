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

TEST_PARSER_TARGET := ../bin/test-parser
TEST_PARSER_SRCS   := tests/test_parser.c \
                      src/eth_parser.c \
                      src/parser_registry.c \
                      src/passthrough_parser.c

TEST_UDS_TARGET := ../bin/test-uds-sender
TEST_UDS_SRCS   := tests/test_uds_sender.c \
                   src/uds_sender.c

TEST_TARGETS := $(TEST_PARSER_TARGET) $(TEST_UDS_TARGET)

BENCH_PARSER_TARGET := ../bin/bench-parser
BENCH_PARSER_SRCS   := tests/bench_parser.c \
                       src/eth_parser.c \
                       src/parser_registry.c \
                       src/passthrough_parser.c

BENCH_UDS_TARGET := ../bin/bench-uds-sender
BENCH_UDS_SRCS   := tests/bench_uds_sender.c \
                    src/uds_sender.c

BENCH_TARGETS := $(BENCH_PARSER_TARGET) $(BENCH_UDS_TARGET)
BENCH_ITERATIONS ?= 100000

.PHONY: all test asan-test bench clean

all: $(TARGET)

$(TARGET): $(OBJS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

$(TEST_PARSER_TARGET): $(TEST_PARSER_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^

$(TEST_UDS_TARGET): $(TEST_UDS_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^

$(BENCH_PARSER_TARGET): $(BENCH_PARSER_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^

$(BENCH_UDS_TARGET): $(BENCH_UDS_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^

test: $(TEST_TARGETS)
	$(TEST_PARSER_TARGET)
	$(TEST_UDS_TARGET)

asan-test: CFLAGS += -fsanitize=address -fno-omit-frame-pointer -g
asan-test: clean $(TEST_TARGETS)
	$(TEST_PARSER_TARGET)
	$(TEST_UDS_TARGET)

bench: $(BENCH_TARGETS)
	$(BENCH_PARSER_TARGET) $(BENCH_ITERATIONS)
	$(BENCH_UDS_TARGET) $(BENCH_ITERATIONS)

clean:
	rm -f $(OBJS) $(TARGET) $(TEST_TARGETS) $(BENCH_TARGETS)
