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

// TestDecode_nilDest verifies an error is returned for a nil dest.
func TestDecode_nilDest(t *testing.T) {
	if err := configfield.Decode(map[string]any{}, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestDecode_nonPointerDest verifies an error is returned when dest isn't a
// pointer.
func TestDecode_nonPointerDest(t *testing.T) {
	type cfg struct {
		Format configfield.Field[string]
	}
	if err := configfield.Decode(map[string]any{}, cfg{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestDecode_nilPointerDest verifies an error is returned for a typed nil
// pointer.
func TestDecode_nilPointerDest(t *testing.T) {
	type cfg struct {
		Format configfield.Field[string]
	}
	var p *cfg
	if err := configfield.Decode(map[string]any{}, p); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestDecode_pointerToNonStruct verifies an error is returned when dest
// points to something other than a struct.
func TestDecode_pointerToNonStruct(t *testing.T) {
	n := 0
	if err := configfield.Decode(map[string]any{}, &n); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type decodeGoodConfig struct {
	Format   configfield.Field[string]
	unexport string
}

// TestDecode_success verifies a matching key overrides the field, and an
// unexported field is skipped without error.
func TestDecode_success(t *testing.T) {
	c := decodeGoodConfig{Format: configfield.Field[string]{Default: "standard"}, unexport: "private"}
	if err := configfield.Decode(map[string]any{"format": "spdx"}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if c.unexport != "private" {
		t.Errorf("unexport = %q, want unchanged %q", c.unexport, "private")
	}
	if got := c.Format.Value(); got != "spdx" {
		t.Errorf("Format.Value() = %q, want %q", got, "spdx")
	}
}

// TestDecode_missingKey_leavesDefault verifies a field whose key is absent
// from section stays at its Default.
func TestDecode_missingKey_leavesDefault(t *testing.T) {
	c := decodeGoodConfig{Format: configfield.Field[string]{Default: "standard"}}
	if err := configfield.Decode(map[string]any{}, &c); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got := c.Format.Value(); got != "standard" {
		t.Errorf("Format.Value() = %q, want %q (untouched)", got, "standard")
	}
}

type decodeInternalFieldConfig struct {
	Secret configfield.Field[string]
}

// TestDecode_internalField_isNotOverridden verifies a field marked Internal
// is never read from section, even when its key is present — and that its
// key being present at all is rejected, not silently ignored.
func TestDecode_internalField_isNotOverridden(t *testing.T) {
	c := decodeInternalFieldConfig{
		Secret: configfield.Field[string]{Default: "hidden", Internal: true},
	}
	err := configfield.Decode(map[string]any{"secret": "leaked"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := c.Secret.Value(); got != "hidden" {
		t.Errorf("Secret.Value() = %q, want %q (Internal field must not be overridden)", got, "hidden")
	}
}

type decodeUnwrappedFieldConfig struct {
	Bad string
}

// TestDecode_unwrappedField_errors verifies decoding a struct with a field
// that isn't a Field returns an error.
func TestDecode_unwrappedField_errors(t *testing.T) {
	err := configfield.Decode(map[string]any{"bad": "x"}, &decodeUnwrappedFieldConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type decodeSetOverrideErrorConfig struct {
	Count configfield.Field[int]
}

// TestDecode_setOverrideError_isPropagated verifies a field-level decode
// failure (wrong type for the destination) surfaces from Decode.
func TestDecode_setOverrideError_isPropagated(t *testing.T) {
	c := decodeSetOverrideErrorConfig{Count: configfield.Field[int]{Default: 1}}
	err := configfield.Decode(map[string]any{"count": "not-a-number"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type decodeUnknownKeyConfig struct {
	Format configfield.Field[string]
}

// TestDecode_unknownKey_rejected verifies a section key that doesn't match
// any field is rejected, rather than silently ignored — catching a typo or
// a stale, no-longer-valid setting.
func TestDecode_unknownKey_rejected(t *testing.T) {
	c := decodeUnknownKeyConfig{Format: configfield.Field[string]{Default: "standard"}}
	err := configfield.Decode(map[string]any{"bogus_key": "spdx"}, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := c.Format.Value(); got != "standard" {
		t.Errorf("Format.Value() = %q, want %q (unrecognized key must not apply)", got, "standard")
	}
}

// TestDecode_unknownKey_nested verifies an unrecognized key nested inside a
// struct-typed field's own section is also rejected.
func TestDecode_unknownKey_nested(t *testing.T) {
	c := decodeGoodNestedConfig{
		Nested: configfield.Field[decodeGoodNestedFixture]{
			Default: decodeGoodNestedFixture{A: configfield.Field[bool]{Default: true}},
		},
	}
	section := map[string]any{"nested": map[string]any{"a": false, "bogus": 1}}
	err := configfield.Decode(section, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type decodeGoodNestedFixture struct {
	A configfield.Field[bool]
}

type decodeGoodNestedConfig struct {
	Nested configfield.Field[decodeGoodNestedFixture]
}

type decodeMultipleBadFieldsConfig struct {
	Count configfield.Field[int]
	Total configfield.Field[int]
}

// TestDecode_collectsEveryFieldError verifies Decode reports every field's
// error, not just the first, so a user fixing several bad values in one
// config file sees all of them at once.
func TestDecode_collectsEveryFieldError(t *testing.T) {
	c := decodeMultipleBadFieldsConfig{
		Count: configfield.Field[int]{Default: 1},
		Total: configfield.Field[int]{Default: 1},
	}
	section := map[string]any{"count": "not-a-number", "total": "also-not-a-number"}
	err := configfield.Decode(section, &c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var joiner interface{ Unwrap() []error }
	if !errors.As(err, &joiner) {
		t.Fatalf("error chain does not contain a joined error: %v", err)
	}
	if got := len(joiner.Unwrap()); got != 2 {
		t.Errorf("collected %d field errors, want 2", got)
	}
}
