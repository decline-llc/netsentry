# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes (current development) |

## Reporting a Vulnerability

**Please do not file public GitHub issues for security vulnerabilities.**

Email: **abcdefg.llc@outlook.com**

Include in your report:
- A description of the vulnerability and its potential impact
- Steps to reproduce or a minimal proof-of-concept
- Any suggested mitigations you may have already identified

We aim to acknowledge reports within **48 hours** and provide an initial
assessment within **7 days**.

## Security Design Notes

- The REST API supports Pre-Shared Key (PSK) authentication
  (`api_auth.enabled: true` in `config.yaml`).  Tokens are read from
  environment variables — never hard-code tokens in config files.
- The engine runs as a non-root user.  The C capture module requires
  `CAP_NET_RAW` (or root) for live capture only; offline pcap analysis
  needs no elevated privileges.
- All SQL queries use parameterised statements; no raw string interpolation.
- The UDS socket (`/tmp/netsentry.sock`) should be owned by the engine
  user and have mode `0600`.

## Known Limitations (v0.1.0)

- No TLS on the HTTP API.  Use a reverse proxy (nginx/caddy) with TLS
  termination in production.
- No IPv6 support.
- No TCP stream reassembly; evasion via fragmented payloads is possible.
