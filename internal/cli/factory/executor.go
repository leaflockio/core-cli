// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"fmt"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/spf13/cobra"
)

// execute wires a cobra.Command from an assembled plan, including its children.
func (f *Factory) execute(plan *assembled, a *app.App) (*cobra.Command, error) {
	def := plan.def
	meta := def.Meta

	cmd := &cobra.Command{
		Use:     meta.Use,
		Short:   meta.Short,
		Long:    meta.Long,
		GroupID: string(def.Group),
		Args:    meta.Args,
	}

	for _, fp := range plan.flags {
		fp.register(cmd.Flags(), a)
	}

	// Nil Handler means this command only holds children — leave RunE nil so
	// cobra's own non-Runnable fallback shows help.
	if def.Handler != nil {
		cmd.RunE = func(cobraCmd *cobra.Command, args []string) error {
			fs := cobraCmd.Flags()

			for _, fp := range plan.flags {
				if fp.hasResolver {
					if err := fp.resolve(fs); err != nil {
						return fmt.Errorf("factory[execute]: command %q: flag %q: %w", def.Meta.Use, fp.name, err)
					}
				}
			}

			return def.Handler(a, args)
		}
	}

	for _, child := range def.Children {
		_, childCmd, err := f.buildNode(child, a, false)
		if err != nil {
			return nil, fmt.Errorf("factory[execute]: command %q: child %w", def.Meta.Use, err)
		}
		cmd.AddCommand(childCmd)
	}

	return cmd, nil
}
