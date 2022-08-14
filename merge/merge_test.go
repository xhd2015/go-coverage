package merge

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	new "github.com/xhd2015/go-coverage/merge/testdata/new"
	old "github.com/xhd2015/go-coverage/merge/testdata/old"
	"github.com/xhd2015/go-coverage/profile"
)

func TestSimpleOld(t *testing.T) {
	old.Calc(20, false)
	old.Calc(2, false)
}

func TestSimpleNew(t *testing.T) {
	new.Calc(context.Background(), 20)
}

// go test -run TestMergeProfile -v ./merge
func TestMergeProfile(t *testing.T) {
	testMergeProfile(t, "TestSimpleOld", "testdata/old/simple.go", "TestSimpleNew", "testdata/new/simple.go")
}

func testMergeProfile(t *testing.T, oldFn string, oldSrcFile string, newFn string, newSrcFile string) {
	tmpDir := path.Join(os.TempDir(), "merge-profile-test")
	err := os.MkdirAll(tmpDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("tmp: %v", tmpDir)

	oldCover := path.Join(tmpDir, "old.cover")
	newCover := path.Join(tmpDir, "new.cover")
	mergedCover := path.Join(tmpDir, "merged.cover")
	run := func(fn string, coverFile string, coverPkg string) {
		cmd := exec.Command("bash", "-x", "-c", fmt.Sprintf("go test -run %s -coverprofile=%s -coverpkg=%s .", fn, coverFile, coverPkg))
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}
	pkgOf := func(s string) string {
		x := path.Dir(s)
		return "./" + strings.TrimPrefix(x, "./")
	}
	// must ensure single file
	run(oldFn, oldCover, pkgOf(oldSrcFile))
	run(newFn, newCover, pkgOf(newSrcFile))

	oldProfile, err := profile.ParseProfileFile(oldCover)
	if err != nil {
		t.Fatal(err)
	}

	newProfile, err := profile.ParseProfileFile(newCover)
	if err != nil {
		t.Fatal(err)
	}

	readString := func(f string) (string, error) {
		bytes, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	oldCodeGetter := func(f string) (string, error) {
		return readString(oldSrcFile)
	}
	newCodeGetter := func(f string) (string, error) {
		return readString(newSrcFile)
	}

	mergedProfile, err := Merge(oldProfile, oldCodeGetter, newProfile, newCodeGetter, MergeOptions{
		GetOldFile: func(newFile string) string {
			// map new file to old file
			if strings.HasSuffix(newFile, newSrcFile) {
				return newFile[:len(newFile)-len(newSrcFile)] + oldSrcFile
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = mergedProfile.Write(mergedCover)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", mergedProfile.String())

	fmtCmd := func(f string) string {
		return fmt.Sprintf("go tool cover -html=%s", f)
	}
	t.Logf("profiles merged, run the followling cmds to see merged profile:\n  %s\n  %s\n  %s", fmtCmd(oldCover), fmtCmd(newCover), fmtCmd(mergedCover))
}
