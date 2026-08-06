// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package gitflags

import "github.com/leaflockio/core-cli/internal/cli/flags"

type stagedFlag struct {
	flags.CommandFlag[*flags.BoolValue]
}

func (f stagedFlag) WithDest(dest *bool) flags.CommandFlag[*flags.BoolValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// Staged is the canonical --staged command flag.
var Staged = stagedFlag{
	flags.CommandFlag[*flags.BoolValue]{
		Value: flags.Bool("staged", "Limit to files staged for commit"),
	},
}
