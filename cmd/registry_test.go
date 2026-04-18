// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"testing"
)

func TestRegistry_allEntriesHaveGroupID(t *testing.T) {
	for i, e := range registry {
		if e.groupID == "" {
			t.Errorf("registry[%d] has empty groupID", i)
		}
	}
}

func TestRegistry_allEntriesHaveFactory(t *testing.T) {
	for i, e := range registry {
		if e.factory == nil {
			t.Errorf("registry[%d] has nil factory", i)
		}
	}
}
