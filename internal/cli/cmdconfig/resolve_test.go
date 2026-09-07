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

func TestResolveConfig_missingDir_returnsEmptySources(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sources.Manifest != nil {
		t.Errorf("Manifest = %v, want nil", sources.Manifest)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone", got)
	}
}

func TestResolveConfig_emptyDir_returnsEmptySources(t *testing.T) {
	dir := t.TempDir()

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone", got)
	}
	if len(sources.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", sources.Warnings)
	}
}

func TestResolveConfig_dirUnreadable_returnsCCF003(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "blocker")
	write(t, parent, "blocker")

	_, err := ResolveConfig(filepath.Join(blocker, "leaf"), []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error when dir cannot be read, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF003 {
		t.Errorf("expected CCF003, got %v", err)
	}
}

func TestResolveConfig_manifestOnly_resolvesManifestCommandSource(t *testing.T) {
	dir := t.TempDir()
	data := []byte("license:\n  format: spdx\n")
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sources.Manifest == nil {
		t.Fatal("Manifest = nil, want non-nil")
	}
	if got := sources.Commands["license"]; got != SourceManifest {
		t.Errorf("Commands[license] = %v, want SourceManifest", got)
	}
	section, ok := sources.ManifestCommands["license"].(map[string]any)
	if !ok || section["format"] != "spdx" {
		t.Errorf("ManifestCommands[license] = %v, want format=spdx", sources.ManifestCommands["license"])
	}
}

func TestResolveConfig_manifestAmbiguousExtension_returnsCCF001(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest.yaml")
	write(t, dir, "manifest.json")

	_, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error for manifest with more than one extension, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF001 {
		t.Errorf("expected CCF001, got %v", err)
	}
}

func TestResolveConfig_manifestUnreadable_returnsCCF005(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte("not: [valid: yaml"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err == nil {
		t.Fatal("expected error for an unreadable manifest, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF005 {
		t.Errorf("expected CCF005, got %v", err)
	}
}

func TestResolveConfig_singleCommandFile_setsSourceFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")

	sources, err := ResolveConfig(dir, []string{"license", "doctor"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceFile {
		t.Errorf("Commands[license] = %v, want SourceFile", got)
	}
	if len(sources.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none for a command that consumes config", sources.Warnings)
	}
}

func TestResolveConfig_multipleCommandFiles_eachSetSourceFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "doctor.yaml")

	sources, err := ResolveConfig(dir, []string{"license", "doctor"}, []string{"license", "doctor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceFile {
		t.Errorf("Commands[license] = %v, want SourceFile", got)
	}
	if got := sources.Commands["doctor"]; got != SourceFile {
		t.Errorf("Commands[doctor] = %v, want SourceFile", got)
	}
}

func TestResolveConfig_recognizedCommandWithoutConfig_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "version.yaml")

	sources, err := ResolveConfig(dir, []string{"version"}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
	if _, ok := sources.Commands["version"]; ok {
		t.Errorf("Commands[version] = %v, want no entry — version does not declare config",
			sources.Commands["version"])
	}
}

func TestResolveConfig_unrecognizedFile_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "licence.yaml") // typo of "license"

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone", got)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}

func TestResolveConfig_unrecognizedFile_ambiguousExtensions_singleWarning(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "typo.yaml")
	write(t, dir, "typo.json")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1 combined warning", sources.Warnings)
	}
}

func TestResolveConfig_manifestAndCommandFile_recordsConflictNotError(t *testing.T) {
	dir := t.TempDir()
	data := []byte("license:\n  format: spdx\n")
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	write(t, dir, "license.yaml")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("a manifest/file conflict must not be a hard error from ResolveConfig: %v", err)
	}
	if len(sources.ManifestConflicts) != 1 || sources.ManifestConflicts[0] != "license" {
		t.Errorf("ManifestConflicts = %v, want [license]", sources.ManifestConflicts)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone — a conflicted command resolves to neither source", got)
	}
}

func TestResolveConfig_commandFileAmbiguousExtension_notFatal(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "license.json")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("a single command's own extension conflict must not be a hard error: %v", err)
	}
	exts, ok := sources.ExtensionConflicts["license"]
	if !ok || len(exts) != 2 {
		t.Errorf("ExtensionConflicts[license] = %v, want 2 extensions", exts)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone — an ambiguous file must not resolve to a source", got)
	}
}

func TestResolveConfig_strayExtensionAlone_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.toml")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone", got)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}

func TestResolveConfig_strayExtensionAlongsideDiscoverable_warnsButSetsSourceFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "license.yaml")
	write(t, dir, "license.toml")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceFile {
		t.Errorf("Commands[license] = %v, want SourceFile", got)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}

func TestResolveConfig_bareManifestNoExtension_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manifest")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sources.Manifest != nil {
		t.Errorf("Manifest = %v, want nil", sources.Manifest)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}

func TestResolveConfig_subdirectoriesIgnored(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "generated")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	write(t, sub, "license.yaml")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sources.Commands["license"]; got != SourceNone {
		t.Errorf("Commands[license] = %v, want SourceNone — a file inside a subdirectory must not be classified", got)
	}
	if len(sources.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", sources.Warnings)
	}
}

func TestResolveConfig_nonDiscoverableExtensionOnUnrecognizedBase_warns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "notes.md")

	sources, err := ResolveConfig(dir, []string{"license"}, []string{"license"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}

// TestResolveConfig_commandWithoutConfig_manifestSectionIgnored is a
// regression test: seedSources only seeds names from configCommands, so a
// manifest section under a command name that doesn't declare Config must
// never surface as a ManifestConflict — resolveCommandSources only walks
// s.Commands, which never contains that name.
func TestResolveConfig_commandWithoutConfig_manifestSectionIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte("version:\n  x: 1\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	write(t, dir, "version.yaml")

	sources, err := ResolveConfig(dir, []string{"version"}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources.ManifestConflicts) != 0 {
		t.Errorf("ManifestConflicts = %v, want none — version does not declare config",
			sources.ManifestConflicts)
	}
	if _, ok := sources.Commands["version"]; ok {
		t.Errorf("Commands[version] = %v, want no entry", sources.Commands["version"])
	}
}

// TestResolveConfig_ambiguousExtensionOnCommandWithoutConfig_doesNotError is
// a regression test for the version.yaml+version.json bug: an ambiguous
// extension for a command that never reads config must stay a warning, not
// escalate into ErrExtensionConflict.
func TestResolveConfig_ambiguousExtensionOnCommandWithoutConfig_doesNotError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "version.yaml")
	write(t, dir, "version.json")

	sources, err := ResolveConfig(dir, []string{"version"}, []string{})
	if err != nil {
		t.Fatalf("an ambiguous extension on a config-less command must not error: %v", err)
	}
	if _, ok := sources.ExtensionConflicts["version"]; ok {
		t.Errorf("ExtensionConflicts[version] = %v, want no entry", sources.ExtensionConflicts["version"])
	}
	if len(sources.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1", sources.Warnings)
	}
}
