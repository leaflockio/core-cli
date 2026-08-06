// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package gitflags

import "github.com/leaflockio/core-cli/internal/cli/flags"

type prFlag struct {
	flags.CommandFlag[*flags.BoolValue]
}

func (f prFlag) WithDest(dest *bool) flags.CommandFlag[*flags.BoolValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// PR is the canonical --pr command flag.
var PR = prFlag{
	flags.CommandFlag[*flags.BoolValue]{
		Value: flags.Bool("pr", "Limit to files changed in the pull request"),
	},
}
