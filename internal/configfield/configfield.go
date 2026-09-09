// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package configfield decodes and validates a typed config value without
// knowing its concrete type.
package configfield

import (
	"io/fs"

	"github.com/leaflockio/core-cli/internal/store"
)

// Validatable requires a Validate method that reports whether a config
// value is internally consistent.
type Validatable interface {
	Validate() error
}

// ConfigLoader decodes a config section and validates the result.
type ConfigLoader interface {
	Load(section map[string]any) error
	Validate() error
}

// CommandConfig is a ConfigLoader that decodes into and validates Dest. PT
// is constrained to *T and Validatable, so every instantiation is required,
// at compile time, to implement Validate.
type CommandConfig[T any, PT interface {
	*T
	Validatable
}] struct {
	Dest PT
}

// Load requires Dest to be shaped entirely from Field values, then decodes
// section into it. A nil section leaves Dest unchanged.
func (c CommandConfig[T, PT]) Load(section map[string]any) error {
	if err := Validate(c.Dest); err != nil {
		return err
	}
	return Decode(section, c.Dest)
}

// Validate calls Dest's own Validate.
func (c CommandConfig[T, PT]) Validate() error {
	return c.Dest.Validate()
}

// Save flattens Dest and writes it to base.
func (c CommandConfig[T, PT]) Save(base string, dirPerm, filePerm fs.FileMode) error {
	m, err := Flatten(c.Dest)
	if err != nil {
		return err
	}
	return store.Save(base, m, dirPerm, filePerm)
}
