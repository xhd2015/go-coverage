# Scenario

**Feature**: GetContent of a tracked empty file returns empty string with nil error

```
# empty blob still succeeds
GitSnapshot -> GetContent(empty.go) -> ("", nil)
```

## Preconditions

- Root fixture committed empty `empty.go`.

## Steps

1. Set `req.File` to `empty.go`.
2. Run calls `GetContent(req.File)`.

## Context

- Empty body must not be confused with missing-file error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	req.File = pathEmpty
	return nil
}
```
