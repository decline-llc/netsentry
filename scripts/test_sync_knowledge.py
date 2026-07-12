#!/usr/bin/env python3

import pathlib
import subprocess
import tempfile
import unittest


class KnowledgeSyncTest(unittest.TestCase):
    def test_generates_idempotent_note_and_moc_link(self):
        source_repo = pathlib.Path(__file__).resolve().parents[1]
        script = source_repo / "scripts" / "sync_knowledge.py"
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()
            subprocess.run(["git", "-C", str(repo), "init", "--quiet"], check=True)

            (repo / "README.md").write_text("# fixture\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "README.md"], check=True)
            self.commit(repo, "fixture: initialize repository")
            before = self.rev_parse(repo, "HEAD")

            (repo / "configs").mkdir()
            (repo / "configs" / "config.yaml").write_text(
                'api_listen_host: "127.0.0.1"\napi_port: 8080\napi_auth_enabled: false\n',
                encoding="utf-8",
            )
            (repo / "configs" / "rules.json").write_text('{"rules": []}\n', encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "configs"], check=True)
            self.commit(repo, "fixture: add configuration")
            after = self.rev_parse(repo, "HEAD")

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

    @staticmethod
    def commit(repo, subject):
        subprocess.run(
            [
                "git", "-C", str(repo),
                "-c", "user.name=NetSentry Test",
                "-c", "user.email=netsentry-test@example.invalid",
                "-c", "commit.gpgsign=false",
                "commit", "--quiet", "-m", subject,
            ],
            check=True,
        )

    @staticmethod
    def rev_parse(repo, revision):
        return subprocess.check_output(
            ["git", "-C", str(repo), "rev-parse", revision], text=True
        ).strip()


if __name__ == "__main__":
    unittest.main()
