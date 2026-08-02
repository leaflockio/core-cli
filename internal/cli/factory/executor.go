// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/spf13/cobra"
)

// execute wires a cobra.Command from an assembled plan, including its children.
func (f *Factory) execute(plan *assembled, a *app.App) (*cobra.Command, error) {
	def := plan.def
	meta := def.Meta

	cmd := &cobra.Command{
		Use:                meta.Use,
		Short:              meta.Short,
		Long:               meta.Long,
		GroupID:            string(def.Group),
		Args:               meta.Args,
		DisableSuggestions: meta.DisableSuggestions,
	}

	// A command declares the group taxonomy for its own children via
	// Groups.
	for _, g := range def.Groups {
		cmd.AddGroup(&cobra.Group{ID: string(g.ID), Title: strings.ToUpper(g.Title)})
	}

	for _, fp := range plan.flags {
		fp.register(cmd.Flags(), a)
	}

	if def.Handler != nil {
		cmd.RunE = func(cobraCmd *cobra.Command, args []string) error {
			fs := cobraCmd.Flags()

			for _, rule := range def.FlagRules {
				if err := rule.Check(fs.Changed); err != nil {
					return err
				}
			}

			for _, fp := range plan.flags {
				if fp.hasResolver {
					if err := fp.resolve(fs); err != nil {
						return fmt.Errorf("factory[execute]: command %q: flag %q: %w", def.Meta.Use, fp.name, err)
					}
				}
			}

			return def.Handler(a, cobraCmd, args)
		}
	} else {
		// A nil Handler means this command only holds children. Giving it a
		// RunE makes it Runnable, so cobra's ValidateArgs — using this
		// command's own Meta.Args — actually runs and rejects an
		// unrecognized subcommand name, at any depth in the tree.
		cmd.RunE = func(cobraCmd *cobra.Command, _ []string) error {
			return cobraCmd.Help()
		}
	}

	for _, child := range def.Children {
		_, childCmd, err := f.buildNode(child, a, plan.level.Next())
		if err != nil {
			return nil, fmt.Errorf("factory[execute]: command %q: child %w", def.Meta.Use, err)
		}
		cmd.AddCommand(childCmd)
	}

	return cmd, nil
}
