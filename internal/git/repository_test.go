// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"errors"
	"testing"
)

var errCmdFailed = errors.New("git command failed")

// --- RepoRoot ---

func TestRepoRoot_returnsTrimmedPath(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("/repo/root\n"), nil
	}

	got, err := RepoRoot("/repo/root/sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/repo/root" {
		t.Errorf("RepoRoot = %q, want %q", got, "/repo/root")
	}
}

func TestRepoRoot_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := RepoRoot("/not/a/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- RemoteURL ---

func TestRemoteURL_returnsTrimmedURL(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("git@github.com:leaflockio/core-cli.git\n"), nil
	}

	got, err := RemoteURL("/repo", "origin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "git@github.com:leaflockio/core-cli.git" {
		t.Errorf("RemoteURL = %q", got)
	}
}

func TestRemoteURL_passesRemoteName(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	var gotArgs []string
	runOutput = func(_ string, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(""), nil
	}

	_, _ = RemoteURL("/repo", "upstream")
	for _, a := range gotArgs {
		if a == "upstream" {
			return
		}
	}
	t.Errorf("expected remote name %q in args, got %v", "upstream", gotArgs)
}

func TestRemoteURL_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := RemoteURL("/repo", "origin"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- ListRemotes ---

func TestListRemotes_returnsNames(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("origin\nupstream\n"), nil
	}

	got, err := ListRemotes("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "origin" || got[1] != "upstream" {
		t.Errorf("ListRemotes = %v, want [origin upstream]", got)
	}
}

func TestListRemotes_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := ListRemotes("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- ListFiles ---

func TestListFiles_returnsAbsolutePaths(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("a.go\nsub/b.go\n"), nil
	}

	got, err := ListFiles("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/repo/a.go", "/repo/sub/b.go"}
	if len(got) != len(want) {
		t.Fatalf("ListFiles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListFiles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListFiles_skipsBlankLines(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("a.go\n\n  \nb.go\n"), nil
	}

	got, err := ListFiles("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(got), got)
	}
}

func TestListFiles_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := ListFiles("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- CurrentBranch ---

func TestCurrentBranch_returnsTrimmedName(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("main\n"), nil
	}

	got, err := CurrentBranch("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main" {
		t.Errorf("CurrentBranch = %q, want %q", got, "main")
	}
}

func TestCurrentBranch_returnsErrorOnDetachedHead(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := CurrentBranch("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- CurrentCommit ---

func TestCurrentCommit_returnsTrimmedSHA(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("abc123\n"), nil
	}

	got, err := CurrentCommit("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("CurrentCommit = %q, want %q", got, "abc123")
	}
}

func TestCurrentCommit_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := CurrentCommit("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- IsShallow ---

func TestIsShallow_true(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("true\n"), nil
	}

	got, err := IsShallow("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected IsShallow to be true")
	}
}

func TestIsShallow_false(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("false\n"), nil
	}

	got, err := IsShallow("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected IsShallow to be false")
	}
}

func TestIsShallow_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := IsShallow("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- RemoteDefaultBranch ---

func TestRemoteDefaultBranch_stripsPrefix(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("refs/remotes/origin/main\n"), nil
	}

	got, err := RemoteDefaultBranch("/repo", "origin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main" {
		t.Errorf("RemoteDefaultBranch = %q, want %q", got, "main")
	}
}

func TestRemoteDefaultBranch_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := RemoteDefaultBranch("/repo", "origin"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- IsDirty ---

func TestIsDirty_trueWhenOutputNonEmpty(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte(" M file.go\n"), nil
	}

	got, err := IsDirty("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected IsDirty to be true")
	}
}

func TestIsDirty_falseWhenOutputEmpty(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte(""), nil
	}

	got, err := IsDirty("/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected IsDirty to be false")
	}
}

func TestIsDirty_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := IsDirty("/repo"); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- Version ---

func TestVersion_returnsTrimmedOutput(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("git version 2.42.0\n"), nil
	}

	got, err := Version()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "git version 2.42.0" {
		t.Errorf("Version = %q", got)
	}
}

func TestVersion_returnsError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := Version(); err == nil {
		t.Error("expected error, got nil")
	}
}
