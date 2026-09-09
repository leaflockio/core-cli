// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield_test

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/configfield"
)

// TestValidate_nil verifies an error is returned for a nil value.
func TestValidate_nil(t *testing.T) {
	if err := configfield.Validate(nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestValidate_notAStruct verifies an error is returned for a non-struct
// value.
func TestValidate_notAStruct(t *testing.T) {
	if err := configfield.Validate(42); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestValidate_pointerToNonStruct verifies a pointer is unwrapped before
// the struct check, so a pointer to a non-struct is still rejected.
func TestValidate_pointerToNonStruct(t *testing.T) {
	n := 42
	if err := configfield.Validate(&n); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestValidate_nilPointer verifies an error is returned for a typed nil
// pointer.
func TestValidate_nilPointer(t *testing.T) {
	var p *validateGoodConfig
	if err := configfield.Validate(p); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateGoodConfig struct {
	Format configfield.Field[string]
}

// TestValidate_structValue verifies a plain struct (not a pointer) is
// accepted.
func TestValidate_structValue(t *testing.T) {
	if err := configfield.Validate(validateGoodConfig{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidate_pointerToStruct verifies a pointer to a struct is accepted.
func TestValidate_pointerToStruct(t *testing.T) {
	if err := configfield.Validate(&validateGoodConfig{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateUnexportedFieldConfig struct {
	Format   configfield.Field[string]
	internal string
}

// TestValidate_skipsUnexportedFields verifies an unexported field, even one
// that isn't a Field, doesn't cause a violation.
func TestValidate_skipsUnexportedFields(t *testing.T) {
	c := validateUnexportedFieldConfig{internal: "private"}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateBadConfig struct {
	Bad     string
	AlsoBad int
}

// TestValidate_unwrappedFields_areRejected verifies every unwrapped field
// is reported, not just the first — Validate collects every violation
// instead of stopping early.
func TestValidate_unwrappedFields_areRejected(t *testing.T) {
	err := configfield.Validate(&validateBadConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var joiner interface{ Unwrap() []error }
	if !errors.As(err, &joiner) {
		t.Fatalf("error chain does not contain a joined error: %v", err)
	}
	if got := len(joiner.Unwrap()); got != 2 {
		t.Errorf("collected %d violations, want 2", got)
	}
}

type validateOptionsConfig struct {
	Format configfield.Field[string]
}

// TestValidate_defaultAmongOptions_passes verifies a Default that matches
// one of its own Options is accepted.
func TestValidate_defaultAmongOptions_passes(t *testing.T) {
	c := validateOptionsConfig{
		Format: configfield.Field[string]{
			Default: "standard",
			Options: []configfield.Option[string]{
				{Value: "standard"},
				{Value: "spdx"},
			},
		},
	}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidate_defaultNotAmongOptions_fails verifies a Default that doesn't
// match any of its own Options is rejected — catching an author's typo or
// an Options list that fell out of sync with Default.
func TestValidate_defaultNotAmongOptions_fails(t *testing.T) {
	c := validateOptionsConfig{
		Format: configfield.Field[string]{
			Default: "not-a-real-option",
			Options: []configfield.Option[string]{
				{Value: "standard"},
				{Value: "spdx"},
			},
		},
	}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateValidateFuncConfig struct {
	Count configfield.Field[int]
}

// TestValidate_defaultAcceptedByValidate_passes verifies a Default accepted
// by the field's own Validate func is accepted.
func TestValidate_defaultAcceptedByValidate_passes(t *testing.T) {
	c := validateValidateFuncConfig{
		Count: configfield.Field[int]{
			Default: 1,
			Validate: func(v int) error {
				if v < 0 {
					return errTestMustBeNonNegative
				}
				return nil
			},
		},
	}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidate_defaultRejectedByValidate_fails verifies a Default rejected
// by the field's own Validate func is reported — catching an author's own
// inconsistent Default.
func TestValidate_defaultRejectedByValidate_fails(t *testing.T) {
	c := validateValidateFuncConfig{
		Count: configfield.Field[int]{
			Default: -1,
			Validate: func(v int) error {
				if v < 0 {
					return errTestMustBeNonNegative
				}
				return nil
			},
		},
	}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateOptionsOnStructConfig struct {
	Nested configfield.Field[optionsOnStructFixture]
}

// TestValidate_optionsOnStructField_rejected verifies Options set on a
// struct-typed Field is rejected outright, rather than attempted.
func TestValidate_optionsOnStructField_rejected(t *testing.T) {
	c := validateOptionsOnStructConfig{
		Nested: configfield.Field[optionsOnStructFixture]{
			Default: optionsOnStructFixture{},
			Options: []configfield.Option[optionsOnStructFixture]{
				{Value: optionsOnStructFixture{}},
			},
		},
	}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateNestedGoodConfig struct {
	A configfield.Field[bool]
}

type validateWithNestedGoodConfig struct {
	Nested configfield.Field[validateNestedGoodConfig]
}

// TestValidate_structTypedField_recursesAndPasses verifies a struct-typed
// Field recurses into T's own fields, and a valid nested shape passes.
func TestValidate_structTypedField_recursesAndPasses(t *testing.T) {
	if err := configfield.Validate(&validateWithNestedGoodConfig{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateNestedBadConfig struct {
	A configfield.Field[bool]
	B string
}

type validateWithNestedBadConfig struct {
	Nested configfield.Field[validateNestedBadConfig]
}

// TestValidate_structTypedField_recursesAndFails verifies a struct-typed
// Field recurses into T's own fields, and an invalid nested shape is
// reported with a path back to the offending field.
func TestValidate_structTypedField_recursesAndFails(t *testing.T) {
	err := configfield.Validate(&validateWithNestedBadConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateDuplicateKeyConfig struct {
	Format      configfield.Field[string]
	FormatAgain configfield.Field[string]
}

// TestValidate_duplicateKey_rejected verifies two sibling fields resolving
// to the same key are rejected — otherwise they'd silently share one
// decoded value, or one would overwrite the other when flattened.
func TestValidate_duplicateKey_rejected(t *testing.T) {
	c := validateDuplicateKeyConfig{
		Format:      configfield.Field[string]{Key: "format"},
		FormatAgain: configfield.Field[string]{Key: "format"},
	}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestValidate_noDuplicateKey_passes verifies distinct keys don't
// false-positive as duplicates.
func TestValidate_noDuplicateKey_passes(t *testing.T) {
	c := validateDuplicateKeyConfig{
		Format:      configfield.Field[string]{Key: "format"},
		FormatAgain: configfield.Field[string]{Key: "format_again"},
	}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateInternalKeyCollisionConfig struct {
	Secret configfield.Field[string]
	Format configfield.Field[string]
}

// TestValidate_duplicateKey_rejectedEvenWhenOneIsInternal verifies a key
// collision is rejected regardless of which sibling is Internal — a single
// user-provided value must never be able to both reject as
// internal-field-set and apply as a real override, depending on which of
// two colliding fields Decode happens to process first.
func TestValidate_duplicateKey_rejectedEvenWhenOneIsInternal(t *testing.T) {
	c := validateInternalKeyCollisionConfig{
		Secret: configfield.Field[string]{Key: "format", Default: "hidden", Internal: true},
		Format: configfield.Field[string]{Key: "format", Default: "standard"},
	}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateUnsupportedTypeConfig struct {
	Ch configfield.Field[chan int]
}

// TestValidate_unsupportedFieldType_rejected verifies a field type
// codec/mapstructure can never decode (a channel) is rejected up front,
// rather than failing later with a confusing low-level error.
func TestValidate_unsupportedFieldType_rejected(t *testing.T) {
	c := validateUnsupportedTypeConfig{}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// plainElementFixture has no Field members — safe as a map/slice/array
// element, decodable directly by ordinary mapstructure.
type plainElementFixture struct {
	Detect string
}

// fieldElementFixture has a Field member — unsafe as a map/slice/array
// element, since mapstructure would decode into it directly with no idea
// Field[T] exists.
type fieldElementFixture struct {
	Detect configfield.Field[string]
}

type validateMapOfPlainStructConfig struct {
	PerFile configfield.Field[map[string]plainElementFixture]
}

// TestValidate_mapOfPlainStruct_passes verifies a map whose value type is a
// plain (non-Field) struct is accepted — the realistic pattern for
// user-authored, dynamically-keyed records like per-file overrides.
func TestValidate_mapOfPlainStruct_passes(t *testing.T) {
	c := validateMapOfPlainStructConfig{}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateMapOfFieldStructConfig struct {
	PerFile configfield.Field[map[string]fieldElementFixture]
}

// TestValidate_mapOfFieldStruct_rejected verifies a map whose value type
// has a Field buried inside it is rejected — mapstructure would corrupt it.
func TestValidate_mapOfFieldStruct_rejected(t *testing.T) {
	c := validateMapOfFieldStructConfig{}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateSliceOfPlainStructConfig struct {
	Legacy configfield.Field[[]plainElementFixture]
}

// TestValidate_sliceOfPlainStruct_passes verifies a slice whose element
// type is a plain (non-Field) struct is accepted — the realistic pattern
// for a user-authored list of records.
func TestValidate_sliceOfPlainStruct_passes(t *testing.T) {
	c := validateSliceOfPlainStructConfig{}
	if err := configfield.Validate(&c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type validateSliceOfFieldStructConfig struct {
	Legacy configfield.Field[[]fieldElementFixture]
}

// TestValidate_sliceOfFieldStruct_rejected verifies a slice whose element
// type has a Field buried inside it is rejected.
func TestValidate_sliceOfFieldStruct_rejected(t *testing.T) {
	c := validateSliceOfFieldStructConfig{}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type validateArrayOfFieldStructConfig struct {
	Fixed configfield.Field[[2]fieldElementFixture]
}

// TestValidate_arrayOfFieldStruct_rejected verifies a fixed-size array
// whose element type has a Field buried inside it is rejected too.
func TestValidate_arrayOfFieldStruct_rejected(t *testing.T) {
	c := validateArrayOfFieldStructConfig{}
	if err := configfield.Validate(&c); err == nil {
		t.Fatal("expected error, got nil")
	}
}
