// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
)

var (
	errBoom              = errors.New("boom")
	errShouldNotBeCalled = errors.New("should never be called")
)

// --- normalizeName ---

func TestNormalizeName_upcases(t *testing.T) {
	got, err := normalizeName("year")
	if err != nil {
		t.Fatalf("normalizeName: %v", err)
	}
	if got != "YEAR" {
		t.Errorf("normalizeName = %q, want %q", got, "YEAR")
	}
}

func TestNormalizeName_acceptsDigitsAndUnderscore(t *testing.T) {
	got, err := normalizeName("file_created_year_2")
	if err != nil {
		t.Fatalf("normalizeName: %v", err)
	}
	if got != "FILE_CREATED_YEAR_2" {
		t.Errorf("normalizeName = %q, want %q", got, "FILE_CREATED_YEAR_2")
	}
}

func TestNormalizeName_rejectsInvalidCharacters(t *testing.T) {
	for _, name := range []string{"MY VAR", "MY-VAR", "MY.VAR", ""} {
		if _, err := normalizeName(name); err == nil {
			t.Errorf("normalizeName(%q) = nil error, want an error", name)
		}
	}
}

func TestNormalizeName_errorWrapsErrInvalidName(t *testing.T) {
	_, err := normalizeName("bad name")
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
}

// --- NewBuiltin ---

func TestNewBuiltin_setsFields(t *testing.T) {
	v, err := NewBuiltin("year", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	if v.name != "YEAR" {
		t.Errorf("name = %q, want %q", v.name, "YEAR")
	}
	if v.origin != OriginBuiltin {
		t.Errorf("origin = %v, want OriginBuiltin", v.origin)
	}
	if v.kind != KindStatic {
		t.Errorf("kind = %v, want KindStatic", v.kind)
	}
	if !v.resolved {
		t.Error("resolved = false, want true")
	}
	if v.value != "2026" {
		t.Errorf("value = %q, want %q", v.value, "2026")
	}
	if v.pattern != "2026" {
		t.Errorf("pattern = %q, want %q (literal value, quoted)", v.pattern, "2026")
	}
	if v.volatility != Stable {
		t.Errorf("volatility = %v, want Stable", v.volatility)
	}
}

func TestNewBuiltin_invalidName(t *testing.T) {
	_, err := NewBuiltin("bad name", "x")
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
	var e *errs.Error
	if errors.As(err, &e) {
		t.Errorf("error is an *errs.Error (%v) — builtin constructors shouldn't classify, only NewUserStatic should", e)
	}
}

// --- NewWireUp ---

func TestNewWireUp_setsFields(t *testing.T) {
	v, err := NewWireUp("year", `\d{4}`, Volatile, func(*app.App) (string, error) { return "2026", nil })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}
	if v.name != "YEAR" {
		t.Errorf("name = %q, want %q", v.name, "YEAR")
	}
	if v.origin != OriginBuiltin {
		t.Errorf("origin = %v, want OriginBuiltin", v.origin)
	}
	if v.kind != KindWireUp {
		t.Errorf("kind = %v, want KindWireUp", v.kind)
	}
	if v.resolved {
		t.Error("resolved = true, want false before first resolve")
	}
	if v.compute == nil {
		t.Error("compute = nil, want the given function")
	}
	if v.pattern != `\d{4}` {
		t.Errorf("pattern = %q, want %q", v.pattern, `\d{4}`)
	}
	if v.volatility != Volatile {
		t.Errorf("volatility = %v, want Volatile", v.volatility)
	}
}

func TestNewWireUp_invalidName(t *testing.T) {
	_, err := NewWireUp("bad name", ".*", Stable, func(*app.App) (string, error) { return "", nil })
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
}

// --- NewUserStatic ---

func TestNewUserStatic_setsFields(t *testing.T) {
	v, err := NewUserStatic("my_var", "hello")
	if err != nil {
		t.Fatalf("NewUserStatic: %v", err)
	}
	if v.name != "MY_VAR" {
		t.Errorf("name = %q, want %q", v.name, "MY_VAR")
	}
	if v.origin != OriginUserDefined {
		t.Errorf("origin = %v, want OriginUserDefined", v.origin)
	}
	if v.kind != KindStatic {
		t.Errorf("kind = %v, want KindStatic", v.kind)
	}
	if !v.resolved {
		t.Error("resolved = false, want true")
	}
	if v.value != "hello" {
		t.Errorf("value = %q, want %q", v.value, "hello")
	}
	if v.pattern != "hello" {
		t.Errorf("pattern = %q, want %q (literal value, quoted)", v.pattern, "hello")
	}
	if v.volatility != Stable {
		t.Errorf("volatility = %v, want Stable", v.volatility)
	}
}

func TestNewUserStatic_invalidNameIsCallerError(t *testing.T) {
	_, err := NewUserStatic("bad name", "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR001)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

// --- NewPerCall ---

func TestNewPerCall_setsFields(t *testing.T) {
	v, err := NewPerCall("file_name", `[^\n]+`, Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	if v.name != "FILE_NAME" {
		t.Errorf("name = %q, want %q", v.name, "FILE_NAME")
	}
	if v.origin != OriginBuiltin {
		t.Errorf("origin = %v, want OriginBuiltin", v.origin)
	}
	if v.kind != KindPerCall {
		t.Errorf("kind = %v, want KindPerCall", v.kind)
	}
	if v.compute != nil {
		t.Error("compute is set, want nil until SetCall")
	}
	if v.pattern != `[^\n]+` {
		t.Errorf("pattern = %q, want %q", v.pattern, `[^\n]+`)
	}
	if v.volatility != Stable {
		t.Errorf("volatility = %v, want Stable", v.volatility)
	}
}

func TestNewPerCall_invalidName(t *testing.T) {
	_, err := NewPerCall("bad name", ".*", Stable)
	if !errors.Is(err, errInvalidName) {
		t.Errorf("error = %v, want wrapping errInvalidName", err)
	}
}

// --- (*Var).resolve ---

func TestVarResolve_staticReturnsValueWithoutCompute(t *testing.T) {
	v, err := NewBuiltin("YEAR", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}

	got, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "2026" {
		t.Errorf("resolve = %q, want %q", got, "2026")
	}
}

func TestVarResolve_staticIgnoresCompute(t *testing.T) {
	v, err := NewBuiltin("YEAR", "2026")
	if err != nil {
		t.Fatalf("NewBuiltin: %v", err)
	}
	v.compute = func(*app.App) (string, error) { return "", errShouldNotBeCalled }

	got, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "2026" {
		t.Errorf("resolve = %q, want %q — resolved:true must skip compute entirely", got, "2026")
	}
}

func TestVarResolve_wireUpCachesAfterFirstCall(t *testing.T) {
	calls := 0
	fn := func(*app.App) (string, error) {
		calls++
		return strconv.Itoa(calls), nil
	}
	v, err := NewWireUp("X", ".*", Stable, fn)
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}

	first, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve (1st): %v", err)
	}
	second, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve (2nd): %v", err)
	}

	if first != "1" || second != "1" {
		t.Errorf("resolve calls = (%q, %q), want (\"1\", \"1\") — WireUp should cache", first, second)
	}
	if calls != 1 {
		t.Errorf("compute called %d times, want 1", calls)
	}
}

func TestVarResolve_wireUpRetriesAfterComputeError(t *testing.T) {
	called := 0
	fn := func(*app.App) (string, error) {
		called++
		if called == 1 {
			return "", errBoom
		}
		return "ok", nil
	}
	v, err := NewWireUp("X", ".*", Stable, fn)
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}

	if _, err := v.resolve(nil); err == nil {
		t.Fatal("expected error on first resolve, got nil")
	} else if !errors.Is(err, errBoom) {
		t.Errorf("resolve error = %v, want wrapping %v", err, errBoom)
	}

	got, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve (retry): %v", err)
	}
	if got != "ok" {
		t.Errorf("resolve (retry) = %q, want %q", got, "ok")
	}
	if called != 2 {
		t.Errorf("compute called %d times, want 2 — a failed resolve must not cache", called)
	}
}

func TestVarResolve_wireUpNilComputeIsUnexpected(t *testing.T) {
	v, err := NewWireUp("X", ".*", Stable, nil)
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}

	_, err = v.resolve(nil)
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

func TestVarResolve_perCallRecomputesEveryCall(t *testing.T) {
	v, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}
	calls := 0
	v.compute = func(*app.App) (string, error) {
		calls++
		return strconv.Itoa(calls), nil
	}

	first, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve (1st): %v", err)
	}
	second, err := v.resolve(nil)
	if err != nil {
		t.Fatalf("resolve (2nd): %v", err)
	}

	if first != "1" || second != "2" {
		t.Errorf("resolve calls = (%q, %q), want (\"1\", \"2\") — PerCall must not cache", first, second)
	}
	if calls != 2 {
		t.Errorf("compute called %d times, want 2", calls)
	}
}

func TestVarResolve_perCallNilComputeIsUnexpected(t *testing.T) {
	v, err := NewPerCall("FILE_NAME", ".*", Stable)
	if err != nil {
		t.Fatalf("NewPerCall: %v", err)
	}

	_, err = v.resolve(nil)
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

func TestVarResolve_computeErrorIsWrappedWithName(t *testing.T) {
	v, err := NewWireUp("GIT_OWNER", ".*", Stable, func(*app.App) (string, error) { return "", errBoom })
	if err != nil {
		t.Fatalf("NewWireUp: %v", err)
	}

	_, err = v.resolve(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want wrapping %v", err, errBoom)
	}
	if !strings.Contains(err.Error(), "GIT_OWNER") {
		t.Errorf("error = %q, want it to mention the variable name", err.Error())
	}
}
