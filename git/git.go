package git

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xhd2015/go-coverage/sh"
)

// find rename
// git diff --find-renames --diff-filter=R   HEAD~10 HEAD|grep -A 3 '^diff --git a/'|grep rename
// FindRenames returns a mapping from new name to old name
func FindRenames(dir string, oldCommit string, newCommit string) (map[string]string, error) {
	repo := &GitRepo{Dir: dir}
	return repo.FindRenames(oldCommit, newCommit)
}

func FindUpdate(dir string, oldCommit string, newCommit string) ([]string, error) {
	repo := &GitRepo{Dir: dir}
	return repo.FindUpdate(oldCommit, newCommit)
}

func FindUpdateAndRenames(dir string, oldCommit string, newCommit string) (map[string]string, error) {
	repo := &GitRepo{Dir: dir}
	updates, err := repo.FindUpdate(oldCommit, newCommit)
	if err != nil {
		return nil, err
	}
	m, err := repo.FindRenames(oldCommit, newCommit)
	if err != nil {
		return nil, err
	}
	for _, u := range updates {
		if _, ok := m[u]; ok {
			return nil, fmt.Errorf("invalid file: %s found both renamed and updated", u)
		}
		m[u] = u
	}
	return m, nil
}

type GitRepo struct {
	Dir string
}

func NewGitRepo(dir string) *GitRepo {
	return &GitRepo{
		Dir: dir,
	}
}
func NewSnapshot(dir string, commit string) *GitSnapshot {
	return &GitSnapshot{
		Dir:    dir,
		Commit: commit,
	}
}

func (c *GitRepo) FindUpdate(oldCommit string, newCommit string) ([]string, error) {
	cmd := fmt.Sprintf(`git -C %s diff --diff-filter=M --name-only %s %s`, sh.Quote(c.Dir), sh.Quote(getRef(oldCommit)), sh.Quote(getRef(newCommit)))
	stdout, _, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		NeedStdOut: true,
	})
	if err != nil {
		return nil, err
	}
	return splitLinesFilterEmpty(stdout), nil
}
func (c *GitRepo) FindRenames(oldCommit string, newCommit string) (map[string]string, error) {
	// example:
	// 	$ git diff --find-renames --diff-filter=R   HEAD~10 HEAD
	// diff --git a/test/stubv2/boot/boot.go b/test/stub/boot/boot.go
	// similarity index 94%
	// rename from test/stubv2/boot/boot.go
	// rename to test/stub/boot/boot.go
	// index e0e86051..56c49801 100644
	// --- a/test/stubv2/boot/boot.go
	// +++ b/test/stub/boot/boot.go
	// @@ -4,8 +4,10 @@ import (
	cmd := fmt.Sprintf(`git -C %s diff --find-renames --diff-filter=R %s %s|grep -A 3 '^diff --git a/'|grep rename`, sh.Quote(c.Dir), sh.Quote(getRef(oldCommit)), sh.Quote(getRef(newCommit)))
	stdout, _, err := sh.RunBashCmdOpts(cmd, sh.RunBashOptions{
		// Verbose:    true,
		NeedStdOut: true,
	})
	if err != nil {
		return nil, err
	}
	lines := splitLinesFilterEmpty(stdout)
	if len(lines)%2 != 0 {
		return nil, fmt.Errorf("internal error, expect git return rename pairs, found:%d", len(lines))
	}

	m := make(map[string]string, len(lines)/2)
	for i := 0; i < len(lines); i += 2 {
		from := strings.TrimPrefix(lines[i], "rename from ")
		to := strings.TrimPrefix(lines[i+1], "rename to ")

		m[to] = from
	}
	return m, nil
}

type GitSnapshot struct {
	Dir    string
	Commit string

	filesInit sync.Once
	files     []string
	filesErr  error
	fileMap   map[string]bool
}

func (c *GitSnapshot) GetContent(file string) (string, error) {
	normFile := strings.TrimPrefix(file, "./")
	if normFile == "" {
		return "", fmt.Errorf("invalid file:%v", file)
	}
	if !c.fileMap[normFile] {
		return "", fmt.Errorf("not a file, maybe a dir:%v", file)
	}

	content, _, err := sh.RunBashWithOpts([]string{
		fmt.Sprintf("git -C %s cat-file -p %s:%s", sh.Quote(c.Dir), sh.Quote(c.ref()), sh.Quote(normFile)),
	}, sh.RunBashOptions{
		NeedStdOut: true,
	})
	return content, err
}
func (c *GitSnapshot) ListFiles() ([]string, error) {
	c.filesInit.Do(func() {
		stdout, _, err := sh.RunBashWithOpts([]string{
			fmt.Sprintf("git -C %s ls-files --with-tree %s", sh.Quote(c.Dir), sh.Quote(c.ref())),
		}, sh.RunBashOptions{
			Verbose:    true,
			NeedStdOut: true,
		})
		if err != nil {
			c.filesErr = err
			return
		}
		c.files = splitLinesFilterEmpty(stdout)
		c.fileMap = make(map[string]bool)
		for _, e := range c.files {
			c.fileMap[e] = true
		}
	})
	return c.files, c.filesErr
}

func (c *GitSnapshot) ref() string {
	return getRef(c.Commit)
}

func getRef(commit string) string {
	if commit == "" {
		return "HEAD"
	}
	return commit
}

func splitLinesFilterEmpty(s string) []string {
	list := strings.Split(s, "\n")
	idx := 0
	for _, e := range list {
		if e != "" {
			list[idx] = e
			idx++
		}
	}
	return list[:idx]
}
