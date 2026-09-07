// Copyright 2026 LeafLock. All rights reserved.
//
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
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/invocation"
)

var errPeekProjectRootFailed = errors.New("peek failed")

// parentStubCommand is a minimal cli.Command that declares Children, for
// exercising resolveSources's one-level-deep walk.
type parentStubCommand struct {
	use      string
	children []cli.Command
}

func (c *parentStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:     &cli.Meta{Use: c.use},
		Handler:  nopHandler,
		Children: c.children,
	}
}

// configStubCommand is a minimal cli.Command that declares a Config, for
// exercising resolveSources's configCommands collection.
type configStubCommand struct {
	use       string
	argsUsage string
}

func (c configStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    &cli.Meta{Use: c.use, ArgsUsage: c.argsUsage},
		Handler: nopHandler,
		Config:  stubConfigLoader{},
	}
}

func TestResolveSources_returns_error_when_peek_fails(t *testing.T) {
	old := peekProjectRoot
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return "", errPeekProjectRootFailed }
	t.Cleanup(func() { peekProjectRoot = old })

	root := &parentStubCommand{use: "root", children: []cli.Command{&stubCommand{use: "child"}}}
	_, err := resolveSources(root, nil)
	if !errors.Is(err, errPeekProjectRootFailed) {
		t.Errorf("error = %v, want errors.Is match for errPeekProjectRootFailed", err)
	}
}

func TestResolveSources_collects_configCommands(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return dir, nil }
	t.Cleanup(func() { peekProjectRoot = old })

	root := &parentStubCommand{use: "root", children: []cli.Command{
		configStubCommand{use: "license"},
		&stubCommand{use: "doctor"},
	}}
	sources, err := resolveSources(root, &app.App{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := sources.Commands["license"]; !ok {
		t.Error("Commands[license] missing, want an entry — license declares Config")
	}
	if _, ok := sources.Commands["doctor"]; ok {
		t.Error("Commands[doctor] present, want no entry — doctor does not declare Config")
	}
}

// TestResolveSources_recognizesCommandWithArgsUsage is a regression test: a
// command like check declares Meta.Use: "check" and ArgsUsage: "[file...]"
// separately. Before that split existed, Use held "check [file...]" and the
// catalog resolveSources builds would never match a real leaf/check.yaml
// against it, misclassifying a valid file as unrecognized.
func TestResolveSources_recognizesCommandWithArgsUsage(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return dir, nil }
	t.Cleanup(func() { peekProjectRoot = old })

	if err := os.WriteFile(filepath.Join(dir, "check.yaml"), []byte("x: 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := &parentStubCommand{use: "root", children: []cli.Command{
		configStubCommand{use: "check", argsUsage: "[file...]"},
	}}
	sources, err := resolveSources(root, &app.App{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources.Warnings) != 0 {
		t.Errorf("expected check.yaml to be recognized via bare Use, got warnings: %v", sources.Warnings)
	}
}

func TestResolveSources_extensionConflict_recordedNotErrored(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return dir, nil }
	t.Cleanup(func() { peekProjectRoot = old })

	if err := os.WriteFile(filepath.Join(dir, "license.yaml"), []byte("x: 1\n"), 0o600); err != nil {
		t.Fatalf("write license.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "license.json"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write license.json: %v", err)
	}

	root := &parentStubCommand{use: "root", children: []cli.Command{configStubCommand{use: "license"}}}
	sources, err := resolveSources(root, &app.App{})
	if err != nil {
		t.Fatalf("an extension conflict must not be a hard error from resolveSources: %v", err)
	}
	exts, ok := sources.ExtensionConflicts["license"]
	if !ok || len(exts) != 2 {
		t.Errorf("ExtensionConflicts[license] = %v, want 2 extensions", exts)
	}
}

func TestCheckConflicts_nilInvocation_doesNotEscalate(t *testing.T) {
	sources := &cmdconfig.Sources{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	if err := checkConflicts(&app.App{}, sources); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckConflicts_extensionConflictOnDifferentCommand_doesNotEscalate(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"doctor"}}}
	sources := &cmdconfig.Sources{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	if err := checkConflicts(a, sources); err != nil {
		t.Fatalf("unexpected error for a conflict on a command other than the one running: %v", err)
	}
}

func TestCheckConflicts_extensionConflictOnInvokedCommand_escalates(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license", "add"}}}
	sources := &cmdconfig.Sources{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	err := checkConflicts(a, sources)
	if err == nil {
		t.Fatal("expected error when the conflict belongs to the command actually being invoked")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF004 {
		t.Errorf("expected CCF004, got %v", err)
	}
}

func TestCheckConflicts_manifestConflictOnDifferentCommand_doesNotEscalate(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"doctor"}}}
	sources := &cmdconfig.Sources{ManifestConflicts: []string{"license"}}
	if err := checkConflicts(a, sources); err != nil {
		t.Fatalf("unexpected error for a conflict on a command other than the one running: %v", err)
	}
}

func TestCheckConflicts_manifestConflictOnInvokedCommand_escalates(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}
	sources := &cmdconfig.Sources{ManifestConflicts: []string{"license"}}
	err := checkConflicts(a, sources)
	if err == nil {
		t.Fatal("expected error when the conflict belongs to the command actually being invoked")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF002 {
		t.Errorf("expected CCF002, got %v", err)
	}
}
