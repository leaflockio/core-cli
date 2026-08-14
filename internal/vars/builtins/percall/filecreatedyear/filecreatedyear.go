// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package filecreatedyear defines the FILE_CREATED_YEAR builtin variable.
package filecreatedyear

import (
	"errors"
	"strconv"
	"time"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/git"
	"github.com/leaflockio/core-cli/internal/vars"
)

// Name is the variable's registered name.
const Name = "FILE_CREATED_YEAR"

// FileCreatedYear is the year of path's earliest commit. A file
// with no commit history yet (new, untracked, or staged-but-uncommitted)
// has no earlier year to report, so this falls back to the current year.
var FileCreatedYear = vars.RegisterBuiltin(vars.NewPerCall(Name))

func Set(v *vars.Vars, dir, path string) error {
	return v.SetCall(Name, func(*app.App) (string, error) {
		date, err := git.FileCreatedDate(dir, path)
		if err != nil {
			var e *errs.Error
			if errors.As(err, &e) && e.Code == errs.GIT006 {
				return strconv.Itoa(time.Now().Year()), nil
			}
			return "", err
		}
		return strconv.Itoa(date.Year()), nil
	})
}
