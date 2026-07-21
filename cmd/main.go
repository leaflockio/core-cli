// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"errors"
	"os"

	"github.com/leaflockio/core-cli/internal/errs"
)

// env is injected at build time via -ldflags:
//
//	-X main.env=<dev|prod>
//
// Version, Commit, and Date are injected into internal/version directly:
//
//	-X github.com/leaflockio/core-cli/internal/version.Version=1.0.0
//	-X github.com/leaflockio/core-cli/internal/version.Commit=abc1234
//	-X github.com/leaflockio/core-cli/internal/version.Date=2026-04-18
var env string

var (
	osExit = os.Exit
	runFn  = run
)

func main() {
	if err := runFn(); err != nil {
		errs.Print(err)
		var e *errs.Error
		if errors.As(err, &e) {
			osExit(e.ExitCode)
			return
		}
		osExit(errs.ExitInternal)
	}
}
