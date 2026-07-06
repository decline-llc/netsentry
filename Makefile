# NetSentry root Makefile

SHELL      := /bin/bash
GO         := go
GOPROXY    ?= https://goproxy.cn,direct
GOCACHE    ?= /tmp/netsentry-go-cache
GO_MODULE  := ./engine
BIN_DIR    := bin
BENCH_ITERATIONS ?= 100000
COVERPROFILE ?= /tmp/netsentry-coverage.out
VERSION    ?= 0.1.0-dev
IMAGE      ?= netsentry:$(VERSION)
DOCKER     ?= docker

.PHONY: all build-c build-go build build-asan test test-coverage deps-check docs-check shell-check asan-test bench fuzz-parser fuzz-parser-long e2e-smoke e2e-pressure sanitize-pcap dist docker-build rc-check lint clean quickstart help

all: build

## build-c   — compile the C capture binary
build-c:
	$(MAKE) -C capture

## build-go  — compile the Go engine binary
build-go:
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) build -o ../$(BIN_DIR)/netsentry-engine \
	    ./cmd/netsentry

## build     — compile both C and Go binaries
build: build-c build-go

## build-asan — compile the C capture binary with AddressSanitizer
build-asan:
	$(MAKE) -C capture build-asan

## test      — run C parser tests and Go unit tests with race detector
test:
	$(MAKE) -C capture test
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test -race -count=1 ./...

## test-coverage — run C tests and Go coverage summary
test-coverage:
	$(MAKE) -C capture test
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test -count=1 -covermode=atomic -coverprofile=$(COVERPROFILE) ./...
	cd $(GO_MODULE) && $(GO) tool cover -func=$(COVERPROFILE) | tail -n 1

## deps-check — verify Go module dependency cache integrity
deps-check:
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) mod verify

## docs-check — scan public docs for retired stale wording
docs-check:
	@bash scripts/docs_check.sh

## shell-check — run shell script syntax checks
shell-check:
	@bash -n scripts/e2e_smoke.sh
	@bash -n scripts/e2e_pressure.sh
	@bash -n scripts/docs_check.sh
	@bash -n scripts/package_release.sh
	@bash -n scripts/rc_check.sh

## asan-test — run C parser tests with AddressSanitizer
asan-test:
	$(MAKE) -C capture asan-test

## bench     — run C parser microbenchmarks and Go benchmarks
bench:
	$(MAKE) -C capture bench BENCH_ITERATIONS=$(BENCH_ITERATIONS)
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test -bench=. -benchtime=10s -benchmem ./...

## fuzz-parser — run deterministic ASan fuzz smoke for the C frame parser
fuzz-parser:
	$(MAKE) -C capture fuzz-parser

## fuzz-parser-long — run a longer deterministic ASan fuzz pass for the C frame parser
fuzz-parser-long:
	$(MAKE) -C capture fuzz-parser-long

## e2e-smoke — run deterministic pcap -> UDS -> engine -> SQLite -> API smoke test
e2e-smoke: build
	@bash scripts/e2e_smoke.sh

## e2e-pressure — run repeat-pcap end-to-end throughput smoke test
e2e-pressure: build
	@bash scripts/e2e_pressure.sh

## sanitize-pcap — write a sanitized pcap: make sanitize-pcap INPUT=in.pcap OUTPUT=out.pcap
sanitize-pcap:
	@test -n "$(INPUT)" || { echo "INPUT is required"; exit 1; }
	@test -n "$(OUTPUT)" || { echo "OUTPUT is required"; exit 1; }
	@python3 scripts/sanitize_pcap.py -i "$(INPUT)" -o "$(OUTPUT)"

## dist      — build a local release archive under dist/
dist: build
	@bash scripts/package_release.sh $(VERSION)

## docker-build — build a local Docker image
docker-build:
	@command -v $(firstword $(DOCKER)) >/dev/null 2>&1 || { echo "$(firstword $(DOCKER)) not found"; exit 1; }
	$(DOCKER) build -t $(IMAGE) .

## rc-check   — run local release-candidate checks, including Docker unless SKIP_DOCKER=1
rc-check:
	@bash scripts/rc_check.sh

## lint      — run go vet and staticcheck (if installed)
lint:
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) $(GO) vet ./...
	@command -v staticcheck >/dev/null 2>&1 && \
	    cd $(GO_MODULE) && GOCACHE=$(GOCACHE) staticcheck ./... || \
	    echo "[lint] staticcheck not found, skipping"

## clean     — remove compiled binaries and object files
clean:
	$(MAKE) -C capture clean
	rm -rf $(BIN_DIR) dist

## quickstart — generate a sample pcap and run a full analysis
quickstart: build
	@echo "==> Generating sample pcap…"
	@command -v python3 >/dev/null 2>&1 || { echo "python3 not found"; exit 1; }
	@python3 scripts/gen_test_pcap.py
	@echo "==> Starting engine in background…"
	@mkdir -p data
	@rm -f data/netsentry.db data/netsentry.db-shm data/netsentry.db-wal
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
