// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package filemodifiedyear defines the FILE_MODIFIED_YEAR builtin variable.
package filemodifiedyear

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
const Name = "FILE_MODIFIED_YEAR"

// FileModifiedYear is the year of path's most recent commit. A
// file with no commit history yet (new, untracked, or
// staged-but-uncommitted) has no prior modification to report, so this
// falls back to the current year.
var FileModifiedYear = vars.RegisterBuiltin(vars.NewPerCall(Name, `\d{4}(-\d{4})?`, vars.Stable))

func Set(v *vars.Vars, dir, path string) error {
	return v.SetCall(Name, func(*app.App) (string, error) {
		date, err := git.FileModifiedDate(dir, path)
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
