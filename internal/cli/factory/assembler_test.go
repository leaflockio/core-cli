// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/spf13/pflag"
)

// — assemble —

func TestAssemble_returns_error_when_handler_nil_and_no_children(t *testing.T) {
	def := cli.Definition{Meta: cli.Meta{Use: "test"}}
	_, err := assemble(&def, levelTop)
	if err == nil {
		t.Fatal("expected error for nil Handler with no children, got nil")
	}
	if !strings.Contains(err.Error(), "handler must not be nil") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "handler must not be nil")
	}
}

func TestAssemble_allows_nil_handler_and_no_children_when_levelRoot(t *testing.T) {
	def := cli.Definition{Meta: cli.Meta{Use: "test"}}
	_, err := assemble(&def, levelRoot)
	if err != nil {
		t.Fatalf("unexpected error for levelRoot with nil Handler and no children: %v", err)
	}
}

func TestAssemble_allows_nil_handler_when_children_declared(t *testing.T) {
	def := &cli.Definition{
		Meta:     cli.Meta{Use: "parent"},
		Children: []cli.Command{&stubCommand{use: "child"}},
	}
	_, err := assemble(def, levelTop)
	if err != nil {
		t.Fatalf("unexpected error for nil Handler with children declared: %v", err)
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
	_, err := assemble(def, levelTop)
	if err == nil {
		t.Fatal("expected error for implicit flag in Definition.Flags, got nil")
	}
	if !strings.Contains(err.Error(), "must not appear in Definition.Flags") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "must not appear in Definition.Flags")
	}
}

func TestAssemble_hasFlags_false_when_no_flags_declared(t *testing.T) {
	plan, err := assemble(minDef("test"), levelTop)
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
	plan, err := assemble(def, levelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.hasFlags {
		t.Error("hasFlags must be true when Definition.Flags is non-empty")
	}
}

func TestAssemble_stores_given_level_on_plan(t *testing.T) {
	for _, lvl := range []level{levelRoot, levelTop, levelNested} {
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
	plan, err := assemble(minDef("test"), levelTop)
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
	plan, err := assemble(def, levelTop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.hasChildren {
		t.Error("hasChildren must be true when Definition.Children is non-empty")
	}
}

func TestAssemble_implicit_flags_go_to_persistent_flags(t *testing.T) {
	plan, err := assemble(minDef("test"), levelTop)
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
	_, err := assemble(def, levelTop)
	if err == nil {
		t.Fatal("expected error for duplicate flag names, got nil")
	}
}

func TestAssemble_returns_error_when_flag_type_unrecognized(t *testing.T) {
	def := minDef("test")
	def.Flags = []flags.Flag{unknownFlag{}}
	_, err := assemble(def, levelTop)
	if err == nil {
		t.Fatal("expected error for unrecognized flag type, got nil")
	}
}

func TestAssemble_returns_error_when_implicit_flag_planning_fails(t *testing.T) {
	original := implicitSystemFlags
	implicitSystemFlags = []flags.Flag{unknownFlag{}}
	defer func() { implicitSystemFlags = original }()

	_, err := assemble(minDef("test"), levelTop)
	if err == nil {
		t.Fatal("expected error when implicit flag planning fails, got nil")
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
	if !strings.Contains(err.Error(), `"verbose" is declared`) {
		t.Errorf("error = %q, want to contain duplicate name report", err.Error())
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
	if !strings.Contains(err.Error(), `shorthand "v"`) {
		t.Errorf("error = %q, want to contain shorthand report", err.Error())
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
	if !strings.Contains(err.Error(), `"flag" is declared`) {
		t.Errorf("error = %q, missing duplicate name report", err.Error())
	}
	if !strings.Contains(err.Error(), `shorthand "f"`) {
		t.Errorf("error = %q, missing shorthand report", err.Error())
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

// — planFlag routing —

func TestPlanFlag_routes_bool_system_flag(t *testing.T) {
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("test", ""),
		Effect: func(_ *app.App) {},
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindSystem {
		t.Errorf("kind = %q, want KindSystem", p.kind)
	}
}

func TestPlanFlag_routes_string_system_flag(t *testing.T) {
	f := flags.SystemFlag[*flags.StringValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.String("test", ""),
		Effect: func(_ *app.App) {},
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindSystem {
		t.Errorf("kind = %q, want KindSystem", p.kind)
	}
}

func TestPlanFlag_routes_string_slice_system_flag(t *testing.T) {
	f := flags.SystemFlag[*flags.StringSliceValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.StringSlice("test", ""),
		Effect: func(_ *app.App) {},
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindSystem {
		t.Errorf("kind = %q, want KindSystem", p.kind)
	}
}

func TestPlanFlag_routes_bool_literal_flag(t *testing.T) {
	f := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("verbose", "")}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindCommand {
		t.Errorf("kind = %q, want KindCommand", p.kind)
	}
	if p.hasResolver {
		t.Error("hasResolver must be false for a literal flag")
	}
}

func TestPlanFlag_routes_string_literal_flag(t *testing.T) {
	f := flags.CommandFlag[*flags.StringValue]{Value: flags.String("output", "")}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindCommand {
		t.Errorf("kind = %q, want KindCommand", p.kind)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if fs.Lookup("output") == nil {
		t.Error("register must register the \"output\" flag on the FlagSet")
	}
}

func TestPlanFlag_routes_string_slice_literal_flag(t *testing.T) {
	f := flags.CommandFlag[*flags.StringSliceValue]{Value: flags.StringSlice("tags", "")}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != flags.KindCommand {
		t.Errorf("kind = %q, want KindCommand", p.kind)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if fs.Lookup("tags") == nil {
		t.Error("register must register the \"tags\" flag on the FlagSet")
	}
}

func TestPlanFlag_routes_string_slice_resolver_flag(t *testing.T) {
	r := &stubStringSliceResolver{}
	f := flags.CommandFlag[*flags.StringSliceValue]{
		Value:    flags.StringSlice("var", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.hasResolver {
		t.Error("hasResolver must be true for a resolver flag")
	}
}

func TestPlanFlag_routes_bool_resolver_flag(t *testing.T) {
	r := &stubBoolResolver{}
	f := flags.CommandFlag[*flags.BoolValue]{
		Value:    flags.Bool("dry-run", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.hasResolver {
		t.Error("hasResolver must be true for a bool resolver flag")
	}
}

func TestPlanFlag_routes_string_resolver_flag(t *testing.T) {
	r := &stubStringResolver{}
	f := flags.CommandFlag[*flags.StringValue]{
		Value:    flags.String("format", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.hasResolver {
		t.Error("hasResolver must be true for a string resolver flag")
	}
}

func TestPlanFlag_returns_error_for_unrecognized_type(t *testing.T) {
	_, err := planFlag(unknownFlag{})
	if err == nil {
		t.Fatal("expected error for unrecognized flag type, got nil")
	}
	if !strings.Contains(err.Error(), "unrecognized type") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unrecognized type")
	}
}

// — system flag effect —
//
// System flag effects fire immediately when pflag parses the flag (via the
// effectValue wrapper — see effectvalue.go), not via a separate post-parse
// check. So these tests assert on state after Parse alone.

func TestPlanBoolSystemFlag_effects_do_not_cross_trigger_between_flags(t *testing.T) {
	var aCalled, bCalled bool
	fa := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("flag-a", ""),
		Effect: func(_ *app.App) { aCalled = true },
	}
	fb := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("flag-b", ""),
		Effect: func(_ *app.App) { bCalled = true },
	}
	pa, err := planFlag(fa)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pb, err := planFlag(fb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	pa.register(fs, nil)
	pb.register(fs, nil)

	if err := fs.Parse([]string{"--flag-a"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !aCalled {
		t.Error("flag-a's effect must fire when only flag-a is set")
	}
	if bCalled {
		t.Error("flag-b's effect must not fire when only flag-a is set")
	}
}

func TestPlanBoolSystemFlag_effect_fires_when_flag_is_changed(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("test-flag", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--test-flag"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !called {
		t.Error("effect must fire when flag is explicitly set")
	}
}

func TestPlanBoolSystemFlag_effect_does_not_fire_when_flag_not_set(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.Bool("test-flag", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if called {
		t.Error("effect must not fire when flag is not set")
	}
}

// — string system flag effect —

func TestPlanStringSystemFlag_effect_fires_when_flag_is_changed(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.StringValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.String("test-str", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--test-str", "val"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !called {
		t.Error("effect must fire when string system flag is explicitly set")
	}
}

func TestPlanStringSystemFlag_effect_does_not_fire_when_flag_not_set(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.StringValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.String("test-str", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	_ = fs.Parse([]string{})

	if called {
		t.Error("effect must not fire when string system flag is not set")
	}
}

// — string slice system flag effect —

func TestPlanStringSliceSystemFlag_effect_fires_when_flag_is_changed(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.StringSliceValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.StringSlice("test-slice", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--test-slice", "a"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !called {
		t.Error("effect must fire when string slice system flag is explicitly set")
	}
}

func TestPlanStringSliceSystemFlag_effect_does_not_fire_when_flag_not_set(t *testing.T) {
	var called bool
	f := flags.SystemFlag[*flags.StringSliceValue]{
		Sub:    flags.SubImplicit,
		Value:  flags.StringSlice("test-slice", ""),
		Effect: func(_ *app.App) { called = true },
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	_ = fs.Parse([]string{})

	if called {
		t.Error("effect must not fire when string slice system flag is not set")
	}
}

// — resolver flag resolve —

func TestPlanBoolResolverFlag_resolve_calls_resolver(t *testing.T) {
	r := &stubBoolResolver{}
	f := flags.CommandFlag[*flags.BoolValue]{
		Value:    flags.Bool("dry-run", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--dry-run"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := p.resolve(fs); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if !r.called {
		t.Error("Resolve must be called on the bool resolver")
	}
	if !r.got {
		t.Error("resolver received false, want true")
	}
}

func TestPlanStringResolverFlag_resolve_calls_resolver(t *testing.T) {
	r := &stubStringResolver{}
	f := flags.CommandFlag[*flags.StringValue]{
		Value:    flags.String("format", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--format", "json"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := p.resolve(fs); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if !r.called {
		t.Error("Resolve must be called on the string resolver")
	}
	if r.got != "json" {
		t.Errorf("resolver received %q, want %q", r.got, "json")
	}
}

func TestPlanStringSliceResolverFlag_resolve_calls_resolver(t *testing.T) {
	r := &stubStringSliceResolver{}
	f := flags.CommandFlag[*flags.StringSliceValue]{
		Value:    flags.StringSlice("var", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p.register(fs, nil)
	if err := fs.Parse([]string{"--var", "a", "--var", "b"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := p.resolve(fs); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if !r.called {
		t.Error("Resolve must be called on the resolver")
	}
	if len(r.got) != 2 || r.got[0] != "a" || r.got[1] != "b" {
		t.Errorf("resolver received %v, want [a b]", r.got)
	}
}

func TestPlanBoolResolverFlag_resolve_returns_error_when_flag_not_registered(t *testing.T) {
	r := &stubBoolResolver{}
	f := flags.CommandFlag[*flags.BoolValue]{
		Value:    flags.Bool("dry-run", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if err := p.resolve(fs); err == nil {
		t.Fatal("expected error when flag not registered in FlagSet, got nil")
	}
}

func TestPlanStringResolverFlag_resolve_returns_error_when_flag_not_registered(t *testing.T) {
	r := &stubStringResolver{}
	f := flags.CommandFlag[*flags.StringValue]{
		Value:    flags.String("format", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if err := p.resolve(fs); err == nil {
		t.Fatal("expected error when flag not registered in FlagSet, got nil")
	}
}

func TestPlanStringSliceResolverFlag_resolve_returns_error_when_flag_not_registered(t *testing.T) {
	r := &stubStringSliceResolver{}
	f := flags.CommandFlag[*flags.StringSliceValue]{
		Value:    flags.StringSlice("var", ""),
		Resolver: r,
	}
	p, err := planFlag(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resolve against an empty FlagSet — flag is not registered, GetStringArray will fail.
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if err := p.resolve(fs); err == nil {
		t.Fatal("expected error when flag not registered in FlagSet, got nil")
	}
}
