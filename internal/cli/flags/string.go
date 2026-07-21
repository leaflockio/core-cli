// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errNullByte    = errors.New("value must not contain null bytes")
	errControlChar = errors.New("value must not contain control characters")
)

// String returns a StringValue with the given name and usage.
func String(name, usage string) *StringValue {
	return &StringValue{
		Meta: Meta{Name: name, Usage: usage},
	}
}

// StringValue is a string flag value descriptor.
type StringValue struct {
	Meta
	defaultVal string
	dest       *string
}

func (f *StringValue) meta() Meta { return f.Meta }

// Validate checks that the default value contains no null bytes or control characters.
func (f *StringValue) Validate() error {
	return validateString(f.Name, f.defaultVal)
}

// Default returns the default value.
func (f *StringValue) Default() string { return f.defaultVal }

// Dest returns the destination pointer.
func (f *StringValue) Dest() *string { return f.dest }

// WithShorthand sets the shorthand character and returns the receiver.
func (f *StringValue) WithShorthand(s string) *StringValue {
	f.Shorthand = s
	return f
}

// WithDefault sets the default value and returns the receiver.
func (f *StringValue) WithDefault(val string) *StringValue {
	f.defaultVal = val
	return f
}

// WithDest binds the destination pointer and returns the receiver.
func (f *StringValue) WithDest(dest *string) *StringValue {
	f.dest = dest
	return f
}

// StringResolver is implemented by flags whose underlying type is string
// but write a custom destination type after parsing.
type StringResolver interface {
	Resolver
	Resolve(raw string) error
}

// validateString guards programmer-defined flag defaults against null bytes and
// ASCII control characters (code points 0x00–0x1F), with the exception of
// horizontal tab (0x09). It does not run on user-supplied runtime input.
//
// Examples of rejected defaults:
//
//	"hello\x00world"  → null byte (0x00) — terminates C strings, corrupts output
//	"line1\nline2"    → newline (0x0A) — breaks help text rendering
//
// Examples of accepted defaults:
//
//	"MIT"             → plain ASCII — typical license identifier default
//	"path/to/file"    → slashes are fine
//	"column\theader"  → tab (0x09) is explicitly allowed
func validateString(name, s string) error {
	if strings.ContainsRune(s, 0) {
		return fmt.Errorf("flag %q: %w", name, errNullByte)
	}
	for _, r := range s {
		if r < 0x20 && r != '\t' {
			return fmt.Errorf("flag %q: %w", name, errControlChar)
		}
	}
	return nil
}
