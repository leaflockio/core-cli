// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"strconv"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
)

// newStaticVars builds a *Vars whose base holds one OriginUserDefined
// Static Var per entry in kv, for tests that only need fixed values.
func newStaticVars(t *testing.T, kv map[string]string) *Vars {
	t.Helper()
	base := make(map[string]*Var, len(kv))
	for k, val := range kv {
		v, err := NewUserStatic(k, val)
		if err != nil {
			t.Fatalf("NewUserStatic(%q): %v", k, err)
		}
		base[k] = v
	}
	return &Vars{base: base}
}

// --- Resolve: substitution ---

func TestResolve_substitutesKnownVars(t *testing.T) {
	v := newStaticVars(t, map[string]string{"NAME": "Leaf", "YEAR": "2026"})

	got, err := v.Resolve("{NAME} {YEAR}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "Leaf 2026" {
		t.Errorf("Resolve = %q, want %q", got, "Leaf 2026")
	}
}

func TestResolve_caseInsensitiveNameMatching(t *testing.T) {
	v := newStaticVars(t, map[string]string{"NAME": "Leaf"})

	got, err := v.Resolve("{name}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "Leaf" {
		t.Errorf("Resolve = %q, want %q", got, "Leaf")
	}
}

func TestResolve_noPlaceholdersReturnsUnchanged(t *testing.T) {
	v := newStaticVars(t, nil)

	got, err := v.Resolve("no vars here")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "no vars here" {
		t.Errorf("Resolve = %q, want unchanged", got)
	}
}

// --- Resolve: composition ---

func TestResolve_compositionExpandsRecursively(t *testing.T) {
	v := newStaticVars(t, map[string]string{
		"YEAR":      "2026",
		"ORG":       "LeafLock",
		"COPYRIGHT": "{YEAR} {ORG}",
	})

	got, err := v.Resolve("{COPYRIGHT}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "2026 LeafLock" {
		t.Errorf("Resolve = %q, want %q", got, "2026 LeafLock")
	}
}

func TestResolve_compositionCycleIsCallerError(t *testing.T) {
	v := newStaticVars(t, map[string]string{"A": "{B}", "B": "{A}"})

	_, err := v.Resolve("{A}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errCompositionCycle) {
		t.Errorf("error = %v, want wrapping errCompositionCycle", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR002 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR002)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

// --- Resolve: unknown / PerCall errors ---

func TestResolve_unknownVariableIsCallerError(t *testing.T) {
	v := newStaticVars(t, nil)

	_, err := v.Resolve("{NOPE}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errUnknownVariable) {
		t.Errorf("error = %v, want wrapping errUnknownVariable", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR003 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR003)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

func TestResolve_perCallWithNoValueIsUnexpected(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	v := &Vars{base: map[string]*Var{"FILE_NAME": perCall}}

	_, err = v.Resolve("{FILE_NAME}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errNoCompute) {
		t.Errorf("error = %v, want wrapping errNoCompute", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d (ExitInternal)", e.ExitCode, errs.ExitInternal)
	}
}

// --- Resolve: PerCall recompute / clearCalls ---

func TestResolve_perCallRecomputesForRepeatedReference(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	calls := 0
	perCall.compute = func(*app.App) (string, error) {
		calls++
		return strconv.Itoa(calls), nil
	}
	v := &Vars{base: map[string]*Var{"FILE_NAME": perCall}}

	got, err := v.Resolve("{FILE_NAME}-{FILE_NAME}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "1-2" {
		t.Errorf("Resolve = %q, want %q — PerCall must not cache within one Resolve", got, "1-2")
	}
}

func TestResolve_clearsPerCallComputeAfterCall(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	perCall.compute = func(*app.App) (string, error) { return "main.go", nil }
	v := &Vars{base: map[string]*Var{"FILE_NAME": perCall}}

	if _, err := v.Resolve("{FILE_NAME}"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if perCall.compute != nil {
		t.Error("compute still set after Resolve, want nil — clearCalls should reset it")
	}
}

func TestResolve_clearsPerCallComputeEvenWhenUnused(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	perCall.compute = func(*app.App) (string, error) { return "main.go", nil }
	v := &Vars{base: map[string]*Var{"FILE_NAME": perCall}}

	if _, err := v.Resolve("no reference to it here"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if perCall.compute != nil {
		t.Error("compute still set after Resolve, want nil — clearCalls runs whether or not the template used it")
	}
}
