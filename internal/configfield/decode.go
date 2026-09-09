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
)

// Decode reads section into dest, a non-nil pointer to a struct that must
// already hold its default values. For every Field[T] whose key is present
// in section, Decode records the override so Value reflects it, unless the
// field is marked Internal, in which case its key being present at all is
// rejected. Every other Field[T] is left exactly as it already was —
// Decode never constructs a replacement Field[T], only mutates the ones
// already there. Any key in section that doesn't belong to any field is
// rejected too.
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

// decodeStruct decodes every exported field of rv from section, returning
// every field's error (each with a dotted path back to its field) and
// every key in section that doesn't belong to any field.
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

// decodeField records fv's override from section, if fv's key is present,
// and reports that key so decodeStruct can track which of section's keys
// were recognized. A missing key leaves fv untouched — it stays at
// Default. A field marked Internal rejects its key being present at all,
// rather than silently ignoring a value the user has no way of knowing was
// never applied.
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
