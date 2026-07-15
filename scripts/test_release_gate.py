#!/usr/bin/env python3
"""Regression tests for release-gate exception isolation."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts import release_gate


R9004_EXCEPTION = """approver: Release Authority
approve_utc: 2026-07-15T09:05:44Z
scope_exempt: R90-04 only: public anonymized real pcap evidence may substitute internal production-derived pcap evidence
effective_increment: R90-04
revoke_condition: This exception expires when R90-04 is complete and does not apply to any later increment
synthetic_prohibited: yes
required_controls: dedicated privacy review, provenance validation, sanitization review, sensitive metadata screening
evidence_note: Approved evidence must be anonymized public real network traffic; synthetic or generated traffic is prohibited
"""


def evidence(**overrides: str) -> str:
    fields = {
        "Corpus description": "anonymized public real network traffic dataset",
        "Evidence class": "public-anonymized-real",
        "Production-derived corpus": "no",
        "Exception applied": "docs/audit/release_exception_r9004.yaml",
        "Exception increment": "R90-04",
        "Privacy review": "approved",
        "Provenance validation": "approved",
        "Sanitization review": "approved",
        "Sensitive metadata screening": "approved",
    }
    fields.update(overrides)
    pcap_fields = "\n".join(f"- {key}: {value}" for key, value in fields.items())
    return f"""## Metadata
- Final decision: approved

## Sustained External C Fuzz Evidence
- Status: pass
- Iterations or duration: 1000000
- Crashes: 0
- ASan findings: no
- Reviewer decision: approved

## Realistic Sanitized Pcap Corpus Evidence
- Status: pass
{pcap_fields}
- Pcap files: 1
- Packets processed: 1
- Query evidence: pass
- Reviewer decision: approved

## Sensitive Information Review
- Raw pcaps staged: no
- Fuzz corpus files staged: no
- Private corpus paths present: no
- Credentials or tokens present: no
- Local operator notes present: no
- Generated archives staged: no

## Final Release Gate Decision
- Sustained external fuzz evidence reviewed: yes
- Realistic sanitized pcap corpus evidence reviewed: yes
- Approved for release: yes
"""


class ReleaseGateTest(unittest.TestCase):
    def validate_r9004(self, **overrides: str) -> list[str]:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            exception = root / "release_exception_r9004.yaml"
            record = root / "evidence.md"
            exception.write_text(R9004_EXCEPTION, encoding="utf-8")
            record.write_text(evidence(**overrides), encoding="utf-8")
            return release_gate.validate(record, exception)

    def test_r9004_public_real_evidence_is_accepted(self) -> None:
        self.assertEqual([], self.validate_r9004())

    def test_r9004_rejects_synthetic_evidence(self) -> None:
        errors = self.validate_r9004(**{"Evidence class": "synthetic"})
        self.assertIn("R90-04 exception pcap Evidence class must be public-anonymized-real", errors)

    def test_r9004_rejects_generated_corpus_description(self) -> None:
        errors = self.validate_r9004(**{"Corpus description": "generated synthetic traffic"})
        self.assertIn(
            "R90-04 exception pcap Corpus description must not describe synthetic or generated traffic",
            errors,
        )

    def test_r9004_rejects_another_increment(self) -> None:
        errors = self.validate_r9004(**{"Exception increment": "R90-05"})
        self.assertIn("R90-04 exception pcap Exception increment must be R90-04", errors)

    def test_r9004_rejects_missing_each_required_control(self) -> None:
        for label in (
            "Privacy review",
            "Provenance validation",
            "Sanitization review",
            "Sensitive metadata screening",
        ):
            with self.subTest(label=label):
                errors = self.validate_r9004(**{label: "pending"})
                self.assertIn(f"R90-04 exception pcap {label} must be approved", errors)

    def test_v010_exception_remains_valid(self) -> None:
        self.assertEqual(
            [],
            release_gate.validate(
                Path("docs/evidence/release-v0.1.0.md"),
                Path("docs/audit/release_exception_v0.1.0.yaml"),
            ),
        )


if __name__ == "__main__":
    unittest.main()
