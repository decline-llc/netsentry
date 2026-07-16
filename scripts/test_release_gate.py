#!/usr/bin/env python3
"""Regression tests for release-gate exception isolation."""

from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from scripts import pcap_evidence, release_gate
from scripts.test_pcap_evidence import metadata, write_pcap


R9004_EXCEPTION = """approver: Release Authority
approve_utc: 2026-07-15T09:05:44Z
scope_exempt: R90-04 only: public anonymized real pcap evidence may substitute internal production-derived pcap evidence
effective_increment: R90-04
revoke_condition: This exception expires when R90-04 is complete and does not apply to any later increment
status: expired
expired_utc: 2026-07-16T08:56:45Z
expired_commit: 009b2a03776987359661c4ab2776f5d04820db34
synthetic_prohibited: yes
required_controls: dedicated privacy review, provenance validation, sanitization review, sensitive metadata screening
evidence_note: Approved evidence must be anonymized public real network traffic; synthetic or generated traffic is prohibited
"""

R9005_EXCEPTION = """approver: user (explicit R90-05 synthetic corpus exception approval in conversation)
approve_utc: 2026-07-16T16:11:36Z
scope_exempt: R90-05 only: the reviewed synthetic pcap corpus may substitute the approved sanitized production-derived pcap corpus
effective_increment: R90-05
effective_version: v0.1.1
revoke_condition: This exception expires when R90-05 is complete and does not apply to R90-06 or any later increment or release
status: expired
expired_utc: 2026-07-16T16:19:54Z
expired_commit: 6c3f9ef276c99c13aa9e985b8c849bb5f0791752
synthetic_allowed: yes
required_controls: explicit synthetic provenance, zero-production-data privacy review, exact packet count and SHA-256 integrity verification, corpus-pressure validation
corpus_sha256: {digest}
packet_count: 1
evidence_note: The approved corpus is synthetic controlled test traffic, is not production-derived, and may satisfy only the R90-05 pcap evidence gate under this exception
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
    def validate_r9004(
        self,
        *,
        exception_text: str = R9004_EXCEPTION,
        **overrides: str,
    ) -> list[str]:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            exception = root / "release_exception_r9004.yaml"
            record = root / "evidence.md"
            exception.write_text(exception_text, encoding="utf-8")
            record.write_text(evidence(**overrides), encoding="utf-8")
            return release_gate.validate(record, exception)

    def validate_without_exception(
        self,
        *,
        manifest_mutator=None,
        include_manifest: bool = True,
        include_corpus: bool = True,
        **overrides: str,
    ) -> list[str]:
        production_fields = {
            "Corpus description": "reviewed sanitized internal production traffic",
            "Evidence class": "production-derived",
            "Production-derived corpus": "yes",
            "Exception applied": "none",
            "Exception increment": "none",
            "Privacy review": "approved",
            "Provenance validation": "approved",
            "Sanitization review": "approved",
            "Sensitive metadata screening": "approved",
        }
        production_fields.update(overrides)
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            record = root / "evidence.md"
            record.write_text(evidence(**production_fields), encoding="utf-8")
            corpus = root / "sample.pcap"
            write_pcap(corpus, packets=1)
            metadata_path = root / "metadata.json"
            metadata_path.write_text(
                json.dumps(metadata(approved=True)),
                encoding="utf-8",
            )
            manifest_path, _report_path = pcap_evidence.generate(
                corpus,
                metadata_path,
                root / "output",
            )
            if manifest_mutator is not None:
                manifest = pcap_evidence.load_json(manifest_path)
                manifest_mutator(manifest)
                manifest_path.write_text(
                    json.dumps(manifest),
                    encoding="utf-8",
                )
            return release_gate.validate(
                record,
                None,
                manifest_path=manifest_path if include_manifest else None,
                corpus_path=corpus if include_corpus else None,
            )

    def validate_r9005(
        self,
        *,
        manifest_mutator=None,
        include_manifest: bool = True,
        include_corpus: bool = True,
        exception_mutator=None,
        **overrides: str,
    ) -> list[str]:
        synthetic_fields = {
            "Corpus description": "reviewed synthetic traffic; not production-derived",
            "Evidence class": "synthetic",
            "Production-derived corpus": "no",
            "Exception applied": release_gate.R9005_EXCEPTION_PATH,
            "Exception increment": "R90-05",
            "Privacy review": "approved",
            "Provenance validation": "approved",
            "Sanitization review": "approved",
            "Sensitive metadata screening": "approved",
        }
        synthetic_fields.update(overrides)
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            corpus = root / "sample.pcap"
            write_pcap(corpus, packets=1)
            artifact = pcap_evidence.inventory(corpus)[0]
            manifest = {
                "task_id": "task-20260716-153438-r90-05",
                "corpus_path": "./sample.pcap",
                "sha256": artifact["sha256"],
                "file_size_bytes": artifact["bytes"],
                "packet_count": artifact["packet_count"],
                "provenance": "Synthetic controlled test traffic. NOT production-derived.",
                "sanitization_description": "No production data included.",
                "sensitive_metadata_handling": "No private identifiers included.",
                "privacy_compliance_statement": "Zero production user data present.",
                "review_approval_status": "approved_exception",
                "exemption_reference": release_gate.R9005_EXCEPTION_PATH,
            }
            if manifest_mutator is not None:
                manifest_mutator(manifest)
            manifest_path = root / "manifest.json"
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
            exception_text = R9005_EXCEPTION.format(digest=artifact["sha256"])
            exception_text = exception_text.replace("packet_count: 1", f"packet_count: {artifact['packet_count']}")
            if exception_mutator is not None:
                exception_text = exception_mutator(exception_text)
            exception_path = root / release_gate.R9005_EXCEPTION
            exception_path.write_text(exception_text, encoding="utf-8")
            record = root / "evidence.md"
            record.write_text(evidence(**synthetic_fields), encoding="utf-8")
            return release_gate.validate(
                record,
                exception_path,
                manifest_path=manifest_path if include_manifest else None,
                corpus_path=corpus if include_corpus else None,
            )

    def test_production_derived_evidence_passes_without_exception(self) -> None:
        self.assertEqual([], self.validate_without_exception())

    def test_no_exception_rejects_non_production_evidence_class(self) -> None:
        errors = self.validate_without_exception(**{"Evidence class": "public-anonymized-real"})
        self.assertIn(
            "without an exception, pcap Evidence class must be production-derived",
            errors,
        )

    def test_no_exception_rejects_substitute_corpus_description(self) -> None:
        errors = self.validate_without_exception(
            **{"Corpus description": "generated synthetic traffic"}
        )
        self.assertIn(
            "production-derived pcap Corpus description must not describe substitute traffic",
            errors,
        )

    def test_no_exception_rejects_false_production_claim(self) -> None:
        errors = self.validate_without_exception(**{"Production-derived corpus": "no"})
        self.assertIn(
            "without an exception, pcap Production-derived corpus must be yes",
            errors,
        )

    def test_no_exception_rejects_exception_reference(self) -> None:
        errors = self.validate_without_exception(
            **{"Exception applied": "docs/audit/release_exception_r9004.yaml"}
        )
        self.assertIn(
            "without an exception, pcap Exception applied must be none",
            errors,
        )

    def test_no_exception_rejects_missing_each_required_control(self) -> None:
        for label in (
            "Privacy review",
            "Provenance validation",
            "Sanitization review",
            "Sensitive metadata screening",
        ):
            with self.subTest(label=label):
                errors = self.validate_without_exception(**{label: "pending"})
                self.assertIn(
                    f"production-derived pcap {label} must be approved",
                    errors,
                )

    def test_no_exception_requires_manifest(self) -> None:
        self.assertIn(
            "production-derived release requires a PCAP evidence manifest",
            self.validate_without_exception(include_manifest=False),
        )

    def test_no_exception_requires_corpus_integrity_verification(self) -> None:
        self.assertIn(
            "production-derived release requires the reviewed PCAP corpus for integrity verification",
            self.validate_without_exception(include_corpus=False),
        )

    def test_no_exception_rejects_unapproved_manifest(self) -> None:
        def make_pending(manifest):
            manifest["review_status"] = "pending"

        errors = self.validate_without_exception(manifest_mutator=make_pending)
        self.assertIn(
            "PCAP evidence manifest: review_status must be approved",
            errors,
        )

    def test_completed_r9005_exception_cannot_be_reused(self) -> None:
        self.assertEqual(
            [release_gate.R9005_RELEASE_REJECTION],
            self.validate_r9005(),
        )

    def test_r9005_exception_rejects_false_production_claim(self) -> None:
        errors = self.validate_r9005(**{"Production-derived corpus": "yes"})
        self.assertIn("R90-05 exception pcap Production-derived corpus must be no", errors)

    def test_r9005_exception_rejects_another_increment(self) -> None:
        errors = self.validate_r9005(**{"Exception increment": "R90-06"})
        self.assertIn("R90-05 exception pcap Exception increment must be R90-05", errors)

    def test_r9005_exception_requires_manifest_and_corpus(self) -> None:
        self.assertIn(
            "R90-05 exception requires the approved synthetic manifest",
            self.validate_r9005(include_manifest=False),
        )
        self.assertIn(
            "R90-05 exception requires the reviewed synthetic corpus for integrity verification",
            self.validate_r9005(include_corpus=False),
        )

    def test_r9005_exception_rejects_tampered_manifest(self) -> None:
        def change_digest(manifest):
            manifest["sha256"] = "0" * 64

        errors = self.validate_r9005(manifest_mutator=change_digest)
        self.assertIn(
            "R90-05 evidence manifest: R90-05 manifest SHA-256 must match the approved exception",
            errors,
        )
        self.assertIn(
            "R90-05 evidence manifest: R90-05 corpus SHA-256 differs from the manifest",
            errors,
        )

    def test_r9005_exception_requires_expired_delivery_metadata(self) -> None:
        errors = self.validate_r9005(
            exception_mutator=lambda text: text.replace("status: expired", "status: active")
        )
        self.assertIn("R90-05 exception status must be expired", errors)

    def test_r9004_public_real_evidence_cannot_approve_release(self) -> None:
        self.assertEqual(
            [release_gate.R9004_RELEASE_REJECTION],
            self.validate_r9004(),
        )

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

    def test_r9004_rejects_invalid_expiry_metadata(self) -> None:
        cases = (
            (
                R9004_EXCEPTION.replace("status: expired", "status: active"),
                "R90-04 exception status must be expired",
            ),
            (
                R9004_EXCEPTION.replace(
                    "expired_utc: 2026-07-16T08:56:45Z",
                    "expired_utc: not-a-time",
                ),
                "R90-04 exception expired_utc must be ISO-8601 UTC",
            ),
            (
                R9004_EXCEPTION.replace(
                    "expired_commit: 009b2a03776987359661c4ab2776f5d04820db34",
                    "expired_commit: 009b2a0",
                ),
                "R90-04 exception expired_commit must be a full lowercase Git SHA",
            ),
        )
        for exception_text, expected in cases:
            with self.subTest(expected=expected):
                self.assertIn(
                    expected,
                    self.validate_r9004(exception_text=exception_text),
                )

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
