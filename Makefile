# NetSentry root Makefile

SHELL      := /bin/bash
GO         := go
GOVULNCHECK ?= govulncheck
ACTIONLINT  ?= actionlint
GOPROXY    ?= https://goproxy.cn,direct
GOCACHE    ?= /tmp/netsentry-go-cache
GO_MODULE  := ./engine
BIN_DIR    := bin
BENCH_ITERATIONS ?= 100000
COVERPROFILE ?= /tmp/netsentry-coverage.out
VERSION    ?= 0.1.0-dev
IMAGE      ?= netsentry:$(VERSION)
DOCKER     ?= docker

.PHONY: all build-c build-go build build-asan test test-unit test-integration test-e2e test-stress test-coverage deps-check supply-chain-check workflow-check docs-check shell-check python-check evidence-check knowledge-check config-check asan-test bench fuzz-parser fuzz-parser-long fuzz-sustained gen-sanitized-corpus e2e-smoke e2e-pressure e2e-corpus-pressure sanitize-pcap pcap-evidence pcap-evidence-check dist release-artifacts docker-build release-gate rc-check lint clean quickstart help

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

## test-unit — run unit/race tests plus C AddressSanitizer tests serially
test-unit: test asan-test

## test-integration — verify and process the external pcap fixture corpus
test-integration: build
	@NETSENTRY_TEST_ASSETS="$(if $(NETSENTRY_TEST_ASSETS),$(NETSENTRY_TEST_ASSETS),$(abspath ../NetSentry_TestAssets))" bash scripts/test_external_corpus.sh

## test-e2e — run the deterministic pcap-to-API system test
test-e2e: e2e-smoke

## test-stress — run configurable end-to-end packet pressure (STRESS_REPEATS defaults to 1000)
test-stress: build
	@PRESSURE_REPEATS="$(if $(STRESS_REPEATS),$(STRESS_REPEATS),1000)" bash scripts/e2e_pressure.sh

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

## workflow-check — validate GitHub Actions syntax with the pinned actionlint tool
workflow-check:
	@command -v $(ACTIONLINT) >/dev/null 2>&1 || { echo "$(ACTIONLINT) not found; install the version in .github/supply-chain-lock.json"; exit 1; }
	@$(ACTIONLINT) .github/workflows/ci.yml .github/workflows/release.yaml .github/workflows/docker-publish.yml

## supply-chain-check — validate immutable Actions, Go policy, external locks, and reachable vulnerabilities
supply-chain-check: workflow-check
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || { echo "$(GOVULNCHECK) not found; install the version in .github/supply-chain-lock.json"; exit 1; }
	@python3 scripts/check_supply_chain.py $(if $(filter 1,$(SUPPLY_CHAIN_FETCH_ASSETS)),--fetch-assets,)
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GOVULNCHECK) ./...

## docs-check — scan public docs for retired stale wording
docs-check:
	@bash scripts/docs_check.sh

## shell-check — run shell script syntax checks
shell-check:
	@bash -n scripts/e2e_smoke.sh
	@bash -n scripts/e2e_pressure.sh
	@bash -n scripts/e2e_corpus_pressure.sh
	@bash -n scripts/test_external_corpus.sh
	@bash -n scripts/fuzz_sustained.sh
	@bash -n scripts/docs_check.sh
	@bash -n scripts/package_release.sh
	@bash -n scripts/rc_check.sh

## python-check — run Python script syntax checks
python-check:
	@python3 -c 'import ast, pathlib; [ast.parse(path.read_text(), filename=str(path)) for path in map(pathlib.Path, ("scripts/check_supply_chain.py", "scripts/gen_test_pcap.py", "scripts/gen_sanitized_corpus.py", "scripts/sanitize_pcap.py", "scripts/test_sanitize_pcap.py", "scripts/pcap_evidence.py", "scripts/test_pcap_evidence.py", "scripts/release_gate.py", "scripts/test_release_gate.py", "scripts/sync_knowledge.py", "scripts/post_push_sync.py", "scripts/test_sync_knowledge.py", "scripts/test_post_push_sync.py", "scripts/fixtures/__init__.py", "scripts/fixtures/post_push_fixture.py"))]'

## evidence-check — run sanitizer, PCAP evidence, and release-gate regressions
evidence-check:
	@python3 -m unittest scripts.test_sanitize_pcap scripts.test_pcap_evidence scripts.test_release_gate

## knowledge-check — test deterministic, idempotent Obsidian knowledge extraction
knowledge-check:
	@python3 -m unittest scripts/test_sync_knowledge.py scripts/test_post_push_sync.py scripts/test_post_push_sync.py

## gen-sanitized-corpus — generate deterministic synthetic pcap corpus outside the repository
gen-sanitized-corpus:
	@python3 scripts/gen_sanitized_corpus.py --output-dir "$(if $(CORPUS_DIR),$(CORPUS_DIR),/tmp/netsentry-sanitized-corpus)" --sets "$(if $(CORPUS_SETS),$(CORPUS_SETS),1)"

## config-check — validate repository config, rule, and suppression files
config-check:
	@mkdir -p $(GOCACHE)
	cd $(GO_MODULE) && GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test -count=1 \
	    ./internal/config ./internal/rule ./internal/alert \
	    -run 'TestLoadRepositoryConfigFile|TestRepositoryRuleFilesUseCanonicalWrappedSchema|TestRepositorySuppressionsFileUsesCanonicalWrappedSchema'

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

## fuzz-sustained — run sustained ASan parser fuzz evidence; optional FUZZ_CORPUS=/path
fuzz-sustained:
	@bash scripts/fuzz_sustained.sh

## e2e-smoke — run deterministic pcap -> UDS -> engine -> SQLite -> API smoke test
e2e-smoke: build
	@bash scripts/e2e_smoke.sh

## e2e-pressure — run repeat-pcap end-to-end throughput smoke test
e2e-pressure: build
	@bash scripts/e2e_pressure.sh

## e2e-corpus-pressure — run local pcap corpus pressure evidence: PCAP_CORPUS=/path make e2e-corpus-pressure
e2e-corpus-pressure:
	@test -n "$(PCAP_CORPUS)" || { echo "PCAP_CORPUS is required"; exit 2; }
	$(MAKE) build
	@bash scripts/e2e_corpus_pressure.sh

## sanitize-pcap — write a sanitized pcap: make sanitize-pcap INPUT=in.pcap OUTPUT=out.pcap
sanitize-pcap:
	@test -n "$(INPUT)" || { echo "INPUT is required"; exit 1; }
	@test -n "$(OUTPUT)" || { echo "OUTPUT is required"; exit 1; }
	@python3 scripts/sanitize_pcap.py -i "$(INPUT)" -o "$(OUTPUT)"

## pcap-evidence — generate path-redacted local manifest/report from reviewed metadata
pcap-evidence:
	@test -n "$(PCAP_CORPUS)" || { echo "PCAP_CORPUS is required"; exit 2; }
	@test -n "$(PCAP_METADATA)" || { echo "PCAP_METADATA is required"; exit 2; }
	@python3 scripts/pcap_evidence.py generate \
	    --pcap "$(PCAP_CORPUS)" \
	    --metadata "$(PCAP_METADATA)" \
	    --output-dir "$(if $(PCAP_EVIDENCE_OUTPUT),$(PCAP_EVIDENCE_OUTPUT),docs/evidence/local/pcap-evidence)"

## pcap-evidence-check — verify approved manifest against the exact local corpus
pcap-evidence-check:
	@test -n "$(PCAP_CORPUS)" || { echo "PCAP_CORPUS is required"; exit 2; }
	@test -n "$(PCAP_EVIDENCE_MANIFEST)" || { echo "PCAP_EVIDENCE_MANIFEST is required"; exit 2; }
	@python3 scripts/pcap_evidence.py validate \
	    --manifest "$(PCAP_EVIDENCE_MANIFEST)" \
	    --pcap "$(PCAP_CORPUS)" \
	    --require-approved

## dist      — build a local release archive under dist/
dist: build
	@bash scripts/package_release.sh $(VERSION)

## release-artifacts — validate a SemVer VERSION and build release archive assets
release-artifacts:
	@[[ "$(VERSION)" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$$ ]] || { echo "VERSION must be SemVer without leading v"; exit 1; }
	$(MAKE) dist VERSION="$(VERSION)"

## release-gate — require reviewed public fuzz and pcap evidence before release
release-gate:
	@python3 scripts/release_gate.py --evidence "$(if $(RELEASE_EVIDENCE),$(RELEASE_EVIDENCE),docs/evidence/release-v0.1.0.md)" $(if $(filter none,$(RELEASE_EXCEPTION)),--no-exception,--exception "$(if $(RELEASE_EXCEPTION),$(RELEASE_EXCEPTION),docs/audit/release_exception_v0.1.0.yaml)") $(if $(PCAP_EVIDENCE_MANIFEST),--pcap-manifest "$(PCAP_EVIDENCE_MANIFEST)") $(if $(PCAP_CORPUS),--pcap-corpus "$(PCAP_CORPUS)")

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
