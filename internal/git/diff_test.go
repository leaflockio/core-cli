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

	"github.com/leaflockio/core-cli/internal/errs"
)

var errDiffCmdFailed = errors.New("git command failed")

// --- ResolvePRFiles ---

func TestResolvePRFiles_returnsFilesOnSuccess(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("a.go\nb.go\n"), nil
	}

	got, err := ResolvePRFiles("/repo", "origin/main", "ACM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "a.go" || got[1] != "b.go" {
		t.Errorf("ResolvePRFiles = %v", got)
	}
}

func TestResolvePRFiles_emptyBaseReturnsGIT003(t *testing.T) {
	_, err := ResolvePRFiles("/repo", "", "ACM")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.GIT003 {
		t.Errorf("expected GIT003 error, got %v", err)
	}
}

func TestResolvePRFiles_diffFailureReturnsGIT002(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errDiffCmdFailed
	}

	_, err := ResolvePRFiles("/repo", "origin/main", "ACM")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.GIT002 {
		t.Errorf("expected GIT002 error, got %v", err)
	}
}

// --- refOn ---

func TestRefOn(t *testing.T) {
	if got := refOn("origin", "main"); got != "origin/main" {
		t.Errorf("refOn = %q, want %q", got, "origin/main")
	}
}

// --- BaseRefFrom ---

func TestBaseRefFrom_composesRef(t *testing.T) {
	got, err := BaseRefFrom("origin", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "origin/main" {
		t.Errorf("BaseRefFrom = %q, want %q", got, "origin/main")
	}
}

func TestBaseRefFrom_errorsWhenRemoteEmpty(t *testing.T) {
	if _, err := BaseRefFrom("", "main"); err == nil {
		t.Error("expected error when remote is empty")
	}
}

func TestBaseRefFrom_errorsWhenBranchEmpty(t *testing.T) {
	if _, err := BaseRefFrom("origin", ""); err == nil {
		t.Error("expected error when branch is empty")
	}
}

// --- noBaseError ---

func TestNoBaseError_returnsGIT003(t *testing.T) {
	var e *errs.Error
	if !errors.As(noBaseError(), &e) || e.Code != errs.GIT003 {
		t.Error("expected GIT003 error")
	}
}

// --- remoteOf ---

func TestRemoteOf_withPrefix(t *testing.T) {
	if got := remoteOf("upstream/release"); got != "upstream" {
		t.Errorf("remoteOf = %q, want %q", got, "upstream")
	}
}

func TestRemoteOf_bareSHAReturnsEmpty(t *testing.T) {
	if got := remoteOf("abc123"); got != "" {
		t.Errorf("remoteOf = %q, want empty", got)
	}
}

// --- ResolveStaged ---

func TestResolveStaged_returnsFiles(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("a.go\nb.go\n"), nil
	}

	files, err := ResolveStaged("/repo", "ACM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 || files[0] != "a.go" || files[1] != "b.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestResolveStaged_emptyOutput(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte(""), nil
	}

	files, err := ResolveStaged("/repo", "ACM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files, got %v", files)
	}
}

func TestResolveStaged_commandErrorReturnsGIT001(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errDiffCmdFailed
	}

	_, err := ResolveStaged("/repo", "ACM")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.GIT001 {
		t.Errorf("expected GIT001 error, got %v", err)
	}
}

func TestResolveStaged_passesDiffFilter(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	var gotArgs []string
	runOutput = func(_ string, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(""), nil
	}

	_, _ = ResolveStaged("/repo", "M")
	for _, a := range gotArgs {
		if a == "--diff-filter=M" {
			return
		}
	}
	t.Errorf("expected --diff-filter=M in args, got %v", gotArgs)
}

// --- gitDiffFiles ---

func TestGitDiffFiles_buildsTripleDotRange(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	var gotArgs []string
	runOutput = func(_ string, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(""), nil
	}

	_, _ = gitDiffFiles("/repo", "origin/main", "ACM")
	for _, a := range gotArgs {
		if a == "origin/main...HEAD" {
			return
		}
	}
	t.Errorf("expected origin/main...HEAD in args, got %v", gotArgs)
}

// --- prFetchError ---

func TestPrFetchError_includesFetchSuggestionWhenRemoteKnown(t *testing.T) {
	var ce *errs.Error
	if !errors.As(prFetchError("origin/main"), &ce) || ce.Code != errs.GIT002 {
		t.Fatalf("expected GIT002 error")
	}
	if len(ce.Contexts) != 2 {
		t.Fatalf("expected 2 contexts, got %d: %+v", len(ce.Contexts), ce.Contexts)
	}
	if ce.Contexts[0].Resolution != "run: git fetch origin main" {
		t.Errorf("Resolution = %q, want %q", ce.Contexts[0].Resolution, "run: git fetch origin main")
	}
}

func TestPrFetchError_omitsFetchSuggestionForBareSHA(t *testing.T) {
	var ce *errs.Error
	if !errors.As(prFetchError("abc123"), &ce) {
		t.Fatalf("expected *errs.Error")
	}
	if len(ce.Contexts) != 1 {
		t.Errorf("expected only the generic context for a bare SHA, got %d: %+v", len(ce.Contexts), ce.Contexts)
	}
}

// --- splitLines ---

func TestSplitLines_empty(t *testing.T) {
	if got := splitLines(""); got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestSplitLines_multiple(t *testing.T) {
	got := splitLines("a.go\nb.go\nc.go")
	if len(got) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(got), got)
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
