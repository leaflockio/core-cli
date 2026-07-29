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
// exercising checkConfigLayout's one-level-deep walk.
type parentStubCommand struct {
	use      string
	children []cli.Command
}

func (c *parentStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:     cli.Meta{Use: c.use},
		Handler:  nopHandler,
		Children: c.children,
	}
}

// configStubCommand is a minimal cli.Command that declares a Config, for
// exercising checkConfigLayout's configCommands collection.
type configStubCommand struct{ use string }

func (c configStubCommand) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Handler: nopHandler,
		Config:  stubConfigLoader{},
	}
}

func TestCheckConfigLayout_returns_error_when_peek_fails(t *testing.T) {
	old := peekProjectRoot
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return "", errPeekProjectRootFailed }
	t.Cleanup(func() { peekProjectRoot = old })

	root := &parentStubCommand{use: "root", children: []cli.Command{&stubCommand{use: "child"}}}
	_, err := checkConfigLayout(root, nil)
	if !errors.Is(err, errPeekProjectRootFailed) {
		t.Errorf("error = %v, want errors.Is match for errPeekProjectRootFailed", err)
	}
}

func TestCheckConfigLayout_collects_config_declaring_children(t *testing.T) {
	old := peekProjectRoot
	dir := t.TempDir()
	peekProjectRoot = func(_ *app.App, _ string) (string, error) { return dir, nil }
	t.Cleanup(func() { peekProjectRoot = old })

	root := &parentStubCommand{use: "root", children: []cli.Command{
		configStubCommand{use: "license"},
		&stubCommand{use: "doctor"},
	}}
	layout, err := checkConfigLayout(root, &app.App{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout == nil {
		t.Fatal("layout should not be nil on success")
	}
}

func TestCheckConfigLayout_returns_error_when_invoked_command_has_extension_conflict(t *testing.T) {
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
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license"}}}

	_, err := checkConfigLayout(root, a)
	if err == nil {
		t.Fatal("expected error when the invoked command's own config has an extension conflict")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF004 {
		t.Errorf("expected CCF004, got %v", err)
	}
}

func TestCheckExtensionConflict_nilInvocation_doesNotEscalate(t *testing.T) {
	layout := &cmdconfig.Layout{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	if err := checkExtensionConflict(&app.App{}, layout); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckExtensionConflict_conflictOnDifferentCommand_doesNotEscalate(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"doctor"}}}
	layout := &cmdconfig.Layout{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	if err := checkExtensionConflict(a, layout); err != nil {
		t.Fatalf("unexpected error for a conflict on a command other than the one running: %v", err)
	}
}

func TestCheckExtensionConflict_conflictOnInvokedCommand_escalates(t *testing.T) {
	a := &app.App{Invocation: &invocation.Invocation{Raw: []string{"license", "add"}}}
	layout := &cmdconfig.Layout{ExtensionConflicts: map[string][]string{"license": {"yaml", "json"}}}
	err := checkExtensionConflict(a, layout)
	if err == nil {
		t.Fatal("expected error when the conflict belongs to the command actually being invoked")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.CCF004 {
		t.Errorf("expected CCF004, got %v", err)
	}
}
