// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package gitowner defines the GIT_OWNER builtin variable.
package gitowner

import (
	"fmt"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "GIT_OWNER"

// GitOwner is the organization or username portion of the current
// repository's default remote URL.
var GitOwner = vars.RegisterBuiltin(vars.NewWireUp(Name, `[A-Za-z0-9][A-Za-z0-9._-]*`, vars.Stable, compute))

func compute(a *app.App) (string, error) {
	if a.Repo.Owner == "" {
		if a.Repo.RemoteErr != nil {
			return "", fmt.Errorf("gitowner: %w", a.Repo.RemoteErr)
		}
		return "", errs.Caller(errs.VAR004,
			"could not determine GIT_OWNER: no remote configured",
			nil,
			errs.Context{
				Cause:      "the repository has no remote to read an owner from",
				Resolution: "add a remote to the repository, or override GIT_OWNER with your own value",
			},
		)
	}
	return a.Repo.Owner, nil
}
