#!/usr/bin/env python3
"""Independent fixture tests for the push-success knowledge sync helper.

Each test creates its own Git repository and Vault temp directory so it never
depends on the calling NetSentry repo or the user's real Vault.
"""

import os
import pathlib
import shutil
import subprocess
import tempfile
import unittest


POST_PUSH = pathlib.Path(__file__).resolve().parents[1] / ".git" / "hooks" / "post-push"


def _git(repo, *args, env=None):
    return subprocess.check_output(
        ["git", "-C", str(repo), *args],
        text=True, stderr=subprocess.DEVNULL,
        env=env,
    ).strip()


def _commit(repo, path, content, subject):
    (repo / path).parent.mkdir(parents=True, exist_ok=True)
    (repo / path).write_text(content, encoding="utf-8")
    _git(repo, "add", path)
    _git(
        repo,
        "-c", "user.name=NetSentry Test",
        "-c", "user.email=netsentry-test@example.invalid",
        "-c", "commit.gpgsign=false",
        "commit", "--quiet", "-m", subject,
    )


def _rev(repo, ref):
    return _git(repo, "rev-parse", ref)


def _is_ancestor(repo, maybe_ancestor, child):
    rc = subprocess.call(
        ["git", "-C", str(repo), "merge-base", "--is-ancestor", maybe_ancestor, child],
        stderr=subprocess.DEVNULL,
    )
    return rc == 0


def _sync(helper_path, vault, repo, sync_range, expect_ok=True):
    env = os.environ.copy()
    env["NETSENTRY_VAULT"] = str(vault)
    vault.mkdir(parents=True, exist_ok=True)
    env["NETSENTRY_SYNC_RANGE"] = sync_range
    cp = subprocess.run(
        [str(helper_path), "sync"],
        capture_output=True, text=True, cwd=str(repo), env=env,
    )
    if expect_ok:
        assert cp.returncode == 0, f"sync failed: {cp.stderr}"
    return cp


class ExactRangeTest(unittest.TestCase):
    def test_only_specified_commits_appear_in_iteration_note(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "README.md", "# Test Repo\n", "docs: initial commit")
            start = _rev(repo, "HEAD")

            _commit(repo, "src/main.go", "package main\n", "feat: add main package")
            mid = _rev(repo, "HEAD")

            _commit(repo, "src/util.go", "package util\n", "feat: add util package")
            end = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{start}..{end}")

            notes = sorted((vault / "04-开发迭代记录").glob("*-自动知识同步.md"))
            self.assertGreaterEqual(len(notes), 1)

            text = notes[0].read_text(encoding="utf-8")
            self.assertIn(mid[:10], text)
            self.assertIn(end[:10], text)
            # First commit's subject should not be in the commit table body
            self.assertNotIn("docs: initial commit", text)


class IdempotencyTest(unittest.TestCase):
    def test_same_range_twice_produces_no_duplicate_index_rows(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "README.md", "# Test\n", "docs: first commit")
            before = _rev(repo, "HEAD")
            _commit(repo, "CHANGES.md", "# Changelog\n", "docs: changelog")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")
            first_lines = _index_lines(vault)

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")
            second_lines = _index_lines(vault)

            self.assertEqual(first_lines, second_lines, "index must be idempotent")


def _index_lines(vault):
    index = vault / "04-开发迭代记录" / "全量提交索引.md"
    if not index.exists():
        return []
    return [line for line in index.read_text(encoding="utf-8").splitlines()
            if line.startswith("|") and "---" not in line]


class NonAncestorRangeTest(unittest.TestCase):
    def test_non_ancestor_range_does_not_crash(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "a.txt", "a\n", "feat: a")
            # Create an orphan branch with unrelated history
            _git(repo, "checkout", "--quiet", "--orphan", "orphan")
            _git(repo, "rm", "-rf", "--quiet", ".")
            _commit(repo, "b.txt", "b\n", "feat: b (orphan)")
            orphan_sha = _rev(repo, "HEAD")
            _git(repo, "checkout", "--quiet", "main" if "main" in _git(repo, "branch", "-a") else "master")
            main_sha = _rev(repo, "HEAD")

            self.assertFalse(_is_ancestor(repo, main_sha, orphan_sha))
            cp = _sync(POST_PUSH, vault, repo, f"{main_sha}..{orphan_sha}", expect_ok=False)
            self.assertEqual(cp.returncode, 0, "non-ancestor range completes without crash")


class EmptyRangeTest(unittest.TestCase):
    def test_empty_range_produces_no_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "x.txt", "x\n", "single commit")
            sha = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{sha}..{sha}")

            notes = list((vault / "04-开发迭代记录").glob("*-自动知识同步.md"))
            self.assertEqual(len(notes), 0, "empty range must not create iteration notes")


class MOCReachabilityTest(unittest.TestCase):
    def test_moc_has_valid_wikilinks_after_sync(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "README.md", "# Fixture\n", "docs: readme")
            before = _rev(repo, "HEAD")
            _commit(repo, "Makefile", "all:\n\techo ok\n", "build: add Makefile")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")

            moc = vault / "00-MOC" / "NetSentry知识总览.md"
            self.assertTrue(moc.exists(), "MOC must exist after sync")
            content = moc.read_text(encoding="utf-8")
            self.assertIn("[[", content, "MOC should contain wikilinks")


class NegativeSecretsTest(unittest.TestCase):

    SECRET_LIKE = ["password", "sudo", "token", "ghp_", "sk-", "PRIVATE KEY", "secret", "credential"]

    def test_vault_output_contains_no_credential_like_strings(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "config.env", 'DB_PASSWORD=secret123\nAPI_TOKEN=ghp_fake\n', "config: add env")
            before = _rev(repo, "HEAD")
            _commit(repo, "src/app.py", "print('hello')\n", "feat: app")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")

            for md_file in vault.rglob("*.md"):
                text = md_file.read_text(encoding="utf-8").lower()
                for keyword in self.SECRET_LIKE:
                    self.assertNotIn(keyword.lower(), text,
                                     f"Vault file {md_file} contains sensitive keyword: {keyword}")

    def test_raw_repository_absolute_path_not_leaked(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "a.txt", "content\n", "feat: initial")
            before = _rev(repo, "HEAD")
            _commit(repo, "b.txt", "more\n", "feat: second")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")

            repo_str = str(repo.resolve())
            for md_file in vault.rglob("*.md"):
                text = md_file.read_text(encoding="utf-8")
                self.assertNotIn(repo_str, text,
                                 f"Vault file {md_file} leaked absolute repo path")


class FailureRecoveryTest(unittest.TestCase):
    def test_sync_after_interrupted_sync_produces_valid_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            _git(repo, "init", "--quiet")
            _commit(repo, "x.txt", "x\n", "feat: first")
            before = _rev(repo, "HEAD")
            _commit(repo, "y.txt", "y\n", "feat: second")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")
            index = vault / "04-开发迭代记录" / "全量提交索引.md"
            moc = vault / "00-MOC" / "NetSentry知识总览.md"

            self.assertTrue(index.exists())
            self.assertTrue(moc.exists())

            # Simulate partial corruption by truncating index
            index.write_text("", encoding="utf-8")
            self.assertEqual(index.stat().st_size, 0)

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")
            self.assertGreater(index.stat().st_size, 0, "re-sync must rebuild empty index")


if __name__ == "__main__":
    unittest.main()


class PreExistingVaultTest(unittest.TestCase):
    """Sync into a Vault that already has notes from previous sessions."""

    def test_sync_into_populated_vault_preserves_existing_notes(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()

            # Pre-populate Vault with an existing note that mimics a prior sync
            existing_iter = (
                vault / "04-开发迭代记录" / "2026-01-01-deadbeef01-自动知识同步.md"
            )
            existing_iter.parent.mkdir(parents=True, exist_ok=True)
            existing_iter.write_text(
                "---\n分类: 开发迭代记录\n---\n\n# 历史同步记录\n\n旧提交\n", encoding="utf-8"
            )
            existing_moc = vault / "00-MOC" / "NetSentry知识总览.md"
            existing_moc.parent.mkdir(parents=True, exist_ok=True)
            existing_moc.write_text(
                "<!-- netsentry-auto-sync:start -->\n- old entry\n<!-- netsentry-auto-sync:end -->\n\n# Rest of MOC\n",
                encoding="utf-8",
            )

            # Create repo and sync
            _git(repo, "init", "--quiet")
            _commit(repo, "README.md", "# Project\n", "docs: initial")
            before = _rev(repo, "HEAD")
            _commit(repo, "CHANGELOG.md", "# Log\n", "docs: changelog")
            after = _rev(repo, "HEAD")

            _sync(POST_PUSH, vault, repo, f"{before}..{after}")

            # Old note must survive
            self.assertTrue(existing_iter.exists(), "pre-existing iteration note preserved")
            # MOC must still have rest of content outside sync block
            moc_text = existing_moc.read_text(encoding="utf-8")
            self.assertIn("Rest of MOC", moc_text, "non-sync MOC sections preserved")
            # Sync block was updated
            self.assertIn("auto-sync:start", moc_text)

    def test_empty_vault_full_sync_rebuilds_everything(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            repo = root / "repo"
            vault = root / "vault"
            repo.mkdir()
            vault.mkdir(parents=True, exist_ok=True)

            _git(repo, "init", "--quiet")
            _commit(repo, "a.txt", "a\n", "one")
            sha = _rev(repo, "HEAD")

            # Full sync from empty vault
            _sync(POST_PUSH, vault, repo, f"{sha}..{sha}", expect_ok=False)
            # Empty range — no notes expected, but no crash
            index = vault / "04-开发迭代记录" / "全量提交索引.md"
            notes = list((vault / "04-开发迭代记录").glob("*-自动知识同步.md"))
            self.assertEqual(len(notes), 0, "empty range produces no notes")
