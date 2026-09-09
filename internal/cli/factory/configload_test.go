// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/configfield"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/invocation"
)

var (
	errLoadFailed     = errors.New("load failed")
	errValidateFailed = errors.New("validate failed")
)

// recordingConfigLoader records whether Load/Validate were called and what
// section Load received, so tests can assert on the skip-when-nil behavior.
type recordingConfigLoader struct {
	loadedSection  map[string]any
	loadCalled     bool
	validateCalled bool
	loadErr        error
	validateErr    error
}

func (r *recordingConfigLoader) Load(section map[string]any) error {
	r.loadCalled = true
	r.loadedSection = section
	return r.loadErr
}

func (r *recordingConfigLoader) Validate() error {
	r.validateCalled = true
	return r.validateErr
}

// configLoaderStubCommand is a cli.Command that carries a caller-supplied
// configfield.ConfigLoader, for asserting on Load/Validate call behavior.
type configLoaderStubCommand struct {
	use string
	cfg configfield.ConfigLoader
}

func (c configLoaderStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    &cli.Meta{Use: c.use},
		Handler: nopHandler,
		Config:  c.cfg,
	}
}

// peekJoin returns a peekProjectRoot override that joins dir and the
// requested name, matching what Workspace.Peek does for real.
func peekJoin(dir string) func(*app.App, string) (string, error) {
	return func(_ *app.App, name string) (string, error) {
		return filepath.Join(dir, name), nil
	}
}

func assertCCF005(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF005 {
		t.Errorf("expected CCF005, got %v", err)
	}
}

// --- findChildByName ---

func TestFindChildByName_matches_caseInsensitively(t *testing.T) {
	root := &parentStubCommand{use: "root", children: []cli.Command{
		&stubCommand{use: "License"},
	}}
	def := findChildByName(root, &app.App{}, "license")
	if def == nil || def.Meta.Use != "License" {
		t.Fatalf("findChildByName() = %v, want the \"License\" child", def)
	}
}

func TestFindChildByName_noMatch_returnsNil(t *testing.T) {
	root := &parentStubCommand{use: "root", children: []cli.Command{
		&stubCommand{use: "doctor"},
	}}
	if def := findChildByName(root, &app.App{}, "license"); def != nil {
		t.Errorf("findChildByName() = %v, want nil", def)
	}
}

// --- resolveSection ---

func TestResolveSection_sourceNone_returnsNil(t *testing.T) {
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceNone}}
	section, err := resolveSection(&app.App{}, "license", sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section != nil {
		t.Errorf("section = %v, want nil", section)
	}
}

func TestResolveSection_sourceFile_readsOwnFile(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("format: spdx\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}
	section, err := resolveSection(&app.App{}, "license", sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section["format"] != "spdx" {
		t.Errorf("section = %v, want format=spdx", section)
	}
}

func TestResolveSection_sourceFile_decodeError_returnsCCF005(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("not: [valid: yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}
	_, err := resolveSection(&app.App{}, "license", sources)
	assertCCF005(t, err)
}

func TestResolveSection_sourceManifest_returnsManifestSection(t *testing.T) {
	sources := &cmdconfig.Sources{
		Commands:         map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceManifest},
		ManifestCommands: map[string]any{"license": map[string]any{"format": "spdx"}},
	}
	section, err := resolveSection(&app.App{}, "license", sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section["format"] != "spdx" {
		t.Errorf("section = %v, want format=spdx", section)
	}
}

func TestResolveSection_sourceManifest_missingSection_returnsNil(t *testing.T) {
	sources := &cmdconfig.Sources{
		Commands:         map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceManifest},
		ManifestCommands: map[string]any{"doctor": map[string]any{"verbose": true}},
	}
	section, err := resolveSection(&app.App{}, "license", sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section != nil {
		t.Errorf("section = %v, want nil", section)
	}
}

// --- loadInvokedConfig ---

func TestLoadInvokedConfig_nilInvocation_noop(t *testing.T) {
	root := &parentStubCommand{use: "root"}
	if err := loadInvokedConfig(root, &app.App{}, &cmdconfig.Sources{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadInvokedConfig_noMatchingChild_noop(t *testing.T) {
	root := &parentStubCommand{use: "root", children: []cli.Command{&stubCommand{use: "doctor"}}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	if err := loadInvokedConfig(root, a, &cmdconfig.Sources{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadInvokedConfig_matchedChildWithoutConfig_noop(t *testing.T) {
	root := &parentStubCommand{use: "root", children: []cli.Command{&stubCommand{use: "license"}}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	if err := loadInvokedConfig(root, a, &cmdconfig.Sources{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadInvokedConfig_peekFails_propagatesError(t *testing.T) {
	old := peekProjectRoot
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return "", errPeekProjectRootFailed }
	t.Cleanup(func() { peekProjectRoot = old })

	rec := &recordingConfigLoader{}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}

	err := loadInvokedConfig(root, a, sources)
	if !errors.Is(err, errPeekProjectRootFailed) {
		t.Errorf("error = %v, want errors.Is match for errPeekProjectRootFailed", err)
	}
}

func TestLoadInvokedConfig_resolveSectionError_propagates(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("not: [valid: yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	rec := &recordingConfigLoader{}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}

	err := loadInvokedConfig(root, a, sources)
	assertCCF005(t, err)
	if rec.loadCalled {
		t.Error("Load must not be called when the section can't be resolved")
	}
}

func TestLoadInvokedConfig_sourceNone_skipsLoadAndValidate(t *testing.T) {
	rec := &recordingConfigLoader{}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceNone}}

	if err := loadInvokedConfig(root, a, sources); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.loadCalled || rec.validateCalled {
		t.Error("Load/Validate must not be called when there is no config section")
	}
}

func TestLoadInvokedConfig_sourceFile_loadsAndValidates(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("format: spdx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	rec := &recordingConfigLoader{}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}

	if err := loadInvokedConfig(root, a, sources); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rec.loadCalled || !rec.validateCalled {
		t.Error("Load and Validate must both be called when a config section exists")
	}
	if rec.loadedSection["format"] != "spdx" {
		t.Errorf("loadedSection = %v, want format=spdx", rec.loadedSection)
	}
}

func TestLoadInvokedConfig_sourceManifest_loadsAndValidates(t *testing.T) {
	rec := &recordingConfigLoader{}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{
		Commands:         map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceManifest},
		ManifestCommands: map[string]any{"license": map[string]any{"format": "spdx"}},
	}

	if err := loadInvokedConfig(root, a, sources); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rec.loadCalled || !rec.validateCalled {
		t.Error("Load and Validate must both be called when a config section exists")
	}
	if rec.loadedSection["format"] != "spdx" {
		t.Errorf("loadedSection = %v, want format=spdx", rec.loadedSection)
	}
}

func TestLoadInvokedConfig_loadError_propagatesAndSkipsValidate(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("format: spdx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	rec := &recordingConfigLoader{loadErr: errLoadFailed}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}

	err := loadInvokedConfig(root, a, sources)
	if !errors.Is(err, errLoadFailed) {
		t.Errorf("error = %v, want errors.Is match for errLoadFailed", err)
	}
	if rec.validateCalled {
		t.Error("Validate must not be called when Load fails")
	}
}

func TestLoadInvokedConfig_validateError_propagates(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("format: spdx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	peekProjectRoot = peekJoin(dir)
	t.Cleanup(func() { peekProjectRoot = old })

	rec := &recordingConfigLoader{validateErr: errValidateFailed}
	root := &parentStubCommand{use: "root", children: []cli.Command{
		configLoaderStubCommand{use: "license", cfg: rec},
	}}
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{Commands: map[string]cmdconfig.CommandSource{"license": cmdconfig.SourceFile}}

	err := loadInvokedConfig(root, a, sources)
	if !errors.Is(err, errValidateFailed) {
		t.Errorf("error = %v, want errors.Is match for errValidateFailed", err)
	}
}
