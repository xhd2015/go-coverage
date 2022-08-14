package code

import (
	"go/ast"
	"strings"
	"testing"
)

// go test -run TestBasic -v ./code
// see coverage:
//    go test -run TestBasic -coverprofile=cover.out -v ./code ;go tool cover -html=cover.out
func TestBasic(t *testing.T) {
	f := parseFile("testdata/basic.go.txt")

	s := Clean(f, CleanOpts{})
	t.Logf("%s", s)
	if strings.Contains(s, "TODO:") {
		t.Fatalf("contains TODO")
	}
}

func parseFile(f string) *ast.File {
	fset, ast, content, err := ParseFile(f)
	if err != nil {
		panic(err)
	}
	_ = fset
	_ = content
	return ast
}
