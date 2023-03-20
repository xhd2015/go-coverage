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
type DeletedLineMapping map[int]bool

// CollectUnchangedLinesMapping
func CollectUnchangedLinesMapping(dir string, oldCommit string, newCommit string) (map[string]LineMapping, map[string]DeletedLineMapping, error) {
	gitDiff := git.NewGitDiff(dir, oldCommit, newCommit)
	return CollectUnchangedLinesMappingWithDetails(gitDiff, nil)
}

func CollectUnchangedLinesMappingWithDetails(gitDiff *git.GitDiff, filterFile func(file string) bool) (map[string]LineMapping, map[string]DeletedLineMapping, error) {
	fileDetails, err := gitDiff.AllFilesDetailsV2()
	if err != nil {
		return nil, nil, err
	}
	mapping := make(map[model.PkgFile]LineMapping, len(fileDetails))
	deleteMapping := make(map[model.PkgFile]DeletedLineMapping, len(fileDetails))
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
			return nil, nil, err
		}
		oldContent, err := gitDiff.GetOldContent(oldFile)
		if err != nil {
			return nil, nil, err
		}

		newLines := strings.Split(newContent, "\n")
		oldLines := strings.Split(oldContent, "\n")

		lineMapping, deletedLines := diff.ComputeBlockMapping(oldLines, newLines)

		trimFile := strings.TrimPrefix(file, "/")
		mapping[trimFile] = lineMapping
		deleteMapping[trimFile] = deletedLines
	}
	return mapping, deleteMapping, nil
}
