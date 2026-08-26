package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func commitFile(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "--", name}, {"commit", "-m", msg}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
}

// conflictRepo builds a repo with base.txt merged divergently on main/feature
// such that `git merge feature` on main conflicts.
func conflictRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	gitCmd("init", "-b", "main")
	gitCmd("config", "user.email", "t@t")
	gitCmd("config", "user.name", "t")
	commitFile(t, dir, "shared.txt", "base\n", "base")

	gitCmd("checkout", "-b", "feature")
	commitFile(t, dir, "shared.txt", "feature side\n", "feature edit")

	gitCmd("checkout", "main")
	commitFile(t, dir, "shared.txt", "main side\n", "main edit")

	merge := exec.Command("git", "merge", "feature")
	merge.Dir = dir
	_ = merge.Run() // expected to fail with conflicts

	conflicts, err := Conflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) == 0 {
		t.Fatal("expected a conflicted merge in fixture")
	}
	return dir
}

func TestConflictsAndResolveSides(t *testing.T) {
	dir := conflictRepo(t)

	conflicts, err := Conflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 || conflicts[0].Path != "shared.txt" || conflicts[0].Code != "UU" {
		t.Fatalf("conflicts = %+v", conflicts)
	}

	if err := ResolveSide(dir, "shared.txt", "bogus"); err == nil {
		t.Fatal("bogus side should error")
	}
	if err := ResolveSide(dir, "shared.txt", "theirs"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "shared.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "feature side") {
		t.Fatalf("theirs content not taken: %q", got)
	}

	remaining, err := Conflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Fatalf("conflicts remain after resolve: %+v", remaining)
	}
}

func TestMergeAbortRestoresPreMergeState(t *testing.T) {
	dir := conflictRepo(t)

	if err := MergeAbort(dir); err != nil {
		t.Fatal(err)
	}
	conflicts, err := Conflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("conflicts after abort: %+v", conflicts)
	}
	got, err := os.ReadFile(filepath.Join(dir, "shared.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "main side\n" {
		t.Fatalf("abort did not restore main content: %q", got)
	}
}

func TestWorktreesListAddRemove(t *testing.T) {
	dir := initRepo(t)

	trees, err := Worktrees(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(trees) != 1 || trees[0].Branch != "main" {
		t.Fatalf("initial worktrees = %+v", trees)
	}

	wtPath := filepath.Join(dir, "..", filepath.Base(dir)+"-wt")
	if err := AddWorktree(dir, wtPath); err != nil {
		t.Fatal(err)
	}
	trees, err = Worktrees(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tree := range trees {
		if filepath.Clean(tree.Path) == filepath.Clean(wtPath) {
			found = true
			if tree.IsCurrent {
				t.Error("new worktree wrongly flagged as current")
			}
			if tree.Branch == "" && !tree.Detached {
				t.Errorf("worktree without branch or detached marker: %+v", tree)
			}
		}
	}
	if !found {
		t.Fatalf("added worktree missing from list: %+v", trees)
	}

	if err := RemoveWorktree(dir, wtPath); err != nil {
		t.Fatal(err)
	}
	trees, _ = Worktrees(dir)
	if len(trees) != 1 {
		t.Fatalf("worktrees after remove = %+v", trees)
	}
}

func TestRecentCommits(t *testing.T) {
	dir := initRepo(t)

	commits, err := RecentCommits(dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "init" {
		t.Fatalf("commits = %+v", commits)
	}
	if commits[0].Hash == "" || len(commits[0].Hash) < 7 {
		t.Fatalf("unexpected hash %q", commits[0].Hash)
	}
}
