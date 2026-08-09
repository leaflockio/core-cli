// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/leaflockio/core-cli/internal/paths"
)

// — assemble —

func TestAssemble_returns_error_when_handler_nil_and_no_children(t *testing.T) {
	def := cli.Definition{Meta: &cli.Meta{Use: "test"}}
	_, err := assemble(&def, level.LevelTop)
	if err == nil {
		t.Fatal("expected error for nil Handler with no children, got nil")
	}
	if !errors.Is(err, errHandlerNil) {
		t.Errorf("error = %v, want errors.Is match for errHandlerNil", err)
	}
}

func TestAssemble_returns_error_when_use_is_empty(t *testing.T) {
	def := cli.Definition{Meta: &cli.Meta{Use: ""}}
	_, err := assemble(&def, level.LevelRoot)
	if err == nil {
		t.Fatal("expected error for empty Use, got nil")
	}
	if !errors.Is(err, errUseEmpty) {
		t.Errorf("error = %v, want errors.Is match for errUseEmpty", err)
	}
}

func TestAssemble_returns_error_when_use_has_whitespace(t *testing.T) {
	def := cli.Definition{Meta: &cli.Meta{Use: "check [file...]"}}
	_, err := assemble(&def, level.LevelRoot)
	if err == nil {
		t.Fatal("expected error for Use containing whitespace, got nil")
	}
	if !errors.Is(err, errUseHasWhitespace) {
		t.Errorf("error = %v, want errors.Is match for errUseHasWhitespace", err)
	}
}

func TestAssemble_allows_nil_handler_and_no_children_when_levelRoot(t *testing.T) {
	def := cli.Definition{Meta: &cli.Meta{Use: "test"}}
	_, err := assemble(&def, level.LevelRoot)
	if err != nil {
		t.Fatalf("unexpected error for level.LevelRoot with nil Handler and no children: %v", err)
	}
}

func TestAssemble_allows_nil_handler_when_children_declared(t *testing.T) {
	def := &cli.Definition{
		Meta:     &cli.Meta{Use: "parent"},
		Children: []cli.Command{&stubCommand{use: "child"}},
	}
	_, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error for nil Handler with children declared: %v", err)
	}
}

func TestAssemble_allows_config_when_levelTop(t *testing.T) {
	def := minDef("license")
	def.Config = stubConfigLoader{}
	_, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error for Config declared at level.LevelTop: %v", err)
	}
}

func TestAssemble_allows_path_registry_when_levelTop(t *testing.T) {
	def := minDef("license")
	registry := paths.NewRegistry()
	if err := registry.Add(paths.KnownPath{Name: "license-config"}); err != nil {
		t.Fatalf("unexpected error adding path: %v", err)
	}
	def.PathRegistry = registry
	_, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error for PathRegistry declared at level.LevelTop: %v", err)
	}
}

func TestAssemble_returns_error_when_config_declared_below_levelTop(t *testing.T) {
	tests := []struct {
		name string
		lvl  level.Level
	}{
		{"root", level.LevelRoot},
		{"nested", level.LevelNested},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := minDef("test")
			def.Config = stubConfigLoader{}
			_, err := assemble(def, tt.lvl)
			if err == nil {
				t.Fatalf("expected error for Config declared at %v, got nil", tt.lvl)
			}
			if !errors.Is(err, errConfigNotTopLevel) {
				t.Errorf("error = %v, want errors.Is match for errConfigNotTopLevel", err)
			}
		})
	}
}

func TestAssemble_returns_error_when_path_registry_declared_below_levelTop(t *testing.T) {
	tests := []struct {
		name string
		lvl  level.Level
	}{
		{"root", level.LevelRoot},
		{"nested", level.LevelNested},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := minDef("test")
			registry := paths.NewRegistry()
			if err := registry.Add(paths.KnownPath{Name: "x"}); err != nil {
				t.Fatalf("unexpected error adding path: %v", err)
			}
			def.PathRegistry = registry
			_, err := assemble(def, tt.lvl)
			if err == nil {
				t.Fatalf("expected error for PathRegistry declared at %v, got nil", tt.lvl)
			}
		})
	}
}

func TestAssemble_returns_error_when_implicit_flag_in_definition(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.SystemFlag[*flags.BoolValue]{
			Sub:    flags.SubImplicit,
			Value:  flags.Bool("custom-implicit", ""),
			Effect: func(_ *app.App) {},
		},
	}
	_, err := assemble(def, level.LevelTop)
	if err == nil {
		t.Fatal("expected error for implicit flag in Definition.Flags, got nil")
	}
	if !errors.Is(err, errImplicitFlagInDef) {
		t.Errorf("error = %v, want errors.Is match for errImplicitFlagInDef", err)
	}
}

func TestAssemble_hasFlags_false_when_no_flags_declared(t *testing.T) {
	plan, err := assemble(minDef("test"), level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.hasFlags {
		t.Error("hasFlags must be false when Definition.Flags is empty")
	}
}

func TestAssemble_hasFlags_true_when_flags_declared(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "")},
	}
	plan, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.hasFlags {
		t.Error("hasFlags must be true when Definition.Flags is non-empty")
	}
}

func TestAssemble_flagRules_do_not_register_their_own_flags(t *testing.T) {
	visible := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("all", "")}
	hidden := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("staged", "")}

	def := minDef("test")
	def.Flags = []flags.Flag{visible}
	def.FlagRules = []flags.Rule{flags.Exclusive{Flags: []flags.Flag{visible, hidden}}}

	plan, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.flags) != 1 {
		t.Errorf("plan.flags has %d entries, want 1 — FlagRules must not register a flag on its own", len(plan.flags))
	}
}

func TestAssemble_stores_given_level_on_plan(t *testing.T) {
	for _, lvl := range []level.Level{level.LevelRoot, level.LevelTop, level.LevelNested} {
		plan, err := assemble(minDef("test"), lvl)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", lvl, err)
		}
		if plan.level != lvl {
			t.Errorf("level = %v, want %v", plan.level, lvl)
		}
	}
}

func TestAssemble_hasChildren_false_when_no_children(t *testing.T) {
	plan, err := assemble(minDef("test"), level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.hasChildren {
		t.Error("hasChildren must be false when Definition.Children is empty")
	}
}

func TestAssemble_hasChildren_true_when_children_declared(t *testing.T) {
	def := minDef("parent")
	def.Children = []cli.Command{&stubCommand{use: "child"}}
	plan, err := assemble(def, level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.hasChildren {
		t.Error("hasChildren must be true when Definition.Children is non-empty")
	}
}

func TestAssemble_implicit_flags_go_to_persistent_flags(t *testing.T) {
	plan, err := assemble(minDef("test"), level.LevelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.persistentFlags) < len(implicitSystemFlags) {
		t.Errorf("plan.persistentFlags has %d entries, want at least %d (implicit)",
			len(plan.persistentFlags), len(implicitSystemFlags))
	}
	if len(plan.flags) != 0 {
		t.Errorf("plan.flags has %d entries, want 0 when no def.Flags declared", len(plan.flags))
	}
}

func TestAssemble_returns_error_when_duplicate_flags(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("dup", "")},
		flags.CommandFlag[*flags.StringValue]{Value: flags.String("dup", "")},
	}
	_, err := assemble(def, level.LevelTop)
	if err == nil {
		t.Fatal("expected error for duplicate flag names, got nil")
	}
}

func TestAssemble_returns_error_when_flag_type_unrecognized(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{unknownFlag{}}
	_, err := assemble(def, level.LevelTop)
	if err == nil {
		t.Fatal("expected error for unrecognized flag type, got nil")
	}
}

func TestAssemble_returns_error_when_implicit_flag_planning_fails(t *testing.T) {
	original := implicitSystemFlags
	implicitSystemFlags = []flags.Flag{unknownFlag{}}
	defer func() { implicitSystemFlags = original }()

	_, err := assemble(minDef("test"), level.LevelTop)
	if err == nil {
		t.Fatal("expected error when implicit flag planning fails, got nil")
	}
}

// — checkHandlerOrChildren —

func TestCheckHandlerOrChildren_nilAtLevelRoot(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}}
	if err := checkHandlerOrChildren(def, level.LevelRoot); err != nil {
		t.Errorf("unexpected error at LevelRoot: %v", err)
	}
}

func TestCheckHandlerOrChildren_nilWhenHandlerSet(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}, Handler: nopHandler}
	if err := checkHandlerOrChildren(def, level.LevelTop); err != nil {
		t.Errorf("unexpected error when Handler is set: %v", err)
	}
}

func TestCheckHandlerOrChildren_nilWhenChildrenDeclared(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}, Children: []cli.Command{&stubCommand{use: "child"}}}
	if err := checkHandlerOrChildren(def, level.LevelTop); err != nil {
		t.Errorf("unexpected error when Children is non-empty: %v", err)
	}
}

func TestCheckHandlerOrChildren_errorsWhenNeitherSet(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}}
	err := checkHandlerOrChildren(def, level.LevelTop)
	if err == nil {
		t.Fatal("expected error for nil Handler with no children, got nil")
	}
	if !errors.Is(err, errHandlerNil) {
		t.Errorf("error = %v, want errors.Is match for errHandlerNil", err)
	}
}

// — checkConfigOwner —

func TestCheckConfigOwner_nilAtLevelTop(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}, Config: stubConfigLoader{}}
	if err := checkConfigOwner(def, level.LevelTop); err != nil {
		t.Errorf("unexpected error at LevelTop: %v", err)
	}
}

func TestCheckConfigOwner_nilWhenNeitherDeclared(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}}
	if err := checkConfigOwner(def, level.LevelRoot); err != nil {
		t.Errorf("unexpected error when neither Config nor PathRegistry is declared: %v", err)
	}
}

func TestCheckConfigOwner_errorsForConfigBelowLevelTop(t *testing.T) {
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}, Config: stubConfigLoader{}}
	err := checkConfigOwner(def, level.LevelRoot)
	if err == nil {
		t.Fatal("expected error for Config declared below LevelTop, got nil")
	}
	if !errors.Is(err, errConfigNotTopLevel) {
		t.Errorf("error = %v, want errors.Is match for errConfigNotTopLevel", err)
	}
}

func TestCheckConfigOwner_errorsForPathRegistryBelowLevelTop(t *testing.T) {
	registry := paths.NewRegistry()
	if err := registry.Add(paths.KnownPath{Name: "x"}); err != nil {
		t.Fatalf("unexpected error adding path: %v", err)
	}
	def := &cli.Definition{Meta: &cli.Meta{Use: "test"}, PathRegistry: registry}
	if err := checkConfigOwner(def, level.LevelNested); err == nil {
		t.Fatal("expected error for PathRegistry declared below LevelTop, got nil")
	}
}

// — checkFlagsNotImplicit —

func TestCheckFlagsNotImplicit_nilWhenNoFlags(t *testing.T) {
	def := minDef("test")
	if err := checkFlagsNotImplicit(def, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckFlagsNotImplicit_nilWhenAllCommandFlags(t *testing.T) {
	def := minDef("test")
	all := []flags.Flag{flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "")}}
	if err := checkFlagsNotImplicit(def, all); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckFlagsNotImplicit_errorsForImplicitFlag(t *testing.T) {
	def := minDef("test")
	all := []flags.Flag{
		flags.SystemFlag[*flags.BoolValue]{
			Sub:    flags.SubImplicit,
			Value:  flags.Bool("custom-implicit", ""),
			Effect: func(_ *app.App) {},
		},
	}
	err := checkFlagsNotImplicit(def, all)
	if err == nil {
		t.Fatal("expected error for implicit flag, got nil")
	}
	if !errors.Is(err, errImplicitFlagInDef) {
		t.Errorf("error = %v, want errors.Is match for errImplicitFlagInDef", err)
	}
}

// — validateMeta —

func TestValidateMeta_returns_nil_for_bare_use(t *testing.T) {
	if err := validateMeta(&cli.Meta{Use: "check"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMeta_returns_error_for_empty_use(t *testing.T) {
	if err := validateMeta(&cli.Meta{Use: ""}); !errors.Is(err, errUseEmpty) {
		t.Errorf("error = %v, want errors.Is match for errUseEmpty", err)
	}
}

func TestValidateMeta_returns_error_for_use_with_space(t *testing.T) {
	if err := validateMeta(&cli.Meta{Use: "check [file...]"}); !errors.Is(err, errUseHasWhitespace) {
		t.Errorf("error = %v, want errors.Is match for errUseHasWhitespace", err)
	}
}

func TestValidateMeta_returns_error_for_use_with_tab(t *testing.T) {
	if err := validateMeta(&cli.Meta{Use: "check\tfile"}); !errors.Is(err, errUseHasWhitespace) {
		t.Errorf("error = %v, want errors.Is match for errUseHasWhitespace", err)
	}
}

func TestValidateMeta_allows_argsUsage_set_separately(t *testing.T) {
	if err := validateMeta(&cli.Meta{Use: "check", ArgsUsage: "[file...]"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// — validateNoDuplicateFlags —

func TestValidateNoDuplicateFlags_returns_nil_when_no_duplicates(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "").WithShorthand("v")},
		flags.CommandFlag[*flags.StringValue]{Value: flags.String("output", "").WithShorthand("o")},
	}
	if err := validateNoDuplicateFlags(def); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNoDuplicateFlags_returns_error_for_duplicate_name(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "")},
		flags.CommandFlag[*flags.StringValue]{Value: flags.String("verbose", "")},
	}
	err := validateNoDuplicateFlags(def)
	if err == nil {
		t.Fatal("expected error for duplicate flag name, got nil")
	}
	if cause := causeOf(t, err); !strings.Contains(cause, `"verbose" is declared`) {
		t.Errorf("cause = %q, want to contain duplicate name report", cause)
	}
}

func TestValidateNoDuplicateFlags_returns_error_for_duplicate_shorthand(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "").WithShorthand("v")},
		flags.CommandFlag[*flags.StringValue]{Value: flags.String("version", "").WithShorthand("v")},
	}
	err := validateNoDuplicateFlags(def)
	if err == nil {
		t.Fatal("expected error for duplicate shorthand, got nil")
	}
	if cause := causeOf(t, err); !strings.Contains(cause, `shorthand "v"`) {
		t.Errorf("cause = %q, want to contain shorthand report", cause)
	}
}

func TestValidateNoDuplicateFlags_reports_all_violations(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("flag", "").WithShorthand("f")},
		flags.CommandFlag[*flags.StringValue]{Value: flags.String("flag", "").WithShorthand("f")},
	}
	err := validateNoDuplicateFlags(def)
	if err == nil {
		t.Fatal("expected error for multiple violations, got nil")
	}
	cause := causeOf(t, err)
	if !strings.Contains(cause, `"flag" is declared`) {
		t.Errorf("cause = %q, missing duplicate name report", cause)
	}
	if !strings.Contains(cause, `shorthand "f"`) {
		t.Errorf("cause = %q, missing shorthand report", cause)
	}
}

func TestValidateNoDuplicateFlags_detects_clash_with_implicit_flags(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{
		flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("no-color", "")},
	}
	err := validateNoDuplicateFlags(def)
	if err == nil {
		t.Fatal("expected error for clash with implicit no-color flag, got nil")
	}
}
