# Contributing to NetSentry

Thank you for your interest in contributing.  This document explains how to
submit changes and what to expect from the review process.

## Development Setup

```bash
# Prerequisites: Go 1.21+, GCC 9+, libpcap-dev, make
git clone https://github.com/decline-llc/netsentry.git
cd netsentry
make build
make test
```

## Branching

- `main` — stable, always green CI
- `feat/<name>` — new features
- `fix/<name>` — bug fixes
- `chore/<name>` — maintenance (docs, deps, refactoring)

## Commit Style

Follow Conventional Commits:

```
feat(engine): add port-blacklist rule type
fix(capture): handle VLAN double-tagging correctly
chore(deps): bump modernc.org/sqlite to v1.30.0
```

## Pull Requests

1. Fork the repository and create a branch from `main`.
2. Write or update tests for the changed code.
3. Ensure `make test` and `make lint` pass locally.
4. Open a pull request with a clear description of the change and why it
   is needed.  Reference any related issue numbers.

## Code Style

- Go: `gofmt -s` formatted, `go vet` clean.
- C: `clang-format` with the project's `.clang-format` (LLVM base, 4-space
  indent, 100-column limit).
- No generated files (binaries, `*.out`, coverage reports) committed.

## Testing

All public functions must have unit tests.  New detection rules require a
corresponding test in `engine/internal/rule/engine_test.go`.

## Security Issues

Do **not** open a public GitHub issue for security vulnerabilities.
See [SECURITY.md](SECURITY.md) for the responsible disclosure process.

## License

By submitting a pull request you agree that your contribution will be
licensed under the [MIT License](LICENSE).
