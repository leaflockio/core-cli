// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/leaflockio/core-cli/internal/util/codec"
)

var (
	errNotAnOption     = errors.New("value is not one of the field's allowed options")
	errOptionsOnStruct = errors.New("options is not supported on a struct-typed field")
)

// Option describes one value a Field accepts, paired with a description of
// when to choose it.
type Option[T any] struct {
	Value       T
	Description string
}

// Field is a single config value. Key, Default, Description, Example,
// Options, and Internal declare it; call Value for the effective value.
type Field[T any] struct {
	// Key is the field's decode/encode key. When empty, it falls back to
	// the struct field's own name converted to snake_case.
	Key string
	// Default is the value used when the field has not been overridden.
	Default T
	// Description documents the field for generated output.
	Description string
	// Example documents one valid value for generated output.
	Example string
	// Options lists the field's allowed values, each with its own
	// description. A nil or empty Options means the field is unrestricted;
	// otherwise Decode rejects any value not among them. Not supported on a
	// struct-typed field — rejected outright rather than attempted.
	Options []Option[T]
	// Internal marks a field as system-managed: never read from user config
	// (Decode skips it) and never written into a generated template
	// (Flatten skips it).
	Internal bool
	// Validate, when set, is called with a candidate value — Default during
	// Validate, or a decoded override during Decode — and rejects it by
	// returning a non-nil error. For constraints Options can't express
	// (ranges, formats, patterns) rather than a closed set of choices.
	Validate func(v T) error

	// overrideValue and overridden are set only by Decode.
	overrideValue T
	overridden    bool
}

// Value returns the field's override if it was explicitly set by user
// config, otherwise Default.
func (f Field[T]) Value() T {
	if f.overridden {
		return f.overrideValue
	}
	return f.Default
}

// isField marks every instantiation of Field[T], regardless of T, as
// satisfying the field interface.
func (Field[T]) isField() {}

// internal reports Internal, for the field interface.
func (f Field[T]) internal() bool { return f.Internal }

// value returns Value as any, for the field interface.
func (f Field[T]) value() any { return f.Value() }

// defaultValue returns Default as any, for the field interface.
func (f Field[T]) defaultValue() any { return f.Default }

// key returns Key, for the field interface.
func (f Field[T]) key() string { return f.Key }

// valueType returns T's reflect.Type, for the field interface. Boxing
// Default into any first (as defaultValue does) would lose T's static type
// whenever T is itself an interface, so this reports it directly instead.
func (f Field[T]) valueType() reflect.Type { return reflect.TypeFor[T]() }

// fieldSetter is implemented by *Field[T] for every T, so a field's
// override can be set generically, without knowing T.
type fieldSetter interface {
	setOverride(raw any) error
}

// Compile-time check that *Field[T] satisfies fieldSetter for every T.
var _ fieldSetter = &Field[struct{}]{}

// setOverride decodes raw into the field's override and marks it as set.
// When T is itself a struct, raw must be a map[string]any: the override is
// seeded from a copy of Default (so nested fields the user didn't configure
// keep their real defaults, not zero values) and then decoded field by
// field, recursively. The decoded value is then checked against Options and
// Validate, when set — a rejection leaves the field's override unset.
func (f *Field[T]) setOverride(raw any) error {
	if reflect.TypeFor[T]().Kind() == reflect.Struct {
		rawMap, ok := raw.(map[string]any)
		if !ok {
			return errNotAMapping
		}
		f.overrideValue = f.Default
		if err := decodeStruct(rawMap, reflect.ValueOf(&f.overrideValue).Elem()); err != nil {
			return err
		}
	} else if err := codec.DecodeValue(raw, &f.overrideValue); err != nil {
		return err
	}
	if err := f.checkValue(f.overrideValue); err != nil {
		return err
	}
	f.overridden = true
	return nil
}

// checkDefault requires Default to be valid — present among Options when
// Options is set, and accepted by Validate when Validate is set — for the
// field interface. Catches an author's own typo or a Default that fell out
// of sync with Options or Validate.
func (f Field[T]) checkDefault() error {
	return f.checkValue(f.Default)
}

// checkValue requires v to be present among Options when Options is
// non-empty, and accepted by Validate when Validate is set.
func (f Field[T]) checkValue(v T) error {
	if err := f.checkValueInOptions(v); err != nil {
		return err
	}
	if f.Validate != nil {
		return f.Validate(v)
	}
	return nil
}

// checkValueInOptions requires v to equal one of f.Options when Options is
// non-empty. T isn't constrained to comparable (it may be a slice or map),
// so equality is checked with reflect.DeepEqual — which is why Options on a
// struct-typed field is rejected outright rather than attempted: any such
// struct is Field-shaped all the way down, and DeepEqual never considers
// two non-nil func values equal, so a nested Validate func would make every
// comparison spuriously fail.
func (f Field[T]) checkValueInOptions(v T) error {
	if len(f.Options) == 0 {
		return nil
	}
	if reflect.TypeFor[T]().Kind() == reflect.Struct {
		return errOptionsOnStruct
	}
	for _, opt := range f.Options {
		if reflect.DeepEqual(v, opt.Value) {
			return nil
		}
	}
	return fmt.Errorf("%v: %w", v, errNotAnOption)
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// toSnakeCase converts a Go identifier such as UpdateLicenseFile to
// update_license_file, keeping acronym runs like TTL or SPDX intact
// (TTL stays ttl, OSIOnly becomes osi_only).
func toSnakeCase(s string) string {
	s = matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	s = matchAllCap.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

// fieldKey returns f's decode/encode key: its own Key if set, otherwise
// sf's name converted to snake_case.
func fieldKey(sf *reflect.StructField, f field) string {
	if k := f.key(); k != "" {
		return k
	}
	return toSnakeCase(sf.Name)
}
