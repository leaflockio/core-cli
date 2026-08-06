// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package verbose_test

import (
	"bytes"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/verbose"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
)

func TestVerbose_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = verbose.Verbose
}

func TestVerbose_definition_name(t *testing.T) {
	if got := verbose.Verbose.Definition().Meta.Name; got != "verbose" {
		t.Errorf("Name = %q, want %q", got, "verbose")
	}
}

func TestVerbose_definition_kind(t *testing.T) {
	if got := verbose.Verbose.Definition().Meta.Kind; got != flags.KindSystem {
		t.Errorf("Kind = %q, want %q", got, flags.KindSystem)
	}
}

func TestVerbose_definition_sub_explicit(t *testing.T) {
	if got := verbose.Verbose.Definition().Meta.Sub; got != flags.SubExplicit {
		t.Errorf("Sub = %q, want %q", got, flags.SubExplicit)
	}
}

func TestVerbose_definition_usage(t *testing.T) {
	if got := verbose.Verbose.Definition().Meta.Usage; got != "Enable verbose output" {
		t.Errorf("Usage = %q, want %q", got, "Enable verbose output")
	}
}

func TestVerbose_definition_shorthand(t *testing.T) {
	if got := verbose.Verbose.Definition().Meta.ShortFlag(); got != "-v" {
		t.Errorf("ShortFlag() = %q, want %q", got, "-v")
	}
}

func TestVerbose_effect_not_nil(t *testing.T) {
	if verbose.Verbose.Effect == nil {
		t.Error("Effect must not be nil")
	}
}

func TestVerbose_effect_callable(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := app.NewBuilder().WithPrinter(printer).Build()
	verbose.Verbose.Effect(a)
}
