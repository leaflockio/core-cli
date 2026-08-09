// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/hooks/ontreeready"
	"github.com/spf13/cobra"
)

// hookRecord is one command's own built node, paired with whichever hooks
// it implements. A nil field means the command doesn't implement that one.
type hookRecord struct {
	self *cobra.Command

	treeReady ontreeready.OnTreeReady
}

// discoverHooks checks cmd against every hook this package knows about.
func discoverHooks(cmd cli.Command, self *cobra.Command, hooks *[]hookRecord) {
	discoverTreeReady(cmd, self, hooks)
}

// discoverTreeReady appends a hookRecord for cmd to hooks when cmd
// implements ontreeready.OnTreeReady.
func discoverTreeReady(cmd cli.Command, self *cobra.Command, hooks *[]hookRecord) {
	if tr, ok := cmd.(ontreeready.OnTreeReady); ok {
		*hooks = append(*hooks, hookRecord{self: self, treeReady: tr})
	}
}

// fireTreeReady calls OnTreeReady on every hookRecord that has one.
func fireTreeReady(hooks []hookRecord, root *cobra.Command) {
	for _, h := range hooks {
		if h.treeReady != nil {
			h.treeReady.OnTreeReady(h.self, root)
		}
	}
}
