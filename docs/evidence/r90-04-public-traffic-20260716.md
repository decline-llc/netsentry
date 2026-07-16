# R90-04 Public Real-Traffic Evidence

> This is an R90-04 validation record, not a release approval. It does not
> satisfy the production-derived requirements for R90-05 or later work.

## Source and Provenance

- Archive: [MAWI Working Group Traffic Archive](https://mawi.wide.ad.jp/mawi/), maintained by the WIDE Project.
- Trace: samplepoint-B `200012281400.dump`, captured from 2000-12-28 14:00:00 to 14:12:06 UTC.
- Source record: [trace metadata](https://mawi.wide.ad.jp/mawi/samplepoint-B/2000/200012281400.html).
- Download: `https://mawi.nezu.wide.ad.jp/mawi/samplepoint-B/2000/200012281400.dump.gz`.
- Download SHA-256: `8e27202b321bde5c0d782520b16a9d69f198652f3263018293cd07f6f1d2fe80`.
- Provenance validation: approved — the archive identifies the trace, time window, capture format, size, packet count, and public download; MAWI identifies samplepoint-B as a WIDE backbone trace.
- Evidence class: public-anonymized-real.
- Production-derived corpus: no.
- Exception applied: `docs/audit/release_exception_r9004.yaml`.
- Exception increment: R90-04.

## Privacy and Sanitization Review

- Privacy review: approved — MAWI's [privacy guideline](https://mawi.wide.ad.jp/mawi/guideline.txt) specifies removal of TCP/UDP payloads and scrambling of addresses; the archive states that trace addresses are scrambled.
- Sanitization review: approved — NetSentry re-sanitized the local download with `python3 scripts/sanitize_pcap.py`; 544,525 IPv4 frames were sanitized and 1,453,535 unsupported frames were zeroed. Raw and sanitized PCAPs remain outside Git.
- Sensitive metadata screening: approved — the committed record contains no PCAP, corpus path, packet payload, local hostname, credential, token, or operator note. The source has a 96-byte capture limit and the archive policy removes user payloads; NetSentry's sanitizer replaced Ethernet, IPv4, and transport-payload identifiers before replay.
- Sanitized output SHA-256: `0115c8d0e280f4a45e88394d1401d94e43aeaca724a17051854c633fa5ebdd06`.

## Corpus-Pressure Validation

- Date: 2026-07-16.
- Command shape: `PCAP_CORPUS=<sanitized-local-pcap> make e2e-corpus-pressure`.
- Status: passed.
- Pcap files: 1.
- Packets processed: 544,525.
- Capture parse errors: 0.
- Dropped packets: 0.
- UDS write errors: 0.
- Alerts generated: 0.
- Aggregated alert rows: 0.
- Query evidence: passed — API query completed with an empty, paginated result (no matching alert rules fired).
- Engine error-log lines: 0.
- Elapsed seconds: 3.537.
- Packet rate: 153,954 packets/second.
- Reviewer decision: approved for R90-04 only.

## Sensitive Information Review

- Raw PCAPs staged: no.
- Sanitized PCAPs staged: no.
- Corpus paths included: no.
- Packet payloads included: no.
- Credentials or tokens present: no.
- Local operator notes present: no.

## Boundary

The R90-04 exception remains limited to public anonymized real traffic. This
record neither approves a release nor changes the requirements of R90-05,
R90-06, or any later increment.
