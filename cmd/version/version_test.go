// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package version

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leaflock/core-cli/internal/terminal"
	"github.com/leaflock/core-cli/internal/ui"
	ver "github.com/leaflock/core-cli/internal/version"
)

func newTestPrinter(t *testing.T) (*ui.Printer, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	return ui.NewPrinter(terminal.New(out, &bytes.Buffer{}, nil)), out
}

func TestPrint_containsAppName(t *testing.T) {
	printer, out := newTestPrinter(t)
	Print(&ver.Info{Version: "1.0.0", Commit: "abc", Date: "2026-01-01"}, printer)

	if !strings.Contains(out.String(), "leaf") {
		t.Errorf("expected output to contain app name, got %q", out.String())
	}
}

func TestPrint_containsVersion(t *testing.T) {
	printer, out := newTestPrinter(t)
	Print(&ver.Info{Version: "1.2.3", Commit: "abc", Date: "2026-01-01"}, printer)

	if !strings.Contains(out.String(), "v1.2.3") {
		t.Errorf("expected output to contain version, got %q", out.String())
	}
}

func TestPrint_containsCommitAndDate(t *testing.T) {
	printer, out := newTestPrinter(t)
	Print(&ver.Info{Version: "1.0.0", Commit: "deadbeef", Date: "2026-04-18"}, printer)

	output := out.String()
	if !strings.Contains(output, "deadbeef") {
		t.Errorf("expected output to contain commit, got %q", output)
	}
	if !strings.Contains(output, "2026-04-18") {
		t.Errorf("expected output to contain date, got %q", output)
	}
}

func TestPrint_noColor_noANSI(t *testing.T) {
	printer, out := newTestPrinter(t)
	printer.SetNoColor(true)
	Print(&ver.Info{Version: "1.0.0", Commit: "abc", Date: "2026-01-01"}, printer)

	if strings.Contains(out.String(), "\x1b[") {
		t.Errorf("expected no ANSI codes when noColor=true, got %q", out.String())
	}
}

func TestNewCmd_metadata(t *testing.T) {
	printer, _ := newTestPrinter(t)
	cmd := NewCmd(&ver.Info{}, printer)

	if cmd.Use != "version" {
		t.Errorf("expected Use %q, got %q", "version", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected non-empty Short")
	}
	if cmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

func TestNewCmd_run_writesToOut(t *testing.T) {
	printer, out := newTestPrinter(t)
	cmd := NewCmd(&ver.Info{Version: "2.0.0", Commit: "xyz", Date: "2026-04-18"}, printer)
	cmd.Run(cmd, nil)

	if !strings.Contains(out.String(), "v2.0.0") {
		t.Errorf("expected output to contain version after Run, got %q", out.String())
	}
}
