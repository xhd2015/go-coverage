package vscode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

// go test -run TestComputeBlockMapping -v ./diff/vscode
func TestComputeBlockMapping(t *testing.T) {
	initTest()
	defer DestroyNow()
	res, err := ComputeBlockMapping(
		[]string{"A", "B", "C"},
		[]string{"A", "B2", "C"},
	)
	if err != nil {
		t.Fatal(err)
	}
	resJSONBytes, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	resJSON := string(resJSONBytes)
	resJSONExpect := `{"0":0,"2":2}`
	if resJSON != resJSONExpect {
		t.Fatalf("expect %s = %+v, actual:%+v", `resJSON`, resJSONExpect, resJSON)
	}
}

// go test -run TestBlockMapping -v ./diff/vscode
func TestBlockMapping(t *testing.T) {
	testBlockMapping(t, "testdata/a_old.txt", "testdata/a_new.txt")
}

func testBlockMapping(t *testing.T, oldFile string, newFile string) {
	m, err := ComputeBlockMapping(
		readLines(oldFile),
		readLines(newFile),
	)
	if err != nil {
		t.Fatal(err)
	}

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

func splitLines(text string) []string {
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
