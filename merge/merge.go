package merge

import (
	"fmt"

	"github.com/xhd2015/go-coverage/code"
	"github.com/xhd2015/go-coverage/cover"
	diff "github.com/xhd2015/go-coverage/diff/myers"
	"github.com/xhd2015/go-coverage/profile"
)

// Merge merge 2 profiles with their code diffs
func Merge(old *profile.Profile, oldCodeGetter func(f string) (string, error), new *profile.Profile, newCodeGetter func(f string) (string, error)) (*profile.Profile, error) {
	oldCouners := old.Counters()
	newCounters := new.Counters()

	mergedCounters := make(map[string][]int, len(newCounters))
	for file, newCounter := range newCounters {
		oldCounter, ok := oldCouners[file]
		if !ok {
			// TODO: detect file rename
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
