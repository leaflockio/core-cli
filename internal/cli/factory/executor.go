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
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

// cobraUse builds cobra's combined Use string from meta.Use (the bare name)
// and meta.ArgsUsage (an optional positional-argument hint).
func cobraUse(meta *cli.Meta) string {
	if meta.ArgsUsage == "" {
		return meta.Use
	}
	return meta.Use + " " + meta.ArgsUsage
}

// execute wires a cobra.Command from plan, including its children.
func (f *Factory) execute(plan *blueprint, a *app.App) (*cobra.Command, error) {
	def := plan.def
	meta := def.Meta

	cmd := &cobra.Command{
		Use:                cobraUse(meta),
		Short:              meta.Short,
		Long:               meta.Long,
		GroupID:            string(def.Group),
		Args:               meta.Args,
		DisableSuggestions: meta.DisableSuggestions,
	}

	for _, g := range def.Groups {
		cmd.AddGroup(&cobra.Group{ID: string(g.ID), Title: strings.ToUpper(g.Title)})
	}

	for _, fp := range plan.flags {
		fp.register(cmd.Flags(), a)
	}

	cmd.RunE = buildRunE(&def, plan, a)

	for _, child := range def.Children {
		_, childCmd, err := f.buildNode(child, a, plan.level.Next())
		if err != nil {
			return nil, fmt.Errorf("factory[execute]: command %q: child %w", def.Meta.Use, err)
		}
		cmd.AddCommand(childCmd)
	}

	return cmd, nil
}

func buildRunE(def *cli.Definition, plan *blueprint, a *app.App) func(*cobra.Command, []string) error {
	if def.Handler == nil {
		return func(cobraCmd *cobra.Command, _ []string) error {
			return cobraCmd.Help()
		}
	}

	return func(cobraCmd *cobra.Command, args []string) error {
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
}
