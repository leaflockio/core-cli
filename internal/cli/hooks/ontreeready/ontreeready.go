// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package ontreeready declares the OnTreeReady hook.
package ontreeready

import (
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

// OnTreeReady is implemented by a command that needs one-time access to its
// own fully-built *cobra.Command, plus the tree's root, once Factory.Build
// has finished wiring every node. Definition only ever describes a
// command's own node, so concerns that must apply uniformly across the
// whole tree can't be expressed there — self is exactly the node the
// factory built for this command, and root is the whole tree.
type OnTreeReady interface {
	cli.Command
	OnTreeReady(self, root *cobra.Command)
}
