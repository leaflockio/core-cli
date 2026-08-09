// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/commands/help"
	"github.com/leaflockio/core-cli/internal/cli/commands/license"
	"github.com/leaflockio/core-cli/internal/cli/commands/version"
)

// commands lists every command wired into the root command tree.
var commands = []cli.Command{
	version.New(),
	license.New(),
	help.New(),
}
