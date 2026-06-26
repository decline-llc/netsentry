# NetSentry

> **Status**: v0.1.0 — In Development (Pre-alpha)

A production-grade lightweight IDS engine focused on offline pcap forensics and edge network threat detection. Powered by C/Go.

---

## Why NetSentry?

- **Fast Forensics**: One-line Docker command, analyze pcap files and see MITRE ATT&CK alerts in 10 seconds.
- **Zero Dependencies**: No database setup, no rule management overhead — single-binary deployment.
- **Lightweight**: ~50 MB memory footprint, runs on Raspberry Pi and edge gateways.
- **Built-in MITRE ATT&CK**: Every alert carries `mitre_tactic` and `mitre_technique_id` out of the box.
- **Production Quality**: AAA test coverage, Prometheus metrics, structured logging, PSK API auth, crash recovery.

---

## Quick Start

### Docker (recommended)

```bash
docker run --rm -v $(pwd)/suspicious.pcap:/data/input.pcap:ro \
    ghcr.io/yourusername/netsentry:v0.1.0 \
    --pcap /data/input.pcap --output table
```

Sample output:

```
┌──────────┬──────────┬──────────────┬─────────────────────┬────────────────────┐
│ SEVERITY │ RULE ID  │ SRC IP       │ MITRE TECHNIQUE     │ PAYLOAD PREVIEW    │
├──────────┼──────────┼──────────────┼─────────────────────┼────────────────────┤
│ HIGH     │ rule-001 │ 10.0.0.99    │ T1190 (Exploit...)  │ GET /search?q=UN...│
│ CRITICAL │ rule-010 │ 192.168.1.50 │ T1571 (Non-Std Po...│ [IP BLACKLIST]     │
└──────────┴──────────┴──────────────┴─────────────────────┴────────────────────┘
Total: 2 alerts | Pcap: 15234 packets | Time: 2.3s
```

### Build from Source

**Prerequisites**: Go 1.21+, GCC 9.0+, libpcap-dev, make, Python 3 + Scapy

```bash
# Ubuntu/Debian
sudo apt install -y build-essential gcc make libpcap-dev golang-go python3-pip
pip3 install scapy

git clone https://github.com/yourusername/netsentry.git
cd netsentry
make quickstart
```

### REST API Mode

```bash
make quickstart   # starts engine + capture, API on :8080
curl http://localhost:8080/api/alerts | jq .
curl http://localhost:8080/api/health?verbose=true | jq .
curl http://localhost:8080/api/metrics
```

---

## Performance Boundaries

| Item | Value |
|------|-------|
| Throughput (offline, single core) | ~50K PPS |
| Memory (idle / full load) | ~30–50 MB / ~80 MB |
| IPv6 | Not supported (v0.1.0) |
| TCP stream reassembly | Not supported (v0.3.0 roadmap) |

NetSentry targets edge devices, pcap forensics, and IDS learning — not 10 Gbps line-rate. For enterprise-scale deployments use [Snort3](https://github.com/snort3/snort3) or [Suricata](https://github.com/OISF/suricata).

---

## Detection Capabilities (v0.1.0)

| Rule Type | Example |
|-----------|---------|
| `payload_match` | SQL injection, XSS, Log4Shell, path traversal, reverse shells |
| `ip_blacklist` | Known C2 infrastructure (CIDR supported) |
| `port_blacklist` | Non-standard ports (e.g. 4444, 31337) |

**Known limitations**: no TCP stream reassembly (split-payload attacks undetectable), no TLS decryption, IPv4 only. See [docs/architecture.md](docs/architecture.md) for the full detection boundary table.

---

## Project Structure

```
capture/    C module — libpcap packet capture + protocol parsing
engine/     Go module — rule engine, HTTP API, alert storage
configs/    Detection rules and config templates
docs/       Architecture, API reference, development guide
testdata/   Sample pcap files (sanitized)
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System architecture, data flow, concurrency design |
| [docs/api-reference.md](docs/api-reference.md) | REST API endpoints, request/response schemas |
| [docs/development.md](docs/development.md) | Build guide, testing strategy, CI/CD, milestones |

---

## Competitive Comparison

| | NetSentry v0.1.0 | Zeek | Suricata |
|--|:--:|:--:|:--:|
| Target | Lightweight pcap forensics / edge IDS | Enterprise NSM | Enterprise IDS/IPS |
| Memory | ~50 MB | ~300 MB | ~500 MB+ |
| Startup | `make quickstart` | `zeek -i eth0` | `suricata -i eth0` |
| MITRE ATT&CK | Built-in | Plugin | Plugin |
| TCP reassembly | v0.3.0 roadmap | Yes | Yes |
| Rule language | JSON | Zeek script | Snort rules (ET Open 30K+) |
| Docker image | < 20 MB | Large | Large |

---

## License

Released under the [MIT License](LICENSE).
