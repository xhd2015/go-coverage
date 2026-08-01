# Scenario

**Feature**: GetContent works when Dir is a healthy linked detached worktree

```
# bare + worktree add --detach; admin gitdir under bare/worktrees still present
root seed -> clone --bare + worktree add --detach -> GitSnapshot{Dir:checkout,Commit}
  -> GetContent(hello.go) -> contentHello
```

## Preconditions

- Root fixture committed `hello.go` with `contentHello`.
- Leaf re-points `req.Dir` to a linked detached worktree of that commit; admin
  gitdir intact.

## Steps

1. Call `repointToLinkedWorktree(t, req, false)` (admin kept).
2. Set `req.File` to `hello.go`.
3. Run calls `GetContent(req.File)` with `Dir` = linked checkout.

## Context

- Preferred healthy-layout leaf for Phase 1 bare-dir design.
- Expected **GREEN** on current master (Cmd.Dir=checkout still resolves via admin).
- Single-file mode (`Concurrent == 0`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	repointToLinkedWorktree(t, req, false)
	req.File = pathHello
	return nil
}
```
