// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package git spins up real, throwaway git repositories for tests that
// need to exercise the actual git binary rather than a fake. It exists
// because internal/git has no seam another package's tests can reach —
// its own tests fake the git call directly, from inside the package.
package git

import (
	"os/exec"
	"strings"
	"testing"
)

// RequireGit skips t unless the git binary is available on PATH.
func RequireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// InitRepo creates a temp git repository with a fixed local identity and
// commit signing disabled, so tests don't depend on the host's global git
// config. Returns the repository's root directory.
func InitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	Run(t, dir, "init", "-q")
	Run(t, dir, "config", "user.email", "leaf-test@example.com")
	Run(t, dir, "config", "user.name", "leaf-test")
	Run(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

// Run runs an arbitrary git subcommand in dir, failing t on error.
func Run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// Add stages paths in dir.
func Add(t *testing.T, dir string, paths ...string) {
	t.Helper()
	Run(t, dir, append([]string{"add"}, paths...)...)
}

// Commit commits dir's staged changes with msg.
func Commit(t *testing.T, dir, msg string) {
	t.Helper()
	Run(t, dir, "commit", "-q", "-m", msg)
}

// CurrentCommit returns dir's HEAD commit hash.
func CurrentCommit(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}
