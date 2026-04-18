// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	helpcmd "github.com/leaflock/core-cli/cmd/help"
	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/util/strutil"
	"github.com/spf13/cobra"
)

// root holds the assembled application and passes it down to subcommands.
type root struct {
	app *app.App
}

// cmd assembles and returns the root cobra command.
func (r *root) cmd() *cobra.Command {
	var f rootFlags
	cmd := r.baseCmd(&f)
	f.register(cmd)
	cmd.CompletionOptions.DisableDefaultCmd = true
	r.addGroups(cmd)
	r.addCommands(cmd)
	helpcmd.Set(cmd, r.app.Printer)
	return cmd
}

// baseCmd constructs the root cobra.Command with its metadata and lifecycle hooks.
func (r *root) baseCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:           config.AppName,
		Short:         "Org-wide toolchain for repo setup, sync, and validation",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if f.noColor {
				r.app.Printer.SetNoColor(true)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}

// addGroups registers command groups in display order and assigns the help command.
func (r *root) addGroups(cmd *cobra.Command) {
	for _, g := range groups {
		cmd.AddGroup(&cobra.Group{ID: g, Title: strutil.Title(g)})
	}
	cmd.SetHelpCommandGroupID(groupGeneral)
}

// addCommands registers all subcommands from the registry onto cmd.
func (r *root) addCommands(cmd *cobra.Command) {
	for _, e := range registry {
		c := e.factory(r.app)
		c.GroupID = e.groupID
		cmd.AddCommand(c)
	}
}
