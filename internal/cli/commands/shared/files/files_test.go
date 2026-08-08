// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package files

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertCode(t *testing.T, err error, want errs.Code) {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != want {
		t.Errorf("expected %s, got %v", want, err)
	}
}

// --- Validate ---

func TestValidate_argsWithStagedIsSRC001(t *testing.T) {
	f := &Files{Staged: true}
	assertCode(t, f.Validate([]string{"a.go"}, true), errs.SRC001)
}

func TestValidate_argsWithPRIsSRC001(t *testing.T) {
	f := &Files{PR: true}
	assertCode(t, f.Validate([]string{"a.go"}, true), errs.SRC001)
}

func TestValidate_noneGivenIsSRC002(t *testing.T) {
	f := &Files{}
	assertCode(t, f.Validate(nil, true), errs.SRC002)
}

func TestValidate_stagedWithoutGitIsGIT001(t *testing.T) {
	f := &Files{Staged: true}
	assertCode(t, f.Validate(nil, false), errs.GIT001)
}

func TestValidate_prWithoutGitIsGIT001(t *testing.T) {
	f := &Files{PR: true}
	assertCode(t, f.Validate(nil, false), errs.GIT001)
}

func TestValidate_argsAloneIsValid(t *testing.T) {
	f := &Files{}
	if err := f.Validate([]string{"a.go"}, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_stagedWithGitIsValid(t *testing.T) {
	f := &Files{Staged: true}
	if err := f.Validate(nil, true); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Resolve: NoGitignore (pure filesystem, no git subprocess needed) ---

func TestResolve_noGitignoreReturnsEverythingByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	got, _, err := (&Files{NoGitignore: true}).Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 files, got %v", got)
	}
}

func TestResolve_argNarrowsToMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	got, _, err := (&Files{NoGitignore: true}).Resolve(dir, []string{"*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

func TestResolve_dotArgTranslatesToMatchAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "sub", "b.go"))

	got, _, err := (&Files{NoGitignore: true}).Resolve(dir, []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 files, got %v", got)
	}
}

func TestResolve_withIncludeNarrows(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	f := (&Files{NoGitignore: true}).WithInclude([]string{"*.go"})
	got, _, err := f.Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

func TestResolve_withExcludeNarrows(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "a_test.go"))

	f := (&Files{NoGitignore: true}).WithExclude([]string{"*_test.go"})
	got, _, err := f.Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

func TestWithInclude_accumulatesMultipleSources(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	f := (&Files{NoGitignore: true}).WithInclude([]string{"*.go"}, []string{"*.md"})
	got, _, err := f.Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 files, got %v", got)
	}
}

// TestResolve_argsAndIncludeNarrowIndependently is a regression test for the
// OR-vs-AND bug fixed this session: "." (matches everything) combined with
// --include must still narrow, not be swallowed by the broader arg pattern.
func TestResolve_argsAndIncludeNarrowIndependently(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	f := (&Files{NoGitignore: true}).WithInclude([]string{"*.go"})
	got, _, err := f.Resolve(dir, []string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

// --- Resolve: Stats ---

func TestResolve_statsReflectEachStage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.go"))
	writeFile(t, filepath.Join(dir, "c.md"))
	writeFile(t, filepath.Join(dir, "vendor", "d.go"))

	f := (&Files{NoGitignore: true}).
		WithInclude([]string{"*.go"}).
		WithExclude([]string{"vendor/**"})
	got, stats, err := f.Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Stats{Discovered: 4, Query: 4, Included: 3, Excluded: 1, Final: 2}
	if stats != want {
		t.Errorf("expected %+v, got %+v", want, stats)
	}
	if len(got) != stats.Final {
		t.Errorf("expected len(got) to match stats.Final (%d), got %d files", stats.Final, len(got))
	}
}

func TestResolve_statsQueryNarrowsFromArgs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.go"))
	writeFile(t, filepath.Join(dir, "c.md"))

	_, stats, err := (&Files{NoGitignore: true}).Resolve(dir, []string{"*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Stats{Discovered: 3, Query: 2, Included: 2, Excluded: 0, Final: 2}
	if stats != want {
		t.Errorf("expected %+v, got %+v", want, stats)
	}
}

func TestResolve_statsQueryEqualsDiscoveredWithNoArgs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.md"))

	_, stats, err := (&Files{NoGitignore: true}).Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Query != stats.Discovered {
		t.Errorf("expected Query (%d) to equal Discovered (%d) with no args", stats.Query, stats.Discovered)
	}
}

// --- Resolve: git-routed sources ---

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	run("config", "commit.gpgsign", "false") // isolate from the host's global signing config
	return dir
}

func gitAdd(t *testing.T, dir string, paths ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"add"}, paths...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "-q", "-m", msg)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

func currentCommit(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestResolve_defaultRoutesToGitListFiles(t *testing.T) {
	dir := initGitRepo(t)
	writeFile(t, filepath.Join(dir, "a.go"))
	gitAdd(t, dir, "a.go")

	got, _, err := (&Files{}).Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

func TestResolve_stagedRoutesToGitResolveStaged(t *testing.T) {
	dir := initGitRepo(t)
	writeFile(t, filepath.Join(dir, "a.go"))
	writeFile(t, filepath.Join(dir, "b.go"))
	gitAdd(t, dir, "a.go")

	got, _, err := (&Files{Staged: true}).Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected only the staged file [a.go], got %v", got)
	}
}

func TestResolve_prRoutesToGitResolvePRFiles(t *testing.T) {
	dir := initGitRepo(t)
	writeFile(t, filepath.Join(dir, "a.go"))
	gitAdd(t, dir, "a.go")
	gitCommit(t, dir, "initial")
	base := currentCommit(t, dir)

	writeFile(t, filepath.Join(dir, "b.go"))
	gitAdd(t, dir, "b.go")
	gitCommit(t, dir, "second")

	got, _, err := (&Files{PR: true, Base: base}).Resolve(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "b.go" {
		t.Errorf("expected only the changed file [b.go], got %v", got)
	}
}

func TestResolve_errorPropagatesFromDiscoverWithZeroStats(t *testing.T) {
	dir := t.TempDir() // not a git repo, so the default git.ListFiles route fails

	got, stats, err := (&Files{}).Resolve(dir, nil)
	if err == nil {
		t.Fatal("expected an error when discover fails")
	}
	if got != nil {
		t.Errorf("expected nil files on error, got %v", got)
	}
	if stats != (Stats{}) {
		t.Errorf("expected zero-value Stats on error, got %+v", stats)
	}
}
