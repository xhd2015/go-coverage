# Scenario

**Feature**: GetContent still returns blob when linked admin gitdir was deleted

```
# sever only bare/worktrees/<id>; keep checkout .git pointer + bare objects
root seed -> bare + detached worktree -> delete admin gitdir
  -> GitSnapshot{Dir:checkout,Commit}.GetContent(hello.go)
  -> contentHello (via bare/common-dir; not via linked admin)
```

## Preconditions

- Root fixture committed `hello.go` with `contentHello`.
- Linked detached worktree of that commit, then **admin gitdir only** removed
  (`gitdir:` target under bare/worktrees); checkout tree and bare objects remain.
- Classic TDD RED driver until implementer resolves git common dir for cat-file /
  list membership instead of relying on linked admin via `Cmd.Dir`.

## Steps

1. Call `repointToLinkedWorktree(t, req, true)` (delete admin after worktree add).
2. Set `req.File` to `hello.go`.
3. Run calls `GetContent(req.File)` with severed-admin checkout as `Dir`.

## Context

- Desired product path: resolve bare/common dir once from `Dir`, then
  `git --git-dir=<bare> cat-file -p <sha>:<path>`.
- Current master uses `Cmd.Dir = checkout` and fails with
  `fatal: not a git repository: …/worktrees/<id>` — leaf must stay RED until fix.
- Single-file mode (`Concurrent == 0`). Public API still `Dir` + `Commit` only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	repointToLinkedWorktree(t, req, true)
	req.File = pathHello
	return nil
}
```
