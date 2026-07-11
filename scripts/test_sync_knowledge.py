#!/usr/bin/env python3

import pathlib
import subprocess
import tempfile
import unittest


class KnowledgeSyncTest(unittest.TestCase):
    def test_generates_idempotent_note_and_moc_link(self):
        repo = pathlib.Path(__file__).resolve().parents[1]
        script = repo / "scripts" / "sync_knowledge.py"
        after = subprocess.check_output(["git", "-C", str(repo), "rev-parse", "HEAD"], text=True).strip()
        before = subprocess.check_output(["git", "-C", str(repo), "rev-parse", "HEAD^"], text=True).strip()
        with tempfile.TemporaryDirectory() as directory:
            vault = pathlib.Path(directory)
            command = [
                "python3", str(script), "--repo", str(repo), "--vault", str(vault),
                "--before", before, "--after", after,
            ]
            subprocess.run(command, check=True, capture_output=True, text=True)
            subprocess.run(command, check=True, capture_output=True, text=True)

            notes = list((vault / "04-开发迭代记录").glob("*-CI知识同步.md"))
            self.assertEqual(len(notes), 1)
            note = notes[0].read_text(encoding="utf-8")
            for value in ("分类:", "最后更新时间:", "版本标记:", "## 数据流基线", "## 当前 MITRE 映射快照", "## 关联", "## 代码引用"):
                self.assertIn(value, note)
            self.assertNotIn(str(repo), note)

            moc = (vault / "00-MOC" / "NetSentry知识总览.md").read_text(encoding="utf-8")
            self.assertEqual(moc.count("<!-- netsentry-ci-knowledge:start -->"), 1)
            self.assertEqual(moc.count("CI知识同步]]"), 1)


if __name__ == "__main__":
    unittest.main()
