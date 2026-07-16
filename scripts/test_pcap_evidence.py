#!/usr/bin/env python3
"""Regression tests for production-derived PCAP evidence manifests."""

from __future__ import annotations

import json
from pathlib import Path
import struct
import tempfile
import unittest

from scripts import pcap_evidence


def write_pcap(path: Path, packets: int = 2) -> None:
    with path.open("wb") as handle:
        handle.write(struct.pack("<IHHIIII", 0xA1B2C3D4, 2, 4, 0, 0, 65535, 1))
        for index in range(packets):
            payload = bytes([index + 1]) * 8
            handle.write(struct.pack("<IIII", 1, index, len(payload), len(payload)))
            handle.write(payload)


def write_pcapng(path: Path) -> None:
    section_body = b"\x4d\x3c\x2b\x1a" + struct.pack("<HHq", 1, 0, -1)
    section_length = 12 + len(section_body)
    enhanced_body = struct.pack("<IIIII", 0, 0, 0, 4, 4) + b"data"
    enhanced_length = 12 + len(enhanced_body)
    with path.open("wb") as handle:
        handle.write(struct.pack("<II", 0x0A0D0D0A, section_length))
        handle.write(section_body)
        handle.write(struct.pack("<I", section_length))
        handle.write(struct.pack("<II", 0x00000006, enhanced_length))
        handle.write(enhanced_body)
        handle.write(struct.pack("<I", enhanced_length))


def metadata(*, approved: bool = False) -> dict:
    review_status = "approved" if approved else "pending"
    reviews = {}
    for name in pcap_evidence.REVIEW_NAMES:
        reviews[name] = {
            "status": review_status,
            "reviewer": "Release Reviewer" if approved else "",
            "reviewed_utc": "2026-07-16T15:00:00Z" if approved else "",
            "notes": f"Reviewed {name} controls",
        }
    return {
        "provenance": {
            "source_description": "Reviewed production traffic mirror",
            "collection_node": "approved-node-alias",
            "collection_start_utc": "2026-07-16T14:00:00Z",
            "collection_end_utc": "2026-07-16T14:05:00Z",
            "custodian_reference": "internal-review-42",
        },
        "sanitization": {
            "tool": "approved-sanitizer 1.0",
            "rules": ["map addresses", "remove application identifiers"],
            "raw_capture_storage": "internal-only",
        },
        "sensitive_metadata_handling": {
            "removed_fields": ["MAC", "IP", "DNS", "SNI", "HTTP URI"],
            "residual_risk": "Reviewed low residual re-identification risk",
        },
        "reviews": reviews,
        "review_status": review_status,
        "approval": (
            {
                "reviewer": "Release Reviewer",
                "reviewed_utc": "2026-07-16T15:00:00Z",
                "decision": "approved",
            }
            if approved
            else None
        ),
    }


class PcapEvidenceTests(unittest.TestCase):
    def test_counts_classic_pcap_and_pcapng_packets(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            pcap = root / "sample.pcap"
            pcapng = root / "sample.pcapng"
            write_pcap(pcap, packets=2)
            write_pcapng(pcapng)

            self.assertEqual(2, pcap_evidence.packet_count(pcap))
            self.assertEqual(1, pcap_evidence.packet_count(pcapng))

    def test_generate_redacts_path_and_records_integrity(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            corpus = root / "private-corpus"
            output = root / "output"
            corpus.mkdir()
            write_pcap(corpus / "sample.pcap", packets=2)
            metadata_path = root / "metadata.json"
            metadata_path.write_text(json.dumps(metadata()), encoding="utf-8")

            manifest_path, report_path = pcap_evidence.generate(corpus, metadata_path, output)
            manifest = pcap_evidence.load_json(manifest_path)

            self.assertTrue(manifest["path_redacted"])
            self.assertEqual(1, manifest["file_count"])
            self.assertEqual(2, manifest["packet_count"])
            self.assertNotIn(str(root), manifest_path.read_text(encoding="utf-8"))
            self.assertNotIn(str(root), report_path.read_text(encoding="utf-8"))
            self.assertEqual([], pcap_evidence.validate_manifest_data(manifest, corpus=corpus))

    def test_approved_manifest_requires_all_human_reviews(self) -> None:
        manifest = {
            "schema_version": 1,
            "generated_utc": "2026-07-16T15:00:00Z",
            "evidence_class": "production-derived",
            "path_redacted": True,
            "file_count": 1,
            "packet_count": 1,
            "artifacts": [
                {
                    "name": "sample.pcap",
                    "bytes": 1,
                    "packet_count": 1,
                    "sha256": "0" * 64,
                }
            ],
            **metadata(approved=True),
        }
        manifest["reviews"]["privacy"]["status"] = "pending"

        errors = pcap_evidence.validate_manifest_data(manifest, require_approved=True)

        self.assertIn("reviews.privacy.status must be approved", errors)

    def test_tampered_corpus_fails_integrity_validation(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            pcap = root / "sample.pcap"
            write_pcap(pcap, packets=1)
            artifact = pcap_evidence.inventory(pcap)
            manifest = {
                "schema_version": 1,
                "generated_utc": "2026-07-16T15:00:00Z",
                "evidence_class": "production-derived",
                "path_redacted": True,
                "file_count": 1,
                "packet_count": 1,
                "artifacts": artifact,
                **metadata(approved=True),
            }
            payload = bytearray(pcap.read_bytes())
            payload[-1] ^= 0xFF
            pcap.write_bytes(payload)

            errors = pcap_evidence.validate_manifest_data(
                manifest,
                corpus=pcap,
                require_approved=True,
            )

            self.assertIn(
                "corpus files, packet counts, byte sizes, or SHA-256 values differ from manifest",
                errors,
            )

    def test_placeholder_metadata_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            pcap = root / "sample.pcap"
            write_pcap(pcap, packets=1)
            values = metadata()
            values["provenance"]["source_description"] = "Replace with source"
            metadata_path = root / "metadata.json"
            metadata_path.write_text(json.dumps(values), encoding="utf-8")

            with self.assertRaisesRegex(pcap_evidence.EvidenceError, "placeholder"):
                pcap_evidence.generate(pcap, metadata_path, root / "output")

    def test_private_path_in_metadata_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            pcap = root / "sample.pcap"
            write_pcap(pcap, packets=1)
            values = metadata()
            values["provenance"]["custodian_reference"] = "/home/operator/private-review"
            metadata_path = root / "metadata.json"
            metadata_path.write_text(json.dumps(values), encoding="utf-8")

            with self.assertRaisesRegex(pcap_evidence.EvidenceError, "forbidden private"):
                pcap_evidence.generate(pcap, metadata_path, root / "output")


if __name__ == "__main__":
    unittest.main()
