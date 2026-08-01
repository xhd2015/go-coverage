# Scenario

**Feature**: shared local git fixture for GitSnapshot.GetContent contract leaves

```
# root seeds a temp repo with known tracked files and records commit SHA
test harness -> git init + commit (hello.go, empty.go, special paths)
  -> req.Dir + req.Commit + req.Expected

# leaves select path / concurrency / git-layout; Run calls GitSnapshot.GetContent
caller -> GitSnapshot{Dir,Commit}.GetContent(file) -> body | error

# Phase 1 leaves re-point Dir to bare-linked detached worktree (optional admin delete)
root fixture -> clone --bare + worktree add --detach -> req.Dir=checkout
```

## Preconditions

- `git` is available on PATH.
- L2 only: temp dirs via `t.TempDir()`; no network, no remotes.
- Parallel-safe: one fresh repo per leaf (Setup chain runs per leaf); absolute
  `req.Dir`; no process-wide env/cwd mutation (`t.Setenv` / `os.Chdir` forbidden).
- Fixture files (paths relative to repo root):
  - `hello.go` — non-empty body (`contentHello`)
  - `empty.go` — empty file
  - `sub/a-b.go` — hyphen path, non-empty body
  - `path with spaces/file.go` — spaces in path, non-empty body
- Linked-worktree helpers (Phase 1): after root seed, `repointToLinkedWorktree`
  builds `bare` + detached checkout from the seed; optional delete of only the
  admin dir under `bare/worktrees/<id>` (checkout `.git` pointer file and bare
  objects remain).

## Steps

1. Create a temp directory and `git init` with local user identity.
2. Write the four fixture files and commit them.
3. Record `HEAD` SHA into `req.Commit`, set `req.Dir`, fill `req.Expected`.
4. Leaf Setup selects `File` / concurrent fields and/or re-points to a linked
   worktree (optionally after admin delete).

## Context

- Module: `github.com/xhd2015/go-coverage`
- Package under test: `github.com/xhd2015/go-coverage/git`
- Hot API: `(*GitSnapshot).GetContent(file string) (string, error)`
- Missing-path error substring (current): `not a file, maybe a dir`
- Desired Phase 1 read path: resolve common/bare git dir from `Dir`, then
  `git --git-dir=<bare> cat-file -p <sha>:<path>` (public API stays `Dir`+`Commit`).
- Spec version **0.0.2**

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	contentHello  = "package hello\n\nconst Msg = \"hello from snapshot\"\n"
	contentEmpty  = ""
	contentAB     = "package ab\n"
	contentSpaces = "package spaces\n"

	pathHello  = "hello.go"
	pathEmpty  = "empty.go"
	pathAB     = "sub/a-b.go"
	pathSpaces = "path with spaces/file.go"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "get-content-doctest@example.com")
	runGit(t, dir, "config", "user.name", "GetContent Doctest")
	runGit(t, dir, "branch", "-M", "master")

	writeRepoFile(t, dir, pathHello, contentHello)
	writeRepoFile(t, dir, pathEmpty, contentEmpty)
	writeRepoFile(t, dir, pathAB, contentAB)
	writeRepoFile(t, dir, pathSpaces, contentSpaces)

	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "seed get-content fixture")

	sha := strings.TrimSpace(runGitOutput(t, dir, "rev-parse", "HEAD"))
	if sha == "" {
		t.Fatal("empty commit SHA")
	}

	req.Dir = dir
	req.Commit = sha
	req.Expected = map[string]string{
		pathHello:  contentHello,
		pathEmpty:  contentEmpty,
		pathAB:     contentAB,
		pathSpaces: contentSpaces,
	}
	return nil
}

// repointToLinkedWorktree turns the root seed repo into a bare + detached linked
// worktree fixture. req.Dir becomes the checkout path; req.Commit is unchanged.
// When deleteAdmin is true, only the linked admin gitdir (bare/worktrees/<id>)
// is removed — checkout .git pointer file and bare objects stay.
func repointToLinkedWorktree(t *testing.T, req *Request, deleteAdmin bool) {
	t.Helper()
	if req.Dir == "" || req.Commit == "" {
		t.Fatal("repointToLinkedWorktree requires root fixture Dir+Commit")
	}
	origin := req.Dir
	commit := req.Commit

	base := t.TempDir()
	bare := filepath.Join(base, "repo.git")
	checkout := filepath.Join(base, "wt-checkout")

	runGitNoC(t, "clone", "--bare", origin, bare)
	runGit(t, bare, "worktree", "add", "--detach", checkout, commit)

	admin := linkedAdminGitDir(t, checkout)
	if st, err := os.Stat(admin); err != nil || !st.IsDir() {
		t.Fatalf("admin gitdir missing or not dir: %s: %v", admin, err)
	}
	if !strings.Contains(admin, string(filepath.Separator)+"worktrees"+string(filepath.Separator)) {
		t.Fatalf("expected linked admin under worktrees/, got %q (bare=%q)", admin, bare)
	}

	if deleteAdmin {
		if err := os.RemoveAll(admin); err != nil {
			t.Fatalf("delete admin gitdir %s: %v", admin, err)
		}
		if _, err := os.Stat(admin); !os.IsNotExist(err) {
			t.Fatalf("admin gitdir still present after delete: %s (stat err=%v)", admin, err)
		}
		// Bare objects must still exist for the desired contract.
		if _, err := os.Stat(bare); err != nil {
			t.Fatalf("bare repo missing after admin delete: %v", err)
		}
	}

	req.Dir = checkout
}

// linkedAdminGitDir reads checkout/.git (gitdir pointer) and returns the absolute
// path of the linked admin directory under the bare repo.
func linkedAdminGitDir(t *testing.T, checkout string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(checkout, ".git"))
	if err != nil {
		t.Fatalf("read checkout .git: %v", err)
	}
	line := strings.TrimSpace(string(data))
	const prefix = "gitdir:"
	if !strings.HasPrefix(strings.ToLower(line), prefix) {
		t.Fatalf("checkout .git is not a gitdir pointer: %q", line)
	}
	// Preserve path case: strip by index of first colon form "gitdir:"
	idx := strings.Index(strings.ToLower(line), prefix)
	rest := strings.TrimSpace(line[idx+len(prefix):])
	if rest == "" {
		t.Fatalf("empty gitdir pointer in %s/.git", checkout)
	}
	if !filepath.IsAbs(rest) {
		rest = filepath.Clean(filepath.Join(checkout, rest))
	} else {
		rest = filepath.Clean(rest)
	}
	return rest
}

func writeRepoFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func runGitNoC(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
```
