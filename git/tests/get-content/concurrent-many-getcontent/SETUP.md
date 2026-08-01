# Scenario

**Feature**: many concurrent GetContent calls on one snapshot return correct bodies

```
# warm ListFiles, then fan-out GetContent on shared GitSnapshot
ListFiles -> N=30 goroutines GetContent({hello.go,empty.go,sub/a-b.go})
  -> all err=nil, content matches Expected
```

## Preconditions

- Root fixture has at least `hello.go`, `empty.go`, `sub/a-b.go`.
- Same `GitSnapshot` instance is shared across goroutines (Run constructs one).

## Steps

1. Set `req.Concurrent = 30`.
2. Set `req.Files` to the three tracked paths above.
3. Run warms `ListFiles` then fans out `GetContent`.

## Context

- Stresses process wait / ECHILD class failures under concurrent bash `git`
  without inspecting process tables (content+err only).
- May be GREEN on current code or flaky under load; not the sole RED gate.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	req.Concurrent = 30
	req.Files = []string{pathHello, pathEmpty, pathAB}
	return nil
}
```
