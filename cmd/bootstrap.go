// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"log/slog"
	"os"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli/commands/root"
	"github.com/leaflockio/core-cli/internal/cli/factory"
	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/invocation"
	"github.com/leaflockio/core-cli/internal/logger"
	"github.com/leaflockio/core-cli/internal/platform"
	"github.com/leaflockio/core-cli/internal/repo"
	"github.com/leaflockio/core-cli/internal/terminal"
	"github.com/leaflockio/core-cli/internal/ui"
	"github.com/leaflockio/core-cli/internal/util/concurrent"
	"github.com/leaflockio/core-cli/internal/version"
	"github.com/leaflockio/core-cli/internal/workspace"
	"github.com/spf13/cobra"
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

	cmd, err := buildCommandTree(a)
	if err != nil {
		return err
	}
	return cmd.Execute()
}

// buildCommandTree builds the root command tree and applies cobra-level
// presentation on top.
func buildCommandTree(a *app.App) (*cobra.Command, error) {
	cmd, err := factory.New().Build(root.New(commands), a)
	if err != nil {
		return nil, err
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.CompletionOptions.DisableDefaultCmd = true
	return cmd, nil
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
	term := terminal.New(os.Stdout, os.Stderr, os.Stdin)
	printer := ui.NewPrinter(term)
	inv := invocation.FromArgs()

	var (
		log      *slog.Logger
		plat     *platform.Platform
		repoInfo *repo.Info
	)
	errs := concurrent.Run(
		func() error {
			var err error
			log, err = logger.Build(&cfg.Log)
			return err
		},
		func() error { plat = platform.Detect(); return nil },
		func() error { repoInfo = repo.Detect(); return nil },
	)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	ws, err := workspace.New("", repoInfo.RootDir)
	if err != nil {
		return nil, err
	}

	return app.NewBuilder().
		WithConfig(cfg).
		WithLogger(log).
		WithPrinter(printer).
		WithPlatform(plat).
		WithVersion(version.Current()).
		WithRepo(repoInfo).
		WithInvocation(inv).
		WithWorkspace(ws).
		Build(), nil
}
