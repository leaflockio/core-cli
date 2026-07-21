// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
)

func TestImplicitSystemFlags_not_empty(t *testing.T) {
	if len(implicitSystemFlags) == 0 {
		t.Fatal("implicitSystemFlags must contain at least one flag")
	}
}

func TestImplicitSystemFlags_no_nil_entries(t *testing.T) {
	for i, f := range implicitSystemFlags {
		if f == nil {
			t.Errorf("implicitSystemFlags[%d] is nil", i)
		}
	}
}

func TestImplicitSystemFlags_all_sub_implicit(t *testing.T) {
	for _, f := range implicitSystemFlags {
		d := f.Definition()
		if d.Meta.Sub != flags.SubImplicit {
			t.Errorf("flag %q: Sub = %q, want SubImplicit", d.Meta.Name, d.Meta.Sub)
		}
	}
}

func TestImplicitSystemFlags_all_kind_system(t *testing.T) {
	for _, f := range implicitSystemFlags {
		d := f.Definition()
		if d.Meta.Kind != flags.KindSystem {
			t.Errorf("flag %q: Kind = %q, want KindSystem", d.Meta.Name, d.Meta.Kind)
		}
	}
}

func TestImplicitSystemFlags_contains_nocolor(t *testing.T) {
	for _, f := range implicitSystemFlags {
		if f.Definition().Meta.Name == "no-color" {
			return
		}
	}
	t.Error("implicitSystemFlags must contain the no-color flag")
}
