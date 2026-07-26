// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package codec_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/util/codec"
)

// fixture is the shared test struct used across all encode/decode tests.
type fixture struct {
	Name      string        `mapstructure:"name"`
	Count     int           `mapstructure:"count"`
	Active    bool          `mapstructure:"active"`
	TTL       time.Duration `mapstructure:"ttl"`
	CreatedAt time.Time     `mapstructure:"created_at"`
}

var fixedTime = time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)

// TestMarshal_JSON verifies compact JSON output for the default Marshal call.
func TestMarshal_JSON(t *testing.T) {
	v := fixture{Name: "leaf", Count: 3, Active: true}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "\n") {
		t.Errorf("expected compact JSON (no newlines), got:\n%s", got)
	}
	if !strings.Contains(got, `"name":"leaf"`) {
		t.Errorf("missing name field in output: %s", got)
	}
}

// TestMarshalWith_JSON_indent verifies indented JSON output.
func TestMarshalWith_JSON_indent(t *testing.T) {
	v := fixture{Name: "leaf", Count: 1}
	data, err := codec.MarshalWith(v, codec.JSON, codec.JSONOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("MarshalWith: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "\n") {
		t.Errorf("expected indented JSON (with newlines), got: %s", got)
	}
	if !strings.Contains(got, "  ") {
		t.Errorf("expected 2-space indent in output: %s", got)
	}
}

// TestMarshal_YAML verifies YAML output.
func TestMarshal_YAML(t *testing.T) {
	v := fixture{Name: "leaf", Count: 2}
	data, err := codec.Marshal(v, codec.YAML)
	if err != nil {
		t.Fatalf("Marshal YAML: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "name: leaf") {
		t.Errorf("missing name field in YAML output: %s", got)
	}
}

// TestMarshalWith_YAML_opts_has_no_effect verifies opts are ignored for YAML.
func TestMarshalWith_YAML_opts_has_no_effect(t *testing.T) {
	v := fixture{Name: "leaf"}
	withOpts, err := codec.MarshalWith(v, codec.YAML, codec.JSONOptions{Indent: "\t"})
	if err != nil {
		t.Fatalf("MarshalWith YAML with opts: %v", err)
	}
	without, err := codec.Marshal(v, codec.YAML)
	if err != nil {
		t.Fatalf("Marshal YAML: %v", err)
	}
	if !bytes.Equal(withOpts, without) {
		t.Errorf("YAML output should be identical regardless of JSONOptions")
	}
}

// TestMarshal_unsupported_format verifies an error is returned for unknown formats.
func TestMarshal_unsupported_format(t *testing.T) {
	_, err := codec.Marshal(fixture{}, codec.Format("toml"))
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !errors.Is(err, codec.ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got: %v", err)
	}
}

// TestUnmarshal_JSON round-trips a struct through JSON.
func TestUnmarshal_JSON(t *testing.T) {
	original := fixture{Name: "leaf", Count: 5, Active: true, TTL: 2 * time.Hour}
	data, err := codec.Marshal(original, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got fixture
	if err := codec.Unmarshal(data, codec.JSON, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Count != original.Count {
		t.Errorf("Count = %d, want %d", got.Count, original.Count)
	}
	if got.Active != original.Active {
		t.Errorf("Active = %v, want %v", got.Active, original.Active)
	}
	if got.TTL != original.TTL {
		t.Errorf("TTL = %v, want %v", got.TTL, original.TTL)
	}
}

// TestUnmarshal_YAML round-trips a struct through YAML.
func TestUnmarshal_YAML(t *testing.T) {
	original := fixture{Name: "leaf", Count: 7}
	data, err := codec.Marshal(original, codec.YAML)
	if err != nil {
		t.Fatalf("Marshal YAML: %v", err)
	}
	var got fixture
	if err := codec.Unmarshal(data, codec.YAML, &got); err != nil {
		t.Fatalf("Unmarshal YAML: %v", err)
	}
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Count != original.Count {
		t.Errorf("Count = %d, want %d", got.Count, original.Count)
	}
}

// TestUnmarshal_JSONC decodes JSON with line and block comments plus a
// trailing comma, none of which are valid in plain JSON.
func TestUnmarshal_JSONC(t *testing.T) {
	data := []byte(`{
  // line comment
  "name": "leaf",
  /* block comment */
  "count": 5,
  "active": true, // trailing comma below
}`)
	var got fixture
	if err := codec.Unmarshal(data, codec.JSONC, &got); err != nil {
		t.Fatalf("Unmarshal JSONC: %v", err)
	}
	if got.Name != "leaf" {
		t.Errorf("Name = %q, want %q", got.Name, "leaf")
	}
	if got.Count != 5 {
		t.Errorf("Count = %d, want %d", got.Count, 5)
	}
	if !got.Active {
		t.Error("Active = false, want true")
	}
}

// TestUnmarshal_JSONC_invalid verifies an error is returned for content that
// is invalid even once comments/trailing commas are stripped.
func TestUnmarshal_JSONC_invalid(t *testing.T) {
	err := codec.Unmarshal([]byte(`{"name": }`), codec.JSONC, &fixture{})
	if err == nil {
		t.Fatal("expected error for invalid JSONC, got nil")
	}
}

// TestMarshal_JSONC verifies Marshal produces plain JSON output for JSONC —
// valid JSON is always valid JSONC, so no special encoding is needed.
func TestMarshal_JSONC(t *testing.T) {
	data, err := codec.Marshal(fixture{Name: "leaf"}, codec.JSONC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), `"name":"leaf"`) {
		t.Errorf("missing name field in output: %s", data)
	}

	var got fixture
	if err := codec.Unmarshal(data, codec.JSONC, &got); err != nil {
		t.Fatalf("round-trip Unmarshal: %v", err)
	}
	if got.Name != "leaf" {
		t.Errorf("Name = %q, want %q", got.Name, "leaf")
	}
}

// TestUnmarshal_unsupported_format verifies an error is returned for unknown formats.
func TestUnmarshal_unsupported_format(t *testing.T) {
	err := codec.Unmarshal([]byte(`{}`), codec.Format("toml"), &fixture{})
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !errors.Is(err, codec.ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got: %v", err)
	}
}

// TestUnmarshal_JSON_invalid verifies an error is returned for malformed JSON.
func TestUnmarshal_JSON_invalid(t *testing.T) {
	err := codec.Unmarshal([]byte(`not json`), codec.JSON, &fixture{})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestUnmarshal_YAML_invalid verifies an error is returned for malformed YAML.
func TestUnmarshal_YAML_invalid(t *testing.T) {
	err := codec.Unmarshal([]byte(":\tbroken: [yaml"), codec.YAML, &fixture{})
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// TestTimeField_JSON verifies time.Time fields round-trip correctly in JSON.
func TestTimeField_JSON(t *testing.T) {
	original := fixture{CreatedAt: fixedTime}
	data, err := codec.Marshal(original, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), "2026-06-23T12:00:00Z") {
		t.Errorf("expected RFC3339 time in output, got: %s", data)
	}
	var got fixture
	if err := codec.Unmarshal(data, codec.JSON, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !got.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, fixedTime)
	}
}

// TestTimeField_YAML verifies time.Time fields round-trip correctly in YAML.
func TestTimeField_YAML(t *testing.T) {
	original := fixture{CreatedAt: fixedTime}
	data, err := codec.Marshal(original, codec.YAML)
	if err != nil {
		t.Fatalf("Marshal YAML: %v", err)
	}
	var got fixture
	if err := codec.Unmarshal(data, codec.YAML, &got); err != nil {
		t.Fatalf("Unmarshal YAML: %v", err)
	}
	if !got.CreatedAt.Equal(fixedTime) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, fixedTime)
	}
}

// TestTimeField_zero verifies a zero time.Time is omitted from JSON output.
func TestTimeField_zero(t *testing.T) {
	v := fixture{Name: "leaf"}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(data), "created_at") {
		t.Errorf("zero time.Time should be omitted from output, got: %s", data)
	}
}

// TestMarshal_non_struct verifies an error is returned when v is not a struct.
func TestMarshal_non_struct(t *testing.T) {
	_, err := codec.Marshal("not a struct", codec.JSON)
	if err == nil {
		t.Fatal("expected error for non-struct input, got nil")
	}
}

// TestMarshal_pointer_to_struct verifies a pointer to struct is accepted.
func TestMarshal_pointer_to_struct(t *testing.T) {
	v := &fixture{Name: "leaf", Count: 1}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal pointer: %v", err)
	}
	if !strings.Contains(string(data), `"name":"leaf"`) {
		t.Errorf("expected name in output, got: %s", data)
	}
}

// tagFallbackFixture has a field with an empty name part in the mapstructure tag,
// exercising the name="" fallback to lowercase field name.
type tagFallbackFixture struct {
	Value   string `mapstructure:",omitempty"`
	Ignored string `mapstructure:"-"`
}

// TestMarshal_tag_name_fallback verifies fields with empty tag names fall back
// to the lowercase field name, and fields tagged "-" are omitted.
func TestMarshal_tag_name_fallback(t *testing.T) {
	v := tagFallbackFixture{Value: "hello", Ignored: "secret"}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), `"value":"hello"`) {
		t.Errorf("expected lowercase field name fallback, got: %s", data)
	}
	if strings.Contains(string(data), "secret") {
		t.Errorf(`field tagged "-" should be omitted, got: %s`, data)
	}
}

// TestMarshalWith_unsupported_format verifies MarshalWith returns error for unknown formats.
func TestMarshalWith_unsupported_format(t *testing.T) {
	_, err := codec.MarshalWith(fixture{}, codec.Format("toml"), codec.JSONOptions{})
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !errors.Is(err, codec.ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got: %v", err)
	}
}

// ptrFixture has a pointer field to exercise the nil pointer path in encodeValue.
type ptrFixture struct {
	Name    string     `mapstructure:"name"`
	Nested  *fixture   `mapstructure:"nested"`
	NilTime *time.Time `mapstructure:"nil_time"`
}

// TestMarshal_nil_pointer verifies nil pointer fields are omitted from output.
func TestMarshal_nil_pointer(t *testing.T) {
	v := ptrFixture{Name: "leaf", Nested: nil}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(data), "nested") {
		t.Errorf("nil pointer field should be omitted, got: %s", data)
	}
}

// TestMarshal_non_nil_pointer verifies non-nil pointer fields are encoded.
func TestMarshal_non_nil_pointer(t *testing.T) {
	inner := &fixture{Name: "inner", Count: 1}
	v := ptrFixture{Name: "leaf", Nested: inner}
	data, err := codec.Marshal(v, codec.JSON)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), "nested") {
		t.Errorf("non-nil pointer field should be present, got: %s", data)
	}
	if !strings.Contains(string(data), "inner") {
		t.Errorf("nested struct name should be present, got: %s", data)
	}
}
