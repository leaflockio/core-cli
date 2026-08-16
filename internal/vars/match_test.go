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

// newVars builds a *Vars from vs, keyed by each Var's own name — for
// tests that need a specific mix of Kind/Origin/Volatility not covered by
// newStaticVars.
func newVars(t *testing.T, vs ...*Var) *Vars {
	t.Helper()
	base := make(map[string]*Var, len(vs))
	for _, v := range vs {
		base[v.name] = v
	}
	return &Vars{base: base}
}

// --- usePattern ---

func TestUsePattern_staticAlwaysFalse(t *testing.T) {
	v, err := NewUserStatic("ORG", "LeafLock")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	for _, c := range []Comparison{Auto, Strict, Loose} {
		if v.usePattern(c) {
			t.Errorf("usePattern(%v) = true, want false for a Static var", c)
		}
	}
}

func TestUsePattern_nonStaticVolatile(t *testing.T) {
	v, err := NewWireUp("YEAR", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	tests := []struct {
		comparison Comparison
		want       bool
	}{
		{Auto, true},
		{Strict, false},
		{Loose, true},
	}
	for _, tt := range tests {
		if got := v.usePattern(tt.comparison); got != tt.want {
			t.Errorf("usePattern(%v) = %v, want %v", tt.comparison, got, tt.want)
		}
	}
}

func TestUsePattern_nonStaticStable(t *testing.T) {
	v, err := NewWireUp("GIT_OWNER", `[A-Za-z0-9]+`, Stable, func(*app.App) (string, error) {
		return "leaflockio", nil
	})
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	tests := []struct {
		comparison Comparison
		want       bool
	}{
		{Auto, false},
		{Strict, false},
		{Loose, true},
	}
	for _, tt := range tests {
		if got := v.usePattern(tt.comparison); got != tt.want {
			t.Errorf("usePattern(%v) = %v, want %v", tt.comparison, got, tt.want)
		}
	}
}

// --- Match: pattern vs literal selection ---

func TestMatch_autoUsesPatternForVolatileVar(t *testing.T) {
	year, err := NewWireUp("YEAR", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	v := newVars(t, year)

	re, err := v.Match("{YEAR}", Auto)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("1999") {
		t.Errorf("expected pattern mode to match a different year (1999), regex = %q", re.String())
	}
	if re.MatchString("abc") {
		t.Errorf("expected pattern mode to reject non-year text, regex = %q", re.String())
	}
}

func TestMatch_autoUsesLiteralForStableVar(t *testing.T) {
	owner, err := NewWireUp("GIT_OWNER", `[A-Za-z0-9]+`, Stable, func(*app.App) (string, error) {
		return "leaflockio", nil
	})
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	v := newVars(t, owner)

	re, err := v.Match("{GIT_OWNER}", Auto)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("leaflockio") {
		t.Errorf("expected literal mode to match the exact value, regex = %q", re.String())
	}
	if re.MatchString("owner2") {
		t.Errorf("expected literal mode to reject a different owner, regex = %q", re.String())
	}
}

func TestMatch_strictForcesLiteralEvenForVolatile(t *testing.T) {
	year, err := NewWireUp("YEAR", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	v := newVars(t, year)

	re, err := v.Match("{YEAR}", Strict)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("2026") {
		t.Errorf("expected Strict to still match the resolved value, regex = %q", re.String())
	}
	if re.MatchString("1999") {
		t.Errorf("expected Strict to reject a different year, regex = %q", re.String())
	}
}

func TestMatch_looseForcesPatternEvenForStable(t *testing.T) {
	owner, err := NewWireUp("GIT_OWNER", `[A-Za-z0-9]+`, Stable, func(*app.App) (string, error) {
		return "leaflockio", nil
	})
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	v := newVars(t, owner)

	re, err := v.Match("{GIT_OWNER}", Loose)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("owner2") {
		t.Errorf("expected Loose to pattern-match a different owner, regex = %q", re.String())
	}
}

func TestMatch_staticAlwaysLiteralRegardlessOfComparison(t *testing.T) {
	org, err := NewUserStatic("ORG", "A+B (Ltd).")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}

	for _, c := range []Comparison{Auto, Strict, Loose} {
		v := newVars(t, org)
		re, err := v.Match("{ORG}", c)
		if err != nil {
			t.Fatalf("Match(%v): %v", c, err)
		}
		if !re.MatchString("A+B (Ltd).") {
			t.Errorf("comparison=%v: expected exact literal match, regex = %q", c, re.String())
		}
		if re.MatchString("AB (Ltd)x") {
			t.Errorf("comparison=%v: expected regex special characters to be escaped, not interpreted, regex = %q",
				c, re.String())
		}
	}
}

// --- Match: literal template text is escaped ---

func TestMatch_literalTemplateTextIsEscaped(t *testing.T) {
	v := newVars(t)

	re, err := v.Match("A(B", Auto)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("A(B") {
		t.Errorf("expected literal unbalanced paren to match itself, regex = %q", re.String())
	}
}

// --- Match: composition ---

func TestMatch_composedStaticVarAppliesUsePatternToNestedRefs(t *testing.T) {
	year, err := NewWireUp("YEAR", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	org, err := NewUserStatic("ORG", "LeafLock")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	copyright, err := NewUserStatic("COPYRIGHT", "{YEAR} {ORG}")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	v := newVars(t, year, org, copyright)

	// COPYRIGHT is itself Static (always literal), but under Loose the
	// nested YEAR reference inside its value should still switch to
	// pattern mode independently.
	re, err := v.Match("{COPYRIGHT}", Loose)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("1999 LeafLock") {
		t.Errorf("expected nested YEAR to pattern-match under Loose, regex = %q", re.String())
	}
	if re.MatchString("1999 OtherOrg") {
		t.Errorf("expected nested ORG (Static) to stay literal even under Loose, regex = %q", re.String())
	}
}

// --- Match: errors ---

func TestMatch_unknownVariableIsCallerError(t *testing.T) {
	v := newVars(t)

	_, err := v.Match("{NOPE}", Auto)
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
}

func TestMatch_compositionCycleIsCallerError(t *testing.T) {
	a, err := NewUserStatic("A", "{B}")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	b, err := NewUserStatic("B", "{A}")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	v := newVars(t, a, b)

	_, err = v.Match("{A}", Auto)
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
}

// --- Match: not anchored ---

func TestMatch_returnsUnanchoredRegex(t *testing.T) {
	year, err := NewWireUp("YEAR", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	v := newVars(t, year)

	re, err := v.Match("{YEAR}", Auto)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("prefix 2026 suffix") {
		t.Errorf("expected an unanchored match within surrounding text, regex = %q", re.String())
	}
}

// --- Match: laziness ---

func TestMatch_onlyRequiresSetCallForReferencedPerCallVars(t *testing.T) {
	referenced, err := NewPerCall("FILE_NAME", `[A-Za-z0-9_.-]+`, Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	unreferenced, err := NewPerCall("FILE_CREATED_YEAR", `\d{4}`, Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	v := newVars(t, referenced, unreferenced)

	if err := v.SetCall("FILE_NAME", func(*app.App) (string, error) { return "main.go", nil }); err != nil {
		t.Fatalf("SetCall: %v", err)
	}
	// FILE_CREATED_YEAR is never SetCall'd, and never referenced by the
	// template below — Match must not require it.

	re, err := v.Match("{FILE_NAME}", Auto)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !re.MatchString("main.go") {
		t.Errorf("expected literal match on FILE_NAME, regex = %q", re.String())
	}
}

// --- Match: clearCalls ---

func TestMatch_clearsPerCallComputeAfterCall(t *testing.T) {
	perCall, err := NewPerCall("FILE_NAME", `[A-Za-z0-9_.-]+`, Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	perCall.compute = func(*app.App) (string, error) { return "main.go", nil }
	v := newVars(t, perCall)

	if _, err := v.Match("{FILE_NAME}", Auto); err != nil {
		t.Fatalf("Match: %v", err)
	}
	if perCall.compute != nil {
		t.Error("compute still set after Match, want nil — clearCalls should reset it")
	}
}
