.PHONY: all quickstart clean build-c build-go test

# Default target
all: build-c build-go

# The "5-minute rule" entry point
quickstart:
	@echo "=== NetSentry Quickstart ==="
	@echo "[1/5] Building C capture module..."
	# TODO: Add gcc build commands here
	@echo "[2/5] Building Go engine..."
	# TODO: Add go build commands here
	@echo "[3/5] Generating test pcap..."
	# TODO: Add python scapy script generation
	@echo "[4/5] Starting services..."
	# TODO: Add startup commands
	@echo "[5/5] Verifying..."
	@echo "Ready! Try: curl http://localhost:8080/api/health"

build-c:
	@echo "Building C module..."
	# cd capture && make

build-go:
	@echo "Building Go module..."
	# go build -o bin/netsentry ./engine/cmd

clean:
	rm -rf bin/
	rm -f data/*.db data/*.sock
