#!/usr/bin/env python3
"""Capture and validate path-redacted local C/Go benchmark evidence."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import platform
import re
import statistics
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any, Callable, Sequence


SCHEMA_VERSION = 1
EVIDENCE_CLASS = "local_synthetic_microbenchmark"
BASELINE_SCHEMA_VERSION = 1
BASELINE_EVIDENCE_CLASS = "observation_only_local_synthetic_microbenchmark"
MINIMUM_BASELINE_SAMPLES = 5
C_CASES = {
    "bench_parser/tcp_plain": ("iterations", "ns_per_packet", "pps"),
    "bench_parser/tcp_vlan": ("iterations", "ns_per_packet", "pps"),
    "bench_parser/tcp_qinq": ("iterations", "ns_per_packet", "pps"),
    "bench_uds_sender/format_packet_json": ("iterations", "ns_per_op", "ops_per_sec"),
    "bench_uds_sender/format_heartbeat_json": ("iterations", "ns_per_op", "ops_per_sec"),
    "bench_uds_sender/uds_send_line": ("iterations", "ns_per_op", "ops_per_sec"),
}
GO_CASES = {
    "BenchmarkMatcherMatch/no_hit": ("ns/op", "MB/s", "B/op", "allocs/op"),
    "BenchmarkMatcherMatch/multi_hit": ("ns/op", "MB/s", "B/op", "allocs/op"),
    "BenchmarkEngineMatch/no_hit": ("ns/op", "MB/s", "B/op", "allocs/op"),
    "BenchmarkEngineMatch/multi_hit": ("ns/op", "MB/s", "B/op", "allocs/op"),
    "BenchmarkStoreWriteBatch/single_alert": ("ns/op", "B/op", "allocs/op", "alerts/op"),
    "BenchmarkStoreWriteBatch/batch_32_alerts": ("ns/op", "B/op", "allocs/op", "alerts/op"),
    "BenchmarkStoreQuery/exact_rule": ("ns/op", "B/op", "allocs/op"),
    "BenchmarkStoreQuery/timestamp_range": ("ns/op", "B/op", "allocs/op"),
}
NUMBER = r"[-+]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][-+]?\d+)?"
C_RESULT_RE = re.compile(
    rf"^(?P<name>bench_(?:parser|uds_sender)/[^ ]+) "
    rf"iterations=(?P<iterations>\d+) "
    rf"(?P<metric>ns_per_packet|ns_per_op)=(?P<latency>{NUMBER}) "
    rf"(?P<rate_name>pps|ops_per_sec)=(?P<rate>{NUMBER})$"
)
GO_RESULT_RE = re.compile(
    rf"^(?P<name>Benchmark[^\s]+)-(?P<cpu>\d+)\s+"
    rf"(?P<iterations>\d+)\s+(?P<metrics>.+)$"
)
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
SENSITIVE_PATH_RE = re.compile(
    r"(?<![A-Za-z0-9_.-])/(?:home|tmp|var/tmp)(?:/[^\s\"'`,;:)\]]*)?"
)


class EvidenceError(ValueError):
    """Raised when benchmark output or evidence is incomplete or malformed."""


def _number(value: str) -> int | float:
    parsed = float(value)
    return int(parsed) if parsed.is_integer() else parsed


def parse_c_output(output: str, expected_iterations: int) -> dict[str, Any]:
    cases: dict[str, Any] = {}
    parser_sink_seen = False
    uds_summary: dict[str, Any] | None = None
    for raw_line in output.splitlines():
        line = raw_line.strip()
        match = C_RESULT_RE.fullmatch(line)
        if match:
            name = match.group("name")
            if name not in C_CASES:
                raise EvidenceError(f"unknown C benchmark case: {name}")
            if name in cases:
                raise EvidenceError(f"duplicate C benchmark case: {name}")
            iterations = int(match.group("iterations"))
            if iterations != expected_iterations:
                raise EvidenceError(
                    f"{name} iterations {iterations}, want {expected_iterations}"
                )
            expected_latency = C_CASES[name][1]
            expected_rate = C_CASES[name][2]
            if match.group("metric") != expected_latency or match.group("rate_name") != expected_rate:
                raise EvidenceError(f"unexpected C metrics for {name}")
            latency = _number(match.group("latency"))
            rate = _number(match.group("rate"))
            if latency <= 0 or rate <= 0:
                raise EvidenceError(f"{name} metrics must be positive")
            cases[name] = {
                "iterations": iterations,
                expected_latency: latency,
                expected_rate: rate,
            }
            continue
        if re.fullmatch(r"bench_parser/sink=\d+", line):
            if parser_sink_seen:
                raise EvidenceError("duplicate C parser sink summary")
            parser_sink_seen = True
            continue
        summary = re.fullmatch(
            rf"bench_uds_sender/sink=(\d+) avg_json_serialize_us=({NUMBER}) write_errors=(\d+)",
            line,
        )
        if summary:
            if uds_summary is not None:
                raise EvidenceError("duplicate C UDS summary")
            uds_summary = {
                "sink": int(summary.group(1)),
                "avg_json_serialize_us": _number(summary.group(2)),
                "write_errors": int(summary.group(3)),
            }
            continue
        if line.startswith(("bench_parser/", "bench_uds_sender/")):
            raise EvidenceError(f"malformed C benchmark output: {line}")

    missing = sorted(set(C_CASES) - set(cases))
    if missing:
        raise EvidenceError(f"missing C benchmark cases: {', '.join(missing)}")
    if not parser_sink_seen:
        raise EvidenceError("missing C parser sink summary")
    if uds_summary is None:
        raise EvidenceError("missing C UDS summary")
    if uds_summary["write_errors"] != 0:
        raise EvidenceError("C UDS benchmark reported write errors")
    return {"cases": cases, "summaries": {"uds_sender": uds_summary}}


def parse_go_output(output: str) -> dict[str, Any]:
    cases: dict[str, Any] = {}
    if any(line.strip() == "FAIL" or line.startswith("FAIL\t") for line in output.splitlines()):
        raise EvidenceError("Go benchmark output contains a failed package")
    for raw_line in output.splitlines():
        line = raw_line.strip()
        match = GO_RESULT_RE.fullmatch(line)
        if not match:
            if line.startswith("Benchmark"):
                raise EvidenceError(f"malformed Go benchmark output: {line}")
            continue
        name = match.group("name")
        if name not in GO_CASES:
            raise EvidenceError(f"unknown Go benchmark case: {name}")
        if name in cases:
            raise EvidenceError(f"duplicate Go benchmark case: {name}")
        tokens = match.group("metrics").split()
        if len(tokens) % 2:
            raise EvidenceError(f"malformed Go metrics for {name}")
        metrics: dict[str, int | float] = {}
        for index in range(0, len(tokens), 2):
            value, unit = tokens[index : index + 2]
            if not re.fullmatch(NUMBER, value):
                raise EvidenceError(f"non-numeric Go metric for {name}: {value}")
            if unit in metrics:
                raise EvidenceError(f"duplicate Go metric for {name}: {unit}")
            metrics[unit] = _number(value)
        expected_metrics = set(GO_CASES[name])
        if set(metrics) != expected_metrics:
            raise EvidenceError(
                f"{name} metrics {sorted(metrics)}, want {sorted(expected_metrics)}"
            )
        if metrics["ns/op"] <= 0 or any(value < 0 for value in metrics.values()):
            raise EvidenceError(f"{name} metrics must be nonnegative with positive ns/op")
        cases[name] = {
            "cpu": int(match.group("cpu")),
            "iterations": int(match.group("iterations")),
            "metrics": metrics,
        }
        if cases[name]["cpu"] <= 0 or cases[name]["iterations"] <= 0:
            raise EvidenceError(f"{name} cpu and iterations must be positive")
    missing = sorted(set(GO_CASES) - set(cases))
    if missing:
        raise EvidenceError(f"missing Go benchmark cases: {', '.join(missing)}")
    return {"cases": cases}


def redact_text(value: str, paths: Sequence[str]) -> str:
    redacted = value
    replacements = sorted(
        {str(Path(path).resolve()) for path in paths if path}, key=len, reverse=True
    )
    for path in replacements:
        redacted = redacted.replace(path, "<redacted-path>")
    return SENSITIVE_PATH_RE.sub("<redacted-path>", redacted)


def _run_text(command: Sequence[str], cwd: Path) -> str:
    result = subprocess.run(
        command, cwd=cwd, text=True, stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT, check=True,
    )
    return result.stdout.strip()


def git_state(repo: Path) -> dict[str, Any]:
    head = _run_text(("git", "rev-parse", "HEAD"), repo)
    tree = _run_text(("git", "rev-parse", "HEAD^{tree}"), repo)
    branch = _run_text(("git", "branch", "--show-current"), repo) or "detached"
    status = subprocess.run(
        ("git", "status", "--porcelain=v1", "--untracked-files=all"),
        cwd=repo, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True,
    ).stdout
    tracked_diff = subprocess.run(
        ("git", "diff", "--binary", "HEAD"), cwd=repo,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True,
    ).stdout
    entries = [line for line in status.splitlines() if line]
    return {
        "head": head,
        "tree": tree,
        "branch": branch,
        "clean": not entries,
        "status_entry_count": len(entries),
        "status_sha256": hashlib.sha256(status.encode()).hexdigest(),
        "tracked_diff_sha256": hashlib.sha256(tracked_diff).hexdigest(),
    }


def environment(repo: Path) -> dict[str, Any]:
    go_values = _run_text(
        ("go", "env", "GOVERSION", "GOOS", "GOARCH", "GOTOOLCHAIN"),
        repo / "engine",
    ).splitlines()
    if len(go_values) != 4:
        raise EvidenceError("go env returned an incomplete toolchain fingerprint")
    os_release: dict[str, str] = {}
    release_path = Path("/etc/os-release")
    if release_path.exists():
        for line in release_path.read_text(encoding="utf-8").splitlines():
            if "=" in line:
                key, value = line.split("=", 1)
                os_release[key] = value.strip().strip('"')
    result = {
        "os": os_release.get("PRETTY_NAME", platform.system()),
        "kernel": platform.release(),
        "architecture": platform.machine(),
        "python": platform.python_version(),
        "go": go_values[0],
        "go_os": go_values[1],
        "go_arch": go_values[2],
        "go_toolchain": go_values[3],
        "gcc": _run_text(("gcc", "--version"), repo).splitlines()[0],
        "make": _run_text(("make", "--version"), repo).splitlines()[0],
    }
    result["fingerprint_sha256"] = hashlib.sha256(
        json.dumps(result, sort_keys=True, separators=(",", ":")).encode()
    ).hexdigest()
    return result


def _utc_now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z")


def execute(
    command: Sequence[str], cwd: Path, redaction_paths: Sequence[str]
) -> dict[str, Any]:
    start = _utc_now()
    start_ns = time.monotonic_ns()
    result = subprocess.run(
        command, cwd=cwd, text=True, stdout=subprocess.PIPE,
        stderr=subprocess.PIPE, check=False,
    )
    elapsed = (time.monotonic_ns() - start_ns) / 1_000_000_000
    return {
        "command": [redact_text(part, redaction_paths) for part in command],
        "working_directory": redact_text(str(cwd.resolve()), redaction_paths),
        "start": start,
        "end": _utc_now(),
        "elapsed_seconds": round(elapsed, 6),
        "exit_status": result.returncode,
        "stdout": redact_text(result.stdout, redaction_paths),
        "stderr": redact_text(result.stderr, redaction_paths),
    }


def build_evidence(
    repo: Path,
    c_iterations: int,
    go_benchtime: str,
    runner: Callable[[Sequence[str], Path, Sequence[str]], dict[str, Any]] = execute,
) -> dict[str, Any]:
    repo = repo.resolve()
    if c_iterations <= 0:
        raise EvidenceError("C iterations must be positive")
    if not re.fullmatch(
        r"(?:[1-9]\d*x|(?:(?:[1-9]\d*)(?:\.\d+)?|0\.\d*[1-9]\d*)(?:ns|us|ms|s|m|h))",
        go_benchtime,
    ):
        raise EvidenceError("Go benchtime must be a positive duration or iteration count")
    redaction_paths = (str(repo), str(Path.home()), tempfile.gettempdir())
    state = git_state(repo)
    env = environment(repo)
    started = _utc_now()
    c_command = (
        "make", "-s", "-C", "capture", "--no-print-directory", "bench",
        f"BENCH_ITERATIONS={c_iterations}",
    )
    go_command = (
        "go", "test", "-count=1", "-run", "^$", "-bench=.",
        f"-benchtime={go_benchtime}", "-benchmem", "./...",
    )
    c_run = runner(c_command, repo, redaction_paths)
    if c_run["exit_status"] != 0:
        raise EvidenceError(f"C benchmark command failed with status {c_run['exit_status']}")
    c_run["parsed"] = parse_c_output(c_run["stdout"], c_iterations)
    go_run = runner(go_command, repo / "engine", redaction_paths)
    if go_run["exit_status"] != 0:
        raise EvidenceError(f"Go benchmark command failed with status {go_run['exit_status']}")
    go_run["parsed"] = parse_go_output(go_run["stdout"])
    evidence = {
        "schema_version": SCHEMA_VERSION,
        "status": "pass",
        "evidence_class": EVIDENCE_CLASS,
        "production_derived": False,
        "threshold_applied": False,
        "release_or_publication_authority": False,
        "start": started,
        "end": _utc_now(),
        "git": state,
        "environment": env,
        "parameters": {
            "c_iterations": c_iterations,
            "go_benchtime": go_benchtime,
            "go_count": 1,
            "go_run_selector": "^$",
            "go_benchmark_selector": ".",
            "go_benchmem": True,
        },
        "commands": {"c": c_run, "go": go_run},
    }
    errors = validation_errors(evidence)
    if errors:
        raise EvidenceError("; ".join(errors))
    return evidence


def validation_errors(evidence: Any) -> list[str]:
    errors: list[str] = []
    if not isinstance(evidence, dict):
        return ["evidence must be an object"]
    if evidence.get("schema_version") != SCHEMA_VERSION:
        errors.append(f"schema_version must be {SCHEMA_VERSION}")
    expected_flags = {
        "status": "pass",
        "evidence_class": EVIDENCE_CLASS,
        "production_derived": False,
        "threshold_applied": False,
        "release_or_publication_authority": False,
    }
    for name, expected in expected_flags.items():
        if evidence.get(name) != expected:
            errors.append(f"{name} must be {expected!r}")
    for name in ("start", "end"):
        if not isinstance(evidence.get(name), str) or not evidence[name].endswith("Z"):
            errors.append(f"{name} must be a UTC timestamp")
    git = evidence.get("git")
    if not isinstance(git, dict):
        errors.append("git must be an object")
    else:
        for name in ("head", "tree"):
            if not isinstance(git.get(name), str) or not SHA_RE.fullmatch(git[name]):
                errors.append(f"git {name} must be a full SHA")
        if not isinstance(git.get("branch"), str) or not git["branch"]:
            errors.append("git branch must be nonempty")
        if not isinstance(git.get("clean"), bool):
            errors.append("git clean must be boolean")
        if not isinstance(git.get("status_entry_count"), int) or isinstance(git.get("status_entry_count"), bool) or git["status_entry_count"] < 0:
            errors.append("git status_entry_count must be nonnegative")
        for name in ("status_sha256", "tracked_diff_sha256"):
            if not isinstance(git.get(name), str) or not re.fullmatch(r"[0-9a-f]{64}", git[name]):
                errors.append(f"git {name} must be a SHA-256")
        if isinstance(git.get("clean"), bool) and isinstance(git.get("status_entry_count"), int):
            if git["clean"] != (git["status_entry_count"] == 0):
                errors.append("git clean and status_entry_count disagree")
    env = evidence.get("environment")
    if not isinstance(env, dict):
        errors.append("environment must be an object")
    else:
        for name in ("os", "kernel", "architecture", "python", "go", "go_os", "go_arch", "go_toolchain", "gcc", "make"):
            if not isinstance(env.get(name), str) or not env[name]:
                errors.append(f"environment {name} must be nonempty")
        if not isinstance(env.get("fingerprint_sha256"), str) or not re.fullmatch(r"[0-9a-f]{64}", env["fingerprint_sha256"]):
            errors.append("environment fingerprint_sha256 must be a SHA-256")
        else:
            fingerprint_fields = dict(env)
            recorded_fingerprint = fingerprint_fields.pop("fingerprint_sha256")
            expected_fingerprint = hashlib.sha256(
                json.dumps(fingerprint_fields, sort_keys=True, separators=(",", ":")).encode()
            ).hexdigest()
            if recorded_fingerprint != expected_fingerprint:
                errors.append("environment fingerprint does not match its fields")
    parameters = evidence.get("parameters")
    if not isinstance(parameters, dict):
        errors.append("parameters must be an object")
    else:
        if not isinstance(parameters.get("c_iterations"), int) or isinstance(parameters.get("c_iterations"), bool) or parameters["c_iterations"] <= 0:
            errors.append("parameters c_iterations must be positive")
        if not isinstance(parameters.get("go_benchtime"), str) or not parameters["go_benchtime"]:
            errors.append("parameters go_benchtime must be nonempty")
        if parameters.get("go_count") != 1 or parameters.get("go_run_selector") != "^$" or parameters.get("go_benchmark_selector") != "." or parameters.get("go_benchmem") is not True:
            errors.append("Go command parameters must preserve the complete uncached benchmark boundary")
    commands = evidence.get("commands")
    if not isinstance(commands, dict) or set(commands) != {"c", "go"}:
        errors.append("commands must contain exactly c and go")
    else:
        for name in ("c", "go"):
            command = commands[name]
            if not isinstance(command, dict):
                errors.append(f"{name} command must be an object")
                continue
            if command.get("exit_status") != 0:
                errors.append(f"{name} command must exit successfully")
            if not isinstance(command.get("command"), list) or not command["command"]:
                errors.append(f"{name} command argv must be nonempty")
            if not isinstance(command.get("stdout"), str) or not isinstance(command.get("stderr"), str):
                errors.append(f"{name} raw output must be strings")
            elapsed = command.get("elapsed_seconds")
            if not isinstance(elapsed, (int, float)) or isinstance(elapsed, bool) or elapsed < 0:
                errors.append(f"{name} elapsed_seconds must be nonnegative")
            parsed = command.get("parsed")
            if not isinstance(parsed, dict):
                errors.append(f"{name} parsed output must be an object")
            elif name == "c" and set(parsed.get("cases", {})) != set(C_CASES):
                errors.append("parsed C benchmark surface is incomplete")
            elif name == "go" and set(parsed.get("cases", {})) != set(GO_CASES):
                errors.append("parsed Go benchmark surface is incomplete")
        if isinstance(parameters, dict):
            expected_c_command = [
                "make", "-s", "-C", "capture", "--no-print-directory", "bench",
                f"BENCH_ITERATIONS={parameters.get('c_iterations')}",
            ]
            expected_go_command = [
                "go", "test", "-count=1", "-run", "^$", "-bench=.",
                f"-benchtime={parameters.get('go_benchtime')}", "-benchmem", "./...",
            ]
            if isinstance(commands.get("c"), dict) and commands["c"].get("command") != expected_c_command:
                errors.append("C command does not match the recorded parameters")
            if isinstance(commands.get("go"), dict) and commands["go"].get("command") != expected_go_command:
                errors.append("Go command does not match the recorded parameters")
            if isinstance(commands.get("c"), dict) and isinstance(commands["c"].get("stdout"), str) and isinstance(parameters.get("c_iterations"), int):
                try:
                    reparsed_c = parse_c_output(commands["c"]["stdout"], parameters["c_iterations"])
                    if reparsed_c != commands["c"].get("parsed"):
                        errors.append("parsed C results do not match raw output")
                except EvidenceError as error:
                    errors.append(f"raw C output is invalid: {error}")
            if isinstance(commands.get("go"), dict) and isinstance(commands["go"].get("stdout"), str):
                try:
                    reparsed_go = parse_go_output(commands["go"]["stdout"])
                    if reparsed_go != commands["go"].get("parsed"):
                        errors.append("parsed Go results do not match raw output")
                except EvidenceError as error:
                    errors.append(f"raw Go output is invalid: {error}")
    rendered = json.dumps(evidence, sort_keys=True)
    if SENSITIVE_PATH_RE.search(rendered):
        errors.append("evidence contains an unredacted sensitive absolute path")
    return errors


def write_evidence(path: Path, evidence: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(json.dumps(evidence, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    os.replace(temporary, path)


def _rounded(value: int | float) -> int | float:
    if isinstance(value, int):
        return value
    rounded = float(f"{value:.12g}")
    return int(rounded) if rounded.is_integer() else rounded


def _display_number(value: int | float) -> str:
    if isinstance(value, int):
        return str(value)
    return f"{value:.6f}".rstrip("0").rstrip(".")


def descriptive_statistics(values: Sequence[int | float]) -> dict[str, Any]:
    if len(values) < MINIMUM_BASELINE_SAMPLES:
        raise EvidenceError(
            f"baseline metric requires at least {MINIMUM_BASELINE_SAMPLES} values"
        )
    numeric = [float(value) for value in values]
    q1, _, q3 = statistics.quantiles(numeric, n=4, method="inclusive")
    median = statistics.median(numeric)
    mean = statistics.fmean(numeric)
    deviation = statistics.stdev(numeric)
    return {
        "values": [_rounded(value) for value in values],
        "count": len(values),
        "minimum": _rounded(min(numeric)),
        "maximum": _rounded(max(numeric)),
        "median": _rounded(median),
        "q1_inclusive": _rounded(q1),
        "q3_inclusive": _rounded(q3),
        "iqr_inclusive": _rounded(q3 - q1),
        "mean": _rounded(mean),
        "sample_standard_deviation": _rounded(deviation),
        "coefficient_of_variation_percent": _rounded(
            0 if mean == 0 else deviation / abs(mean) * 100
        ),
        "range_percent_of_median": _rounded(
            0 if median == 0 else (max(numeric) - min(numeric)) / abs(median) * 100
        ),
    }


def _sample_metric_values(samples: Sequence[dict[str, Any]]) -> dict[str, Any]:
    metrics: dict[str, Any] = {"c": {}, "go": {}}
    for case in sorted(C_CASES):
        case_metrics: dict[str, Any] = {}
        for metric in C_CASES[case][1:]:
            case_metrics[metric] = descriptive_statistics(
                [sample["commands"]["c"]["parsed"]["cases"][case][metric] for sample in samples]
            )
        metrics["c"][case] = case_metrics
    metrics["c"]["bench_uds_sender/summary"] = {
        "avg_json_serialize_us": descriptive_statistics(
            [
                sample["commands"]["c"]["parsed"]["summaries"]["uds_sender"]["avg_json_serialize_us"]
                for sample in samples
            ]
        )
    }
    for case in sorted(GO_CASES):
        metrics["go"][case] = {
            metric: descriptive_statistics(
                [sample["commands"]["go"]["parsed"]["cases"][case]["metrics"][metric] for sample in samples]
            )
            for metric in GO_CASES[case]
        }
    return metrics


def aggregate_samples(paths: Sequence[Path]) -> dict[str, Any]:
    if len(paths) < MINIMUM_BASELINE_SAMPLES:
        raise EvidenceError(
            f"baseline requires at least {MINIMUM_BASELINE_SAMPLES} samples"
        )
    ordered_paths = sorted((Path(path) for path in paths), key=lambda path: path.name)
    names = [path.name for path in ordered_paths]
    if len(names) != len(set(names)):
        raise EvidenceError("baseline sample basenames must be unique")
    samples: list[dict[str, Any]] = []
    references: list[dict[str, Any]] = []
    for path in ordered_paths:
        try:
            raw = path.read_bytes()
            sample = json.loads(raw)
        except (OSError, json.JSONDecodeError) as error:
            raise EvidenceError(f"cannot read sample {path.name}: {error}") from error
        errors = validation_errors(sample)
        if errors:
            raise EvidenceError(f"invalid sample {path.name}: {'; '.join(errors)}")
        samples.append(sample)
        references.append(
            {
                "file": path.name,
                "sha256": hashlib.sha256(raw).hexdigest(),
                "start": sample["start"],
                "end": sample["end"],
            }
        )

    reference = samples[0]
    for index, sample in enumerate(samples, start=1):
        name = references[index - 1]["file"]
        if sample["git"]["clean"] is not True or sample["git"]["status_entry_count"] != 0:
            raise EvidenceError(f"sample {name} must record a clean Git state")
        for field in (
            "head", "tree", "clean", "status_entry_count", "status_sha256",
            "tracked_diff_sha256",
        ):
            if sample["git"][field] != reference["git"][field]:
                raise EvidenceError(f"sample {name} Git {field} differs")
        if sample["environment"] != reference["environment"]:
            raise EvidenceError(f"sample {name} environment differs")
        if sample["parameters"] != reference["parameters"]:
            raise EvidenceError(f"sample {name} parameters differ")
        for family in ("c", "go"):
            if sample["commands"][family]["command"] != reference["commands"][family]["command"]:
                raise EvidenceError(f"sample {name} {family} command differs")

    return {
        "schema_version": BASELINE_SCHEMA_VERSION,
        "status": "observation_only",
        "evidence_class": BASELINE_EVIDENCE_CLASS,
        "production_derived": False,
        "threshold_applied": False,
        "portable_or_cross_host_claim": False,
        "release_or_publication_authority": False,
        "quartile_method": "inclusive",
        "variation_method": "sample_standard_deviation_over_absolute_mean_percent",
        "sampling_mode": "sequential_uncached_same_host",
        "sample_count": len(samples),
        "start": references[0]["start"],
        "end": references[-1]["end"],
        "git": {
            "head": reference["git"]["head"],
            "tree": reference["git"]["tree"],
            "branch": reference["git"]["branch"],
            "clean": True,
        },
        "environment": reference["environment"],
        "parameters": reference["parameters"],
        "commands": {
            family: reference["commands"][family]["command"] for family in ("c", "go")
        },
        "samples": references,
        "metrics": _sample_metric_values(samples),
    }


def baseline_validation_errors(
    baseline: Any, sample_paths: Sequence[Path]
) -> list[str]:
    if not isinstance(baseline, dict):
        return ["baseline must be an object"]
    try:
        expected = aggregate_samples(sample_paths)
    except EvidenceError as error:
        return [str(error)]
    errors: list[str] = []
    if baseline != expected:
        errors.append("baseline does not match recomputation from raw samples")
    rendered = json.dumps(baseline, sort_keys=True)
    if SENSITIVE_PATH_RE.search(rendered):
        errors.append("baseline contains an unredacted sensitive absolute path")
    return errors


def write_baseline_markdown(path: Path, baseline: dict[str, Any]) -> None:
    lines = [
        "# R90-74 Repeated Single-Host Benchmark Baseline",
        "",
        "## Evidence Boundary",
        "",
        f"- Status: `{baseline['status']}`",
        f"- Evidence class: `{baseline['evidence_class']}`",
        f"- Samples: {baseline['sample_count']} sequential uncached complete captures",
        f"- Commit: `{baseline['git']['head']}`",
        f"- Tree: `{baseline['git']['tree']}`",
        f"- Environment fingerprint: `{baseline['environment']['fingerprint_sha256']}`",
        f"- Parameters: C iterations `{baseline['parameters']['c_iterations']}`, Go benchtime `{baseline['parameters']['go_benchtime']}`",
        "- Quartiles: inclusive method; variation: sample standard deviation / absolute mean.",
        "- Threshold applied: false",
        "- Production-derived or portable/cross-host claim: false",
        "- Release, tag, or publication authority: false",
        "",
        "## Samples",
        "",
        "| File | SHA-256 | Start | End |",
        "| --- | --- | --- | --- |",
    ]
    for sample in baseline["samples"]:
        lines.append(
            f"| `{sample['file']}` | `{sample['sha256']}` | {sample['start']} | {sample['end']} |"
        )
    lines.extend(["", "## Descriptive Results", ""])
    for family, title in (("c", "C Benchmarks"), ("go", "Go Benchmarks")):
        lines.extend(
            [
                f"### {title}",
                "",
                "| Case | Metric | Median | Inclusive IQR | CV % | Min | Max |",
                "| --- | --- | ---: | ---: | ---: | ---: | ---: |",
            ]
        )
        for case in sorted(baseline["metrics"][family]):
            for metric in sorted(baseline["metrics"][family][case]):
                summary = baseline["metrics"][family][case][metric]
                lines.append(
                    f"| `{case}` | `{metric}` | {_display_number(summary['median'])} | "
                    f"{_display_number(summary['iqr_inclusive'])} | "
                    f"{_display_number(summary['coefficient_of_variation_percent'])} | "
                    f"{_display_number(summary['minimum'])} | {_display_number(summary['maximum'])} |"
                )
        lines.append("")
    lines.extend(
        [
            "## Interpretation",
            "",
            "These values describe five sequential observations from one exact clean commit and one unchanged local environment. Every raw sample is retained beside this summary and the aggregate is independently recomputable from their SHA-256-bound contents.",
            "",
            "Variation is reported, not judged against a pass/fail budget. This baseline does not establish production capacity, cross-host portability, an SLO, a release gate, or publication authority.",
        ]
    )
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def command_capture(args: argparse.Namespace) -> int:
    try:
        evidence = build_evidence(args.repo, args.c_iterations, args.go_benchtime)
        output = args.output
        if output is None:
            run_id = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
            output = args.repo / "docs" / "evidence" / "local" / "benchmark" / f"benchmark-evidence-{run_id}.json"
        write_evidence(output, evidence)
    except (EvidenceError, OSError, subprocess.SubprocessError) as error:
        print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    print(f"[benchmark-evidence] ok: {output}")
    return 0


def command_validate(args: argparse.Namespace) -> int:
    try:
        evidence = json.loads(args.input.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    errors = validation_errors(evidence)
    if errors:
        for error in errors:
            print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    print(f"[benchmark-evidence] ok: {args.input}")
    return 0


def command_aggregate(args: argparse.Namespace) -> int:
    try:
        baseline = aggregate_samples(args.samples)
        write_evidence(args.json, baseline)
        write_baseline_markdown(args.markdown, baseline)
    except (EvidenceError, OSError) as error:
        print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    print(f"[benchmark-evidence] baseline ok: {args.json}")
    print(f"[benchmark-evidence] baseline report: {args.markdown}")
    return 0


def command_validate_baseline(args: argparse.Namespace) -> int:
    try:
        baseline = json.loads(args.input.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    errors = baseline_validation_errors(baseline, args.samples)
    if errors:
        for error in errors:
            print(f"[benchmark-evidence] {error}", file=sys.stderr)
        return 1
    print(f"[benchmark-evidence] baseline ok: {args.input}")
    return 0


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(description=__doc__)
    commands = root.add_subparsers(dest="command", required=True)
    capture = commands.add_parser("capture")
    capture.add_argument("--repo", type=Path, default=Path.cwd())
    capture.add_argument("--output", type=Path)
    capture.add_argument("--c-iterations", type=int, default=100_000)
    capture.add_argument("--go-benchtime", default="10s")
    capture.set_defaults(func=command_capture)
    validate = commands.add_parser("validate")
    validate.add_argument("--input", type=Path, required=True)
    validate.set_defaults(func=command_validate)
    aggregate = commands.add_parser("aggregate")
    aggregate.add_argument("--samples", type=Path, nargs="+", required=True)
    aggregate.add_argument("--json", type=Path, required=True)
    aggregate.add_argument("--markdown", type=Path, required=True)
    aggregate.set_defaults(func=command_aggregate)
    validate_baseline = commands.add_parser("validate-baseline")
    validate_baseline.add_argument("--input", type=Path, required=True)
    validate_baseline.add_argument("--samples", type=Path, nargs="+", required=True)
    validate_baseline.set_defaults(func=command_validate_baseline)
    return root


def main() -> int:
    args = parser().parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
