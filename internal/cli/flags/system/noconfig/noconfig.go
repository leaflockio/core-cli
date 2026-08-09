// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package noconfig declares the canonical, engine-owned --no-config system
// flag.
//
// # A project config problem must never block a command from running
//
// A malformed config file, or two conflicting files for the same command,
// otherwise stops every command from running at all — even one unrelated
// to whatever's wrong with the config. A user in the middle of diagnosing
// that problem still needs to run leaf commands, without first deleting or
// fixing the file. --no-config is that escape hatch: it skips project
// config-file layout validation and loading entirely for the invocation,
// leaving every command's config-backed fields at their defaults.
//
// # A dedicated package for a single implicit flag
//
// NoConfig has Sub: [flags.SubImplicit] — it's applied to every command
// automatically, and no command ever imports it directly; only the engine's
// list of implicit flags does. Giving it its own package keeps that
// exclusivity unambiguous at the one place it's referenced.
//
// # Effect is a no-op — the real behavior can't live there
//
// factory.Build's config-layout validation and loading run before cobra
// parses any flags, so a SystemFlag.Effect (which only fires during flag
// parsing) is already too late to gate them. The actual skip is a raw
// argv check against the invocation, done directly in factory.Build,
// referencing NoConfig's own Definition().Meta.Name rather than a second
// hardcoded flag-name string. Effect still has to be a real function —
// [flags.SystemFlag]'s registration calls it unconditionally — so it's a
// deliberate no-op, not an oversight.
package noconfig

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
)

// NoConfig is the canonical --no-config system flag.
var NoConfig = flags.SystemFlag[*flags.BoolValue]{
	Sub:    flags.SubImplicit,
	Value:  flags.Bool("no-config", "Skip project config files"),
	Effect: func(*app.App) {},
}
