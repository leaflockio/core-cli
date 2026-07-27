// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package level

import (
	"testing"

	"github.com/spf13/cobra"
)

// buildTree returns root -> top -> nested -> deeper, all wired via AddCommand.
func buildTree() (*cobra.Command, *cobra.Command, *cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "tool"}
	top := &cobra.Command{Use: "group"}
	nested := &cobra.Command{Use: "action"}
	deeper := &cobra.Command{Use: "sub"}
	root.AddCommand(top)
	top.AddCommand(nested)
	nested.AddCommand(deeper)
	return root, top, nested, deeper
}

func TestOf(t *testing.T) {
	root, top, nested, deeper := buildTree()

	tests := []struct {
		name string
		cmd  *cobra.Command
		want Level
	}{
		{"root", root, LevelRoot},
		{"top", top, LevelTop},
		{"nested", nested, LevelNested},
		{"deeper is still nested", deeper, LevelNested},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Of(tt.cmd); got != tt.want {
				t.Errorf("Of(%s) = %v, want %v", tt.cmd.Use, got, tt.want)
			}
		})
	}
}

func TestRootOf(t *testing.T) {
	root, top, nested, deeper := buildTree()

	for _, cmd := range []*cobra.Command{root, top, nested, deeper} {
		if got := RootOf(cmd); got != root {
			t.Errorf("RootOf(%s) = %s, want %s", cmd.Use, got.Use, root.Use)
		}
	}
}

func TestTopLevelOf(t *testing.T) {
	root, top, nested, deeper := buildTree()

	tests := []struct {
		name string
		cmd  *cobra.Command
		want *cobra.Command
	}{
		{"root returns itself", root, root},
		{"top returns itself", top, top},
		{"nested returns top", nested, top},
		{"deeper returns top", deeper, top},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TopLevelOf(tt.cmd); got != tt.want {
				t.Errorf("TopLevelOf(%s) = %s, want %s", tt.cmd.Use, got.Use, tt.want.Use)
			}
		})
	}
}
