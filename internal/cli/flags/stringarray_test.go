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

func TestStringSlice_constructor(t *testing.T) {
	f := flags.StringSlice("tags", "list of tags")
	if f.Name != "tags" {
		t.Errorf("Name = %q, want %q", f.Name, "tags")
	}
	if f.Usage != "list of tags" {
		t.Errorf("Usage = %q, want %q", f.Usage, "list of tags")
	}
}

func TestStringSliceValue_Validate_empty(t *testing.T) {
	if err := flags.StringSlice("tags", "").Validate(); err != nil {
		t.Errorf("Validate() on empty slice = %v, want nil", err)
	}
}

func TestStringSliceValue_Validate_valid_elements(t *testing.T) {
	f := flags.StringSlice("tags", "").WithDefault([]string{"MIT", "Apache-2.0"})
	if err := f.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestStringSliceValue_Validate_rejects_null_byte_in_element(t *testing.T) {
	f := flags.StringSlice("tags", "").WithDefault([]string{"valid", "bad\x00val"})
	if err := f.Validate(); err == nil {
		t.Error("Validate() should reject element with null byte")
	}
}

func TestStringSliceValue_Validate_rejects_control_char_in_element(t *testing.T) {
	f := flags.StringSlice("tags", "").WithDefault([]string{"line1\nline2"})
	if err := f.Validate(); err == nil {
		t.Error("Validate() should reject element with control character")
	}
}

func TestStringSliceValue_Default_nil_by_default(t *testing.T) {
	if flags.StringSlice("tags", "").Default() != nil {
		t.Error("Default() should be nil before WithDefault")
	}
}

func TestStringSliceValue_WithDefault(t *testing.T) {
	vals := []string{"MIT", "Apache-2.0"}
	f := flags.StringSlice("tags", "").WithDefault(vals)
	if len(f.Default()) != 2 {
		t.Errorf("Default() length = %d, want 2", len(f.Default()))
	}
}

func TestStringSliceValue_Dest_nil_by_default(t *testing.T) {
	if flags.StringSlice("tags", "").Dest() != nil {
		t.Error("Dest() should be nil before WithDest")
	}
}

func TestStringSliceValue_WithDest(t *testing.T) {
	var dest []string
	f := flags.StringSlice("tags", "").WithDest(&dest)
	if f.Dest() != &dest {
		t.Error("Dest() should point to the bound variable")
	}
}

func TestStringSliceValue_WithShorthand(t *testing.T) {
	f := flags.StringSlice("tags", "").WithShorthand("t")
	if f.Shorthand != "t" {
		t.Errorf("Shorthand = %q, want %q", f.Shorthand, "t")
	}
}

// stringSliceResolverStub satisfies StringSliceResolver for interface verification.
type stringSliceResolverStub struct {
	flags.CommandFlag[*flags.StringSliceValue]
}

func (s stringSliceResolverStub) IsResolver()            {}
func (s stringSliceResolverStub) Resolve([]string) error { return nil }

func TestStringSliceResolver_satisfied_by_implementation(t *testing.T) {
	var _ flags.StringSliceResolver = stringSliceResolverStub{}
}

func TestStringSliceValue_meta_via_CommandFlag(t *testing.T) {
	f := flags.CommandFlag[*flags.StringSliceValue]{Value: flags.StringSlice("tags", "list of tags")}
	d := f.Definition()
	if d.Meta.Name != "tags" {
		t.Errorf("Meta.Name = %q, want %q", d.Meta.Name, "tags")
	}
	if d.Meta.Kind != flags.KindCommand {
		t.Errorf("Meta.Kind = %q, want %q", d.Meta.Kind, flags.KindCommand)
	}
}
