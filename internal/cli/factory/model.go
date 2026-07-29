// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/cli"
	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/level"
	"github.com/spf13/pflag"
)

// assembled is the validated plan for a command, ready for the executor.
type assembled struct {
	def cli.Definition

	// level is this command's position in the tree.
	level level.Level

	// hasFlags is true when the command declared flags in its Definition.
	hasFlags bool

	// hasChildren is true when the command declared subcommands.
	hasChildren bool

	// flags holds the plans for this command's own Definition.Flags. The
	// executor registers these on the built command's local Flags(), so they
	// apply only to this command, not its children.
	flags []assembledFlag

	// persistentFlags holds the plans for implicitSystemFlags. Only
	// Factory.Build registers these, on the root's PersistentFlags.
	persistentFlags []assembledFlag
}

// assembledFlag is the execution record for a single flag.
type assembledFlag struct {
	name string
	kind flags.FlagKind
	sub  flags.FlagSubcategory

	// hasShorthand is true when the flag declared a single-character alias.
	hasShorthand bool

	// hasResolver is true when the flag opted into custom value writing via a Resolver.
	// Only applicable to command kind flags.
	hasResolver bool

	// register writes the flag into a pflag.FlagSet at command build time.
	register func(*pflag.FlagSet, *app.App)

	// resolve reads the raw parsed value from pflag and passes it to the
	// flag's typed Resolve() method. Only set when hasResolver is true.
	resolve func(*pflag.FlagSet) error
}
