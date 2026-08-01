# Scenario

**Feature**: shared local git fixture for GitSnapshot.GetContent contract leaves

```
# root seeds a temp repo with known tracked files and records commit SHA
test harness -> git init + commit (hello.go, empty.go, special paths)
  -> req.Dir + req.Commit + req.Expected

# leaves select path / concurrency; Run calls GitSnapshot.GetContent
caller -> GitSnapshot{Dir,Commit}.GetContent(file) -> body | error
```

## Preconditions

- `git` is available on PATH.
- L2 only: temp dirs via `t.TempDir()`; no network, no remotes.
- Parallel-safe: one fresh repo per leaf (Setup chain runs per leaf); absolute
  `req.Dir`; no process-wide env/cwd mutation.
- Fixture files (paths relative to repo root):
  - `hello.go` — non-empty body (`contentHello`)
  - `empty.go` — empty file
  - `sub/a-b.go` — hyphen path, non-empty body
  - `path with spaces/file.go` — spaces in path, non-empty body

## Steps

1. Create a temp directory and `git init` with local user identity.
2. Write the four fixture files and commit them.
3. Record `HEAD` SHA into `req.Commit`, set `req.Dir`, fill `req.Expected`.
4. Leaf Setup selects `File` or concurrent fields on the same fixture.

## Context

- Module: `github.com/xhd2015/go-coverage`
- Package under test: `github.com/xhd2015/go-coverage/git`
- Hot API: `(*GitSnapshot).GetContent(file string) (string, error)`
- Missing-path error substring (current): `not a file, maybe a dir`
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
