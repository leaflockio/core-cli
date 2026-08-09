// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package ontreeready declares the OnTreeReady hook and Notify, the walk
// that fires it.
package ontreeready

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/spf13/cobra"
)

// OnTreeReady is implemented by a command that needs one-time access to the
// fully assembled root *cobra.Command, once Factory.Build has finished
// wiring every node. Definition only ever describes a command's own node —
// concerns that must apply uniformly across the whole tree can't be
// expressed there. The OnTreeReady method runs once per implementing
// command, after Build's recursive wiring completes, regardless of where
// in the tree that command sits.
type OnTreeReady interface {
	cli.Command
	OnTreeReady(root *cobra.Command)
}

// Notify walks cmd's declared tree and calls OnTreeReady on every command
// that implements it, passing the fully-built root.
func Notify(cmd cli.Command, a *app.App, root *cobra.Command) {
	def := cmd.Define(a)
	if tr, ok := cmd.(OnTreeReady); ok {
		tr.OnTreeReady(root)
	}
	for _, child := range def.Children {
		Notify(child, a, root)
	}
}
