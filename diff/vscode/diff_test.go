package vscode

import (
	"encoding/json"
	"testing"
	"time"
)

func initTest() {
	disableDebugLog = false
}

// go test -run TestDiff -v ./diff/vscode
func TestDiff(t *testing.T) {
	initTest()
	defer DestroyNow()
	res, err := Diff(&Request{
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
	res2, err := Diff(&Request{
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

// go test -run TestKeepAliveAfter20s -v ./diff/vscode
func TestKeepAliveAfter20s(t *testing.T) {
	initTest()
	defer DestroyNow()
	res, err := Diff(&Request{
		OldLines: []string{"A", "B", "C"},
		NewLines: []string{"A", "B2", "C"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = res
	time.Sleep(20 * time.Second)

	res, err = Diff(&Request{
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
}

// go test  -bench=BenchmarkDiff -benchtime=10s -run=NONE -v ./diff/vscode
// result:  -           10845941 ns/op = 10.8ms/op, the myers is 1585 ns/op,  its 6842x slower than that native go implementation.
// latency: - ms
func BenchmarkDiff(b *testing.B) {
	// defer DestroyNow() // DON'T DO THIS, this will effectively close all thing
	for i := 0; i < b.N; i++ {
		res, err := Diff(&Request{
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
