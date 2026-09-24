// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/leaflockio/core-cli/internal/errs"
)

var (
	errNotPointer       = errors.New("dest must be a non-nil pointer to a struct")
	errInternalFieldSet = errors.New("field is internal and cannot be set from config")
	errUnknownKey       = errors.New("unrecognized key")
	errNotASequence     = errors.New("expected a sequence")
)

// Decode reads section into dest, a non-nil pointer to a struct that must
// already hold its default values.
func Decode(section map[string]any, dest any) error {
	if dest == nil {
		return errs.Unexpected(fmt.Errorf("configfield[decode]: %w", errNilValue))
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errs.Unexpected(fmt.Errorf("configfield[decode]: %w", errNotPointer))
	}
	if rv.Elem().Kind() != reflect.Struct {
		return errs.Unexpected(fmt.Errorf("configfield[decode]: %w", errNotStruct))
	}
	return decodeStruct(section, rv.Elem())
}

// decodeStructValue asserts raw is a map[string]any and decodes it into
// dest via decodeStruct.
func decodeStructValue(raw any, dest reflect.Value) error {
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return errNotAMapping
	}
	return decodeStruct(rawMap, dest)
}

// decodeStruct decodes every exported field of rv from section.
func decodeStruct(section map[string]any, rv reflect.Value) error {
	var violations []error
	known := make(map[string]bool, len(section))
	for sf, fv := range rv.Fields() {
		if !sf.IsExported() {
			continue
		}
		key, err := decodeField(section, &sf, fv)
		if key != "" {
			known[key] = true
		}
		if err != nil {
			violations = append(violations, fmt.Errorf("%s.%w", sf.Name, err))
		}
	}
	for _, key := range slices.Sorted(maps.Keys(section)) {
		if !known[key] {
			violations = append(violations, fmt.Errorf("%q: %w", key, errUnknownKey))
		}
	}
	return errors.Join(violations...)
}

// decodeField records fv's override from section, if fv's key is present.
func decodeField(section map[string]any, sf *reflect.StructField, fv reflect.Value) (string, error) {
	f, ok := fv.Interface().(field)
	if !ok {
		return "", errs.Unexpected(fmt.Errorf("configfield[decode]: field %q: %w", sf.Name, errFieldNotWrapped))
	}
	key := fieldKey(sf, f)
	raw, present := section[key]
	if !present {
		return key, nil
	}
	if f.internal() {
		return key, fmt.Errorf("%q: %w", key, errInternalFieldSet)
	}
	setter, ok := fv.Addr().Interface().(fieldSetter)
	if !ok {
		return key, errs.Unexpected(fmt.Errorf("configfield[decode]: field %q: %w", sf.Name, errFieldNotWrapped))
	}
	if err := setter.setOverride(raw); err != nil {
		return key, fmt.Errorf("%q: %w", key, err)
	}
	return key, nil
}

// decodeElements decodes raw into dest, a string-keyed map or slice whose
// element type is Field-shaped.
func decodeElements(raw any, dest reflect.Value) error {
	if dest.Kind() == reflect.Slice {
		return decodeElementSlice(raw, dest)
	}
	if dest.Kind() == reflect.Map {
		return decodeElementMap(raw, dest)
	}
	return nil
}

// decodeElement decodes raw into a fresh elemType instance.
func decodeElement(elemType reflect.Type, label string, raw any) (reflect.Value, error) {
	itemMap, ok := raw.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("%s: %w", label, errNotAMapping)
	}
	elem := reflect.New(elemType).Elem()
	if err := decodeStruct(itemMap, elem); err != nil {
		return reflect.Value{}, fmt.Errorf("%s: %w", label, err)
	}
	return elem, nil
}

// decodeElementSlice decodes raw into dest, a slice whose element type is
// Field-shaped.
func decodeElementSlice(raw any, dest reflect.Value) error {
	items, ok := raw.([]any)
	if !ok {
		return errNotASequence
	}
	elemType := dest.Type().Elem()
	var violations []error
	out := reflect.MakeSlice(dest.Type(), len(items), len(items))
	for i, item := range items {
		elem, err := decodeElement(elemType, fmt.Sprintf("[%d]", i), item)
		if err != nil {
			violations = append(violations, err)
			continue
		}
		out.Index(i).Set(elem)
	}
	if len(violations) > 0 {
		return errors.Join(violations...)
	}
	dest.Set(out)
	return nil
}

// decodeElementMap decodes raw into dest, a string-keyed map whose value
// type is Field-shaped.
func decodeElementMap(raw any, dest reflect.Value) error {
	items, ok := raw.(map[string]any)
	if !ok {
		return errNotAMapping
	}
	elemType := dest.Type().Elem()
	var violations []error
	out := reflect.MakeMapWithSize(dest.Type(), len(items))
	for _, key := range slices.Sorted(maps.Keys(items)) {
		elem, err := decodeElement(elemType, fmt.Sprintf("%q", key), items[key])
		if err != nil {
			violations = append(violations, err)
			continue
		}
		out.SetMapIndex(reflect.ValueOf(key), elem)
	}
	if len(violations) > 0 {
		return errors.Join(violations...)
	}
	dest.Set(out)
	return nil
}
