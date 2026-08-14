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

var errBadRegistration = errors.New("bad registration")

// withRegistry replaces the package-level builtin registry for the
// duration of t, restoring the original afterward. New reads registry and
// errRegistration indirectly through builtins(), so tests that exercise
// New need to control what's registered without depending on any real
// builtin package.
func withRegistry(t *testing.T, reg map[string]*Var, regErr error) {
	t.Helper()
	origRegistry, origErr := registry, errRegistration
	registry, errRegistration = reg, regErr
	t.Cleanup(func() {
		registry, errRegistration = origRegistry, origErr
	})
}

// --- New ---

func TestNew_seedsFromRegistryAndResolvesWireUpEagerly(t *testing.T) {
	calls := 0
	wireUp, err := NewWireUp("YEAR", func(*app.App) (string, error) {
		calls++
		return "2026", nil
	})
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	static, err := NewBuiltin("ORG", "LeafLock")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	withRegistry(t, map[string]*Var{"YEAR": wireUp, "ORG": static}, nil)

	v, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if len(v.base) != 2 {
		t.Errorf("len(base) = %d, want 2", len(v.base))
	}
	if calls != 1 {
		t.Errorf("WireUp compute called %d times during New, want 1 (eager resolve)", calls)
	}
	if got := v.base["YEAR"]; got == nil || !got.resolved || got.value != "2026" {
		t.Errorf("base[YEAR] = %+v, want resolved with value 2026", got)
	}
}

func TestNew_perCallLeftUndeclaredUntilSetCall(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME")
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	withRegistry(t, map[string]*Var{"FILE_NAME": perCall}, nil)

	v, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := v.base["FILE_NAME"]
	if got == nil {
		t.Fatal("base[FILE_NAME] missing")
	}
	if got.compute != nil {
		t.Error("compute is set, want nil — PerCall vars stay undeclared until SetCall")
	}
}

func TestNew_wireUpResolveErrorPropagates(t *testing.T) {
	broken, err := NewWireUp("BROKEN", func(*app.App) (string, error) { return "", errBoom })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	withRegistry(t, map[string]*Var{"BROKEN": broken}, nil)

	_, err = New(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want wrapping %v", err, errBoom)
	}
}

func TestNew_registrationErrorPropagatesAsUnexpected(t *testing.T) {
	withRegistry(t, map[string]*Var{}, errBadRegistration)

	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBadRegistration) {
		t.Errorf("error = %v, want wrapping %v", err, errBadRegistration)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d (ExitInternal)", e.ExitCode, errs.ExitInternal)
	}
}

// --- SetCall ---

func TestSetCall_attachesComputeToDeclaredPerCall(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME")
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	v := &Vars{base: map[string]*Var{"FILE_NAME": perCall}}

	if err := v.SetCall("file_name", func(*app.App) (string, error) { return "main.go", nil }); err != nil {
		t.Fatalf("SetCall: %v", err)
	}
	if perCall.compute == nil {
		t.Fatal("compute not attached")
	}
	got, err := perCall.resolve(nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "main.go" {
		t.Errorf("resolve = %q, want %q", got, "main.go")
	}
}

func TestSetCall_invalidNameIsUnexpected(t *testing.T) {
	v := &Vars{base: map[string]*Var{}}

	err := v.SetCall("bad name", func(*app.App) (string, error) { return "", nil })
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d (ExitInternal)", e.ExitCode, errs.ExitInternal)
	}
}

func TestSetCall_unknownNameIsUnexpected(t *testing.T) {
	v := &Vars{base: map[string]*Var{}}

	err := v.SetCall("NOPE", func(*app.App) (string, error) { return "", nil })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errSetCallUnknown) {
		t.Errorf("error = %v, want wrapping errSetCallUnknown", err)
	}
}

func TestSetCall_nonPerCallNameIsUnexpected(t *testing.T) {
	static, err := NewBuiltin("YEAR", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	v := &Vars{base: map[string]*Var{"YEAR": static}}

	err = v.SetCall("YEAR", func(*app.App) (string, error) { return "", nil })
	if !errors.Is(err, errSetCallUnknown) {
		t.Errorf("error = %v, want wrapping errSetCallUnknown — YEAR is Static, not PerCall", err)
	}
}

// --- Upsert ---

func TestUpsert_addsAndReplaces(t *testing.T) {
	existing, err := NewUserStatic("ORG", "Old")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	v := &Vars{base: map[string]*Var{"ORG": existing}}

	replacement, err := NewUserStatic("ORG", "New")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	addition, err := NewUserStatic("TEAM", "Platform")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}

	if err := v.Upsert(replacement, addition); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if v.base["ORG"].value != "New" {
		t.Errorf("base[ORG].value = %q, want %q", v.base["ORG"].value, "New")
	}
	if v.base["TEAM"].value != "Platform" {
		t.Errorf("base[TEAM].value = %q, want %q", v.base["TEAM"].value, "Platform")
	}
}

func TestUpsert_rejectsNonUserDefinedAndLeavesVarsUnmodified(t *testing.T) {
	existing, err := NewUserStatic("ORG", "Old")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	v := &Vars{base: map[string]*Var{"ORG": existing}}

	builtin, err := NewBuiltin("YEAR", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	goodUpdate, err := NewUserStatic("ORG", "New")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}

	err = v.Upsert(goodUpdate, builtin)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errUpsertOrigin) {
		t.Errorf("error = %v, want wrapping errUpsertOrigin", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d (ExitInternal)", e.ExitCode, errs.ExitInternal)
	}

	if v.base["ORG"].value != "Old" {
		t.Errorf("base[ORG].value = %q, want %q — Upsert must validate before mutating anything",
			v.base["ORG"].value, "Old")
	}
}
