// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd(t *testing.T) *cobra.Command {
	t.Helper()
	return &cobra.Command{Use: "leaf"}
}

func TestRootFlags_noColorRegisteredAsPersistent(t *testing.T) {
	var f rootFlags
	cmd := newTestCmd(t)
	f.register(cmd)

	if cmd.PersistentFlags().Lookup("no-color") == nil {
		t.Error("expected --no-color to be registered as a persistent flag")
	}
}

func TestRootFlags_noLocalVersionFlag(t *testing.T) {
	var f rootFlags
	cmd := newTestCmd(t)
	f.register(cmd)

	if cmd.Flags().Lookup("version") != nil {
		t.Error("expected --version flag to be absent after dropping it")
	}
}
