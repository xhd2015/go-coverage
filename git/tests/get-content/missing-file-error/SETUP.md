# Scenario

**Feature**: GetContent of a path not in the snapshot file list fails clearly

```
# path never committed / not in ls-files map
GitSnapshot -> GetContent(does-not-exist.go) -> error "not a file, maybe a dir"
```

## Preconditions

- Root fixture does not track `does-not-exist.go`.

## Steps

1. Set `req.File` to `does-not-exist.go`.
2. Run calls `GetContent(req.File)`.

## Context

- Current gate is list-based (`fileMap`); message matches production copy.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	req.File = "does-not-exist.go"
	return nil
}
```
