// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/leaflockio/core-cli/internal/terminal"
)

// Printer writes styled output to the writers provided by a Terminal. Commands
// receive a Printer rather than writing directly to os.Stdout or os.Stderr,
// which keeps output routable and testable via a bytes.Buffer.
type Printer struct {
	term    *terminal.Terminal
	noColor bool
	verbose bool
}

// NewPrinter constructs a Printer backed by t. Color is automatically disabled
// when the NO_COLOR environment variable is set.
func NewPrinter(t *terminal.Terminal) *Printer {
	return &Printer{
		term:    t,
		noColor: os.Getenv("NO_COLOR") != "",
	}
}

// SetNoColor enables or disables color output. Used by the --no-color flag.
func (p *Printer) SetNoColor(v bool) { p.noColor = v }

// SetVerbose enables or disables verbose output. Used by the --verbose flag.
func (p *Printer) SetVerbose(v bool) { p.verbose = v }

// Verbose prints a progress message to Out when verbose mode is enabled.
// Commands that include flags.Verbose in their Definition.Flags call this
// throughout their handler to surface progress information to the user.
func (p *Printer) Verbose(msg string) {
	if p.verbose {
		fmt.Fprintln(p.term.Out, p.render(&StyleMuted, msg))
	}
}

// Out returns the writer for normal output.
func (p *Printer) Out() io.Writer { return p.term.Out }

// Err returns the writer for error output.
func (p *Printer) Err() io.Writer { return p.term.Err }

// IsTTY reports whether the terminal is interactive.
func (p *Printer) IsTTY() bool { return p.term.IsTTY }

// render applies style to text, returning plain text when color is disabled.
func (p *Printer) render(style *lipgloss.Style, text string) string {
	if p.noColor {
		return text
	}
	return style.Render(text)
}

// Primary returns text styled as a primary chrome element (app name, main command).
func (p *Printer) Primary(text string) string { return p.render(&StylePrimary, text) }

// Secondary returns text styled as a secondary chrome element (subcommand names).
func (p *Printer) Secondary(text string) string { return p.render(&StyleSecondary, text) }

// Description returns text styled as descriptive content (command short descriptions).
func (p *Printer) Description(text string) string { return p.render(&StyleDescription, text) }

// Muted returns text styled as muted/secondary metadata for inline use.
func (p *Printer) Muted(text string) string { return p.render(&StyleMuted, text) }

// Flag returns text styled as a flag name (--flag, -f).
func (p *Printer) Flag(text string) string { return p.render(&StyleFlag, text) }

// Header returns text styled as a bold section header.
func (p *Printer) Header(text string) string { return p.render(&StyleHeader, text) }

// Success prints a success message to Out.
func (p *Printer) Success(msg string) {
	fmt.Fprintln(p.term.Out, p.render(&StyleSuccess, "✓ "+msg))
}

// Info prints an informational message to Out.
func (p *Printer) Info(msg string) {
	fmt.Fprintln(p.term.Out, p.render(&StyleInfo, "→ "+msg))
}

// Warning prints a warning message to Out.
func (p *Printer) Warning(msg string) {
	fmt.Fprintln(p.term.Out, p.render(&StyleWarning, "! "+msg))
}

// Error prints an error message to Err.
func (p *Printer) Error(msg string) {
	fmt.Fprintln(p.term.Err, p.render(&StyleError, "✗ "+msg))
}
