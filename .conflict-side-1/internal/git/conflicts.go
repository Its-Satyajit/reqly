package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ConflictFile is a file with unresolved merge conflicts (unmerged index
// states such as UU, AA, DU...).
type ConflictFile struct {
	Path string
	Code string // raw two-letter porcelain code, e.g. "UU"
}

// Conflicts returns the unmerged files of the repo containing dir.
func Conflicts(dir string) ([]ConflictFile, error) {
	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w: %s", err, strings.TrimSpace(string(out)))
	}
	var conflicts []ConflictFile
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		code := line[:2]
		if !isUnmerged(code) {
			continue
		}
		path := line[3:]
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		conflicts = append(conflicts, ConflictFile{Path: path, Code: code})
	}
	return conflicts, nil
}

func isUnmerged(code string) bool {
	switch code {
	case "DD", "AU", "UD", "UA", "DU", "AA", "UU":
		return true
	}
	return false
}

// MergeAbort aborts the in-progress merge and restores the pre-merge state.
func MergeAbort(dir string) error {
	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge --abort: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ResolveSide selects one side of a conflicted file and stages the result.
// side must be "ours" or "theirs".
func ResolveSide(dir, path, side string) error {
	if side != "ours" && side != "theirs" {
		return fmt.Errorf("invalid resolve side %q: want ours or theirs", side)
	}
	checkout := exec.Command("git", "checkout", "--"+side, "--", path)
	checkout.Dir = dir
	if out, err := checkout.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout --%s: %w: %s", side, err, strings.TrimSpace(string(out)))
	}
	add := exec.Command("git", "add", "--", path)
	add.Dir = dir
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
