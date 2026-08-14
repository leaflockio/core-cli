// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package filecreatedyear

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/vars"
)

// requireGit skips t if the git binary isn't available — there's no
// mockable seam for internal/git from outside that package, so these
// tests exercise the real command.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// runGit runs a git command in dir with a fixed local author/committer
// identity, so the test doesn't depend on any ambient git config.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=leaf-test", "GIT_AUTHOR_EMAIL=leaf-test@example.com",
		"GIT_COMMITTER_NAME=leaf-test", "GIT_COMMITTER_EMAIL=leaf-test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestFileCreatedYear_fallsBackToCurrentYearWithNoCommitHistory(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init")
	// The repo needs at least one commit so `git log` on the query below
	// succeeds (empty output = GIT006), rather than failing outright
	// because the repo has no commits at all.
	if err := os.WriteFile(filepath.Join(dir, "dummy.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	runGit(t, dir, "add", "dummy.txt")
	runGit(t, dir, "commit", "-m", "init")

	if err := os.WriteFile(filepath.Join(dir, "untracked.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	if err := Set(v, dir, "untracked.go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := v.Resolve("{FILE_CREATED_YEAR}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := strconv.Itoa(time.Now().Year())
	if got != want {
		t.Errorf("Resolve = %q, want %q (current-year fallback)", got, want)
	}
}

func TestFileCreatedYear_registersAndResolvesFromRealCommit(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	runGit(t, dir, "add", "main.go")
	runGit(t, dir, "commit", "-m", "add main.go")

	if FileCreatedYear == nil {
		t.Fatal("FileCreatedYear is nil — registration failed")
	}
	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	if err := Set(v, dir, "main.go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := v.Resolve("{FILE_CREATED_YEAR}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := strconv.Itoa(time.Now().Year())
	if got != want {
		t.Errorf("Resolve = %q, want %q (just committed)", got, want)
	}
}
