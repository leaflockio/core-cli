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
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/spf13/pflag"
)

var (
	errHandlerNil        = errors.New("handler must not be nil unless the command declares children")
	errConfigNotTopLevel = errors.New("config and path registry may only be declared by a top-level command")
	errImplicitFlagInDef = errors.New("must not appear in Definition.Flags; it is added by the engine automatically")
	errDuplicateFlags    = errors.New("duplicate flag declarations")
	errInvalidFlagType   = errors.New("must be a CommandFlag or SystemFlag")
	errUndeclaredGroup   = errors.New("group is not declared in the parent's Groups")
	errUseEmpty          = errors.New("use must not be empty")
	errUseHasWhitespace  = errors.New("use must be a bare command name, with no whitespace")
)

// assemble validates def and produces an assembled plan, using a only to
// resolve each child's own Definition (via Define) for group validation —
// children are not built here, only peeked at.
func assemble(def *cli.Definition, lvl level.Level, a *app.App) (*assembled, error) {
	if err := validateMeta(def.Meta); err != nil {
		return nil, err
	}

	if lvl != level.LevelRoot && def.Handler == nil && len(def.Children) == 0 {
		return nil, errs.Unexpected(
			fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, errHandlerNil),
			errs.Context{
				Cause:      fmt.Sprintf("command %q declares neither a Handler nor Children", def.Meta.Use),
				Resolution: "add a Handler, or declare at least one child command",
			},
		)
	}

	if lvl != level.LevelTop && (def.Config != nil || !def.PathRegistry.IsEmpty()) {
		return nil, errs.Unexpected(
			fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, errConfigNotTopLevel),
			errs.Context{
				Cause:      fmt.Sprintf("command %q declares Config or PathRegistry but isn't top-level", def.Meta.Use),
				Resolution: "only a direct child of the root command may declare Config or PathRegistry",
			},
		)
	}

	if err := validateNoDuplicateFlags(def); err != nil {
		return nil, err
	}

	if err := validateChildGroups(def, a); err != nil {
		return nil, err
	}

	implicitFlagPlans := make([]assembledFlag, 0, len(implicitSystemFlags))
	for _, f := range implicitSystemFlags {
		p, err := planFlag(f)
		if err != nil {
			return nil, fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, err)
		}
		implicitFlagPlans = append(implicitFlagPlans, p)
	}

	all := allFlags(def)
	definedFlagPlans := make([]assembledFlag, 0, len(all))
	for _, f := range all {
		d := f.Definition()
		if d.Meta.Sub == flags.SubImplicit {
			return nil, errs.Unexpected(
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
		p, err := planFlag(f)
		if err != nil {
			return nil, fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, err)
		}
		definedFlagPlans = append(definedFlagPlans, p)
	}

	return &assembled{
		def:             *def,
		level:           lvl,
		hasFlags:        len(all) > 0,
		hasChildren:     len(def.Children) > 0,
		flags:           definedFlagPlans,
		persistentFlags: implicitFlagPlans,
	}, nil
}

// allFlags returns every flag def registers.
func allFlags(def *cli.Definition) []flags.Flag {
	return def.Flags
}

// validateMeta checks a command's Meta for authoring mistakes. Currently
// just Use's shape; a natural home for further Meta-level checks later.
//
// Use must be a non-empty, whitespace-free bare command name. Every
// framework lookup that matches on command identity (config-file names,
// invoked-command routing) compares against Use directly, so a usage
// pattern accidentally embedded in it (e.g. "check [file...]") silently
// breaks those lookups instead of erroring — put that in ArgsUsage via
// WithArgsUsage instead.
func validateMeta(meta *cli.Meta) error {
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

// validateChildGroups requires every child's own Group to be either empty or
// a member of def's own declared Groups — the exact set execute() registers
// on def's built command, so a child's GroupID is always backed by a real
// registration.
func validateChildGroups(def *cli.Definition, a *app.App) error {
	declared := make(map[cli.GroupID]bool, len(def.Groups))
	for _, g := range def.Groups {
		declared[g.ID] = true
	}

	msgs := make([]string, 0, len(def.Children))
	for _, child := range def.Children {
		childDef := child.Define(a)
		if childDef.Group == "" || declared[childDef.Group] {
			continue
		}
		msgs = append(msgs, fmt.Sprintf("child %q: group %q is not declared in %q's Groups",
			childDef.Meta.Use, childDef.Group, def.Meta.Use))
	}
	if len(msgs) == 0 {
		return nil
	}

	return errs.Unexpected(
		fmt.Errorf("factory[assemble]: command %q: %s: %w", def.Meta.Use, strings.Join(msgs, "; "), errUndeclaredGroup),
		errs.Context{
			Cause:      fmt.Sprintf("command %q: %s", def.Meta.Use, strings.Join(msgs, "; ")),
			Resolution: "add the group to the parent's WithGroups, or remove it from the child's WithGroup",
		},
	)
}

// validateNoDuplicateFlags counts every flag name and shorthand across implicit
// and definition flags in one pass, then reports all violations together.
func validateNoDuplicateFlags(def *cli.Definition) error {
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

// planFlag dispatches to the correct plan function based on the concrete flag type.
func planFlag(f flags.Flag) (assembledFlag, error) {
	switch v := f.(type) {
	case flags.SystemFlag[*flags.BoolValue]:
		return planBoolSystemFlag(v)
	case flags.SystemFlag[*flags.StringValue]:
		return planStringSystemFlag(v)
	case flags.SystemFlag[*flags.StringSliceValue]:
		return planStringSliceSystemFlag(v)
	case flags.CommandFlag[*flags.BoolValue]:
		if r, ok := v.Resolver.(flags.BoolResolver); ok {
			return planBoolResolverFlag(f, r)
		}
		return planBoolLiteralFlag(v)
	case flags.CommandFlag[*flags.StringValue]:
		if r, ok := v.Resolver.(flags.StringResolver); ok {
			return planStringResolverFlag(f, r)
		}
		return planStringLiteralFlag(v)
	case flags.CommandFlag[*flags.StringSliceValue]:
		if r, ok := v.Resolver.(flags.StringSliceResolver); ok {
			return planStringSliceResolverFlag(f, r)
		}
		return planStringSliceLiteralFlag(v)
	}
	meta := f.Definition().Meta
	return assembledFlag{}, errs.Unexpected(
		fmt.Errorf("flag %q: unrecognized type %T: %w", meta.Name, f, errInvalidFlagType),
		errs.Context{
			Cause:      fmt.Sprintf("flag %q has type %T, which is not a CommandFlag or SystemFlag", meta.Name, f),
			Resolution: "declare the flag using flags.CommandFlag or flags.SystemFlag",
		},
	)
}

// — System flags —

func planBoolSystemFlag(f flags.SystemFlag[*flags.BoolValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		sub:          meta.Sub,
		hasShorthand: meta.Shorthand != "",
		register: func(fs *pflag.FlagSet, a *app.App) {
			baseRegister(fs)
			wrapWithEffect(fs, name, func() { effect(a) })
		},
	}, nil
}

func planStringSystemFlag(f flags.SystemFlag[*flags.StringValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		sub:          meta.Sub,
		hasShorthand: meta.Shorthand != "",
		register: func(fs *pflag.FlagSet, a *app.App) {
			baseRegister(fs)
			wrapWithEffect(fs, name, func() { effect(a) })
		},
	}, nil
}

func planStringSliceSystemFlag(f flags.SystemFlag[*flags.StringSliceValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		sub:          meta.Sub,
		hasShorthand: meta.Shorthand != "",
		register: func(fs *pflag.FlagSet, a *app.App) {
			baseRegister(fs)
			wrapWithEffect(fs, name, func() { effect(a) })
		},
	}, nil
}

// — Literal command flags —

func planBoolLiteralFlag(f flags.CommandFlag[*flags.BoolValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

func planStringLiteralFlag(f flags.CommandFlag[*flags.StringValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

func planStringSliceLiteralFlag(f flags.CommandFlag[*flags.StringSliceValue]) (assembledFlag, error) {
	meta := f.Definition().Meta
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return assembledFlag{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

// — Resolver flags —

func planBoolResolverFlag(f flags.Flag, r flags.BoolResolver) (assembledFlag, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, false, nil)
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		hasResolver:  true,
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
		resolve: func(fs *pflag.FlagSet) error {
			raw, err := fs.GetBool(name)
			if err != nil {
				return err
			}
			return r.Resolve(raw)
		},
	}, nil
}

// planStringResolverFlag builds an assembledFlag for a string CommandFlag that
// carries a StringResolver.
func planStringResolverFlag(f flags.Flag, r flags.StringResolver) (assembledFlag, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, "", nil)
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		hasResolver:  true,
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
		resolve: func(fs *pflag.FlagSet) error {
			raw, err := fs.GetString(name)
			if err != nil {
				return err
			}
			return r.Resolve(raw)
		},
	}, nil
}

// planStringSliceResolverFlag builds an assembledFlag for a string-slice
// CommandFlag that carries a StringSliceResolver.
func planStringSliceResolverFlag(f flags.Flag, r flags.StringSliceResolver) (assembledFlag, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, nil, nil)
	return assembledFlag{
		name:         name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		hasResolver:  true,
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
		resolve: func(fs *pflag.FlagSet) error {
			raw, err := fs.GetStringArray(name)
			if err != nil {
				return err
			}
			return r.Resolve(raw)
		},
	}, nil
}
