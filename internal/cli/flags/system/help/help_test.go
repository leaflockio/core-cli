// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package help_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/help"
)

func TestHelp_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = help.Help
}

func TestHelp_definition_name(t *testing.T) {
	if got := help.Help.Definition().Meta.Name; got != "help" {
		t.Errorf("Name = %q, want %q", got, "help")
	}
}

func TestHelp_definition_shorthand(t *testing.T) {
	if got := help.Help.Definition().Meta.Shorthand; got != "h" {
		t.Errorf("Shorthand = %q, want %q", got, "h")
	}
}

func TestHelp_definition_kind(t *testing.T) {
	if got := help.Help.Definition().Meta.Kind; got != flags.KindSystem {
		t.Errorf("Kind = %q, want %q", got, flags.KindSystem)
	}
}

func TestHelp_definition_sub_implicit(t *testing.T) {
	if got := help.Help.Definition().Meta.Sub; got != flags.SubImplicit {
		t.Errorf("Sub = %q, want %q", got, flags.SubImplicit)
	}
}

func TestHelp_definition_usage(t *testing.T) {
	want := "help for this command"
	if got := help.Help.Definition().Meta.Usage; got != want {
		t.Errorf("Usage = %q, want %q", got, want)
	}
}

func TestHelp_effect_not_nil(t *testing.T) {
	if help.Help.Effect == nil {
		t.Error("Effect must not be nil")
	}
}

func TestHelp_effect_callable(t *testing.T) {
	help.Help.Effect(nil)
}
