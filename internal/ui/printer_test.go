// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/terminal"
)

func newTestPrinter(t *testing.T) (*Printer, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	return NewPrinter(terminal.New(out, errBuf, nil)), out, errBuf
}

func TestNewPrinter(t *testing.T) {
	printer, _, _ := newTestPrinter(t)
	if printer == nil {
		t.Fatal("expected non-nil Printer")
	}
}

func TestPrinter_Out(t *testing.T) {
	out := &bytes.Buffer{}
	printer := NewPrinter(terminal.New(out, &bytes.Buffer{}, nil))

	if printer.Out() != out {
		t.Error("Out() did not return the expected writer")
	}
}

func TestPrinter_Err(t *testing.T) {
	errBuf := &bytes.Buffer{}
	printer := NewPrinter(terminal.New(&bytes.Buffer{}, errBuf, nil))

	if printer.Err() != errBuf {
		t.Error("Err() did not return the expected writer")
	}
}

func TestPrinter_IsTTY(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	if printer.IsTTY() {
		t.Error("expected IsTTY=false for a buffer-backed terminal")
	}
}

func TestNewPrinter_noColorFromEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	printer, _, _ := newTestPrinter(t)

	if !printer.noColor {
		t.Error("expected noColor=true when NO_COLOR env var is set")
	}
}

func TestPrinter_SetNoColor(t *testing.T) {
	printer, _, _ := newTestPrinter(t)
	printer.SetNoColor(true)

	if !printer.noColor {
		t.Error("expected noColor=true after SetNoColor(true)")
	}
}

func TestPrinter_render_plainWhenNoColor(t *testing.T) {
	printer, _, _ := newTestPrinter(t)
	printer.SetNoColor(true)

	result := printer.render(&StyleSuccess, "hello")
	if result != "hello" {
		t.Errorf("expected plain %q when noColor=true, got %q", "hello", result)
	}
}

func TestPrinter_render_styledWhenColor(t *testing.T) {
	printer, _, _ := newTestPrinter(t)
	printer.SetNoColor(false)

	result := printer.render(&StyleSuccess, "hello")
	if !strings.Contains(result, "hello") {
		t.Errorf("expected result to contain %q, got %q", "hello", result)
	}
}

func TestPrinter_Primary(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	if !strings.Contains(printer.Primary("leaf"), "leaf") {
		t.Error("Primary() did not contain the input text")
	}
}

func TestPrinter_Secondary(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	if !strings.Contains(printer.Secondary("version"), "version") {
		t.Error("Secondary() did not contain the input text")
	}
}

func TestPrinter_Description(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	if !strings.Contains(printer.Description("Developer tooling"), "Developer tooling") {
		t.Error("Description() did not contain the input text")
	}
}

func TestPrinter_Success(t *testing.T) {
	printer, out, errBuf := newTestPrinter(t)
	printer.Success("installed")

	if !strings.Contains(out.String(), "installed") {
		t.Errorf("expected Out to contain %q, got %q", "installed", out.String())
	}
	if errBuf.Len() != 0 {
		t.Error("expected Err to be empty for Success")
	}
}

func TestPrinter_Info(t *testing.T) {
	printer, out, errBuf := newTestPrinter(t)
	printer.Info("running")

	if !strings.Contains(out.String(), "running") {
		t.Errorf("expected Out to contain %q, got %q", "running", out.String())
	}
	if errBuf.Len() != 0 {
		t.Error("expected Err to be empty for Info")
	}
}

func TestPrinter_Warning(t *testing.T) {
	printer, out, errBuf := newTestPrinter(t)
	printer.Warning("deprecated")

	if !strings.Contains(out.String(), "deprecated") {
		t.Errorf("expected Out to contain %q, got %q", "deprecated", out.String())
	}
	if errBuf.Len() != 0 {
		t.Error("expected Err to be empty for Warning")
	}
}

func TestPrinter_Error(t *testing.T) {
	printer, out, errBuf := newTestPrinter(t)
	printer.Error("something failed")

	if !strings.Contains(errBuf.String(), "something failed") {
		t.Errorf("expected Err to contain %q, got %q", "something failed", errBuf.String())
	}
	if out.Len() != 0 {
		t.Error("expected Out to be empty for Error")
	}
}

func TestPrinter_Muted(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	result := printer.Muted("secondary info")
	if !strings.Contains(result, "secondary info") {
		t.Errorf("expected Muted to contain %q, got %q", "secondary info", result)
	}
}

func TestPrinter_Flag(t *testing.T) {
	printer, _, _ := newTestPrinter(t)

	result := printer.Flag("--verbose")
	if !strings.Contains(result, "--verbose") {
		t.Errorf("expected Flag to contain %q, got %q", "--verbose", result)
	}
}
