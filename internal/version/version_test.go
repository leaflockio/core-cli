// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package version

import "testing"

func TestCurrent_returnsNonNil(t *testing.T) {
	if Current() == nil {
		t.Fatal("expected non-nil Info from Current()")
	}
}

func TestCurrent_defaultValues(t *testing.T) {
	info := Current()

	if info.Version != "dev" {
		t.Errorf("expected default Version %q, got %q", "dev", info.Version)
	}
	if info.Commit != "none" {
		t.Errorf("expected default Commit %q, got %q", "none", info.Commit)
	}
	if info.Date != "unknown" {
		t.Errorf("expected default Date %q, got %q", "unknown", info.Date)
	}
}

func TestCurrent_reflectsBuildVars(t *testing.T) {
	original := Version
	Version = "1.2.3"
	defer func() { Version = original }()

	if Current().Version != "1.2.3" {
		t.Errorf("expected Version %q, got %q", "1.2.3", Current().Version)
	}
}
