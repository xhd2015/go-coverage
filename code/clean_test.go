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

	s := Clean(f, CleanOpts{
		Log: true,
	})
	t.Logf("%s", s)
	if strings.Contains(s, "TODO:") {
		t.Fatalf("contains TODO")
	}
}

// NOTE: when running TestGo1_18Generic, we expect:
//    go1.18 to compile successfully and clean code correctly
//    go1.17 and under to compile also successfully, but clean code failed (because test data syntax is unknown to go1.17)
//
// go test -run TestGo1_18Generic -v ./code
// see coverage:
//    go test -run TestGo1_18Generic -coverprofile=cover.out -v ./code ;go tool cover -html=cover.out
func TestGo1_18Generic(t *testing.T) {
	f := parseFile("testdata/go1.18-generic.go.txt")

	s := Clean(f, CleanOpts{
		Log:             true,
		LogIndent:       4,
		DisallowUnknown: true,
	})
	t.Logf("%s", s)
	if strings.Contains(s, "TODO:") {
		t.Fatalf("contains TODO")
	}
}

// go test -run TestLoadProject -v ./code
func TestLoadProject(t *testing.T) {
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

// go test -run TestCleanSwitch -v ./code
func TestCleanSwitch(t *testing.T) {
	f := parseFile("testdata/switch.go.txt")

	s := Clean(f, CleanOpts{
		// Log: true,
	})
	t.Logf("%s", s)
	if strings.Contains(s, "TODO:") {
		t.Fatalf("contains TODO")
	}
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
