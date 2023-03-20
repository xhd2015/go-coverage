package myers

import (
	"encoding/json"
	"testing"
)

// go test  -bench=BenchmarkDiff -benchtime=10s -run=NONE -v ./diff/myers
// result:  -            1585 ns/op
// latency: - ms
func BenchmarkDiff(b *testing.B) {
	// defer DestroyNow() // DON'T DO THIS, this will effectively close all thing
	for i := 0; i < b.N; i++ {
		res, _ := ComputeBlockMapping(
			[]string{"A", "B", "C"},
			[]string{"A", "B2", "C"},
		)
		resJSONBytes, err := json.Marshal(res)
		if err != nil {
			b.Fatal(err)
		}
		resJSON := string(resJSONBytes)
		resJSONExpect := `{"0":0,"2":2}`
		if resJSON != resJSONExpect {
			b.Fatalf("expect %s = %+v, actual:%+v", `resJSON`, resJSONExpect, resJSON)
		}
	}
}
