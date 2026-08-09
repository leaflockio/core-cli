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

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/spf13/pflag"
)

var errInvalidFlagType = errors.New("must be a CommandFlag or SystemFlag")

// planFlag dispatches to the correct plan function based on the concrete flag type.
func planFlag(f flags.Flag) (flagSpec, error) {
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
	return flagSpec{}, errs.Unexpected(
		fmt.Errorf("flag %q: unrecognized type %T: %w", meta.Name, f, errInvalidFlagType),
		errs.Context{
			Cause:      fmt.Sprintf("flag %q has type %T, which is not a CommandFlag or SystemFlag", meta.Name, f),
			Resolution: "declare the flag using flags.CommandFlag or flags.SystemFlag",
		},
	)
}

// — System flags —

func planBoolSystemFlag(f flags.SystemFlag[*flags.BoolValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
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

func planStringSystemFlag(f flags.SystemFlag[*flags.StringValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
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

func planStringSliceSystemFlag(f flags.SystemFlag[*flags.StringSliceValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	name, effect := meta.Name, f.Effect
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
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

func planBoolLiteralFlag(f flags.CommandFlag[*flags.BoolValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

func planStringLiteralFlag(f flags.CommandFlag[*flags.StringValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

func planStringSliceLiteralFlag(f flags.CommandFlag[*flags.StringSliceValue]) (flagSpec, error) {
	meta := f.Definition().Meta
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, f.Value.Default(), f.Value.Dest())
	return flagSpec{
		name:         meta.Name,
		kind:         meta.Kind,
		hasShorthand: meta.Shorthand != "",
		register:     func(fs *pflag.FlagSet, _ *app.App) { baseRegister(fs) },
	}, nil
}

// — Resolver flags —

func planBoolResolverFlag(f flags.Flag, r flags.BoolResolver) (flagSpec, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := boolRegistrar(meta.Name, meta.Shorthand, meta.Usage, false, nil)
	return flagSpec{
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

// planStringResolverFlag builds an flagSpec for a string CommandFlag that
// carries a StringResolver.
func planStringResolverFlag(f flags.Flag, r flags.StringResolver) (flagSpec, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := stringRegistrar(meta.Name, meta.Shorthand, meta.Usage, "", nil)
	return flagSpec{
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

// planStringSliceResolverFlag builds an flagSpec for a string-slice
// CommandFlag that carries a StringSliceResolver.
func planStringSliceResolverFlag(f flags.Flag, r flags.StringSliceResolver) (flagSpec, error) {
	meta := f.Definition().Meta
	name := meta.Name
	baseRegister := stringArrayRegistrar(meta.Name, meta.Shorthand, meta.Usage, nil, nil)
	return flagSpec{
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
