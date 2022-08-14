package merge

import (
	"fmt"
	"strings"

	"github.com/xhd2015/go-coverage/code"
	"github.com/xhd2015/go-coverage/cover"
	diff "github.com/xhd2015/go-coverage/diff/myers"
	"github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/profile"
)

func MergeGit(old *profile.Profile, new *profile.Profile, modPrefix string, dir string, oldCommit string, newCommit string) (*profile.Profile, error) {
	newToOld, err := git.FindUpdateAndRenames(dir, oldCommit, newCommit)
	if err != nil {
		return nil, err
	}

	oldGit := git.NewSnapshot(dir, oldCommit)
	newGit := git.NewSnapshot(dir, newCommit)
	getOldFile := func(newFile string) string {
		file := strings.TrimPrefix(newFile, modPrefix)
		file = strings.TrimPrefix(file, "/")
		file = strings.TrimPrefix(file, ".")
		return newToOld[file]
	}

	return Merge(old, oldGit.GetContent, new, newGit.GetContent, MergeOptions{
		GetOldFile: getOldFile,
	})
}

type MergeOptions struct {
	GetOldFile func(newFile string) string
}

// Merge merge 2 profiles with their code diffs
func Merge(old *profile.Profile, oldCodeGetter func(f string) (string, error), new *profile.Profile, newCodeGetter func(f string) (string, error), opts MergeOptions) (*profile.Profile, error) {
	oldCouners := old.Counters()
	newCounters := new.Counters()

	mergedCounters := make(map[string][]int, len(newCounters))
	for file, newCounter := range newCounters {
		var oldMustExist bool
		oldFile := file
		if opts.GetOldFile != nil {
			oldFile = opts.GetOldFile(file)
			if oldFile == "" {
				mergedCounters[file] = newCounter
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
