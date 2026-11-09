package gitops_test

import (
	"strings"
	"testing"

	coveragegit "github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/line"
)

func TestGitDiffSecurityOldFileLookupStaysOnVerifiedDetails(t *testing.T) {
	f := makeDiffFixture(t, t.TempDir())
	probe, marker := markerProbe(t)
	oldRef, newRef := "old|"+probe, "new`"+probe+"`"
	runGit(t, f.dir, "branch", oldRef, f.old)
	runGit(t, f.dir, "branch", newRef, f.new)
	diff := coveragegit.NewGitDiff(f.dir, oldRef, newRef)
	for _, tc := range []struct {
		file, old string
		ok        bool
	}{
		{"modified.go", "modified.go", true},
		{"rename/new/item.go", "rename/old/item.go", true},
		{"unchanged.go", "", false},
		{"added.go", "", false},
		{"deleted.go", "", false},
		{"absent.go", "", false},
	} {
		old, ok, err := diff.GetOldFile(tc.file)
		if err != nil || old != tc.old || ok != tc.ok {
			t.Fatalf("GetOldFile(%q)=(%q,%v,%v), want (%q,%v)", tc.file, old, ok, err, tc.old, tc.ok)
		}
		assertNoMarker(t, marker)
	}
	content, err := diff.GetOldContentNewFile("modified.go")
	if err != nil || !strings.Contains(content, "return 1") {
		t.Fatalf("old snapshot content: %q %v", content, err)
	}
	// Match the real line-annotation caller's nested GetOldFile predicate.
	ranger, err := line.MakeGitDiffRanger(diff, func(file string) bool {
		_, ok, err := diff.GetOldFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return ok
	})
	if err != nil {
		t.Fatal(err)
	}
	seen := false
	if err := ranger(func(file, oldFile, next, old string) bool {
		if file == "modified.go" {
			seen = true
			if oldFile != "modified.go" || !strings.Contains(old, "return 1") || !strings.Contains(next, "return 2") {
				t.Fatal("ranger returned wrong actual content")
			}
		}
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("the real nested old-file lookup skipped a modified file")
	}
	assertNoMarker(t, marker)
}
