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

	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
)

var (
	errHandlerNil        = errors.New("handler must not be nil unless the command declares children")
	errConfigNotTopLevel = errors.New("config and path registry may only be declared by a top-level command")
	errImplicitFlagInDef = errors.New("must not appear in Definition.Flags; it is added by the engine automatically")
	errDuplicateFlags    = errors.New("duplicate flag declarations")
	errUseEmpty          = errors.New("use must not be empty")
	errUseHasWhitespace  = errors.New("use must be a bare command name, with no whitespace")
)

// assembleChecks are assemble's validation gates, ordered cheapest/most-
// local first.
var assembleChecks = []func(*cli.Definition, level.Level) error{
	validateMeta,
	checkHandlerOrChildren,
	checkConfigOwner,
	validateNoDuplicateFlags,
	checkFlagsNotImplicit,
}

// runChecks runs every assembleChecks gate against def, stopping at the
// first failure.
func runChecks(def *cli.Definition, lvl level.Level) error {
	for _, check := range assembleChecks {
		if err := check(def, lvl); err != nil {
			return err
		}
	}
	return nil
}

// assemble validates def and produces a blueprint.
func assemble(def *cli.Definition, lvl level.Level) (*blueprint, error) {
	if err := runChecks(def, lvl); err != nil {
		return nil, err
	}

	all := allFlags(def)

	implicitFlagPlans, err := planFlags(implicitSystemFlags, def.Meta.Use)
	if err != nil {
		return nil, err
	}
	definedFlagPlans, err := planFlags(all, def.Meta.Use)
	if err != nil {
		return nil, err
	}

	return &blueprint{
		def:             *def,
		level:           lvl,
		hasFlags:        len(all) > 0,
		hasChildren:     len(def.Children) > 0,
		flags:           definedFlagPlans,
		persistentFlags: implicitFlagPlans,
	}, nil
}

func planFlags(fs []flags.Flag, use string) ([]flagSpec, error) {
	specs := make([]flagSpec, 0, len(fs))
	for _, f := range fs {
		p, err := planFlag(f)
		if err != nil {
			return nil, fmt.Errorf("factory[assemble]: command %q: %w", use, err)
		}
		specs = append(specs, p)
	}
	return specs, nil
}

// checkHandlerOrChildren requires def to declare a Handler or at least one
// child, except at the tree root — a handlerless, childless leaf command
// would never do anything when invoked.
func checkHandlerOrChildren(def *cli.Definition, lvl level.Level) error {
	if lvl == level.LevelRoot || def.Handler != nil || len(def.Children) > 0 {
		return nil
	}
	return errs.Unexpected(
		fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, errHandlerNil),
		errs.Context{
			Cause:      fmt.Sprintf("command %q declares neither a Handler nor Children", def.Meta.Use),
			Resolution: "add a Handler, or declare at least one child command",
		},
	)
}

// checkConfigOwner requires Config and PathRegistry to be declared only by
// a top-level command — the one level factory.Build ever loads config for.
func checkConfigOwner(def *cli.Definition, lvl level.Level) error {
	if lvl == level.LevelTop || (def.Config == nil && def.PathRegistry.IsEmpty()) {
		return nil
	}
	return errs.Unexpected(
		fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, errConfigNotTopLevel),
		errs.Context{
			Cause:      fmt.Sprintf("command %q declares Config or PathRegistry but isn't top-level", def.Meta.Use),
			Resolution: "only a direct child of the root command may declare Config or PathRegistry",
		},
	)
}

// checkFlagsNotImplicit rejects any flag in def's own declared Flags marked
// SubImplicit — that subcategory is reserved for implicitSystemFlags, added
// automatically by the engine, never by a command's own Definition.
func checkFlagsNotImplicit(def *cli.Definition, _ level.Level) error {
	for _, f := range allFlags(def) {
		d := f.Definition()
		if d.Meta.Sub != flags.SubImplicit {
			continue
		}
		return errs.Unexpected(
			fmt.Errorf(
				"factory[assemble]: command %q: implicit system flag %q: %w",
				def.Meta.Use, d.Meta.Name, errImplicitFlagInDef,
			),
			errs.Context{
				Cause: fmt.Sprintf(
					"command %q declared implicit system flag %q in Definition.Flags",
					def.Meta.Use, d.Meta.Name,
				),
				Resolution: "remove it — implicit system flags are added automatically by the engine",
			},
		)
	}
	return nil
}

// allFlags returns every flag def registers.
func allFlags(def *cli.Definition) []flags.Flag {
	return def.Flags
}

// validateMeta checks a command's Meta for authoring mistakes.
func validateMeta(def *cli.Definition, _ level.Level) error {
	meta := def.Meta
	if meta.Use == "" {
		return errs.Unexpected(errUseEmpty, errs.Context{
			Cause:      "a command's Meta.Use was empty",
			Resolution: "every command must declare a non-empty Use — its bare name",
		})
	}
	if strings.ContainsAny(meta.Use, " \t") {
		return errs.Unexpected(
			fmt.Errorf("factory[assemble]: command %q: %w", meta.Use, errUseHasWhitespace),
			errs.Context{
				Cause: fmt.Sprintf("command %q's Use contains whitespace", meta.Use),
				Resolution: "Use must be the bare command name only — put a positional-argument " +
					`hint in ArgsUsage instead, e.g. WithArgsUsage("[file...]")`,
			},
		)
	}
	return nil
}

// validateNoDuplicateFlags counts every flag name and shorthand across implicit
// and definition flags in one pass, then reports all violations together.
func validateNoDuplicateFlags(def *cli.Definition, _ level.Level) error {
	nameCount := map[string]int{}
	shortCount := map[string]int{}

	own := allFlags(def)
	all := make([]flags.Flag, 0, len(implicitSystemFlags)+len(own))
	all = append(all, implicitSystemFlags...)
	all = append(all, own...)

	for _, f := range all {
		m := f.Definition().Meta
		nameCount[m.Name]++
		if m.Shorthand != "" {
			shortCount[m.Shorthand]++
		}
	}

	reported := map[string]bool{}
	var msgs []string
	for _, f := range all {
		m := f.Definition().Meta
		if nameCount[m.Name] > 1 && !reported["n:"+m.Name] {
			msgs = append(msgs, fmt.Sprintf("flag %q is declared %d times", m.Name, nameCount[m.Name]))
			reported["n:"+m.Name] = true
		}
		if m.Shorthand != "" && shortCount[m.Shorthand] > 1 && !reported["s:"+m.Shorthand] {
			msgs = append(msgs, fmt.Sprintf("shorthand %q is used by multiple flags", m.Shorthand))
			reported["s:"+m.Shorthand] = true
		}
	}
	if len(msgs) > 0 {
		return errs.Unexpected(
			fmt.Errorf(
				"factory[assemble]: command %q: %s: %w",
				def.Meta.Use, strings.Join(msgs, "; "), errDuplicateFlags,
			),
			errs.Context{
				Cause:      fmt.Sprintf("command %q: %s", def.Meta.Use, strings.Join(msgs, "; ")),
				Resolution: "rename or remove the duplicate flag declarations",
			},
		)
	}
	return nil
}
