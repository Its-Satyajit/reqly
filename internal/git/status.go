package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// FileStatus is one entry of `git status --porcelain`.
type FileStatus struct {
	// Path is the repo-relative file path (for renames, the new path).
	Path string
	// X is the index (staged) status code, Y the worktree code.
	X, Y rune
	// Staged reports whether the change has index-side changes.
	Staged bool
}

// StatusResult is the outcome of a repository status check.
type StatusResult struct {
	Branch    string
	Head      string // HEAD ref name when detached, empty otherwise
	Files     []FileStatus
	Clean     bool
	RepoFound bool
}

// IsRepo reports whether dir is inside a git work tree.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// CurrentBranch returns the checked-out branch of the repo containing dir.
// Works on unborn branches (fresh init, no commits) too.
func CurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", nil // detached HEAD or bare repo
	}
	return strings.TrimSpace(string(out)), nil
}

// Status collects branch and per-file status for the repo containing dir.
func Status(dir string) (*StatusResult, error) {
	if !IsRepo(dir) {
		return &StatusResult{RepoFound: false}, nil
	}
	branch, err := CurrentBranch(dir)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "status", "--porcelain=v1", "-b")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	res := parsePorcelain(string(out))
	res.Branch = branch
	res.RepoFound = true
	return res, nil
}

// Stage stages the given repo-relative paths (like `git add --`).
func Stage(dir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Commit records all staged changes with the given subject-only message.
func Commit(dir, message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message must not be empty")
	}
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// parsePorcelain parses `git status --porcelain=v1 -b` output. The leading
// `## <branch>` line is consumed here only for cleanliness detection; callers
// overwrite Branch from rev-parse.
func parsePorcelain(out string) *StatusResult {
	res := &StatusResult{}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			continue
		}
		if len(line) < 4 {
			continue
		}
		x, y := rune(line[0]), rune(line[1])
		path := line[3:]
		// Renames/copies render as "R  old -> new"; keep the new path.
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		res.Files = append(res.Files, FileStatus{
			Path:   path,
			X:      x,
			Y:      y,
			Staged: x != ' ' && x != '?',
		})
	}
	res.Clean = len(res.Files) == 0
	return res
}

// Unstage removes paths from the index (like `git restore --staged --`).
func Unstage(dir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"restore", "--staged", "--"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git restore: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
