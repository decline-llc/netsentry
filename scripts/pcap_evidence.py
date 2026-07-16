#!/usr/bin/env python3
"""Generate and validate reviewed PCAP evidence manifests without leaking paths."""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import re
import struct
import sys
from typing import Any


SCHEMA_VERSION = 1
SHA256 = re.compile(r"^[0-9a-f]{64}$")
REVIEW_NAMES = ("provenance", "privacy", "sanitization", "sensitive_metadata")
PLACEHOLDERS = ("todo", "tbd", "xxx", "replace me", "replace with", "placeholder")
FORBIDDEN_PUBLIC_TEXT = ("/home/", "/tmp/", "-----begin ", "ghp_", "github_pat_")
PCAP_MAGICS = {
    b"\xd4\xc3\xb2\xa1": "<",
    b"\xa1\xb2\xc3\xd4": ">",
    b"\x4d\x3c\xb2\xa1": "<",
    b"\xa1\xb2\x3c\x4d": ">",
}
PCAPNG_MAGIC = b"\x0a\x0d\x0d\x0a"
PCAPNG_PACKET_BLOCKS = {0x00000002, 0x00000003, 0x00000006}


class EvidenceError(RuntimeError):
    """A deterministic evidence contract or integrity failure."""


def require(condition: bool, message: str) -> None:
    if not condition:
        raise EvidenceError(message)


def load_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        value = json.load(handle)
    require(isinstance(value, dict), f"{path}: top-level JSON must be an object")
    return value


def parse_utc(value: str, label: str) -> datetime:
    try:
        parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except (AttributeError, ValueError) as exc:
        raise EvidenceError(f"{label} must be ISO-8601 UTC") from exc
    require(parsed.tzinfo is not None, f"{label} must include a UTC offset")
    return parsed.astimezone(timezone.utc)


def meaningful(value: Any, label: str) -> str:
    require(isinstance(value, str) and value.strip(), f"{label} must be non-empty")
    normalized = value.strip()
    lowered = normalized.casefold()
    require(not any(marker in lowered for marker in PLACEHOLDERS), f"{label} contains a placeholder")
    return normalized


def reject_sensitive_text(value: Any, label: str = "manifest") -> None:
    if isinstance(value, dict):
        for key, item in value.items():
            reject_sensitive_text(item, f"{label}.{key}")
    elif isinstance(value, list):
        for index, item in enumerate(value):
            reject_sensitive_text(item, f"{label}[{index}]")
    elif isinstance(value, str):
        lowered = value.casefold()
        for forbidden in FORBIDDEN_PUBLIC_TEXT:
            require(forbidden not in lowered, f"{label} contains forbidden private or credential-like text")


def packet_count_pcap(path: Path) -> int:
    with path.open("rb") as handle:
        header = handle.read(24)
        require(len(header) == 24, f"{path.name}: truncated pcap header")
        endian = PCAP_MAGICS.get(header[:4])
        require(endian is not None, f"{path.name}: unsupported pcap magic")
        record = struct.Struct(f"{endian}IIII")
        count = 0
        while True:
            record_header = handle.read(record.size)
            if not record_header:
                return count
            require(len(record_header) == record.size, f"{path.name}: truncated pcap record header")
            _ts_sec, _ts_fraction, captured_length, _original_length = record.unpack(record_header)
            require(
                len(handle.read(captured_length)) == captured_length,
                f"{path.name}: truncated pcap packet",
            )
            count += 1


def packet_count_pcapng(path: Path) -> int:
    data = path.read_bytes()
    require(len(data) >= 28 and data[:4] == PCAPNG_MAGIC, f"{path.name}: invalid pcapng header")
    offset = 0
    endian: str | None = None
    count = 0
    while offset < len(data):
        require(offset + 12 <= len(data), f"{path.name}: truncated pcapng block")
        block_type_raw = data[offset : offset + 4]
        if block_type_raw == PCAPNG_MAGIC:
            byte_order = data[offset + 8 : offset + 12]
            if byte_order == b"\x4d\x3c\x2b\x1a":
                endian = "<"
            elif byte_order == b"\x1a\x2b\x3c\x4d":
                endian = ">"
            else:
                raise EvidenceError(f"{path.name}: invalid pcapng byte-order magic")
        require(endian is not None, f"{path.name}: packet block appears before section header")
        block_type, block_length = struct.unpack_from(f"{endian}II", data, offset)
        require(block_length >= 12 and block_length % 4 == 0, f"{path.name}: invalid pcapng block length")
        require(offset + block_length <= len(data), f"{path.name}: truncated pcapng block body")
        trailing_length = struct.unpack_from(f"{endian}I", data, offset + block_length - 4)[0]
        require(trailing_length == block_length, f"{path.name}: pcapng block length mismatch")
        if block_type in PCAPNG_PACKET_BLOCKS:
            count += 1
        offset += block_length
    require(offset == len(data), f"{path.name}: trailing pcapng bytes")
    return count


def packet_count(path: Path) -> int:
    with path.open("rb") as handle:
        magic = handle.read(4)
    if magic in PCAP_MAGICS:
        return packet_count_pcap(path)
    if magic == PCAPNG_MAGIC:
        return packet_count_pcapng(path)
    raise EvidenceError(f"{path.name}: expected pcap or pcapng input")


def corpus_files(corpus: Path) -> list[Path]:
    require(corpus.exists(), f"corpus does not exist: {corpus}")
    if corpus.is_file():
        files = [corpus]
    else:
        files = sorted(
            path
            for path in corpus.rglob("*")
            if path.is_file() and path.suffix.casefold() in {".pcap", ".pcapng"}
        )
    require(bool(files), f"no pcap or pcapng files found: {corpus}")
    return files


def file_record(path: Path, root: Path) -> dict[str, Any]:
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    name = path.name if root.is_file() else path.relative_to(root).as_posix()
    return {
        "name": name,
        "bytes": path.stat().st_size,
        "packet_count": packet_count(path),
        "sha256": digest,
    }


def inventory(corpus: Path) -> list[dict[str, Any]]:
    return [file_record(path, corpus) for path in corpus_files(corpus)]


def validate_review(review: Any, label: str, require_approved: bool) -> None:
    require(isinstance(review, dict), f"reviews.{label} must be an object")
    status = review.get("status")
    require(status in {"pending", "approved", "rejected"}, f"reviews.{label}.status is invalid")
    if require_approved:
        require(status == "approved", f"reviews.{label}.status must be approved")
    if status == "approved":
        meaningful(review.get("reviewer"), f"reviews.{label}.reviewer")
        reviewed_at = parse_utc(review.get("reviewed_utc"), f"reviews.{label}.reviewed_utc")
        require(
            reviewed_at <= datetime.now(timezone.utc),
            f"reviews.{label}.reviewed_utc must not be in the future",
        )
    meaningful(review.get("notes"), f"reviews.{label}.notes")


def validate_manifest_data(
    manifest: dict[str, Any],
    *,
    corpus: Path | None = None,
    require_approved: bool = False,
) -> list[str]:
    errors: list[str] = []
    try:
        reject_sensitive_text(manifest)
        require(manifest.get("schema_version") == SCHEMA_VERSION, "schema_version must be 1")
        require(manifest.get("evidence_class") == "production-derived", "evidence_class must be production-derived")
        generated = parse_utc(manifest.get("generated_utc"), "generated_utc")
        require(generated <= datetime.now(timezone.utc), "generated_utc must not be in the future")
        require(manifest.get("path_redacted") is True, "path_redacted must be true")

        artifacts = manifest.get("artifacts")
        require(isinstance(artifacts, list) and artifacts, "artifacts must be a non-empty list")
        seen: set[str] = set()
        for index, artifact in enumerate(artifacts):
            require(isinstance(artifact, dict), f"artifacts[{index}] must be an object")
            name = meaningful(artifact.get("name"), f"artifacts[{index}].name")
            pure_name = Path(name)
            require(not pure_name.is_absolute() and ".." not in pure_name.parts, f"artifacts[{index}].name is unsafe")
            require(name not in seen, f"duplicate artifact name: {name}")
            seen.add(name)
            require(isinstance(artifact.get("bytes"), int) and artifact["bytes"] > 0, f"{name}: bytes must be positive")
            require(
                isinstance(artifact.get("packet_count"), int) and artifact["packet_count"] > 0,
                f"{name}: packet_count must be positive",
            )
            require(SHA256.fullmatch(artifact.get("sha256", "")) is not None, f"{name}: sha256 is invalid")

        require(
            manifest.get("file_count") == len(artifacts),
            "file_count must match artifacts",
        )
        require(
            manifest.get("packet_count") == sum(item["packet_count"] for item in artifacts),
            "packet_count must match artifacts",
        )

        provenance = manifest.get("provenance")
        require(isinstance(provenance, dict), "provenance must be an object")
        for field in ("source_description", "collection_node", "collection_start_utc", "collection_end_utc", "custodian_reference"):
            meaningful(provenance.get(field), f"provenance.{field}")
        start = parse_utc(provenance["collection_start_utc"], "provenance.collection_start_utc")
        end = parse_utc(provenance["collection_end_utc"], "provenance.collection_end_utc")
        require(end >= start, "provenance collection end precedes start")
        require(end <= datetime.now(timezone.utc), "provenance collection end must not be in the future")

        sanitization = manifest.get("sanitization")
        require(isinstance(sanitization, dict), "sanitization must be an object")
        meaningful(sanitization.get("tool"), "sanitization.tool")
        rules = sanitization.get("rules")
        require(isinstance(rules, list) and rules, "sanitization.rules must be a non-empty list")
        for index, rule in enumerate(rules):
            meaningful(rule, f"sanitization.rules[{index}]")
        require(
            sanitization.get("raw_capture_storage") == "internal-only",
            "sanitization.raw_capture_storage must be internal-only",
        )

        handling = manifest.get("sensitive_metadata_handling")
        require(isinstance(handling, dict), "sensitive_metadata_handling must be an object")
        removed = handling.get("removed_fields")
        require(isinstance(removed, list) and removed, "sensitive_metadata_handling.removed_fields must be non-empty")
        for index, field in enumerate(removed):
            meaningful(field, f"sensitive_metadata_handling.removed_fields[{index}]")
        meaningful(handling.get("residual_risk"), "sensitive_metadata_handling.residual_risk")

        reviews = manifest.get("reviews")
        require(isinstance(reviews, dict), "reviews must be an object")
        for name in REVIEW_NAMES:
            validate_review(reviews.get(name), name, require_approved)

        status = manifest.get("review_status")
        require(status in {"pending", "approved", "rejected"}, "review_status is invalid")
        if require_approved:
            require(status == "approved", "review_status must be approved")
        if status == "approved":
            approval = manifest.get("approval")
            require(isinstance(approval, dict), "approval must be an object when approved")
            meaningful(approval.get("reviewer"), "approval.reviewer")
            approved_at = parse_utc(approval.get("reviewed_utc"), "approval.reviewed_utc")
            require(approved_at <= datetime.now(timezone.utc), "approval.reviewed_utc must not be in the future")
            require(approval.get("decision") == "approved", "approval.decision must be approved")
            for name in REVIEW_NAMES:
                require(reviews[name]["status"] == "approved", f"reviews.{name}.status must be approved")

        if corpus is not None:
            actual = inventory(corpus)
            require(actual == artifacts, "corpus files, packet counts, byte sizes, or SHA-256 values differ from manifest")
    except (EvidenceError, OSError, json.JSONDecodeError) as exc:
        errors.append(str(exc))
    return errors


def generate(corpus: Path, metadata_path: Path, output_dir: Path) -> tuple[Path, Path]:
    metadata = load_json(metadata_path)
    artifacts = inventory(corpus)
    manifest = {
        "schema_version": SCHEMA_VERSION,
        "generated_utc": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "evidence_class": "production-derived",
        "path_redacted": True,
        "file_count": len(artifacts),
        "packet_count": sum(item["packet_count"] for item in artifacts),
        "artifacts": artifacts,
        "provenance": metadata.get("provenance"),
        "sanitization": metadata.get("sanitization"),
        "sensitive_metadata_handling": metadata.get("sensitive_metadata_handling"),
        "reviews": metadata.get("reviews"),
        "review_status": metadata.get("review_status", "pending"),
        "approval": metadata.get("approval"),
    }
    errors = validate_manifest_data(manifest)
    require(not errors, errors[0] if errors else "manifest validation failed")

    output_dir.mkdir(parents=True, exist_ok=True)
    manifest_path = output_dir / "pcap-evidence-manifest.json"
    report_path = output_dir / "pcap-evidence-report.md"
    manifest_path.write_text(json.dumps(manifest, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    report_lines = [
        "# PCAP Evidence Review",
        "",
        "> Generated facts are not approval. Review status is copied from the human-reviewed metadata input.",
        "",
        "## Corpus Inventory",
        "",
        f"- Evidence class: {manifest['evidence_class']}",
        "- Corpus path included: no",
        f"- Files: {manifest['file_count']}",
        f"- Packets: {manifest['packet_count']}",
        f"- Review status: {manifest['review_status']}",
        "",
        "| File | Bytes | Packets | SHA-256 |",
        "| --- | ---: | ---: | --- |",
    ]
    for artifact in artifacts:
        report_lines.append(
            f"| `{artifact['name']}` | {artifact['bytes']} | {artifact['packet_count']} | `{artifact['sha256']}` |"
        )
    report_lines.extend(
        [
            "",
            "## Provenance",
            "",
            f"- Source: {manifest['provenance']['source_description']}",
            f"- Collection node: {manifest['provenance']['collection_node']}",
            f"- Collection window: {manifest['provenance']['collection_start_utc']} to {manifest['provenance']['collection_end_utc']}",
            f"- Custodian reference: {manifest['provenance']['custodian_reference']}",
            "",
            "## Sanitization",
            "",
            f"- Tool: {manifest['sanitization']['tool']}",
            f"- Rules: {'; '.join(manifest['sanitization']['rules'])}",
            f"- Raw capture storage: {manifest['sanitization']['raw_capture_storage']}",
            "",
            "## Sensitive Metadata Handling",
            "",
            f"- Removed fields: {'; '.join(manifest['sensitive_metadata_handling']['removed_fields'])}",
            f"- Residual risk: {manifest['sensitive_metadata_handling']['residual_risk']}",
            "",
            "## Reviews",
            "",
        ]
    )
    for name in REVIEW_NAMES:
        review = manifest["reviews"][name]
        report_lines.append(f"- {name}: {review['status']} — {review['notes']}")
    report_path.write_text("\n".join(report_lines) + "\n", encoding="utf-8")
    return manifest_path, report_path


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    generate_parser = subparsers.add_parser("generate")
    generate_parser.add_argument("--pcap", required=True, type=Path)
    generate_parser.add_argument("--metadata", required=True, type=Path)
    generate_parser.add_argument("--output-dir", required=True, type=Path)

    validate_parser = subparsers.add_parser("validate")
    validate_parser.add_argument("--manifest", required=True, type=Path)
    validate_parser.add_argument("--pcap", type=Path)
    validate_parser.add_argument("--require-approved", action="store_true")

    args = parser.parse_args()
    try:
        if args.command == "generate":
            manifest_path, report_path = generate(args.pcap, args.metadata, args.output_dir)
            print(f"[pcap-evidence] wrote {manifest_path}")
            print(f"[pcap-evidence] wrote {report_path}")
            return 0
        manifest = load_json(args.manifest)
        errors = validate_manifest_data(
            manifest,
            corpus=args.pcap,
            require_approved=args.require_approved,
        )
        if errors:
            raise EvidenceError(errors[0])
        print(f"[pcap-evidence] valid: {args.manifest}")
        return 0
    except (EvidenceError, OSError, json.JSONDecodeError) as exc:
        print(f"[pcap-evidence] failed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
