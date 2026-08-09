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

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/hooks/ontreeready"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/spf13/cobra"
)

var _ ontreeready.OnTreeReady = (*command)(nil)

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

// newHelpChild returns the same "help" node the factory would build from
// New()'s Definition, without going through the factory itself.
func newHelpChild() *cobra.Command {
	return &cobra.Command{
		Use:  "help [command]",
		Args: cobra.ArbitraryArgs,
		RunE: cli.ShowHelp,
	}
}

// --- New / Define / run ---

func TestNew_returns_a_command(t *testing.T) {
	if help := New(); help == nil {
		t.Error("New() returned nil")
	}
}

func TestCommand_Define_meta(t *testing.T) {
	c := &command{}
	def := c.Define(nil)
	if got := def.Meta.Use; got != "help" {
		t.Errorf("Use = %q, want %q", got, "help")
	}
	if got := def.Meta.ArgsUsage; got != "[command]" {
		t.Errorf("ArgsUsage = %q, want %q", got, "[command]")
	}
	if def.Handler == nil {
		t.Error("Handler should not be nil")
	}
}

func TestCommand_run_delegatesToShowHelp(t *testing.T) {
	printer, out := newTestPrinter(t)
	root := &cobra.Command{Use: "leaf", Short: "root", Args: cli.NewMeta("leaf", "", "").Args}
	root.SetHelpFunc(func(c *cobra.Command, _ []string) { _, _ = out.WriteString(c.Short) })
	child := newCmd("version", "Print the version")
	root.AddCommand(child)

	c := &command{}
	if err := c.run(nil, root, []string{"version"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	_ = printer
	if out.String() != "Print the version" {
		t.Errorf("expected version's help to render, got %q", out.String())
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
	if !strings.Contains(got, "GENERAL") {
		t.Errorf("expected a GENERAL section header when no Groups are declared, got %q", got)
	}
}

// TestRenderCommands_ungroupedHasBlankLineBeforeSection is a regression
// test: the no-groups-at-all path used to call renderUngrouped directly,
// with no leading blank line or header, so the command list ran straight
// into the USAGE block with no visual separation.
func TestRenderCommands_ungroupedHasBlankLineBeforeSection(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddCommand(newCmd("version", "Print version"))

	got := renderCommands(parent, printer)
	if !strings.HasPrefix(got, "\n") {
		t.Errorf("expected output to start with a blank line before the GENERAL section, got %q", got)
	}
}

func TestRenderCommands_grouped(t *testing.T) {
	printer, _ := newTestPrinter(t)
	parent := newCmd("leaf", "")
	parent.AddGroup(&cobra.Group{ID: "tools", Title: "TOOLS"})
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
	if !strings.Contains(got, "GENERAL") {
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

// --- OnTreeReady ---

func TestCommand_OnTreeReady_usesStashedPrinter(t *testing.T) {
	printer, out := newTestPrinter(t)
	root := newCmd("leaf", "A great tool")
	self := newHelpChild()
	root.AddCommand(self)

	c := &command{}
	c.Define(&app.App{Printer: printer})
	c.OnTreeReady(self, root)
	_ = root.Help()

	if !strings.Contains(out.String(), "A great tool") {
		t.Errorf("expected help output in printer.Out(), got %q", out.String())
	}
}

// --- setHelp ---

func TestSet_writesToPrinterOut(t *testing.T) {
	printer, out := newTestPrinter(t)
	root := newCmd("leaf", "A great tool")
	self := newHelpChild()
	root.AddCommand(self)
	setHelp(self, root, printer)
	_ = root.Help()

	if !strings.Contains(out.String(), "A great tool") {
		t.Errorf("expected help output in printer.Out(), got %q", out.String())
	}
}

func TestSet_helpCommand_showsKnownCommandHelp(t *testing.T) {
	printer, out := newTestPrinter(t)
	root := &cobra.Command{Use: "leaf", Args: cli.NewMeta("leaf", "", "").Args}
	root.AddCommand(newCmd("version", "Print the version"))
	self := newHelpChild()
	root.AddCommand(self)
	setHelp(self, root, printer)

	root.SetArgs([]string{"help", "version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Print the version") {
		t.Errorf("expected version's help in output, got %q", out.String())
	}
}

func TestSet_helpCommand_rejectsUnknownTopic(t *testing.T) {
	printer, _ := newTestPrinter(t)
	root := &cobra.Command{Use: "leaf", Args: cli.NewMeta("leaf", "", "").Args, SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newCmd("version", "Print the version"))
	self := newHelpChild()
	root.AddCommand(self)
	setHelp(self, root, printer)

	root.SetArgs([]string{"help", "banana"})
	if err := root.Execute(); err == nil {
		t.Error("expected an error for an unrecognized help topic, got nil")
	}
}
