// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield_test

import (
	"testing"

	"github.com/leaflockio/core-cli/internal/configfield"
)

// TestFlatten_nilSrc verifies an error is returned for a nil src.
func TestFlatten_nilSrc(t *testing.T) {
	if _, err := configfield.Flatten(nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestFlatten_nilPointerSrc verifies an error is returned for a typed nil
// pointer.
func TestFlatten_nilPointerSrc(t *testing.T) {
	type cfg struct {
		Format configfield.Field[string]
	}
	var p *cfg
	if _, err := configfield.Flatten(p); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestFlatten_notAStruct verifies an error is returned for a non-struct
// value.
func TestFlatten_notAStruct(t *testing.T) {
	if _, err := configfield.Flatten(42); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type flattenGoodConfig struct {
	Format   configfield.Field[string]
	unexport string
}

// TestFlatten_structValue verifies a plain struct (not a pointer) is
// accepted, and an unexported field is skipped.
func TestFlatten_structValue(t *testing.T) {
	c := flattenGoodConfig{Format: configfield.Field[string]{Default: "standard"}, unexport: "private"}
	m, err := configfield.Flatten(c)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if len(m) != 1 {
		t.Errorf("m = %#v, want only the format key (unexported field must not appear)", m)
	}
	if got := m["format"]; got != "standard" {
		t.Errorf(`m["format"] = %v, want "standard"`, got)
	}
}

// TestFlatten_pointerToStruct verifies a pointer to a struct is accepted.
func TestFlatten_pointerToStruct(t *testing.T) {
	c := &flattenGoodConfig{Format: configfield.Field[string]{Default: "standard"}}
	m, err := configfield.Flatten(c)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if got := m["format"]; got != "standard" {
		t.Errorf(`m["format"] = %v, want "standard"`, got)
	}
}

type flattenInternalConfig struct {
	Format configfield.Field[string]
	Secret configfield.Field[string]
}

// TestFlatten_internalField_isSkipped verifies a field marked Internal
// isn't written into the output map.
func TestFlatten_internalField_isSkipped(t *testing.T) {
	c := flattenInternalConfig{
		Format: configfield.Field[string]{Default: "standard"},
		Secret: configfield.Field[string]{Default: "hidden", Internal: true},
	}
	m, err := configfield.Flatten(&c)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if _, ok := m["secret"]; ok {
		t.Errorf("m contains Internal field: %#v", m)
	}
	if got := m["format"]; got != "standard" {
		t.Errorf(`m["format"] = %v, want "standard"`, got)
	}
}

type flattenUnwrappedFieldConfig struct {
	Bad string
}

// TestFlatten_unwrappedField_errors verifies flattening a struct with a
// field that isn't a Field returns an error.
func TestFlatten_unwrappedField_errors(t *testing.T) {
	_, err := configfield.Flatten(&flattenUnwrappedFieldConfig{Bad: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type flattenNestedConfig struct {
	A configfield.Field[bool]
	B configfield.Field[bool]
}

type flattenWithNestedConfig struct {
	Nested configfield.Field[flattenNestedConfig]
}

// TestFlatten_structTypedField_isNested verifies a struct-typed Field's
// resolved value is flattened into a nested map rather than assigned
// directly.
func TestFlatten_structTypedField_isNested(t *testing.T) {
	c := flattenWithNestedConfig{
		Nested: configfield.Field[flattenNestedConfig]{
			Default: flattenNestedConfig{
				A: configfield.Field[bool]{Default: true},
				B: configfield.Field[bool]{Default: false},
			},
		},
	}
	m, err := configfield.Flatten(&c)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	nested, ok := m["nested"].(map[string]any)
	if !ok {
		t.Fatalf(`m["nested"] = %#v, want a nested map`, m["nested"])
	}
	if got := nested["a"]; got != true {
		t.Errorf(`nested["a"] = %v, want true`, got)
	}
	if got := nested["b"]; got != false {
		t.Errorf(`nested["b"] = %v, want false`, got)
	}
}

type flattenNestedBadConfig struct {
	A configfield.Field[bool]
	B string
}

type flattenWithNestedBadConfig struct {
	Nested configfield.Field[flattenNestedBadConfig]
}

// TestFlatten_structTypedField_nestedErrorIsPropagated verifies an error
// while flattening a struct-typed Field's nested value is propagated.
func TestFlatten_structTypedField_nestedErrorIsPropagated(t *testing.T) {
	c := flattenWithNestedBadConfig{
		Nested: configfield.Field[flattenNestedBadConfig]{Default: flattenNestedBadConfig{}},
	}
	_, err := configfield.Flatten(&c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
