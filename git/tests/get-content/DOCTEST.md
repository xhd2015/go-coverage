# GitSnapshot.GetContent — blob content contract

Contract tests for `(*GitSnapshot).GetContent` reading committed file bodies from a
local git fixture (no network). Semantics are locked for the hardened
**direct `git cat-file`** path (implementer removes `bash -c`); content leaves may
already be GREEN on the current bash-based implementation.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Participants**

- **caller** — builds `GitSnapshot{Dir, Commit}` after a local `git init` + commit
  fixture and asks for blob text of a tracked path.
- **GitSnapshot** — snapshot of one commit in a work-tree directory; caches
  `ls-files` membership for path checks.
- **GetContent(file)** — returns the committed blob body for a tracked file path
  (normalized by trimming a leading `./`), or a clear error when the path is not
  a file in the snapshot list.
- **ListFiles** — lists tracked paths at the snapshot commit; used by concurrent
  leaves to warm the file map before fan-out.
- **local git fixture** — temp repo with known files and a recorded commit SHA
  (L2 only; no remotes).

**Behaviors**

- Tracked non-empty file → `(body, nil)` equal to the committed bytes as string.
- Tracked empty file → `("", nil)`.
- Path not in the snapshot file list → error mentioning `not a file, maybe a dir`
  (current list-based gate); no silent empty string.
- Paths with shell-sensitive characters (spaces, `sub/a-b.go`) → same content
  contract without requiring a shell string for the file path.
- After `ListFiles`, many concurrent `GetContent` calls on the same snapshot
  succeed with correct bodies (regression guard for ECHILD / process wait issues
  under load; does not require process inspection).

## Decision Tree

Most significant factor is **path / content outcome** of a single `GetContent`
call; concurrent multi-call is a separate access-pattern leaf on the same fixture.

```
get-content/   (GitSnapshot.GetContent @ committed SHA)
├── existing-file-content/       # hello.go → committed body
├── empty-file-ok/               # empty.go → "" , err=nil
├── missing-file-error/          # path not in ls-files → error
├── special-path-chars/          # path with spaces / a-b → content
└── concurrent-many-getcontent/  # N goroutines GetContent, all correct
```

### Parameter significance (high → low)

1. **Path outcome** — exists (non-empty / empty / special chars) vs missing.
2. **Access pattern** — single call vs concurrent many calls on one snapshot.

## Test Case Index

| # | Path | Preconditions | Expected |
|---|------|---------------|----------|
| 1 | `existing-file-content/` | Fixture commit includes `hello.go` with known body | `Content` equals body; err nil |
| 2 | `empty-file-ok/` | Fixture includes empty `empty.go` | `Content == ""`; err nil |
| 3 | `missing-file-error/` | Path not tracked in commit | err non-nil; message contains `not a file, maybe a dir` |
| 4 | `special-path-chars/` | `path with spaces/file.go` tracked | Content equals committed body; err nil |
| 5 | `concurrent-many-getcontent/` | ListFiles warmed; 30 concurrent GetContent over 3 files | All hits err-free; content matches expected map |

## How to Run

```sh
# from go-coverage module root
doctest vet ./git/tests/get-content
doctest test -v ./git/tests/get-content
```

```go
import (
	"sync"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/go-coverage/git"
)

// ConcurrentHit is one GetContent outcome under concurrent fan-out.
type ConcurrentHit struct {
	File    string
	Content string
	ErrMsg  string
}

type Request struct {
	Dir    string
	Commit string

	// Single-file mode when Concurrent == 0.
	File string

	// Concurrent mode: Concurrent > 0 fans out GetContent over Files (round-robin).
	Concurrent int
	Files      []string

	// Expected bodies keyed by path (seeded by root Setup for asserts).
	Expected map[string]string
}

type Response struct {
	Content string
	Hits    []ConcurrentHit
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if req.Dir == "" {
		t.Fatal("req.Dir must be set by Setup")
	}
	snap := git.NewSnapshot(req.Dir, req.Commit)

	if req.Concurrent > 0 {
		if len(req.Files) == 0 {
			t.Fatal("concurrent mode requires req.Files")
		}
		if _, err := snap.ListFiles(); err != nil {
			return nil, err
		}
		n := req.Concurrent
		type raw struct {
			idx     int
			file    string
			content string
			err     error
		}
		ch := make(chan raw, n)
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			i := i
			f := req.Files[i%len(req.Files)]
			go func() {
				defer wg.Done()
				c, err := snap.GetContent(f)
				ch <- raw{idx: i, file: f, content: c, err: err}
			}()
		}
		go func() {
			wg.Wait()
			close(ch)
		}()

		resp := &Response{Hits: make([]ConcurrentHit, n)}
		var first error
		for r := range ch {
			msg := ""
			if r.err != nil {
				msg = r.err.Error()
				if first == nil {
					first = r.err
				}
			}
			resp.Hits[r.idx] = ConcurrentHit{
				File:    r.file,
				Content: r.content,
				ErrMsg:  msg,
			}
		}
		return resp, first
	}

	if req.File == "" {
		t.Fatal("req.File must be set by Setup for single-file mode")
	}
	content, err := snap.GetContent(req.File)
	return &Response{Content: content}, err
}
```
