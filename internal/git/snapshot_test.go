// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

var errUnexpectedArgs = errors.New("unexpected args")

func TestDetect_notInstalled(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(_ string) (string, error) { return "", errCmdFailed }

	s, errs := Detect("/anywhere")
	if s.Installed {
		t.Error("expected Installed to be false")
	}
	if s.IsRepo {
		t.Error("expected IsRepo to be false when git isn't installed")
	}
	if errs != nil {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestDetect_installedButNotRepo(t *testing.T) {
	origLookPath := lookPath
	origOut := runOutput
	defer func() { lookPath = origLookPath; runOutput = origOut }()
	lookPath = func(_ string) (string, error) { return "git", nil }
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	s, errs := Detect("/not/a/repo")
	if !s.Installed {
		t.Error("expected Installed to be true")
	}
	if s.IsRepo {
		t.Error("expected IsRepo to be false")
	}
	if s.RootDir != "" {
		t.Errorf("expected RootDir empty, got %q", s.RootDir)
	}
	if errs != nil {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestDetect_populatesRepoFields(t *testing.T) {
	origLookPath := lookPath
	origOut := runOutput
	defer func() { lookPath = origLookPath; runOutput = origOut }()
	lookPath = func(_ string) (string, error) { return "git", nil }

	runOutput = func(_ string, args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "rev-parse --show-toplevel":
			return []byte("/repo\n"), nil
		case "remote":
			return []byte("origin\n"), nil
		case "remote get-url origin":
			return []byte("git@github.com:leaflockio/core-cli.git\n"), nil
		case "symbolic-ref --short HEAD":
			return []byte("main\n"), nil
		case "rev-parse HEAD":
			return []byte("abc123\n"), nil
		case "rev-parse --is-shallow-repository":
			return []byte("false\n"), nil
		case "symbolic-ref refs/remotes/origin/HEAD":
			return []byte("refs/remotes/origin/main\n"), nil
		case "status --porcelain":
			return []byte(""), nil
		case "--version":
			return []byte("git version 2.42.0\n"), nil
		}
		return nil, fmt.Errorf("%w: %v", errUnexpectedArgs, args)
	}

	s, errs := Detect("/repo/sub")
	if errs != nil {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !s.Installed || !s.IsRepo {
		t.Fatalf("expected Installed and IsRepo true, got %+v", s)
	}
	if s.RootDir != "/repo" {
		t.Errorf("RootDir = %q, want %q", s.RootDir, "/repo")
	}
	if s.DefaultRemote != "origin" {
		t.Errorf("DefaultRemote = %q, want %q", s.DefaultRemote, "origin")
	}
	if s.RemoteURL != "git@github.com:leaflockio/core-cli.git" {
		t.Errorf("RemoteURL = %q", s.RemoteURL)
	}
	if s.Branch != "main" {
		t.Errorf("Branch = %q, want %q", s.Branch, "main")
	}
	if s.CommitSHA != "abc123" {
		t.Errorf("CommitSHA = %q, want %q", s.CommitSHA, "abc123")
	}
	if s.Shallow {
		t.Error("expected Shallow to be false")
	}
	if s.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", s.DefaultBranch, "main")
	}
	if s.Dirty {
		t.Error("expected Dirty to be false")
	}
	if s.Version != "git version 2.42.0" {
		t.Errorf("Version = %q", s.Version)
	}
}

func TestDetect_leavesFieldEmptyWhenIndividualCallFails(t *testing.T) {
	origLookPath := lookPath
	origOut := runOutput
	defer func() { lookPath = origLookPath; runOutput = origOut }()
	lookPath = func(_ string) (string, error) { return "git", nil }

	runOutput = func(_ string, args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "rev-parse --show-toplevel":
			return []byte("/repo\n"), nil
		case "remote":
			return []byte("origin\n"), nil
		case "symbolic-ref --short HEAD":
			return nil, errCmdFailed
		}
		return []byte(""), nil
	}

	s, errs := Detect("/repo")
	if !s.IsRepo {
		t.Fatal("expected IsRepo true")
	}
	if s.Branch != "" {
		t.Errorf("Branch = %q, want empty on detached HEAD", s.Branch)
	}
	if len(errs) == 0 {
		t.Error("expected at least one collected error")
	}
}

func TestDetect_noRemotesLeavesDependentFieldsEmpty(t *testing.T) {
	origLookPath := lookPath
	origOut := runOutput
	defer func() { lookPath = origLookPath; runOutput = origOut }()
	lookPath = func(_ string) (string, error) { return "git", nil }

	runOutput = func(_ string, args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "rev-parse --show-toplevel":
			return []byte("/repo\n"), nil
		case "remote":
			return []byte(""), nil
		}
		return []byte(""), nil
	}

	s, _ := Detect("/repo")
	if s.DefaultRemote != "" {
		t.Errorf("DefaultRemote = %q, want empty", s.DefaultRemote)
	}
	if s.RemoteURL != "" {
		t.Errorf("RemoteURL = %q, want empty when DefaultRemote is empty", s.RemoteURL)
	}
	if s.DefaultBranch != "" {
		t.Errorf("DefaultBranch = %q, want empty when DefaultRemote is empty", s.DefaultBranch)
	}
}

// --- Snapshot.BaseRef ---

func TestSnapshotBaseRef_composesFromResolvedFields(t *testing.T) {
	s := &Snapshot{DefaultRemote: "origin", DefaultBranch: "main"}
	got, err := s.BaseRef()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "origin/main" {
		t.Errorf("BaseRef = %q, want %q", got, "origin/main")
	}
}

func TestSnapshotBaseRef_errorsWhenUnresolved(t *testing.T) {
	s := &Snapshot{}
	if _, err := s.BaseRef(); err == nil {
		t.Error("expected error when DefaultRemote/DefaultBranch are empty")
	}
}

// --- resolveDefaultRemote ---

func TestResolveDefaultRemote_singleRemote(t *testing.T) {
	if got := resolveDefaultRemote([]string{"upstream"}); got != "upstream" {
		t.Errorf("resolveDefaultRemote = %q, want %q", got, "upstream")
	}
}

func TestResolveDefaultRemote_multipleWithOrigin(t *testing.T) {
	if got := resolveDefaultRemote([]string{"upstream", "origin"}); got != "origin" {
		t.Errorf("resolveDefaultRemote = %q, want %q", got, "origin")
	}
}

func TestResolveDefaultRemote_multipleWithoutOrigin(t *testing.T) {
	if got := resolveDefaultRemote([]string{"upstream", "fork"}); got != "" {
		t.Errorf("resolveDefaultRemote = %q, want empty", got)
	}
}

func TestResolveDefaultRemote_none(t *testing.T) {
	if got := resolveDefaultRemote(nil); got != "" {
		t.Errorf("resolveDefaultRemote = %q, want empty", got)
	}
}
