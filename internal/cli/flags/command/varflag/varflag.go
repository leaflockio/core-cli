// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package varflag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

var errVarInvalidFormat = errors.New("expected KEY=VALUE")

type varResolver struct {
	dest *[]*vars.Var
}

func (r *varResolver) IsResolver() {}

// Resolve parses each KEY=VALUE entry from raw into a user-defined Var.
func (r *varResolver) Resolve(raw []string) error {
	if r.dest == nil {
		return nil
	}
	result := make([]*vars.Var, 0, len(raw))
	for _, entry := range raw {
		k, v, ok := strings.Cut(entry, "=")
		if !ok {
			return errs.Caller(errs.VAR006,
				fmt.Sprintf("invalid --var value %q", entry),
				fmt.Errorf("%w: %s", errVarInvalidFormat, entry),
				errs.Context{
					Cause:      fmt.Sprintf("%q has no \"=\" separating a key from a value", entry),
					Resolution: "use the form --var KEY=VALUE",
				},
			)
		}
		nv, err := vars.NewUserStatic(k, v)
		if err != nil {
			return err
		}
		result = append(result, nv)
	}
	*r.dest = result
	return nil
}

type varFlag struct {
	flags.CommandFlag[*flags.StringSliceValue]
}

// WithDest returns a CommandFlag bound to dest.
func (f varFlag) WithDest(dest *[]*vars.Var) flags.CommandFlag[*flags.StringSliceValue] {
	cf := f.CommandFlag
	cf.Resolver = &varResolver{dest: dest}
	return cf
}

// Var is the canonical --var command flag for repeatable KEY=VALUE pairs.
var Var = varFlag{
	CommandFlag: flags.CommandFlag[*flags.StringSliceValue]{
		Value: flags.StringSlice("var", "Set a variable as KEY=VALUE"),
	},
}
