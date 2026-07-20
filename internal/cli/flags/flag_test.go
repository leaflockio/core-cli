// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import "github.com/leaflockio/core-cli/internal/cli/flags"

// Compile-time checks that both flag kinds satisfy the Flag interface.
var (
	_ flags.Flag = flags.CommandFlag[*flags.BoolValue]{}
	_ flags.Flag = flags.SystemFlag[*flags.BoolValue]{}
)
