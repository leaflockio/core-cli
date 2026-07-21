// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package nocolor declares the canonical, engine-owned --no-color system
// flag.
//
// # A dedicated package for a single implicit flag
//
// NoColor has Sub: [flags.SubImplicit] — it's applied to every command
// automatically, and no command ever imports it directly; only the engine's
// list of implicit flags does. Giving it its own package keeps that
// exclusivity unambiguous at the one place it's referenced.
//
// # Effect always disables color, regardless of the parsed value
//
// [flags.SystemFlag.Effect] has no parameter for the flag's parsed value —
// it only fires once the flag has been explicitly set. NoColor's Effect
// therefore always disables color unconditionally; there is no way for it
// to distinguish `--no-color` from `--no-color=false`. Both disable color.
package nocolor

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
)

// NoColor is the canonical --no-color system flag.
var NoColor = flags.SystemFlag[*flags.BoolValue]{
	Sub:    flags.SubImplicit,
	Value:  flags.Bool("no-color", "Disable color output"),
	Effect: func(a *app.App) { a.Printer.SetNoColor(true) },
}
