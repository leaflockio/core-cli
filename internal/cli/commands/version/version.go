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
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/ui"
	ver "github.com/leaflockio/core-cli/internal/version"
)

// New returns the version command.
func New() cli.Command {
	return &command{}
}

type command struct {
	short bool
}

func (c *command) Define(a *app.App) *cli.Definition {
	return cli.NewDefinition(cli.NewMeta("version", "Print the current version", "")).
		WithFlags([]flags.Flag{
			flags.CommandFlag[*flags.BoolValue]{
				Value: flags.Bool("short", "Print the version number only").WithShorthand("s").WithDest(&c.short),
			},
		}).
		WithHandler(c.run)
}

func (c *command) run(a *app.App, _ []string) error {
	if c.short {
		fmt.Fprintln(a.Printer.Out(), a.Version.Version)
		return nil
	}
	printVersion(a.Version, a.Printer)
	return nil
}

func printVersion(info *ver.Info, printer *ui.Printer) {
	name := printer.Primary(config.AppName)
	version := printer.Secondary("v" + info.Version)

	line1 := lipgloss.JoinHorizontal(lipgloss.Left, name, "  ", version)

	label := printer.Muted
	value := printer.Description
	dot := printer.Muted(" · ")

	line2 := lipgloss.JoinHorizontal(lipgloss.Left,
		label("commit "), value(info.Commit),
		dot,
		label("built "), value(info.Date),
	)

	fmt.Fprintln(printer.Out(), line1)
	fmt.Fprintln(printer.Out(), line2)
}
