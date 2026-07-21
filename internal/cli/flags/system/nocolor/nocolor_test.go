// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package nocolor_test

import (
	"bytes"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/nocolor"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
)

func TestNoColor_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = nocolor.NoColor
}

func TestNoColor_definition_name(t *testing.T) {
	if got := nocolor.NoColor.Definition().Meta.Name; got != "no-color" {
		t.Errorf("Name = %q, want %q", got, "no-color")
	}
}

func TestNoColor_definition_kind(t *testing.T) {
	if got := nocolor.NoColor.Definition().Meta.Kind; got != flags.KindSystem {
		t.Errorf("Kind = %q, want %q", got, flags.KindSystem)
	}
}

func TestNoColor_definition_sub_implicit(t *testing.T) {
	if got := nocolor.NoColor.Definition().Meta.Sub; got != flags.SubImplicit {
		t.Errorf("Sub = %q, want %q", got, flags.SubImplicit)
	}
}

func TestNoColor_definition_usage(t *testing.T) {
	if got := nocolor.NoColor.Definition().Meta.Usage; got != "Disable color output" {
		t.Errorf("Usage = %q, want %q", got, "Disable color output")
	}
}

func TestNoColor_effect_not_nil(t *testing.T) {
	if nocolor.NoColor.Effect == nil {
		t.Error("Effect must not be nil")
	}
}

func TestNoColor_effect_callable(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := app.NewBuilder().WithPrinter(printer).Build()
	nocolor.NoColor.Effect(a)
}
