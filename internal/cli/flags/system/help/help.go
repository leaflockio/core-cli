// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package help declares the canonical, engine-owned --help system flag.
//
// Effect is left empty on purpose — cobra's own execute() already reads
// this flag and shows help, so there's nothing left to hook. Registering
// it here, as an implicit system flag on the root's PersistentFlags,
// makes cobra treat it as inherited instead of re-adding its own local
// copy per command.
package help

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/flags"
)

// Help is the canonical --help system flag.
var Help = flags.SystemFlag[*flags.BoolValue]{
	Sub:    flags.SubImplicit,
	Value:  flags.Bool("help", "help for this command").WithShorthand("h"),
	Effect: func(*app.App) {},
}
