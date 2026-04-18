// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// flagVar is implemented by each typed flag, binding its value onto a pflag.FlagSet.
type flagVar interface {
	bind(flags *pflag.FlagSet, name, usage string)
}

// flagDef describes a single CLI flag: its name, usage text, and value binding.
type flagDef struct {
	name  string
	usage string
	v     flagVar
}

// boolFlag binds a bool flag to a target pointer with a default value.
type boolFlag struct {
	target *bool
	def    bool
}

func (b *boolFlag) bind(flags *pflag.FlagSet, name, usage string) {
	flags.BoolVar(b.target, name, b.def, usage)
}

// rootFlags holds the values bound to root-level CLI flags.
type rootFlags struct {
	noColor bool
}

// defs returns the flag definitions for the root command.
// To add a new root flag, append a flagDef entry here.
func (f *rootFlags) defs() []flagDef {
	return []flagDef{
		{name: "no-color", usage: "Disable color output", v: &boolFlag{target: &f.noColor}},
	}
}

// register binds all root-level flags onto cmd as persistent flags.
func (f *rootFlags) register(cmd *cobra.Command) {
	for _, d := range f.defs() {
		d.v.bind(cmd.PersistentFlags(), d.name, d.usage)
	}
}
