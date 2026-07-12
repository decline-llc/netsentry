#!/usr/bin/env python3
"""Generate an idempotent, substantive Obsidian knowledge note for a Git range."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import pathlib
import re
import subprocess
import sys

ZERO_SHA = "0" * 40
MOC_MARKER_START = "<!-- netsentry-ci-knowledge:start -->"
MOC_MARKER_END = "<!-- netsentry-ci-knowledge:end -->"


def git(repo: pathlib.Path, *args: str) -> str:
    return subprocess.check_output(
        ["git", "-C", str(repo), *args], text=True, stderr=subprocess.PIPE
    ).strip()


def resolve_range(repo: pathlib.Path, before: str, after: str) -> tuple[str, str]:
    after = git(repo, "rev-parse", after)
    if not before or before == ZERO_SHA:
        try:
            before = git(repo, "rev-parse", f"{after}^")
        except subprocess.CalledProcessError:
            before = after
    else:
        before = git(repo, "rev-parse", before)
    return before, after


def changed_files(repo: pathlib.Path, before: str, after: str) -> list[str]:
    if before == after:
        output = git(repo, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", after)
    else:
        output = git(repo, "diff", "--name-only", before, after)
    return sorted(line for line in output.splitlines() if line)


TOPICS = (
    (("capture/",), "抓包解析边界", "pcap/DLT 与 packet length 属于第一信任边界；任何新协议读取都必须先证明 offset 和 captured length 自洽。", "[[02-功能模块拆解/Capture抓包解析]]"),
    (("engine/internal/receiver/",), "IPC 契约", "UDS 的文件权限不是完整输入验证；frame size、Base64、IP、timestamp 和 payload length 必须在入队前 fail-closed。", "[[03-技术栈知识点/Unix Domain Socket JSONL协议]]"),
    (("engine/internal/rule/", "configs/rules"), "规则引擎", "规则先完成集合/MITRE/类型配置校验和编译，再原子发布不可变 snapshot；AC 命中只产生候选。", "[[02-功能模块拆解/Go Engine规则引擎]]"),
    (("engine/internal/alert/",), "告警与存储", "并发 worker 可以并行匹配，但 recovery log 与 SQLite 的 append/read/truncate/write 必须保持串行临界区。", "[[02-功能模块拆解/SQLite告警存储]]"),
    (("engine/internal/api/", "engine/internal/config/", "configs/config"), "API 与配置安全", "HTTP 默认应保持 loopback；非 loopback 管理面必须认证，并具有 connection/header/body 资源上限。", "[[02-功能模块拆解/HTTP API与可观测性]]"),
    (("Makefile", "scripts/", "capture/tests/", "engine/"), "验证门禁", "行为变更至少关联 unit/race；解析/IPC 变更还需 ASan 或 corpus integration；跨组件变更需要 E2E。", "[[06-测试与发布规范/测试矩阵与发布门禁]]"),
    ((".github/",), "CI/CD", "自动化必须保存输入 revision、校验结果和失败边界；生成知识是可审查草稿，不替代代码与测试证据。", "[[02-功能模块拆解/CI-CD与发布流程]]"),
)


def topic_rows(files: list[str]) -> list[tuple[str, str, str]]:
    rows: list[tuple[str, str, str]] = []
    for prefixes, title, insight, link in TOPICS:
        if any(path == prefix or path.startswith(prefix) for path in files for prefix in prefixes):
            rows.append((title, insight, link))
    return rows or [("项目演进", "本次范围未命中已知模块分类，需要在 `_待整理` 中人工补充分层影响。", "[[01-项目核心架构/系统架构与数据流]]")]


def load_rules(repo: pathlib.Path) -> list[dict]:
    path = repo / "configs" / "rules.json"
    if not path.is_file():
        return []
    payload = json.loads(path.read_text(encoding="utf-8"))
    return payload.get("rules", payload if isinstance(payload, list) else [])


def mitre_rows(rules: list[dict]) -> list[tuple[str, str, str, str]]:
    rows = set()
    for rule in rules:
        for technique in rule.get("mitre_techniques", []):
            rows.add((
                str(technique.get("technique_id", "")),
                str(technique.get("tactic", "")),
                str(technique.get("technique_name", "")),
                str(rule.get("id", "")),
            ))
    return sorted(rows)


def config_value(repo: pathlib.Path, key: str, default: str) -> str:
    path = repo / "configs" / "config.yaml"
    if not path.is_file():
        return default
    pattern = re.compile(rf"^\s*{re.escape(key)}:\s*(.*?)\s*$")
    for line in path.read_text(encoding="utf-8").splitlines():
        match = pattern.match(line)
        if match:
            return match.group(1).strip('"\'')
    return default


def render_note(repo: pathlib.Path, files: list[str], before: str, after: str) -> tuple[str, str, str]:
    short = after[:10]
    date = git(repo, "show", "-s", "--format=%cs", after) or dt.date.today().isoformat()
    subject = git(repo, "show", "-s", "--format=%s", after).replace("\n", " ")
    topics = topic_rows(files)
    rules = load_rules(repo)
    mappings = mitre_rows(rules)
    links = sorted({row[2] for row in topics})
    filename = f"{date}-{short}-CI知识同步.md"
    relative = f"04-开发迭代记录/{filename}"

    lines = [
        "---",
        "分类: 开发迭代记录",
        "标签: [ci, knowledge-sync, architecture, audit]",
        "关联模块: [capture, engine, rules, storage, api, ci]",
        f"最后更新时间: {date}",
        f"对应代码相对路径: [git range {before[:10]}..{short}]",
        f"版本标记: {short}",
        "---",
        "",
        f"# CI 知识同步：{date} / {short}",
        "",
        f"提交主题：{subject}",
        "",
        "本笔记由确定性脚本从代码范围、当前配置与规则 schema 提取。它记录可复核的不变量和受影响知识面，不把 commit message 当作技术结论。涉及设计取舍的内容仍需在 `_待整理` 复核后合并到稳定笔记。",
        "",
        "## 数据流基线",
        "",
        "```mermaid",
        "flowchart LR",
        "    P[pcap / live frame] --> C[C capture]",
        "    C -->|UDS JSONL| R[Go receiver]",
        "    R --> W[worker pool]",
        "    W --> E[immutable ruleState]",
        "    E --> D[(SQLite / recovery log)]",
        "    D --> A[HTTP API / metrics]",
        "```",
        "",
        "## 受影响的知识主题",
        "",
    ]
    for title, insight, link in topics:
        lines.append(f"- **{title}**：{insight} {link}")

    lines += [
        "",
        "## 当前运行不变量",
        "",
        f"- API 默认监听：`{config_value(repo, 'api_listen_host', 'unknown')}:{config_value(repo, 'api_port', 'unknown')}`。",
        f"- API mutation 认证默认值：`api_auth_enabled={config_value(repo, 'api_auth_enabled', 'unknown')}`；非 loopback 必须在配置校验中 fail-closed。",
        f"- 当前 seed rule 数：{len(rules)}；MITRE tuple 数：{len(mappings)}。",
        "- 网络 signature 代表与 technique 一致的 indicator，不等价于确认 technique 已成功执行。",
        "",
        "## 当前 MITRE 映射快照",
        "",
        "| Technique | Tactic | Canonical name | Rule |",
        "|---|---|---|---|",
    ]
    if mappings:
        lines.extend(f"| {technique} | {tactic} | {name} | {rule_id} |" for technique, tactic, name, rule_id in mappings)
    else:
        lines.append("| - | - | 未检测到 rule mapping | - |")

    lines += [
        "",
        "## 变更范围（用于定位，不作为知识结论）",
        "",
    ]
    lines.extend(f"- `{path}`" for path in files)
    lines += [
        "",
        "## 人工复核问题",
        "",
        "- 信任边界、资源上限或数据格式是否变化？",
        "- rule mapping 是否仍与 ATT&CK tactic/name 和网络证据强度一致？",
        "- 是否新增已知绕过、迁移需求、性能预算或 rollback 条件？",
        "- 哪些结论应合并进架构/模块/技术点稳定笔记？",
        "",
        "## 关联",
        "",
        " · ".join(links + ["[[00-MOC/NetSentry知识总览]]"]),
        "",
        "## 代码引用",
        "",
        f"`git diff {before[:10]}..{short}`；`configs/config.yaml`；`configs/rules.json`；`scripts/sync_knowledge.py`。",
        "",
    ]
    return relative, "\n".join(lines), subject


def update_moc(vault: pathlib.Path, relative_note: str, after: str, subject: str) -> None:
    moc = vault / "00-MOC" / "NetSentry知识总览.md"
    moc.parent.mkdir(parents=True, exist_ok=True)
    text = moc.read_text(encoding="utf-8") if moc.exists() else "# NetSentry 知识总览\n"
    link = f"[[{relative_note[:-3]}]]"
    new_entry = f"- `{after[:10]}` {subject}：{link}"
    if MOC_MARKER_START in text and MOC_MARKER_END in text:
        before_marker, tail = text.split(MOC_MARKER_START, 1)
        marker_body, after_marker = tail.split(MOC_MARKER_END, 1)
        entries = [line for line in marker_body.strip().splitlines() if line.startswith("- ")]
        entries = [entry for entry in entries if link not in entry]
        entries = ([new_entry] + entries)[:20]
        block = MOC_MARKER_START + "\n\n" + "\n".join(entries) + "\n\n" + MOC_MARKER_END
        text = before_marker.rstrip() + "\n\n" + block + after_marker
    else:
        text = text.rstrip() + "\n\n## CI 知识同步入口\n\n" + MOC_MARKER_START + "\n\n" + new_entry + "\n\n" + MOC_MARKER_END + "\n"
    moc.write_text(text, encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=pathlib.Path, default=pathlib.Path(__file__).resolve().parents[1])
    parser.add_argument("--vault", type=pathlib.Path, required=True)
    parser.add_argument("--before", default="HEAD^")
    parser.add_argument("--after", default="HEAD")
    args = parser.parse_args()

    repo = args.repo.resolve()
    vault = args.vault.resolve()
    before, after = resolve_range(repo, args.before, args.after)
    files = changed_files(repo, before, after)
    relative, note, subject = render_note(repo, files, before, after)
    note_path = vault / relative
    note_path.parent.mkdir(parents=True, exist_ok=True)
    note_path.write_text(note, encoding="utf-8")
    update_moc(vault, relative, after, subject)
    print(json.dumps({"status": "ok", "range": f"{before}..{after}", "note": relative, "changed_files": len(files)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, json.JSONDecodeError, subprocess.CalledProcessError) as exc:
        print(f"knowledge sync failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
