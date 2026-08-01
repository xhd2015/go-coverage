## Expected

- Run error is nil (admin-deleted checkout must still yield content via bare).
- `resp.Content` equals committed `hello.go` body (`req.Expected[pathHello]`).
- On failure, must **not** treat `not a git repository` as acceptable — that is
  the current broken Cmd.Dir/admin path the Phase 1 contract forbids.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not a git repository") {
			t.Fatalf("GetContent(%q) after admin delete: must resolve via bare/common-dir, got not-a-git-repository: %v", req.File, err)
		}
		t.Fatalf("GetContent(%q) after admin delete: unexpected error: %v", req.File, err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	want := req.Expected[pathHello]
	if resp.Content != want {
		t.Fatalf("GetContent(%q) content = %q, want %q", req.File, resp.Content, want)
	}
}
```
