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

func TestString_constructor(t *testing.T) {
	f := flags.String("output", "output path")
	if f.Name != "output" {
		t.Errorf("Name = %q, want %q", f.Name, "output")
	}
	if f.Usage != "output path" {
		t.Errorf("Usage = %q, want %q", f.Usage, "output path")
	}
}

func TestStringValue_Validate_valid(t *testing.T) {
	if err := flags.String("output", "").WithDefault("MIT").Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestStringValue_Validate_tab_allowed(t *testing.T) {
	if err := flags.String("output", "").WithDefault("col\theader").Validate(); err != nil {
		t.Errorf("Validate() with tab = %v, want nil", err)
	}
}

func TestStringValue_Validate_rejects_null_byte(t *testing.T) {
	err := flags.String("output", "").WithDefault("bad\x00val").Validate()
	if err == nil {
		t.Error("Validate() should reject null byte")
	}
}

func TestStringValue_Validate_rejects_control_char(t *testing.T) {
	err := flags.String("output", "").WithDefault("line1\nline2").Validate()
	if err == nil {
		t.Error("Validate() should reject control characters")
	}
}

func TestStringValue_Default_empty(t *testing.T) {
	if flags.String("output", "").Default() != "" {
		t.Error("Default() should be empty before WithDefault")
	}
}

func TestStringValue_WithDefault(t *testing.T) {
	f := flags.String("output", "").WithDefault("MIT")
	if f.Default() != "MIT" {
		t.Errorf("Default() = %q, want %q", f.Default(), "MIT")
	}
}

func TestStringValue_Dest_nil_by_default(t *testing.T) {
	if flags.String("output", "").Dest() != nil {
		t.Error("Dest() should be nil before WithDest")
	}
}

func TestStringValue_WithDest(t *testing.T) {
	var dest string
	f := flags.String("output", "").WithDest(&dest)
	if f.Dest() != &dest {
		t.Error("Dest() should point to the bound variable")
	}
}

func TestStringValue_WithShorthand(t *testing.T) {
	f := flags.String("output", "").WithShorthand("o")
	if f.Shorthand != "o" {
		t.Errorf("Shorthand = %q, want %q", f.Shorthand, "o")
	}
}

// stringResolverStub satisfies StringResolver for interface verification.
type stringResolverStub struct {
	flags.CommandFlag[*flags.StringValue]
}

func (s stringResolverStub) IsResolver()          {}
func (s stringResolverStub) Resolve(string) error { return nil }

func TestStringResolver_satisfied_by_implementation(t *testing.T) {
	var _ flags.StringResolver = stringResolverStub{}
}
