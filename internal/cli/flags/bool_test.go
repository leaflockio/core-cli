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

func TestBool_constructor(t *testing.T) {
	f := flags.Bool("verbose", "enable verbose output")
	if f.Name != "verbose" {
		t.Errorf("Name = %q, want %q", f.Name, "verbose")
	}
	if f.Usage != "enable verbose output" {
		t.Errorf("Usage = %q, want %q", f.Usage, "enable verbose output")
	}
}

func TestBoolValue_Validate_always_nil(t *testing.T) {
	if err := flags.Bool("verbose", "").Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestBoolValue_Default_zero(t *testing.T) {
	if flags.Bool("verbose", "").Default() != false {
		t.Error("Default() should be false before WithDefault")
	}
}

func TestBoolValue_WithDefault(t *testing.T) {
	f := flags.Bool("verbose", "").WithDefault(true)
	if f.Default() != true {
		t.Errorf("Default() = %v, want true", f.Default())
	}
}

func TestBoolValue_Dest_nil_by_default(t *testing.T) {
	if flags.Bool("verbose", "").Dest() != nil {
		t.Error("Dest() should be nil before WithDest")
	}
}

func TestBoolValue_WithDest(t *testing.T) {
	var dest bool
	f := flags.Bool("verbose", "").WithDest(&dest)
	if f.Dest() != &dest {
		t.Error("Dest() should point to the bound variable")
	}
}

func TestBoolValue_WithShorthand(t *testing.T) {
	f := flags.Bool("verbose", "").WithShorthand("v")
	if f.Shorthand != "v" {
		t.Errorf("Shorthand = %q, want %q", f.Shorthand, "v")
	}
}

// boolResolverStub satisfies BoolResolver for interface verification.
type boolResolverStub struct {
	flags.CommandFlag[*flags.BoolValue]
}

func (s boolResolverStub) IsResolver()        {}
func (s boolResolverStub) Resolve(bool) error { return nil }

func TestBoolResolver_satisfied_by_implementation(t *testing.T) {
	var _ flags.BoolResolver = boolResolverStub{}
}
