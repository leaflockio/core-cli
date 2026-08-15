// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package filecreatedyear

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	gittest "github.com/leaflockio/core-cli/internal/testutil/git"
	"github.com/leaflockio/core-cli/internal/vars"
)

func TestFileCreatedYear_fallsBackToCurrentYearWithNoCommitHistory(t *testing.T) {
	gittest.RequireGit(t)
	dir := gittest.InitRepo(t)
	// The repo needs at least one commit so `git log` on the query below
	// succeeds (empty output = GIT006), rather than failing outright
	// because the repo has no commits at all.
	if err := os.WriteFile(filepath.Join(dir, "dummy.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	gittest.Add(t, dir, "dummy.txt")
	gittest.Commit(t, dir, "init")

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
	gittest.RequireGit(t)
	dir := gittest.InitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	gittest.Add(t, dir, "main.go")
	gittest.Commit(t, dir, "add main.go")

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
