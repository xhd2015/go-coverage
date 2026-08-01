# Scenario

**Feature**: GetContent returns the committed body of a tracked non-empty file

```
# snapshot at fixture SHA -> GetContent("hello.go") -> exact body
GitSnapshot{Dir,Commit} -> GetContent(hello.go) -> contentHello
```

## Preconditions

- Root fixture committed `hello.go` with `contentHello`.

## Steps

1. Set `req.File` to `hello.go`.
2. Run calls `GetContent(req.File)`.

## Context

- Single-file mode (`Concurrent == 0`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	req.File = pathHello
	return nil
}
```
