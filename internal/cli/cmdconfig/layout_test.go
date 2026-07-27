// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

func write(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x: 1\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestDetectLayout_missingDir_isNone(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone", layout.Mode)
	}
}

func TestDetectLayout_emptyDir_isNone(t *testing.T) {
	dir := t.TempDir()

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone", layout.Mode)
	}
}

func TestDetectLayout_dirUnreadable_returnsCCF003(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "blocker")
	write(t, parent, "blocker")

	_, err := DetectLayout(filepath.Join(blocker, "leaf"), []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error when dir cannot be read, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF003 {
		t.Errorf("expected CCF003, got %v", err)
	}
}

func TestDetectLayout_manifestOnly_isFlat(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest.yaml")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutFlat {
		t.Errorf("Mode = %v, want LayoutFlat", layout.Mode)
	}
}

func TestDetectLayout_manifestAmbiguousExtension_returnsCCF001(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest.yaml")
	write(t, dir, "manifest.json")

	_, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error for manifest with more than one extension, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF001 {
		t.Errorf("expected CCF001, got %v", err)
	}
}

func TestDetectLayout_singleCommandFile_isModular(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")

	layout, err := DetectLayout(dir, []string{"license", "doctor"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutModular {
		t.Errorf("Mode = %v, want LayoutModular", layout.Mode)
	}
	if len(layout.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none for a command that consumes config", layout.Warnings)
	}
}

func TestDetectLayout_multipleCommandFiles_areModular(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "doctor.yaml")

	layout, err := DetectLayout(dir, []string{"license", "doctor"}, []string{"license", "doctor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutModular {
		t.Errorf("Mode = %v, want LayoutModular", layout.Mode)
	}
}

func TestDetectLayout_recognizedCommandWithoutConfig_warnsInert(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "version.yaml")

	layout, err := DetectLayout(dir, []string{"version"}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutModular {
		t.Errorf("Mode = %v, want LayoutModular", layout.Mode)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}

func TestDetectLayout_unrecognizedFile_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "licence.yaml") // typo of "license"

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone", layout.Mode)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}

func TestDetectLayout_unrecognizedFile_ambiguousExtensions_singleWarning(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "typo.yaml")
	write(t, dir, "typo.json")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1 combined warning", layout.Warnings)
	}
}

func TestDetectLayout_manifestAndCommandFile_returnsCCF002(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest.yaml")
	write(t, dir, "license.yaml")

	_, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error when manifest coexists with a per-command file, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF002 {
		t.Errorf("expected CCF002, got %v", err)
	}
}

func TestDetectLayout_commandFileAmbiguousExtension_notFatal(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "license.json")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("a single command's own extension conflict must not be a hard error: %v", err)
	}
	exts, ok := layout.ExtensionConflicts["license"]
	if !ok || len(exts) != 2 {
		t.Errorf("ExtensionConflicts[license] = %v, want 2 extensions", exts)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone — an ambiguous command file must not count as recognized", layout.Mode)
	}
}

func TestDetectLayout_strayExtensionAlone_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.toml")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone", layout.Mode)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}

func TestDetectLayout_strayExtensionAlongsideDiscoverable_warnsButModular(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "license.toml")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutModular {
		t.Errorf("Mode = %v, want LayoutModular", layout.Mode)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}

func TestDetectLayout_bareManifestNoExtension_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone", layout.Mode)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}

func TestDetectLayout_subdirectoriesIgnored(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "generated")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	write(t, sub, "license.yaml")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Mode != LayoutNone {
		t.Errorf("Mode = %v, want LayoutNone — a file inside a subdirectory must not be classified", layout.Mode)
	}
}

func TestDetectLayout_nonDiscoverableExtensionOnUnrecognizedBase_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "notes.md")

	layout, err := DetectLayout(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", layout.Warnings)
	}
}
