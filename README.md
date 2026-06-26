# netsentry
Production-grade lightweight IDS engine for pcap offline analysis &amp; MITRE ATT&amp;CK mapping. Powered by C/Go.
NetSentry
Status: v0.1.0 In Development (Pre-alpha)
A production-grade lightweight IDS engine focused on offline pcap analysis and edge network threat detection.
✨ Why NetSentry?
- Fast Forensics: One-click docker deployment, analyze pcap files and output MITRE ATT&CK alerts within 10 seconds.
- Zero Dependencies: No database required, no complicated rule management, supports single-binary deployment.
- Lightweight: ~50MB memory footprint, compatible with edge devices such as Raspberry Pi and network gateways.
- Built-in MITRE ATT&CK: All alert logs carry standard mitre_tactic and technique_id mapping by default.
⚡ Quick Start
The project is currently in active development. Build from source for testing.
Prerequisites
- Go 1.21+
- GCC 9.0+
- libpcap-dev
make quickstart
📁 Project Structure
- capture/: C-based packet capture & parsing module based on libpcap.
- engine/: Go-based core engine, including rule matching, API service and data storage.
- configs/: Default detection rules and configuration templates.
📄 Documentation
See Design Document for detailed architecture and technical specifications.
📜 License
Released under the MIT License. SeeLICENSE file for full terms.
