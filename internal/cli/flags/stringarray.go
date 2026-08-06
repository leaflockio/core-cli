// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

// StringSlice returns a StringSliceValue with the given name and usage.
func StringSlice(name, usage string) *StringSliceValue {
	return &StringSliceValue{
		Meta: Meta{Name: name, Usage: usage},
	}
}

// StringSliceValue is a repeatable string flag value descriptor.
type StringSliceValue struct {
	Meta
	defaultVal []string
	dest       *[]string
}

func (f *StringSliceValue) meta() Meta { return f.Meta }

// Validate checks that none of the default elements contain null bytes or control characters.
func (f *StringSliceValue) Validate() error {
	for _, v := range f.defaultVal {
		if err := validateString(f.Name, v); err != nil {
			return err
		}
	}
	return nil
}

// Default returns the default value.
func (f *StringSliceValue) Default() []string { return f.defaultVal }

// Dest returns the destination pointer.
func (f *StringSliceValue) Dest() *[]string { return f.dest }

// WithShorthand sets the shorthand character and returns the receiver.
func (f *StringSliceValue) WithShorthand(s string) *StringSliceValue {
	f.Shorthand = s
	return f
}

// WithDefault sets the default value and returns the receiver.
func (f *StringSliceValue) WithDefault(val []string) *StringSliceValue {
	f.defaultVal = val
	return f
}

// WithDest returns a copy of f with dest bound as its destination pointer.
// Copying rather than mutating f in place means any two callers sharing the
// same underlying Value never have one caller's binding silently
// overwritten by another's.
func (f *StringSliceValue) WithDest(dest *[]string) *StringSliceValue {
	c := *f
	c.dest = dest
	return &c
}

// StringSliceResolver is implemented by flags whose underlying type is []string
// but write a custom destination type after parsing.
type StringSliceResolver interface {
	Resolver
	Resolve(raw []string) error
}
