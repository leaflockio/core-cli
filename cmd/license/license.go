// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package license

import (
	"github.com/leaflock/core-cli/internal/app"
	"github.com/spf13/cobra"
)

// New returns the "leaf license" subcommand.
func New(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "license",
		Short: "Manage license headers in source files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}
