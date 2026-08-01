// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package license

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

// New returns the license command.
func New() cli.Command {
	return &command{}
}

type command struct{}

func (c *command) Define(_ *app.App) *cli.Definition {
	return cli.NewDefinition(cli.NewMeta("license", "Manage license headers in source files", "")).
		WithHandler(c.run)
}

func (c *command) run(a *app.App, _ *cobra.Command, _ []string) error {
	a.Printer.Info("license: not yet implemented")
	return nil
}
