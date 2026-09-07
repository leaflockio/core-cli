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
	"github.com/leaflockio/core-cli/internal/cli/flags/system/noconfig"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/leaflockio/core-cli/internal/store"
)

// loadConfig is the config-loading entry point for tool.
func loadConfig(cmd cli.Command, a *app.App) error {
	if noConfigRequested(a) {
		return nil
	}

	sources, err := resolveSources(cmd, a)
	if err != nil {
		return err
	}

	if err := checkConflicts(a, sources); err != nil {
		return err
	}

	return loadInvokedConfig(cmd, a, sources)
}

// noConfigRequested reports whether --no-config was passed. Reads
// a.Invocation instead of parsed flags because this runs before cobra
// parses anything (see the noconfig package doc for why).
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

// loadInvokedConfig loads and validates config for the command currently
// being invoked, using sources to locate it. Returns nil when a has no
// Invocation, no child matches the invoked command, or the matched
// command declares no Config.
func loadInvokedConfig(cmd cli.Command, a *app.App, sources *cmdconfig.Sources) error {
	if a.Invocation == nil {
		return nil
	}
	invoked := strings.ToLower(a.Invocation.CommandAt(level.LevelTop))

	def := findChildByName(cmd, a, invoked)
	if def == nil || def.Config == nil {
		return nil
	}

	section, err := resolveSection(a, invoked, sources)
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

// resolveSection reads name's config section: from its own dedicated
// file when sources reports SourceFile, from sources.ManifestCommands
// when sources reports SourceManifest, or nil when sources reports
// SourceNone.
func resolveSection(a *app.App, name string, sources *cmdconfig.Sources) (map[string]any, error) {
	var section map[string]any
	switch sources.Commands[name] {
	case cmdconfig.SourceFile:
		path, err := peekProjectRoot(a, name)
		if err != nil {
			return nil, err
		}
		if _, err := store.Load(path, &section); err != nil {
			return nil, errConfigUnreadable(name, err)
		}
	case cmdconfig.SourceManifest:
		if v, ok := sources.ManifestCommands[name].(map[string]any); ok {
			section = v
		}
	case cmdconfig.SourceNone:
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
