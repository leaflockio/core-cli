// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"errors"

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

var errNilCmd = errors.New("factory[build]: cmd must not be nil")

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
		return nil, errNilCmd
	}

	plan, cobraCmd, err := f.buildNode(cmd, a, true)
	if err != nil {
		return nil, err
	}

	registerPersistent(cobraCmd, plan.persistentFlags, a)

	return cobraCmd, nil
}

// buildNode assembles and wires cmd, recursing into children via itself.
func (f *Factory) buildNode(cmd cli.Command, a *app.App, isAppRoot bool) (*assembled, *cobra.Command, error) {
	def := cmd.Define(a)

	plan, err := assemble(def, isAppRoot)
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
