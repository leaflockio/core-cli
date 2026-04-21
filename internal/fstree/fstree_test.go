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
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var (
	errLsFiles            = errors.New("ls-files failed")
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

// --- gitWalk ---

func TestGitWalk_returnsFiles(t *testing.T) {
	orig := gitLsFilesCmd
	defer func() { gitLsFilesCmd = orig }()
	gitLsFilesCmd = func(_ string) ([]byte, error) {
		return []byte("main.go\ninternal/foo.go\n"), nil
	}

	files, err := gitWalk("repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %v", files)
	}
}

func TestGitWalk_commandError(t *testing.T) {
	orig := gitLsFilesCmd
	defer func() { gitLsFilesCmd = orig }()
	gitLsFilesCmd = func(_ string) ([]byte, error) {
		return nil, errLsFiles
	}

	_, err := gitWalk("repo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- parseGitLines ---

func TestParseGitLines_basic(t *testing.T) {
	got := parseGitLines([]byte("main.go\ninternal/foo.go\n"), "repo")
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(got), got)
	}
}

func TestParseGitLines_empty(t *testing.T) {
	got := parseGitLines([]byte(""), "repo")
	if len(got) != 0 {
		t.Errorf("expected no files, got %v", got)
	}
}

func TestParseGitLines_blankLinesSkipped(t *testing.T) {
	got := parseGitLines([]byte("a.go\n\n  \nb.go\n"), "repo")
	if len(got) != 2 {
		t.Errorf("expected 2 files, got %v", got)
	}
}

func TestParseGitLines_joinsRoot(t *testing.T) {
	root := t.TempDir()
	got := parseGitLines([]byte("pkg/foo.go\n"), root)
	want := filepath.Join(root, "pkg", "foo.go")
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected %q, got %v", want, got)
	}
}

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

func TestWalk_gitignoreSucceeds(t *testing.T) {
	orig := gitLsFilesCmd
	defer func() { gitLsFilesCmd = orig }()
	gitLsFilesCmd = func(_ string) ([]byte, error) {
		return []byte("tracked.go\n"), nil
	}

	files, err := Walk("repo", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from gitWalk, got %v", files)
	}
}

func TestWalk_gitignoreFallsBackToFsWalk(t *testing.T) {
	origGit := gitLsFilesCmd
	origWalk := walkDir
	defer func() { gitLsFilesCmd = origGit; walkDir = origWalk }()

	gitLsFilesCmd = func(_ string) ([]byte, error) {
		return nil, errLsFiles // git unavailable
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		_ = fn(filepath.Join("repo", "file.go"), stubDirEntry{name: "file.go", isDir: false}, nil)
		return nil
	}

	files, err := Walk("repo", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from fsWalk fallback, got %v", files)
	}
}

func TestWalk_noGitignoreUsesFsWalk(t *testing.T) {
	origGit := gitLsFilesCmd
	origWalk := walkDir
	defer func() { gitLsFilesCmd = origGit; walkDir = origWalk }()

	called := false
	gitLsFilesCmd = func(_ string) ([]byte, error) {
		called = true
		return nil, errLsFiles
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		_ = fn(filepath.Join("repo", "file.go"), stubDirEntry{name: "file.go", isDir: false}, nil)
		return nil
	}

	files, err := Walk("repo", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected gitLsFilesCmd to not be called when useGitignore=false")
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from fsWalk, got %v", files)
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

// --- buildDoublestarPattern ---

func TestBuildDoublestarPattern_isAnchored(t *testing.T) {
	p := buildDoublestarPattern("*.go")
	if !strings.HasPrefix(p, "^") || !strings.HasSuffix(p, "$") {
		t.Errorf("expected anchored pattern, got %q", p)
	}
}

func TestBuildDoublestarPattern_doubleStarBecomesAny(t *testing.T) {
	p := buildDoublestarPattern("**/foo.go")
	if !strings.Contains(p, reAny) {
		t.Errorf("expected %q for **, got %q", reAny, p)
	}
}

func TestBuildDoublestarPattern_singleStarBecomesNonSep(t *testing.T) {
	p := buildDoublestarPattern("*.go")
	if !strings.Contains(p, reNonSep) {
		t.Errorf("expected %q for *, got %q", reNonSep, p)
	}
}

func TestBuildDoublestarPattern_questionBecomesNonSepSingle(t *testing.T) {
	p := buildDoublestarPattern("file?.go")
	if !strings.Contains(p, reNonSepSingle) {
		t.Errorf("expected %q for ?, got %q", reNonSepSingle, p)
	}
}

func TestBuildDoublestarPattern_dotEscaped(t *testing.T) {
	p := buildDoublestarPattern("file.go")
	if !strings.Contains(p, reLiteralDot) {
		t.Errorf("expected %q for literal dot, got %q", reLiteralDot, p)
	}
}

// --- doublestarRegexp ---

func TestDoublestarRegexp_doubleStarWithTrailingSlash(t *testing.T) {
	re := doublestarRegexp("**/foo.go")
	if !re.MatchString("internal/pkg/foo.go") {
		t.Error("expected **/foo.go to match internal/pkg/foo.go")
	}
}

func TestDoublestarRegexp_doubleStarWithoutTrailingSlash(t *testing.T) {
	re := doublestarRegexp("**foo.go")
	if !re.MatchString("foo.go") {
		t.Error("expected **foo.go to match foo.go")
	}
}

func TestDoublestarRegexp_doubleStarAtEnd(t *testing.T) {
	re := doublestarRegexp("vendor/**")
	if !re.MatchString("vendor/some/pkg/file.go") {
		t.Error("expected vendor/** to match vendor/some/pkg/file.go")
	}
}

func TestDoublestarRegexp_singleStar(t *testing.T) {
	re := doublestarRegexp("*.go")
	if !re.MatchString("main.go") {
		t.Error("expected *.go to match main.go")
	}
	if re.MatchString("a/main.go") {
		t.Error("expected *.go to not match across path separators")
	}
}

func TestDoublestarRegexp_questionMark(t *testing.T) {
	re := doublestarRegexp("file?.go")
	if !re.MatchString("fileA.go") {
		t.Error("expected file?.go to match fileA.go")
	}
	if re.MatchString("fileAB.go") {
		t.Error("expected file?.go to not match fileAB.go")
	}
}

func TestDoublestarRegexp_dot(t *testing.T) {
	re := doublestarRegexp("file.go")
	if !re.MatchString("file.go") {
		t.Error("expected file.go to match file.go")
	}
	if re.MatchString("fileXgo") {
		t.Error("expected dot to be treated as literal, not wildcard")
	}
}

func TestDoublestarRegexp_literal(t *testing.T) {
	re := doublestarRegexp("foobar")
	if !re.MatchString("foobar") {
		t.Error("expected foobar to match foobar")
	}
	if re.MatchString("foobaz") {
		t.Error("expected foobar to not match foobaz")
	}
}

func TestDoublestarRegexp_invalidPatternMatchesNothing(t *testing.T) {
	// "(**" produces an unclosed group in the regexp, triggering the fallback.
	re := doublestarRegexp("(**")
	if re.MatchString("anything") {
		t.Error("expected invalid pattern to match nothing")
	}
}

// --- gitLsFilesCmd real implementation ---

func TestGitLsFilesCmd_realImpl(t *testing.T) {
	// Exercise the real function body; working dir is inside a git repo so the
	// command succeeds. Any error is treated as git being unavailable.
	out, err := gitLsFilesCmd(".")
	if err != nil {
		t.Skipf("git ls-files not available: %v", err)
	}
	_ = out
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
