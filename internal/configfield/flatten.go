// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield

import (
	"fmt"
	"reflect"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/util/codec"
)

// Flatten produces a map[string]any from src, a struct or a pointer to
// one.
func Flatten(src any) (map[string]any, error) {
	if src == nil {
		return nil, errs.Unexpected(fmt.Errorf("configfield[flatten]: %w", errNilValue))
	}

	rv := reflect.ValueOf(src)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, errs.Unexpected(fmt.Errorf("configfield[flatten]: %w", errNilValue))
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errs.Unexpected(fmt.Errorf("configfield[flatten]: %w", errNotStruct))
	}

	return flattenStruct(rv)
}

// flattenStruct builds the map for every exported field of rv.
func flattenStruct(rv reflect.Value) (map[string]any, error) {
	m := make(map[string]any, rv.NumField())
	for sf, fv := range rv.Fields() {
		if !sf.IsExported() {
			continue
		}
		if err := flattenField(m, &sf, fv); err != nil {
			return nil, fmt.Errorf("%s.%w", sf.Name, err)
		}
	}
	return m, nil
}

// flattenField writes fv's resolved value into m under its key, unless
// the field is Internal.
func flattenField(m map[string]any, sf *reflect.StructField, fv reflect.Value) error {
	f, ok := fv.Interface().(field)
	if !ok {
		return errs.Unexpected(fmt.Errorf("configfield[flatten]: field %q: %w", sf.Name, errFieldNotWrapped))
	}
	if f.internal() {
		return nil
	}
	encoded, err := flattenValue(f.value())
	if err != nil {
		return err
	}
	m[fieldKey(sf, f)] = encoded
	return nil
}

// flattenValue encodes val, a Field[T]'s resolved value, into a form safe
// for a config map.
func flattenValue(val any) (any, error) {
	valRv := reflect.ValueOf(val)
	if valRv.Kind() == reflect.Struct {
		return flattenStruct(valRv)
	}
	if fieldTypedElem(valRv.Type()) {
		return flattenElements(valRv)
	}
	return codec.EncodeValue(val)
}

// flattenElements flattens rv, a string-keyed map or slice whose element
// type is Field-shaped, one element at a time via flattenStruct.
func flattenElements(rv reflect.Value) (any, error) {
	if rv.Kind() == reflect.Slice {
		return flattenElementSlice(rv)
	}
	return flattenElementMap(rv)
}

// flattenElementSlice flattens rv, a slice whose element type is
// Field-shaped, into a plain slice of nested maps.
func flattenElementSlice(rv reflect.Value) (any, error) {
	out := make([]any, rv.Len())
	for i := range rv.Len() {
		nested, err := flattenStruct(rv.Index(i))
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}
		out[i] = nested
	}
	return out, nil
}

// flattenElementMap flattens rv, a string-keyed map whose value type is
// Field-shaped, into a plain map of nested maps.
func flattenElementMap(rv reflect.Value) (any, error) {
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		nested, err := flattenStruct(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("%q: %w", iter.Key(), err)
		}
		out[fmt.Sprint(iter.Key().Interface())] = nested
	}
	return out, nil
}
