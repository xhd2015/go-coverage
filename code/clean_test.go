package code

import (
	"context"
	goast "go/ast"
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/xhd2015/go-coverage/ast"
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

// go test -run TestProject -v ./code
func TestProject(t *testing.T) {
	ctx := context.Background()
	dir := os.Getenv("TEST_DIR")
	if dir == "" {
		t.Fatalf("requires dir")
	}
	_, _, pkgs, err := ast.LoadSyntaxOnly(ctx, dir, []string{"./src/..."}, []string{"-mod=vendor"})
	if err != nil {
		t.Fatal(err)
	}

	packages.Visit(pkgs, func(p *packages.Package) bool {
		for _, f := range p.Syntax {
			Clean(f, CleanOpts{})
		}
		return true
	}, nil)

}

func parseFile(f string) *goast.File {
	fset, ast, content, err := ParseFile(f)
	if err != nil {
		panic(err)
	}
	_ = fset
	_ = content
	return ast
}
