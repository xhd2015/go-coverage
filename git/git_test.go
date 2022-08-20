package git

import (
	"os"
	"testing"
)

var dir string

func init() {
	dir = os.Getenv("TEST_DIR")
}

// go test -run TestFindUpdate -v ./git
func TestFindUpdate(t *testing.T) {
	files, err := FindUpdate(dir, "HEAD~10", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("files:%v", files)
}

// go test -run TestFindRename -v ./git
func TestFindRename(t *testing.T) {
	files, err := FindRenames(dir, "origin/master~50", "origin/master")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("files:%v", files)
}

// go test -run FindUpdateAndRenames -v ./git
func TestFindUpdateAndRenames(t *testing.T) {
	files, err := FindUpdateAndRenames(dir, "HEAD~10", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("files:%v", files)
}

// go test -run TestGitSnapshot -v ./git
func TestGitSnapshot(t *testing.T) {
	git := &GitSnapshot{
		Dir:    dir,
		Commit: "HEAD",
	}
	files, err := git.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("files:%v", files)

	content, err := git.GetContent(files[0])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content:%v", content)
}
