// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/errs"
)

// --- FileCreatedDate ---

func TestFileCreatedDate_parsesEarliestCommitDate(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	var gotArgs []string
	runOutput = func(_ string, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte("2020-01-15T10:00:00-05:00\n"), nil
	}

	got, err := FileCreatedDate("/repo", "path/to/file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, _ := time.Parse(time.RFC3339, "2020-01-15T10:00:00-05:00")
	if !got.Equal(want) {
		t.Errorf("FileCreatedDate = %v, want %v", got, want)
	}

	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "--reverse") {
		t.Errorf("expected --reverse in args %v", gotArgs)
	}
	if !strings.Contains(joined, "--follow") {
		t.Errorf("expected --follow in args %v", gotArgs)
	}
	if !strings.HasSuffix(joined, "-- path/to/file.go") {
		t.Errorf("expected path as final arg after --, got %v", gotArgs)
	}
}

// --- FileModifiedDate ---

func TestFileModifiedDate_parsesMostRecentCommitDate(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	var gotArgs []string
	runOutput = func(_ string, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte("2024-06-01T08:30:00Z\n"), nil
	}

	got, err := FileModifiedDate("/repo", "path/to/file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, _ := time.Parse(time.RFC3339, "2024-06-01T08:30:00Z")
	if !got.Equal(want) {
		t.Errorf("FileModifiedDate = %v, want %v", got, want)
	}

	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "-1") {
		t.Errorf("expected -1 in args %v", gotArgs)
	}
	if strings.Contains(joined, "--reverse") {
		t.Errorf("expected no --reverse in args %v", gotArgs)
	}
}

// --- fileLogDate error paths (shared by both) ---

func TestFileLogDate_propagatesCommandError(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return nil, errCmdFailed
	}

	if _, err := FileCreatedDate("/repo", "file.go"); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFileLogDate_noCommitHistory(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte(""), nil
	}

	_, err := FileCreatedDate("/repo", "file.go")
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.GIT006 {
		t.Errorf("expected GIT006 error, got %v", err)
	}
}

func TestFileLogDate_malformedDate(t *testing.T) {
	orig := runOutput
	defer func() { runOutput = orig }()
	runOutput = func(_ string, _ ...string) ([]byte, error) {
		return []byte("not-a-date\n"), nil
	}

	if _, err := FileCreatedDate("/repo", "file.go"); err == nil {
		t.Error("expected a date-parsing error, got nil")
	}
}
