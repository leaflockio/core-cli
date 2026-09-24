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
func (f *Field[T]) setOverride(raw any) error {
	if err := f.decodeOverride(raw); err != nil {
		return err
	}
	if err := f.checkValue(f.overrideValue); err != nil {
		return err
	}
	f.overridden = true
	return nil
}

// decodeOverride decodes raw into f.overrideValue, dispatching by T's
// shape.
func (f *Field[T]) decodeOverride(raw any) error {
	t := reflect.TypeFor[T]()
	dest := reflect.ValueOf(&f.overrideValue).Elem()
	if t.Kind() == reflect.Struct {
		f.overrideValue = f.Default
		return decodeStructValue(raw, dest)
	}
	if fieldTypedElem(t) {
		return decodeElements(raw, dest)
	}
	return codec.DecodeValue(raw, &f.overrideValue)
}

// isFieldShaped reports whether t is a struct with at least one exported
// field implementing the field interface.
func isFieldShaped(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && hasFieldMember(t)
}

// fieldTypedElem reports whether t is a string-keyed map or slice whose
// element type is Field-shaped.
func fieldTypedElem(t reflect.Type) bool {
	isSlice := t.Kind() == reflect.Slice
	isStringKeyedMap := t.Kind() == reflect.Map && t.Key().Kind() == reflect.String
	if !isSlice && !isStringKeyedMap {
		return false
	}
	return isFieldShaped(t.Elem())
}

// checkDefault reports whether Default is valid according to Options and
// Validate.
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

// checkValueInOptions reports whether v is one of f.Options, when Options
// is non-empty.
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

// toSnakeCase converts a Go identifier to snake_case.
//
//	APIKey -> api_key
func toSnakeCase(s string) string {
	s = matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	s = matchAllCap.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

// fieldKey returns f's decode/encode key.
func fieldKey(sf *reflect.StructField, f field) string {
	if k := f.key(); k != "" {
		return k
	}
	return toSnakeCase(sf.Name)
}
