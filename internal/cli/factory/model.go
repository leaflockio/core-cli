// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import (
	"github.com/leaflock/core-cli/internal/app"
	"github.com/leaflock/core-cli/internal/cli"
	"github.com/leaflock/core-cli/internal/cli/flags"
	"github.com/spf13/pflag"
)

// assembled is the validated plan for a command, ready for the executor.
type assembled struct {
	def cli.Definition

	// hasFlags is true when the command declared flags beyond the implicit set.
	hasFlags bool

	// hasChildren is true when the command declared subcommands.
	hasChildren bool

	flags []assembledFlag
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
	register func(*pflag.FlagSet)

	// effect fires the flag's side-effect when the flag was explicitly set.
	// Only set for system kind flags.
	effect func(*pflag.FlagSet, *app.App)

	// resolve reads the raw parsed value from pflag and passes it to the
	// flag's typed Resolve() method. Only set when hasResolver is true.
	resolve func(*pflag.FlagSet) error
}
