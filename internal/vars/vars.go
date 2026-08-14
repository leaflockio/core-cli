// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package vars resolves named template variables to their string values.
package vars

import (
	"errors"
	"fmt"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
)

var (
	errSetCallUnknown = errors.New("vars: not a declared PerCall variable on this Vars")
	errUpsertOrigin   = errors.New("vars: not a user-defined variable")
)

// Vars holds a resolved set of variables for the current invocation.
type Vars struct {
	app  *app.App
	base map[string]*Var
}

// New returns a Vars seeded from every registered builtin. WireUp
// variables are resolved once, immediately, against a. PerCall variables
// are only declared — they stay unresolved until a consumer supplies
// their context. Use [Vars.Upsert] afterward to layer in user-defined
// variables.
func New(a *app.App) (*Vars, error) {
	reg, err := builtins()
	if err != nil {
		return nil, err
	}
	base := make(map[string]*Var, len(reg))
	for name, v := range reg {
		if v.kind == KindWireUp {
			if _, err := v.resolve(a); err != nil {
				return nil, err
			}
		}
		base[name] = v
	}
	return &Vars{app: a, base: base}, nil
}

// SetCall attaches this call's compute function to an already-declared
// PerCall variable — it never creates one.
func (v *Vars) SetCall(name string, compute func(*app.App) (string, error)) error {
	name, err := normalizeName(name)
	if err != nil {
		return errs.Unexpected(err)
	}
	target, ok := v.base[name]
	if !ok || target.kind != KindPerCall {
		return errs.Unexpected(fmt.Errorf("%w: %s", errSetCallUnknown, name))
	}
	target.compute = compute
	return nil
}

// Upsert adds or replaces user-defined variables on this Vars: a name
// already present in base is replaced, any other name is added. It
// validates every given Var before changing anything, so a rejected call
// leaves Vars unmodified.
func (v *Vars) Upsert(vars ...*Var) error {
	for _, nv := range vars {
		if nv.origin != OriginUserDefined {
			return errs.Unexpected(fmt.Errorf("%w: %s", errUpsertOrigin, nv.name))
		}
	}
	for _, nv := range vars {
		v.base[nv.name] = nv
	}
	return nil
}
