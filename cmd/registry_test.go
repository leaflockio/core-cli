// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"testing"
)

func TestCommands_defineSucceeds(t *testing.T) {
	for i, c := range commands {
		def := c.Define(nil)
		if def == nil {
			t.Fatalf("commands[%d]: expected non-nil Definition", i)
		}
		if def.Meta.Use == "" {
			t.Errorf("commands[%d]: expected non-empty Meta.Use", i)
		}
	}
}
