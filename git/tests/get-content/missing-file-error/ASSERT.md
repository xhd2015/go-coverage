## Expected

- Run returns a non-nil error.
- Error string contains `not a file, maybe a dir` (and preferably the requested path).
- Must not succeed with empty content.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err == nil {
		got := ""
		if resp != nil {
			got = resp.Content
		}
		t.Fatalf("GetContent(%q): want error, got nil (content=%q)", req.File, got)
	}
	msg := err.Error()
	if !strings.Contains(msg, "not a file, maybe a dir") {
		t.Fatalf("GetContent(%q): error %q, want substring %q", req.File, msg, "not a file, maybe a dir")
	}
	if !strings.Contains(msg, req.File) {
		t.Fatalf("GetContent(%q): error %q should mention the path", req.File, msg)
	}
}
```
