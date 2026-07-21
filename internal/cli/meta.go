// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli

import "github.com/spf13/cobra"

// Meta holds the identity of a command — its name, summary, and argument rules.
type Meta struct {
	Use   string
	Short string
	Long  string
	// Args is the positional argument validator. Defaults to cobra.NoArgs.
	Args cobra.PositionalArgs
}

// NewMeta returns a Meta with defaults applied. Fields that have defaults can
// be individually overridden using their corresponding WithX method on this type.
func NewMeta(use, short, long string) Meta {
	return Meta{
		Use:   use,
		Short: short,
		Long:  long,
		Args:  cobra.NoArgs,
	}
}

// WithArgs returns a copy of the Meta with Args replaced. Use this for
// commands that accept positional arguments.
func (m Meta) WithArgs(args cobra.PositionalArgs) Meta {
	m.Args = args
	return m
}
