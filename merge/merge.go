package merge

import (
	"fmt"
	"strings"

	"github.com/xhd2015/go-coverage/code"
	"github.com/xhd2015/go-coverage/cover"
	diff "github.com/xhd2015/go-coverage/diff/myers"
	"github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/profile"
	"github.com/xhd2015/go-coverage/sh"
)

func MergeGit(old *profile.Profile, new *profile.Profile, modPrefix string, dir string, oldCommit string, newCommit string) (*profile.Profile, error) {
	if modPrefix == "" || modPrefix == "auto" {
		var err error
		modPrefix, err = GetModPath(dir)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(modPrefix, "/") || strings.HasSuffix(modPrefix, "/") {
		return nil, fmt.Errorf("modPrefix must not start or end with '/':%s", modPrefix)
	}

	newToOld, err := git.FindUpdateAndRenames(dir, oldCommit, newCommit)
	if err != nil {
		return nil, err
	}

	oldGit := git.NewSnapshot(dir, oldCommit)
	newGit := git.NewSnapshot(dir, newCommit)
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
		return oldGit.GetContent(strings.TrimPrefix(file, modPrefix+"/"))
	}
	getNewContent := func(file string) (string, error) {
		return newGit.GetContent(strings.TrimPrefix(file, modPrefix+"/"))
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
func Merge(old *profile.Profile, oldCodeGetter func(f string) (string, error), new *profile.Profile, newCodeGetter func(f string) (string, error), opts MergeOptions) (*profile.Profile, error) {
	oldCouners := old.Counters()
	newCounters := new.Counters()

	mergedCounters := make(map[string][]int, len(newCounters))
	for file, newCounter := range newCounters {
		var oldMustExist bool
		oldFile := file
		if opts.GetUpdatedFile != nil {
			oldFile = opts.GetUpdatedFile(file)
			if oldFile == "" {
				oldCounter, ok := oldCouners[file]
				if !ok {
					mergedCounters[file] = newCounter
				} else {
					if len(newCounter) != len(oldCounter) {
						return nil, fmt.Errorf("unchanged file found different lenght of counters: file=%s, old=%d, new=%d", file, len(oldCounter), len(newCounter))
					}
					// plain merge
					addedCounters := make([]int, len(newCounter))
					for i := 0; i < len(newCounter); i++ {
						addedCounters[i] = newCounter[i] + oldCounter[i]
					}
					mergedCounters[file] = addedCounters
				}
				continue
			}
			oldMustExist = true
		}

		oldCounter, ok := oldCouners[oldFile]
		if !ok {
			if oldMustExist {
				return nil, fmt.Errorf("counters not found for old file %s", oldFile)
			}
			mergedCounters[file] = newCounter
			continue
		}
		oldCode, err := oldCodeGetter(file)
		if err != nil {
			return nil, err
		}
		newCode, err := newCodeGetter(file)
		if err != nil {
			return nil, err
		}
		mergedCounter, err := MergeFileCounter(oldCounter, oldCode, newCounter, newCode)
		if err != nil {
			return nil, err
		}
		mergedCounters[file] = mergedCounter
	}

	res := new.Clone()
	res.ResetCounters(mergedCounters)
	return res, nil
}

// MergeFileCounter merge counters of the same file between two commits,
// the algorithm takes semantic update into consideration, making the
// merge more accurate while strict.
func MergeFileCounter(oldCounter []int, oldCode string, newCounter []int, newCode string) (mergedCounters []int, err error) {
	oldFset, oldAst, err := code.ParseCodeString("old.go", oldCode)
	if err != nil {
		return
	}
	newFset, newAst, err := code.ParseCodeString("new.go", newCode)
	if err != nil {
		return
	}
	oldBlocks := cover.CollectStmts(oldFset, oldAst, []byte(oldCode))
	if len(oldCounter) != len(oldBlocks) {
		err = fmt.Errorf("inconsistent old block(%d) and counter(%d)", len(oldBlocks), len(oldCounter))
		return
	}
	newBlocks := cover.CollectStmts(newFset, newAst, []byte(newCode))
	if len(newCounter) != len(newBlocks) {
		err = fmt.Errorf("inconsistent new block(%d) and counter(%d)", len(newBlocks), len(newCounter))
		return
	}

	newToOld := diff.ComputeBlockMapping(oldBlocks, newBlocks)
	mergedCounters = append([]int(nil), newCounter...)

	for i, c := range newCounter {
		if oldIdx, ok := newToOld[i]; ok {
			c += oldCounter[oldIdx]
		}
		mergedCounters[i] = c
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
