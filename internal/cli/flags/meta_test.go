// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
)

func TestMeta_zero_value(t *testing.T) {
	var m flags.Meta
	if m.Kind != "" || m.Sub != "" || m.Name != "" || m.Shorthand != "" || m.Usage != "" {
		t.Error("zero Meta should have all empty fields")
	}
}

func TestMeta_fields(t *testing.T) {
	m := flags.Meta{
		Kind:      flags.KindSystem,
		Sub:       flags.SubImplicit,
		Name:      "no-color",
		Shorthand: "n",
		Usage:     "disable color output",
	}
	if m.Kind != flags.KindSystem {
		t.Errorf("Kind = %q, want %q", m.Kind, flags.KindSystem)
	}
	if m.Sub != flags.SubImplicit {
		t.Errorf("Sub = %q, want %q", m.Sub, flags.SubImplicit)
	}
	if m.Name != "no-color" {
		t.Errorf("Name = %q, want %q", m.Name, "no-color")
	}
	if m.Shorthand != "n" {
		t.Errorf("Shorthand = %q, want %q", m.Shorthand, "n")
	}
	if m.Usage != "disable color output" {
		t.Errorf("Usage = %q, want %q", m.Usage, "disable color output")
	}
}
