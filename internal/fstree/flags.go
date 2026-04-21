// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package fstree

import "github.com/spf13/cobra"

// Flags holds the filesystem-scope flag values for commands that support
// the --all file selection mode. AddTo registers the flags onto a command.
type Flags struct {
	All         bool     // --all          : all source files (filesystem walk)
	NoGitignore bool     // --no-gitignore : disable gitignore-aware traversal
	Include     []string // --include      : add include glob pattern (repeatable)
	Exclude     []string // --exclude      : add exclude glob pattern (repeatable)
}

// AddTo registers the filesystem-scope flags onto cmd as local flags.
func (f *Flags) AddTo(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.All, "all", false,
		"all source files on disk")
	cmd.Flags().BoolVar(&f.NoGitignore, "no-gitignore", false,
		"disable gitignore-aware traversal for --all (includes files ignored by git)")
	cmd.Flags().StringArrayVar(&f.Include, "include", nil,
		"include files matching glob pattern (repeatable, additive to config)")
	cmd.Flags().StringArrayVar(&f.Exclude, "exclude", nil,
		"exclude files matching glob pattern (repeatable, additive to config)")
}
