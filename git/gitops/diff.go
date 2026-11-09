package gitops

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/go-coverage/model"
)

const COMMIT_WORKING = "WORKING"

type FileDetail = model.FileDetail

type DiffCommitOptions struct {
	PathPatterns []string `json:"pathPatterns"` // only include these patterns, example: src/**/*.go
}

// DiffCommit finds file changes between two files
func DiffCommit(dir string, ref string, compareRef string, options *DiffCommitOptions) (fileDetailsMap map[string]*FileDetail, err error) {
	if ref == "" {
		err = fmt.Errorf("requires ref")
		return
	}
	if compareRef == "" {
		err = fmt.Errorf("requires compareRef")
		return
	}
	if compareRef == COMMIT_WORKING {
		err = fmt.Errorf("compareRef cannot be WORKING commit")
		return
	}

	// Resolve caller-controlled revisions once. Subsequent commands receive only
	// verified commit object IDs (or no new revision for the working tree).
	var newOID string
	if ref != COMMIT_WORKING {
		newOID, err = resolveDiffCommit(dir, ref)
		if err != nil {
			return nil, err
		}
	}
	oldOID, err := resolveDiffCommit(dir, compareRef)
	if err != nil {
		return nil, err
	}
	paths := []string{"--"}
	if options != nil {
		paths = append(paths, options.PathPatterns...)
	}
	list := func(args ...string) ([]string, error) {
		out, runErr := diffGitOutput(dir, append(args, paths...)...)
		if runErr != nil {
			return nil, runErr
		}
		return splitLinesFilterEmpty(strings.TrimSpace(out)), nil
	}

	// Keep the existing ls-files/index and WORKING semantics and map assembly.
	newFilesArgs := []string{"ls-files", "--with-tree", newOID}
	if ref == COMMIT_WORKING {
		newFilesArgs = []string{"ls-files", "--exclude-standard", "-co"}
	}
	allNewFiles, err := list(newFilesArgs...)
	if err != nil {
		return nil, err
	}
	allOldFiles, err := list("ls-files", "--with-tree", oldOID)
	if err != nil {
		return nil, err
	}
	revisions := []string{oldOID}
	if newOID != "" {
		revisions = append(revisions, newOID)
	}
	addedFiles, err := list(append([]string{"diff", "--diff-filter=A", "--name-only", "--ignore-submodules"}, revisions...)...)
	if err != nil {
		return nil, err
	}
	var untrackedFiles []string
	if ref == COMMIT_WORKING {
		untrackedFiles, err = list("ls-files", "--no-empty-directory", "--exclude-standard", "--others", "--full-name")
		if err != nil {
			return nil, err
		}
	}
	modifiedFiles, err := list(append([]string{"-c", "core.fileMode=false", "diff", "--diff-filter=M", "--name-only", "--ignore-submodules"}, revisions...)...)
	if err != nil {
		return nil, err
	}
	renameArgs := append([]string{"diff", "--diff-filter=R", "--summary", "--ignore-submodules"}, revisions...)
	renamedFilesWithSummaryGroup, err := diffGitOutput(dir, append(renameArgs, paths...)...)
	if err != nil {
		return nil, err
	}

	fileDetailsMap = make(map[string]*FileDetail, len(allNewFiles))
	for _, file := range allNewFiles {
		fileDetailsMap[file] = &FileDetail{}
	}
	for _, file := range allOldFiles {
		_, ok := fileDetailsMap[file]
		if !ok {
			// deleted files has two type of changes:
			// 1.deleted
			// 2.renamed to another file
			// an example:
			//   HEAD:   ab.txt
			// HEAD~1:   a.txt b.txt
			// results:  ab.txt: renamedFrom=a.txt,contentChanged=true
			//           a.txt:  deleted = true
			//           b.txt:  deleted = true
			fileDetailsMap[file] = &FileDetail{
				Deleted: true, // it may be renmaed also
			}
		}
	}
	// add untracked files
	for _, file := range untrackedFiles {
		fd := fileDetailsMap[file]
		if fd != nil {
			if fd.Deleted {
				// untracked file in allOld, mark as update
				fd.Deleted = false
				fd.ContentChanged = true
				continue
			}
			// untracked file in allNewFiles
			fd.IsNew = true
			continue
		}
		// neither in allNew nor allOld, mark as new
		fileDetailsMap[file] = &FileDetail{
			IsNew: true,
		}
	}

	for _, file := range addedFiles {
		fileDetailsMap[file].IsNew = true
	}
	// NOTE: this only contains un-renamed updates
	for _, file := range modifiedFiles {
		fileDetailsMap[file].ContentChanged = true
	}

	// check renames and updates
	// iterate for each 'rename ' at line begin
	// do not consider filename containing newline,"{" and "}", and space in the end or start
	parseRenames(renamedFilesWithSummaryGroup, func(newFile, oldFile, percent string) {
		d := fileDetailsMap[newFile]

		d.RenamedFrom = oldFile
		d.ContentChanged = d.ContentChanged || percent != "100%"
	})

	return
}

// diffGitOutput deliberately does not use RunCommand: revisions, pathspecs and
// the worktree directory must never be interpreted by a shell.
func diffGitOutput(dir string, args ...string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("requires dir")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}

func resolveDiffCommit(dir, ref string) (string, error) {
	if strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("invalid commit reference: option-like value")
	}
	for _, char := range ref {
		if char < 0x20 || char == 0x7f {
			return "", fmt.Errorf("invalid commit reference: control character")
		}
	}
	out, err := diffGitOutput(dir, "rev-parse", "--verify", "--quiet", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	oid := strings.TrimSpace(out)
	if len(oid) != 40 && len(oid) != 64 {
		return "", fmt.Errorf("invalid resolved commit object ID")
	}
	if _, err := hex.DecodeString(oid); err != nil {
		return "", fmt.Errorf("invalid resolved commit object ID: %w", err)
	}
	return oid, nil
}

func parseRenames(renamedFilesWithSummary string, fn func(newFile string, oldFile string, percent string)) {
	renames := splitLinesFilterEmpty(renamedFilesWithSummary)
	for _, line := range renames {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "rename ") {
			continue
		}
		s := strings.TrimSpace(line[len("rename "):])

		// parse percent
		//  rename src/{module_funder/id/submodule_funder_seabank/cl => module_funder_bke/id}/bcl/task/repair_clawback_task.go (64%)
		var percent string
		pEidx := strings.LastIndex(s, ")")
		if pEidx >= 0 {
			s = strings.TrimSpace(s[:pEidx])
			pIdx := strings.LastIndex(s, "(")
			if pIdx >= 0 {
				percent = strings.TrimSpace(s[pIdx+1:])
				s = strings.TrimSpace(s[:pIdx])
			}
		}

		bIdx := strings.Index(s, "{")
		if bIdx < 0 {
			continue
		}
		bEIdx := strings.LastIndex(s, "}")
		if bEIdx < 0 {
			continue
		}
		prefix := s[:bIdx]
		var suffix string
		if bEIdx+1 < len(s) {
			suffix = s[bEIdx+1:]
		}
		s = s[bIdx+1 : bEIdx]
		sep := " => "
		toIdx := strings.Index(s, sep)
		if toIdx < 0 {
			continue
		}
		oldPath := s[:toIdx]
		var newPath string
		if toIdx+len(sep) < len(s) {
			newPath = s[toIdx+len(sep):]
		}

		file := joinPath(prefix, newPath, suffix)
		oldFile := joinPath(prefix, oldPath, suffix)

		fn(file, oldFile, percent)
	}
}

func joinPath(p ...string) string {
	j := 0
	for i := 0; i < len(p); i++ {
		e := strings.TrimPrefix(p[i], "/")
		e = strings.TrimSuffix(e, "/")
		if e != "" {
			p[j] = e
			j++
		}
	}
	return strings.Join(p[:j], "/")
}
