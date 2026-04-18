// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"bytes"
	"testing"

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/terminal"
	"github.com/leaflock/core-cli/internal/ui"
	"github.com/leaflock/core-cli/internal/util/strutil"
	"github.com/leaflock/core-cli/internal/version"
)

func newTestApp(t *testing.T) *app.App {
	t.Helper()
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	return app.NewBuilder().
		WithPrinter(printer).
		WithVersion(&version.Info{}).
		Build()
}

func TestRoot_use(t *testing.T) {
	cmd := (&root{app: newTestApp(t)}).cmd()
	if cmd.Use != "leaf" {
		t.Errorf("expected Use %q, got %q", "leaf", cmd.Use)
	}
}

func TestRoot_short(t *testing.T) {
	cmd := (&root{app: newTestApp(t)}).cmd()
	if cmd.Short == "" {
		t.Error("expected non-empty Short")
	}
}

func TestRoot_silenced(t *testing.T) {
	cmd := (&root{app: newTestApp(t)}).baseCmd(&rootFlags{})
	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage=true")
	}
	if !cmd.SilenceErrors {
		t.Error("expected SilenceErrors=true")
	}
}

func TestRoot_addGroups_registersAll(t *testing.T) {
	r := &root{app: newTestApp(t)}
	cmd := r.baseCmd(&rootFlags{})
	r.addGroups(cmd)

	ids := make(map[string]bool)
	for _, g := range cmd.Groups() {
		ids[g.ID] = true
	}
	for _, want := range []string{groupGeneral, groupTools} {
		if !ids[want] {
			t.Errorf("expected group %q to be registered", want)
		}
	}
}

func TestRoot_addGroups_titlesFromIDs(t *testing.T) {
	r := &root{app: newTestApp(t)}
	cmd := r.baseCmd(&rootFlags{})
	r.addGroups(cmd)

	for _, g := range cmd.Groups() {
		want := strutil.Title(g.ID)
		if g.Title != want {
			t.Errorf("group %q: expected Title %q, got %q", g.ID, want, g.Title)
		}
	}
}

func TestRoot_addCommands_registersAll(t *testing.T) {
	r := &root{app: newTestApp(t)}
	cmd := r.baseCmd(&rootFlags{})
	r.addCommands(cmd)

	names := make(map[string]bool)
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"version", "license"} {
		if !names[want] {
			t.Errorf("expected subcommand %q to be registered", want)
		}
	}
}

func TestRoot_addCommands_setsGroupID(t *testing.T) {
	r := &root{app: newTestApp(t)}
	cmd := r.baseCmd(&rootFlags{})
	r.addCommands(cmd)

	for _, c := range cmd.Commands() {
		if c.GroupID == "" {
			t.Errorf("command %q has empty GroupID", c.Name())
		}
	}
}

func TestRoot_runE_help(t *testing.T) {
	cmd := (&root{app: newTestApp(t)}).cmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Errorf("expected nil error from RunE, got %v", err)
	}
}

func TestRoot_persistentPreRun_noColor(t *testing.T) {
	a := newTestApp(t)
	cmd := (&root{app: a}).cmd()
	cmd.SetArgs([]string{"--no-color", "version"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	// After --no-color, Primary should return plain text with no styling applied.
	if got := a.Printer.Primary("x"); got != "x" {
		t.Errorf("expected plain text after --no-color, got %q", got)
	}
}
