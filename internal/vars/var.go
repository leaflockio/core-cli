// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
)

var nameRe = regexp.MustCompile(`^[A-Z0-9_]+$`)

var errInvalidName = errors.New("vars: invalid variable name")

// normalizeName upper-cases name and rejects any character outside
// [A-Z0-9_].
func normalizeName(name string) (string, error) {
	upper := strings.ToUpper(name)
	if !nameRe.MatchString(upper) {
		return "", fmt.Errorf("%w %q: only letters, digits, and underscore are allowed", errInvalidName, name)
	}
	return upper, nil
}

// Kind distinguishes when a Var's value is computed.
type Kind int

const (
	KindStatic Kind = iota
	KindWireUp
	KindPerCall
)

// Origin distinguishes who defined a Var.
type Origin int

const (
	OriginBuiltin Origin = iota
	OriginUserDefined
)

// Var is a variable resolved once.
type Var struct {
	name     string
	origin   Origin
	kind     Kind
	compute  func(*app.App) (string, error)
	resolved bool
	value    string
}

// errNoCompute is the static base error wrapped when resolve finds no
// compute function set — for WireUp this means a builtin was constructed
// wrong; for PerCall it means whatever's calling Resolve forgot to call
// SetCall for this name first.
var errNoCompute = errors.New("vars: no compute function set")

// resolve returns v's value. Static and WireUp variables compute once and
// cache the result for the lifetime of this Var. PerCall variables always
// recompute, SetCall attaches a fresh compute right before each Resolve
// pass and clearCalls wipes it after, so there's nothing worth caching
// across calls.
func (v *Var) resolve(a *app.App) (string, error) {
	if v.kind != KindPerCall && v.resolved {
		return v.value, nil
	}
	if v.compute == nil {
		return "", errs.Unexpected(fmt.Errorf("%w: %s", errNoCompute, v.name))
	}

	value, err := v.compute(a)
	if err != nil {
		return "", fmt.Errorf("vars: resolving %s: %w", v.name, err)
	}
	if v.kind != KindPerCall {
		v.value = value
		v.resolved = true
	}
	return value, nil
}

// NewBuiltin constructs a builtin Static variable, a fixed value.
func NewBuiltin(name, value string) (*Var, error) {
	n, err := normalizeName(name)
	if err != nil {
		return nil, err
	}
	return &Var{name: n, origin: OriginBuiltin, kind: KindStatic, resolved: true, value: value}, nil
}

// NewWireUp constructs a builtin WireUp variable: computed once, eagerly,
// the first time it's resolved.
func NewWireUp(name string, fn func(*app.App) (string, error)) (*Var, error) {
	n, err := normalizeName(name)
	if err != nil {
		return nil, err
	}
	return &Var{name: n, origin: OriginBuiltin, kind: KindWireUp, compute: fn}, nil
}

// NewUserStatic constructs a user-defined Static variable, a fixed value.
func NewUserStatic(name, value string) (*Var, error) {
	n, err := normalizeName(name)
	if err != nil {
		return nil, errs.Caller(errs.VAR001,
			fmt.Sprintf("invalid variable name %q", name),
			err,
			errs.Context{
				Cause:      "variable names may only contain letters, digits, and underscore",
				Resolution: fmt.Sprintf("rename %q so it only uses letters, digits, and underscore", name),
			},
		)
	}
	return &Var{name: n, origin: OriginUserDefined, kind: KindStatic, resolved: true, value: value}, nil
}

// NewPerCall declares a builtin PerCall variable, known to every Vars that
// includes it, but not resolvable until a consumer supplies its context —
// via a builtin-specific setter.
func NewPerCall(name string) (*Var, error) {
	n, err := normalizeName(name)
	if err != nil {
		return nil, err
	}
	return &Var{name: n, origin: OriginBuiltin, kind: KindPerCall}, nil
}
