// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"fmt"

	"github.com/leaflockio/core-cli/internal/errs"
)

// registry records every Var passed to RegisterBuiltin, keyed by name.
var registry = map[string]*Var{}

// errRegistration accumulates every error RegisterBuiltin has hit so far.
var errRegistration error

var (
	errNotBuiltin        = errors.New("vars: RegisterBuiltin called with non-builtin variable")
	errAlreadyRegistered = errors.New("vars: builtin variable already registered")
)

// RegisterBuiltin registers v.
func RegisterBuiltin(v *Var, err error) *Var {
	if err != nil {
		errRegistration = errors.Join(errRegistration, err)
		return nil
	}
	if v.origin != OriginBuiltin {
		errRegistration = errors.Join(errRegistration, fmt.Errorf("%w: %s", errNotBuiltin, v.name))
		return nil
	}
	if _, exists := registry[v.name]; exists {
		errRegistration = errors.Join(errRegistration, fmt.Errorf("%w: %s", errAlreadyRegistered, v.name))
		return nil
	}
	registry[v.name] = v
	return v
}

func builtins() (map[string]*Var, error) {
	if errRegistration != nil {
		return nil, errs.Unexpected(errRegistration, errs.Context{
			Cause: "a builtin variable under internal/vars/builtins failed to register — " +
				"an invalid name, a name collision with another builtin, or the wrong " +
				"origin passed to RegisterBuiltin",
			Resolution: "find and fix the offending declaration under internal/vars/builtins.",
		})
	}
	return registry, nil
}
