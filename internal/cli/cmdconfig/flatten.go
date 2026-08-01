// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig

import (
	"fmt"
	"reflect"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/util/codec"
)

// Flatten produces a map[string]any from src, a struct or a pointer to
// one. For every Field[T] (except ones marked Internal, which are
// skipped), it emits one key mapped to its resolved Value. When a resolved
// value is itself a struct, it's flattened into a nested map rather than
// assigned directly.
func Flatten(src any) (map[string]any, error) {
	if src == nil {
		return nil, errs.Unexpected(fmt.Errorf("cmdconfig[flatten]: %w", errNilValue))
	}

	rv := reflect.ValueOf(src)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, errs.Unexpected(fmt.Errorf("cmdconfig[flatten]: %w", errNilValue))
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errs.Unexpected(fmt.Errorf("cmdconfig[flatten]: %w", errNotStruct))
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

// flattenField writes fv's resolved value into m under its key, unless the
// field is Internal. A resolved value that's itself a struct is flattened
// into a nested map rather than assigned directly. Anything else — a
// scalar, or a slice/map that may itself contain a plain struct — goes
// through codec.EncodeValue, so a struct nested inside a slice or map is
// keyed by its mapstructure tags rather than its raw Go field names.
func flattenField(m map[string]any, sf *reflect.StructField, fv reflect.Value) error {
	f, ok := fv.Interface().(field)
	if !ok {
		return errs.Unexpected(fmt.Errorf("cmdconfig[flatten]: field %q: %w", sf.Name, errFieldNotWrapped))
	}
	if f.internal() {
		return nil
	}
	key := fieldKey(sf, f)

	val := f.value()
	valRv := reflect.ValueOf(val)
	if valRv.Kind() == reflect.Struct {
		nested, err := flattenStruct(valRv)
		if err != nil {
			return err
		}
		m[key] = nested
		return nil
	}
	encoded, err := codec.EncodeValue(val)
	if err != nil {
		return err
	}
	m[key] = encoded
	return nil
}
