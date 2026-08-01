// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/leaflockio/core-cli/internal/errs"
)

// field is implemented by every instantiation of Field[T], regardless of T,
// so reflection-based code can work with one generically without knowing
// its T.
type field interface {
	isField()
	internal() bool
	value() any
	defaultValue() any
	key() string
	valueType() reflect.Type
	checkDefault() error
}

var fieldType = reflect.TypeFor[field]()

// Compile-time check that Field[T] satisfies field for every T.
var _ field = Field[struct{}]{}

var (
	errNotStruct         = errors.New("value must be a struct")
	errNilValue          = errors.New("value must not be nil")
	errFieldNotWrapped   = errors.New("field is not a Field")
	errNotAMapping       = errors.New("expected a mapping")
	errUnsupportedType   = errors.New("field type cannot be decoded")
	errDuplicateFieldKey = errors.New("duplicate key")
	errFieldTypedElement = errors.New("map, slice, and array elements cannot contain a Field")
)

// unsupportedKinds are T kinds that codec/mapstructure can never decode
// into or encode meaningfully, so Field[T] rejects them outright rather
// than letting them fail later with a confusing low-level error.
var unsupportedKinds = map[reflect.Kind]bool{
	reflect.Chan:          true,
	reflect.Func:          true,
	reflect.UnsafePointer: true,
	reflect.Complex64:     true,
	reflect.Complex128:    true,
	reflect.Interface:     true,
}

// containerKinds are T kinds whose element type needs checking for a
// buried Field — codec/mapstructure decodes into a map, slice, or array
// element via its own struct logic, with no idea Field[T] exists, so an
// element struct containing one would get silently corrupted rather than
// decoded correctly.
var containerKinds = map[reflect.Kind]bool{
	reflect.Map:   true,
	reflect.Slice: true,
	reflect.Array: true,
}

// hasFieldMember reports whether t — expected to be a struct — has any
// exported field implementing field. Only checks t's own direct fields, not
// fields nested further inside them.
func hasFieldMember(t reflect.Type) bool {
	for sf := range t.Fields() {
		if sf.IsExported() && sf.Type.Implements(fieldType) {
			return true
		}
	}
	return false
}

// Validate reports an error if any exported field, at any depth, of the
// given value's struct type is not a Field, has a type that can't be
// decoded, has a Default that isn't valid per its own Options or Validate,
// or shares its key with a sibling field. The value may be a struct or a
// pointer to one.
func Validate(v any) error {
	if v == nil {
		return errs.Unexpected(fmt.Errorf("cmdconfig[validate]: %w", errNilValue))
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return errs.Unexpected(fmt.Errorf("cmdconfig[validate]: %w", errNilValue))
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errs.Unexpected(fmt.Errorf("cmdconfig[validate]: %s: %w", rv.Type(), errNotStruct))
	}

	if err := validateStruct(rv); err != nil {
		return errs.Unexpected(fmt.Errorf("cmdconfig[validate]: %w", err))
	}
	return nil
}

// validateStruct checks every exported field of rv, collecting every
// violation found (each with a dotted path back to its field) rather than
// stopping at the first — including any two sibling fields that resolve to
// the same key, which would otherwise silently share one decoded value or
// have one overwrite the other when flattened.
func validateStruct(rv reflect.Value) error {
	var violations []error
	keyOwners := make(map[string]string)
	for sf, fv := range rv.Fields() {
		if !sf.IsExported() {
			continue
		}
		f, err := validateField(fv)
		if err != nil {
			violations = append(violations, fmt.Errorf("%s.%w", sf.Name, err))
			continue
		}
		key := fieldKey(&sf, f)
		if owner, dup := keyOwners[key]; dup {
			violations = append(violations,
				fmt.Errorf("%s: key %q: %w (already used by %s)", sf.Name, key, errDuplicateFieldKey, owner))
			continue
		}
		keyOwners[key] = sf.Name
	}
	return errors.Join(violations...)
}

// validateField requires fv to be a Field[T] for some T that's decodable
// and, if a map, slice, or array, has an element type with no Field buried
// inside it. Requires Default to be among its own Options (when Options is
// set), recursing into T's own Default when T is itself a struct. Returns
// the field interface value on success so validateStruct can check it for
// a duplicate key without asserting it a second time.
func validateField(fv reflect.Value) (field, error) {
	if !fv.Type().Implements(fieldType) {
		return nil, errFieldNotWrapped
	}
	f, ok := fv.Interface().(field)
	if !ok {
		return nil, errFieldNotWrapped
	}
	if err := f.checkDefault(); err != nil {
		return nil, err
	}
	vt := f.valueType()
	if unsupportedKinds[vt.Kind()] {
		return nil, fmt.Errorf("%s: %w", vt, errUnsupportedType)
	}
	if containerKinds[vt.Kind()] {
		elem := vt.Elem()
		if elem.Kind() == reflect.Struct && hasFieldMember(elem) {
			return nil, fmt.Errorf("%s: %w", vt, errFieldTypedElement)
		}
	}
	if vt.Kind() == reflect.Struct {
		if err := validateStruct(reflect.ValueOf(f.defaultValue())); err != nil {
			return nil, err
		}
	}
	return f, nil
}
