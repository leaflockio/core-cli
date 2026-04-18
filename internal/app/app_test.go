// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package app

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/leaflock/core-cli/internal/platform"
	"github.com/leaflock/core-cli/internal/terminal"
	"github.com/leaflock/core-cli/internal/ui"
	"github.com/leaflock/core-cli/internal/version"
)

func TestNewBuilder_returnsNonNil(t *testing.T) {
	if NewBuilder() == nil {
		t.Fatal("expected non-nil Builder")
	}
}

func TestBuilder_WithConfig(t *testing.T) {
	a := NewBuilder().WithConfig(nil).Build()
	if a.Config != nil {
		t.Error("expected Config to be nil")
	}
}

func TestBuilder_WithLogger(t *testing.T) {
	log := slog.Default()
	a := NewBuilder().WithLogger(log).Build()

	if a.Log != log {
		t.Error("Log not set correctly")
	}
}

func TestBuilder_WithPrinter(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	a := NewBuilder().WithPrinter(printer).Build()

	if a.Printer != printer {
		t.Error("Printer not set correctly")
	}
}

func TestBuilder_WithPlatform(t *testing.T) {
	plat := platform.Detect()
	a := NewBuilder().WithPlatform(plat).Build()

	if a.Platform != plat {
		t.Error("Platform not set correctly")
	}
}

func TestBuilder_WithVersion(t *testing.T) {
	info := &version.Info{Version: "1.2.3"}
	a := NewBuilder().WithVersion(info).Build()

	if a.Version.Version != "1.2.3" {
		t.Errorf("expected Version %q, got %q", "1.2.3", a.Version.Version)
	}
}

func TestBuilder_Build_fullChain(t *testing.T) {
	printer := ui.NewPrinter(terminal.New(&bytes.Buffer{}, &bytes.Buffer{}, nil))
	plat := platform.Detect()
	log := slog.Default()
	info := &version.Info{Version: "2.0.0", Commit: "abc", Date: "2026-04-18"}

	a := NewBuilder().
		WithLogger(log).
		WithPrinter(printer).
		WithPlatform(plat).
		WithVersion(info).
		Build()

	if a.Log != log {
		t.Error("Log not set in full chain")
	}
	if a.Printer != printer {
		t.Error("Printer not set in full chain")
	}
	if a.Platform != plat {
		t.Error("Platform not set in full chain")
	}
	if a.Version != info {
		t.Error("Version not set in full chain")
	}
}
