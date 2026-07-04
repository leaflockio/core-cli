// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package factory builds cobra commands from cli.Command definitions.
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

// Build wires cmd into an executable command, including any declared children.
func (f *Factory) Build(cmd cli.Command, a *app.App) (*cobra.Command, error) {
	if cmd == nil {
		return nil, errNilCmd
	}

	def := cmd.Define(a)

	plan, err := assemble(def)
	if err != nil {
		return nil, err
	}

	return f.execute(plan, a)
}
