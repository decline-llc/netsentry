"""Deterministic pre-built Git fixture repository for post-push sync testing.

Provides ``build_fixture_repo(target_dir)`` that creates a Git repository with
known commit history, diverse file types, and fixed timestamps so commit SHAs
are reproducible across machines.  This avoids temp-directory-per-test overhead
and gives every test the same well-known history to verify against.
"""

import pathlib
import subprocess
import textwrap

FIXTURE_AUTHOR_NAME = "NetSentry Fixture"
FIXTURE_AUTHOR_EMAIL = "netsentry-fixture@example.invalid"
# Unix epoch seconds — same author & committer timestamp for every commit so
# SHAs are deterministic across machines and Git versions.
FIXTURE_EPOCH = 1700000000
FIXTURE_DATE = "@{:d}".format(FIXTURE_EPOCH)


def _git(repo, *args, env=None):
    return subprocess.check_output(
        ["git", "-C", str(repo), *args],
        text=True, stderr=subprocess.DEVNULL,
        env=env,
    ).strip()


def _env():
    import os
    e = os.environ.copy()
    e["GIT_AUTHOR_DATE"] = FIXTURE_DATE
    e["GIT_COMMITTER_DATE"] = FIXTURE_DATE
    return e


def _commit(repo, path, content, subject):
    (repo / path).parent.mkdir(parents=True, exist_ok=True)
    (repo / path).write_text(content, encoding="utf-8")
    _git(repo, "add", str(path), env=_env())
    _git(
        repo,
        "-c", "user.name={}".format(FIXTURE_AUTHOR_NAME),
        "-c", "user.email={}".format(FIXTURE_AUTHOR_EMAIL),
        "-c", "commit.gpgsign=false",
        "commit", "--quiet", "-m", subject,
        env=_env(),
    )


def build_fixture_repo(target_dir):
    """Create a deterministic Git repo at *target_dir* with known history.

    Returns a dict with keys:

    * ``repo`` — ``pathlib.Path`` to the created repo root
    * ``shas`` — list of full commit SHAs in reverse chronological order (newest first)
    * ``short_shas`` — same but first-10-char abbreviations
    * ``root_sha`` — the root commit SHA
    * ``head_sha`` — the current HEAD SHA
    """
    repo = pathlib.Path(target_dir)
    repo.mkdir(parents=True, exist_ok=True)

    _git(repo, "init", "--quiet", "-b", "main")

    # ---- 8 deterministic commits covering diverse file types ----
    files = textwrap.dedent("""\
        # Fixture Project

        A deterministic test fixture for the NetSentry post-push knowledge sync helper.

        ## Build

        ```bash
        make
        ```

        ## License

        MIT
        """)
    _commit(repo, "README.md", files, "docs: initial project README")

    makefile = textwrap.dedent("""\
        .PHONY: build test clean

        build:
        \\tgo build ./...

        test:
        \\tgo test ./...

        clean:
        \\trm -f bin/*
        """)
    _commit(repo, "Makefile", makefile, "build: add Makefile with Go targets")

    main_go = textwrap.dedent("""\
        package main

        import "fmt"

        func main() {
        \\tfmt.Println("NetSentry fixture")
        }
        """)
    _commit(repo, "cmd/fixture/main.go", main_go, "feat: add Go main entrypoint")

    config_yaml = textwrap.dedent("""\
        api_listen_host: "127.0.0.1"
        api_port: 8080
        log_level: info
        """)
    _commit(repo, "configs/config.yaml", config_yaml, "config: add default config")

    rules_json = textwrap.dedent("""\
        {
          "rules": [
            {"id": "R001", "type": "keyword", "keywords": ["test"]}
          ]
        }
        """)
    _commit(repo, "configs/rules.json", rules_json, "feat: add seed detection rule")

    # A commit that modifies an existing file (tests the diff-tree path reporting)
    updated_config = textwrap.dedent("""\
        api_listen_host: "127.0.0.1"
        api_port: 8080
        api_auth_enabled: true
        log_level: debug
        """)
    _commit(repo, "configs/config.yaml", updated_config,
            "config: enable API auth and debug logging")

    changelog = textwrap.dedent("""\
        # Changelog

        ## 0.1.0

        - Initial fixture release
        """)
    _commit(repo, "CHANGELOG.md", changelog, "docs: add changelog")

    test_go = textwrap.dedent("""\
        package main

        import "testing"

        func TestFixture(t *testing.T) {
        \\tt.Log("fixture test ok")
        }
        """)
    _commit(repo, "cmd/fixture/main_test.go", test_go, "test: add unit test for fixture")

    # Collect all SHAs (newest first)
    shas = _git(repo, "log", "--format=%H").splitlines()
    result = {
        "repo": repo,
        "shas": shas,
        "short_shas": [s[:10] for s in shas],
        "root_sha": shas[-1] if shas else None,
        "head_sha": shas[0] if shas else None,
    }
    return result
