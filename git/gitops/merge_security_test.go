package gitops_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	coveragegit "github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/line"
	"github.com/xhd2015/go-coverage/merge"
	"github.com/xhd2015/go-coverage/profile"
)

func TestMergeGitSecurityGetModPathUsesLiteralDirectory(t *testing.T) {
	probe, marker := markerProbe(t)
	for _, component := range []string{"module|" + probe, "module`" + probe + "`", "module with spaces and 'quotes'"} {
		dir := filepath.Join(t.TempDir(), component)
		putFile(t, dir, "go.mod", "module example.test/literal-directory\n\ngo 1.16\n")
		got, err := merge.GetModPath(dir)
		if err != nil || got != "example.test/literal-directory" {
			t.Fatalf("literal module directory: path=%q err=%v", got, err)
		}
		assertNoMarker(t, marker)
	}
	if _, err := merge.GetModPath(""); err == nil {
		t.Fatal("empty module directory changed from an error to the current directory")
	}
	if _, err := merge.GetModPath(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing module directory accepted")
	}
}

func TestMergeGitSecurityCoverageAndLineMappingWithLiteralRefs(t *testing.T) {
	probe, marker := markerProbe(t)
	dir := filepath.Join(t.TempDir(), "repo`"+probe+"`")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	makeDiffFixture(t, dir)
	oldSource := "package fixture\nfunc Calc(n int) int {\n return n + 1\n}\n"
	newSource := "package fixture\n\nfunc Calc(n int) int {\n return n + 1\n}\n"
	putFile(t, dir, "calc.go", oldSource)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "covered old function")
	oldOID := runGit(t, dir, "rev-parse", "HEAD")
	putFile(t, dir, "calc.go", newSource)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "shift unchanged function")
	newOID := runGit(t, dir, "rev-parse", "HEAD")
	oldRef, newRef := "old|"+probe, "new`"+probe+"`"
	runGit(t, dir, "branch", oldRef, oldOID)
	runGit(t, dir, "branch", newRef, newOID)
	parse := func(lineNo, count int) merge.Profile {
		t.Helper()
		p, err := profile.ParseProfile(fmt.Sprintf("mode: count\nexample.test/security/calc.go:%d.22,%d.2 1 %d\n", lineNo, lineNo+2, count))
		if err != nil {
			t.Fatal(err)
		}
		return merge.NewStdProfile(p)
	}
	mod, merged, _, err := merge.MergeGitDiff(parse(2, 3), parse(3, 2), "auto", dir, oldRef, newRef)
	if err != nil {
		t.Fatal(err)
	}
	if mod != "example.test/security" {
		t.Fatalf("module changed: %q", mod)
	}
	output := merged.(*merge.StdProfile).Output()
	if len(output.Blocks) != 1 || output.Blocks[0].Count != 5 {
		t.Fatalf("real mapped counters did not merge: %s", output.String())
	}
	mapping, _, err := line.CollectUnchangedLinesMapping(dir, oldRef, newRef)
	if err != nil {
		t.Fatal(err)
	}
	if mapping["calc.go"][3] != 2 {
		t.Fatalf("return-line mapping changed: %#v", mapping["calc.go"])
	}
	ranger, err := line.MakeGitDiffRanger(coveragegit.NewGitDiff(dir, oldRef, newRef), func(file string) bool { return file == "calc.go" })
	if err != nil {
		t.Fatal(err)
	}
	seen := false
	if err := ranger(func(file, oldFile, gotNew, gotOld string) bool {
		seen = true
		if file != "calc.go" || oldFile != "calc.go" || gotNew != newSource || gotOld != oldSource {
			t.Fatalf("snapshot content changed: %q %q", gotNew, gotOld)
		}
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("actual changed-file ranger was not executed")
	}
	for _, oldSide := range []bool{false, true} {
		old, next := oldRef, "missing|"+probe
		if oldSide {
			old, next = "missing|"+probe, newRef
		}
		if _, err := merge.MergeGit(parse(2, 3), parse(3, 2), "auto", dir, old, next); err == nil {
			t.Fatal("invalid ref passed through coverage merge")
		}
	}
	assertNoMarker(t, marker)
}
