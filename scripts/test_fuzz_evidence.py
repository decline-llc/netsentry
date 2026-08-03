import argparse
import json
import tempfile
import unittest
from pathlib import Path

from scripts import fuzz_evidence


class FuzzEvidenceTest(unittest.TestCase):
    def build(self, root: Path, *, corpus: str = "", include_paths: str = "0"):
        parser_log = root / "parser.log"
        formatter_log = root / "formatter.log"
        parser_log.write_text("fuzz_parser: ok iterations=100\n", encoding="utf-8")
        formatter_log.write_text(
            "fuzz_uds_formatter: ok iterations=100\n", encoding="utf-8"
        )
        return argparse.Namespace(
            run_id="20260803T090000Z",
            start="2026-08-03T09:00:00Z",
            end="2026-08-03T09:00:01Z",
            start_ns=1,
            end_ns=1_000_000_001,
            iterations=100,
            parser_status=0,
            parser_start_ns=1,
            parser_end_ns=400_000_001,
            formatter_status=0,
            formatter_start_ns=400_000_001,
            formatter_end_ns=1_000_000_001,
            corpus_status=0,
            corpus=corpus,
            corpus_files=1 if corpus else 0,
            include_paths=include_paths,
            compiler_version="cc 1.0",
            make_version="GNU Make 4.3",
            bash_version="GNU bash 5.2",
            parser_log=parser_log,
            formatter_log=formatter_log,
        )

    def test_complete_dual_harness_evidence_passes(self):
        with tempfile.TemporaryDirectory() as directory:
            summary = fuzz_evidence.build_summary(self.build(Path(directory)))
        self.assertEqual([], fuzz_evidence.validation_errors(summary, require_pass=True))
        self.assertEqual({"parser", "formatter"}, set(summary["harnesses"]))
        self.assertEqual("local_synthetic", summary["evidence_class"])
        self.assertFalse(summary["production_derived"])
        self.assertFalse(summary["release_or_publication_authority"])

    def test_failed_or_missing_harness_is_rejected_for_pass_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.build(Path(directory))
            args.formatter_status = 1
            summary = fuzz_evidence.build_summary(args)
        errors = fuzz_evidence.validation_errors(summary, require_pass=True)
        self.assertTrue(any("formatter harness must pass" in error for error in errors))
        summary["harnesses"].pop("formatter")
        errors = fuzz_evidence.validation_errors(summary)
        self.assertIn("harnesses must contain exactly parser and formatter", errors)

    def test_iteration_mismatch_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            summary = fuzz_evidence.build_summary(self.build(Path(directory)))
        summary["harnesses"]["parser"]["iterations"] = 99
        errors = fuzz_evidence.validation_errors(summary)
        self.assertIn("parser iterations must equal the recorded budget", errors)

    def test_harness_log_must_confirm_iteration_budget(self):
        with tempfile.TemporaryDirectory() as directory:
            summary = fuzz_evidence.build_summary(self.build(Path(directory)))
        summary["harnesses"]["formatter"]["log_tail"] = [
            "fuzz_uds_formatter: ok iterations=99"
        ]
        errors = fuzz_evidence.validation_errors(summary, require_pass=True)
        self.assertIn(
            "formatter log_tail must confirm the recorded iteration budget", errors
        )

    def test_corpus_path_is_redacted_from_json_and_markdown(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = str(root / "operator private corpus")
            args = self.build(root, corpus=private)
            args.parser_log.write_text(
                f"replayed {private}/seed.bin\n", encoding="utf-8"
            )
            summary = fuzz_evidence.build_summary(args)
            markdown = root / "summary.md"
            fuzz_evidence.write_markdown(markdown, summary)
            rendered = json.dumps(summary) + markdown.read_text(encoding="utf-8")
        self.assertNotIn(private, rendered)
        self.assertEqual("redacted", summary["corpus"]["path"])
        self.assertTrue(summary["corpus"]["path_redacted"])
        self.assertIn("<redacted-corpus>/seed.bin", rendered)

    def test_sanitizer_marker_blocks_pass_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.build(Path(directory))
            args.parser_log.write_text("ERROR: AddressSanitizer: boom\n", encoding="utf-8")
            summary = fuzz_evidence.build_summary(args)
        errors = fuzz_evidence.validation_errors(summary, require_pass=True)
        self.assertTrue(any("parser harness must pass" in error for error in errors))
        self.assertTrue(any("zero sanitizer findings" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
