// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig_test

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
)

// errTestMustBeNonNegative is the error a test Validate func rejects a
// negative value with, shared by every test in this package that needs one.
var errTestMustBeNonNegative = errors.New("must be non-negative")

// TestField_Value_defaultWhenNotOverridden verifies Value returns Default
// when the field has never been decoded.
func TestField_Value_defaultWhenNotOverridden(t *testing.T) {
	f := cmdconfig.Field[string]{Default: "standard"}
	if got := f.Value(); got != "standard" {
		t.Errorf("Value() = %q, want %q", got, "standard")
	}
}

// TestField_Value_overrideAfterDecode verifies Value returns the override
// once Decode has recorded one.
func TestField_Value_overrideAfterDecode(t *testing.T) {
	type cfg struct {
		Format cmdconfig.Field[string]
	}
	c := cfg{Format: cmdconfig.Field[string]{Default: "standard"}}
	if err := cmdconfig.Decode(map[string]any{"format": "spdx"}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Format.Value(); got != "spdx" {
		t.Errorf("Value() = %q, want %q", got, "spdx")
	}
}

// TestField_key_explicit verifies decode honors an explicit Key over the
// snake_case fallback.
func TestField_key_explicit(t *testing.T) {
	type cfg struct {
		Format cmdconfig.Field[string]
	}
	c := cfg{Format: cmdconfig.Field[string]{Key: "fmt", Default: "standard"}}
	if err := cmdconfig.Decode(map[string]any{"fmt": "spdx"}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Format.Value(); got != "spdx" {
		t.Errorf("Value() = %q, want %q", got, "spdx")
	}
}

// TestField_key_snakeCaseFallback verifies decode falls back to the struct
// field's own name, converted to snake_case, when Key is empty — including
// an acronym-bearing name, which must not be split letter by letter.
func TestField_key_snakeCaseFallback(t *testing.T) {
	type cfg struct {
		UpdateLicenseFile cmdconfig.Field[bool]
		TTL               cmdconfig.Field[int]
	}
	c := cfg{
		UpdateLicenseFile: cmdconfig.Field[bool]{Default: false},
		TTL:               cmdconfig.Field[int]{Default: 24},
	}
	section := map[string]any{
		"update_license_file": true,
		"ttl":                 48,
	}
	if err := cmdconfig.Decode(section, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.UpdateLicenseFile.Value(); got != true {
		t.Errorf("UpdateLicenseFile.Value() = %v, want true", got)
	}
	if got := c.TTL.Value(); got != 48 {
		t.Errorf("TTL.Value() = %v, want 48", got)
	}
}

// TestField_setOverride_scalarSuccess verifies a scalar-typed Field decodes
// straightforwardly.
func TestField_setOverride_scalarSuccess(t *testing.T) {
	type cfg struct {
		Count cmdconfig.Field[int]
	}
	c := cfg{Count: cmdconfig.Field[int]{Default: 1}}
	if err := cmdconfig.Decode(map[string]any{"count": 5}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Count.Value(); got != 5 {
		t.Errorf("Count.Value() = %d, want 5", got)
	}
}

// TestField_setOverride_scalarTypeMismatch verifies a scalar-typed Field's
// decode error is propagated.
func TestField_setOverride_scalarTypeMismatch(t *testing.T) {
	type cfg struct {
		Count cmdconfig.Field[int]
	}
	c := cfg{Count: cmdconfig.Field[int]{Default: 1}}
	err := cmdconfig.Decode(map[string]any{"count": "not-a-number"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestField_setOverride_optionAllowed verifies a value matching one of
// Options is accepted.
func TestField_setOverride_optionAllowed(t *testing.T) {
	type cfg struct {
		Format cmdconfig.Field[string]
	}
	c := cfg{
		Format: cmdconfig.Field[string]{
			Default: "standard",
			Options: []cmdconfig.Option[string]{
				{Value: "standard", Description: "proprietary"},
				{Value: "spdx", Description: "public open-source"},
			},
		},
	}
	if err := cmdconfig.Decode(map[string]any{"format": "spdx"}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Format.Value(); got != "spdx" {
		t.Errorf("Format.Value() = %q, want %q", got, "spdx")
	}
}

// TestField_setOverride_optionRejected verifies a value not among Options
// is rejected, and the override is not applied.
func TestField_setOverride_optionRejected(t *testing.T) {
	type cfg struct {
		Format cmdconfig.Field[string]
	}
	c := cfg{
		Format: cmdconfig.Field[string]{
			Default: "standard",
			Options: []cmdconfig.Option[string]{
				{Value: "standard", Description: "proprietary"},
				{Value: "spdx", Description: "public open-source"},
			},
		},
	}
	err := cmdconfig.Decode(map[string]any{"format": "made-up"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := c.Format.Value(); got != "standard" {
		t.Errorf("Format.Value() = %q, want %q (rejected override must not apply)", got, "standard")
	}
}

// TestField_setOverride_validateAllowed verifies a value accepted by
// Validate is applied.
func TestField_setOverride_validateAllowed(t *testing.T) {
	type cfg struct {
		Count cmdconfig.Field[int]
	}
	c := cfg{
		Count: cmdconfig.Field[int]{
			Default: 1,
			Validate: func(v int) error {
				if v < 0 {
					return errTestMustBeNonNegative
				}
				return nil
			},
		},
	}
	if err := cmdconfig.Decode(map[string]any{"count": 5}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Count.Value(); got != 5 {
		t.Errorf("Count.Value() = %d, want 5", got)
	}
}

// TestField_setOverride_validateRejected verifies a value rejected by
// Validate is not applied.
func TestField_setOverride_validateRejected(t *testing.T) {
	type cfg struct {
		Count cmdconfig.Field[int]
	}
	c := cfg{
		Count: cmdconfig.Field[int]{
			Default: 1,
			Validate: func(v int) error {
				if v < 0 {
					return errTestMustBeNonNegative
				}
				return nil
			},
		},
	}
	err := cmdconfig.Decode(map[string]any{"count": -5}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := c.Count.Value(); got != 1 {
		t.Errorf("Count.Value() = %d, want 1 (rejected override must not apply)", got)
	}
}

type optionsOnStructFixture struct {
	A cmdconfig.Field[bool]
}

// TestField_setOverride_optionsOnStructField_rejected verifies Options set
// on a struct-typed Field is rejected outright, rather than attempted.
func TestField_setOverride_optionsOnStructField_rejected(t *testing.T) {
	type cfg struct {
		Nested cmdconfig.Field[optionsOnStructFixture]
	}
	c := cfg{
		Nested: cmdconfig.Field[optionsOnStructFixture]{
			Default: optionsOnStructFixture{},
			Options: []cmdconfig.Option[optionsOnStructFixture]{
				{Value: optionsOnStructFixture{}},
			},
		},
	}
	err := cmdconfig.Decode(map[string]any{"nested": map[string]any{"a": true}}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// nestedFieldFixture is a struct-typed Field's T, for the recursive
// setOverride cases below.
type nestedFieldFixture struct {
	A cmdconfig.Field[bool]
	B cmdconfig.Field[bool]
}

// TestField_setOverride_structSuccess verifies a struct-typed Field decodes
// recursively, seeding the override from a copy of Default so untouched
// nested fields keep their real default rather than a zero value.
func TestField_setOverride_structSuccess(t *testing.T) {
	type cfg struct {
		Nested cmdconfig.Field[nestedFieldFixture]
	}
	c := cfg{
		Nested: cmdconfig.Field[nestedFieldFixture]{
			Default: nestedFieldFixture{
				A: cmdconfig.Field[bool]{Default: true},
				B: cmdconfig.Field[bool]{Default: true},
			},
		},
	}
	section := map[string]any{
		"nested": map[string]any{"a": false},
	}
	if err := cmdconfig.Decode(section, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	resolved := c.Nested.Value()
	if got := resolved.A.Value(); got != false {
		t.Errorf("Nested.A.Value() = %v, want false (overridden)", got)
	}
	if got := resolved.B.Value(); got != true {
		t.Errorf("Nested.B.Value() = %v, want true (real default preserved)", got)
	}
}

// TestField_setOverride_structRawNotAMapping verifies an error is returned
// when a struct-typed Field's raw value isn't a map.
func TestField_setOverride_structRawNotAMapping(t *testing.T) {
	type cfg struct {
		Nested cmdconfig.Field[nestedFieldFixture]
	}
	c := cfg{Nested: cmdconfig.Field[nestedFieldFixture]{Default: nestedFieldFixture{}}}
	err := cmdconfig.Decode(map[string]any{"nested": "not-a-map"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// badNestedFieldFixture has a field that isn't a Field[T], for exercising
// the recursive decode-failure path.
type badNestedFieldFixture struct {
	A cmdconfig.Field[bool]
	B string
}

// TestField_setOverride_structRecursiveDecodeError verifies an error from
// the recursive decode of a struct-typed Field is propagated.
func TestField_setOverride_structRecursiveDecodeError(t *testing.T) {
	type cfg struct {
		Nested cmdconfig.Field[badNestedFieldFixture]
	}
	c := cfg{Nested: cmdconfig.Field[badNestedFieldFixture]{Default: badNestedFieldFixture{}}}
	section := map[string]any{
		"nested": map[string]any{"a": true},
	}
	err := cmdconfig.Decode(section, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
