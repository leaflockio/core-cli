// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package help

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newTestPrinter(t *testing.T) (*ui.Printer, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	return ui.NewPrinter(terminal.New(out, &bytes.Buffer{}, nil)), out
}

func newCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run:   func(_ *cobra.Command, _ []string) {},
	}
}

// --- renderDescription ---

func TestRenderDescription_withShort(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "A great tool")

	got := renderDescription(cmd, printer)
	if !strings.Contains(got, "A great tool") {
		t.Errorf("expected description, got %q", got)
	}
	if !strings.HasSuffix(got, "\n\n") {
		t.Errorf("expected trailing newlines, got %q", got)
	}
}

func TestRenderDescription_empty(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")

	if got := renderDescription(cmd, printer); got != "" {
		t.Errorf("expected empty string for no Short, got %q", got)
	}
}

// --- renderUsage ---

func TestRenderUsage_noSubcommands(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")

	got := renderUsage(cmd, printer)
	if !strings.Contains(got, "USAGE:") {
		t.Errorf("expected Usage header, got %q", got)
	}
	if strings.Contains(got, "[command]") {
		t.Errorf("expected no [command] line for leaf with no subcommands, got %q", got)
	}
}

func TestRenderUsage_withSubcommands(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddCommand(newCmd("version", "Print version"))

	got := renderUsage(parent, printer)
	if !strings.Contains(got, "[command]") {
		t.Errorf("expected [command] line for command with subcommands, got %q", got)
	}
}

func TestRenderUsage_containsBothNames(t *testing.T) {
	printer, _ := newTestPrinter(t)
	printer.SetNoColor(true)
	parent := newCmd("leaf", "")
	child := newCmd("license", "")
	parent.AddCommand(child)

	got := renderUsage(child, printer)
	if !strings.Contains(got, "leaf") {
		t.Errorf("expected parent name in usage, got %q", got)
	}
	if !strings.Contains(got, "license") {
		t.Errorf("expected command name in usage, got %q", got)
	}
}

// --- styledCommandPath ---

func TestStyledCommandPath_rootOnly(t *testing.T) {
	printer, _ := newTestPrinter(t)
	printer.SetNoColor(true)

	got := styledCommandPath("leaf", printer)
	if got != "leaf" {
		t.Errorf("got %q, want %q", got, "leaf")
	}
}

func TestStyledCommandPath_withSubcommand(t *testing.T) {
	printer, _ := newTestPrinter(t)
	printer.SetNoColor(true)

	got := styledCommandPath("leaf license", printer)
	if got != "leaf license" {
		t.Errorf("got %q, want %q", got, "leaf license")
	}
}

// TestStyledCommandPath_groupsAncestorsTogether guards against splitting on
// the first space instead of the last: "leaf license" (the ancestors) must
// be styled as one unit, distinct from "add" (the invoked command itself).
func TestStyledCommandPath_groupsAncestorsTogether(t *testing.T) {
	printer, _ := newTestPrinter(t)

	got := styledCommandPath("leaf license add", printer)
	want := printer.Primary("leaf license") + " " + printer.Secondary("add")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- visibleCommands / renderCommands ---

func TestVisibleCommands_filtersHidden(t *testing.T) {
	parent := newCmd("leaf", "")
	visible := newCmd("version", "Print version")
	hidden := newCmd("internal", "")
	hidden.Hidden = true
	parent.AddCommand(visible, hidden)

	got := visibleCommands(parent)
	if len(got) != 1 || got[0].Use != "version" {
		t.Errorf("expected only visible commands, got %v", got)
	}
}

func TestRenderCommands_noSubcommands(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")

	if got := renderCommands(cmd, printer); got != "" {
		t.Errorf("expected empty string for no subcommands, got %q", got)
	}
}

func TestRenderCommands_ungrouped(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddCommand(newCmd("version", "Print version"))

	got := renderCommands(parent, printer)
	if !strings.Contains(got, "version") {
		t.Errorf("expected version in output, got %q", got)
	}
}

func TestRenderCommands_grouped(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddGroup(&cobra.Group{ID: "tools", Title: "Tools"})
	sub := newCmd("license", "Manage licenses")
	sub.GroupID = "tools"
	parent.AddCommand(sub)

	got := renderCommands(parent, printer)
	if !strings.Contains(got, "TOOLS") {
		t.Errorf("expected group title in output, got %q", got)
	}
	if !strings.Contains(got, "license") {
		t.Errorf("expected license in output, got %q", got)
	}
}

func TestRenderCommands_ungroupedFallsToGeneral(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddGroup(&cobra.Group{ID: "tools", Title: "Tools"})
	parent.AddCommand(newCmd("orphan", "No group"))

	got := renderCommands(parent, printer)
	if !strings.Contains(got, "w") {
		t.Errorf("expected General section, got %q", got)
	}
}

// --- maxNameWidth ---

func TestMaxNameWidth(t *testing.T) {
	cmds := []*cobra.Command{
		newCmd("hi", ""),
		newCmd("version", ""),
		newCmd("go", ""),
	}
	if got := maxNameWidth(cmds); got != len("version") {
		t.Errorf("expected %d, got %d", len("version"), got)
	}
}

// --- renderFlags ---

func TestRenderFlags_noFlags(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")
	if got := renderFlags(cmd, printer); got != "" {
		t.Errorf("expected empty output for no flags, got %q", got)
	}
}

func TestRenderFlags_localFlag(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	got := renderFlags(cmd, printer)
	if !strings.Contains(got, "FLAGS:") {
		t.Errorf("expected Flags section, got %q", got)
	}
	if !strings.Contains(got, "verbose") {
		t.Errorf("expected verbose flag, got %q", got)
	}
}

func TestRenderFlags_inheritedFlag(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.PersistentFlags().Bool("no-color", false, "Disable color")
	child := newCmd("version", "")
	parent.AddCommand(child)
	// Cobra populates inherited flags after command tree is built.
	_ = parent.Execute()

	got := renderFlags(child, printer)
	if !strings.Contains(got, "GLOBAL FLAGS:") {
		t.Errorf("expected Global Flags section, got %q", got)
	}
	if !strings.Contains(got, "no-color") {
		t.Errorf("expected no-color in global flags, got %q", got)
	}
}

func TestRenderFlags_hiddenFlagExcluded(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")
	cmd.Flags().Bool("secret", false, "Hidden flag")
	if err := cmd.Flags().MarkHidden("secret"); err != nil {
		t.Fatalf("MarkHidden: %v", err)
	}

	got := renderFlags(cmd, printer)
	if strings.Contains(got, "secret") {
		t.Errorf("expected hidden flag to be excluded, got %q", got)
	}
}

// --- formatFlag ---

func TestFormatFlag_withShorthand(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose")
	f := cmd.Flags().Lookup("verbose")

	got := formatFlag(f, printer)
	if !strings.Contains(got, "-v,") {
		t.Errorf("expected shorthand -v, got %q", got)
	}
	if !strings.Contains(got, "--verbose") {
		t.Errorf("expected --verbose, got %q", got)
	}
}

func TestFormatFlag_withoutShorthand(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := newCmd("leaf", "")
	cmd.Flags().Bool("no-color", false, "Disable color")
	f := cmd.Flags().Lookup("no-color")

	got := formatFlag(f, printer)
	if strings.Contains(got, "-,") {
		t.Errorf("unexpected shorthand in output, got %q", got)
	}
	if !strings.Contains(got, "--no-color") {
		t.Errorf("expected --no-color, got %q", got)
	}
}

// --- renderFooter ---

func TestRenderFooter_withSubcommands(t *testing.T) {
	parent := newCmd("leaf", "")
	parent.AddCommand(newCmd("version", ""))

	got := renderFooter(parent)
	if !strings.Contains(got, "--help") {
		t.Errorf("expected --help hint, got %q", got)
	}
}

func TestRenderFooter_noSubcommands(t *testing.T) {
	cmd := newCmd("leaf", "")
	if got := renderFooter(cmd); got != "" {
		t.Errorf("expected empty footer for no subcommands, got %q", got)
	}
}

// --- Set ---

func TestSet_writesToPrinterOut(t *testing.T) {
	printer, out := newTestPrinter(t)
	cmd := newCmd("leaf", "A great tool")
	Set(cmd, printer)
	_ = cmd.Help()

	if !strings.Contains(out.String(), "A great tool") {
		t.Errorf("expected help output in printer.Out(), got %q", out.String())
	}
}
