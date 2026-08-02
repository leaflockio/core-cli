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

	layout, err := checkConfigLayout(cmd, a)
	if err != nil {
		return nil, err
	}

	if err := loadInvokedConfig(cmd, a, layout); err != nil {
		return nil, err
	}

	plan, cobraCmd, err := f.buildNode(cmd, a, level.LevelRoot)
	if err != nil {
		return nil, err
	}

	registerPersistent(cobraCmd, plan.persistentFlags, a)

	return cobraCmd, nil
}

// buildNode assembles and wires cmd, recursing into children via itself.
func (f *Factory) buildNode(cmd cli.Command, a *app.App, lvl level.Level) (*assembled, *cobra.Command, error) {
	def := cmd.Define(a)

	plan, err := assemble(def, lvl, a)
	if err != nil {
		return nil, nil, err
	}

	cobraCmd, err := f.execute(plan, a)
	if err != nil {
		return nil, nil, err
	}

	return plan, cobraCmd, nil
}

// registerPersistent registers persistentFlags on cmd's PersistentFlags.
func registerPersistent(cmd *cobra.Command, persistentFlags []assembledFlag, a *app.App) {
	for _, fp := range persistentFlags {
		fp.register(cmd.PersistentFlags(), a)
	}
}
