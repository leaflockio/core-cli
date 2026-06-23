// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package fsutil

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// blockerPath creates a file at <root>/blocker and returns a path that
// uses it as a directory component, triggering MkdirAll failures.
func blockerPath(t *testing.T, root, suffix string) string {
	t.Helper()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	return filepath.Join(blocker, suffix)
}

// --- EnsureDir ---

func TestEnsureDir_createsNestedDirs(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a", "b", "c")

	got, err := EnsureDir(path, 0o700)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
}

func TestEnsureDir_idempotent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "existing")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := EnsureDir(path, 0o700)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
}

func TestEnsureDir_error(t *testing.T) {
	path := blockerPath(t, t.TempDir(), "sub")

	_, err := EnsureDir(path, 0o700)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fsutil: create directory") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// --- EnsureParent ---

func TestEnsureParent_createsParent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a", "b", "file.txt")

	got, err := EnsureParent(path, 0o700)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("parent dir missing: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("file should not be created")
	}
}

func TestEnsureParent_idempotent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")

	got, err := EnsureParent(path, 0o700)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
}

func TestEnsureParent_error(t *testing.T) {
	path := blockerPath(t, t.TempDir(), filepath.Join("sub", "file.txt"))

	_, err := EnsureParent(path, 0o700)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fsutil: create parent for") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// --- WriteFile ---

func TestWriteFile_createsFileAndParent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sub", "file.txt")
	data := []byte("hello")

	if err := WriteFile(path, data, 0o700, 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestWriteFile_truncatesExisting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")

	if err := WriteFile(path, []byte("first"), 0o700, 0o600); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteFile(path, []byte("second"), 0o700, 0o600); err != nil {
		t.Fatalf("second write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}

func TestWriteFile_ensureParentError(t *testing.T) {
	path := blockerPath(t, t.TempDir(), filepath.Join("sub", "file.txt"))

	err := WriteFile(path, []byte("data"), 0o700, 0o600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteFile_writeError(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	err := WriteFile(target, []byte("data"), 0o700, 0o600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fsutil: write file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// --- CreateFile ---

func TestCreateFile_createsFileAndParent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sub", "file.txt")

	f, err := CreateFile(path, 0o700, 0o600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing: %v", err)
	}
}

func TestCreateFile_opensExisting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	f, err := CreateFile(path, 0o700, 0o600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Close()
}

func TestCreateFile_ensureParentError(t *testing.T) {
	path := blockerPath(t, t.TempDir(), filepath.Join("sub", "file.txt"))

	_, err := CreateFile(path, 0o700, 0o600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateFile_openError(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := CreateFile(target, 0o700, 0o600)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fsutil: create file") {
		t.Errorf("unexpected error message: %v", err)
	}
}
