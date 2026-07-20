// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/logger"
)

var (
	errTestConfig = errors.New("config error")
	errTestApp    = errors.New("app error")
)

// brokenCmd declares neither a Handler nor Children, tripping the factory's
// assemble guard when wired in as a non-root command.
type brokenCmd struct{}

func (brokenCmd) Define(*app.App) *cli.Definition {
	return &cli.Definition{Meta: cli.Meta{Use: "broken"}}
}

// withBrokenCommands temporarily swaps commands for a slice that fails to
// assemble, restoring the original on cleanup.
func withBrokenCommands(t *testing.T) {
	t.Helper()
	original := commands
	commands = []cli.Command{brokenCmd{}}
	t.Cleanup(func() { commands = original })
}

// minimalCfg returns a Config with the minimum valid logger settings.
func minimalCfg() *config.Config {
	return &config.Config{
		Log: logger.Config{
			Level:   logger.LevelInfo,
			Console: logger.ConsoleConfig{Enabled: true, Format: logger.FormatText},
		},
	}
}

func TestRunWith_configError(t *testing.T) {
	err := runWith("",
		func(string) (*config.Config, error) { return nil, errTestConfig },
		nil,
	)
	if !errors.Is(err, errTestConfig) {
		t.Errorf("expected config error, got %v", err)
	}
}

func TestRunWith_appError(t *testing.T) {
	err := runWith("",
		func(string) (*config.Config, error) { return &config.Config{}, nil },
		func(*config.Config) (*app.App, error) { return nil, errTestApp },
	)
	if !errors.Is(err, errTestApp) {
		t.Errorf("expected app error, got %v", err)
	}
}

func TestRunWith_buildCommandTreeError(t *testing.T) {
	withBrokenCommands(t)

	err := runWith("",
		func(string) (*config.Config, error) { return &config.Config{}, nil },
		func(*config.Config) (*app.App, error) { return &app.App{}, nil },
	)
	if err == nil {
		t.Error("expected error when buildCommandTree fails")
	}
}

func TestBuildCommandTree_propagatesFactoryError(t *testing.T) {
	withBrokenCommands(t)

	if _, err := buildCommandTree(&app.App{}); err == nil {
		t.Error("expected error when factory.Build fails")
	}
}

// TestRunWith_executesCommand verifies that the root command executes successfully.
func TestRunWith_executesCommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"leaf"}

	err := runWith("",
		func(string) (*config.Config, error) { return minimalCfg(), nil },
		buildApp,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRun_devDefaults verifies run() succeeds end-to-end using dev config defaults.
func TestRun_devDefaults(t *testing.T) {
	t.Setenv("LEAF_CONFIG_DIR", t.TempDir())

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"leaf"}

	if err := run(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInitConfig_invalidEnv verifies that an unknown env string returns an error.
func TestInitConfig_invalidEnv(t *testing.T) {
	if _, err := initConfig("invalid"); err == nil {
		t.Error("expected error for invalid env")
	}
}

// TestInitConfig_invalidConfigFile verifies that a malformed config file returns an error.
func TestInitConfig_invalidConfigFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(": invalid: yaml: :"), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	t.Setenv("LEAF_CONFIG_DIR", dir)

	if _, err := initConfig(""); err == nil {
		t.Error("expected error for invalid config file")
	}
}

// TestInitConfig_validDefaults verifies that an empty config dir loads successfully using defaults.
func TestInitConfig_validDefaults(t *testing.T) {
	t.Setenv("LEAF_CONFIG_DIR", t.TempDir())

	cfg, err := initConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

// TestBuildApp_setsAllFields verifies that all App fields are populated from a valid config.
func TestBuildApp_setsAllFields(t *testing.T) {
	a, err := buildApp(minimalCfg())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Config == nil {
		t.Error("expected Config to be set")
	}
	if a.Log == nil {
		t.Error("expected Log to be set")
	}
	if a.Printer == nil {
		t.Error("expected Printer to be set")
	}
	if a.Platform == nil {
		t.Error("expected Platform to be set")
	}
	if a.Version == nil {
		t.Error("expected Version to be set")
	}
	if a.Invocation == nil {
		t.Error("expected Invocation to be set")
	}
	if a.Repo == nil {
		t.Error("expected Repo to be set")
	}
	if a.Workspace == nil {
		t.Error("expected Workspace to be set")
	}
}

// TestBuildApp_propagatesLoggerError verifies that a logger build failure is propagated.
func TestBuildApp_propagatesLoggerError(t *testing.T) {
	cfg := &config.Config{
		Log: logger.Config{
			Level: logger.LevelInfo,
			// Neither console nor file enabled — logger.Build returns ErrNoOutputs.
		},
	}
	if _, err := buildApp(cfg); !errors.Is(err, logger.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got %v", err)
	}
}

func TestBuildApp_propagatesWorkspaceError(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := buildApp(minimalCfg()); err == nil {
		t.Error("expected error when workspace.New fails")
	}
}
