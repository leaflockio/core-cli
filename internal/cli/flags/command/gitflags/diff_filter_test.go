// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package gitflags_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags/command/gitflags"
)

func TestDiffFilter_definition(t *testing.T) {
	def := gitflags.DiffFilter.Definition()
	if def.Meta.Name != "diff-filter" {
		t.Errorf("Name = %q, want %q", def.Meta.Name, "diff-filter")
	}
	if gitflags.DiffFilter.Value.Default() != "ACM" {
		t.Errorf("Default() = %q, want %q", gitflags.DiffFilter.Value.Default(), "ACM")
	}
}

func TestDiffFilter_WithDest(t *testing.T) {
	var dest string
	cf := gitflags.DiffFilter.WithDest(&dest)
	if cf.Value.Dest() != &dest {
		t.Error("Dest() should point to the bound variable")
	}
}

// TestDiffFilter_WithDest_independentAcrossCalls guards against the same
// regression WithDest had before it was made copy-on-write: two calls on
// the shared DiffFilter var must not clobber each other's binding.
func TestDiffFilter_WithDest_independentAcrossCalls(t *testing.T) {
	var destA, destB string
	a := gitflags.DiffFilter.WithDest(&destA)
	b := gitflags.DiffFilter.WithDest(&destB)

	if a.Value.Dest() != &destA {
		t.Errorf("a.Value.Dest() = %p, want %p", a.Value.Dest(), &destA)
	}
	if b.Value.Dest() != &destB {
		t.Errorf("b.Value.Dest() = %p, want %p", b.Value.Dest(), &destB)
	}
	if a.Value.Dest() == b.Value.Dest() {
		t.Error("a and b should have independent Dest pointers")
	}
}

func TestDiffFilter_Validate_validCases(t *testing.T) {
	cases := []string{
		"ACM",  // the default
		"acm",  // lowercase (exclude) is valid too
		"AcM",  // mixed case
		"ACM*", // letters + trailing star
		"acm*", // lowercase + trailing star
		"A",    // single letter
		"*",    // bare star: degenerate but syntactically valid — matches real git, which accepts it without error
		"",     // empty: no letters to reject
	}
	for _, value := range cases {
		if err := gitflags.DiffFilter.Validate(value); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", value, err)
		}
	}
}

func TestDiffFilter_Validate_invalidCases(t *testing.T) {
	cases := []string{
		"Z",   // not a recognized diff-filter letter
		"AZ",  // valid letter mixed with an invalid one
		"Z*",  // invalid letter before the trailing star
		"A*C", // star not at the end
		"**",  // star repeated
		"***", // star repeated further
		"AC ", // a character that isn't a letter or star at all
		"1",   // a digit
	}
	for _, value := range cases {
		if err := gitflags.DiffFilter.Validate(value); err == nil {
			t.Errorf("Validate(%q) = nil, want an error", value)
		}
	}
}
