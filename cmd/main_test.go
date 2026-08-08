// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

var errTestInternal = errors.New("unexpected failure")

func TestMain_noError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	runFn = func() error { return nil }
	exitCalled := false
	osExit = func(int) { exitCalled = true }

	out := captureStderr(t, main)

	if exitCalled {
		t.Error("expected osExit not to be called when run succeeds")
	}
	if !strings.Contains(out, "done in") {
		t.Errorf("expected duration to be printed even on success, got %q", out)
	}
}

func TestMain_callerError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	callerErr := errs.Caller(errs.ENV001, "bad env", nil)
	runFn = func() error { return callerErr }

	var got int
	osExit = func(code int) { got = code }

	out := captureStderr(t, main)

	if got != errs.ExitUser {
		t.Errorf("expected ExitUser (%d), got %d", errs.ExitUser, got)
	}
	assertErrorBeforeDuration(t, out)
}

func TestMain_internalError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	runFn = func() error { return errTestInternal }

	var got int
	osExit = func(code int) { got = code }

	out := captureStderr(t, main)

	if got != errs.ExitInternal {
		t.Errorf("expected ExitInternal (%d), got %d", errs.ExitInternal, got)
	}
	assertErrorBeforeDuration(t, out)
}

// assertErrorBeforeDuration checks that the error message errs.Print writes
// appears before the trailing duration line, not after.
func assertErrorBeforeDuration(t *testing.T, out string) {
	t.Helper()
	errIdx := strings.Index(out, "error")
	durIdx := strings.Index(out, "done in")
	if errIdx == -1 || durIdx == -1 {
		t.Fatalf("expected both an error message and a duration line, got %q", out)
	}
	if errIdx > durIdx {
		t.Errorf("expected error message before duration line, got %q", out)
	}
}
