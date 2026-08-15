// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package spdxid defines the SPDX_ID builtin variable.
package spdxid

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "SPDX_ID"

// SPDXID is the SPDX identifier classified from the repository's detected
// LICENSE file.
var SPDXID = vars.RegisterBuiltin(vars.NewWireUp(Name, `[A-Za-z0-9.+-]+`, vars.Stable, compute))

func compute(a *app.App) (string, error) {
	if a.Repo.License.SPDXID == "" {
		return "", errs.Caller(errs.VAR005,
			"could not determine SPDX_ID: no classifiable license file",
			nil,
			errs.Context{
				Cause: "no LICENSE file at the repository root could be classified to a known SPDX identifier",
				Resolution: "add a recognizable LICENSE file to the repository, " +
					"or override SPDX_ID with your own value",
			},
		)
	}
	return a.Repo.License.SPDXID, nil
}
