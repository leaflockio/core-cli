// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/spf13/cobra"
)

var errResolve = errors.New("resolve failed")

func TestExecute_builds_command_with_correct_metadata(t *testing.T) {
	plan := assembledPlan("mycmd", "short desc", "long desc")
	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Use != "mycmd" {
		t.Errorf("Use = %q, want %q", cmd.Use, "mycmd")
	}
	if cmd.Short != "short desc" {
		t.Errorf("Short = %q, want %q", cmd.Short, "short desc")
	}
	if cmd.Long != "long desc" {
		t.Errorf("Long = %q, want %q", cmd.Long, "long desc")
	}
}

func TestExecute_registers_flags_on_command(t *testing.T) {
	plan := assembledPlan("mycmd", "", "")
	plan.flags = append(plan.flags, mustPlanFlag(
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose")},
	))

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Flags().Lookup("verbose") == nil {
		t.Error("flag 'verbose' must be registered on the command")
	}
}

func TestExecute_registers_system_flag_that_fires_effect_on_parse(t *testing.T) {
	var effectCalled bool
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubExplicit,
		Value:  flags.Bool("test-sys", ""),
		Effect: func(_ *app.App) { effectCalled = true },
	}
	plan := assembledPlan("mycmd", "", "")
	plan.flags = []assembledFlag{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := cmd.Flags().Parse([]string{"--test-sys"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !effectCalled {
		t.Error("system flag effect must fire on Parse, before RunE ever runs")
	}
}

func TestExecute_registered_system_flag_does_not_fire_effect_when_not_set(t *testing.T) {
	var effectCalled bool
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubExplicit,
		Value:  flags.Bool("test-sys", ""),
		Effect: func(_ *app.App) { effectCalled = true },
	}
	plan := assembledPlan("mycmd", "", "")
	plan.flags = []assembledFlag{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.Flags().Parse([]string{}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if effectCalled {
		t.Error("system flag effect must not fire when flag is not set")
	}
}

func TestExecute_RunE_runs_resolver(t *testing.T) {
	r := &stubStringSliceResolver{}
	f := flags.CommandFlag[*flags.StringSliceValue]{
		Value:    flags.StringSlice("tags", ""),
		Resolver: r,
	}
	plan := assembledPlan("mycmd", "", "")
	plan.flags = []assembledFlag{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.Flags().Parse([]string{"--tags", "foo"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if !r.called {
		t.Error("resolver must be called during RunE")
	}
}

func TestExecute_RunE_returns_resolver_error(t *testing.T) {
	r := &errStringSliceResolver{err: errResolve}
	f := flags.CommandFlag[*flags.StringSliceValue]{
		Value:    flags.StringSlice("tags", ""),
		Resolver: r,
	}
	plan := assembledPlan("mycmd", "", "")
	plan.flags = []assembledFlag{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected error from resolver, got nil")
	}
}

func TestExecute_RunE_calls_handler(t *testing.T) {
	var handlerCalled bool
	plan := assembledPlan("mycmd", "", "")
	plan.def.Handler = func(_ *app.App, _ *cobra.Command, _ []string) error {
		handlerCalled = true
		return nil
	}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if !handlerCalled {
		t.Error("handler must be called during RunE")
	}
}

func TestExecute_RunE_passes_the_running_command_to_handler(t *testing.T) {
	var gotCmd *cobra.Command
	plan := assembledPlan("mycmd", "", "")
	plan.def.Handler = func(_ *app.App, cmd *cobra.Command, _ []string) error {
		gotCmd = cmd
		return nil
	}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if gotCmd != cmd {
		t.Error("handler must receive the same *cobra.Command that RunE was invoked with")
	}
}

func TestExecute_RunE_is_nil_when_handler_is_nil(t *testing.T) {
	plan := assembledPlan("parent", "", "")
	plan.def.Handler = nil
	plan.def.Children = []cli.Command{&stubCommand{use: "child"}}
	plan.hasChildren = true

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.RunE != nil {
		t.Error("RunE must be nil when Handler is nil — cobra's own non-Runnable fallback shows help; " +
			"system flag effects still fire at parse time regardless")
	}
	if cmd.Runnable() {
		t.Error("command must not be Runnable when Handler is nil")
	}
}

func TestExecute_adds_children_as_subcommands(t *testing.T) {
	child := &stubCommand{use: "child"}
	plan := assembledPlan("parent", "", "")
	plan.def.Children = []cli.Command{child}
	plan.hasChildren = true

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range cmd.Commands() {
		if sub.Use == "child" {
			return
		}
	}
	t.Error("child command must be registered as a subcommand")
}

func TestExecute_returns_error_when_child_build_fails(t *testing.T) {
	plan := assembledPlan("parent", "", "")
	plan.def.Children = []cli.Command{&nilHandlerCommand{}}
	plan.hasChildren = true

	_, err := (&Factory{}).execute(plan, nil)
	if err == nil {
		t.Fatal("expected error when child build fails, got nil")
	}
}

func TestExecute_non_root_plan_children_are_levelNested(t *testing.T) {
	plan := assembledPlan("license", "", "")
	plan.level = levelTop
	plan.def.Children = []cli.Command{&nilHandlerCommand{}}
	plan.hasChildren = true

	_, err := (&Factory{}).execute(plan, nil)
	if err == nil {
		t.Fatal("expected error: a non-root plan's child is levelNested, never exempt " +
			"from the Handler-or-Children guard")
	}
}

func TestExecute_does_not_register_persistentFlags_on_flagset(t *testing.T) {
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("no-color", ""),
		Effect: func(_ *app.App) {},
	}
	plan := assembledPlan("mycmd", "", "")
	plan.persistentFlags = []assembledFlag{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.PersistentFlags().Lookup("no-color") != nil {
		t.Error("execute must not register persistentFlags — only Build does, at the tree root")
	}
	if cmd.Flags().Lookup("no-color") != nil {
		t.Error("execute must not register persistentFlags on local Flags() either")
	}
	if cmd.PersistentPreRunE != nil {
		t.Error("execute must not set PersistentPreRunE — only Build does, at the tree root")
	}
}
