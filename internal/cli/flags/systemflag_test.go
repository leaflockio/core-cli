// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
)

func TestSystemFlag_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = flags.SystemFlag[*flags.BoolValue]{}
}

func TestSystemFlag_Definition_not_nil(t *testing.T) {
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:   flags.SubImplicit,
		Value: flags.Bool("no-color", "disable color output"),
	}
	if f.Definition() == nil {
		t.Fatal("Definition() must not return nil")
	}
}

func TestSystemFlag_Definition_kind_is_system(t *testing.T) {
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:   flags.SubImplicit,
		Value: flags.Bool("no-color", "disable color output"),
	}
	if got := f.Definition().Meta.Kind; got != flags.KindSystem {
		t.Errorf("Meta.Kind = %q, want %q", got, flags.KindSystem)
	}
}

func TestSystemFlag_Definition_carries_sub(t *testing.T) {
	implicit := flags.SystemFlag[*flags.BoolValue]{
		Sub:   flags.SubImplicit,
		Value: flags.Bool("no-color", "disable color output"),
	}
	if got := implicit.Definition().Meta.Sub; got != flags.SubImplicit {
		t.Errorf("SubImplicit: Meta.Sub = %q, want %q", got, flags.SubImplicit)
	}

	explicit := flags.SystemFlag[*flags.BoolValue]{
		Sub:   flags.SubExplicit,
		Value: flags.Bool("verbose", "enable verbose output"),
	}
	if got := explicit.Definition().Meta.Sub; got != flags.SubExplicit {
		t.Errorf("SubExplicit: Meta.Sub = %q, want %q", got, flags.SubExplicit)
	}
}

func TestSystemFlag_Definition_carries_name_and_usage(t *testing.T) {
	f := flags.SystemFlag[*flags.BoolValue]{
		Sub:   flags.SubImplicit,
		Value: flags.Bool("no-color", "disable color output"),
	}
	d := f.Definition()
	if d.Meta.Name != "no-color" {
		t.Errorf("Meta.Name = %q, want %q", d.Meta.Name, "no-color")
	}
	if d.Meta.Usage != "disable color output" {
		t.Errorf("Meta.Usage = %q, want %q", d.Meta.Usage, "disable color output")
	}
}

func TestSystemFlag_Effect_field(t *testing.T) {
	called := false
	f := flags.SystemFlag[*flags.BoolValue]{
		Effect: func(*app.App) { called = true },
	}
	f.Effect(nil)
	if !called {
		t.Error("Effect was not called")
	}
}

func TestSystemFlag_Effect_nil_by_default(t *testing.T) {
	var f flags.SystemFlag[*flags.BoolValue]
	if f.Effect != nil {
		t.Error("Effect should be nil when not set")
	}
}
