import copy
import hashlib
import json
import subprocess
import tempfile
import unittest
from pathlib import Path

from scripts import benchmark_evidence


C_OUTPUT = """\
bench_parser/tcp_plain iterations=100 ns_per_packet=10.25 pps=97560976
bench_parser/tcp_vlan iterations=100 ns_per_packet=11.50 pps=86956522
bench_parser/tcp_qinq iterations=100 ns_per_packet=12.75 pps=78431373
bench_parser/sink=12345
bench_uds_sender/format_packet_json iterations=100 ns_per_op=101.25 ops_per_sec=9876543
bench_uds_sender/format_heartbeat_json iterations=100 ns_per_op=81.50 ops_per_sec=12269939
bench_uds_sender/uds_send_line iterations=100 ns_per_op=901.75 ops_per_sec=1108955
bench_uds_sender/sink=67890 avg_json_serialize_us=0.21 write_errors=0
"""

GO_LINES = [
    "BenchmarkMatcherMatch/no_hit-8 100 10.5 ns/op 100.0 MB/s 8 B/op 1 allocs/op",
    "BenchmarkMatcherMatch/multi_hit-8 100 20 ns/op 90 MB/s 16 B/op 2 allocs/op",
    "BenchmarkEngineMatch/no_hit-8 100 30 ns/op 80 MB/s 24 B/op 3 allocs/op",
    "BenchmarkEngineMatch/multi_hit-8 100 40 ns/op 70 MB/s 32 B/op 4 allocs/op",
    "BenchmarkStoreWriteBatch/single_alert-8 10 1000 ns/op 100 B/op 5 allocs/op 1 alerts/op",
    "BenchmarkStoreWriteBatch/batch_32_alerts-8 10 2000 ns/op 200 B/op 6 allocs/op 32 alerts/op",
    "BenchmarkStoreQuery/exact_rule-8 100 50 ns/op 40 B/op 2 allocs/op",
    "BenchmarkStoreQuery/timestamp_range-8 100 60 ns/op 48 B/op 3 allocs/op",
]
GO_OUTPUT = "goos: linux\ngoarch: amd64\n" + "\n".join(GO_LINES) + "\nPASS\nok example 1s\n"


class BenchmarkEvidenceParserTest(unittest.TestCase):
    def test_complete_c_surface_parses_typed_metrics(self):
        parsed = benchmark_evidence.parse_c_output(C_OUTPUT, 100)
        self.assertEqual(set(benchmark_evidence.C_CASES), set(parsed["cases"]))
        self.assertEqual(10.25, parsed["cases"]["bench_parser/tcp_plain"]["ns_per_packet"])
        self.assertEqual(0, parsed["summaries"]["uds_sender"]["write_errors"])

    def test_every_missing_c_case_is_rejected(self):
        lines = C_OUTPUT.splitlines()
        for name in benchmark_evidence.C_CASES:
            with self.subTest(name=name):
                incomplete = "\n".join(line for line in lines if not line.startswith(name + " "))
                with self.assertRaisesRegex(benchmark_evidence.EvidenceError, "missing C benchmark cases"):
                    benchmark_evidence.parse_c_output(incomplete, 100)

    def test_malformed_duplicate_unknown_and_write_error_c_output_is_rejected(self):
        mutations = [
            C_OUTPUT + C_OUTPUT.splitlines()[0] + "\n",
            C_OUTPUT + "bench_parser/unknown iterations=100 ns_per_packet=1 pps=1\n",
            C_OUTPUT.replace("ns_per_packet=10.25", "ns_per_packet=bad"),
            C_OUTPUT.replace("write_errors=0", "write_errors=1"),
            C_OUTPUT.replace("iterations=100", "iterations=99", 1),
        ]
        for output in mutations:
            with self.subTest(output=output[-100:]):
                with self.assertRaises(benchmark_evidence.EvidenceError):
                    benchmark_evidence.parse_c_output(output, 100)

    def test_complete_go_surface_parses_suffixes_and_metrics(self):
        parsed = benchmark_evidence.parse_go_output(GO_OUTPUT)
        self.assertEqual(set(benchmark_evidence.GO_CASES), set(parsed["cases"]))
        case = parsed["cases"]["BenchmarkStoreWriteBatch/batch_32_alerts"]
        self.assertEqual(8, case["cpu"])
        self.assertEqual(32, case["metrics"]["alerts/op"])

    def test_every_missing_go_case_is_rejected(self):
        for removed in GO_LINES:
            with self.subTest(removed=removed.split()[0]):
                output = "\n".join(line for line in GO_OUTPUT.splitlines() if line != removed)
                with self.assertRaisesRegex(benchmark_evidence.EvidenceError, "missing Go benchmark cases"):
                    benchmark_evidence.parse_go_output(output)

    def test_malformed_duplicate_unknown_and_failed_go_output_is_rejected(self):
        mutations = [
            GO_OUTPUT + GO_LINES[0] + "\n",
            GO_OUTPUT + "BenchmarkUnknown/case-8 1 1 ns/op\n",
            GO_OUTPUT.replace("10.5 ns/op", "bad ns/op", 1),
            GO_OUTPUT.replace("100.0 MB/s ", "", 1),
            GO_OUTPUT + "FAIL\texample/pkg [build failed]\n",
        ]
        for output in mutations:
            with self.subTest(output=output[-100:]):
                with self.assertRaises(benchmark_evidence.EvidenceError):
                    benchmark_evidence.parse_go_output(output)


class BenchmarkEvidenceContractTest(unittest.TestCase):
    def base_evidence(self):
        evidence = {
            "schema_version": 1,
            "status": "pass",
            "evidence_class": benchmark_evidence.EVIDENCE_CLASS,
            "production_derived": False,
            "threshold_applied": False,
            "release_or_publication_authority": False,
            "start": "2026-08-06T00:00:00Z",
            "end": "2026-08-06T00:00:01Z",
            "git": {
                "head": "a" * 40,
                "tree": "b" * 40,
                "branch": "main",
                "clean": True,
                "status_entry_count": 0,
                "status_sha256": "c" * 64,
                "tracked_diff_sha256": "d" * 64,
            },
            "environment": {
                "os": "Example Linux",
                "kernel": "1.0",
                "architecture": "x86_64",
                "python": "3.12",
                "go": "go1.25",
                "go_os": "linux",
                "go_arch": "amd64",
                "go_toolchain": "auto",
                "gcc": "gcc 13",
                "make": "GNU Make 4.3",
                "fingerprint_sha256": "",
            },
            "parameters": {
                "c_iterations": 100,
                "go_benchtime": "100x",
                "go_count": 1,
                "go_run_selector": "^$",
                "go_benchmark_selector": ".",
                "go_benchmem": True,
            },
            "commands": {
                "c": {
                    "command": ["make", "-s", "-C", "capture", "--no-print-directory", "bench", "BENCH_ITERATIONS=100"],
                    "working_directory": "<redacted-path>",
                    "start": "2026-08-06T00:00:00Z",
                    "end": "2026-08-06T00:00:01Z",
                    "elapsed_seconds": 1,
                    "exit_status": 0,
                    "stdout": C_OUTPUT,
                    "stderr": "",
                    "parsed": benchmark_evidence.parse_c_output(C_OUTPUT, 100),
                },
                "go": {
                    "command": ["go", "test", "-count=1", "-run", "^$", "-bench=.", "-benchtime=100x", "-benchmem", "./..."],
                    "working_directory": "<redacted-path>",
                    "start": "2026-08-06T00:00:00Z",
                    "end": "2026-08-06T00:00:01Z",
                    "elapsed_seconds": 1,
                    "exit_status": 0,
                    "stdout": GO_OUTPUT,
                    "stderr": "",
                    "parsed": benchmark_evidence.parse_go_output(GO_OUTPUT),
                },
            },
        }
        fields = dict(evidence["environment"])
        fields.pop("fingerprint_sha256")
        evidence["environment"]["fingerprint_sha256"] = hashlib.sha256(
            json.dumps(fields, sort_keys=True, separators=(",", ":")).encode()
        ).hexdigest()
        return evidence

    def test_complete_contract_passes_and_records_clean_or_dirty_exact_state(self):
        evidence = self.base_evidence()
        self.assertEqual([], benchmark_evidence.validation_errors(evidence))
        evidence["git"]["clean"] = False
        evidence["git"]["status_entry_count"] = 2
        self.assertEqual([], benchmark_evidence.validation_errors(evidence))

    def test_partial_surface_and_unbounded_claim_are_rejected(self):
        evidence = self.base_evidence()
        evidence["commands"]["go"]["parsed"]["cases"].pop(next(iter(benchmark_evidence.GO_CASES)))
        evidence["threshold_applied"] = True
        errors = benchmark_evidence.validation_errors(evidence)
        self.assertIn("parsed Go benchmark surface is incomplete", errors)
        self.assertTrue(any("threshold_applied" in error for error in errors))

    def test_parsed_results_must_match_raw_output(self):
        evidence = self.base_evidence()
        evidence["commands"]["c"]["parsed"]["cases"]["bench_parser/tcp_plain"]["pps"] = 1
        evidence["commands"]["go"]["parsed"]["cases"]["BenchmarkMatcherMatch/no_hit"]["metrics"]["ns/op"] = 1
        errors = benchmark_evidence.validation_errors(evidence)
        self.assertIn("parsed C results do not match raw output", errors)
        self.assertIn("parsed Go results do not match raw output", errors)

    def test_default_redaction_removes_repository_home_and_temporary_paths(self):
        with tempfile.TemporaryDirectory(prefix="operator private benchmark ") as directory:
            private = str(Path(directory) / "source tree")
            rendered = benchmark_evidence.redact_text(
                f"repo={private} home={Path.home()}/secret tmp=/tmp/private/run",
                (private, str(Path.home()), "/tmp"),
            )
        self.assertNotIn(private, rendered)
        self.assertNotIn(str(Path.home()), rendered)
        self.assertNotIn("/tmp/private", rendered)
        self.assertIn("<redacted-path>", rendered)

    def test_unredacted_sensitive_path_is_rejected(self):
        evidence = self.base_evidence()
        evidence["commands"]["c"]["stderr"] = "failed under /home/operator/private"
        self.assertIn(
            "evidence contains an unredacted sensitive absolute path",
            benchmark_evidence.validation_errors(evidence),
        )

    def test_json_round_trip_preserves_raw_and_parsed_results(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "evidence.json"
            benchmark_evidence.write_evidence(path, self.base_evidence())
            loaded = __import__("json").loads(path.read_text(encoding="utf-8"))
        self.assertEqual([], benchmark_evidence.validation_errors(loaded))
        self.assertIn("BenchmarkMatcherMatch/no_hit", loaded["commands"]["go"]["stdout"])

    def test_validation_rejects_clean_status_disagreement(self):
        evidence = copy.deepcopy(self.base_evidence())
        evidence["git"]["status_entry_count"] = 1
        self.assertIn(
            "git clean and status_entry_count disagree",
            benchmark_evidence.validation_errors(evidence),
        )

    def test_git_state_distinguishes_clean_and_dirty_without_exposing_paths(self):
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory)
            subprocess.run(("git", "init", "-q"), cwd=repo, check=True)
            subprocess.run(("git", "config", "user.email", "test@example.invalid"), cwd=repo, check=True)
            subprocess.run(("git", "config", "user.name", "Test"), cwd=repo, check=True)
            tracked = repo / "tracked.txt"
            tracked.write_text("clean\n", encoding="utf-8")
            subprocess.run(("git", "add", "tracked.txt"), cwd=repo, check=True)
            subprocess.run(("git", "commit", "-qm", "fixture"), cwd=repo, check=True)
            clean = benchmark_evidence.git_state(repo)
            tracked.write_text("dirty\n", encoding="utf-8")
            (repo / "private operator note.txt").write_text("local\n", encoding="utf-8")
            dirty = benchmark_evidence.git_state(repo)
        self.assertTrue(clean["clean"])
        self.assertEqual(0, clean["status_entry_count"])
        self.assertFalse(dirty["clean"])
        self.assertEqual(2, dirty["status_entry_count"])
        self.assertNotEqual(clean["status_sha256"], dirty["status_sha256"])
        self.assertNotEqual(clean["tracked_diff_sha256"], dirty["tracked_diff_sha256"])


class BenchmarkBaselineTest(unittest.TestCase):
    def build_samples(self, root: Path):
        paths = []
        for index in range(5):
            evidence = BenchmarkEvidenceContractTest().base_evidence()
            evidence["start"] = f"2026-08-06T00:00:0{index}Z"
            evidence["end"] = f"2026-08-06T00:00:1{index}Z"
            c_case = evidence["commands"]["c"]["parsed"]["cases"]["bench_parser/tcp_plain"]
            c_case["ns_per_packet"] = 10 * (index + 1)
            c_case["pps"] = 1000 - 100 * index
            evidence["commands"]["c"]["stdout"] = C_OUTPUT.replace(
                "ns_per_packet=10.25 pps=97560976",
                f"ns_per_packet={10 * (index + 1)} pps={1000 - 100 * index}",
            )
            go_case = evidence["commands"]["go"]["parsed"]["cases"]["BenchmarkMatcherMatch/no_hit"]
            go_case["metrics"]["ns/op"] = 10 * (index + 1)
            evidence["commands"]["go"]["stdout"] = GO_OUTPUT.replace(
                "100 10.5 ns/op", f"100 {10 * (index + 1)} ns/op", 1
            )
            path = root / f"sample-{index + 1:02d}.json"
            benchmark_evidence.write_evidence(path, evidence)
            paths.append(path)
        return paths

    def test_five_matched_samples_produce_defined_statistics(self):
        with tempfile.TemporaryDirectory() as directory:
            paths = self.build_samples(Path(directory))
            baseline = benchmark_evidence.aggregate_samples(paths)
        stats = baseline["metrics"]["c"]["bench_parser/tcp_plain"]["ns_per_packet"]
        self.assertEqual([10, 20, 30, 40, 50], stats["values"])
        self.assertEqual(30, stats["median"])
        self.assertEqual(20, stats["q1_inclusive"])
        self.assertEqual(40, stats["q3_inclusive"])
        self.assertEqual(20, stats["iqr_inclusive"])
        self.assertEqual(5, stats["count"])
        self.assertEqual(benchmark_evidence.BASELINE_EVIDENCE_CLASS, baseline["evidence_class"])
        self.assertFalse(baseline["threshold_applied"])
        self.assertFalse(baseline["portable_or_cross_host_claim"])

    def test_fewer_than_five_and_duplicate_basenames_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            paths = self.build_samples(root)
            with self.assertRaisesRegex(benchmark_evidence.EvidenceError, "at least 5"):
                benchmark_evidence.aggregate_samples(paths[:4])
            other = root / "other"
            other.mkdir()
            duplicate = other / paths[0].name
            duplicate.write_bytes(paths[0].read_bytes())
            with self.assertRaisesRegex(benchmark_evidence.EvidenceError, "basenames must be unique"):
                benchmark_evidence.aggregate_samples(paths + [duplicate])

    def test_context_and_contract_drift_are_rejected(self):
        mutations = [
            ("dirty", "clean Git state"),
            ("head", "Git head differs"),
            ("tree", "Git tree differs"),
            ("parameters", "parameters differ"),
            ("environment", "environment differs"),
        ]
        for mutation, message in mutations:
            with self.subTest(mutation=mutation):
                with tempfile.TemporaryDirectory() as directory:
                    paths = self.build_samples(Path(directory))
                    sample = json.loads(paths[-1].read_text())
                    if mutation == "dirty":
                        sample["git"]["clean"] = False
                        sample["git"]["status_entry_count"] = 1
                    elif mutation == "head":
                        sample["git"]["head"] = "f" * 40
                    elif mutation == "tree":
                        sample["git"]["tree"] = "e" * 40
                    elif mutation == "parameters":
                        sample["parameters"]["go_benchtime"] = "1s"
                        sample["commands"]["go"]["command"][6] = "-benchtime=1s"
                    elif mutation == "environment":
                        sample["environment"]["kernel"] = "different"
                        fields = dict(sample["environment"])
                        fields.pop("fingerprint_sha256")
                        sample["environment"]["fingerprint_sha256"] = hashlib.sha256(
                            json.dumps(fields, sort_keys=True, separators=(",", ":")).encode()
                        ).hexdigest()
                    paths[-1].write_text(json.dumps(sample), encoding="utf-8")
                    with self.assertRaisesRegex(benchmark_evidence.EvidenceError, message):
                        benchmark_evidence.aggregate_samples(paths)

    def test_metric_surface_drift_is_rejected_by_raw_reparse(self):
        with tempfile.TemporaryDirectory() as directory:
            paths = self.build_samples(Path(directory))
            sample = json.loads(paths[-1].read_text())
            sample["commands"]["go"]["parsed"]["cases"]["BenchmarkStoreQuery/exact_rule"]["metrics"].pop("ns/op")
            paths[-1].write_text(json.dumps(sample), encoding="utf-8")
            with self.assertRaisesRegex(benchmark_evidence.EvidenceError, "parsed Go results do not match raw output"):
                benchmark_evidence.aggregate_samples(paths)

    def test_revalidation_detects_tampered_digest_and_aggregate_value(self):
        with tempfile.TemporaryDirectory() as directory:
            paths = self.build_samples(Path(directory))
            baseline = benchmark_evidence.aggregate_samples(paths)
            self.assertEqual([], benchmark_evidence.baseline_validation_errors(baseline, paths))
            tampered = copy.deepcopy(baseline)
            tampered["samples"][0]["sha256"] = "0" * 64
            self.assertIn(
                "baseline does not match recomputation from raw samples",
                benchmark_evidence.baseline_validation_errors(tampered, paths),
            )
            tampered = copy.deepcopy(baseline)
            tampered["metrics"]["c"]["bench_parser/tcp_plain"]["ns_per_packet"]["median"] = 1
            self.assertIn(
                "baseline does not match recomputation from raw samples",
                benchmark_evidence.baseline_validation_errors(tampered, paths),
            )

    def test_markdown_is_observation_only_and_contains_no_sample_path(self):
        with tempfile.TemporaryDirectory(prefix="private baseline ") as directory:
            root = Path(directory)
            paths = self.build_samples(root)
            baseline = benchmark_evidence.aggregate_samples(paths)
            report = root / "report.md"
            benchmark_evidence.write_baseline_markdown(report, baseline)
            rendered = report.read_text()
        self.assertNotIn(str(root), rendered)
        self.assertIn("Threshold applied: false", rendered)
        self.assertIn("does not establish production capacity", rendered)


if __name__ == "__main__":
    unittest.main()
