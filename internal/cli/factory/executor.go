// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

var errUndeclaredGroup = errors.New("group is not declared in the parent's Groups")

// cobraUse builds cobra's combined Use string from meta.Use (the bare name)
// and meta.ArgsUsage (an optional positional-argument hint).
func cobraUse(meta *cli.Meta) string {
	if meta.ArgsUsage == "" {
		return meta.Use
	}
	return meta.Use + " " + meta.ArgsUsage
}

// execute wires a cobra.Command from plan, including its children.
func (f *Factory) execute(plan *blueprint, a *app.App, hooks *[]hookRecord) (*cobra.Command, error) {
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

	declaredGroups := make(map[cli.GroupID]bool, len(def.Groups))
	for _, g := range def.Groups {
		cmd.AddGroup(&cobra.Group{ID: string(g.ID), Title: strings.ToUpper(g.Title)})
		declaredGroups[g.ID] = true
	}

	for _, fp := range plan.flags {
		fp.register(cmd.Flags(), a)
	}

	cmd.RunE = buildRunE(&def, plan, a)

	for _, child := range def.Children {
		childPlan, childCmd, err := f.buildNode(child, a, plan.level.Next(), hooks)
		if err != nil {
			return nil, fmt.Errorf("factory[execute]: command %q: child %w", def.Meta.Use, err)
		}
		if err := checkChildGroup(&def, &childPlan.def, declaredGroups); err != nil {
			return nil, err
		}
		cmd.AddCommand(childCmd)
	}

	return cmd, nil
}

// checkChildGroup requires childDef's own Group to be either empty or a
// member of declared — the set derived from parentDef's own Groups, the
// exact set execute() registers on parentDef's built command, so a child's
// GroupID is always backed by a real registration. Checked here, right
// after buildNode has already built childDef for real, rather than via a
// separate peek at Definition before any child is built — a second
// Define() call per child that assemble() used to make solely for this
// check.
func checkChildGroup(parentDef, childDef *cli.Definition, declared map[cli.GroupID]bool) error {
	if childDef.Group == "" || declared[childDef.Group] {
		return nil
	}
	msg := fmt.Sprintf("child %q: group %q is not declared in %q's Groups",
		childDef.Meta.Use, childDef.Group, parentDef.Meta.Use)
	return errs.Unexpected(
		fmt.Errorf("factory[execute]: command %q: %s: %w", parentDef.Meta.Use, msg, errUndeclaredGroup),
		errs.Context{
			Cause:      fmt.Sprintf("command %q: %s", parentDef.Meta.Use, msg),
			Resolution: "add the group to the parent's WithGroups, or remove it from the child's WithGroup",
		},
	)
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
