// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package root implements the top-level command that every other command is
// nested under.
package root

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/config"
)

type command struct {
	children []cli.Command
}

// New returns the root command with children nested under it. Which commands
// are included is decided by the caller — this package stays agnostic to that.
func New(children []cli.Command) cli.Command {
	return &command{children: children}
}

const shortDesc = "Enhance developer experience, right from the command line."

func (c *command) Define(_ *app.App) *cli.Definition {
	return cli.NewDefinition(cli.NewMeta(config.AppName, shortDesc, "")).
		WithChildren(c.children)
}
