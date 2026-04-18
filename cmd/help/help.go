// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package help

import (
	"fmt"
	"strings"

	"github.com/leaflock/core-cli/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	indentCmdWidth   = 2
	indentGroupWidth = 4
	colSepWidth      = 2
	flagWidth        = 16
)

var (
	indentCmd   = strings.Repeat(" ", indentCmdWidth)
	indentGroup = strings.Repeat(" ", indentGroupWidth)
	colSep      = strings.Repeat(" ", colSepWidth)
)

// Set registers the styled help function on cmd, writing output to printer.Out().
func Set(cmd *cobra.Command, printer *ui.Printer) {
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		fmt.Fprint(printer.Out(), render(c, printer))
	})
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
	b.WriteString("Usage:\n")
	styledPath := styledCommandPath(cmd.CommandPath(), printer)
	useLine := strings.Replace(cmd.UseLine(), cmd.CommandPath(), styledPath, 1)
	b.WriteString(indentCmd + useLine + "\n")
	if cmd.HasAvailableSubCommands() {
		b.WriteString(indentCmd + styledPath + " [command]\n")
	}
	return b.String()
}

// styledCommandPath styles each segment of the command path individually:
// the root (e.g. "leaf") gets Primary, subcommands (e.g. "license") get Secondary.
func styledCommandPath(path string, printer *ui.Printer) string {
	parts := strings.SplitN(path, " ", 2)
	if len(parts) == 1 {
		return printer.Primary(parts[0])
	}
	return printer.Primary(parts[0]) + " " + printer.Secondary(parts[1])
}

func renderCommands(cmd *cobra.Command, printer *ui.Printer) string {
	visible := visibleCommands(cmd)
	if len(visible) == 0 {
		return ""
	}

	width := maxNameWidth(visible)
	var b strings.Builder
	b.WriteString("\nAvailable Commands:\n")

	groups := cmd.Groups()
	if len(groups) == 0 {
		b.WriteString(renderUngrouped(printer, visible, width, indentCmd))
		return b.String()
	}

	b.WriteString(renderGrouped(printer, visible, groups, width))
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
		b.WriteString("\n" + indentCmd + g.Title + "\n")
		b.WriteString(renderUngrouped(printer, cmds, width, indentGroup))
	}

	if len(ungrouped) > 0 {
		b.WriteString("\n" + indentCmd + "Other\n")
		b.WriteString(renderUngrouped(printer, ungrouped, width, indentGroup))
	}

	return b.String()
}

func renderUngrouped(printer *ui.Printer, cmds []*cobra.Command, width int, indent string) string {
	var b strings.Builder
	for _, c := range cmds {
		name := printer.Secondary(fmt.Sprintf("%s%-*s", indent, width, c.Name()))
		b.WriteString(name + colSep + c.Short + "\n")
	}
	return b.String()
}

func renderFlags(cmd *cobra.Command, printer *ui.Printer) string {
	local := cmd.LocalFlags()
	inherited := cmd.InheritedFlags()

	var b strings.Builder
	if local.HasAvailableFlags() {
		b.WriteString("\nFlags:\n")
		local.VisitAll(func(f *pflag.Flag) {
			if !f.Hidden {
				b.WriteString(formatFlag(f, printer))
			}
		})
	}

	if inherited.HasAvailableFlags() {
		b.WriteString("\nGlobal Flags:\n")
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
