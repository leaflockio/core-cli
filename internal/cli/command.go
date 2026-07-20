// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli

import "github.com/leaflockio/core-cli/internal/app"

// Command is implemented by every command package.
// Define returns the complete declaration of the command — its identity, group,
// flags, handler, and any nested subcommands.
type Command interface {
	Define(a *app.App) *Definition
}
