package cover

import (
	"strings"
	"testing"

	"github.com/xhd2015/go-coverage/code"
)

// go test -run TestCollector -v ./cover/
func TestCollector(t *testing.T) {
	fset, ast, content, err := code.ParseFile("testdata/simple.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	stmts := CollectStmts(fset, ast, content)
	t.Logf("stmts:%v", strings.Join(stmts, "<<<\n"))
}
