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
	"sort"
	"testing"
)

var errPermDenied = errors.New("permission denied")

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
}

func sorted(files []string) []string {
	out := append([]string(nil), files...)
	sort.Strings(out)
	return out
}

// --- Resolve: literal file / directory ---

// chdirTo switches the process cwd to dir for the duration of the test.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// wantCanonical returns the absolute, symlink-resolved form of path — the
// same ground truth Resolve itself computes — so tests are robust on
// systems where t.TempDir() sits behind a symlink (e.g. macOS's /tmp).
func wantCanonical(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func TestResolve_literalFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	mustWriteFile(t, f)
	want := wantCanonical(t, f)

	got, err := Resolve([]string{f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

func TestResolve_literalDirectory(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	b := filepath.Join(dir, "sub", "b.go")
	mustWriteFile(t, a)
	mustWriteFile(t, b)

	got, err := Resolve([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sorted([]string{wantCanonical(t, a), wantCanonical(t, b)})
	if got2 := sorted(got); len(got2) != 2 || got2[0] != want[0] || got2[1] != want[1] {
		t.Errorf("expected %v, got %v", want, got2)
	}
}

func TestResolve_literalPathNotFound(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.go")

	_, err := Resolve([]string{missing})
	if err == nil {
		t.Fatal("expected error for missing literal path")
	}
}

func TestResolve_relativeArgReturnsAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	mustWriteFile(t, f)
	want := wantCanonical(t, f)
	chdirTo(t, dir)

	got, err := Resolve([]string{"main.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
	if !filepath.IsAbs(got[0]) {
		t.Errorf("expected an absolute path, got %q", got[0])
	}
}

func TestResolve_dedupesAcrossRelativeAndAbsoluteSpelling(t *testing.T) {
	dir := t.TempDir()
	abs := filepath.Join(dir, "main.go")
	mustWriteFile(t, abs)
	want := wantCanonical(t, abs)
	chdirTo(t, dir)

	got, err := Resolve([]string{"main.go", abs})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected deduped [%s], got %v", want, got)
	}
}

func TestResolve_directoryWalkReturnsAbsolutePaths(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sub", "main.go")
	mustWriteFile(t, f)
	want := wantCanonical(t, f)
	chdirTo(t, dir)

	got, err := Resolve([]string{"sub"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

func TestResolve_globPatternReturnsAbsolutePaths(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	mustWriteFile(t, f)
	want := wantCanonical(t, f)
	chdirTo(t, dir)

	got, err := Resolve([]string{"*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

func TestResolve_absPathErrorPropagates(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	mustWriteFile(t, f)

	orig := absPath
	defer func() { absPath = orig }()
	absPath = func(_ string) (string, error) { return "", errPermDenied }

	_, err := Resolve([]string{f})
	if err == nil {
		t.Fatal("expected error to propagate from absPath")
	}
}

func TestResolve_evalSymlinksErrorPropagates(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	mustWriteFile(t, f)

	orig := evalSymlinks
	defer func() { evalSymlinks = orig }()
	evalSymlinks = func(_ string) (string, error) { return "", errPermDenied }

	_, err := Resolve([]string{f})
	if err == nil {
		t.Fatal("expected error to propagate from evalSymlinks")
	}
}

// --- Resolve: glob patterns ---

func TestResolve_globPatternMatches(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	ts := filepath.Join(dir, "a.ts")
	mustWriteFile(t, a)
	mustWriteFile(t, ts)

	got, err := Resolve([]string{filepath.Join(dir, "*.go")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := wantCanonical(t, a)
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

func TestResolve_globPatternNoMatchesIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.go"))

	got, err := Resolve([]string{filepath.Join(dir, "*.ts")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestResolve_starPatternDoesNotMatchNestedFiles(t *testing.T) {
	dir := t.TempDir()
	top := filepath.Join(dir, "main.go")
	nested := filepath.Join(dir, "sub", "foo.go")
	mustWriteFile(t, top)
	mustWriteFile(t, nested)

	got, err := Resolve([]string{filepath.Join(dir, "*.go")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := wantCanonical(t, top)
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected only [%s] (non-recursive *), got %v", want, got)
	}
}

func TestResolve_doubleStarPatternMatchesNestedFiles(t *testing.T) {
	dir := t.TempDir()
	top := filepath.Join(dir, "main.go")
	nested := filepath.Join(dir, "sub", "foo.go")
	mustWriteFile(t, top)
	mustWriteFile(t, nested)

	got, err := Resolve([]string{filepath.Join(dir, "**", "*.go")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sorted([]string{wantCanonical(t, top), wantCanonical(t, nested)})
	if got2 := sorted(got); len(got2) != 2 || got2[0] != want[0] || got2[1] != want[1] {
		t.Errorf("expected %v, got %v", want, got2)
	}
}

func TestResolve_bracePatternMatches(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	tsFile := filepath.Join(dir, "main.ts")
	pyFile := filepath.Join(dir, "main.py")
	mustWriteFile(t, goFile)
	mustWriteFile(t, tsFile)
	mustWriteFile(t, pyFile)

	got, err := Resolve([]string{filepath.Join(dir, "*.{go,ts}")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sorted([]string{wantCanonical(t, goFile), wantCanonical(t, tsFile)})
	if got2 := sorted(got); len(got2) != 2 || got2[0] != want[0] || got2[1] != want[1] {
		t.Errorf("expected %v, got %v", want, got2)
	}
}

func TestResolve_globPatternStaticPrefixOnlyWalksThatDir(t *testing.T) {
	dir := t.TempDir()
	inSub := filepath.Join(dir, "sub", "b.go")
	outsideSub := filepath.Join(dir, "other", "c.go")
	mustWriteFile(t, inSub)
	mustWriteFile(t, outsideSub)

	got, err := Resolve([]string{filepath.Join(dir, "sub", "*.go")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := wantCanonical(t, inSub)
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

func TestResolve_globPatternNonexistentPrefixIsEmpty(t *testing.T) {
	dir := t.TempDir()

	got, err := Resolve([]string{filepath.Join(dir, "nonexistent", "*.go")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

// --- Resolve: multiple args ---

func TestResolve_dedupesOverlappingArgs(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	mustWriteFile(t, a)

	got, err := Resolve([]string{dir, a})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := wantCanonical(t, a)
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected deduped [%s], got %v", want, got)
	}
}

func TestResolve_unionsDistinctArgs(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	b := filepath.Join(dir, "b.go")
	mustWriteFile(t, a)
	mustWriteFile(t, b)

	got, err := Resolve([]string{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sorted([]string{wantCanonical(t, a), wantCanonical(t, b)})
	if got2 := sorted(got); len(got2) != 2 || got2[0] != want[0] || got2[1] != want[1] {
		t.Errorf("expected %v, got %v", want, got2)
	}
}

func TestResolve_emptyPaths(t *testing.T) {
	got, err := Resolve(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no files, got %v", got)
	}
}

func TestResolve_stopsOnFirstError(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	mustWriteFile(t, a)
	missing := filepath.Join(dir, "missing.go")

	_, err := Resolve([]string{missing, a})
	if err == nil {
		t.Fatal("expected error from missing literal path")
	}
}

func TestResolveOne_globWalkErrorPropagates(t *testing.T) {
	origWalk := walkDir
	defer func() { walkDir = origWalk }()
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		_ = fn("root", nil, errPermDenied) // fsWalk's callback swallows entry errors
		return errPermDenied               // force walkDir itself to fail
	}

	_, err := resolveOne("*.go")
	if err == nil {
		t.Fatal("expected error to propagate from fsWalk")
	}
}

func TestResolveOne_directoryWalkErrorPropagates(t *testing.T) {
	dir := t.TempDir()

	origWalk := walkDir
	defer func() { walkDir = origWalk }()
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		_ = fn("root", nil, errPermDenied)
		return errPermDenied
	}

	_, err := resolveOne(dir)
	if err == nil {
		t.Fatal("expected error to propagate from fsWalk for a literal directory")
	}
}

func TestRebase_canonicalizeErrorPropagates(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.go")
	mustWriteFile(t, f)

	orig := absPath
	defer func() { absPath = orig }()
	absPath = func(_ string) (string, error) { return "", errPermDenied }

	_, err := rebase(dir, []string{f})
	if err == nil {
		t.Fatal("expected error to propagate from canonicalize")
	}
}

func TestRebase_relErrorPropagates(t *testing.T) {
	// root must exist so canonicalize succeeds; filepath.Rel then fails
	// because root is absolute and the file entry is relative — it has no
	// cwd to reconcile them against.
	dir := t.TempDir()
	_, err := rebase(dir, []string{filepath.Join("rel", "file.go")})
	if err == nil {
		t.Fatal("expected error to propagate from filepath.Rel")
	}
}

func TestRebase_emptyFilesSkipsCanonicalize(t *testing.T) {
	orig := absPath
	defer func() { absPath = orig }()
	called := false
	absPath = func(p string) (string, error) {
		called = true
		return orig(p)
	}

	got, err := rebase("nonexistent-root", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
	if called {
		t.Error("expected canonicalize to be skipped for empty files")
	}
}

// --- resolveOne: stat error that isn't "not found", but path has glob meta ---

func TestResolveOne_statErrorWithGlobMetaIsTreatedAsPattern(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	mustWriteFile(t, a)

	orig := statFile
	defer func() { statFile = orig }()
	statFile = func(_ string) (os.FileInfo, error) { return nil, errPermDenied }

	got, err := resolveOne(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := wantCanonical(t, a)
	if len(got) != 1 || got[0] != want {
		t.Errorf("expected [%s], got %v", want, got)
	}
}

// --- hasGlobMeta ---

func TestHasGlobMeta_star(t *testing.T) {
	if !hasGlobMeta("*.go") {
		t.Error("expected true for *.go")
	}
}

func TestHasGlobMeta_question(t *testing.T) {
	if !hasGlobMeta("file?.go") {
		t.Error("expected true for file?.go")
	}
}

func TestHasGlobMeta_bracket(t *testing.T) {
	if !hasGlobMeta("[abc].go") {
		t.Error("expected true for [abc].go")
	}
}

func TestHasGlobMeta_literal(t *testing.T) {
	if hasGlobMeta("main.go") {
		t.Error("expected false for main.go")
	}
}

// --- globPrefix ---

func TestGlobPrefix_noDirComponent(t *testing.T) {
	if got := globPrefix("*.go"); got != "." {
		t.Errorf("globPrefix(*.go) = %q, want %q", got, ".")
	}
}

func TestGlobPrefix_singleDir(t *testing.T) {
	want := filepath.FromSlash("src")
	if got := globPrefix("src/*.go"); got != want {
		t.Errorf("globPrefix(src/*.go) = %q, want %q", got, want)
	}
}

func TestGlobPrefix_doubleStarDir(t *testing.T) {
	want := filepath.FromSlash("src")
	if got := globPrefix("src/**/*.go"); got != want {
		t.Errorf("globPrefix(src/**/*.go) = %q, want %q", got, want)
	}
}

func TestGlobPrefix_nestedDir(t *testing.T) {
	want := filepath.FromSlash("internal/foo")
	if got := globPrefix("internal/foo/bar*.go"); got != want {
		t.Errorf("globPrefix(internal/foo/bar*.go) = %q, want %q", got, want)
	}
}

func TestGlobPrefix_absolutePath(t *testing.T) {
	want := filepath.FromSlash("/abs/path")
	if got := globPrefix("/abs/path/*.go"); got != want {
		t.Errorf("globPrefix(/abs/path/*.go) = %q, want %q", got, want)
	}
}

func TestGlobPrefix_noWildcard(t *testing.T) {
	want := filepath.FromSlash("src/foo.go")
	if got := globPrefix("src/foo.go"); got != want {
		t.Errorf("globPrefix(src/foo.go) = %q, want %q", got, want)
	}
}
