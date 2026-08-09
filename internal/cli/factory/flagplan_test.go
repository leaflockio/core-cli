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
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/spf13/pflag"
)

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
	if !errors.Is(err, errInvalidFlagType) {
		t.Errorf("error = %v, want errors.Is match for errInvalidFlagType", err)
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
