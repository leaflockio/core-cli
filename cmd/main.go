// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"errors"
	"os"

	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/errs"
	"github.com/leaflock/core-cli/internal/logger"
)

// version and env are injected at build time via -ldflags.
//
//	-X main.version=<semver>
//	-X main.env=<dev|prod>
var (
	version string
	env     string
)

func main() {
	if err := run(); err != nil {
		errs.Print(err)
		var e *errs.Error
		if errors.As(err, &e) {
			os.Exit(e.ExitCode)
		}
		os.Exit(errs.ExitInternal)
	}
}

func run() error {
	e, err := config.EnvResolve(env)
	if err != nil {
		return err
	}

	cfg, err := config.Load(e)
	if err != nil {
		return err
	}

	errs.Configure(errs.DisplayConfig{
		ShowCode:       cfg.Errors.ShowCode,
		ShowResolution: cfg.Errors.ShowResolution,
		ShowUnderlying: cfg.Errors.ShowUnderlying,
	})

	log, err := logger.Build(&cfg.Log)
	if err != nil {
		return err
	}

	_ = version
	_ = log

	return nil
}
