// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"testing"
)

func TestGroups_notEmpty(t *testing.T) {
	if len(groups) == 0 {
		t.Error("expected at least one group defined")
	}
}

func TestGroups_allKnownIDs(t *testing.T) {
	known := map[string]bool{groupGeneral: true, groupTools: true}
	for _, g := range groups {
		if !known[g] {
			t.Errorf("unknown group ID %q in groups", g)
		}
	}
}
