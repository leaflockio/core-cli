// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// Bool returns a BoolValue with the given name and usage.
func Bool(name, usage string) *BoolValue {
	return &BoolValue{
		Meta: Meta{Name: name, Usage: usage},
	}
}

// BoolValue is a boolean flag value descriptor.
type BoolValue struct {
	Meta
	defaultVal bool
	dest       *bool
}

func (f *BoolValue) meta() Meta { return f.Meta }

// Validate performs no checks — bool flags carry no user-supplied text.
func (f *BoolValue) Validate() error { return nil }

// Default returns the default value.
func (f *BoolValue) Default() bool { return f.defaultVal }

// Dest returns the destination pointer.
func (f *BoolValue) Dest() *bool { return f.dest }

// WithShorthand sets the shorthand character and returns the receiver.
func (f *BoolValue) WithShorthand(s string) *BoolValue {
	f.Shorthand = s
	return f
}

// WithDefault sets the default value and returns the receiver.
func (f *BoolValue) WithDefault(val bool) *BoolValue {
	f.defaultVal = val
	return f
}

// WithDest returns a copy of f with dest bound as its destination pointer.
// Copying rather than mutating f in place means any two callers sharing the
// same underlying Value never have one caller's binding silently
// overwritten by another's.
func (f *BoolValue) WithDest(dest *bool) *BoolValue {
	c := *f
	c.dest = dest
	return &c
}

// BoolResolver is implemented by flags whose underlying type is bool
// but write a custom destination type after parsing.
type BoolResolver interface {
	Resolver
	Resolve(raw bool) error
}
