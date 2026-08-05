// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package fstree

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

var (
	errWalkPerm           = errors.New("permission denied")
	errInfoNotImplemented = errors.New("not implemented")
)

// stubDirEntry is a minimal fs.DirEntry used to simulate walk callbacks.
type stubDirEntry struct {
	name  string
	isDir bool
}

func (s stubDirEntry) Name() string               { return s.name }
func (s stubDirEntry) IsDir() bool                { return s.isDir }
func (s stubDirEntry) Type() fs.FileMode          { return 0 }
func (s stubDirEntry) Info() (fs.FileInfo, error) { return nil, errInfoNotImplemented }

// --- shouldSkipDir ---

func TestShouldSkipDir_gitDir(t *testing.T) {
	if !shouldSkipDir(".git") {
		t.Error("expected .git to be skipped")
	}
}

func TestShouldSkipDir_regularDir(t *testing.T) {
	if shouldSkipDir("vendor") {
		t.Error("expected vendor to not be skipped")
	}
}

func TestShouldSkipDir_emptyName(t *testing.T) {
	if shouldSkipDir("") {
		t.Error("expected empty name to not be skipped")
	}
}

// --- fsWalk ---

func TestFsWalk_returnsFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.go"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := fsWalk(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %v", files)
	}
}

func TestFsWalk_skipsGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	files, err := fsWalk(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "main.go" {
		t.Errorf("expected only main.go, got %v", files)
	}
}

func TestFsWalk_entryErrorIsSkipped(t *testing.T) {
	orig := walkDir
	defer func() { walkDir = orig }()
	walkDir = func(root string, fn fs.WalkDirFunc) error {
		// Simulate an OS-level error on a single entry; fsWalk should skip it.
		_ = fn(filepath.Join(root, "bad"), nil, errWalkPerm)
		return nil
	}

	files, err := fsWalk("any")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files after entry error, got %v", files)
	}
}

// --- Walk ---

func TestWalk_delegatesToFsWalk(t *testing.T) {
	origWalk := walkDir
	defer func() { walkDir = origWalk }()

	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		_ = fn(filepath.Join("repo", "file.go"), stubDirEntry{name: "file.go", isDir: false}, nil)
		return nil
	}

	files, err := Walk("repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %v", files)
	}
}

// --- Filter ---

func TestFilter_includeMatchNoExclude(t *testing.T) {
	files := []string{"a.go", "b.ts"}
	got := Filter(files, []string{"*.go"}, nil)
	if len(got) != 1 || got[0] != "a.go" {
		t.Errorf("expected [a.go], got %v", got)
	}
}

func TestFilter_excludeWinsOverInclude(t *testing.T) {
	files := []string{"a.go"}
	got := Filter(files, []string{"*.go"}, []string{"a.go"})
	if len(got) != 0 {
		t.Errorf("expected exclude to win, got %v", got)
	}
}

func TestFilter_noIncludeMatch(t *testing.T) {
	files := []string{"a.ts"}
	got := Filter(files, []string{"*.go"}, nil)
	if len(got) != 0 {
		t.Errorf("expected empty when include does not match, got %v", got)
	}
}

func TestFilter_emptyIncludeExcludesAll(t *testing.T) {
	files := []string{"a.go", "b.go"}
	got := Filter(files, nil, nil)
	if len(got) != 0 {
		t.Errorf("expected all excluded when include is empty, got %v", got)
	}
}

func TestFilter_emptyFiles(t *testing.T) {
	got := Filter(nil, []string{"*.go"}, nil)
	if len(got) != 0 {
		t.Errorf("expected empty result for nil files, got %v", got)
	}
}

func TestFilter_matchAllIncludesEverything(t *testing.T) {
	files := []string{"main.go", "config/settings.json", "a/b/c/deep.txt"}
	got := Filter(files, []string{MatchAll}, nil)
	if len(got) != len(files) {
		t.Errorf("expected MatchAll to include every file, got %v", got)
	}
}

// --- matchesAny ---

func TestMatchesAny_noPatterns(t *testing.T) {
	if matchesAny("a.go", nil) {
		t.Error("expected false for no patterns")
	}
}

func TestMatchesAny_fullPathMatch(t *testing.T) {
	if !matchesAny(filepath.Join("internal", "foo.go"), []string{"**/*.go"}) {
		t.Error("expected full path to match **/*.go")
	}
}

func TestMatchesAny_baseNameMatch(t *testing.T) {
	if !matchesAny(filepath.Join("internal", "foo.go"), []string{"foo.go"}) {
		t.Error("expected base name to match foo.go")
	}
}

func TestMatchesAny_noMatch(t *testing.T) {
	if matchesAny("main.ts", []string{"*.go"}) {
		t.Error("expected no match for .ts file against *.go pattern")
	}
}

// --- globMatch ---

func TestGlobMatch_simplePatternMatch(t *testing.T) {
	if !globMatch("*.go", "main.go") {
		t.Error("expected *.go to match main.go")
	}
}

func TestGlobMatch_simplePatternNoMatch(t *testing.T) {
	if globMatch("*.go", "main.ts") {
		t.Error("expected *.go to not match main.ts")
	}
}

func TestGlobMatch_doubleStarMatch(t *testing.T) {
	if !globMatch("**/*.go", "internal/pkg/foo.go") {
		t.Error("expected **/*.go to match internal/pkg/foo.go")
	}
}

func TestGlobMatch_doubleStarNoMatch(t *testing.T) {
	if globMatch("**/*.go", "internal/pkg/foo.ts") {
		t.Error("expected **/*.go to not match foo.ts")
	}
}

func TestGlobMatch_singleStarDoesNotCrossSeparator(t *testing.T) {
	if globMatch("*.go", "a/main.go") {
		t.Error("expected *.go to not match across a path separator")
	}
}

func TestGlobMatch_questionMark(t *testing.T) {
	if !globMatch("file?.go", "fileA.go") {
		t.Error("expected file?.go to match fileA.go")
	}
	if globMatch("file?.go", "fileAB.go") {
		t.Error("expected file?.go to not match fileAB.go")
	}
}

func TestGlobMatch_braceAlternation(t *testing.T) {
	if !globMatch("*.{go,ts}", "main.go") {
		t.Error("expected *.{go,ts} to match main.go")
	}
	if !globMatch("*.{go,ts}", "main.ts") {
		t.Error("expected *.{go,ts} to match main.ts")
	}
	if globMatch("*.{go,ts}", "main.py") {
		t.Error("expected *.{go,ts} to not match main.py")
	}
}

func TestGlobMatch_invalidPatternMatchesNothing(t *testing.T) {
	if globMatch("[", "anything") {
		t.Error("expected invalid pattern to match nothing")
	}
}

// --- Flags.AddTo ---

func TestFlagsAddTo_registersAllFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	f := &Flags{}
	f.AddTo(cmd)
	for _, name := range []string{"all", "no-gitignore", "include", "exclude"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}
