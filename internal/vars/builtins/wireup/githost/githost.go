// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package githost defines the GIT_HOST builtin variable.
package githost

import (
	"fmt"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "GIT_HOST"

// GitHost is the hosting provider host parsed from the current
// repository's default remote URL.
var GitHost = vars.RegisterBuiltin(vars.NewWireUp(Name, compute))

func compute(a *app.App) (string, error) {
	if a.Repo.Host == "" {
		if a.Repo.RemoteErr != nil {
			return "", fmt.Errorf("githost: %w", a.Repo.RemoteErr)
		}
		return "", errs.Caller(errs.VAR004,
			"could not determine GIT_HOST: no remote configured",
			nil,
			errs.Context{
				Cause:      "the repository has no remote to read a host from",
				Resolution: "add a remote to the repository, or override GIT_HOST with your own value",
			},
		)
	}
	return a.Repo.Host, nil
}
