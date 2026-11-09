package gitops_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/go-coverage/git/gitops"
)

type diffFixture struct{ dir, old, new string }

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func putFile(t *testing.T, dir, name, content string) {
	t.Helper()
	file := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func makeDiffFixture(t *testing.T, dir string) diffFixture {
	t.Helper()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.name", "security fixture")
	runGit(t, dir, "config", "user.email", "fixture@example.invalid")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	runGit(t, dir, "config", "diff.renames", "true")
	putFile(t, dir, "go.mod", "module example.test/security\n\ngo 1.16\n")
	putFile(t, dir, "unchanged.go", "package fixture\nfunc Keep() int { return 7 }\n")
	putFile(t, dir, "modified.go", "package fixture\nfunc Value() int { return 1 }\n")
	putFile(t, dir, "deleted.go", "package fixture\nfunc Gone() int { return 8 }\n")
	putFile(t, dir, "rename/old/item.go", "package fixture\nfunc Renamed() int { return 9 }\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "old")
	old := runGit(t, dir, "rev-parse", "HEAD")
	putFile(t, dir, "modified.go", "package fixture\nfunc Value() int { return 2 }\n")
	putFile(t, dir, "added.go", "package fixture\nfunc Added() int { return 3 }\n")
	if err := os.Remove(filepath.Join(dir, "deleted.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "rename/new"), 0755); err != nil {
		t.Fatal(err)
	}
	// Both absolute paths are fixed descendants of this fixture directory.
	from, to := filepath.Join(dir, "rename/old/item.go"), filepath.Join(dir, "rename/new/item.go")
	if err := os.Rename(from, to); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-qm", "new")
	return diffFixture{dir, old, runGit(t, dir, "rev-parse", "HEAD")}
}

func markerProbe(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "executed")
	probe := filepath.Join(dir, "probe")
	body := "#!/bin/sh\nprintf executed > '" + strings.ReplaceAll(marker, "'", "'\\''") + "'\n"
	if err := os.WriteFile(probe, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	return probe, marker
}

func assertNoMarker(t *testing.T, marker string) {
	t.Helper()
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("shell marker exists or stat failed: %v", err)
	}
}

func TestDiffCommitSecurityLiteralExistingRefs(t *testing.T) {
	f := makeDiffFixture(t, t.TempDir())
	opts := &gitops.DiffCommitOptions{PathPatterns: []string{"*.go"}}
	want, err := gitops.DiffCommit(f.dir, f.new, f.old, opts)
	if err != nil {
		t.Fatal(err)
	}
	probe, marker := markerProbe(t)
	for _, ref := range []string{"literal|" + probe, "literal`" + probe + "`"} {
		for _, oldSide := range []bool{false, true} {
			name := ref
			if oldSide {
				name = "old/" + ref
			}
			point := f.new
			if oldSide {
				point = f.old
			}
			runGit(t, f.dir, "branch", name, point)
			newRef, oldRef := name, f.old
			if oldSide {
				newRef, oldRef = f.new, name
			}
			got, err := gitops.DiffCommit(f.dir, newRef, oldRef, opts)
			if err != nil {
				t.Fatalf("legal literal ref rejected (%q): %v", name, err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("literal ref changed details: got=%+v want=%+v", got, want)
			}
			assertNoMarker(t, marker)
		}
	}
}

func TestDiffCommitSecurityRejectsInvalidRefsAndPreservesSentinels(t *testing.T) {
	f := makeDiffFixture(t, t.TempDir())
	probe, marker := markerProbe(t)
	bad := []string{"", "missing|" + probe, "missing`" + probe + "`", "--help", "--output=" + marker, "HEAD\n", "HEAD\r", "HEAD\t", "HEAD\x00", "HEAD\x7f", "does-not-exist"}
	for _, ref := range bad {
		for _, oldSide := range []bool{false, true} {
			newRef, oldRef := ref, f.old
			if oldSide {
				newRef, oldRef = f.new, ref
			}
			if _, err := gitops.DiffCommit(f.dir, newRef, oldRef, nil); err == nil {
				t.Fatalf("invalid ref accepted: oldSide=%v ref=%q", oldSide, ref)
			}
			assertNoMarker(t, marker)
		}
	}
	if _, err := gitops.DiffCommit(f.dir, f.new, gitops.COMMIT_WORKING, nil); err == nil {
		t.Fatal("WORKING comparison accepted")
	}
	if _, err := gitops.DiffCommit("", f.new, f.old, nil); err == nil {
		t.Fatal("empty repository accepted")
	}
}

func TestDiffCommitCompatibilityDetailsAndWorkingTree(t *testing.T) {
	f := makeDiffFixture(t, t.TempDir())
	opts := &gitops.DiffCommitOptions{PathPatterns: []string{"*.go"}}
	got, err := gitops.DiffCommit(f.dir, "HEAD", "HEAD~1", opts)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]*gitops.FileDetail{
		"unchanged.go":       {},
		"modified.go":        {ContentChanged: true},
		"deleted.go":         {Deleted: true},
		"added.go":           {IsNew: true},
		"rename/old/item.go": {Deleted: true},
		"rename/new/item.go": {RenamedFrom: "rename/old/item.go"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("details differ: got=%#v want=%#v", got, want)
	}
	putFile(t, f.dir, "modified.go", "package fixture\nfunc Value() int { return 4 }\n")
	putFile(t, f.dir, "untracked.go", "package fixture\n")
	putFile(t, f.dir, "staged.go", "package fixture\n")
	runGit(t, f.dir, "add", "staged.go")
	got, err = gitops.DiffCommit(f.dir, gitops.COMMIT_WORKING, f.new, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got["modified.go"] == nil || !got["modified.go"].ContentChanged || got["untracked.go"] == nil || !got["untracked.go"].IsNew || got["staged.go"] == nil || !got["staged.go"].IsNew {
		t.Fatalf("working-tree semantics changed: %#v", got)
	}
}

func TestDiffCommitSecurityDirectoryAndPathspecAreLiteral(t *testing.T) {
	probe, marker := markerProbe(t)
	for _, component := range []string{"repo|" + probe, "repo`" + probe + "`", "repo with spaces and 'quote'"} {
		dir := filepath.Join(t.TempDir(), component)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f := makeDiffFixture(t, dir)
		literal := "file|" + probe
		putFile(t, dir, literal, "literal file\n")
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-qm", "literal path")
		got, err := gitops.DiffCommit(dir, "HEAD", f.new, &gitops.DiffCommitOptions{PathPatterns: []string{literal}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[literal] == nil || !got[literal].IsNew {
			t.Fatalf("literal pathspec lost: %#v", got)
		}
		assertNoMarker(t, marker)
	}
}

func TestDiffCommitCompatibilitySHA256(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", "--object-format=sha256", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Git has no SHA-256 repository support: %v %s", err, out)
	}
	f := makeDiffFixture(t, dir)
	if len(f.new) != 64 {
		t.Fatalf("wanted SHA-256 OID: %q", f.new)
	}
	if _, err := gitops.DiffCommit(dir, f.new, f.old, nil); err != nil {
		t.Fatal(err)
	}
}
