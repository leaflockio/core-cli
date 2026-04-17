// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package pathutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// noModMarker is a marker name that will never exist on any real filesystem,
// used to force FindModuleRoot to always return false in fallback tests.
const noModMarker = "NO_MODULE_ROOT_MARKER_XYZ"

func TestFindModuleRoot_found(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ModuleRootMarker), []byte("module test\n"), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	sub := filepath.Join(root, "internal", "util")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, ok := FindModuleRoot(sub)
	if !ok {
		t.Fatal("expected module root to be found, got false")
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindModuleRoot_notFound(t *testing.T) {
	old := ModuleRootMarker
	ModuleRootMarker = noModMarker
	t.Cleanup(func() { ModuleRootMarker = old })

	_, ok := FindModuleRoot(t.TempDir())
	if ok {
		t.Error("expected module root not to be found, got true")
	}
}

func TestFindModuleRoot_customMarker(t *testing.T) {
	old := ModuleRootMarker
	ModuleRootMarker = "WORKSPACE"
	t.Cleanup(func() { ModuleRootMarker = old })

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "WORKSPACE"), []byte(""), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	got, ok := FindModuleRoot(root)
	if !ok {
		t.Fatal("expected module root to be found with custom marker, got false")
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestRelPath_withModuleRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ModuleRootMarker), []byte("module test\n"), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	file := filepath.Join(root, "internal", "config", "debug.go")
	got := RelPath(file)
	want := "internal/config/debug.go"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRelPath_fallback(t *testing.T) {
	old := ModuleRootMarker
	ModuleRootMarker = noModMarker
	t.Cleanup(func() { ModuleRootMarker = old })

	file := filepath.Join(t.TempDir(), "pkg", "file.go")
	got := RelPath(file)

	if !strings.HasSuffix(got, "pkg/file.go") {
		t.Errorf("got %q, want suffix pkg/file.go", got)
	}
	if parts := strings.Split(got, "/"); len(parts) != FallbackDepth {
		t.Errorf("got %d components in %q, want %d", len(parts), got, FallbackDepth)
	}
}

func TestRelPath_customFallbackDepth(t *testing.T) {
	oldMarker := ModuleRootMarker
	ModuleRootMarker = noModMarker
	t.Cleanup(func() { ModuleRootMarker = oldMarker })

	oldDepth := FallbackDepth
	FallbackDepth = 3
	t.Cleanup(func() { FallbackDepth = oldDepth })

	file := filepath.Join(t.TempDir(), "a", "b", "c", "file.go")
	got := RelPath(file)

	if parts := strings.Split(got, "/"); len(parts) != 3 {
		t.Errorf("got %d components in %q, want 3", len(parts), got)
	}
	if !strings.HasSuffix(got, "b/c/file.go") {
		t.Errorf("got %q, want suffix b/c/file.go", got)
	}
}

func TestRelPath_fallbackDepthExceedsComponents(t *testing.T) {
	oldMarker := ModuleRootMarker
	ModuleRootMarker = noModMarker
	t.Cleanup(func() { ModuleRootMarker = oldMarker })

	oldDepth := FallbackDepth
	FallbackDepth = 5
	t.Cleanup(func() { FallbackDepth = oldDepth })

	// Path has only 1 component — fewer than FallbackDepth, returns as-is.
	got := RelPath(filepath.Join(t.TempDir(), "file.go"))
	if !strings.HasSuffix(got, "file.go") {
		t.Errorf("got %q, want suffix file.go", got)
	}
}
