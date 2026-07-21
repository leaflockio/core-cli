// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package builtin

import (
	"slices"
	"testing"
)

// --- Lookup ---

func TestLookup_canonicalNames(t *testing.T) {
	for _, name := range []string{"proprietary", "open-source"} {
		content, ok := Lookup(name)
		if !ok {
			t.Errorf("Lookup(%q): expected ok=true", name)
		}
		if content == "" {
			t.Errorf("Lookup(%q): expected non-empty content", name)
		}
	}
}

func TestLookup_aliases(t *testing.T) {
	cases := []struct{ alias, canonical string }{
		{"private", "proprietary"},
		{"public", "open-source"},
	}
	for _, tc := range cases {
		aliasContent, aliasOK := Lookup(tc.alias)
		canonicalContent, canonicalOK := Lookup(tc.canonical)
		if !aliasOK {
			t.Errorf("Lookup(%q): expected ok=true", tc.alias)
		}
		if !canonicalOK {
			t.Errorf("Lookup(%q): expected ok=true", tc.canonical)
		}
		if aliasContent != canonicalContent {
			t.Errorf("Lookup(%q) content differs from Lookup(%q)", tc.alias, tc.canonical)
		}
	}
}

func TestLookup_caseInsensitive(t *testing.T) {
	for _, name := range []string{"Proprietary", "PROPRIETARY", "Private", "Open-Source", "PUBLIC"} {
		_, ok := Lookup(name)
		if !ok {
			t.Errorf("Lookup(%q): expected ok=true (case-insensitive)", name)
		}
	}
}

func TestLookup_unknown(t *testing.T) {
	_, ok := Lookup("unknown")
	if ok {
		t.Error("Lookup(\"unknown\"): expected ok=false")
	}
}

func TestLookup_empty(t *testing.T) {
	_, ok := Lookup("")
	if ok {
		t.Error("Lookup(\"\"): expected ok=false")
	}
}

// --- IsPrivate ---

func TestIsPrivate_privateNames(t *testing.T) {
	for _, name := range []string{"proprietary", "private", "Proprietary", "PRIVATE"} {
		if !IsPrivate(name) {
			t.Errorf("IsPrivate(%q): expected true", name)
		}
	}
}

func TestIsPrivate_nonPrivateNames(t *testing.T) {
	for _, name := range []string{"open-source", "public", "MIT", "Apache-2.0", "", "unknown"} {
		if IsPrivate(name) {
			t.Errorf("IsPrivate(%q): expected false", name)
		}
	}
}

// --- Names ---

func TestNames_returnsCanonicalOnly(t *testing.T) {
	names := Names()
	want := []string{"proprietary", "open-source"}
	for _, w := range want {
		if !slices.Contains(names, w) {
			t.Errorf("Names(): expected %q to be present", w)
		}
	}
	for _, alias := range []string{"private", "public"} {
		if slices.Contains(names, alias) {
			t.Errorf("Names(): alias %q should not be in canonical names", alias)
		}
	}
}

func TestNames_stable(t *testing.T) {
	first := Names()
	second := Names()
	if len(first) != len(second) {
		t.Error("Names(): successive calls returned different lengths")
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("Names(): successive calls differ at index %d: %q vs %q", i, first[i], second[i])
		}
	}
}
