// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	licensecmd "github.com/leaflock/core-cli/cmd/license"
	versioncmd "github.com/leaflock/core-cli/cmd/version"
	"github.com/leaflock/core-cli/internal/app"
	"github.com/spf13/cobra"
)

// registry maps each subcommand factory to its display group.
// Add new commands here — root.go picks them up automatically.
var registry = []struct {
	groupID string
	factory func(*app.App) *cobra.Command
}{
	{groupGeneral, versioncmd.New},
	{groupTools, licensecmd.New},
}
