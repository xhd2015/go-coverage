package line

import (
	"io/ioutil"
	"testing"

	"github.com/xhd2015/go-coverage/code"
)

// go test -run TestCollectEmptyLines -v ./line
func TestCollectEmptyLines(t *testing.T) {
	testCode, err := ioutil.ReadFile("./testdata/line_test.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	fset, file, err := code.ParseCodeString("lines_test.go", string(testCode))
	if err != nil {
		t.Fatal(err)
	}

	lines := CollectEmptyLinesForFile(fset, file)

	t.Logf("%+v", lines)

}
