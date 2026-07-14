#!/usr/bin/env python3
"""Versioned entrypoint for the local post-push knowledge synchronizer."""

from __future__ import annotations

import os
import pathlib
import subprocess
import sys

from sync_knowledge import sync_range


def sync(repo: pathlib.Path, vault: pathlib.Path, range_spec: str) -> dict[str, object]:
    """Synchronize an explicit ``before..after`` range without Git-hook state."""
    before, separator, after = range_spec.partition("..")
    if not separator or not before or not after:
        raise ValueError("sync range must be before..after")
    result = sync_range(repo, vault, before, after)
    write_full_index(repo, vault)
    return result


def write_full_index(repo: pathlib.Path, vault: pathlib.Path) -> None:
    """Rebuild the local Vault's commit index from versioned Python code."""
    rows = subprocess.check_output(
        ["git", "-C", str(repo), "log", "--all", "--date=short", "--format=%H%x09%cs%x09%s"],
        text=True,
    ).splitlines()
    lines = [
        "---", "分类: 开发迭代记录", "标签: [git, history, auto-sync]",
        "关联模块: [all]", "对应代码相对路径: [git log --all]", "---", "",
        "# 全量提交索引", "",
        "| 日期 | Commit | 提交说明 |", "|---|---|---|",
    ]
    for row in rows:
        commit, date, subject = row.split("\t", 2)
        lines.append(f"| {date} | {commit[:10]} | {subject.replace('|', '\\|')} |")
    index = vault / "04-开发迭代记录" / "全量提交索引.md"
    index.parent.mkdir(parents=True, exist_ok=True)
    index.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main(argv: list[str] | None = None) -> int:
    argv = list(sys.argv[1:] if argv is None else argv)
    if argv[:1] != ["sync"]:
        raise ValueError("only explicit 'sync' mode is supported")
    repo = pathlib.Path.cwd()
    vault = pathlib.Path(os.environ.get(
        "NETSENTRY_VAULT", "/home/virtual-machine/Desktop/NetSentry-Knowledge"
    )).expanduser()
    range_spec = os.environ.get("NETSENTRY_SYNC_RANGE", "")
    if not range_spec:
        before = os.popen("git rev-parse HEAD~1").read().strip()
        after = os.popen("git rev-parse HEAD").read().strip()
        range_spec = f"{before}..{after}"
    sync(repo, vault, range_spec)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError) as exc:
        print(f"knowledge sync failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
