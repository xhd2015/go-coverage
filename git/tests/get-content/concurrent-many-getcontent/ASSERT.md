## Expected

- Aggregate Run error is nil (every concurrent GetContent succeeded).
- `len(resp.Hits) == req.Concurrent` (30).
- Each hit has empty `ErrMsg` and `Content == req.Expected[hit.File]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("concurrent GetContent: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.Hits) != req.Concurrent {
		t.Fatalf("hits = %d, want %d", len(resp.Hits), req.Concurrent)
	}
	for i, hit := range resp.Hits {
		if hit.ErrMsg != "" {
			t.Fatalf("hit[%d] file=%q err=%q", i, hit.File, hit.ErrMsg)
		}
		want, ok := req.Expected[hit.File]
		if !ok {
			t.Fatalf("hit[%d] unexpected file %q", i, hit.File)
		}
		if hit.Content != want {
			t.Fatalf("hit[%d] file=%q content = %q, want %q", i, hit.File, hit.Content, want)
		}
	}
}
```
