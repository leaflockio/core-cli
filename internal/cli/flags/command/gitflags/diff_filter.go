// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package gitflags

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
)

type diffFilterFlag struct {
	flags.CommandFlag[*flags.StringValue]
}

func (f diffFilterFlag) WithDest(dest *string) flags.CommandFlag[*flags.StringValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// diffFilterChars are the git diff-filter status letters this flag accepts.
// Uppercase includes files with that status, lowercase excludes them.
const diffFilterChars = "ACDMRTUXB"

// Validate reports whether value is a well-formed git diff-filter string.
func (f diffFilterFlag) Validate(value string) error {
	letters := strings.TrimSuffix(value, "*")
	for _, c := range letters {
		if !strings.ContainsRune(diffFilterChars, unicode.ToUpper(c)) {
			return errs.Caller(errs.GIT004,
				fmt.Sprintf("invalid --diff-filter character %q", string(c)),
				nil,
				errs.Context{
					Cause: fmt.Sprintf("%q is not a recognized diff-filter character", string(c)),
					Resolution: fmt.Sprintf(
						"use only characters from %s (case-insensitive), optionally followed by a single trailing *",
						diffFilterChars,
					),
				},
			)
		}
	}
	return nil
}

// DiffFilter is the canonical --diff-filter command flag.
var DiffFilter = diffFilterFlag{
	flags.CommandFlag[*flags.StringValue]{
		Value: flags.String(
			"diff-filter",
			"Which kinds of file changes to include",
		).WithDefault("ACM"),
	},
}
