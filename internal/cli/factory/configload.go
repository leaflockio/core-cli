// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/cmdconfig"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/leaflockio/core-cli/internal/store"
)

// loadInvokedConfig loads and validates config for the command currently
// being invoked, using layout to locate it. Returns nil when a has no
// Invocation, no child matches the invoked command, or the matched command
// declares no Config.
func loadInvokedConfig(cmd cli.Command, a *app.App, layout *cmdconfig.Layout) error {
	if a.Invocation == nil {
		return nil
	}
	invoked := strings.ToLower(a.Invocation.CommandAt(level.LevelTop))

	def := findChildByName(cmd, a, invoked)
	if def == nil || def.Config == nil {
		return nil
	}

	section, err := resolveSection(a, invoked, layout.Mode)
	if err != nil {
		return err
	}
	if section == nil {
		return nil
	}

	if err := def.Config.Load(section); err != nil {
		return err
	}
	return def.Config.Validate()
}

// findChildByName returns cmd's direct child whose lowercased Meta.Use
// matches name, or nil if none does.
func findChildByName(cmd cli.Command, a *app.App, name string) *cli.Definition {
	for _, child := range cmd.Define(a).Children {
		def := child.Define(a)
		if strings.EqualFold(def.Meta.Use, name) {
			return def
		}
	}
	return nil
}

// resolveSection reads name's config section according to mode. Returns a
// nil map when no file backs the section.
func resolveSection(a *app.App, name string, mode cmdconfig.LayoutMode) (map[string]any, error) {
	var section map[string]any

	switch mode {
	case cmdconfig.LayoutModular:
		path, err := peekProjectRoot(a, name)
		if err != nil {
			return nil, err
		}
		if _, err := store.Load(path, &section); err != nil {
			return nil, errConfigUnreadable(name, err)
		}

	case cmdconfig.LayoutFlat:
		var manifest map[string]any
		if _, err := store.Load(manifestPath(a), &manifest); err != nil {
			return nil, errConfigUnreadable(name, err)
		}
		if s, ok := manifest[name].(map[string]any); ok {
			section = s
		}

	case cmdconfig.LayoutNone:
		// section stays nil — no config directory at all.
	}

	return section, nil
}

// errConfigUnreadable reports that name's config file exists but could not
// be read or decoded.
func errConfigUnreadable(name string, err error) error {
	return errs.Caller(
		errs.CCF005,
		"config file could not be read",
		err,
		errs.Context{
			Cause:      "the config file for \"" + name + "\" exists but could not be read or decoded",
			Resolution: "check the file's syntax and permissions",
		},
	)
}
