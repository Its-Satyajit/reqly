package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Worktree is one entry of `git worktree list`.
type Worktree struct {
	Path      string
	Branch    string
	IsCurrent bool
	IsBare    bool
	Detached  bool // detached HEAD worktree (no branch)
}

// Worktrees lists the worktrees of the repo containing dir. IsCurrent marks
// the worktree that contains dir itself.
func Worktrees(dir string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return parseWorktrees(string(out)), nil
}

func parseWorktrees(out string) []Worktree {
	var trees []Worktree
	var cur *Worktree
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			if cur != nil {
				trees = append(trees, *cur)
			}
			cur = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			cur.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			cur.Detached = true
		case line == "bare":
			cur.IsBare = true
		}
	}
	if cur != nil {
		trees = append(trees, *cur)
	}
	return trees
}

// AddWorktree creates a new worktree at path on a new branch named after its
// directory (git's default when no start point is given).
func AddWorktree(repoDir, path string) error {
	cmd := exec.Command("git", "worktree", "add", "--", path)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoveWorktree deletes the worktree at path.
func RemoveWorktree(repoDir, path string) error {
	cmd := exec.Command("git", "worktree", "remove", "--", path)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RecentCommit is one entry of the shortlog.
type RecentCommit struct {
	Hash    string
	Subject string
}

// RecentCommits returns up to limit commits from HEAD, newest first.
func RecentCommits(dir string, limit int) ([]RecentCommit, error) {
	if limit <= 0 || limit > 100 {
		limit = 5
	}
	cmd := exec.Command("git", "log",
		fmt.Sprintf("-%d", limit),
		"--pretty=format:%h%x09%s")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w: %s", err, strings.TrimSpace(string(out)))
	}
	var commits []RecentCommit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		commits = append(commits, RecentCommit{Hash: parts[0], Subject: parts[1]})
	}
	return commits, nil
}
