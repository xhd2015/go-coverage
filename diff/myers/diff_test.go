package myers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

// go test -run TestSame -v ./diff/myers
func TestSame(t *testing.T) {

	ops := operations(
		[]string{
			"what a luck",
			"a",
		},
		[]string{
			"b",
			"what the beck",
		},
	)

	t.Logf("ops:%+v", jsonstr(ops))
}

func TestSimilar(t *testing.T) {
	ops := operations(
		[]string{
			"1: one",
			"3: three",
		},
		[]string{
			"1: one",
			"2: two",
			"3: three",
		},
	)

	t.Logf("ops:%+v", jsonstr(ops))
}

func TestFileA(t *testing.T) {
	ops := operations(
		readLines("testdata/a_old.txt"),
		readLines("testdata/a_new.txt"),
	)

	t.Logf("ops:%+v", jsonstr(ops))
}

// go test -run TestBlockMapping -v ./diff/myers
func TestBlockMapping(t *testing.T) {
	m := ComputeBlockMapping(
		readLines("testdata/a_old.txt"),
		readLines("testdata/a_new.txt"),
	)
	s := fmt.Sprintf("%+v", m)
	// NOTE: 0-baesd
	exp := `map[1:1 2:2 3:3 4:4 5:5 10:6 16:8]`
	if s != exp {
		t.Fatalf("expect %s = %+v, actual:%+v", `s`, exp, s)
	}
}

func jsonstr(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(s)
}
func readLines(f string) []string {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return splitLines(string(data))
}
