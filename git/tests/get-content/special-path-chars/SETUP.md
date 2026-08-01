# Scenario

**Feature**: GetContent works for paths that must not depend on shell word-splitting

```
# spaces in path — argv-safe cat-file contract
GitSnapshot -> GetContent("path with spaces/file.go") -> contentSpaces
```

## Preconditions

- Root fixture committed `path with spaces/file.go` with `contentSpaces`.
- Also seeds `sub/a-b.go` (used by concurrent leaf); this leaf focuses on spaces.

## Steps

1. Set `req.File` to `path with spaces/file.go`.
2. Run calls `GetContent(req.File)`.

## Context

- Locks semantics when implementer switches from `bash -c` string to direct
  `git cat-file -p ref:file` argv (no process inspection in assert).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	req.File = pathSpaces
	return nil
}
```
