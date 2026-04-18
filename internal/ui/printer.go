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

	"github.com/leaflock/core-cli/internal/terminal"
)

// Printer writes styled output to the writers provided by a Terminal. Commands
// receive a Printer rather than writing directly to os.Stdout or os.Stderr,
// which keeps output routable and testable via a bytes.Buffer.
type Printer struct {
	term *terminal.Terminal
}

// NewPrinter constructs a Printer backed by t.
func NewPrinter(t *terminal.Terminal) *Printer {
	return &Printer{term: t}
}

// Out returns the writer for normal output.
func (p *Printer) Out() io.Writer { return p.term.Out }

// Err returns the writer for error output.
func (p *Printer) Err() io.Writer { return p.term.Err }

// IsTTY reports whether the terminal is interactive.
func (p *Printer) IsTTY() bool { return p.term.IsTTY }

// Success prints a success message to Out.
func (p *Printer) Success(msg string) {
	fmt.Fprintln(p.term.Out, StyleSuccess.Render("✓ "+msg))
}

// Info prints an informational message to Out.
func (p *Printer) Info(msg string) {
	fmt.Fprintln(p.term.Out, StyleInfo.Render("→ "+msg))
}

// Warning prints a warning message to Out.
func (p *Printer) Warning(msg string) {
	fmt.Fprintln(p.term.Out, StyleWarning.Render("! "+msg))
}

// Error prints an error message to Err.
func (p *Printer) Error(msg string) {
	fmt.Fprintln(p.term.Err, StyleError.Render("✗ "+msg))
}

// Muted prints a secondary message to Out.
func (p *Printer) Muted(msg string) {
	fmt.Fprintln(p.term.Out, StyleMuted.Render(msg))
}
