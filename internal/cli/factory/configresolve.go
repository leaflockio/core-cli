// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"slices"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/level"
)

// peekProjectRoot resolves name's path within the config directory. Pass ""
// for the config directory's own path.
var peekProjectRoot = func(a *app.App, name string) (string, error) {
	return a.Workspace.ForProjectRoot().Peek(name)
}

// resolveSources resolves the config directory and every command's
// config source, using the command names declared by cmd's top-level
// children.
func resolveSources(cmd cli.Command, a *app.App) (*cmdconfig.Sources, error) {
	dir, err := peekProjectRoot(a, "")
	if err != nil {
		return nil, err
	}

	root := cmd.Define(a)
	allCommands := make([]string, 0, len(root.Children))
	configCommands := make([]string, 0, len(root.Children))
	for _, child := range root.Children {
		def := child.Define(a)
		name := strings.ToLower(def.Meta.Use)
		allCommands = append(allCommands, name)
		if def.Config != nil {
			configCommands = append(configCommands, name)
		}
	}

	return cmdconfig.ResolveConfig(dir, allCommands, configCommands)
}

// checkConflicts escalates an extension or manifest conflict into an
// error only when it belongs to the command currently being invoked. A
// conflict on a different command is not escalated. Returns nil when a
// has no Invocation set.
func checkConflicts(a *app.App, sources *cmdconfig.Sources) error {
	if a.Invocation == nil {
		return nil
	}
	invoked := strings.ToLower(a.Invocation.CommandAt(level.LevelTop))

	if exts, conflicted := sources.ExtensionConflicts[invoked]; conflicted {
		return cmdconfig.ErrExtensionConflict(invoked, exts)
	}
	if slices.Contains(sources.ManifestConflicts, invoked) {
		return cmdconfig.ErrManifestConflict(invoked)
	}
	return nil
}
