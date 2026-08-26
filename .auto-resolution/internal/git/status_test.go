package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name      string
		out       string
		wantFiles []FileStatus
		wantClean bool
	}{
		{
			name:      "empty output is clean",
			out:       "## main\n",
			wantClean: true,
		},
		{
			name: "staged modification",
			out:  "## main\nM  a.request.json\n",
			wantFiles: []FileStatus{
				{Path: "a.request.json", X: 'M', Y: ' ', Staged: true},
			},
		},
		{
			name: "unstaged modification",
			out:  "## main\n M b.yaml\n",
			wantFiles: []FileStatus{
				{Path: "b.yaml", X: ' ', Y: 'M', Staged: false},
			},
		},
		{
			name: "untracked",
			out:  "## main\n?? new-dir/\n",
			wantFiles: []FileStatus{
				{Path: "new-dir/", X: '?', Y: '?', Staged: false},
			},
		},
		{
			name: "rename keeps new path",
			out:  "## feature\nR  old.json -> new.json\n",
			wantFiles: []FileStatus{
				{Path: "new.json", X: 'R', Y: ' ', Staged: true},
			},
		},
		{
			name: "mixed states",
			out:  "## dev\nMM c.json\n D d.json\nA  e.json\n",
			wantFiles: []FileStatus{
				{Path: "c.json", X: 'M', Y: 'M', Staged: true},
				{Path: "d.json", X: ' ', Y: 'D', Staged: false},
				{Path: "e.json", X: 'A', Y: ' ', Staged: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePorcelain(tt.out)
			if got.Clean != tt.wantClean {
				t.Errorf("Clean = %v, want %v", got.Clean, tt.wantClean)
			}
			if len(got.Files) != len(tt.wantFiles) {
				t.Fatalf("got %d files, want %d: %+v", len(got.Files), len(tt.wantFiles), got.Files)
			}
			for i, f := range got.Files {
				w := tt.wantFiles[i]
				if f != w {
					t.Errorf("file[%d] = %+v, want %+v", i, f, w)
				}
			}
		})
	}
}

// initRepo creates a temp git repo with an initial commit and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	return dir
}

func TestStatusRoundTripAndStageCommit(t *testing.T) {
	dir := initRepo(t)

	st, err := Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.RepoFound || st.Branch != "main" || !st.Clean {
		t.Fatalf("initial status = %+v", st)
	}

	if err := os.WriteFile(filepath.Join(dir, "env.request.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err = Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Files) != 1 || st.Files[0].Path != "env.request.json" {
		t.Fatalf("dirty status = %+v", st.Files)
	}

	if err := Stage(dir, []string{"env.request.json"}); err != nil {
		t.Fatal(err)
	}
	st, err = Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Files[0].Staged {
		t.Fatalf("expected staged entry, got %+v", st.Files[0])
	}

	if err := Commit(dir, ""); err == nil {
		t.Fatal("empty message should error")
	}
	if err := Commit(dir, "feat: add env request"); err != nil {
		t.Fatal(err)
	}
	st, err = Status(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Clean {
		t.Fatalf("repo not clean after commit: %+v", st.Files)
	}
}

func TestIsRepoFalseOutsideRepo(t *testing.T) {
	if IsRepo(t.TempDir()) {
		t.Fatal("temp dir reported as repo")
	}
}

func TestStatusOutsideRepoReportsNotFound(t *testing.T) {
	st, err := Status(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if st.RepoFound {
		t.Fatal("RepoFound should be false outside a repo")
	}
}
