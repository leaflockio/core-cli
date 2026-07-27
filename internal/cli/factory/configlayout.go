// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
)

// peekProjectRoot resolves the config directory's own path.
var peekProjectRoot = func(a *app.App) (string, error) {
	return a.Workspace.ForProjectRoot().Peek("")
}

// checkConfigLayout resolves the config directory and validates its layout
// before any command gets built, using the command names declared by cmd's
// top-level children.
func checkConfigLayout(cmd cli.Command, a *app.App) error {
	root := cmd.Define(a)

	allCommands := make([]string, 0, len(root.Children))
	configCommands := make([]string, 0, len(root.Children))
	for _, child := range root.Children {
		def := child.Define(a)
		allCommands = append(allCommands, def.Meta.Use)
		if def.Config != nil {
			configCommands = append(configCommands, def.Meta.Use)
		}
	}

	dir, err := peekProjectRoot(a)
	if err != nil {
		return err
	}

	layout, err := cmdconfig.DetectLayout(dir, allCommands, configCommands)
	if err != nil {
		return err
	}

	return checkExtensionConflict(a, layout)
}

// checkExtensionConflict escalates an entry in layout.ExtensionConflicts to
// an error only when it belongs to the command currently being invoked. A
// conflict on a different command is not escalated. Returns nil when a has
// no Invocation set.
func checkExtensionConflict(a *app.App, layout *cmdconfig.Layout) error {
	if a.Invocation == nil {
		return nil
	}
	invoked := a.Invocation.CommandAt(level.LevelTop)
	exts, conflicted := layout.ExtensionConflicts[invoked]
	if !conflicted {
		return nil
	}
	return errs.Caller(
		errs.CCF004,
		"config file exists under more than one extension",
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("%s has files with extensions: %s", invoked, strings.Join(exts, ", ")),
			Resolution: "keep only one and remove the others",
		},
	)
}
