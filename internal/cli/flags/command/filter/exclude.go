// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package filter

import "github.com/leaflockio/core-cli/internal/cli/flags"

type excludeFlag struct {
	flags.CommandFlag[*flags.StringSliceValue]
}

func (f excludeFlag) WithDest(dest *[]string) flags.CommandFlag[*flags.StringSliceValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// Exclude is the canonical --exclude command flag.
var Exclude = excludeFlag{
	flags.CommandFlag[*flags.StringSliceValue]{
		Value: flags.StringSlice("exclude", "Glob pattern of files to exclude"),
	},
}
