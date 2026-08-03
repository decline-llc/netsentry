# R90-64 Sustained Parser and Formatter Fuzz Evidence

## Evidence Scope

- Run ID: `20260803T091620Z`
- Run date: 2026-08-03 UTC
- Source baseline: `33bc37d9ff71932d6e4ea49cf414f3ed0008415a`
  plus the reviewed R90-64 working diff
- Command: `FUZZ_SUSTAINED_ITERATIONS=1000000 make fuzz-sustained`
- Evidence class: local synthetic deterministic mutations
- External corpus: not provided
- Production-derived traffic: no
- Release, tag, or publication authority: no

This is a bounded robustness baseline for the two versioned C fuzz harnesses.
It is not corpus provenance, realistic traffic, throughput, release-gate, or
production evidence.

## Tool Preflight

- Platform: Linux x86-64 with glibc 2.39
- Compiler: Ubuntu GCC 13.3.0
- Make: GNU Make 4.3
- Bash: GNU Bash 5.2.21
- Python: 3.12.3
- Sanitizer build: both harness targets were forcibly rebuilt from the
  repository `-fsanitize=address -fno-omit-frame-pointer -g` configuration
  immediately before execution

No extra third-party tool or external corpus was required. Paths, hostnames,
raw logs, generated JSON, and generated Markdown remain outside Git.

## Results

| Harness | Iterations | Exit status | Elapsed | Sanitizer findings |
| --- | ---: | ---: | ---: | ---: |
| C frame parser | 1,000,000 | 0 | 0.389 s | 0 |
| C UDS JSON formatter | 1,000,000 | 0 | 117.810 s | 0 |

- Overall elapsed time: 118.317 seconds.
- Both harnesses printed their exact accepted iteration count.
- The versioned evidence validator accepted the exact two-harness set,
  matching positive iteration budgets, individual pass/exit status, zero
  observed sanitizer findings, no-corpus fields, and explicit
  `local_synthetic` / non-production / non-publication classifications.
- Focused evidence regressions also proved that a missing or failed harness,
  an iteration mismatch, and a sanitizer marker cannot be accepted as passing
  evidence. A corpus fixture whose directory name required path handling was
  replayed by the parser in a separate 100-iteration integration run; neither
  generated JSON nor Markdown exposed its absolute path.

## Review Decision

Accepted for R90-64 local engineering evidence only. The result closes the
planned sustained dual-harness baseline; it does not modify or satisfy the
blocked R90-59 publication gate.
