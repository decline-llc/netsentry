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
- `toolchain go1.26.8` pins the latest reviewed patch in the selected supported
  Go 1.26 execution line. The 2026-10-01 review used the
  [official release policy/history](https://go.dev/doc/devel/release) and
  [download metadata](https://go.dev/dl/?mode=json): the supported lines are
  1.27 and 1.26, with latest patches 1.27.1 and 1.26.8. Selecting 1.26 limits
  the version jump from main's prior 1.25.14 pin.
- The lock records the official Linux amd64 archive identity: 66,897,291 bytes,
  SHA-256 `d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b`.
  This is upstream metadata; this increment did not download/rehash the archive.

R90-131 updates current-main metadata with static review. Behavioral, native/RC,
workflow execution, vulnerability scanning and knowledge suites remain **not run;
delegated by user**. R90-59's successful Go 1.26.8 candidate validation does not
establish current-main compatibility or zero reachable findings. The language
baseline, dependency versions, security tools and workflow sources retain their
prior definitions.

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
