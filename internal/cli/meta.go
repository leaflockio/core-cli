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
	// ArgsUsage is an optional cobra-style positional-argument hint shown in
	// help output, e.g. "[file...]". Empty when the command takes none.
	ArgsUsage string
	// Args is the positional argument validator. Defaults to a validator that
	// rejects an unrecognized subcommand name, suggesting close matches.
	Args cobra.PositionalArgs
	// DisableSuggestions turns off close-match suggestions for an
	// unrecognized subcommand name. Defaults to false.
	DisableSuggestions bool
}

// NewMeta returns a Meta with defaults applied. Fields that have defaults can
// be individually overridden using their corresponding WithX method on this type.
func NewMeta(use, short, long string) *Meta {
	return &Meta{
		Use:   use,
		Short: short,
		Long:  long,
		Args:  unknownCommand,
	}
}

// WithArgs returns a copy of the Meta with Args replaced. Use this for
// commands that accept positional arguments.
func (m *Meta) WithArgs(args cobra.PositionalArgs) *Meta {
	c := *m
	c.Args = args
	return &c
}

// WithArgsUsage returns a copy of the Meta with ArgsUsage set.
func (m *Meta) WithArgsUsage(pattern string) *Meta {
	c := *m
	c.ArgsUsage = pattern
	return &c
}

// WithDisableSuggestions returns a copy of the Meta with DisableSuggestions
// set to true.
func (m *Meta) WithDisableSuggestions() *Meta {
	c := *m
	c.DisableSuggestions = true
	return &c
}
