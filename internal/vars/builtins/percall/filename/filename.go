// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package filename defines the FILE_NAME builtin variable.
package filename

import (
	"path/filepath"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "FILE_NAME"

// FileName is path's base filename, per call.
var FileName = vars.RegisterBuiltin(vars.NewPerCall(Name, `[A-Za-z0-9_.-]+`, vars.Stable))

func Set(v *vars.Vars, path string) error {
	return v.SetCall(Name, func(*app.App) (string, error) {
		return filepath.Base(path), nil
	})
}
