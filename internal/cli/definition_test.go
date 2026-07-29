// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/paths"
	"github.com/spf13/cobra"
)

// stubConfigLoader is a minimal cmdconfig.ConfigLoader for Definition.Config tests.
type stubConfigLoader struct{}

func (stubConfigLoader) Load(_ map[string]any) error { return nil }
func (stubConfigLoader) Validate() error             { return nil }

func TestNewDefinition_defaults(t *testing.T) {
	d := cli.NewDefinition(cli.NewMeta("foo", "short", ""))
	if d.Group != cli.GroupCLI {
		t.Errorf("Group = %q, want %q", d.Group, cli.GroupCLI)
	}
	if d.Meta.Use != "foo" {
		t.Errorf("Meta.Use = %q, want %q", d.Meta.Use, "foo")
	}
	if d.PathRegistry != nil {
		t.Error("PathRegistry should be nil by default")
	}
	if d.Config != nil {
		t.Error("Config should be nil by default")
	}
}

func TestDefinition_WithGroup(t *testing.T) {
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).WithGroup(cli.GroupProject)
	if d.Group != cli.GroupProject {
		t.Errorf("Group = %q, want %q", d.Group, cli.GroupProject)
	}
}

func TestDefinition_WithFlags(t *testing.T) {
	var v bool
	f := flags.CommandFlag[*flags.BoolValue]{
		Value: flags.Bool("verbose", "").WithShorthand("v").WithDest(&v),
	}
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).WithFlags([]flags.Flag{f})
	if len(d.Flags) != 1 {
		t.Errorf("Flags length = %d, want 1", len(d.Flags))
	}
}

func TestDefinition_WithPathRegistry(t *testing.T) {
	r := paths.NewRegistry()
	if err := r.Add(paths.KnownPath{Name: "manifest"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).WithPathRegistry(r)
	if d.PathRegistry != r {
		t.Error("PathRegistry not set correctly")
	}
}

func TestDefinition_WithConfig(t *testing.T) {
	c := stubConfigLoader{}
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).WithConfig(c)
	if d.Config != c {
		t.Error("Config not set correctly")
	}
}

func TestDefinition_WithHandler(t *testing.T) {
	handler := func(a *app.App, cmd *cobra.Command, args []string) error { return nil }
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).WithHandler(handler)
	if d.Handler == nil {
		t.Error("Handler should not be nil after WithHandler")
	}
}

func TestDefinition_WithChildren(t *testing.T) {
	child := &stubCommand{}
	d := cli.NewDefinition(cli.NewMeta("foo", "short", "")).
		WithChildren([]cli.Command{child})
	if len(d.Children) != 1 {
		t.Errorf("Children length = %d, want 1", len(d.Children))
	}
}

// stubCommand is a minimal Command implementation for testing.
type stubCommand struct{}

func (s *stubCommand) Define(_ *app.App) *cli.Definition {
	return cli.NewDefinition(cli.NewMeta("stub", "stub command", ""))
}
