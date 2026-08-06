// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package dryrun

import "github.com/leaflockio/core-cli/internal/cli/flags"

type dryRunFlag struct {
	flags.CommandFlag[*flags.BoolValue]
}

func (f dryRunFlag) WithDest(dest *bool) flags.CommandFlag[*flags.BoolValue] {
	f.Value = f.Value.WithDest(dest)
	return f.CommandFlag
}

// DryRun is the canonical --dry-run command flag.
var DryRun = dryRunFlag{
	flags.CommandFlag[*flags.BoolValue]{
		Value: flags.Bool("dry-run", "Report what would be written without writing it"),
	},
}
