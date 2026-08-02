// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/cli/factory"
	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/invocation"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/leaflockio/core-cli/internal/workspace"
	"github.com/spf13/cobra"
)

// testWorkspace returns a Workspace rooted at a fresh temp dir, isolated
// from the real filesystem and from other tests.
func testWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	ws, err := workspace.New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	return ws
}

// testApp returns a minimal *app.App with a Workspace set — Build now reads
// a.Workspace to locate the project config directory before assembling.
func testApp(t *testing.T) *app.App {
	t.Helper()
	return app.NewBuilder().WithWorkspace(testWorkspace(t)).Build()
}

// testAppWithConfigDir is like testApp, but also returns the resolved
// project config directory path so a test can write files into it before
// calling Build.
func testAppWithConfigDir(t *testing.T) (*app.App, string) {
	t.Helper()
	repoRoot := t.TempDir()
	ws, err := workspace.New(t.TempDir(), repoRoot)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	return app.NewBuilder().WithWorkspace(ws).Build(), filepath.Join(repoRoot, config.AppName)
}

func TestFactory_Build_returns_error_for_nil_command(t *testing.T) {
	_, err := factory.New().Build(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil command, got nil")
	}
}

func TestFactory_Build_returns_error_for_nil_app(t *testing.T) {
	_, err := factory.New().Build(&stubCmd{use: "test"}, nil)
	if err == nil {
		t.Fatal("expected error for nil app, got nil")
	}
}

func TestFactory_Build_returns_error_for_nil_workspace(t *testing.T) {
	a := app.NewBuilder().Build()
	_, err := factory.New().Build(&stubCmd{use: "test"}, a)
	if err == nil {
		t.Fatal("expected error for nil Workspace, got nil")
	}
}

func TestFactory_Build_returns_error_when_config_layout_conflicts(t *testing.T) {
	a, configDir := testAppWithConfigDir(t)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "manifest.yaml"), []byte("x: 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "child.yaml"), []byte("x: 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile child: %v", err)
	}

	parent := &parentStubCmd{use: "parent", children: []cli.Command{&stubCmd{use: "child"}}}
	_, err := factory.New().Build(parent, a)
	if err == nil {
		t.Fatal("expected error when config layout conflicts, got nil")
	}
}

func TestFactory_Build_returns_cobra_command(t *testing.T) {
	cmd, err := factory.New().Build(&stubCmd{use: "test"}, testApp(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Build must return a non-nil cobra.Command")
	}
	if cmd.Use != "test" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test")
	}
}

func TestFactory_Build_returns_error_when_child_definition_invalid(t *testing.T) {
	parent := &parentStubCmd{use: "parent", children: []cli.Command{&nilHandlerCmd{}}}
	_, err := factory.New().Build(parent, testApp(t))
	if err == nil {
		t.Fatal("expected error for invalid child definition, got nil")
	}
}

func TestFactory_Build_allows_root_with_nil_handler_and_no_children(t *testing.T) {
	cmd, err := factory.New().Build(&nilHandlerCmd{}, testApp(t))
	if err != nil {
		t.Fatalf("unexpected error for root with nil Handler and no Children: %v", err)
	}
	if cmd == nil {
		t.Fatal("Build must return a non-nil cobra.Command")
	}
}

// TestFactory_Build_returns_error_when_grandchild_definition_invalid guards
// against the exemption leaking past the literal command passed to Build: a
// mid-tree command with its own children (e.g. a "license" group) must not
// be treated as a root just because it's the top of its own subtree.
func TestFactory_Build_returns_error_when_grandchild_definition_invalid(t *testing.T) {
	mid := &parentStubCmd{use: "mid", children: []cli.Command{&nilHandlerCmd{}}}
	root := &parentStubCmd{use: "root", children: []cli.Command{mid}}

	_, err := factory.New().Build(root, testApp(t))
	if err == nil {
		t.Fatal("expected error for invalid grandchild definition, got nil")
	}
}

func TestFactory_Build_registers_no_color_as_persistent_flag_on_root(t *testing.T) {
	cmd, err := factory.New().Build(&stubCmd{use: "test"}, testApp(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.PersistentFlags().Lookup("no-color") == nil {
		t.Error("expected --no-color to be registered as a persistent flag on the built root command")
	}
}

func TestFactory_Build_child_does_not_get_its_own_no_color_flag(t *testing.T) {
	parent := &parentStubCmd{use: "parent", children: []cli.Command{&stubCmd{use: "child"}}}

	cmd, err := factory.New().Build(parent, testApp(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := cmd.Commands()[0]
	if child.Flags().Lookup("no-color") != nil {
		t.Error("child command must not have its own local --no-color flag")
	}
	if child.PersistentFlags().Lookup("no-color") != nil {
		t.Error("child command must not have its own persistent --no-color flag")
	}
}

func TestFactory_Build_no_color_effect_fires_on_parse(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := app.NewBuilder().WithPrinter(printer).WithWorkspace(testWorkspace(t)).Build()

	cmd, err := factory.New().Build(&stubCmd{use: "test"}, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd.SetArgs([]string{"--no-color"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if got := printer.Primary("x"); got != "x" {
		t.Errorf("expected plain text after --no-color, got %q", got)
	}
}

func TestFactory_Build_no_color_effect_fires_when_set_after_subcommand(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := app.NewBuilder().WithPrinter(printer).WithWorkspace(testWorkspace(t)).Build()

	parent := &parentStubCmd{use: "parent", children: []cli.Command{&stubCmd{use: "child"}}}
	cmd, err := factory.New().Build(parent, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd.SetArgs([]string{"child", "--no-color"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if got := printer.Primary("x"); got != "x" {
		t.Errorf("expected plain text after --no-color set on child, got %q", got)
	}
}

// TestFactory_Build_no_color_effect_fires_on_handlerless_bare_invocation guards
// against a regression where a Handler-less command (like root) invoked with no
// subcommand — e.g. `leaf --no-color` — relied on cobra's built-in non-Runnable
// shortcut, which returns flag.ErrHelp before cobra's preRun() ever runs,
// silently skipping PersistentPreRunE. Firing the effect at flag-parse time
// instead (see effectvalue.go) sidesteps that shortcut entirely: parsing always
// happens, Runnable or not, so the effect fires regardless.
func TestFactory_Build_no_color_effect_fires_on_handlerless_bare_invocation(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := app.NewBuilder().WithPrinter(printer).WithWorkspace(testWorkspace(t)).Build()

	parent := &handlerlessParentStubCmd{use: "parent", children: []cli.Command{&stubCmd{use: "child"}}}
	cmd, err := factory.New().Build(parent, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"--no-color"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if got := printer.Primary("x"); got != "x" {
		t.Errorf("expected plain text after --no-color on bare handler-less invocation, got %q", got)
	}
}

type stubCmd struct{ use string }

func (c *stubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
	}
}

type parentStubCmd struct {
	use      string
	children []cli.Command
}

func (c *parentStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:     cli.Meta{Use: c.use},
		Handler:  func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		Children: c.children,
	}
}

type handlerlessParentStubCmd struct {
	use      string
	children []cli.Command
}

func (c *handlerlessParentStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:     cli.Meta{Use: c.use},
		Children: c.children,
	}
}

type nilHandlerCmd struct{}

func (c *nilHandlerCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{Meta: cli.Meta{Use: "bad"}}
}

type groupedParentStubCmd struct {
	use      string
	groups   []cli.Group
	children []cli.Command
}

func (c *groupedParentStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:     cli.Meta{Use: c.use},
		Groups:   c.groups,
		Handler:  func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		Children: c.children,
	}
}

type groupedChildStubCmd struct {
	use   string
	group cli.GroupID
}

func (c *groupedChildStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Group:   c.group,
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
	}
}

func TestFactory_Build_returns_error_when_child_group_undeclared(t *testing.T) {
	parent := &parentStubCmd{use: "parent", children: []cli.Command{&groupedChildStubCmd{use: "child", group: "read"}}}
	_, err := factory.New().Build(parent, testApp(t))
	if err == nil {
		t.Fatal("expected error when a child's Group isn't declared in its parent's Groups, got nil")
	}
}

func TestFactory_Build_allows_child_group_declared_in_parent_groups(t *testing.T) {
	parent := &groupedParentStubCmd{
		use:      "parent",
		groups:   []cli.Group{{ID: "read", Title: "read"}},
		children: []cli.Command{&groupedChildStubCmd{use: "child", group: "read"}},
	}
	_, err := factory.New().Build(parent, testApp(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// nopConfigLoader is a minimal cmdconfig.ConfigLoader for Build-level tests.
type nopConfigLoader struct{}

func (nopConfigLoader) Load(_ map[string]any) error { return nil }
func (nopConfigLoader) Validate() error             { return nil }

type configStubCmd struct{ use string }

func (c *configStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		Config:  nopConfigLoader{},
	}
}

// recordingConfigLoader records the section it was given, for Build-level
// tests that assert on what actually got loaded.
type recordingConfigLoader struct {
	loadedSection map[string]any
}

func (r *recordingConfigLoader) Load(section map[string]any) error {
	r.loadedSection = section
	return nil
}

func (r *recordingConfigLoader) Validate() error { return nil }

type configLoaderStubCmd struct {
	use string
	cfg cmdconfig.ConfigLoader
}

func (c *configLoaderStubCmd) Define(_ *app.App) *cli.Definition {
	return &cli.Definition{
		Meta:    cli.Meta{Use: c.use},
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		Config:  c.cfg,
	}
}

func TestFactory_Build_flatMode_loadsInvokedCommandConfig(t *testing.T) {
	repoRoot := t.TempDir()
	ws, err := workspace.New(t.TempDir(), repoRoot)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	configDir := filepath.Join(repoRoot, config.AppName)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	manifest := "child:\n  name: leaf\n"
	if err := os.WriteFile(filepath.Join(configDir, "manifest.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("WriteFile manifest: %v", err)
	}
	a := app.NewBuilder().
		WithWorkspace(ws).
		WithInvocation(&invocation.Invocation{Raw: []string{"child"}}).
		Build()

	rec := &recordingConfigLoader{}
	parent := &parentStubCmd{use: "parent", children: []cli.Command{&configLoaderStubCmd{use: "child", cfg: rec}}}
	if _, err := factory.New().Build(parent, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.loadedSection["name"] != "leaf" {
		t.Errorf("loadedSection = %v, want name=leaf", rec.loadedSection)
	}
}

func TestFactory_Build_returns_error_when_invoked_config_unreadable(t *testing.T) {
	repoRoot := t.TempDir()
	ws, err := workspace.New(t.TempDir(), repoRoot)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	configDir := filepath.Join(repoRoot, config.AppName)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "child.yaml"), []byte("not: [valid: yaml"), 0o600); err != nil {
		t.Fatalf("WriteFile child: %v", err)
	}
	a := app.NewBuilder().
		WithWorkspace(ws).
		WithInvocation(&invocation.Invocation{Raw: []string{"child"}}).
		Build()

	parent := &parentStubCmd{use: "parent", children: []cli.Command{&configStubCmd{use: "child"}}}
	_, err = factory.New().Build(parent, a)
	if err == nil {
		t.Fatal("expected error when the invoked command's config file can't be decoded, got nil")
	}
}
