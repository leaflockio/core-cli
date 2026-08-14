// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package date defines the DATE builtin variable.
package date

import (
	"time"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Date is the current date, in ISO 8601 (YYYY-MM-DD) form.
var Date = vars.RegisterBuiltin(vars.NewWireUp("DATE", compute))

func compute(*app.App) (string, error) {
	return time.Now().Format(time.DateOnly), nil
}
