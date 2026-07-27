// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package cmdconfig decodes and validates a command's config value without
// knowing its concrete type.
package cmdconfig

import "github.com/leaflockio/core-cli/internal/util/codec"

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

// CommandConfig binds a command's config struct as its Definition's config
// declaration. PT is constrained to *T and Validatable, so every command's
// config type is required, at compile time, to implement Validate.
type CommandConfig[T any, PT interface {
	*T
	Validatable
}] struct {
	Dest PT
}

// Load decodes section into Dest. A nil section leaves Dest unchanged.
func (c CommandConfig[T, PT]) Load(section map[string]any) error {
	return codec.DecodeMap(section, c.Dest)
}

// Validate calls Dest's own Validate.
func (c CommandConfig[T, PT]) Validate() error {
	return c.Dest.Validate()
}
