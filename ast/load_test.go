package ast

import (
	"context"
	"os"
	"testing"
	"time"

	"golang.org/x/tools/go/packages"
)

// go test -run TestLoad_NeedName -v ./ast
func TestLoad_NeedName(t *testing.T) {
	modes := []LoadMode{
		LoadMode_NeedName,
	}
	// load_test.go:34: loaded 1 packages,modes=[LoadMode(NeedName)], cost: 214.856076ms
	doTestLoadSingle(t, modes)
}

// go test -run TestLoad_NeedName_NeedDeps -v ./ast
func TestLoad_NeedName_NeedDeps(t *testing.T) {
	modes := []LoadMode{
		LoadMode_NeedName,
		LoadMode_NeedDeps, // adds imported fields
	}
	// loaded 1 packages,modes=[LoadMode(NeedName) LoadMode(NeedDeps)], cost: 894.678306ms
	doTestLoadSingle(t, modes)
}

// go test -run TestLoad_NeedImports -v ./ast
func TestLoad_NeedImports(t *testing.T) {
	modes := []LoadMode{
		LoadMode_NeedImports, // adds imported fields
	}
	// loaded 927 packages,modes=[LoadMode(NeedImports)], cost: 832.167989ms
	doTestLoadSingle(t, modes)
}

// go test -run TestLoad_NeedSyntaxSrcAll -v ./ast
func TestLoad_NeedSyntaxSrcAll(t *testing.T) {
	modes := []LoadMode{
		LoadMode_NeedSyntax, // adds imported fields
		LoadMode_NeedName,
	}
	// loaded 55 packages,modes=[LoadMode(NeedSyntax)], cost: 5.137941135s
	// loaded 55 packages,modes=[LoadMode(NeedSyntax)], cost: 908.40577ms  (even without cache?)

	// loaded 55 packages,modes=[LoadMode(NeedSyntax) LoadMode(NeedName)], cost: 1.280476403s
	doTestLoad(t, modes, []string{"./src/..."})
}

func doTestLoadSingle(t *testing.T, modes []LoadMode) {
	doTestLoad(t, modes, []string{"./src"})
}
func doTestLoad(t *testing.T, modes []LoadMode, args []string) {
	testRepoDir := os.Getenv("TEST_REPO_DIR")
	if testRepoDir == "" {
		t.Fatalf("requires TEST_REPO_DIR")
	}

	begin := time.Now()
	ctx := context.Background()
	pkgs, _, err := LoadPackages(ctx, testRepoDir, args, &LoadOptions{
		Modes:      modes,
		BuildFlags: []string{"-mod=vendor", "-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var allPkgs []*packages.Package
	packages.Visit(pkgs, func(p *packages.Package) bool {
		allPkgs = append(allPkgs, p)
		return true
	}, nil)
	t.Logf("loaded %v packages,modes=%v, cost: %v", len(allPkgs), modes, time.Since(begin))
}
