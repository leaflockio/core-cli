// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/nocolor"
	"github.com/leaflockio/core-cli/internal/cli/flags/system/noconfig"
)

// implicitSystemFlags are registered on every command unconditionally.
// Developers must not add these to Definition.Flags.
var implicitSystemFlags = []flags.Flag{
	nocolor.NoColor,
	noconfig.NoConfig,
}
