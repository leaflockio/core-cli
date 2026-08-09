// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/noconfig"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/spf13/cobra"
)

var (
	errNilCmd       = errors.New("factory[build]: cmd must not be nil")
	errNilApp       = errors.New("factory[build]: a must not be nil")
	errNilWorkspace = errors.New("factory[build]: a.Workspace must not be nil")
)

// Factory builds and wires commands for execution.
type Factory struct{}

// New returns a new Factory.
func New() *Factory {
	return &Factory{}
}

// Build wires cmd's full command tree and registers implicit system flags
// on the root's PersistentFlags. Call it once, with the top-level command
// — descendants are reached via Definition.Children.
func (f *Factory) Build(cmd cli.Command, a *app.App) (*cobra.Command, error) {
	if cmd == nil {
		return nil, errs.Unexpected(errNilCmd, errs.Context{
			Cause:      "Factory.Build was called with a nil cli.Command",
			Resolution: "pass the root command to Build, never nil",
		})
	}
	if a == nil {
		return nil, errs.Unexpected(errNilApp, errs.Context{
			Cause:      "Factory.Build was called with a nil *app.App",
			Resolution: "construct a real *app.App before calling Build",
		})
	}
	if a.Workspace == nil {
		return nil, errs.Unexpected(errNilWorkspace, errs.Context{
			Cause:      "the *app.App passed to Factory.Build has no Workspace set",
			Resolution: "call WithWorkspace when constructing the *app.App",
		})
	}

	if !noConfigRequested(a) {
		layout, err := checkConfigLayout(cmd, a)
		if err != nil {
			return nil, err
		}
		if err := loadInvokedConfig(cmd, a, layout); err != nil {
			return nil, err
		}
	}

	var hooks []hookRecord
	plan, cobraCmd, err := f.buildNode(cmd, a, level.LevelRoot, &hooks)
	if err != nil {
		return nil, err
	}

	registerPersistent(cobraCmd, plan.persistentFlags, a)
	fireTreeReady(hooks, cobraCmd)

	return cobraCmd, nil
}

// noConfigRequested reports whether --no-config (or its shorthand, if one
// is ever registered) was passed. Checked directly against a's raw
// invocation, not cobra's parsed flags — this runs before cobra parses
// anything (see the noconfig package doc for why).
func noConfigRequested(a *app.App) bool {
	if a.Invocation == nil {
		return false
	}
	meta := noconfig.NoConfig.Definition().Meta
	fm := a.Invocation.FlagMap()
	if _, ok := fm[meta.LongFlag()]; ok {
		return true
	}
	if short := meta.ShortFlag(); short != "" {
		if _, ok := fm[short]; ok {
			return true
		}
	}
	return false
}

// buildNode assembles and wires cmd, recursing into children via itself.
func (f *Factory) buildNode(
	cmd cli.Command, a *app.App, lvl level.Level, hooks *[]hookRecord,
) (*blueprint, *cobra.Command, error) {
	def := cmd.Define(a)

	plan, err := assemble(def, lvl)
	if err != nil {
		return nil, nil, err
	}

	cobraCmd, err := f.execute(plan, a, hooks)
	if err != nil {
		return nil, nil, err
	}

	discoverHooks(cmd, cobraCmd, hooks)

	return plan, cobraCmd, nil
}

// registerPersistent registers persistentFlags on cmd's PersistentFlags.
func registerPersistent(cmd *cobra.Command, persistentFlags []flagSpec, a *app.App) {
	for _, fp := range persistentFlags {
		fp.register(cmd.PersistentFlags(), a)
	}
}
