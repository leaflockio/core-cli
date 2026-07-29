// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package level

import "github.com/spf13/cobra"

// Of classifies cmd's own position by walking its Parent chain.
func Of(cmd *cobra.Command) Level {
	if cmd.Parent() == nil {
		return LevelRoot
	}
	if cmd.Parent().Parent() == nil {
		return LevelTop
	}
	return LevelNested
}

// RootOf walks cmd's Parent chain and returns its root ancestor. Returns cmd
// itself if cmd is already the root.
func RootOf(cmd *cobra.Command) *cobra.Command {
	for cmd.Parent() != nil {
		cmd = cmd.Parent()
	}
	return cmd
}

// TopLevelOf walks cmd's Parent chain and returns its top-level ancestor —
// the direct child of the root. Returns cmd itself if cmd is the root or
// already top-level.
func TopLevelOf(cmd *cobra.Command) *cobra.Command {
	for cmd.Parent() != nil && cmd.Parent().Parent() != nil {
		cmd = cmd.Parent()
	}
	return cmd
}
