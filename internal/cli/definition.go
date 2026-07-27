// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/paths"
	"github.com/spf13/cobra"
)

// Definition is the complete declaration of a command — what it is, what it
// accepts, what protects it, what it exposes, and what it does.
type Definition struct {
	// Meta holds the command's identity — Use, Short, Long, Args.
	Meta Meta

	// Group is the display group for this command. Defaults to GroupCLI.
	Group GroupID

	// Flags declares the flags this command accepts.
	Flags []flags.Flag

	// PathRegistry declares the paths this command registers.
	PathRegistry *paths.Registry

	// Config declares this command's config type.
	Config cmdconfig.ConfigLoader

	// Handler is the command's execution logic.
	Handler func(a *app.App, cmd *cobra.Command, args []string) error

	// Children declares the subcommands nested under this command.
	Children []Command
}

// NewDefinition returns a Definition with Group defaulting to GroupCLI.
// Use the WithX methods to override individual fields.
func NewDefinition(meta Meta) *Definition {
	return &Definition{
		Meta:  meta,
		Group: GroupCLI,
	}
}

// WithGroup sets Group and returns the receiver.
func (d *Definition) WithGroup(group GroupID) *Definition {
	d.Group = group
	return d
}

// WithFlags sets Flags and returns the receiver.
func (d *Definition) WithFlags(f []flags.Flag) *Definition {
	d.Flags = f
	return d
}

// WithPathRegistry sets PathRegistry and returns the receiver.
func (d *Definition) WithPathRegistry(r *paths.Registry) *Definition {
	d.PathRegistry = r
	return d
}

// WithConfig sets Config and returns the receiver.
func (d *Definition) WithConfig(c cmdconfig.ConfigLoader) *Definition {
	d.Config = c
	return d
}

// WithHandler sets Handler and returns the receiver.
func (d *Definition) WithHandler(h func(a *app.App, cmd *cobra.Command, args []string) error) *Definition {
	d.Handler = h
	return d
}

// WithChildren sets Children and returns the receiver.
func (d *Definition) WithChildren(children []Command) *Definition {
	d.Children = children
	return d
}
