// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package app

import (
	"log/slog"

	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/platform"
	"github.com/leaflock/core-cli/internal/ui"
	"github.com/leaflock/core-cli/internal/version"
	"github.com/leaflock/core-cli/internal/workspace"
)

// App is the central DI container. It is constructed once in main and passed
// as a single unit to every command. Adding a new infrastructure dependency
// only requires a new field, a new With* method, and updating main — no other
// wiring changes needed.
type App struct {
	Config    *config.Config
	Log       *slog.Logger
	Printer   *ui.Printer
	Platform  *platform.Platform
	Version   *version.Info
	Workspace *workspace.Workspace // filesystem path manager, never nil
}

// Builder constructs an App using a fluent chain of With* calls.
type Builder struct {
	cfg       *config.Config
	log       *slog.Logger
	printer   *ui.Printer
	plat      *platform.Platform
	version   *version.Info
	workspace *workspace.Workspace
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithConfig sets the application configuration.
func (b *Builder) WithConfig(cfg *config.Config) *Builder {
	b.cfg = cfg
	return b
}

// WithLogger sets the structured logger.
func (b *Builder) WithLogger(log *slog.Logger) *Builder {
	b.log = log
	return b
}

// WithPrinter sets the terminal printer.
func (b *Builder) WithPrinter(printer *ui.Printer) *Builder {
	b.printer = printer
	return b
}

// WithPlatform sets the detected host platform.
func (b *Builder) WithPlatform(plat *platform.Platform) *Builder {
	b.plat = plat
	return b
}

// WithVersion sets the binary version info.
func (b *Builder) WithVersion(v *version.Info) *Builder {
	b.version = v
	return b
}

// WithWorkspace sets the filesystem path manager.
func (b *Builder) WithWorkspace(ws *workspace.Workspace) *Builder {
	b.workspace = ws
	return b
}

// Build assembles and returns the App.
func (b *Builder) Build() *App {
	return &App{
		Config:    b.cfg,
		Log:       b.log,
		Printer:   b.printer,
		Platform:  b.plat,
		Version:   b.version,
		Workspace: b.workspace,
	}
}
