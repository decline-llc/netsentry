# NetSentry root Makefile

SHELL      := /bin/bash
GO         := go
GOPROXY    ?= https://goproxy.cn,direct
GO_MODULE  := ./engine
BIN_DIR    := bin

.PHONY: all build-c build-go build test asan-test bench lint clean quickstart help

all: build

## build-c   — compile the C capture binary
build-c:
	$(MAKE) -C capture

## build-go  — compile the Go engine binary
build-go:
	@mkdir -p $(BIN_DIR)
	cd $(GO_MODULE) && GOPROXY=$(GOPROXY) $(GO) build -o ../$(BIN_DIR)/netsentry-engine \
	    ./cmd/netsentry

## build     — compile both C and Go binaries
build: build-c build-go

## test      — run C parser tests and Go unit tests with race detector
test:
	$(MAKE) -C capture test
	cd $(GO_MODULE) && GOPROXY=$(GOPROXY) $(GO) test -race -count=1 ./...

## asan-test — run C parser tests with AddressSanitizer
asan-test:
	$(MAKE) -C capture asan-test

## bench     — run Go benchmarks (10 s each)
bench:
	cd $(GO_MODULE) && GOPROXY=$(GOPROXY) $(GO) test -bench=. -benchtime=10s -benchmem ./...

## lint      — run go vet and staticcheck (if installed)
lint:
	cd $(GO_MODULE) && $(GO) vet ./...
	@command -v staticcheck >/dev/null 2>&1 && \
	    cd $(GO_MODULE) && staticcheck ./... || \
	    echo "[lint] staticcheck not found, skipping"

## clean     — remove compiled binaries and object files
clean:
	$(MAKE) -C capture clean
	rm -rf $(BIN_DIR)

## quickstart — generate a sample pcap and run a full analysis
quickstart: build
	@echo "==> Generating sample pcap…"
	@command -v python3 >/dev/null 2>&1 || { echo "python3 not found"; exit 1; }
	@python3 scripts/gen_test_pcap.py
	@echo "==> Starting engine in background…"
	@mkdir -p data
	$(BIN_DIR)/netsentry-engine -config configs/config.yaml &
	@ENGINE_PID=$$!; \
	    sleep 2; \
	    echo "==> Running capture against test.pcap…"; \
	    $(BIN_DIR)/netsentry-capture -r /tmp/netsentry_test.pcap; \
	    sleep 1; \
	    echo "==> Fetching alerts…"; \
	    curl -s http://localhost:8080/api/alerts | python3 -m json.tool; \
	    kill $$ENGINE_PID 2>/dev/null; \
	    echo "==> Done."

## help      — show this message
help:
	@grep -E '^##' Makefile | sed 's/## //'
