// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package varflag_test

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/cli/flags/command/varflag"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/vars"
)

// resolverOf asserts cf carries a flags.StringSliceResolver and returns it,
// failing the test instead of panicking if it doesn't.
func resolverOf(t *testing.T, cf flags.CommandFlag[*flags.StringSliceValue]) flags.StringSliceResolver {
	t.Helper()
	r, ok := cf.Resolver.(flags.StringSliceResolver)
	if !ok {
		t.Fatalf("Resolver = %T, want a flags.StringSliceResolver", cf.Resolver)
	}
	return r
}

func TestVar_definition(t *testing.T) {
	def := varflag.Var.Definition()
	if def.Meta.Name != "var" {
		t.Errorf("Name = %q, want %q", def.Meta.Name, "var")
	}
}

// TestVar_WithDest_producesUsableVars proves the resolved []*vars.Var
// isn't just parsed text — it's directly usable with Upsert and Resolve,
// the actual consuming API.
func TestVar_WithDest_producesUsableVars(t *testing.T) {
	var dest []*vars.Var
	cf := varflag.Var.WithDest(&dest)

	r := resolverOf(t, cf)
	if err := r.Resolve([]string{"ORG=LeafLock"}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(dest) != 1 {
		t.Fatalf("len(dest) = %d, want 1", len(dest))
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	if err := v.Upsert(dest...); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := v.Resolve("{ORG}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "LeafLock" {
		t.Errorf("Resolve({ORG}) = %q, want %q", got, "LeafLock")
	}
}

// TestVar_WithDest_independentAcrossCalls guards against the same
// regression class DiffFilter's WithDest test covers: two calls on the
// shared Var must not clobber each other's binding.
func TestVar_WithDest_independentAcrossCalls(t *testing.T) {
	var destA, destB []*vars.Var
	a := varflag.Var.WithDest(&destA)
	b := varflag.Var.WithDest(&destB)

	ra := resolverOf(t, a)
	rb := resolverOf(t, b)

	if err := ra.Resolve([]string{"A=1"}); err != nil {
		t.Fatalf("Resolve (a): %v", err)
	}
	if err := rb.Resolve([]string{"B=2"}); err != nil {
		t.Fatalf("Resolve (b): %v", err)
	}

	if len(destA) != 1 || len(destB) != 1 {
		t.Fatalf("destA = %d entries, destB = %d entries, want 1 each", len(destA), len(destB))
	}
}

func TestVar_Resolve_parsesMultipleEntries(t *testing.T) {
	var dest []*vars.Var
	r := resolverOf(t, varflag.Var.WithDest(&dest))

	if err := r.Resolve([]string{"ORG=LeafLock", "YEAR=2026"}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(dest) != 2 {
		t.Errorf("len(dest) = %d, want 2", len(dest))
	}
}

func TestVar_Resolve_valueMayContainEquals(t *testing.T) {
	var dest []*vars.Var
	r := resolverOf(t, varflag.Var.WithDest(&dest))

	if err := r.Resolve([]string{"EXPR=a=b"}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	if err := v.Upsert(dest...); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := v.Resolve("{EXPR}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "a=b" {
		t.Errorf("Resolve({EXPR}) = %q, want %q", got, "a=b")
	}
}

func TestVar_Resolve_missingEqualsIsCallerError(t *testing.T) {
	var dest []*vars.Var
	r := resolverOf(t, varflag.Var.WithDest(&dest))

	err := r.Resolve([]string{"ORG"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR006 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR006)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

// TestVar_Resolve_invalidNamePropagatesFromNewUserStatic proves the flag
// doesn't duplicate vars' own name validation — it surfaces whatever
// vars.NewUserStatic returns (VAR001) as-is.
func TestVar_Resolve_invalidNamePropagatesFromNewUserStatic(t *testing.T) {
	var dest []*vars.Var
	r := resolverOf(t, varflag.Var.WithDest(&dest))

	err := r.Resolve([]string{"bad name=x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR001)
	}
}

func TestVar_Resolve_nilDestIsNoop(t *testing.T) {
	r := resolverOf(t, varflag.Var.WithDest(nil))

	if err := r.Resolve([]string{"ORG=LeafLock"}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
}
