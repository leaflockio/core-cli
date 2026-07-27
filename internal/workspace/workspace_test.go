// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/spf13/cobra"
)

var (
	errHomeResolutionFailed  = errors.New("home directory resolution failed")
	errCacheResolutionFailed = errors.New("cache dir resolution failed")
)

// Build a cobra command tree from the given sub-command names and return the
// deepest command. For example, sub "pr", "create" returns the create command
// whose CommandPath() is "gh pr create".
func newCmd(sub ...string) *cobra.Command {
	root := &cobra.Command{Use: "gh"}
	parent := root
	for _, s := range sub {
		child := &cobra.Command{Use: s}
		parent.AddCommand(child)
		parent = child
	}
	return parent
}

// --- New ---

func TestNew_explicitHomeDir(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()

	ws, err := New(home, repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantEntityRoot := filepath.Join(home, config.EntityFolder)
	wantUserRoot := filepath.Join(home, config.EntityFolder, config.AppName)
	wantRepoRoot := filepath.Join(repo, config.AppName)

	if ws.entityRoot != wantEntityRoot {
		t.Errorf("entityRoot: got %q, want %q", ws.entityRoot, wantEntityRoot)
	}
	if ws.userRoot != wantUserRoot {
		t.Errorf("userRoot: got %q, want %q", ws.userRoot, wantUserRoot)
	}
	if ws.repoRoot != wantRepoRoot {
		t.Errorf("repoRoot: got %q, want %q", ws.repoRoot, wantRepoRoot)
	}
}

func TestNew_emptyHomeDirResolvesFromOS(t *testing.T) {
	repo := t.TempDir()

	ws, err := New("", repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.userRoot == "" {
		t.Error("expected userRoot to be set")
	}
}

func TestNew_homeResolutionError(t *testing.T) {
	old := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errHomeResolutionFailed }
	t.Cleanup(func() { osUserHomeDir = old })

	_, err := New("", t.TempDir())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.WSP001 {
		t.Errorf("expected WSP001, got %v", err)
	}
}

// --- CredentialsPath ---

func TestCredentialsPath(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	got := ws.CredentialsPath()
	want := filepath.Join(home, config.EntityFolder, config.CredentialsFile)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- ForUser ---

func TestForUser_depthRoot(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	cs := ws.ForUser(newCmd(), level.LevelRoot)

	dir, err := cs.Dir("cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, config.EntityFolder, config.AppName, "cache")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
	assertPerm(t, dir, 0o700)
}

func TestForUser_depthCommand(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	cs := ws.ForUser(newCmd("pr", "create"), level.LevelTop)

	dir, err := cs.Dir("cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("pr", "cache")) {
		t.Errorf("expected path to end with pr/cache, got %q", dir)
	}
}

func TestForUser_depthFull(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	cs := ws.ForUser(newCmd("pr", "create"), level.LevelNested)

	dir, err := cs.Dir("cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("pr", "create", "cache")) {
		t.Errorf("expected path to end with pr/create/cache, got %q", dir)
	}
}

func TestForUser_rootCmdWithLevelTop(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	// root command has no sub-path — commandSubPath returns ""
	cs := ws.ForUser(newCmd(), level.LevelTop)

	dir, err := cs.Dir("cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, config.EntityFolder, config.AppName, "cache")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestForUser_rootCmdWithLevelNested(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	cs := ws.ForUser(newCmd(), level.LevelNested)

	dir, err := cs.Dir("cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, config.EntityFolder, config.AppName, "cache")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

// --- ForRepo ---

func TestForRepo_depthRoot(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForRepo(newCmd(), level.LevelRoot)

	dir, err := cs.Dir("out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName, "out")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
	assertPerm(t, dir, 0o755)
}

func TestForRepo_depthCommand(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForRepo(newCmd("pr", "create"), level.LevelTop)

	dir, err := cs.Dir("out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("pr", "out")) {
		t.Errorf("expected path to end with pr/out, got %q", dir)
	}
}

func TestForRepo_depthFull(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForRepo(newCmd("pr", "create"), level.LevelNested)

	dir, err := cs.Dir("out")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("pr", "create", "out")) {
		t.Errorf("expected path to end with pr/create/out, got %q", dir)
	}
}

// --- ForProjectRoot ---

func TestForProjectRoot(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForProjectRoot()

	dir, err := cs.Dir("templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName, "templates")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
	assertPerm(t, dir, 0o755)
}

// --- ForGenerated ---

func TestForGenerated_depthCommand(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForGenerated(newCmd("license"), level.LevelTop)

	file, err := cs.File("license.lock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName, "generated", "license", "license.lock")
	if file != want {
		t.Errorf("got %q, want %q", file, want)
	}
}

func TestForGenerated_depthRoot(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForGenerated(newCmd("license"), level.LevelRoot)

	file, err := cs.File("out.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName, "generated", "out.txt")
	if file != want {
		t.Errorf("got %q, want %q", file, want)
	}
}

// --- ForCache ---

func TestForCache(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)

	cs := ws.ForCache(newCmd("license"), "templates")

	dir, err := cs.Dir("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("license", "templates", "abc123")) {
		t.Errorf("expected path to end with license/templates/abc123, got %q", dir)
	}
	assertPerm(t, dir, 0o755)
}

func TestForCache_purposeIsolation(t *testing.T) {
	ws, _ := New(t.TempDir(), t.TempDir())

	templates := ws.ForCache(newCmd("license"), "templates")
	spdx := ws.ForCache(newCmd("license"), "spdx")

	td, err := templates.Dir("h1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sd, err := spdx.Dir("h1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td == sd {
		t.Errorf("expected different purposes to resolve to different dirs, both got %q", td)
	}
}

func TestForCache_disabledWhenUnresolvable(t *testing.T) {
	old := osUserCacheDir
	osUserCacheDir = func() (string, error) { return "", errCacheResolutionFailed }
	t.Cleanup(func() { osUserCacheDir = old })

	ws, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cs := ws.ForCache(newCmd("license"), "templates")

	if _, err := cs.Dir("h1"); err == nil {
		t.Fatal("Dir: expected error, got nil")
	} else {
		var e *errs.Error
		if !errors.As(err, &e) || e.Code != errs.WSP002 {
			t.Errorf("Dir: expected WSP002, got %v", err)
		}
	}

	if _, err := cs.File("meta.json"); err == nil {
		t.Fatal("File: expected error, got nil")
	} else {
		var e *errs.Error
		if !errors.As(err, &e) || e.Code != errs.WSP002 {
			t.Errorf("File: expected WSP002, got %v", err)
		}
	}
}

// --- Dir ---

func TestCommandSpace_Dir_creates(t *testing.T) {
	ws, _ := New(t.TempDir(), t.TempDir())
	cs := ws.ForUser(newCmd("pr"), level.LevelTop)

	dir, err := cs.Dir("templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
}

func TestCommandSpace_Dir_error(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	// plant a file where the base directory would be created
	blocker := filepath.Join(home, config.EntityFolder, config.AppName, "pr")
	if err := os.MkdirAll(filepath.Dir(blocker), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cs := ws.ForUser(newCmd("pr"), level.LevelTop)
	_, err := cs.Dir("templates")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- File ---

func TestCommandSpace_File_createsParent(t *testing.T) {
	ws, _ := New(t.TempDir(), t.TempDir())
	cs := ws.ForRepo(newCmd("pr"), level.LevelTop)

	path, err := cs.File("pr.lock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "pr.lock" {
		t.Errorf("expected filename pr.lock, got %q", filepath.Base(path))
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("parent dir missing: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("file should not be created")
	}
}

func TestCommandSpace_File_error(t *testing.T) {
	home := t.TempDir()
	ws, _ := New(home, t.TempDir())

	blocker := filepath.Join(home, config.EntityFolder, config.AppName, "pr")
	if err := os.MkdirAll(filepath.Dir(blocker), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cs := ws.ForUser(newCmd("pr"), level.LevelTop)
	_, err := cs.File("pr.lock")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Peek ---

func TestCommandSpace_Peek_emptyNameReturnsBasePath(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)
	cs := ws.ForProjectRoot()

	path, err := cs.Peek("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName)
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("Peek must not create anything on disk")
	}
}

func TestCommandSpace_Peek_returnsNamedPath(t *testing.T) {
	repo := t.TempDir()
	ws, _ := New(t.TempDir(), repo)
	cs := ws.ForProjectRoot()

	path, err := cs.Peek("manifest.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(repo, config.AppName, "manifest.yaml")
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("Peek must not create anything on disk")
	}
}

func TestCommandSpace_Peek_error(t *testing.T) {
	old := osUserCacheDir
	osUserCacheDir = func() (string, error) { return "", errCacheResolutionFailed }
	t.Cleanup(func() { osUserCacheDir = old })

	ws, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cs := ws.ForCache(newCmd("license"), "templates")
	if _, err := cs.Peek("h1"); err == nil {
		t.Fatal("expected error, got nil")
	} else {
		var e *errs.Error
		if !errors.As(err, &e) || e.Code != errs.WSP002 {
			t.Errorf("expected WSP002, got %v", err)
		}
	}
}

// assertPerm checks that the directory at path has the expected permission bits.
func assertPerm(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("perm %s: got %04o, want %04o", path, got, want)
	}
}
