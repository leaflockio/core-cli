// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package filter

import "github.com/leaflockio/core-cli/internal/cli/flags"

type includeFlag struct {
	flags.CommandFlag[*flags.StringSliceValue]
}

func (f includeFlag) WithDest(dest *[]string) flags.CommandFlag[*flags.StringSliceValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// Include is the canonical --include command flag.
var Include = includeFlag{
	flags.CommandFlag[*flags.StringSliceValue]{
		Value: flags.StringSlice("include", "Glob pattern of files to include"),
	},
}
