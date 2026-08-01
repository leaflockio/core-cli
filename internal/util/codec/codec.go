// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package codec provides format-agnostic encode/decode helpers for structs
// that carry mapstructure tags. JSON, YAML, and JSONC are supported through
// a shared map[string]any intermediate: a struct is first converted to a
// map via mapstructure (preserving tag names as keys), and the map is then
// encoded to the target format. Decoding reverses the process. An
// already-built map, or a single value in place of a whole struct's fields,
// may be given directly wherever a struct is otherwise expected, skipping
// the struct-specific conversion step.
//
// Structs must use mapstructure tags to control field names:
//
//	type Example struct {
//	    Name      string    `mapstructure:"name"`
//	    CreatedAt time.Time `mapstructure:"created_at"`
//	}
//
// time.Time values are encoded as RFC3339 strings and decoded back
// transparently. Duration fields are decoded from strings (e.g. "24h").
package codec

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/tailscale/hujson"
	"go.yaml.in/yaml/v3"
)

var timeType = reflect.TypeFor[time.Time]()

// Sentinel errors.
var (
	// ErrUnsupportedFormat is returned when a format other than JSON or YAML is given.
	ErrUnsupportedFormat = errors.New("codec: unsupported format")
	// ErrNotStruct is returned when the value passed to Marshal is neither a
	// struct nor an already-built map[string]any.
	ErrNotStruct = errors.New("codec: value is not a struct")

	// Internal sentinel returned by encodeValue to signal that a field should
	// be omitted from the output (zero time.Time, nil pointer).
	errOmit = errors.New("omit")
)

// Format identifies the serialization format.
type Format string

const (
	// JSON encodes and decodes using encoding/json.
	JSON Format = "json"
	// YAML encodes and decodes using go.yaml.in/yaml/v3.
	YAML Format = "yaml"
	// JSONC decodes using hujson, then encoding/json; encodes identically to JSON.
	JSONC Format = "jsonc"
)

// JSONOptions controls JSON-specific encoding behavior.
// The zero value produces compact output with no indentation.
type JSONOptions struct {
	// Indent is the string used for each indentation level.
	// When empty, compact output is produced. Typical values: "  " (2 spaces)
	// or "\t" (tab).
	Indent string
}

// marshalJSON encodes m using compact or indented output based on opts.Indent.
func (o JSONOptions) marshalJSON(m map[string]any) ([]byte, error) {
	if o.Indent != "" {
		return json.MarshalIndent(m, "", o.Indent)
	}
	return json.Marshal(m)
}

// Marshal encodes v into the given format with compact output for JSON.
// YAML indentation is handled by the YAML encoder and is not configurable.
// The v parameter must be a struct or pointer to struct with mapstructure
// tags, or an already-built map[string]any, which is encoded as-is. Fields
// of type time.Time are written as RFC3339.
func Marshal(v any, format Format) ([]byte, error) {
	return MarshalWith(v, format, JSONOptions{})
}

// MarshalWith encodes v into the given format, applying opts for JSON output.
// When opts.Indent is set, JSON output is pretty-printed with that indent
// string. The opts parameter has no effect on YAML output.
func MarshalWith(v any, format Format, opts JSONOptions) ([]byte, error) {
	m, err := structToMap(v)
	if err != nil {
		return nil, fmt.Errorf("codec: marshal: %w", err)
	}
	switch format {
	case JSON, JSONC:
		out, err := opts.marshalJSON(m)
		if err != nil {
			return nil, fmt.Errorf("codec: marshal json: %w", err)
		}
		return out, nil
	case YAML:
		out, err := yaml.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("codec: marshal yaml: %w", err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}
}

// Unmarshal decodes data in the given format into v. The v parameter must be
// a non-nil pointer to a struct with mapstructure tags. RFC3339 strings are
// decoded back into time.Time fields automatically.
func Unmarshal(data []byte, format Format, v any) error {
	m, err := DecodeBytes(data, format)
	if err != nil {
		return err
	}
	return DecodeMap(m, v)
}

// DecodeBytes decodes data in the given format into a map[string]any.
func DecodeBytes(data []byte, format Format) (map[string]any, error) {
	var m map[string]any
	switch format {
	case JSON:
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("codec: unmarshal json: %w", err)
		}
	case YAML:
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("codec: unmarshal yaml: %w", err)
		}
	case JSONC:
		standard, err := hujson.Standardize(data)
		if err != nil {
			return nil, fmt.Errorf("codec: unmarshal jsonc: %w", err)
		}
		if err := json.Unmarshal(standard, &m); err != nil {
			return nil, fmt.Errorf("codec: unmarshal jsonc: %w", err)
		}
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}
	return m, nil
}

// DecodeMap populates v from a map[string]any, matching keys to
// mapstructure tags. The v parameter must be a non-nil pointer to a struct.
// RFC3339 strings are decoded back into time.Time fields automatically.
func DecodeMap(m map[string]any, v any) error {
	return decodeInto(m, v)
}

// DecodeValue populates dest from a single already-decoded value — a
// scalar, slice, or map. The dest parameter must be a non-nil pointer. An
// RFC3339 string decodes into a time.Time, and a duration string decodes
// into a time.Duration.
func DecodeValue(raw, dest any) error {
	return decodeInto(raw, dest)
}

// EncodeValue converts v — a scalar, slice, map, or struct with
// mapstructure tags, at any nesting depth — into a form safe to hand
// directly to json.Marshal or yaml.Marshal: every struct becomes a
// map[string]any keyed by its mapstructure tags, wherever it appears,
// including inside a slice or map. A time.Time becomes an RFC3339 string.
func EncodeValue(v any) (any, error) {
	encoded, err := encodeValue(reflect.ValueOf(v))
	if errors.Is(err, errOmit) {
		err = nil
	}
	return encoded, err
}

// structToMap converts a struct to map[string]any using mapstructure tag names
// as keys. Uses reflection so that time.Time fields are converted to RFC3339
// strings rather than being recursed into as plain structs (which mapstructure
// decoder would do, producing an empty map for time.Time). An already-built
// map[string]any is returned as-is, skipping struct conversion entirely.
func structToMap(v any) (map[string]any, error) {
	if m, ok := v.(map[string]any); ok {
		return m, nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: got %T", ErrNotStruct, v)
	}
	return encodeStruct(rv)
}

func encodeStruct(rv reflect.Value) (map[string]any, error) {
	rt := rv.Type()
	m := make(map[string]any, rt.NumField())
	for i := range rt.NumField() {
		field := rt.Field(i)
		fv := rv.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		encoded, err := encodeValue(fv)
		if errors.Is(err, errOmit) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("codec: field %q: %w", name, err)
		}
		m[name] = encoded
	}
	return m, nil
}

func encodeValue(rv reflect.Value) (any, error) {
	if !rv.IsValid() {
		return nil, errOmit
	}
	if rv.Type() == timeType {
		t, ok := rv.Interface().(time.Time)
		if !ok || t.IsZero() {
			return nil, errOmit
		}
		return t.UTC().Format(time.RFC3339), nil
	}
	kind := rv.Kind()
	if kind == reflect.Ptr {
		if rv.IsNil() {
			return nil, errOmit
		}
		return encodeValue(rv.Elem())
	}
	if kind == reflect.Struct {
		return encodeStruct(rv)
	}
	if kind == reflect.Slice || kind == reflect.Array {
		return encodeSequence(rv)
	}
	if kind == reflect.Map {
		return encodeMap(rv)
	}
	return rv.Interface(), nil
}

// encodeSequence encodes each element of a slice or array, so a struct
// element (at any position) goes through encodeStruct rather than being
// passed through with its raw Go field names.
func encodeSequence(rv reflect.Value) (any, error) {
	out := make([]any, rv.Len())
	for i := range rv.Len() {
		encoded, err := encodeValue(rv.Index(i))
		if errors.Is(err, errOmit) {
			out[i] = nil
			continue
		}
		if err != nil {
			return nil, err
		}
		out[i] = encoded
	}
	return out, nil
}

// encodeMap encodes each value of a map, so a struct value goes through
// encodeStruct rather than being passed through with its raw Go field
// names. Keys are converted to string via fmt.Sprint, matching how config
// maps are always string-keyed in practice.
func encodeMap(rv reflect.Value) (any, error) {
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		encoded, err := encodeValue(iter.Value())
		if errors.Is(err, errOmit) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out[fmt.Sprint(iter.Key().Interface())] = encoded
	}
	return out, nil
}

// decodeInto populates v from raw using mapstructure, decoding an RFC3339
// string into a time.Time and a duration string into a time.Duration.
func decodeInto(raw, v any) error {
	cfg := &mapstructure.DecoderConfig{
		Result: v,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			mapstructure.StringToTimeDurationHookFunc(),
		),
	}
	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return dec.Decode(raw)
}
