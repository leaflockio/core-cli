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

func TestFlagKind_system(t *testing.T) {
	if flags.KindSystem != "system" {
		t.Errorf("KindSystem = %q, want %q", flags.KindSystem, "system")
	}
}

func TestFlagKind_command(t *testing.T) {
	if flags.KindCommand != "command" {
		t.Errorf("KindCommand = %q, want %q", flags.KindCommand, "command")
	}
}

func TestFlagSubcategory_implicit(t *testing.T) {
	if flags.SubImplicit != "implicit" {
		t.Errorf("SubImplicit = %q, want %q", flags.SubImplicit, "implicit")
	}
}

func TestFlagSubcategory_explicit(t *testing.T) {
	if flags.SubExplicit != "explicit" {
		t.Errorf("SubExplicit = %q, want %q", flags.SubExplicit, "explicit")
	}
}
