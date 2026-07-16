#!/usr/bin/env python3
"""Regression tests for the non-PCAP release gate."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts import release_gate


def evidence(
    *,
    metadata_decision: str = "approved",
    fuzz_status: str = "pass",
    fuzz_iterations: str = "1000000",
    fuzz_crashes: str = "0",
    asan_findings: str = "no",
    fuzz_review: str = "approved",
    pcap_section: str = "",
    credentials_present: str = "no",
    final_fuzz_review: str = "yes",
    approved_for_release: str = "yes",
) -> str:
    return f"""## Metadata
- Final decision: {metadata_decision}

## Sustained External C Fuzz Evidence
- Status: {fuzz_status}
- Iterations or duration: {fuzz_iterations}
- Crashes: {fuzz_crashes}
- ASan findings: {asan_findings}
- Reviewer decision: {fuzz_review}

{pcap_section}
## Sensitive Information Review
- Raw pcaps staged: yes
- Fuzz corpus files staged: no
- Private corpus paths present: yes
- Credentials or tokens present: {credentials_present}
- Local operator notes present: no
- Generated archives staged: no

## Final Release Gate Decision
- Sustained external fuzz evidence reviewed: {final_fuzz_review}
- Approved for release: {approved_for_release}
"""


class ReleaseGateTest(unittest.TestCase):
    def validate(self, text: str, **kwargs) -> list[str]:
        with tempfile.TemporaryDirectory() as directory:
            record = Path(directory) / "evidence.md"
            record.write_text(text, encoding="utf-8")
            return release_gate.validate(record, **kwargs)

    def test_release_passes_without_any_pcap_section(self) -> None:
        self.assertEqual([], self.validate(evidence(), exception_path=None))

    def test_all_pcap_fields_are_ignored(self) -> None:
        section = """## Realistic Sanitized Pcap Corpus Evidence
- Status: failed
- Corpus description: unrestricted
- Evidence class: unknown
- Production-derived corpus: unknown
- Exception applied: missing
- Exception increment: another increment
- Privacy review: rejected
- Provenance validation: rejected
- Sanitization review: rejected
- Sensitive metadata screening: rejected
- Pcap files: 0
- Packets processed: 0
- Query evidence: failed
- Reviewer decision: rejected

"""
        self.assertEqual(
            [],
            self.validate(
                evidence(pcap_section=section),
                exception_path=Path("missing-exception.yaml"),
                manifest_path=Path("missing-manifest.json"),
                corpus_path=Path("missing-corpus.pcap"),
            ),
        )

    def test_historical_release_records_still_pass(self) -> None:
        for name in ("release-v0.1.0.md", "release-v0.1.1.md"):
            with self.subTest(name=name):
                self.assertEqual(
                    [],
                    release_gate.validate(
                        Path("docs/evidence") / name,
                        Path("missing-exception.yaml"),
                        manifest_path=Path("missing-manifest.json"),
                        corpus_path=Path("missing-corpus.pcap"),
                    ),
                )

    def test_metadata_final_decision_remains_required(self) -> None:
        errors = self.validate(evidence(metadata_decision="pending"), exception_path=None)
        self.assertIn("Metadata/Final decision must be approved", errors)

    def test_fuzz_controls_remain_required(self) -> None:
        cases = (
            ({"fuzz_status": "failed"}, "fuzz Status must be pass"),
            ({"fuzz_iterations": "999999"}, "fuzz iterations must be at least 1000000"),
            ({"fuzz_crashes": "1"}, "fuzz Crashes must be 0"),
            ({"asan_findings": "yes"}, "fuzz ASan findings must be no/none/0"),
            ({"fuzz_review": "pending"}, "fuzz Reviewer decision must be approved"),
            ({"final_fuzz_review": "no"}, "final fuzz review must be yes"),
        )
        for overrides, expected in cases:
            with self.subTest(expected=expected):
                self.assertIn(
                    expected,
                    self.validate(evidence(**overrides), exception_path=None),
                )

    def test_final_release_approval_remains_required(self) -> None:
        errors = self.validate(evidence(approved_for_release="no"), exception_path=None)
        self.assertIn("Approved for release must be yes", errors)

    def test_sensitive_information_gate_remains_required(self) -> None:
        errors = self.validate(evidence(credentials_present="yes"), exception_path=None)
        self.assertIn("Sensitive Information Review/Credentials or tokens present must be no", errors)

    def test_private_or_credential_like_public_text_is_rejected(self) -> None:
        for forbidden in ("/home/private/corpus", "/tmp/private", "ghp_example"):
            with self.subTest(forbidden=forbidden):
                errors = self.validate(evidence() + forbidden, exception_path=None)
                self.assertTrue(
                    any("forbidden private or credential-like content" in error for error in errors)
                )


if __name__ == "__main__":
    unittest.main()
