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

	"github.com/leaflockio/core-cli/internal/comment"
	"github.com/leaflockio/core-cli/internal/invocation"
	"github.com/leaflockio/core-cli/internal/platform"
	"github.com/leaflockio/core-cli/internal/repo"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/leaflockio/core-cli/internal/version"
	"github.com/leaflockio/core-cli/internal/workspace"
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

func TestBuilder_WithRepo(t *testing.T) {
	r := &repo.Info{IsGit: true, RootDir: t.TempDir()}
	a := NewBuilder().WithRepo(r).Build()

	if a.Repo != r {
		t.Error("Repo not set correctly")
	}
}

func TestBuilder_WithInvocation(t *testing.T) {
	inv := &invocation.Invocation{Raw: []string{"run", "--all"}}
	a := NewBuilder().WithInvocation(inv).Build()

	if a.Invocation != inv {
		t.Error("Invocation not set correctly")
	}
}

func TestBuilder_WithWorkspace(t *testing.T) {
	ws, err := workspace.New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	a := NewBuilder().WithWorkspace(ws).Build()

	if a.Workspace != ws {
		t.Error("Workspace not set correctly")
	}
}

func TestApp_ApplyCommentOverrides(t *testing.T) {
	prev, ok := comment.LanguageForExtension("go")
	if !ok {
		t.Fatal("precondition failed: go should be a known extension")
	}
	t.Cleanup(func() {
		comment.ApplyOverrides(map[string]comment.Language{"go": prev}, nil)
	})

	a := NewBuilder().Build()
	a.ApplyCommentOverrides(map[string]comment.Language{"go": comment.CSSStyle}, nil)

	got, ok := comment.LanguageForExtension("go")
	if !ok {
		t.Fatal("LanguageForExtension(go) ok = false, want true")
	}
	if got.Default != comment.CSSStyle.Default {
		t.Errorf("LanguageForExtension(go) = %v, want CSSStyle", got)
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
