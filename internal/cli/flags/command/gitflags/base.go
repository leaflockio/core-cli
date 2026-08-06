// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package gitflags

import "github.com/leaflockio/core-cli/internal/cli/flags"

type baseFlag struct {
	flags.CommandFlag[*flags.StringValue]
}

func (f baseFlag) WithDest(dest *string) flags.CommandFlag[*flags.StringValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// Base is the canonical --base command flag.
var Base = baseFlag{
	flags.CommandFlag[*flags.StringValue]{
		Value: flags.String("base", "Branch or commit to compare against"),
	},
}
