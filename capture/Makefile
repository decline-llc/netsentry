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
HEADERS := include/eth_parser.h \
           include/packet_types.h \
           include/parser.h \
           include/uds_sender.h

OBJS    := $(SRCS:.c=.o)
TARGET  := ../bin/netsentry-capture
ASAN_TARGET := ../bin/netsentry-capture-asan
ASAN_CFLAGS := $(CFLAGS) -fsanitize=address -fno-omit-frame-pointer -g

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

FUZZ_PARSER_TARGET := ../bin/fuzz-parser
FUZZ_PARSER_SRCS   := tests/fuzz_parser.c \
                      src/eth_parser.c \
                      src/parser_registry.c \
                      src/passthrough_parser.c
FUZZ_ITERATIONS ?= 5000

.PHONY: all build-asan test asan-test bench fuzz-parser clean

all: $(TARGET)

$(TARGET): $(OBJS)
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -c -o $@ $<

build-asan: $(ASAN_TARGET)

$(ASAN_TARGET): $(SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(ASAN_CFLAGS) -o $@ $^ $(LDFLAGS)

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

$(FUZZ_PARSER_TARGET): $(FUZZ_PARSER_SRCS)
	@mkdir -p $(dir $@)
	$(CC) $(ASAN_CFLAGS) -o $@ $^

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

fuzz-parser: $(FUZZ_PARSER_TARGET)
	ASAN_OPTIONS=detect_leaks=0 FUZZ_ITERATIONS=$(FUZZ_ITERATIONS) $(FUZZ_PARSER_TARGET)

clean:
	rm -f $(OBJS) $(TARGET) $(ASAN_TARGET) $(TEST_TARGETS) $(BENCH_TARGETS) $(FUZZ_PARSER_TARGET)
