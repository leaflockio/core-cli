// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package verbose declares the canonical, explicit-opt-in --verbose system
// flag.
//
// # Explicit, not implicit
//
// Verbose has Sub: [flags.SubExplicit] — it is never applied automatically.
// Only commands that emit progress output opt in.
//
// # Effect always enables verbose, regardless of the parsed value
//
// [flags.SystemFlag.Effect] has no parameter for the flag's parsed value —
// it only fires once the flag has been explicitly set. Verbose's Effect
// therefore always enables verbose output unconditionally; there is no way
// for it to distinguish `--verbose` from `--verbose=false`. Both enable it.
package verbose

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
)

// Verbose is the canonical --verbose system flag.
var Verbose = flags.SystemFlag[*flags.BoolValue]{
	Sub:    flags.SubExplicit,
	Value:  flags.Bool("verbose", "Enable verbose output").WithShorthand("v"),
	Effect: func(a *app.App) { a.Printer.SetVerbose(true) },
}
