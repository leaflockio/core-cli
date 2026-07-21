// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

// testCode is a stable code used across tests. It is not a real error
// code — it exists only to exercise the errs machinery.
const testCode errs.Code = "TEST001"

var errUnderlying = errors.New("underlying error")

func TestError_Error(t *testing.T) {
	e := &errs.Error{Message: "something failed"}
	if got := e.Error(); got != "something failed" {
		t.Errorf("Error() = %q, want %q", got, "something failed")
	}
}

func TestError_Unwrap(t *testing.T) {
	e := &errs.Error{Err: errUnderlying}
	if !errors.Is(e, errUnderlying) {
		t.Error("errors.Is failed to match through Unwrap")
	}
}

func TestCaller(t *testing.T) {
	e := errs.Caller(testCode, "bad input", errUnderlying)

	if e.Code != testCode {
		t.Errorf("Code = %q, want %q", e.Code, testCode)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d", e.ExitCode, errs.ExitUser)
	}
	if !errors.Is(e, errUnderlying) {
		t.Error("errors.Is failed through Caller error chain")
	}
}

func TestCaller_withContexts(t *testing.T) {
	ctx := errs.Context{Cause: "value is wrong", Resolution: "set it to one of: a, b, c"}
	e := errs.Caller(testCode, "bad input", nil, ctx)

	if len(e.Contexts) != 1 {
		t.Fatalf("len(Contexts) = %d, want 1", len(e.Contexts))
	}
	if e.Contexts[0].Cause != ctx.Cause {
		t.Errorf("Cause = %q, want %q", e.Contexts[0].Cause, ctx.Cause)
	}
	if e.Contexts[0].Resolution != ctx.Resolution {
		t.Errorf("Resolution = %q, want %q", e.Contexts[0].Resolution, ctx.Resolution)
	}
}

func TestUnexpected(t *testing.T) {
	e := errs.Unexpected(errUnderlying)

	if e.ExitCode != errs.ExitInternal {
		t.Errorf("ExitCode = %d, want %d", e.ExitCode, errs.ExitInternal)
	}
	if !errors.Is(e, errUnderlying) {
		t.Error("errors.Is failed through Unexpected error chain")
	}
}

func TestValidation(t *testing.T) {
	e := errs.Validation(testCode, "header missing", errUnderlying)

	if e.Code != testCode {
		t.Errorf("Code = %q, want %q", e.Code, testCode)
	}
	if e.ExitCode != errs.ExitValidation {
		t.Errorf("ExitCode = %d, want %d", e.ExitCode, errs.ExitValidation)
	}
	if !errors.Is(e, errUnderlying) {
		t.Error("errors.Is failed through Validation error chain")
	}
}

func TestValidation_withContexts(t *testing.T) {
	ctx := errs.Context{Cause: "file lacks header", Resolution: "run leaf license add"}
	e := errs.Validation(testCode, "header missing", nil, ctx)

	if len(e.Contexts) != 1 {
		t.Fatalf("len(Contexts) = %d, want 1", len(e.Contexts))
	}
	if e.Contexts[0].Cause != ctx.Cause {
		t.Errorf("Cause = %q, want %q", e.Contexts[0].Cause, ctx.Cause)
	}
	if e.Contexts[0].Resolution != ctx.Resolution {
		t.Errorf("Resolution = %q, want %q", e.Contexts[0].Resolution, ctx.Resolution)
	}
}

func TestErrorsAs_throughWrap(t *testing.T) {
	wrapped := fmt.Errorf("outer: %w", errs.Caller(testCode, "bad input", nil))

	var e *errs.Error
	if !errors.As(wrapped, &e) {
		t.Fatal("errors.As failed to unwrap to *errs.Error")
	}
	if e.Code != testCode {
		t.Errorf("Code = %q, want %q", e.Code, testCode)
	}
}
