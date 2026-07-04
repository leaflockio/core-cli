// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package nocolor provides the canonical --no-color system flag.
package nocolor

import (
	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli/flags"
)

// NoColor is the canonical --no-color system flag.
var NoColor = flags.SystemFlag[*flags.BoolValue]{
	Sub:    flags.SubImplicit,
	Value:  flags.Bool("no-color", "Disable color output"),
	Effect: func(a *app.App) { a.Printer.SetNoColor(true) },
}
