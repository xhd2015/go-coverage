## Expected

- Run error is nil.
- `resp.Content` equals committed body for `path with spaces/file.go`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("GetContent(%q): unexpected error: %v", req.File, err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	want := req.Expected[pathSpaces]
	if resp.Content != want {
		t.Fatalf("GetContent(%q) content = %q, want %q", req.File, resp.Content, want)
	}
}
```
