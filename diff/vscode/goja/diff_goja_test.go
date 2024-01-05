package goja

import (
	"encoding/json"
	"io/ioutil"
	"strings"
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

// go test -run TestDiffLines -v ./diff/vscode/goja
func TestDiffLines(t *testing.T) {
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
func readLines(file string) []string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(content), "\n")
}

// go test -run TestDiffFile -v ./diff/vscode/goja
func TestDiffFile(t *testing.T) {
	oldFile := "../testdata/a_old_3.txt"
	newFile := "../testdata/a_new_3.txt"

	res, err := Diff(&vscode.Request{
		OldLines: readLines(oldFile),
		NewLines: readLines(newFile),
	})
	if err != nil {
		t.Fatal(err)
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("res: %+v", string(resJSON))
}

// go test -run TestDiffContent -v ./diff/vscode/goja
func TestDiffContent(t *testing.T) {
	oldContent := `package main

func calculateLateCharge(principal string) string {
	lateChargeRate := GetConfig("late_charge_rate")
	if lateChargeRate == "" {
		return ZERO
	}
	return utils.Multiple(lateChargeRate, principal)
}`

	newContent := `package main

func calculateLateChargeV2(principal string) string {
	lateChargeRate := GetConfig("late_charge_rate_v2")
	deductRate := GetConfig("deduct_rate")
	if lateChargeRate == "" {
		log.Errorf("missing config late_charge_rate")
		return ZERO
	}
	if deductRate != ""{
		lateChargeRate = utils.Multiple(lateChargeRate, deductRate)
	}
	return utils.Multiple(lateChargeRate, principal)
}`
	res, err := Diff(&vscode.Request{
		OldLines: strings.Split(oldContent, "\n"),
		NewLines: strings.Split(newContent, "\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("res: %+v", string(resJSON))
}
