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

	"github.com/leaflock/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

var (
	errCommandFailed = errors.New("git command failed")
	errFetchFailed   = errors.New("fetch failed")
	errDiffFailed    = errors.New("diff failed")
)

// --- splitLines ---

func TestSplitLines_empty(t *testing.T) {
	if got := splitLines(""); got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestSplitLines_single(t *testing.T) {
	got := splitLines("file.go")
	if len(got) != 1 || got[0] != "file.go" {
		t.Errorf("expected [file.go], got %v", got)
	}
}

func TestSplitLines_multiple(t *testing.T) {
	got := splitLines("a.go\nb.go\nc.go")
	if len(got) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(got), got)
	}
}

func TestSplitLines_trailingNewline(t *testing.T) {
	got := splitLines("a.go\nb.go\n")
	if len(got) != 2 {
		t.Errorf("expected 2 lines after trailing newline, got %d: %v", len(got), got)
	}
}

func TestSplitLines_blankLinesSkipped(t *testing.T) {
	got := splitLines("a.go\n\nb.go\n\n")
	if len(got) != 2 {
		t.Errorf("expected blank lines to be skipped, got %d: %v", len(got), got)
	}
}

func TestSplitLines_whitespaceOnlyLine(t *testing.T) {
	got := splitLines("a.go\n   \nb.go")
	if len(got) != 2 {
		t.Errorf("expected whitespace-only line to be skipped, got %d: %v", len(got), got)
	}
}

// --- ResolveStaged ---

func TestResolveStaged_returnsFiles(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	cmdOutput = func(_ ...string) ([]byte, error) {
		return []byte("a.go\nb.go\n"), nil
	}

	files, err := ResolveStaged("ACM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 || files[0] != "a.go" || files[1] != "b.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestResolveStaged_emptyOutput(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	cmdOutput = func(_ ...string) ([]byte, error) {
		return []byte(""), nil
	}

	files, err := ResolveStaged("ACM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files, got %v", files)
	}
}

func TestResolveStaged_commandError(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	cmdOutput = func(_ ...string) ([]byte, error) {
		return nil, errCommandFailed
	}

	_, err := ResolveStaged("ACM")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ce *errs.Error
	if !errors.As(err, &ce) || ce.Code != errs.LIC005 {
		t.Errorf("expected LIC005 error, got %v", err)
	}
}

func TestResolveStaged_passesDiffFilter(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	var gotArgs []string
	cmdOutput = func(args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(""), nil
	}

	_, _ = ResolveStaged("M")
	for _, a := range gotArgs {
		if a == "--diff-filter=M" {
			return
		}
	}
	t.Errorf("expected --diff-filter=M in args, got %v", gotArgs)
}

// --- ResolvePRFiles ---

func TestResolvePRFiles_diffSucceeds(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	cmdOutput = func(_ ...string) ([]byte, error) {
		return []byte("x.go\ny.go\n"), nil
	}

	files, err := ResolvePRFiles("origin/main", "ACM", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %v", files)
	}
}

func TestResolvePRFiles_noAutoFetch(t *testing.T) {
	orig := cmdOutput
	defer func() { cmdOutput = orig }()
	cmdOutput = func(_ ...string) ([]byte, error) {
		return nil, errCommandFailed
	}

	_, err := ResolvePRFiles("origin/main", "ACM", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ce *errs.Error
	if !errors.As(err, &ce) || ce.Code != errs.LIC006 {
		t.Errorf("expected LIC006 error, got %v", err)
	}
}

func TestResolvePRFiles_fetchSucceedsThenDiff(t *testing.T) {
	origOut := cmdOutput
	origRun := cmdRun
	defer func() { cmdOutput = origOut; cmdRun = origRun }()

	calls := 0
	cmdOutput = func(_ ...string) ([]byte, error) {
		calls++
		if calls == 1 {
			return nil, errCommandFailed // first diff fails
		}
		return []byte("changed.go\n"), nil // retry succeeds
	}
	cmdRun = func(_ ...string) error {
		return nil // fetch succeeds
	}

	files, err := ResolvePRFiles("origin/main", "ACM", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "changed.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestResolvePRFiles_fetchFailsFallbackToMaster(t *testing.T) {
	origOut := cmdOutput
	origRun := cmdRun
	defer func() { cmdOutput = origOut; cmdRun = origRun }()

	cmdRun = func(_ ...string) error {
		return errFetchFailed // fetch of main fails
	}
	calls := 0
	cmdOutput = func(_ ...string) ([]byte, error) {
		calls++
		if calls == 1 {
			return nil, errDiffFailed // initial diff of main fails
		}
		return []byte("fallback.go\n"), nil // diff of master succeeds
	}

	files, err := ResolvePRFiles(defaultRemote+"/"+defaultBranch, "ACM", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "fallback.go" {
		t.Errorf("expected fallback.go from master, got %v", files)
	}
}

func TestResolvePRFiles_allPathsFail(t *testing.T) {
	origOut := cmdOutput
	origRun := cmdRun
	defer func() { cmdOutput = origOut; cmdRun = origRun }()

	cmdRun = func(_ ...string) error { return errFetchFailed }
	cmdOutput = func(_ ...string) ([]byte, error) { return nil, errDiffFailed }

	_, err := ResolvePRFiles(defaultRemote+"/"+defaultBranch, "ACM", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ce *errs.Error
	if !errors.As(err, &ce) || ce.Code != errs.LIC006 {
		t.Errorf("expected LIC006 error, got %v", err)
	}
}

func TestResolvePRFiles_nonMainBaseFetchFails(t *testing.T) {
	origOut := cmdOutput
	origRun := cmdRun
	defer func() { cmdOutput = origOut; cmdRun = origRun }()

	cmdRun = func(_ ...string) error { return errFetchFailed }
	cmdOutput = func(_ ...string) ([]byte, error) { return nil, errDiffFailed }

	_, err := ResolvePRFiles("origin/feature-x", "ACM", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ce *errs.Error
	if !errors.As(err, &ce) || ce.Code != errs.LIC006 {
		t.Errorf("expected LIC006 error, got %v", err)
	}
}

// --- Flags.AddTo ---

func TestFlagsAddTo_registersAllFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var f Flags
	f.AddTo(cmd)

	for _, name := range []string{"staged", "pr", "base", "diff-filter"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}

func TestFlagsAddTo_diffFilterDefault(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var f Flags
	f.AddTo(cmd)

	fl := cmd.Flags().Lookup("diff-filter")
	if fl == nil {
		t.Fatal("diff-filter flag not registered")
		return
	}
	if fl.DefValue != "ACM" {
		t.Errorf("expected default ACM, got %s", fl.DefValue)
	}
}
