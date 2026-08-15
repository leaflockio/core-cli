// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package year defines the YEAR builtin variable.
package year

import (
	"strconv"
	"time"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Year is the current year.
var Year = vars.RegisterBuiltin(vars.NewWireUp("YEAR", `\d{4}(-\d{4})?`, vars.Volatile, compute))

func compute(*app.App) (string, error) {
	return strconv.Itoa(time.Now().Year()), nil
}
