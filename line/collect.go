package line

import (
	"strings"

	diff "github.com/xhd2015/go-coverage/diff/myers"
	"github.com/xhd2015/go-coverage/git"
	"github.com/xhd2015/go-coverage/model"
)

// type ChangeType string

// const (
// 	ChangeTypeUnchanged ChangeType = "unchanged"
// 	ChangeTypeNew       ChangeType = "new"
// 	ChangeTypeUpdated   ChangeType = "updated"
// )

// type FileLines struct {
// 	ChangeType  ChangeType
// 	LineMapping map[int]int // effective when ChangeTypeUpdated==updated
// }

type LineMapping map[int]int

// CollectUnchangedLinesMapping
func CollectUnchangedLinesMapping(dir string, oldCommit string, newCommit string) (map[string]LineMapping, error) {
	gitDiff := git.NewGitDiff(dir, oldCommit, newCommit)
	return CollectUnchangedLinesMappingWithDetails(gitDiff, nil)
}

func CollectUnchangedLinesMappingWithDetails(gitDiff *git.GitDiff, filterFile func(file string) bool) (map[string]LineMapping, error) {
	fileDetails, err := gitDiff.AllFilesDetailsV2()
	if err != nil {
		return nil, err
	}
	mapping := make(map[model.PkgFile]LineMapping, len(fileDetails))
	for file, fd := range fileDetails {
		if filterFile != nil && !filterFile(file) {
			continue
		}
		if fd.IsNew || !fd.ContentChanged {
			continue
		}
		oldFile := file
		if fd.RenamedFrom != "" {
			oldFile = fd.RenamedFrom
		}

		// get content
		newContent, err := gitDiff.GetNewContent(file)
		if err != nil {
			return nil, err
		}
		oldContent, err := gitDiff.GetOldContent(oldFile)
		if err != nil {
			return nil, err
		}

		newLines := strings.Split(newContent, "\n")
		oldLines := strings.Split(oldContent, "\n")

		lineMapping := diff.ComputeBlockMapping(oldLines, newLines)

		mapping[strings.TrimPrefix(file, "/")] = lineMapping
	}
	return mapping, nil
}
