// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package help implements the "help" command and the styled rendering used
// for every command's help output.
package help

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	indentCmdWidth = 2
	colSepWidth    = 2
	flagWidth      = 16
)

var (
	indentCmd = strings.Repeat(" ", indentCmdWidth)
	colSep    = strings.Repeat(" ", colSepWidth)
)

// New returns the "help" command.
func New() cli.Command {
	return &command{}
}

type command struct {
	printer *ui.Printer
}

func (c *command) Define(a *app.App) *cli.Definition {
	if a != nil {
		c.printer = a.Printer
	}
	meta := cli.NewMeta("help", "Help about any command", "").
		WithArgsUsage("[command]").
		WithArgs(cobra.ArbitraryArgs)
	return cli.NewDefinition(meta).WithHandler(c.run)
}

func (c *command) run(_ *app.App, cmd *cobra.Command, args []string) error {
	return cli.ShowHelp(cmd, args)
}

func (c *command) OnTreeReady(self, root *cobra.Command) {
	setHelp(self, root, c.printer)
}

// setHelp registers the styled help function on root and designates self
// as root's official help command.
func setHelp(self, root *cobra.Command, printer *ui.Printer) {
	root.SetHelpFunc(func(c *cobra.Command, _ []string) {
		fmt.Fprint(printer.Out(), render(c, printer))
	})
	root.SetHelpCommand(self)
}

func render(cmd *cobra.Command, printer *ui.Printer) string {
	var b strings.Builder
	b.WriteString(renderDescription(cmd, printer))
	b.WriteString(renderUsage(cmd, printer))
	b.WriteString(renderCommands(cmd, printer))
	b.WriteString(renderFlags(cmd, printer))
	b.WriteString(renderFooter(cmd))
	return b.String()
}

func renderDescription(cmd *cobra.Command, printer *ui.Printer) string {
	if cmd.Short == "" {
		return ""
	}
	return printer.Description(cmd.Short) + "\n\n"
}

func renderUsage(cmd *cobra.Command, printer *ui.Printer) string {
	var b strings.Builder
	b.WriteString(printer.Header("USAGE:"))
	b.WriteString("\n")
	styledPath := styledCommandPath(cmd.CommandPath(), printer)
	useLine := strings.Replace(cmd.UseLine(), cmd.CommandPath(), styledPath, 1)
	b.WriteString(indentCmd)
	b.WriteString(useLine)
	b.WriteString("\n")
	if cmd.HasAvailableSubCommands() {
		b.WriteString(indentCmd)
		b.WriteString(styledPath)
		b.WriteString(" [command]\n")
	}
	return b.String()
}

// styledCommandPath styles the command path in two parts. Every segment
// except the last is the invoked command's ancestors, and gets Primary. The
// last segment — the invoked command itself — gets Secondary.
func styledCommandPath(path string, printer *ui.Printer) string {
	i := strings.LastIndex(path, " ")
	if i == -1 {
		return printer.Primary(path)
	}
	return printer.Primary(path[:i]) + " " + printer.Secondary(path[i+1:])
}

func renderCommands(cmd *cobra.Command, printer *ui.Printer) string {
	visible := visibleCommands(cmd)
	if len(visible) == 0 {
		return ""
	}

	width := maxNameWidth(visible)
	groups := cmd.Groups()
	if len(groups) == 0 {
		return renderUngroupedSection(printer, visible, width)
	}

	return renderGrouped(printer, visible, groups, width)
}

func renderUngroupedSection(printer *ui.Printer, cmds []*cobra.Command, width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(printer.Header("GENERAL"))
	b.WriteString("\n")
	b.WriteString(renderUngrouped(printer, cmds, width, indentCmd))
	return b.String()
}

func renderGrouped(printer *ui.Printer, visible []*cobra.Command, groups []*cobra.Group, width int) string {
	grouped := make(map[string][]*cobra.Command, len(groups))
	var ungrouped []*cobra.Command
	for _, c := range visible {
		if c.GroupID != "" {
			grouped[c.GroupID] = append(grouped[c.GroupID], c)
		} else {
			ungrouped = append(ungrouped, c)
		}
	}

	var b strings.Builder
	for _, g := range groups {
		cmds := grouped[g.ID]
		if len(cmds) == 0 {
			continue
		}
		b.WriteString("\n")
		b.WriteString(printer.Header(g.Title))
		b.WriteString("\n")
		b.WriteString(renderUngrouped(printer, cmds, width, indentCmd))
	}

	if len(ungrouped) > 0 {
		b.WriteString(renderUngroupedSection(printer, ungrouped, width))
	}

	return b.String()
}

func renderUngrouped(printer *ui.Printer, cmds []*cobra.Command, width int, indent string) string {
	var b strings.Builder
	for _, c := range cmds {
		name := printer.Secondary(fmt.Sprintf("%s%-*s", indent, width, c.Name()))
		b.WriteString(name)
		b.WriteString(colSep)
		b.WriteString(c.Short)
		b.WriteByte('\n')
	}
	return b.String()
}

func renderFlags(cmd *cobra.Command, printer *ui.Printer) string {
	local := cmd.LocalFlags()
	inherited := cmd.InheritedFlags()

	var b strings.Builder
	if local.HasAvailableFlags() {
		b.WriteString("\n")
		b.WriteString(printer.Header("FLAGS:"))
		b.WriteString("\n")
		local.VisitAll(func(f *pflag.Flag) {
			if !f.Hidden {
				b.WriteString(formatFlag(f, printer))
			}
		})
	}

	if inherited.HasAvailableFlags() {
		b.WriteString("\n")
		b.WriteString(printer.Header("GLOBAL FLAGS:"))
		b.WriteString("\n")
		inherited.VisitAll(func(f *pflag.Flag) {
			if !f.Hidden {
				b.WriteString(formatFlag(f, printer))
			}
		})
	}

	return b.String()
}

func renderFooter(cmd *cobra.Command) string {
	if !cmd.HasAvailableSubCommands() {
		return ""
	}
	return fmt.Sprintf("\nUse \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
}

func visibleCommands(cmd *cobra.Command) []*cobra.Command {
	all := cmd.Commands()
	visible := make([]*cobra.Command, 0, len(all))
	for _, c := range all {
		if !c.Hidden {
			visible = append(visible, c)
		}
	}
	return visible
}

func maxNameWidth(cmds []*cobra.Command) int {
	width := 0
	for _, c := range cmds {
		if n := len(c.Name()); n > width {
			width = n
		}
	}
	return width
}

func formatFlag(f *pflag.Flag, printer *ui.Printer) string {
	var name string
	if f.Shorthand != "" {
		name = printer.Flag(fmt.Sprintf("  -%s, --%-*s", f.Shorthand, flagWidth, f.Name))
	} else {
		name = printer.Flag(fmt.Sprintf("      --%-*s", flagWidth, f.Name))
	}
	return name + f.Usage + "\n"
}
