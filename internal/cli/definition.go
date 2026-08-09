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
	Meta *Meta

	// Group is which of the parent's Groups this command belongs to.
	// Validated against the parent Definition's own Groups — never this
	// Definition's. Empty by default, meaning the command renders ungrouped.
	Group GroupID

	// Groups declares the groups this Definition's own children may set
	// their Group to. Unrelated to this Definition's own Group. Empty by
	// default, meaning children render ungrouped.
	Groups []Group

	// Flags declares the flags this command accepts.
	Flags []flags.Flag

	// FlagRules declares relationships between flags, Checked after flags
	// are parsed, before Handler runs. Every flag a Rule references must
	// also be declared in Flags; FlagRules never registers a flag on its own.
	FlagRules []flags.Rule

	// PathRegistry declares the paths this command registers.
	PathRegistry *paths.Registry

	// Config declares this command's config type.
	Config cmdconfig.ConfigLoader

	// Handler is the command's execution logic.
	Handler func(a *app.App, cmd *cobra.Command, args []string) error

	// Children declares the subcommands nested under this command.
	Children []Command
}

// NewDefinition returns a Definition with Meta set. Use the WithX methods
// to set any other fields.
func NewDefinition(meta *Meta) *Definition {
	return &Definition{
		Meta: meta,
	}
}

// WithGroup sets Group and returns the receiver.
func (d *Definition) WithGroup(group GroupID) *Definition {
	d.Group = group
	return d
}

// WithGroups sets Groups and returns the receiver.
func (d *Definition) WithGroups(groups []Group) *Definition {
	d.Groups = groups
	return d
}

// WithFlags sets Flags and returns the receiver.
func (d *Definition) WithFlags(f []flags.Flag) *Definition {
	d.Flags = f
	return d
}

// WithFlagRules sets FlagRules and returns the receiver.
func (d *Definition) WithFlagRules(rules []flags.Rule) *Definition {
	d.FlagRules = rules
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
