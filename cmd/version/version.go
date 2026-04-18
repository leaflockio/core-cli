// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package version

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/ui"
	ver "github.com/leaflock/core-cli/internal/version"
	"github.com/spf13/cobra"
)

// Print writes the styled version output to printer.Out().
func Print(info *ver.Info, printer *ui.Printer) {
	name := printer.Primary(config.AppName)
	version := printer.Muted("v" + info.Version)
	commit := printer.Muted(info.Commit)
	date := printer.Muted(info.Date)

	fmt.Fprintln(printer.Out(), lipgloss.JoinHorizontal(lipgloss.Left, name, "  ", version, "  ", commit, "  ", date))
}

// New returns the "leaf version" subcommand.
func New(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			Print(a.Version, a.Printer)
		},
	}
}
