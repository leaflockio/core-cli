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
	errFieldTypedElement = errors.New("map and slice elements cannot contain a Field")
)

// unsupportedKinds are T kinds Field[T] rejects outright — codec/
// mapstructure can never decode into or encode most of them meaningfully.
var unsupportedKinds = map[reflect.Kind]bool{
	reflect.Chan:          true,
	reflect.Func:          true,
	reflect.UnsafePointer: true,
	reflect.Complex64:     true,
	reflect.Complex128:    true,
	reflect.Interface:     true,
	reflect.Array:         true,
}

// containerKinds are T kinds whose element type needs checking for a
// buried Field.
var containerKinds = map[reflect.Kind]bool{
	reflect.Map:   true,
	reflect.Slice: true,
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

// Validate reports whether v, a struct or pointer to one, is correctly
// shaped for Decode.
func Validate(v any) error {
	if v == nil {
		return errs.Unexpected(fmt.Errorf("configfield[validate]: %w", errNilValue))
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return errs.Unexpected(fmt.Errorf("configfield[validate]: %w", errNilValue))
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errs.Unexpected(fmt.Errorf("configfield[validate]: %s: %w", rv.Type(), errNotStruct))
	}

	if err := validateStruct(rv); err != nil {
		return errs.Unexpected(fmt.Errorf("configfield[validate]: %w", err))
	}
	return nil
}

// validateStruct checks every exported field of rv and collects every
// violation found.
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

// validateField requires fv to be a decodable Field[T].
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
		if isFieldShaped(elem) {
			if !fieldTypedElem(vt) {
				return nil, fmt.Errorf("%s: %w", vt, errFieldTypedElement)
			}
			if err := validateStruct(reflect.New(elem).Elem()); err != nil {
				return nil, err
			}
		}
	}
	if vt.Kind() == reflect.Struct {
		if err := validateStruct(reflect.ValueOf(f.defaultValue())); err != nil {
			return nil, err
		}
	}
	return f, nil
}
