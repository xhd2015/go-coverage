## Expected

- Run error is nil.
- `resp.Content` is exactly `""`.

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
	if resp.Content != "" {
		t.Fatalf("GetContent(%q) content = %q, want empty string", req.File, resp.Content)
	}
	if want := req.Expected[pathEmpty]; resp.Content != want {
		t.Fatalf("expected map mismatch for empty file: got %q want %q", resp.Content, want)
	}
}
```
