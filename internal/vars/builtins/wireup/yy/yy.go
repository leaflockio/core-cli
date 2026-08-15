// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package yy defines the YY builtin variable.
package yy

import (
	"fmt"
	"time"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/vars"
)

// YY is the current year's last two digits.
var YY = vars.RegisterBuiltin(vars.NewWireUp("YY", `\d{2}`, vars.Volatile, compute))

func compute(*app.App) (string, error) {
	return fmt.Sprintf("%02d", time.Now().Year()%100), nil
}
