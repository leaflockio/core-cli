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

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/leaflock/core-cli/internal/cli/flags"
	"github.com/spf13/pflag"
)

var (
	errHandlerNil        = errors.New("handler must not be nil unless the command declares children")
	errImplicitFlagInDef = errors.New("must not appear in Definition.Flags; it is added by the engine automatically")
	errDuplicateFlags    = errors.New("duplicate flag declarations")
	errInvalidFlagType   = errors.New("must be a CommandFlag or SystemFlag")
)

// assemble validates def and produces an assembled plan.
func assemble(def *cli.Definition, isAppRoot bool) (*assembled, error) {
	if !isAppRoot && def.Handler == nil && len(def.Children) == 0 {
		return nil, fmt.Errorf("factory[assemble]: command %q: %w", def.Meta.Use, errHandlerNil)
	}

	if err := validateNoDuplicateFlags(def); err != nil {
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

	definedFlagPlans := make([]assembledFlag, 0, len(def.Flags))
	for _, f := range def.Flags {
		d := f.Definition()
		if d.Meta.Sub == flags.SubImplicit {
			return nil, fmt.Errorf(
				"factory[assemble]: command %q: implicit system flag %q: %w",
				def.Meta.Use, d.Meta.Name, errImplicitFlagInDef,
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
		hasFlags:        len(def.Flags) > 0,
		hasChildren:     len(def.Children) > 0,
		flags:           definedFlagPlans,
		persistentFlags: implicitFlagPlans,
	}, nil
}

// validateNoDuplicateFlags counts every flag name and shorthand across implicit
// and definition flags in one pass, then reports all violations together.
func validateNoDuplicateFlags(def *cli.Definition) error {
	nameCount := map[string]int{}
	shortCount := map[string]int{}

	all := make([]flags.Flag, 0, len(implicitSystemFlags)+len(def.Flags))
	all = append(all, implicitSystemFlags...)
	all = append(all, def.Flags...)

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
		return fmt.Errorf(
			"factory[assemble]: command %q: %s: %w",
			def.Meta.Use, strings.Join(msgs, "; "), errDuplicateFlags,
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
	return assembledFlag{}, fmt.Errorf("flag %q: unrecognized type %T: %w", meta.Name, f, errInvalidFlagType)
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
