package git

import (
	"fmt"
	"sync"
)

type GitDiff struct {
	dir       string
	oldCommit string
	newCommit string

	oldGit *GitSnapshot
	newGit *GitSnapshot

	mergeOnce   sync.Once
	newToOld    map[string]string // merged updated and renamed
	newToOldErr error

	renameOnce sync.Once
	renames    map[string]string // renamed, new to old
	renameErr  error

	updateOnce  sync.Once
	updateFiles []string // file list
	updateErr   error
}

func NewGitDiff(dir string, oldCommit string, newCommit string) *GitDiff {
	return &GitDiff{
		dir:       dir,
		oldCommit: oldCommit,
		newCommit: newCommit,
		oldGit:    NewSnapshot(dir, oldCommit),
		newGit:    NewSnapshot(dir, newCommit),
	}
}

func (c *GitDiff) AllFiles() ([]string, error) {
	return c.newGit.ListFiles()
}
func (c *GitDiff) GetUpdateAndRenames() (newToOld map[string]string, err error) {
	c.mergeOnce.Do(func() {
		renames, err := c.GetRenames()
		if err != nil {
			c.newToOldErr = err
			return
		}
		updates, err := c.GetUpdates()
		if err != nil {
			c.newToOldErr = err
			return
		}
		newToOld := make(map[string]string, len(updates)+len(renames))
		for k, v := range renames {
			newToOld[k] = v
		}
		for _, u := range updates {
			if _, ok := renames[u]; ok {
				c.newToOldErr = fmt.Errorf("invalid file: %s found both renamed and updated", u)
				return
			}
			newToOld[u] = u
		}
		c.newToOld = newToOld
	})
	return c.newToOld, c.newToOldErr
}
func (c *GitDiff) GetRenames() (newToOld map[string]string, err error) {
	c.renameOnce.Do(func() {
		c.renames, c.renameErr = FindRenames(c.dir, c.oldCommit, c.newCommit)
	})
	return c.renames, c.renameErr
}
func (c *GitDiff) GetUpdates() ([]string, error) {
	c.updateOnce.Do(func() {
		repo := &GitRepo{Dir: c.dir}
		updates, err := repo.FindUpdate(c.oldCommit, c.newCommit)
		if err != nil {
			c.updateErr = err
			return
		}
		c.updateFiles = updates
	})
	return c.updateFiles, c.updateErr
}

func (c *GitDiff) GetNewContent(newFile string) (string, error) {
	return c.newGit.GetContent(newFile)
}

func (c *GitDiff) GetOldContent(oldFile string) (string, error) {
	return c.oldGit.GetContent(oldFile)
}
func (c *GitDiff) GetOldContentNewFile(newFile string) (string, error) {
	oldFile, ok, err := c.GetOldFile(newFile)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("file does not exist in old:%s", newFile)
	}
	return c.oldGit.GetContent(oldFile)
}

func (c *GitDiff) GetOldFile(newFile string) (oldFile string, ok bool, err error) {
	newToOld, err := c.GetUpdateAndRenames()
	if err != nil {
		return "", false, err
	}
	oldFile, ok = newToOld[newFile]
	return
}