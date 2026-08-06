# CI workflow note

## status

**added** (and pushed)

## branch / remote / push

| Field | Value |
|-------|--------|
| Branch | `master-2026-08-06-use-go-best-practice-to-review-current-project` |
| Remote | `origin` → `ssh://git@github.com/xhd2015/go-coverage.git` |
| Upstream | `origin/master-2026-08-06-use-go-best-practice-to-review-current-project` (created/updated with `-u`) |
| Push result | **success** (`bfc4fb4..80f7577`) |
| Commit SHA | `80f7577e58314ec8b65833b8c6685e84da1423c8` (short: `80f7577`) |

## paths changed (in CI commit)

| Path | Action |
|------|--------|
| `.github/workflows/test.yml` | added |
| `script/ci/coverage-package-table.py` | added (go-coverage module filter; skips `cmd/`) |

Commit message: `ci: add GitHub Actions test workflow aligned with doctest pattern`

## how to view Actions for this push

- Repo: https://github.com/xhd2015/go-coverage  
- Actions (all): https://github.com/xhd2015/go-coverage/actions  
- Branch filter: https://github.com/xhd2015/go-coverage/actions?query=branch%3Amaster-2026-08-06-use-go-best-practice-to-review-current-project  
- Workflow file on branch: https://github.com/xhd2015/go-coverage/blob/master-2026-08-06-use-go-best-practice-to-review-current-project/.github/workflows/test.yml  
- Commit: https://github.com/xhd2015/go-coverage/commit/80f7577e58314ec8b65833b8c6685e84da1423c8  

## how this differs from doctest’s workflow

Reference: doctest worktree `.github/workflows/test.yml`.

| Aspect | doctest | go-coverage |
|--------|---------|-------------|
| Triggers | `push` + `pull_request` | same |
| Runner | `ubuntu-latest` | same |
| `COVERPKG` | `github.com/xhd2015/doctest/...` | `github.com/xhd2015/go-coverage/...` |
| setup-go | `go-version-file: go.mod` | **`go-version: stable`** — module still declares `go 1.16`; too old for `doctest@latest` |
| Install doctest | `go install ./cmd/doctest` | `go install github.com/xhd2015/doctest/cmd/doctest@latest` |
| Doctest stages | discovery + e2e | same (`git/tests/get-content/` trees; e2e unlabeled today) |
| xgo merge + summary + artifacts | yes + package table | same pattern + `script/ci/coverage-package-table.py` adapted for this module |

## notes

- Pre-commit hook `git-hook-github-workflow-test` warned that the workflow differs from a local “recommended” template; intentionally kept **doctest-aligned** pattern as requested.
- Existing package/test failures in the tree will fail the new CI job until fixed (out of scope for this ship).
- No PR opened.
