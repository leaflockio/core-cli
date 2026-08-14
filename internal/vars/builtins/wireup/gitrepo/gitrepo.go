// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package gitrepo defines the GIT_REPO builtin variable.
package gitrepo

import (
	"fmt"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "GIT_REPO"

// GitRepo is the repository name portion of the current repository's
// default remote URL.
var GitRepo = vars.RegisterBuiltin(vars.NewWireUp(Name, compute))

func compute(a *app.App) (string, error) {
	if a.Repo.RepoName == "" {
		if a.Repo.RemoteErr != nil {
			return "", fmt.Errorf("gitrepo: %w", a.Repo.RemoteErr)
		}
		return "", errs.Caller(errs.VAR004,
			"could not determine GIT_REPO: no remote configured",
			nil,
			errs.Context{
				Cause:      "the repository has no remote to read a repo name from",
				Resolution: "add a remote to the repository, or override GIT_REPO with your own value",
			},
		)
	}
	return a.Repo.RepoName, nil
}
