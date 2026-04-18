// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"os"

	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/errs"
	"github.com/leaflock/core-cli/internal/logger"
	"github.com/leaflock/core-cli/internal/platform"
	"github.com/leaflock/core-cli/internal/terminal"
	"github.com/leaflock/core-cli/internal/ui"
	"github.com/leaflock/core-cli/internal/version"
)

// run wires the build-time env into the startup sequence.
func run() error {
	return runWith(env, initConfig, buildApp)
}

// runWith resolves config, builds the app, and executes the root command.
func runWith(
	envStr string,
	cfgFn func(string) (*config.Config, error),
	appFn func(*config.Config) (*app.App, error),
) error {
	cfg, err := cfgFn(envStr)
	if err != nil {
		return err
	}
	a, err := appFn(cfg)
	if err != nil {
		return err
	}
	return (&root{app: a}).cmd().Execute()
}

// initConfig resolves the active environment, loads configuration, and
// applies the error display settings from that config.
func initConfig(envStr string) (*config.Config, error) {
	e, err := config.EnvResolve(envStr)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(e)
	if err != nil {
		return nil, err
	}
	errs.Configure(errs.DisplayConfig{
		ShowCode:       cfg.Errors.ShowCode,
		ShowResolution: cfg.Errors.ShowResolution,
		ShowUnderlying: cfg.Errors.ShowUnderlying,
	})
	return cfg, nil
}

// buildApp constructs all infrastructure from cfg and returns the assembled App.
func buildApp(cfg *config.Config) (*app.App, error) {
	log, err := logger.Build(&cfg.Log)
	if err != nil {
		return nil, err
	}
	term := terminal.New(os.Stdout, os.Stderr, os.Stdin)
	printer := ui.NewPrinter(term)
	plat := platform.Detect()
	return app.NewBuilder().
		WithConfig(cfg).
		WithLogger(log).
		WithPrinter(printer).
		WithPlatform(plat).
		WithVersion(version.Current()).
		Build(), nil
}
