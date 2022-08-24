package merge

import (
	"fmt"
	"strings"

	"github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/sh"
)

func MergeGit(old Profile, new Profile, modPrefix string, dir string, oldCommit string, newCommit string) (Profile, error) {
	if modPrefix == "" || modPrefix == "auto" {
		var err error
		modPrefix, err = GetModPath(dir)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(modPrefix, "/") || strings.HasSuffix(modPrefix, "/") {
		return nil, fmt.Errorf("modPrefix must not start or end with '/':%s", modPrefix)
	}

	gitDiff := git.NewGitDiff(dir, oldCommit, newCommit)
	newToOld, err := gitDiff.GetUpdateAndRenames()
	if err != nil {
		return nil, err
	}
	getUpdatedFile := func(newFile string) string {
		file := strings.TrimPrefix(newFile, modPrefix)
		file = strings.TrimPrefix(file, "/")
		file = strings.TrimPrefix(file, ".")
		oldFile := newToOld[file]
		if oldFile == "" {
			return ""
		}
		return modPrefix + "/" + oldFile
	}

	getOldContent := func(file string) (string, error) {
		return gitDiff.GetOldContent(strings.TrimPrefix(file, modPrefix+"/"))
	}
	getNewContent := func(file string) (string, error) {
		return gitDiff.GetNewContent(strings.TrimPrefix(file, modPrefix+"/"))
	}

	return Merge(old, getOldContent, new, getNewContent, MergeOptions{
		GetUpdatedFile: getUpdatedFile,
	})
}

type MergeOptions struct {
	// GetUpdatedFile only return file that have changed
	GetUpdatedFile func(newFile string) string
}

// Merge merge 2 profiles with their code diffs
func Merge(old Profile, oldCodeGetter func(f string) (string, error), newProfile Profile, newCodeGetter func(f string) (string, error), opts MergeOptions) (Profile, error) {
	var err error
	res := newProfile.Clone()
	newProfile.RangeCounters(func(pkgFile string, newCounters Counters) bool {
		var oldMustExist bool
		oldFile := pkgFile
		if opts.GetUpdatedFile != nil {
			oldFile = opts.GetUpdatedFile(pkgFile)
			if oldFile == "" {
				oldCounters := old.GetCounters(pkgFile)
				if oldCounters == nil {
					res.SetCounters(pkgFile, newCounters)
				} else {
					if newCounters.Len() != oldCounters.Len() {
						err = fmt.Errorf("unchanged file found different lenght of counters: file=%s, old=%d, new=%d", pkgFile, oldCounters.Len(), newCounters.Len())
						return false
					}
					// plain merge
					addedCounters := newCounters.New(newCounters.Len())
					for i := 0; i < newCounters.Len(); i++ {
						addedCounters.Set(i, newCounters.Get(i).Add(oldCounters.Get(i)))
					}
					res.SetCounters(pkgFile, addedCounters)
				}
				return true
			}
			oldMustExist = true
		}

		oldCounter := old.GetCounters(oldFile)
		if oldCounter == nil {
			if oldMustExist {
				err = fmt.Errorf("counters not found for old file %s", oldFile)
				return false
			}
			res.SetCounters(pkgFile, newCounters)
			return true
		}
		var oldCode string
		var newCode string
		oldCode, err = oldCodeGetter(pkgFile)
		if err != nil {
			return false
		}
		newCode, err = newCodeGetter(pkgFile)
		if err != nil {
			return false
		}
		var mergedCounter Counters
		mergedCounter, err = MergeFileNameCounters(oldCounter, oldFile, oldCode, newCounters, pkgFile, newCode)
		if err != nil {
			return false
		}
		res.SetCounters(pkgFile, mergedCounter)
		return true
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MergeFileCounter merge counters of the same file between two commits,
// the algorithm takes semantic update into consideration, making the
// merge more accurate while strict.
func MergeFileCounter(oldCounter Counters, oldCode string, newCounter Counters, newCode string) (mergedCounters Counters, err error) {
	return MergeFileNameCounters(oldCounter, "old.go", oldCode, newCounter, "new.go", newCode)
}

func MergeFileNameCounters(oldCounter Counters, oldFileName string, oldCode string, newCounter Counters, newFileName string, newCode string) (mergedCounters Counters, err error) {
	newToOld, err := ComputeFileBlockMapping(oldFileName, oldCode, newFileName, newCode)
	mergedCounters = newCounter.New(newCounter.Len())
	for i := 0; i < newCounter.Len(); i++ {
		c := newCounter.Get(i)
		if oldIdx, ok := newToOld[i]; ok {
			c = c.Add(oldCounter.Get(oldIdx))
		}
		mergedCounters.Set(i, c)
	}
	return
}

func GetModPath(dir string) (modPath string, err error) {
	// try to read mod from dir
	var mod struct {
		Module struct {
			Path string
		}
	}
	_, _, err = sh.RunBashCmdOpts(fmt.Sprintf(`cd %s && go mod edit -json`, sh.Quote(dir)), sh.RunBashOptions{
		StdoutToJSON: &mod,
	})
	if err != nil {
		err = fmt.Errorf("get module path: %v", err)
		return
	}
	modPath = mod.Module.Path
	return
}
