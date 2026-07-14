#!/usr/bin/env python3
"""Direct unit tests for the versioned post-push knowledge synchronizer."""

import pathlib
import subprocess
import sys
import tempfile
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from post_push_sync import sync


def git(repo, *args):
    return subprocess.check_output(["git", "-C", str(repo), *args], text=True).strip()


def commit(repo, path, content, subject):
    file_path = repo / path
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_text(content, encoding="utf-8")
    git(repo, "add", path)
    git(repo, "-c", "user.name=Test", "-c", "user.email=test@example.invalid",
        "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", subject)


class PostPushSyncTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = pathlib.Path(self.temp.name)
        self.repo = self.root / "repo"
        self.vault = self.root / "vault"
        self.repo.mkdir()
        git(self.repo, "init", "--quiet")
        commit(self.repo, "README.md", "# fixture\n", "docs: initial")
        self.before = git(self.repo, "rev-parse", "HEAD")
        commit(self.repo, "src/app.py", "print('ok')\n", "feat: app")
        self.after = git(self.repo, "rev-parse", "HEAD")

    def tearDown(self):
        self.temp.cleanup()

    def run_sync(self):
        return sync(self.repo, self.vault, f"{self.before}..{self.after}")

    def note(self):
        notes = list((self.vault / "04-开发迭代记录").glob("*-CI知识同步.md"))
        self.assertEqual(len(notes), 1)
        return notes[0]

    def test_direct_sync_returns_range(self):
        self.assertIn(self.after[:10], self.run_sync()["range"])

    def test_direct_sync_creates_note(self):
        self.assertTrue(self.note_after_sync().exists())

    def note_after_sync(self):
        self.run_sync()
        return self.note()

    def test_note_records_changed_file(self):
        self.assertIn("src/app.py", self.note_after_sync().read_text(encoding="utf-8"))

    def test_note_records_commit_subject(self):
        self.assertIn("feat: app", self.note_after_sync().read_text(encoding="utf-8"))

    def test_moc_is_created(self):
        self.run_sync()
        self.assertTrue((self.vault / "00-MOC" / "NetSentry知识总览.md").exists())

    def test_moc_links_note(self):
        self.run_sync()
        self.assertIn("[[04-开发迭代记录/", (self.vault / "00-MOC" / "NetSentry知识总览.md").read_text(encoding="utf-8"))

    def test_full_index_records_synced_commit(self):
        self.run_sync()
        index = (self.vault / "04-开发迭代记录" / "全量提交索引.md").read_text(encoding="utf-8")
        self.assertIn(self.after[:10], index)

    def test_repeat_sync_is_idempotent(self):
        first = self.note_after_sync().read_text(encoding="utf-8")
        self.run_sync()
        self.assertEqual(first, self.note().read_text(encoding="utf-8"))

    def test_invalid_range_is_rejected(self):
        with self.assertRaises(ValueError):
            sync(self.repo, self.vault, "not-a-range")

    def test_missing_before_revision_is_rejected(self):
        with self.assertRaises(subprocess.CalledProcessError):
            sync(self.repo, self.vault, f"missing..{self.after}")

    def test_missing_after_revision_is_rejected(self):
        with self.assertRaises(subprocess.CalledProcessError):
            sync(self.repo, self.vault, f"{self.before}..missing")

    def test_repository_absolute_path_is_not_written(self):
        text = self.note_after_sync().read_text(encoding="utf-8")
        self.assertNotIn(str(self.repo.resolve()), text)

    def test_second_range_creates_second_note(self):
        self.run_sync()
        before = self.after
        commit(self.repo, "src/next.py", "pass\n", "feat: next")
        after = git(self.repo, "rev-parse", "HEAD")
        sync(self.repo, self.vault, f"{before}..{after}")
        self.assertEqual(len(list((self.vault / "04-开发迭代记录").glob("*-CI知识同步.md"))), 2)

    def test_note_contains_required_frontmatter(self):
        text = self.note_after_sync().read_text(encoding="utf-8")
        self.assertIn("分类:", text)

    def test_note_contains_review_section(self):
        self.assertIn("## 人工复核问题", self.note_after_sync().read_text(encoding="utf-8"))

    def test_sync_is_callable_without_hook_files(self):
        self.assertEqual(self.run_sync()["status"], "ok")


if __name__ == "__main__":
    unittest.main()
