// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"errors"
	"testing"

	"github.com/leaflock/core-cli/internal/errs"
)

var errTestInternal = errors.New("unexpected failure")

func TestMain_noError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	runFn = func() error { return nil }
	exitCalled := false
	osExit = func(int) { exitCalled = true }

	main()

	if exitCalled {
		t.Error("expected osExit not to be called when run succeeds")
	}
}

func TestMain_callerError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	callerErr := errs.Caller(errs.ENV001, "bad env", nil)
	runFn = func() error { return callerErr }

	var got int
	osExit = func(code int) { got = code }

	main()

	if got != errs.ExitUser {
		t.Errorf("expected ExitUser (%d), got %d", errs.ExitUser, got)
	}
}

func TestMain_internalError(t *testing.T) {
	origRun, origExit := runFn, osExit
	defer func() { runFn, osExit = origRun, origExit }()

	runFn = func() error { return errTestInternal }

	var got int
	osExit = func(code int) { got = code }

	main()

	if got != errs.ExitInternal {
		t.Errorf("expected ExitInternal (%d), got %d", errs.ExitInternal, got)
	}
}
