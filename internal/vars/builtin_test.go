// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
)

// --- RegisterBuiltin ---

func TestRegisterBuiltin_addsToRegistry(t *testing.T) {
	withRegistry(t, map[string]*Var{}, nil)

	v, err := NewWireUp("YEAR", func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}

	if got := RegisterBuiltin(v, nil); got != v {
		t.Errorf("RegisterBuiltin = %v, want %v", got, v)
	}
	if registry["YEAR"] != v {
		t.Errorf("registry[YEAR] = %v, want %v", registry["YEAR"], v)
	}
	if errRegistration != nil {
		t.Errorf("errRegistration = %v, want nil", errRegistration)
	}
}

func TestRegisterBuiltin_propagatesConstructorError(t *testing.T) {
	withRegistry(t, map[string]*Var{}, nil)

	got := RegisterBuiltin(NewWireUp("bad name", func(*app.App) (string, error) { return "", nil }))
	if got != nil {
		t.Errorf("RegisterBuiltin = %v, want nil", got)
	}
	if !errors.Is(errRegistration, errInvalidName) {
		t.Errorf("errRegistration = %v, want wrapping errInvalidName", errRegistration)
	}
}

func TestRegisterBuiltin_rejectsNonBuiltinOrigin(t *testing.T) {
	withRegistry(t, map[string]*Var{}, nil)

	v, err := NewUserStatic("FOO", "bar")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}

	if got := RegisterBuiltin(v, nil); got != nil {
		t.Errorf("RegisterBuiltin = %v, want nil", got)
	}
	if !errors.Is(errRegistration, errNotBuiltin) {
		t.Errorf("errRegistration = %v, want wrapping errNotBuiltin", errRegistration)
	}
	if _, exists := registry["FOO"]; exists {
		t.Error("registry[FOO] exists, want a rejected registration to not be stored")
	}
}

func TestRegisterBuiltin_rejectsDuplicateName(t *testing.T) {
	withRegistry(t, map[string]*Var{}, nil)

	first, err := NewWireUp("YEAR", func(*app.App) (string, error) { return "first", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	if got := RegisterBuiltin(first, nil); got != first {
		t.Fatalf("first RegisterBuiltin = %v, want %v", got, first)
	}

	second, err := NewWireUp("YEAR", func(*app.App) (string, error) { return "second", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	if got := RegisterBuiltin(second, nil); got != nil {
		t.Errorf("second RegisterBuiltin = %v, want nil", got)
	}
	if !errors.Is(errRegistration, errAlreadyRegistered) {
		t.Errorf("errRegistration = %v, want wrapping errAlreadyRegistered", errRegistration)
	}
	if registry["YEAR"] != first {
		t.Error("registry[YEAR] changed, want the first registration to remain")
	}
}

func TestRegisterBuiltin_accumulatesErrorsViaJoin(t *testing.T) {
	withRegistry(t, map[string]*Var{}, nil)

	userVar, err := NewUserStatic("FOO", "bar")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	RegisterBuiltin(userVar, nil)
	RegisterBuiltin(NewWireUp("bad name", func(*app.App) (string, error) { return "", nil }))

	if !errors.Is(errRegistration, errNotBuiltin) {
		t.Errorf("errRegistration = %v, want wrapping errNotBuiltin", errRegistration)
	}
	if !errors.Is(errRegistration, errInvalidName) {
		t.Errorf("errRegistration = %v, want also wrapping errInvalidName — errors.Join should accumulate both",
			errRegistration)
	}
}

// --- builtins ---

func TestBuiltins_returnsRegistryWhenNoErrors(t *testing.T) {
	v, err := NewBuiltin("YEAR", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	withRegistry(t, map[string]*Var{"YEAR": v}, nil)

	got, err := builtins()
	if err != nil {
		t.Fatalf("builtins: %v", err)
	}
	if len(got) != 1 || got["YEAR"] != v {
		t.Errorf("builtins() = %v, want {YEAR: %v}", got, v)
	}
}

func TestBuiltins_wrapsRegistrationErrAsUnexpected(t *testing.T) {
	withRegistry(t, map[string]*Var{}, errBoom)

	got, err := builtins()
	if got != nil {
		t.Errorf("builtins() map = %v, want nil", got)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want wrapping %v", err, errBoom)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d (ExitInternal)", e.ExitCode, errs.ExitInternal)
	}
}
