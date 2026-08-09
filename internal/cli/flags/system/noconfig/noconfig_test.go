// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package noconfig_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/noconfig"
)

func TestNoConfig_satisfies_Flag_interface(t *testing.T) {
	var _ flags.Flag = noconfig.NoConfig
}

func TestNoConfig_definition_name(t *testing.T) {
	if got := noconfig.NoConfig.Definition().Meta.Name; got != "no-config" {
		t.Errorf("Name = %q, want %q", got, "no-config")
	}
}

func TestNoConfig_definition_kind(t *testing.T) {
	if got := noconfig.NoConfig.Definition().Meta.Kind; got != flags.KindSystem {
		t.Errorf("Kind = %q, want %q", got, flags.KindSystem)
	}
}

func TestNoConfig_definition_sub_implicit(t *testing.T) {
	if got := noconfig.NoConfig.Definition().Meta.Sub; got != flags.SubImplicit {
		t.Errorf("Sub = %q, want %q", got, flags.SubImplicit)
	}
}

func TestNoConfig_definition_usage(t *testing.T) {
	want := "Skip project config files"
	if got := noconfig.NoConfig.Definition().Meta.Usage; got != want {
		t.Errorf("Usage = %q, want %q", got, want)
	}
}

func TestNoConfig_effect_not_nil(t *testing.T) {
	if noconfig.NoConfig.Effect == nil {
		t.Error("Effect must not be nil")
	}
}

func TestNoConfig_effect_callable(t *testing.T) {
	noconfig.NoConfig.Effect(nil)
}
