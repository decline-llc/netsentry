#!/usr/bin/env python3
"""Validate NetSentry's immutable CI, Go, tooling, and external-fixture locks."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile
import urllib.parse
import urllib.request


ROOT = pathlib.Path(__file__).resolve().parents[1]
LOCK_PATH = ROOT / ".github" / "supply-chain-lock.json"
SHA40 = re.compile(r"^[0-9a-f]{40}$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")
VERSION = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+$")
USES = re.compile(r"^\s*uses:\s*([^@\s]+)@([^\s#]+)(?:\s+#\s*(\S+))?\s*$")


class PolicyError(RuntimeError):
    """A deterministic supply-chain policy violation."""


def require(condition: bool, message: str) -> None:
    if not condition:
        raise PolicyError(message)


def load_json(path: pathlib.Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        value = json.load(handle)
    require(isinstance(value, dict), f"{path}: top-level JSON must be an object")
    return value


def safe_relative(value: str, label: str) -> pathlib.PurePosixPath:
    path = pathlib.PurePosixPath(value)
    require(value == path.as_posix(), f"{label}: path must use canonical POSIX form")
    require(not path.is_absolute() and ".." not in path.parts, f"{label}: unsafe path {value!r}")
    require(bool(path.parts), f"{label}: empty path")
    return path


def validate_actions(lock: dict) -> tuple[int, int]:
    entries = lock.get("actions")
    require(isinstance(entries, list) and entries, "lock.actions must be a non-empty list")
    locked: dict[str, dict] = {}
    for entry in entries:
        name = entry.get("uses", "")
        require(name and name not in locked, f"duplicate or empty Action lock: {name!r}")
        require(SHA40.fullmatch(entry.get("commit", "")) is not None, f"{name}: invalid commit")
        require(re.fullmatch(r"v\d+\.\d+\.\d+", entry.get("version", "")) is not None, f"{name}: version must be exact SemVer tag")
        require(entry.get("runtime") == "node24", f"{name}: reviewed runtime must be node24")
        require(entry.get("source") == f"https://github.com/{name}/releases/tag/{entry['version']}", f"{name}: release source mismatch")
        locked[name] = entry

    tools = lock.get("tools")
    require(isinstance(tools, list) and tools, "lock.tools must be a non-empty list")
    install_lines = []
    for tool in tools:
        require(SHA40.fullmatch(tool.get("commit", "")) is not None, f"{tool.get('module')}: invalid tool commit")
        require(re.fullmatch(r"v\d+\.\d+\.\d+", tool.get("version", "")) is not None, f"{tool.get('module')}: tool version must be exact")
        source = urllib.parse.urlparse(tool.get("source", ""))
        require(source.scheme == "https" and source.netloc == "github.com", f"{tool.get('module')}: invalid source")
        install_lines.append(f"go install {tool['module']}@{tool['version']}")

    workflows = sorted((ROOT / ".github" / "workflows").glob("*.y*ml"))
    require(bool(workflows), "no GitHub Actions workflows found")
    seen: set[str] = set()
    uses_count = 0
    for workflow in workflows:
        text = workflow.read_text(encoding="utf-8")
        for line_number, line in enumerate(text.splitlines(), 1):
            match = USES.match(line)
            if not match:
                require(
                    re.match(r"^\s*uses\s*:", line) is None,
                    f"{workflow}:{line_number}: unsupported uses syntax; use an unquoted owner/repository@commit",
                )
                continue
            name, ref, comment = match.groups()
            if name.startswith("./"):
                continue
            uses_count += 1
            require(name in locked, f"{workflow}:{line_number}: unreviewed Action {name}")
            expected = locked[name]
            require(SHA40.fullmatch(ref) is not None, f"{workflow}:{line_number}: Action ref is not an immutable commit")
            require(ref == expected["commit"], f"{workflow}:{line_number}: {name} commit differs from lock")
            require(comment == expected["version"], f"{workflow}:{line_number}: {name} version comment differs from lock")
            seen.add(name)
        if "make rc-check" in text:
            require("make supply-chain-check" in text, f"{workflow}: rc-check lacks supply-chain-check")
            for install in install_lines:
                require(install in text, f"{workflow}: missing pinned tool install {install}")

    missing = sorted(set(locked) - seen)
    require(not missing, f"locked Actions are unused: {', '.join(missing)}")
    return len(locked), uses_count


def validate_go_policy(lock: dict) -> str:
    policy = lock.get("go")
    require(isinstance(policy, dict), "lock.go must be an object")
    language = policy.get("language_version", "")
    toolchain = policy.get("ci_toolchain", "")
    require(VERSION.fullmatch(language) is not None, "Go language version must include an exact patch")
    require(VERSION.fullmatch(toolchain) is not None, "Go toolchain version must include an exact patch")
    require(toolchain in policy.get("supported_release_lines_at_review", []), "CI toolchain not in reviewed supported releases")
    require(policy.get("source") == "https://go.dev/dl/?mode=json", "Go release source must be go.dev")

    go_mod = (ROOT / "engine" / "go.mod").read_text(encoding="utf-8")
    language_match = re.search(r"^go\s+(\d+\.\d+\.\d+)\s*$", go_mod, re.MULTILINE)
    toolchain_match = re.search(r"^toolchain\s+go(\d+\.\d+\.\d+)\s*$", go_mod, re.MULTILINE)
    require(language_match is not None, "engine/go.mod must have an exact go directive")
    require(toolchain_match is not None, "engine/go.mod must have an exact toolchain directive")
    require(language_match.group(1) == language, "go directive differs from supply-chain lock")
    require(toolchain_match.group(1) == toolchain, "toolchain directive differs from supply-chain lock")

    environment = os.environ.copy()
    environment.setdefault("GOTOOLCHAIN", "auto")
    runtime = subprocess.check_output(
        ["go", "env", "GOVERSION"], cwd=ROOT / "engine", env=environment, text=True
    ).strip()
    require(runtime == f"go{toolchain}", f"runtime {runtime} differs from locked go{toolchain}")
    return runtime


def validate_manifest(lock: dict) -> tuple[pathlib.Path, dict]:
    policy = lock.get("external_assets")
    require(isinstance(policy, dict), "lock.external_assets must be an object")
    relative = safe_relative(policy.get("manifest", ""), "external manifest")
    manifest_path = ROOT / relative
    manifest = load_json(manifest_path)
    require(manifest.get("schema_version") == 1, "external manifest schema_version must be 1")

    handling = manifest.get("policy", {})
    require(handling.get("inert_test_data_only") is True, "external assets must be inert test data")
    require(handling.get("execute_or_replay_on_network") is False, "network replay must be prohibited")
    require(handling.get("integrity") == "sha256", "external asset integrity must be sha256")

    sources = manifest.get("sources")
    files = manifest.get("files")
    require(isinstance(sources, dict) and sources, "external manifest sources missing")
    require(isinstance(files, list), "external manifest files must be a list")
    require(len(files) == policy.get("expected_files"), "external file count differs from lock")

    source_paths: dict[str, str] = {}
    for name, source in sources.items():
        revision = source.get("revision", "")
        require(SHA40.fullmatch(revision) is not None, f"{name}: revision must be a full commit")
        repository = urllib.parse.urlparse(source.get("repository", ""))
        require(repository.scheme == "https" and repository.netloc == "github.com", f"{name}: repository must be GitHub HTTPS")
        require(bool(source.get("license")), f"{name}: license declaration missing")
        license_path = safe_relative(source.get("license_file", ""), f"{name} license")
        license_url = urllib.parse.urlparse(source.get("license_url", ""))
        require(license_url.scheme == "https" and license_url.netloc == "raw.githubusercontent.com", f"{name}: license URL must be pinned raw GitHub content")
        require(revision in license_url.path, f"{name}: license URL is not pinned to revision")
        source_paths[name] = license_path.as_posix()

    seen_paths: set[str] = set()
    entries_by_path: dict[str, dict] = {}
    for entry in files:
        path = safe_relative(entry.get("path", ""), "external file").as_posix()
        require(path not in seen_paths, f"duplicate external path: {path}")
        seen_paths.add(path)
        entries_by_path[path] = entry
        source_name = entry.get("source", "")
        require(source_name in sources, f"{path}: unknown source {source_name}")
        require(SHA256.fullmatch(entry.get("sha256", "")) is not None, f"{path}: invalid sha256")
        require(isinstance(entry.get("bytes"), int) and entry["bytes"] > 0, f"{path}: invalid byte count")
        require(bool(entry.get("purpose")), f"{path}: purpose missing")
        url = urllib.parse.urlparse(entry.get("url", ""))
        require(url.scheme == "https" and url.netloc == "raw.githubusercontent.com", f"{path}: URL must be raw GitHub HTTPS")
        require(not url.query and not url.fragment, f"{path}: URL query/fragment prohibited")
        require(sources[source_name]["revision"] in url.path, f"{path}: URL not pinned to source revision")

    for source_name, license_path in source_paths.items():
        require(license_path in entries_by_path, f"{source_name}: license file not integrity-locked")
        license_entry = entries_by_path[license_path]
        require(license_entry.get("source") == source_name, f"{source_name}: license source mismatch")
        require(license_entry.get("url") == sources[source_name]["license_url"], f"{source_name}: license URL mismatch")
    return manifest_path, manifest


def fetch_and_verify(manifest: dict) -> int:
    with tempfile.TemporaryDirectory(prefix="netsentry-external-assets-") as directory:
        root = pathlib.Path(directory)
        for entry in manifest["files"]:
            request = urllib.request.Request(
                entry["url"], headers={"User-Agent": "NetSentry-supply-chain-check/1"}
            )
            with urllib.request.urlopen(request, timeout=30) as response:
                payload = response.read(entry["bytes"] + 1)
            require(len(payload) == entry["bytes"], f"{entry['path']}: fetched byte count mismatch")
            digest = hashlib.sha256(payload).hexdigest()
            require(digest == entry["sha256"], f"{entry['path']}: fetched sha256 mismatch")
            target = root / safe_relative(entry["path"], "fetched asset")
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(payload)
            print(f"[supply-chain] PASS {entry['path']} bytes={len(payload)} sha256={digest}")
        return len(manifest["files"])


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--fetch-assets",
        action="store_true",
        help="download all locked fixtures/licenses to an ephemeral directory and verify bytes/hash",
    )
    args = parser.parse_args()

    lock = load_json(LOCK_PATH)
    require(lock.get("schema_version") == 1, "supply-chain lock schema_version must be 1")
    action_count, uses_count = validate_actions(lock)
    runtime = validate_go_policy(lock)
    manifest_path, manifest = validate_manifest(lock)
    fetched = fetch_and_verify(manifest) if args.fetch_assets else 0
    print(
        "[supply-chain] ok: "
        f"actions={action_count} uses={uses_count} runtime={runtime} "
        f"manifest={manifest_path.relative_to(ROOT)} assets={len(manifest['files'])} fetched={fetched}"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, PolicyError, json.JSONDecodeError, subprocess.CalledProcessError) as exc:
        print(f"supply-chain check failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
