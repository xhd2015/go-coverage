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
	testBlockMapping(t, "testdata/a_old.txt", "testdata/a_new.txt")
}

// go test -run TestBlockMappingV2 -v ./diff/myers
func TestBlockMappingV2(t *testing.T) {
	testBlockMappingV2(t, "testdata/b_del_update_old.txt", "testdata/b_del_update_new.txt")
}

// go test -run TestBlockMapping -v ./diff/myers
func testBlockMapping(t *testing.T, oldFile string, newFile string) {
	m := ComputeBlockMapping(
		readLines(oldFile),
		readLines(newFile),
	)
	s := fmt.Sprintf("%+v", m)
	// NOTE: 0-baesd
	exp := `map[1:1 2:2 3:3 4:4 5:5 10:6 16:8]`
	if s != exp {
		t.Fatalf("expect %s = %+v, actual:%+v", `s`, exp, s)
	}
}

func testBlockMappingV2(t *testing.T, oldFile string, newFile string) {
	m := ComputeBlockMappingV2(
		readLines(oldFile),
		readLines(newFile),
	)
	s := fmt.Sprintf("%+v", m)
	t.Logf("res:%v", s)
	// // NOTE: 0-baesd
	// exp := `map[1:1 2:2 3:3 4:4 5:5 10:6 16:8]`
	// if s != exp {
	// 	t.Fatalf("expect %s = %+v, actual:%+v", `s`, exp, s)
	// }
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
