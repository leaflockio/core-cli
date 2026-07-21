// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"testing"

	"github.com/spf13/pflag"
)

// — effectValue.Set —

func TestEffectValue_Set_fires_effect_after_successful_set(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")
	inner := fs.Lookup("flag").Value

	var called bool
	v := &effectValue{Value: inner, effect: func() { called = true }}

	if err := v.Set("true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("effect must fire after a successful Set")
	}
}

func TestEffectValue_Set_does_not_fire_effect_when_underlying_set_fails(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")
	inner := fs.Lookup("flag").Value

	var called bool
	v := &effectValue{Value: inner, effect: func() { called = true }}

	if err := v.Set("not-a-bool"); err == nil {
		t.Fatal("expected error from underlying Set, got nil")
	}
	if called {
		t.Error("effect must not fire when underlying Set fails")
	}
}

// — delegated methods —

func TestEffectValue_String_delegates_to_wrapped_value(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", true, "")
	inner := fs.Lookup("flag").Value

	v := &effectValue{Value: inner}
	if v.String() != inner.String() {
		t.Errorf("String() = %q, want %q", v.String(), inner.String())
	}
}

func TestEffectValue_Type_delegates_to_wrapped_value(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")
	inner := fs.Lookup("flag").Value

	v := &effectValue{Value: inner}
	if v.Type() != "bool" {
		t.Errorf("Type() = %q, want %q", v.Type(), "bool")
	}
}

// — IsBoolFlag —

func TestEffectValue_IsBoolFlag_true_when_wrapped_value_is_bool(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")
	inner := fs.Lookup("flag").Value

	v := &effectValue{Value: inner}
	if !v.IsBoolFlag() {
		t.Error("IsBoolFlag() must be true when the wrapped Value is a bool flag")
	}
}

func TestEffectValue_IsBoolFlag_false_when_wrapped_value_lacks_interface(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("flag", "", "")
	inner := fs.Lookup("flag").Value

	v := &effectValue{Value: inner}
	if v.IsBoolFlag() {
		t.Error("IsBoolFlag() must be false when the wrapped Value does not implement it")
	}
}

// — wrapWithEffect —

func TestWrapWithEffect_fires_effect_on_parse(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")

	var called bool
	wrapWithEffect(fs, "flag", func() { called = true })

	if err := fs.Parse([]string{"--flag"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !called {
		t.Error("effect must fire when the wrapped flag is parsed")
	}
}

func TestWrapWithEffect_does_not_fire_when_flag_not_passed(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")

	var called bool
	wrapWithEffect(fs, "flag", func() { called = true })

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if called {
		t.Error("effect must not fire when the flag is never set")
	}
}

func TestWrapWithEffect_preserves_GetBool(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("flag", false, "")
	wrapWithEffect(fs, "flag", func() {})

	if err := fs.Parse([]string{"--flag"}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	got, err := fs.GetBool("flag")
	if err != nil {
		t.Fatalf("GetBool failed: %v", err)
	}
	if !got {
		t.Error("GetBool must reflect the parsed value through the wrapped Value")
	}
}
