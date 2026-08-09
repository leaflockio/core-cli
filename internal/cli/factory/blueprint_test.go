// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/level"
)

func TestBlueprint_zero_value(t *testing.T) {
	var a blueprint
	if a.level != level.LevelRoot {
		t.Errorf("level must be LevelRoot in zero value, got %v", a.level)
	}
	if a.hasFlags {
		t.Error("hasFlags must be false in zero value")
	}
	if a.hasChildren {
		t.Error("hasChildren must be false in zero value")
	}
	if a.flags != nil {
		t.Error("flags must be nil in zero value")
	}
	if a.persistentFlags != nil {
		t.Error("persistentFlags must be nil in zero value")
	}
}

func TestFlagSpec_zero_value(t *testing.T) {
	var f flagSpec
	if f.name != "" {
		t.Errorf("name must be empty in zero value, got %q", f.name)
	}
	if f.hasShorthand {
		t.Error("hasShorthand must be false in zero value")
	}
	if f.hasResolver {
		t.Error("hasResolver must be false in zero value")
	}
	if f.register != nil {
		t.Error("register must be nil in zero value")
	}
	if f.resolve != nil {
		t.Error("resolve must be nil in zero value")
	}
}
