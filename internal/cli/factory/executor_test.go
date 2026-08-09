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
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/spf13/cobra"
)

var errResolve = errors.New("resolve failed")

func TestExecute_builds_command_with_correct_metadata(t *testing.T) {
	plan := minBlueprint("mycmd", "short desc", "long desc")
	cmd, err := (&Factory{}).execute(plan, nil, nil)
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

// TestExecute_appendsArgsUsageToCommandUse is a regression test: the real
// cobra command's Use must combine Meta.Use with Meta.ArgsUsage, while
// Meta.Use itself (read everywhere else in the framework for command
// identity) stays just the bare name.
func TestExecute_appendsArgsUsageToCommandUse(t *testing.T) {
	plan := &blueprint{
		def: cli.Definition{
			Meta:    &cli.Meta{Use: "check", ArgsUsage: "[file...]"},
			Handler: func(_ *app.App, _ *cobra.Command, _ []string) error { return nil },
		},
	}
	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Use != "check [file...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "check [file...]")
	}
}

// — cobraUse —

func TestCobraUse_returnsBareUseWhenNoArgsUsage(t *testing.T) {
	got := cobraUse(&cli.Meta{Use: "check"})
	if got != "check" {
		t.Errorf("cobraUse = %q, want %q", got, "check")
	}
}

func TestCobraUse_appendsArgsUsage(t *testing.T) {
	got := cobraUse(&cli.Meta{Use: "check", ArgsUsage: "[file...]"})
	if got != "check [file...]" {
		t.Errorf("cobraUse = %q, want %q", got, "check [file...]")
	}
}

// — buildRunE —

func TestBuildRunE_rendersHelp_whenHandlerNil(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "parent"}}
	runE := buildRunE(def, &blueprint{}, nil)

	cmd := &cobra.Command{Use: "parent"}
	if err := runE(cmd, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildRunE_callsHandler_whenHandlerSet(t *testing.T) {
	var handlerCalled bool
	def := &cli.Definition{
		Meta: &cli.Meta{Use: "mycmd"},
		Handler: func(_ *app.App, _ *cobra.Command, _ []string) error {
			handlerCalled = true
			return nil
		},
	}
	runE := buildRunE(def, &blueprint{}, nil)

	cmd := &cobra.Command{Use: "mycmd"}
	if err := runE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler must be called when Handler is set")
	}
}

func TestExecute_RunE_rejects_exclusive_flags_set_together(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.FlagRules = []flags.Rule{flags.Exclusive{Flags: boolFlags("all", "staged")}}
	plan.flags = []flagSpec{
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("all", "")}),
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("staged", "")}),
	}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.Flags().Parse([]string{"--all", "--staged"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Error("expected RunE to reject two flags from the same exclusive group")
	}
}

func TestExecute_RunE_does_not_call_handler_when_exclusive_flags_conflict(t *testing.T) {
	var handlerCalled bool
	plan := minBlueprint("mycmd", "", "")
	plan.def.Handler = func(_ *app.App, _ *cobra.Command, _ []string) error {
		handlerCalled = true
		return nil
	}
	plan.def.FlagRules = []flags.Rule{flags.Exclusive{Flags: boolFlags("all", "staged")}}
	plan.flags = []flagSpec{
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("all", "")}),
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("staged", "")}),
	}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.Flags().Parse([]string{"--all", "--staged"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	_ = cmd.RunE(cmd, nil)
	if handlerCalled {
		t.Error("Handler must not be called when exclusive flags conflict")
	}
}

func TestExecute_RunE_allows_single_exclusive_flag(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.FlagRules = []flags.Rule{flags.Exclusive{Flags: boolFlags("all", "staged")}}
	plan.flags = []flagSpec{
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("all", "")}),
		mustPlanFlag(flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("staged", "")}),
	}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.Flags().Parse([]string{"--all"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("unexpected error for a single exclusive flag: %v", err)
	}
}

func TestExecute_registers_declared_groups_on_command(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.Groups = []cli.Group{{ID: "read", Title: "read"}, {ID: "write", Title: "write"}}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmd.ContainsGroup("read") {
		t.Error("expected group \"read\" to be registered on the command")
	}
	if !cmd.ContainsGroup("write") {
		t.Error("expected group \"write\" to be registered on the command")
	}
}

func TestExecute_normalizes_group_title_to_upper_case(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.Groups = []cli.Group{{ID: "read", Title: "read ops"}}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, g := range cmd.Groups() {
		if g.ID == "read" && g.Title != "READ OPS" {
			t.Errorf("Title = %q, want %q", g.Title, "READ OPS")
		}
	}
}

func TestExecute_registers_no_groups_when_none_declared(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmd.Groups()) != 0 {
		t.Errorf("Groups() = %v, want none registered when Definition.Groups is empty", cmd.Groups())
	}
}

func TestExecute_sets_groupID_from_def_group(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.Group = "read"

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.GroupID != "read" {
		t.Errorf("GroupID = %q, want %q", cmd.GroupID, "read")
	}
}

func TestExecute_propagates_disableSuggestions_to_command(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.def.Meta.DisableSuggestions = true

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmd.DisableSuggestions {
		t.Error("DisableSuggestions = false, want true to be propagated from Meta")
	}
}

func TestExecute_registers_flags_on_command(t *testing.T) {
	plan := minBlueprint("mycmd", "", "")
	plan.flags = append(plan.flags, mustPlanFlag(
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "enable verbose")},
	))

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.flags = []flagSpec{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.flags = []flagSpec{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.flags = []flagSpec{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.flags = []flagSpec{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected error from resolver, got nil")
	}
}

func TestExecute_RunE_calls_handler(t *testing.T) {
	var handlerCalled bool
	plan := minBlueprint("mycmd", "", "")
	plan.def.Handler = func(_ *app.App, _ *cobra.Command, _ []string) error {
		handlerCalled = true
		return nil
	}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.def.Handler = func(_ *app.App, cmd *cobra.Command, _ []string) error {
		gotCmd = cmd
		return nil
	}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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

func TestExecute_RunE_renders_help_when_handler_is_nil(t *testing.T) {
	plan := minBlueprint("parent", "", "")
	plan.def.Handler = nil
	plan.def.Children = []cli.Command{&stubCommand{use: "child"}}
	plan.hasChildren = true

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.RunE == nil {
		t.Fatal("RunE must be set when Handler is nil, so cobra's ValidateArgs runs for this command too, " +
			"instead of being skipped by cobra's non-Runnable fallback")
	}
	if !cmd.Runnable() {
		t.Error("command must be Runnable when Handler is nil")
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("RunE must render help and return nil when called with no args: %v", err)
	}
}

func TestExecute_RunE_rejects_unknown_subcommand_when_handler_is_nil(t *testing.T) {
	plan := minBlueprint("parent", "", "")
	plan.def.Handler = nil
	plan.def.Meta.Args = cli.NewMeta("parent", "", "").Args
	plan.def.Children = []cli.Command{&stubCommand{use: "child"}}
	plan.hasChildren = true

	cmd, err := (&Factory{}).execute(plan, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cmd.ValidateArgs([]string{"bogus"}); err == nil {
		t.Error("ValidateArgs must reject an unrecognized subcommand name now that the command is Runnable")
	}
}

func TestExecute_adds_children_as_subcommands(t *testing.T) {
	child := &stubCommand{use: "child"}
	plan := minBlueprint("parent", "", "")
	plan.def.Children = []cli.Command{child}
	plan.hasChildren = true

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("parent", "", "")
	plan.def.Children = []cli.Command{&nilHandlerCommand{}}
	plan.hasChildren = true

	_, err := (&Factory{}).execute(plan, nil, nil)
	if err == nil {
		t.Fatal("expected error when child build fails, got nil")
	}
}

func TestExecute_non_root_plan_children_are_levelNested(t *testing.T) {
	plan := minBlueprint("license", "", "")
	plan.level = level.LevelTop
	plan.def.Children = []cli.Command{&nilHandlerCommand{}}
	plan.hasChildren = true

	_, err := (&Factory{}).execute(plan, nil, nil)
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
	plan := minBlueprint("mycmd", "", "")
	plan.persistentFlags = []flagSpec{mustPlanFlag(f)}

	cmd, err := (&Factory{}).execute(plan, nil, nil)
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
