# Supply-Chain Policy

NetSentry treats CI definitions, build tools, Go toolchains, dependencies, and external pcap fixtures as executable or integrity-sensitive inputs. `.github/supply-chain-lock.json` is the reviewed lock that connects those inputs to immutable upstream evidence.

## Action policy

Every external `uses:` entry must use a full 40-character commit SHA. The adjacent comment records the exact release tag reviewed for humans; the comment is not the trust anchor. `scripts/check_supply_chain.py` rejects mutable tags, unknown Actions, lock/workflow drift, unused lock entries, and Actions whose reviewed runtime is not Node 24.

For an Action update:

1. Read the upstream release and `action.yml` from the Action's official repository.
2. Confirm the runtime and hosted-runner minimum.
3. Resolve the exact release tag, including an annotated tag, to its commit object.
4. Update the workflow SHA, readable version comment, and lock entry together.
5. Run `make workflow-check` and `make supply-chain-check` before review.

## Go and security tools

`engine/go.mod` separates the module language baseline from the CI compiler:

- `go 1.22.2` preserves the current language/module semantics.
- `toolchain go1.25.12` pins a supported n-1 Go patch release reviewed from `https://go.dev/dl/?mode=json`.

`actions/setup-go` reads the toolchain directive. The supply-chain checker also runs `go env GOVERSION` inside `engine/` and rejects a runtime that differs from the lock. CI installs `govulncheck` and `actionlint` from exact Go module versions; their upstream release commits are recorded in the lock. `govulncheck ./...` must report zero reachable vulnerabilities.

## External fixture and license integrity

Pcap bytes and license copies remain outside the source repository. The tracked `testdata/external-pcaps/manifest.json` records nine immutable URLs, upstream commits, byte counts, SHA-256 values, purposes, and license relationships for PcapPlusPlus and Zeek.

With `SUPPLY_CHAIN_FETCH_ASSETS=1`, CI downloads each entry only to a temporary directory, rejects size/hash drift, and deletes the directory at process exit. The checker also rejects unpinned URLs, unsafe relative paths, missing license entries, source/license mismatches, network-replay permission, and manifest/lock count drift. It never opens, executes, or replays a pcap.

## Commands

```bash
# Fast offline lock/policy and vulnerability verification
make supply-chain-check

# CI-equivalent remote fixture/license verification
SUPPLY_CHAIN_FETCH_ASSETS=1 make supply-chain-check

# Existing full local integration using sibling fixture bytes
make test-integration
```

The external lock proves provenance and byte integrity, not that traffic is safe or production-representative. Production-derived evidence still requires authorization, sanitization, and human review.
