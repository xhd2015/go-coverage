package goja

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dop251/goja"

	"github.com/xhd2015/go-coverage/diff/vscode"
)

// go test -run TestGojaUsage -v ./diff/vscode/goja
func TestGojaUsage(t *testing.T) {
	runtime := goja.New()
	type Obj struct {
		Hello string // must be exported, cannot have `json:"hello"`
	}
	err := runtime.Set("obj", &Obj{
		Hello: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = runtime.Set("varWorld", "world")
	if err != nil {
		t.Fatal(err)

	}
	res, err := runtime.RunString(`obj.Hello+" " + varWorld;`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("res: %v", res.ToString())
}

// go test -run TestGojaDiff -v ./diff/vscode/goja
func TestGojaDiff(t *testing.T) {
	runtime := goja.New()
	req := vscode.Request{
		OldLines: []string{"hello", "world"},
	}
	err := runtime.Set("request", req)
	if err != nil {
		t.Fatal(err)
	}
	res, err := runtime.RunString(`globalThis.request.OldLines;`)
	if err != nil {
		t.Fatal(err)
	}
	v := res.Export()
	t.Logf("res: %T %v", v, v)
}

// go test -run TestDiff -v ./diff/vscode/goja
func TestDiff(t *testing.T) {
	res, err := Diff(&vscode.Request{
		OldLines: []string{"A", "B", "C"},
		NewLines: []string{"A", "B2", "C"},
	})
	if err != nil {
		t.Fatal(err)
	}
	resJSONBytes, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	resJSON := string(resJSONBytes)
	resJSONExpect := `{"quitEarly":false,"changes":[{"originalStartLineNumber":2,"originalEndLineNumber":2,"modifiedStartLineNumber":2,"modifiedEndLineNumber":2}]}`
	if resJSON != resJSONExpect {
		t.Fatalf("expect %s = %+v, actual:%+v", `resJSON`, resJSONExpect, resJSON)
	}

	time.Sleep(100 * time.Millisecond)
	res2, err := Diff(&vscode.Request{
		OldLines: []string{"A1", "B", "C"},
		NewLines: []string{"A", "B2", "C"},
	})
	if err != nil {
		t.Fatal(err)
	}
	res2JSONBytes, err := json.Marshal(res2)
	if err != nil {
		t.Fatal(err)
	}
	res2JSON := string(res2JSONBytes)
	res2JSONExpect := `{"quitEarly":false,"changes":[{"originalStartLineNumber":1,"originalEndLineNumber":2,"modifiedStartLineNumber":1,"modifiedEndLineNumber":2}]}`
	if res2JSON != res2JSONExpect {
		t.Fatalf("expect %s = %+v, actual:%+v", `res2JSON`, res2JSONExpect, res2JSON)
	}
}

// go test  -bench=BenchmarkDiff -benchtime=10s -run=NONE -v ./diff/vscode/goja
// result:  -          590299 ns/op = 0.59ms/op, the stdin-stdout is 109311564 ns/op = 109ms/op, the myers is 1585 ns/op,  its 372x slower than that native go implementation, but 185x faster than stdin-stdout
// latency: - ms
func BenchmarkDiff(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := Diff(&vscode.Request{
			OldLines: []string{"A", "B", "C"},
			NewLines: []string{"A", "B2", "C"},
		})
		if err != nil {
			b.Fatal(err)
		}
		resJSONBytes, err := json.Marshal(res)
		if err != nil {
			b.Fatal(err)
		}
		resJSON := string(resJSONBytes)
		resJSONExpect := `{"quitEarly":false,"changes":[{"originalStartLineNumber":2,"originalEndLineNumber":2,"modifiedStartLineNumber":2,"modifiedEndLineNumber":2}]}`
		if resJSON != resJSONExpect {
			b.Fatalf("expect %s = %+v, actual:%+v", `resJSON`, resJSONExpect, resJSON)
		}
	}
}
