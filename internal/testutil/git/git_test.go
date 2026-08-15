// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gittest "github.com/leaflockio/core-cli/internal/testutil/git"
)

func TestInitRepo_disablesCommitSigning(t *testing.T) {
	gittest.RequireGit(t)
	dir := gittest.InitRepo(t)

	cmd := exec.Command("git", "config", "commit.gpgsign")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git config commit.gpgsign: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "false" {
		t.Errorf("commit.gpgsign = %q, want %q", got, "false")
	}
}

func TestAddCommitCurrentCommit_roundTrip(t *testing.T) {
	gittest.RequireGit(t)
	dir := gittest.InitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	gittest.Add(t, dir, "a.txt")
	gittest.Commit(t, dir, "add a.txt")

	got := gittest.CurrentCommit(t, dir)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, out)
	}
	want := strings.TrimSpace(string(out))

	if got != want {
		t.Errorf("CurrentCommit = %q, want %q", got, want)
	}
	if len(got) != 40 {
		t.Errorf("CurrentCommit = %q, want a 40-character hash", got)
	}
}
